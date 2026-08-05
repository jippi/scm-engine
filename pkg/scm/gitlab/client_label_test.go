package gitlab_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/scm/gitlab"
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/stretchr/testify/require"
)

// labelServer stands in for the GitLab labels API and counts how many requests
// it served, which is how the cache and the pagination loop are observed.
type labelServer struct {
	*httptest.Server

	requests atomic.Int64
}

func newLabelClient(t *testing.T, handler http.HandlerFunc) (scm.LabelClient, *labelServer) {
	t.Helper()

	server := &labelServer{}
	server.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.requests.Add(1)

		handler(w, r)
	}))
	t.Cleanup(server.Close)

	ctx := state.WithToken(t.Context(), "token")
	ctx = state.WithBaseURL(ctx, server.URL)
	ctx = state.WithProjectID(ctx, "jippi/scm-engine")

	client, err := gitlab.NewClient(ctx, nil)
	require.NoError(t, err)

	return client.Labels(), server
}

func labelContext(t *testing.T) (ctx context.Context) { //nolint:nonamedreturns // documents the return
	t.Helper()

	ctx = state.WithProjectID(t.Context(), "jippi/scm-engine")
	ctx = state.WithDryRun(ctx, false)

	return ctx
}

func TestLabelClient_List(t *testing.T) {
	t.Parallel()

	client, server := newLabelClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"id":1,"name":"bug","color":"#FF0000"},{"id":2,"name":"feature"}]`)
	})

	labels, err := client.List(labelContext(t))
	require.NoError(t, err)
	require.Len(t, labels, 2)
	require.Equal(t, "bug", labels[0].Name)
	require.Equal(t, "#FF0000", labels[0].Color)
	require.Equal(t, "feature", labels[1].Name)

	require.Equal(t, int64(1), server.requests.Load())
}

// The label list is cached for the lifetime of an evaluation, so repeated
// lookups do not re-query GitLab.
func TestLabelClient_List_isCached(t *testing.T) {
	t.Parallel()

	client, server := newLabelClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"id":1,"name":"bug"}]`)
	})

	ctx := labelContext(t)

	first, err := client.List(ctx)
	require.NoError(t, err)

	second, err := client.List(ctx)
	require.NoError(t, err)

	require.Equal(t, first, second)
	require.Equal(t, int64(1), server.requests.Load(), "the second call must be served from cache")
}

// GitLab paginates labels, so every page has to be followed or a project with
// more than 100 labels would silently look like it only had the first page.
func TestLabelClient_List_followsPagination(t *testing.T) {
	t.Parallel()

	client, server := newLabelClient(t, func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}

		w.Header().Set("Content-Type", "application/json")

		switch page {
		case "1":
			w.Header().Set("X-Next-Page", "2")
			fmt.Fprint(w, `[{"id":1,"name":"page-one"}]`)

		default:
			w.Header().Set("X-Next-Page", "")
			fmt.Fprint(w, `[{"id":2,"name":"page-two"}]`)
		}
	})

	labels, err := client.List(labelContext(t))
	require.NoError(t, err)

	require.Len(t, labels, 2)
	require.Equal(t, "page-one", labels[0].Name)
	require.Equal(t, "page-two", labels[1].Name)
	require.Equal(t, int64(2), server.requests.Load())
}

func TestLabelClient_List_propagatesError(t *testing.T) {
	t.Parallel()

	client, _ := newLabelClient(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"message":"401 Unauthorized"}`, http.StatusUnauthorized)
	})

	_, err := client.List(labelContext(t))
	require.Error(t, err)
}

func TestLabelClient_Create(t *testing.T) {
	t.Parallel()

	client, server := newLabelClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"id":10,"name":"brand-new","color":"#00FF00"}`)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[]`)
	})

	label, _, err := client.Create(labelContext(t), &scm.CreateLabelOptions{
		Name:  scm.Ptr("brand-new"),
		Color: scm.Ptr("#00FF00"),
	})
	require.NoError(t, err)
	require.Equal(t, "brand-new", label.Name)

	require.Positive(t, server.requests.Load())
}

// Creating or updating a label invalidates the cache, otherwise the rest of the
// evaluation would keep seeing the stale list.
func TestLabelClient_Create_invalidatesCache(t *testing.T) {
	t.Parallel()

	var created atomic.Bool

	client, server := newLabelClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodPost {
			created.Store(true)
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"id":2,"name":"added"}`)

			return
		}

		if created.Load() {
			fmt.Fprint(w, `[{"id":1,"name":"existing"},{"id":2,"name":"added"}]`)

			return
		}

		fmt.Fprint(w, `[{"id":1,"name":"existing"}]`)
	})

	ctx := labelContext(t)

	before, err := client.List(ctx)
	require.NoError(t, err)
	require.Len(t, before, 1)

	_, _, err = client.Create(ctx, &scm.CreateLabelOptions{Name: scm.Ptr("added")})
	require.NoError(t, err)

	after, err := client.List(ctx)
	require.NoError(t, err)
	require.Len(t, after, 2, "the cache must have been invalidated by the create")

	require.GreaterOrEqual(t, server.requests.Load(), int64(3))
}

func TestLabelClient_Update(t *testing.T) {
	t.Parallel()

	client, _ := newLabelClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodPut {
			fmt.Fprint(w, `{"id":1,"name":"bug","description":"updated"}`)

			return
		}

		fmt.Fprint(w, `[]`)
	})

	label, _, err := client.Update(labelContext(t), &scm.UpdateLabelOptions{
		Name:        scm.Ptr("bug"),
		Description: scm.Ptr("updated"),
	})
	require.NoError(t, err)
	require.Equal(t, "updated", label.Description)
}

func TestLabelClient_Update_propagatesError(t *testing.T) {
	t.Parallel()

	client, _ := newLabelClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			http.Error(w, `{"message":"404 Not Found"}`, http.StatusNotFound)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[]`)
	})

	_, _, err := client.Update(labelContext(t), &scm.UpdateLabelOptions{Name: scm.Ptr("missing")})
	require.Error(t, err)
}
