package github

import (
	"context"
	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	"golang.org/x/oauth2"
)

var _ scm.EvalContext = (*Context)(nil)

func NewContext(ctx context.Context, _, token string) (*Context, error) {
	httpClient := oauth2.NewClient(
		ctx,
		oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: token,
			},
		),
	)

	owner, repo := ownerAndRepo(ctx)

	client := graphql.NewClient("https://api.github.com/graphql", httpClient)

	var (
		evalContext *Context
		variables   = map[string]any{
			"owner": owner,
			"repo":  repo,
			"pr":    state.MergeRequestIDInt(ctx),
		}
	)

	if err := client.Query(ctx, &evalContext, variables); err != nil {
		return nil, err
	}

	// Initialize null-able types
	evalContext.ActionGroups = make(map[string]any)

	// move PullRequest to root context
	evalContext.PullRequest = evalContext.Repository.PullRequest
	evalContext.Repository.PullRequest = nil

	// Move 'files' to MR context without nesting
	evalContext.PullRequest.Files = evalContext.PullRequest.ResponseFiles.Nodes
	evalContext.PullRequest.ResponseFiles = nil

	// Move 'labels' to MR context without nesting
	evalContext.PullRequest.Labels = evalContext.PullRequest.ResponseLabels.Nodes
	evalContext.PullRequest.ResponseLabels = nil

	if len(evalContext.PullRequest.ResponseFirstCommits.Nodes) > 0 {
		evalContext.PullRequest.FirstCommit = evalContext.PullRequest.ResponseFirstCommits.Nodes[0].Commit

		tmp := time.Since(evalContext.PullRequest.FirstCommit.CommittedDate)
		evalContext.PullRequest.TimeSinceFirstCommit = &tmp
	}

	evalContext.PullRequest.ResponseFirstCommits = nil

	if len(evalContext.PullRequest.ResponseLastCommits.Nodes) > 0 {
		evalContext.PullRequest.LastCommit = evalContext.PullRequest.ResponseLastCommits.Nodes[0].Commit

		tmp := time.Since(evalContext.PullRequest.LastCommit.CommittedDate)
		evalContext.PullRequest.TimeSinceLastCommit = &tmp
	}

	evalContext.PullRequest.ResponseLastCommits = nil

	if evalContext.PullRequest.FirstCommit != nil && evalContext.PullRequest.LastCommit != nil {
		tmp := evalContext.PullRequest.FirstCommit.CommittedDate.Sub(evalContext.PullRequest.LastCommit.CommittedDate).Round(time.Hour)
		evalContext.PullRequest.TimeBetweenFirstAndLastCommit = &tmp
	}

	return evalContext, nil
}

func (c *Context) IsValid() bool {
	return c != nil
}

func (c *Context) SetWebhookEvent(in any) {
	c.WebhookEvent = in
}

func (c *Context) SetContext(ctx context.Context) {
	c.Context = ctx
}

func (c *Context) GetDescription() string {
	return c.PullRequest.Body
}

func (c *Context) CanUseConfigurationFileFromChangeRequest(ctx context.Context) bool {
	return true
}

func (c *Context) TrackActionGroupExecution(name string) {
	c.ActionGroups[name] = true
}

func (c *Context) HasExecutedActionGroup(name string) bool {
	_, ok := c.ActionGroups[name]

	return ok
}

func (c *Context) AllowPipelineFailure(ctx context.Context) bool {
	return len(c.PullRequest.findModifiedFiles(state.ConfigFilePath(ctx))) == 1
}

func (c *Context) GetCodeOwners() []scm.Actor {
	return []scm.Actor{}
}
