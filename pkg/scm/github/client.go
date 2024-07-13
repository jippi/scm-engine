package github

import (
	"context"

	go_github "github.com/google/go-github/v63/github"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
)

// Ensure the GitLab client implements the [scm.Client]
var _ scm.Client = (*Client)(nil)

// Client is a wrapper around the GitLab specific implementation of [scm.Client] interface
type Client struct {
	wrapped *go_github.Client

	labels        *LabelClient
	mergeRequests *MergeRequestClient
}

// NewClient creates a new GitLab client
func NewClient(ctx context.Context) *Client {
	client := go_github.NewClient(nil).WithAuthToken(state.Token(ctx))

	return &Client{wrapped: client}
}

// Labels returns a client target at managing labels/tags
func (client *Client) Labels() scm.LabelClient {
	if client.labels == nil {
		client.labels = NewLabelClient(client)
	}

	return client.labels
}

// MergeRequests returns a client target at managing merge/pull requests
func (client *Client) MergeRequests() scm.MergeRequestClient {
	if client.mergeRequests == nil {
		client.mergeRequests = NewMergeRequestClient(client)
	}

	return client.mergeRequests
}

// EvalContext creates a new evaluation context for GitLab specific usage
func (client *Client) EvalContext(ctx context.Context) (scm.EvalContext, error) {
	res, err := NewContext(ctx, "https://api.github.com/", state.Token(ctx))
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Start pipeline
func (client *Client) Start(ctx context.Context) error {
	return nil
}

// Stop pipeline
func (client *Client) Stop(ctx context.Context, err error) error {
	return nil
}
