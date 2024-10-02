package gitlab

import (
	"context"
	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	slogctx "github.com/veqryn/slog-context"
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
			"project_id": graphql.ID(state.ProjectID(ctx)),
			"mr_id":      state.MergeRequestID(ctx),
		}
	)

	if err := client.Query(ctx, &evalContext, variables); err != nil {
		return nil, err
	}

	if evalContext.Project == nil || evalContext.Project.MergeRequest == nil {
		return nil, nil //nolint:nilnil
	}

	// Initialize null-able types
	evalContext.ActionGroups = make(map[string]any)

	// Move project labels into a un-nested expr exposed field
	evalContext.Project.Labels = evalContext.Project.ResponseLabels.Nodes
	evalContext.Project.ResponseLabels.Nodes = nil

	// Move merge request labels into a un-nested expr exposed field
	evalContext.MergeRequest = evalContext.Project.MergeRequest
	evalContext.Project.MergeRequest = nil

	evalContext.MergeRequest.Assignees = evalContext.MergeRequest.CurrentAssignees.Nodes
	evalContext.MergeRequest.CurrentAssignees = nil

	evalContext.MergeRequest.Reviewers = evalContext.MergeRequest.CurrentReviewers.Nodes
	evalContext.MergeRequest.CurrentReviewers = nil

	// Copy "current user" into MR
	evalContext.MergeRequest.CurrentUser = evalContext.CurrentUser

	evalContext.MergeRequest.Labels = evalContext.MergeRequest.ResponseLabels.Nodes
	evalContext.MergeRequest.ResponseLabels = nil

	evalContext.Group = evalContext.Project.ResponseGroup
	evalContext.Project.ResponseGroup = nil

	evalContext.MergeRequest.Notes = evalContext.MergeRequest.ResponseNotes.Nodes
	evalContext.MergeRequest.ResponseNotes.Nodes = nil

	if len(evalContext.MergeRequest.ResponseFirstCommits.Nodes) > 0 {
		evalContext.MergeRequest.FirstCommit = &evalContext.MergeRequest.ResponseFirstCommits.Nodes[0]

		tmp := time.Since(*evalContext.MergeRequest.FirstCommit.CommittedDate)
		evalContext.MergeRequest.TimeSinceFirstCommit = &tmp
	}

	evalContext.MergeRequest.ResponseFirstCommits = nil

	if len(evalContext.MergeRequest.ResponseLastCommits.Nodes) > 0 {
		evalContext.MergeRequest.LastCommit = &evalContext.MergeRequest.ResponseLastCommits.Nodes[0]

		tmp := time.Since(*evalContext.MergeRequest.LastCommit.CommittedDate)
		evalContext.MergeRequest.TimeSinceLastCommit = &tmp
	}

	evalContext.MergeRequest.ResponseLastCommits = nil

	if evalContext.MergeRequest.FirstCommit != nil && evalContext.MergeRequest.LastCommit != nil {
		tmp := evalContext.MergeRequest.FirstCommit.CommittedDate.Sub(*evalContext.MergeRequest.LastCommit.CommittedDate).Round(time.Hour)
		evalContext.MergeRequest.TimeBetweenFirstAndLastCommit = &tmp
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
	if c.MergeRequest.Description == nil {
		return ""
	}

	return *c.MergeRequest.Description
}

func (c *Context) CanUseConfigurationFileFromChangeRequest(ctx context.Context) bool {
	// If the Merge Request has diverged from HEAD we can't trust the configuration
	if c.MergeRequest.DivergedFromTargetBranch {
		slogctx.Warn(ctx, "The Merge Request branch has diverged from HEAD; will use the scm-engine config from HEAD instead")

		return false
	}

	// If the Merge Request is not up to date with HEAD we can't trust the configuration
	if c.MergeRequest.ShouldBeRebased {
		slogctx.Warn(ctx, "The Merge Request branch is not up to date with HEAD; will use the scm-engine config from HEAD instead")

		return false
	}

	slogctx.Info(ctx, "The Merge Request branch is up to date with HEAD; will use the scm-engine config from the branch")

	return true
}

// AllowPipelineFailure controls if the CI pipeline are allowed to fail
//
// We allow the pipeline to fail with an error if the SCM-Engine configuration file
// is changed within the merge request, effectively allowing us to lint the configuration
// file when changing it but failing "open" in all other cases.
func (c *Context) AllowPipelineFailure(ctx context.Context) bool {
	return len(c.MergeRequest.findModifiedFiles(state.ConfigFilePath(ctx))) == 1
}

func (c *Context) TrackActionGroupExecution(group string) {
	// Ungrouped actions shouldn't be tracked
	if len(group) == 0 {
		return
	}

	c.ActionGroups[group] = true
}

func (c *Context) HasExecutedActionGroup(group string) bool {
	// Ungrouped actions shouldn't be tracked
	if len(group) == 0 {
		return false
	}

	_, ok := c.ActionGroups[group]

	return ok
}

// GetCodeOwners returns the eligible code owners for the merge request
//
// This is based on the elibible approvers in the rules of the merge request approval
// state api.
func (c *Context) GetCodeOwners() []string {
	ownerSet := make(map[string]bool)

	for _, rule := range c.MergeRequest.ApprovalState.Rules {
		// Multiple code owner paths could be matched when sections are used, so
		// flatten the list of eligible approvers
		if rule.Type.String() != "CODE_OWNER" {
			continue
		}

		if rule.EligibleApprovers != nil {
			for _, user := range rule.EligibleApprovers {
				ownerSet[user.Username] = true
			}
		}
	}

	owners := make([]string, 0, len(ownerSet))
	for owner := range ownerSet {
		owners = append(owners, owner)
	}

	return owners
}
