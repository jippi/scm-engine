package github

import (
	"context"
	"log/slog"
	"strings"

	go_github "github.com/google/go-github/v64/github"
	"github.com/jippi/scm-engine/pkg/scm"
	slogctx "github.com/veqryn/slog-context"
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
		IncludeAncestorGroups: scm.Ptr(true),
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
	owner, repo := ownerAndRepo(ctx)

	githubLabels, response, err := client.client.wrapped.Issues.ListLabels(ctx, owner, repo, &go_github.ListOptions{PerPage: opt.PerPage, Page: opt.Page})
	if err != nil {
		return nil, convertResponse(response), err
	}

	labels := make([]*scm.Label, 0)

	for _, label := range githubLabels {
		labels = append(labels, convertLabel(label))
	}

	return labels, convertResponse(response), nil
}

func (client *LabelClient) Create(ctx context.Context, opt *scm.CreateLabelOptions) (*scm.Label, *scm.Response, error) {
	// Invalidate cache
	client.cache = nil

	owner, repo := ownerAndRepo(ctx)

	label, resp, err := client.client.wrapped.Issues.CreateLabel(ctx, owner, repo, &go_github.Label{
		Name:        opt.Name,
		Description: opt.Description,
		Color:       scm.Ptr(strings.TrimPrefix(*opt.Color, "#")),
	})

	return convertLabel(label), convertResponse(resp), err
}

func (client *LabelClient) Update(ctx context.Context, opt *scm.UpdateLabelOptions) (*scm.Label, *scm.Response, error) {
	// Invalidate cache
	client.cache = nil

	owner, repo := ownerAndRepo(ctx)

	updateLabel := &go_github.Label{}
	updateLabel.Name = opt.Name
	updateLabel.Color = scm.Ptr(strings.TrimPrefix(*opt.Color, "#"))
	updateLabel.Description = opt.Description

	label, resp, err := client.client.wrapped.Issues.EditLabel(ctx, owner, repo, *opt.Name, updateLabel)

	return convertLabel(label), convertResponse(resp), err
}

func convertLabel(label *go_github.Label) *scm.Label {
	if label == nil {
		return nil
	}

	return &scm.Label{
		Name:        *label.Name,
		Description: *label.Description,
		Color:       *label.Color,
	}
}
