package github_test

import (
	"testing"

	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/scm/github"
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/stretchr/testify/require"
)

func TestContextPullRequest_IsApproved(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		decision github.PullRequestReviewDecision
		want     bool
	}{
		{name: "approved", decision: github.PullRequestReviewDecisionApproved, want: true},
		{name: "changes requested", decision: github.PullRequestReviewDecisionChangesRequested, want: false},
		{name: "review required", decision: github.PullRequestReviewDecisionReviewRequired, want: false},
		{name: "no decision yet", decision: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pullRequest := github.ContextPullRequest{ReviewDecision: tt.decision}
			require.Equal(t, tt.want, pullRequest.IsApproved())
		})
	}
}

func TestContextPullRequest_StateIs(t *testing.T) {
	t.Parallel()

	pullRequest := github.ContextPullRequest{State: github.PullRequestStateOpen}

	require.True(t, pullRequest.StateIs("OPEN"))
	require.False(t, pullRequest.StateIs("CLOSED"))
	require.True(t, pullRequest.StateIs("CLOSED", "MERGED", "OPEN"))
	require.False(t, pullRequest.StateIs("CLOSED", "MERGED"))
	require.False(t, pullRequest.StateIs(), "with no states to compare against nothing can match")
}

func TestContextPullRequest_StateIs_panicsOnUnknownState(t *testing.T) {
	t.Parallel()

	pullRequest := github.ContextPullRequest{State: github.PullRequestStateOpen}

	require.PanicsWithError(t, `unknown state value: "open"`, func() {
		// GitHub states are upper case, so the lower case spelling is a typo
		pullRequest.StateIs("open")
	})
}

func TestContextPullRequest_StateIs_acceptsEveryGeneratedState(t *testing.T) {
	t.Parallel()

	for _, state := range github.AllPullRequestState {
		t.Run(string(state), func(t *testing.T) {
			t.Parallel()

			pullRequest := github.ContextPullRequest{State: state}
			require.True(t, pullRequest.StateIs(string(state)))
		})
	}
}

func TestContextPullRequest_HasLabel(t *testing.T) {
	t.Parallel()

	pullRequest := github.ContextPullRequest{
		Labels: []github.ContextLabel{{Name: "bug"}, {Name: "type/ci"}},
	}

	tests := []struct {
		name  string
		label string
		want  bool
	}{
		{name: "first label", label: "bug", want: true},
		{name: "later label", label: "type/ci", want: true},
		{name: "unknown label", label: "feature", want: false},
		{name: "prefix does not match", label: "type", want: false},
		{name: "case sensitive", label: "Bug", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, pullRequest.HasLabel(tt.label))
			require.Equal(t, !tt.want, pullRequest.HasNoLabel(tt.label))
		})
	}
}

func TestContextPullRequest_ModifiedFiles(t *testing.T) {
	t.Parallel()

	pullRequest := github.ContextPullRequest{
		Files: []github.PullRequestChangedFile{
			{Path: "main.go"},
			{Path: "pkg/config/config.go"},
			{Path: "docs/index.md"},
		},
	}

	tests := []struct {
		name     string
		patterns []string
		want     []string
	}{
		{name: "directory prefix", patterns: []string{"docs/"}, want: []string{"docs/index.md"}},
		{name: "exact file", patterns: []string{"main.go"}, want: []string{"main.go"}},
		{name: "extension glob", patterns: []string{"*.md"}, want: []string{"docs/index.md"}},
		{name: "no match", patterns: []string{"nope/"}, want: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, pullRequest.ModifiedFilesList(tt.patterns...))
			require.Equal(t, len(tt.want) > 0, pullRequest.ModifiedFiles(tt.patterns...))
		})
	}
}

func TestContext_IsValid(t *testing.T) {
	t.Parallel()

	require.True(t, (&github.Context{}).IsValid())

	var nilContext *github.Context
	require.False(t, nilContext.IsValid())
}

