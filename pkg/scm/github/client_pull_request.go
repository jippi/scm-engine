package github

import (
	"context"
	"io"
	"net/http"

	go_github "github.com/google/go-github/v62/github"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
)

var _ scm.MergeRequestClient = (*MergeRequestClient)(nil)

type MergeRequestClient struct {
	client *Client
}

func NewMergeRequestClient(client *Client) *MergeRequestClient {
	return &MergeRequestClient{client: client}
}

func (client *MergeRequestClient) Update(ctx context.Context, opt *scm.UpdateMergeRequestOptions) (*scm.Response, error) {
	owner, repo := ownerAndRepo(ctx)

	// Add labels
	if _, resp, err := client.client.wrapped.Issues.AddLabelsToIssue(ctx, owner, repo, state.MergeRequestIDInt(ctx), *opt.AddLabels); err != nil {
		return convertResponse(resp), err
	}

	// Remove labels
	if opt.RemoveLabels != nil && len(*opt.RemoveLabels) > 0 {
		for _, label := range *opt.RemoveLabels {
			if resp, err := client.client.wrapped.Issues.RemoveLabelForIssue(ctx, owner, repo, state.MergeRequestIDInt(ctx), label); err != nil {
				switch resp.StatusCode {
				case http.StatusNotFound:
					// Ignore

				default:
					return convertResponse(resp), err
				}
			}
		}
	}

	// Update MR
	updatePullRequest := &go_github.PullRequest{
		Locked: opt.DiscussionLocked,
	}

	_, resp, err := client.client.wrapped.PullRequests.Edit(ctx, owner, repo, state.MergeRequestIDInt(ctx), updatePullRequest)

	return convertResponse(resp), err
}

func (client *MergeRequestClient) GetRemoteConfig(ctx context.Context, filename, ref string) (io.Reader, error) {
	return nil, nil
}

func (client *MergeRequestClient) List(ctx context.Context, options *scm.ListMergeRequestsOptions) ([]scm.ListMergeRequest, error) {
	return nil, nil
}
