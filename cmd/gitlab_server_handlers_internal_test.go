//nolint:testpackage // the webhook handler is built from unexported helpers and package level state
package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jippi/scm-engine/pkg/state"
	"github.com/stretchr/testify/require"
)

var jsonHeaders = map[string]string{"Content-Type": "application/json"}

func TestGitLabStatusHandler(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()

	GitLabStatusHandler(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/_status", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), "scm-engine status: OK")
}

// newWebhook builds the handler against a local stand-in for GitLab, so a
// payload that gets past validation fails on a controlled 404 instead of
// reaching the network.
//
// The handler reads its state from the request context, which in production
// comes from the http.Server BaseContext set up in Server(); the returned
// function mirrors that wiring.
func newWebhook(t *testing.T, secret string) func(string, map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"message":"404 Not Found"}`, http.StatusNotFound)
	}))
	t.Cleanup(upstream.Close)

	ctx := state.WithProvider(t.Context(), "gitlab")
	ctx = state.WithToken(ctx, "token")
	ctx = state.WithBaseURL(ctx, upstream.URL)
	ctx = state.WithBackstageURL(ctx, "")
	ctx = state.WithBackstageToken(ctx, "")
	ctx = state.WithConfigFilePath(ctx, ".scm-engine.yml")
	ctx = state.WithGlobalConfigFilePath(ctx, "")
	ctx = state.WithDryRun(ctx, true)
	ctx = state.WithUpdatePipeline(ctx, false, "")

	handler := GitLabWebhookHandler(ctx, secret)

	return func(body string, headers map[string]string) *httptest.ResponseRecorder {
		req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/gitlab", strings.NewReader(body))

		for key, value := range headers {
			req.Header.Set(key, value)
		}

		recorder := httptest.NewRecorder()
		handler(recorder, req)

		return recorder
	}
}

func TestGitLabWebhookHandler_rejectsBadSecret(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		headers map[string]string
	}{
		{name: "no token header", headers: jsonHeaders},
		{
			name:    "wrong token",
			headers: map[string]string{"Content-Type": "application/json", "X-Gitlab-Token": "nope"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := newWebhook(t, "expected-secret")(`{}`, tt.headers)

			require.Equal(t, http.StatusForbidden, recorder.Code)
			require.Contains(t, recorder.Body.String(), "Missing or invalid X-Gitlab-Token")
		})
	}
}

func TestGitLabWebhookHandler_acceptsMatchingSecret(t *testing.T) {
	t.Parallel()

	recorder := newWebhook(t, "expected-secret")(`{"event_type":"nope"}`, map[string]string{
		"Content-Type":   "application/json",
		"X-Gitlab-Token": "expected-secret",
	})

	require.NotEqual(t, http.StatusForbidden, recorder.Code)
}

// With no secret configured the token header must not be required at all.
func TestGitLabWebhookHandler_secretIsOptional(t *testing.T) {
	t.Parallel()

	recorder := newWebhook(t, "")(`{"event_type":"nope"}`, jsonHeaders)

	require.NotEqual(t, http.StatusForbidden, recorder.Code)
}

func TestGitLabWebhookHandler_requiresJSONContentType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		contentType string
	}{
		{name: "missing", contentType: ""},
		{name: "form encoded", contentType: "application/x-www-form-urlencoded"},
		{name: "text", contentType: "text/plain"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			headers := map[string]string{}
			if tt.contentType != "" {
				headers["Content-Type"] = tt.contentType
			}

			recorder := newWebhook(t, "")(`{}`, headers)

			require.Equal(t, http.StatusNotAcceptable, recorder.Code)
			require.Contains(t, recorder.Body.String(), "Content-Type: application/json")
		})
	}
}

// An empty body used to fall through into the JSON decoder after having already
// written a 400, producing two responses for a single request.
func TestGitLabWebhookHandler_rejectsEmptyBody(t *testing.T) {
	t.Parallel()

	recorder := newWebhook(t, "")("", jsonHeaders)

	require.Equal(t, http.StatusBadRequest, recorder.Code)

	body := recorder.Body.String()
	require.Contains(t, body, "The POST body is empty")
	require.NotContains(t, body, "could not decode POST body",
		"the handler must stop after reporting the empty body, not fall through to the decoder")
}

func TestGitLabWebhookHandler_rejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	recorder := newWebhook(t, "")(`{"broken":`, jsonHeaders)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), "could not decode POST body")
}

func TestGitLabWebhookHandler_rejectsUnknownEventType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{name: "unsupported event", body: `{"event_type":"pipeline"}`},
		{name: "no event type at all", body: `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := newWebhook(t, "")(tt.body, jsonHeaders)

			require.Equal(t, http.StatusInternalServerError, recorder.Code)
			require.Contains(t, recorder.Body.String(), "unknown event type")
		})
	}
}

// The two supported event types carry the Merge Request IID and commit SHA in
// different places in the payload, so both shapes have to be understood.
func TestGitLabWebhookHandler_acceptsSupportedEventTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{
			name: "merge_request event",
			body: `{"event_type":"merge_request","project":{"path_with_namespace":"jippi/scm-engine"},` +
				`"object_attributes":{"iid":42,"last_commit":{"id":"abc123"}}}`,
		},
		{
			name: "note event",
			body: `{"event_type":"note","project":{"path_with_namespace":"jippi/scm-engine"},` +
				`"merge_request":{"iid":42,"last_commit":{"id":"abc123"}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := newWebhook(t, "")(tt.body, jsonHeaders)

			// The payload is understood, so the handler gets past validation and
			// on to fetching the config, which the stand-in answers with a 404.
			require.NotEqual(t, http.StatusInternalServerError, recorder.Code)
			require.NotContains(t, recorder.Body.String(), "unknown event type")
		})
	}
}
