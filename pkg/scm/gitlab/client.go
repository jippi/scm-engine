package gitlab

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	slogctx "github.com/veqryn/slog-context"
	go_gitlab "github.com/xanzy/go-gitlab"
)

var pipelineName = scm.Ptr("scm-engine")

// Ensure the GitLab client implements the [scm.Client]
var _ scm.Client = (*Client)(nil)

// Client is a wrapper around the GitLab specific implementation of [scm.Client] interface
type Client struct {
	wrapped *go_gitlab.Client

	labels        *LabelClient
	mergeRequests *MergeRequestClient
}

// NewClient creates a new GitLab client
func NewClient(ctx context.Context) (*Client, error) {
	client, err := go_gitlab.NewClient(state.Token(ctx), go_gitlab.WithBaseURL(state.BaseURL(ctx)))
	if err != nil {
		return nil, err
	}

	return &Client{wrapped: client}, nil
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
	return NewContext(ctx, graphqlBaseURL(client.wrapped.BaseURL()), state.Token(ctx))
}

// Start pipeline
func (client *Client) Start(ctx context.Context) error {
	if !state.ShouldUpdatePipeline(ctx) {
		return nil
	}

	_, response, err := client.wrapped.Commits.SetCommitStatus(state.ProjectID(ctx), state.CommitSHA(ctx), &go_gitlab.SetCommitStatusOptions{
		State:       go_gitlab.Running,
		Context:     pipelineName,
		Description: scm.Ptr("Currently evaluating MR"),
	})

	switch response.StatusCode {
	// GitLab returns '400 {message: {name: [has already been taken]}}' if the context/pipeline name we use
	// are already used by a regular CI job. We treat that as a non-failure and continue on after logging
	case http.StatusBadRequest:
		slogctx.Warn(ctx, "could not update commit pipeline status", slog.Any("err", err))

		return nil

	default:
		return err
	}
}

// Stop pipeline
func (client *Client) Stop(ctx context.Context, err error) error {
	if !state.ShouldUpdatePipeline(ctx) {
		return nil
	}

	status := go_gitlab.Success
	message := "OK"

	if err != nil {
		status = go_gitlab.Failed
		message = err.Error()
	}

	_, response, err := client.wrapped.Commits.SetCommitStatus(state.ProjectID(ctx), state.CommitSHA(ctx), &go_gitlab.SetCommitStatusOptions{
		State:       status,
		Context:     pipelineName,
		Description: scm.Ptr(message),
	})

	switch response.StatusCode {
	// GitLab returns '400 {message: {name: [has already been taken]}}' if the context/pipeline name we use
	// are already used by a regular CI job. We treat that as a non-failure and continue on after logging
	case http.StatusBadRequest:
		slogctx.Warn(ctx, "could not update commit pipeline status", slog.Any("err", err))

		return nil

	default:
		return err
	}
}

func graphqlBaseURL(inputURL *url.URL) string {
	var buf strings.Builder
	if inputURL.Scheme != "" {
		buf.WriteString(inputURL.Scheme)
		buf.WriteByte(':')
	}

	if inputURL.Scheme != "" || inputURL.Host != "" || inputURL.User != nil {
		if inputURL.OmitHost && inputURL.Host == "" && inputURL.User == nil {
			// omit empty host
		} else {
			if inputURL.Host != "" || inputURL.Path != "" || inputURL.User != nil {
				buf.WriteString("//")
			}

			if ui := inputURL.User; ui != nil {
				buf.WriteString(ui.String())
				buf.WriteByte('@')
			}

			if h := inputURL.Host; h != "" {
				buf.WriteString(inputURL.Host)
			}
		}
	}

	return buf.String()
}
