package gitlab

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	slogctx "github.com/veqryn/slog-context"
	go_gitlab "github.com/xanzy/go-gitlab"
)

var _ scm.LabelClient = (*LabelClient)(nil)

type LabelClient struct {
	client *Client

	cache []*scm.Label
}

func NewLabelClient(client *Client) *LabelClient {
	return &LabelClient{client: client}
}

func (client *LabelClient) List(ctx context.Context) ([]*scm.Label, error) {
	// Check cache
	if len(client.cache) != 0 {
		return client.cache, nil
	}

	var results []*scm.Label

	// Load all existing labels
	opts := &scm.ListLabelsOptions{
		IncludeAncestorGroups: go_gitlab.Ptr(true),
		ListOptions: scm.ListOptions{
			PerPage: 100,
			Page:    1,
		},
	}

	for {
		slogctx.Info(ctx, "Reading labels page", slog.Int("page", opts.Page))

		labels, resp, err := client.list(ctx, opts)
		if err != nil {
			return nil, err
		}

		results = append(results, labels...)

		if resp.NextPage == 0 {
			break
		}

		opts.ListOptions.Page = resp.NextPage
	}

	// Store cache
	client.cache = results

	return results, nil
}

func (client *LabelClient) list(ctx context.Context, opt *scm.ListLabelsOptions) ([]*scm.Label, *scm.Response, error) {
	project, err := ParseID(state.ProjectID(ctx))
	if err != nil {
		return nil, nil, err
	}

	endpoint := fmt.Sprintf("projects/%s/labels", go_gitlab.PathEscape(project))

	options := []go_gitlab.RequestOptionFunc{
		go_gitlab.WithContext(ctx),
	}

	req, err := client.client.wrapped.NewRequest(http.MethodGet, endpoint, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var labels []*scm.Label

	resp, err := client.client.wrapped.Do(req, &labels)
	if err != nil {
		return nil, convertResponse(resp), err
	}

	return labels, convertResponse(resp), nil
}

func (client *LabelClient) Create(ctx context.Context, opt *scm.CreateLabelOptions) (*scm.Label, *scm.Response, error) {
	// Invalidate cache
	client.cache = nil

	project, err := ParseID(state.ProjectID(ctx))
	if err != nil {
		return nil, nil, err
	}

	endpoint := fmt.Sprintf("projects/%s/labels", go_gitlab.PathEscape(project))

	options := []go_gitlab.RequestOptionFunc{
		go_gitlab.WithContext(ctx),
	}

	req, err := client.client.wrapped.NewRequest(http.MethodPost, endpoint, opt, options)
	if err != nil {
		return nil, nil, err
	}

	label := new(scm.Label)

	resp, err := client.client.wrapped.Do(req, label)
	if err != nil {
		return nil, convertResponse(resp), err
	}

	return label, convertResponse(resp), nil
}

func (client *LabelClient) Update(ctx context.Context, opt *scm.UpdateLabelOptions) (*scm.Label, *scm.Response, error) {
	// Invalidate cache
	client.cache = nil

	project, err := ParseID(state.ProjectID(ctx))
	if err != nil {
		return nil, nil, err
	}

	endpoint := fmt.Sprintf("projects/%s/labels", go_gitlab.PathEscape(project))

	options := []go_gitlab.RequestOptionFunc{
		go_gitlab.WithContext(ctx),
	}

	req, err := client.client.wrapped.NewRequest(http.MethodPut, endpoint, opt, options)
	if err != nil {
		return nil, nil, err
	}

	label := new(scm.Label)

	resp, err := client.client.wrapped.Do(req, label)
	if err != nil {
		return nil, convertResponse(resp), err
	}

	return label, convertResponse(resp), nil
}
