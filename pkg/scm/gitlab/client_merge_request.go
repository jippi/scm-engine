package gitlab

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jippi/gitlab-labeller/pkg/scm"
	"github.com/jippi/gitlab-labeller/pkg/state"
	go_gitlab "github.com/xanzy/go-gitlab"
)

var _ scm.MergeRequestClient = (*MergeRequestClient)(nil)

type MergeRequestClient struct {
	client *Client
}

func NewMergeRequestClient(client *Client) *MergeRequestClient {
	return &MergeRequestClient{client: client}
}

func (client *MergeRequestClient) Update(ctx context.Context, opt *scm.UpdateMergeRequestOptions) (*scm.Response, error) {
	project, err := ParseID(state.ProjectIDFromContext(ctx))
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("projects/%s/merge_requests/%s", go_gitlab.PathEscape(project), state.MergeRequestIDFromContext(ctx))

	options := []go_gitlab.RequestOptionFunc{
		go_gitlab.WithContext(ctx),
	}

	req, err := client.client.wrapped.NewRequest(http.MethodPut, endpoint, opt, options)
	if err != nil {
		return nil, err
	}

	m := new(go_gitlab.MergeRequest)

	resp, err := client.client.wrapped.Do(req, m)

	return convertResponse(resp), err
}