func TestContext_GetDescription(t *testing.T) {
	t.Parallel()

	evalContext := &github.Context{
		PullRequest: &github.ContextPullRequest{Body: "the body"},
	}

	require.Equal(t, "the body", evalContext.GetDescription())
}

func TestContext_GetLabels(t *testing.T) {
	t.Parallel()

	empty := &github.Context{PullRequest: &github.ContextPullRequest{}}
	require.Empty(t, empty.GetLabels())

	withLabels := &github.Context{
		PullRequest: &github.ContextPullRequest{
			Labels: []github.ContextLabel{{Name: "bug"}, {Name: "type/ci"}},
		},
	}
	require.Equal(t, []string{"bug", "type/ci"}, withLabels.GetLabels())
}

func TestContext_SetWebhookEvent(t *testing.T) {
	t.Parallel()

	evalContext := &github.Context{}
	evalContext.SetWebhookEvent(map[string]string{"action": "opened"})

	require.Equal(t, map[string]string{"action": "opened"}, evalContext.WebhookEvent)
}

func TestContext_SetContext(t *testing.T) {
	t.Parallel()

	evalContext := &github.Context{}
	evalContext.SetContext(state.WithProjectID(t.Context(), "jippi/scm-engine"))

	require.Equal(t, "jippi/scm-engine", state.ProjectID(evalContext.Context))
}

// GitHub always reads the configuration from the Pull Request branch, unlike
// GitLab which falls back to HEAD when the branch has diverged.
func TestContext_CanUseConfigurationFileFromChangeRequest(t *testing.T) {
	t.Parallel()

	require.True(t, (&github.Context{}).CanUseConfigurationFileFromChangeRequest(t.Context()))
}

func TestContext_AllowPipelineFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		diff []github.PullRequestChangedFile
		want bool
	}{
		{name: "no files changed", diff: nil, want: false},
		{name: "config file changed", diff: []github.PullRequestChangedFile{{Path: ".scm-engine.yml"}}, want: true},
		{name: "unrelated file changed", diff: []github.PullRequestChangedFile{{Path: "main.go"}}, want: false},
		{
			name: "config file among others",
			diff: []github.PullRequestChangedFile{{Path: "main.go"}, {Path: ".scm-engine.yml"}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := state.WithConfigFilePath(t.Context(), ".scm-engine.yml")
			evalContext := &github.Context{PullRequest: &github.ContextPullRequest{Files: tt.diff}}

			require.Equal(t, tt.want, evalContext.AllowPipelineFailure(ctx))
		})
	}
}

func TestContext_ActionGroupTracking(t *testing.T) {
	t.Parallel()

	evalContext := &github.Context{ActionGroups: map[string]any{}}

	require.False(t, evalContext.HasExecutedActionGroup("stale"))

	evalContext.TrackActionGroupExecution("stale")
	require.True(t, evalContext.HasExecutedActionGroup("stale"))
	require.False(t, evalContext.HasExecutedActionGroup("other"))
}

// Actions without a group must not be tracked. Without this guard the first
// ungrouped action claims the empty group and every later ungrouped action is
// skipped as a duplicate, so only one of them would ever run.
func TestContext_ActionGroupTracking_ignoresUngrouped(t *testing.T) {
	t.Parallel()

	evalContext := &github.Context{ActionGroups: map[string]any{}}

	evalContext.TrackActionGroupExecution("")

	require.False(t, evalContext.HasExecutedActionGroup(""))
	require.Empty(t, evalContext.ActionGroups)

	// A second ungrouped action must still be allowed to run
	evalContext.TrackActionGroupExecution("")
	require.False(t, evalContext.HasExecutedActionGroup(""))
}

// These are documented as unimplemented on GitHub; assert the contract so a
// future implementation has to update the tests deliberately.
func TestContext_unimplementedAccessors(t *testing.T) {
	t.Parallel()

	evalContext := &github.Context{}

	require.Empty(t, evalContext.GetCodeOwners())
	require.Empty(t, evalContext.GetReviewers())
	require.Equal(t, scm.Actor{}, evalContext.GetAuthor())
}
