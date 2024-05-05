package gitlab

import (
	"context"
	"net/url"
	"strings"

	"github.com/jippi/gitlab-labeller/pkg/scm"
	go_gitlab "github.com/xanzy/go-gitlab"
)

// Ensure the GitLab client implements the [scm.Client]
var _ scm.Client = (*Client)(nil)

// Client is a wrapper around the GitLab specific implementation of [scm.Client] interface
type Client struct {
	wrapped *go_gitlab.Client
	token   string

	labels        *LabelClient
	mergeRequests *MergeRequestClient
}

// NewClient creates a new GitLab client
func NewClient(token, baseurl string) (*Client, error) {
	client, err := go_gitlab.NewClient(token, go_gitlab.WithBaseURL(baseurl))
	if err != nil {
		return nil, err
	}

	return &Client{wrapped: client, token: token}, nil
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
	res, err := NewContext(ctx, graphqlBaseURL(client.wrapped.BaseURL()), client.token)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func graphqlBaseURL(u *url.URL) string {
	var buf strings.Builder
	if u.Scheme != "" {
		buf.WriteString(u.Scheme)
		buf.WriteByte(':')
	}

	if u.Scheme != "" || u.Host != "" || u.User != nil {
		if u.OmitHost && u.Host == "" && u.User == nil {
			// omit empty host
		} else {
			if u.Host != "" || u.Path != "" || u.User != nil {
				buf.WriteString("//")
			}

			if ui := u.User; ui != nil {
				buf.WriteString(ui.String())
				buf.WriteByte('@')
			}

			if h := u.Host; h != "" {
				buf.WriteString(u.Host)
			}
		}
	}

	return buf.String()
}
