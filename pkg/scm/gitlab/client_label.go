package gitlab

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jippi/gitlab-labeller/pkg/scm"
	"github.com/jippi/gitlab-labeller/pkg/state"
	go_gitlab "github.com/xanzy/go-gitlab"
)

var _ scm.LabelClient = (*LabelClient)(nil)

type LabelClient struct {
	client *Client
}

func NewLabelClient(client *Client) *LabelClient {
	return &LabelClient{client: client}
}

func (client *LabelClient) List(ctx context.Context) ([]*scm.Label, error) {
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
		fmt.Println("Reading labels page", opts.Page)
		labels, resp, err := client.list(ctx, opts)
		if err != nil {
			panic(err)
		}

		results = append(results, labels...)

		if resp.NextPage == 0 {
			break
		}

		opts.ListOptions.Page = resp.NextPage
	}

	return results, nil
}

func (client *LabelClient) list(ctx context.Context, opt *scm.ListLabelsOptions) ([]*scm.Label, *scm.Response, error) {
	project, err := ParseID(state.ProjectIDFromContext(ctx))
	if err != nil {
		return nil, nil, err
	}

	u := fmt.Sprintf("projects/%s/labels", go_gitlab.PathEscape(project))

	options := []go_gitlab.RequestOptionFunc{
		go_gitlab.WithContext(ctx),
	}

	req, err := client.client.wrapped.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var l []*scm.Label
	resp, err := client.client.wrapped.Do(req, &l)
	if err != nil {
		return nil, convertResponse(resp), err
	}

	return l, convertResponse(resp), nil
}

func (client *LabelClient) Create(ctx context.Context, opt *scm.CreateLabelOptions) (*scm.Label, *scm.Response, error) {
	project, err := ParseID(state.ProjectIDFromContext(ctx))
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/labels", go_gitlab.PathEscape(project))

	options := []go_gitlab.RequestOptionFunc{
		go_gitlab.WithContext(ctx),
	}

	req, err := client.client.wrapped.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	l := new(scm.Label)
	resp, err := client.client.wrapped.Do(req, l)
	if err != nil {
		return nil, convertResponse(resp), err
	}

	return l, convertResponse(resp), nil
}

func (client *LabelClient) Update(ctx context.Context, opt *scm.UpdateLabelOptions) (*scm.Label, *scm.Response, error) {
	project, err := ParseID(state.ProjectIDFromContext(ctx))
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/labels", go_gitlab.PathEscape(project))

	options := []go_gitlab.RequestOptionFunc{
		go_gitlab.WithContext(ctx),
	}

	req, err := client.client.wrapped.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	l := new(scm.Label)
	resp, err := client.client.wrapped.Do(req, l)
	if err != nil {
		return nil, convertResponse(resp), err
	}

	return l, convertResponse(resp), nil
}
