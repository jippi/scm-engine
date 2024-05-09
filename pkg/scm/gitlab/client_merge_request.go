package gitlab

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/hasura/go-graphql-client"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	go_gitlab "github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

var _ scm.MergeRequestClient = (*MergeRequestClient)(nil)

type MergeRequestClient struct {
	client *Client
}

func NewMergeRequestClient(client *Client) *MergeRequestClient {
	return &MergeRequestClient{client: client}
}

func (client *MergeRequestClient) Update(ctx context.Context, opt *scm.UpdateMergeRequestOptions) (*scm.Response, error) {
	project, err := ParseID(state.ProjectID(ctx))
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("projects/%s/merge_requests/%s", go_gitlab.PathEscape(project), state.MergeRequestID(ctx))

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

func (client *MergeRequestClient) GetRemoteConfig(ctx context.Context, filename, ref string) (io.Reader, error) {
	project, err := ParseID(state.ProjectID(ctx))
	if err != nil {
		return nil, err
	}

	file, _, err := client.client.wrapped.RepositoryFiles.GetRawFile(project, filename, &go_gitlab.GetRawFileOptions{Ref: go_gitlab.Ptr(ref)})
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(file), nil
}

func (client *MergeRequestClient) List(ctx context.Context, options *scm.ListMergeRequestsOptions) ([]scm.ListMergeRequest, error) {
	httpClient := oauth2.NewClient(
		ctx,
		oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: client.client.token,
			},
		),
	)

	graphqlClient := graphql.NewClient(graphqlBaseURL(client.client.wrapped.BaseURL())+"/api/graphql", httpClient)

	var (
		result    *ListMergeRequestsQuery
		variables = map[string]any{
			"project_id": graphql.ID(state.ProjectID(ctx)),
			"state":      MergeRequestState(options.State),
			"first":      options.First,
		}
	)

	if err := graphqlClient.Query(ctx, &result, variables); err != nil {
		return nil, err
	}

	hits := []scm.ListMergeRequest{}
	for _, x := range result.Project.MergeRequests.Nodes {
		hits = append(hits, scm.ListMergeRequest{ID: x.ID})
	}

	return hits, nil
}
