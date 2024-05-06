package gitlab

import (
	"context"
	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/jippi/gitlab-labeller/pkg/scm"
	"github.com/jippi/gitlab-labeller/pkg/state"
	"golang.org/x/oauth2"
)

var _ scm.EvalContext = (*Context)(nil)

func NewContext(ctx context.Context, baseURL, token string) (*Context, error) {
	httpClient := oauth2.NewClient(
		ctx,
		oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: token,
			},
		),
	)

	client := graphql.NewClient(baseURL+"/api/graphql", httpClient)

	var (
		evalContext *Context
		variables   = map[string]any{
			"project_id": graphql.ID(state.ProjectIDFromContext(ctx)),
			"mr_id":      state.MergeRequestIDFromContext(ctx),
		}
	)

	if err := client.Query(ctx, &evalContext, variables); err != nil {
		return nil, err
	}

	if evalContext.Project.MergeRequest == nil {
		return nil, nil //nolint:nilnil
	}

	// Move project labels into a un-nested expr exposed field
	evalContext.Project.Labels = evalContext.Project.ResponseLabels.Nodes
	evalContext.Project.ResponseLabels.Nodes = nil

	// Move merge request labels into a un-nested expr exposed field
	evalContext.MergeRequest = evalContext.Project.MergeRequest
	evalContext.Project.MergeRequest = nil

	evalContext.MergeRequest.Labels = evalContext.MergeRequest.ResponseLabels.Nodes
	evalContext.MergeRequest.ResponseLabels = nil

	evalContext.Group = evalContext.Project.ResponseGroup
	evalContext.Project.ResponseGroup = nil

	if len(evalContext.MergeRequest.ResponseFirstCommits.Nodes) > 0 {
		evalContext.MergeRequest.FirstCommit = &evalContext.MergeRequest.ResponseFirstCommits.Nodes[0]

		tmp := time.Since(evalContext.MergeRequest.FirstCommit.CommittedDate)
		evalContext.MergeRequest.TimeSinceFirstCommit = &tmp
	}

	evalContext.MergeRequest.ResponseFirstCommits = nil

	if len(evalContext.MergeRequest.ResponseLastCommits.Nodes) > 0 {
		evalContext.MergeRequest.LastCommit = &evalContext.MergeRequest.ResponseLastCommits.Nodes[0]

		tmp := time.Since(evalContext.MergeRequest.LastCommit.CommittedDate)
		evalContext.MergeRequest.TimeSinceLastCommit = &tmp
	}

	evalContext.MergeRequest.ResponseLastCommits = nil

	if evalContext.MergeRequest.FirstCommit != nil && evalContext.MergeRequest.LastCommit != nil {
		tmp := evalContext.MergeRequest.FirstCommit.CommittedDate.Sub(evalContext.MergeRequest.LastCommit.CommittedDate).Round(time.Hour)
		evalContext.MergeRequest.TimeBetweenFirstAndLastCommit = &tmp
	}

	return evalContext, nil
}

func (c *Context) IsValid() bool {
	return c != nil
}
