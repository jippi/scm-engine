package gitlab_test

import (
	"testing"

	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/scm/gitlab"
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/stretchr/testify/require"
)

func TestContext_IsValid(t *testing.T) {
	t.Parallel()

	require.True(t, (&gitlab.Context{}).IsValid())

	var nilContext *gitlab.Context
	require.False(t, nilContext.IsValid())
}

func TestContext_GetDescription(t *testing.T) {
	t.Parallel()

	empty := &gitlab.Context{MergeRequest: &gitlab.ContextMergeRequest{}}
	require.Empty(t, empty.GetDescription(), "a nil description reads as an empty string")

	withBody := &gitlab.Context{
		MergeRequest: &gitlab.ContextMergeRequest{Description: scm.Ptr("hello world")},
	}
	require.Equal(t, "hello world", withBody.GetDescription())
}

func TestContext_SetWebhookEvent(t *testing.T) {
	t.Parallel()

	evalContext := &gitlab.Context{}
	require.Nil(t, evalContext.WebhookEvent)

	evalContext.SetWebhookEvent(map[string]string{"object_kind": "merge_request"})
	require.Equal(t, map[string]string{"object_kind": "merge_request"}, evalContext.WebhookEvent)
}

func TestContext_SetContext(t *testing.T) {
	t.Parallel()

	evalContext := &gitlab.Context{}
	ctx := state.WithProjectID(t.Context(), "jippi/scm-engine")

	evalContext.SetContext(ctx)

	require.Equal(t, "jippi/scm-engine", state.ProjectID(evalContext.Context))
}

func TestContext_GetLabels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mr   *gitlab.ContextMergeRequest
		want []string
	}{
		{
			name: "no labels",
			mr:   &gitlab.ContextMergeRequest{},
			want: []string{},
		},
		{
			name: "labels are returned in order",
			mr: &gitlab.ContextMergeRequest{
				Labels: []gitlab.ContextLabel{{Title: "bug"}, {Title: "type/ci"}},
			},
			want: []string{"bug", "type/ci"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			evalContext := &gitlab.Context{MergeRequest: tt.mr}
			require.Equal(t, tt.want, evalContext.GetLabels())
		})
	}
}

func TestContext_GetAuthor(t *testing.T) {
	t.Parallel()

	evalContext := &gitlab.Context{
		MergeRequest: &gitlab.ContextMergeRequest{
			Author: &gitlab.ContextUser{
				ID:          "gid://gitlab/User/42",
				Username:    "jippi",
				Bot:         false,
				PublicEmail: scm.Ptr("jippi@example.com"),
			},
		},
	}

	author := evalContext.GetAuthor()
	require.Equal(t, "gid://gitlab/User/42", author.ID)
	require.Equal(t, "jippi", author.Username)
	require.False(t, author.IsBot)
	require.Equal(t, scm.Ptr("jippi@example.com"), author.Email)
}

func TestContext_GetReviewers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mr   *gitlab.ContextMergeRequest
		want scm.Actors
	}{
		{
			name: "no reviewers",
			mr:   &gitlab.ContextMergeRequest{},
			want: scm.Actors{},
		},
		{
			name: "reviewers are converted to actors",
			mr: &gitlab.ContextMergeRequest{
				Reviewers: []*gitlab.ContextUser{
					{ID: "1", Username: "one"},
					{ID: "2", Username: "two", Bot: true},
				},
			},
			want: scm.Actors{
				{ID: "1", Username: "one"},
				{ID: "2", Username: "two", IsBot: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			evalContext := &gitlab.Context{MergeRequest: tt.mr}
			require.Equal(t, tt.want, evalContext.GetReviewers())
		})
	}
}

func TestContext_ActionGroupTracking(t *testing.T) {
	t.Parallel()

	evalContext := &gitlab.Context{ActionGroups: map[string]any{}}

	require.False(t, evalContext.HasExecutedActionGroup("stale"))

	evalContext.TrackActionGroupExecution("stale")
	require.True(t, evalContext.HasExecutedActionGroup("stale"))
	require.False(t, evalContext.HasExecutedActionGroup("other"))

	// Tracking the same group twice is harmless
	evalContext.TrackActionGroupExecution("stale")
	require.True(t, evalContext.HasExecutedActionGroup("stale"))
}

// Actions without a group must never be recorded, otherwise a single ungrouped
// action would block every other ungrouped action from running.
func TestContext_ActionGroupTracking_ignoresUngrouped(t *testing.T) {
	t.Parallel()

	evalContext := &gitlab.Context{ActionGroups: map[string]any{}}

	evalContext.TrackActionGroupExecution("")

	require.False(t, evalContext.HasExecutedActionGroup(""))
	require.Empty(t, evalContext.ActionGroups)
}

func TestContext_CanUseConfigurationFileFromChangeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mr   *gitlab.ContextMergeRequest
		want bool
	}{
		{
			name: "branch is up to date with HEAD",
			mr:   &gitlab.ContextMergeRequest{},
			want: true,
		},
		{
			name: "branch has diverged from HEAD",
			mr:   &gitlab.ContextMergeRequest{DivergedFromTargetBranch: true},
			want: false,
		},
		{
			name: "branch needs a rebase",
			mr:   &gitlab.ContextMergeRequest{ShouldBeRebased: true},
			want: false,
		},
		{
			name: "both diverged and needing a rebase",
			mr:   &gitlab.ContextMergeRequest{DivergedFromTargetBranch: true, ShouldBeRebased: true},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			evalContext := &gitlab.Context{MergeRequest: tt.mr}
			require.Equal(t, tt.want, evalContext.CanUseConfigurationFileFromChangeRequest(t.Context()))
		})
	}
}

// The pipeline is allowed to fail only when the Merge Request changes the
// scm-engine configuration itself, so linting a config change surfaces as a
// failure while unrelated changes fail open.
func TestContext_AllowPipelineFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		diff     []gitlab.ContextDiffStat
		wantFail bool
	}{
		{
			name:     "no files changed",
			diff:     nil,
			wantFail: false,
		},
		{
			name:     "the configuration file was changed",
			diff:     []gitlab.ContextDiffStat{{Path: ".scm-engine.yml"}},
			wantFail: true,
		},
		{
			name: "an unrelated file was changed",
			diff: []gitlab.ContextDiffStat{{Path: "main.go"}},

			wantFail: false,
		},
		{
			name: "the configuration file alongside other files",
			diff: []gitlab.ContextDiffStat{{Path: "main.go"}, {Path: ".scm-engine.yml"}},

			wantFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := state.WithConfigFilePath(t.Context(), ".scm-engine.yml")
			evalContext := &gitlab.Context{MergeRequest: &gitlab.ContextMergeRequest{DiffStats: tt.diff}}

			require.Equal(t, tt.wantFail, evalContext.AllowPipelineFailure(ctx))
		})
	}
}

func TestContextUser_ToActor(t *testing.T) {
	t.Parallel()

	user := gitlab.ContextUser{
		ID:          "gid://gitlab/User/7",
		Username:    "someone",
		Bot:         true,
		PublicEmail: scm.Ptr("someone@example.com"),
	}

	require.Equal(t, scm.Actor{
		ID:       "gid://gitlab/User/7",
		Username: "someone",
		IsBot:    true,
		Email:    scm.Ptr("someone@example.com"),
	}, user.ToActor())
}

func TestContextUser_ToActor_withoutEmail(t *testing.T) {
	t.Parallel()

	actor := gitlab.ContextUser{ID: "1", Username: "someone"}.ToActor()

	require.Nil(t, actor.Email)
	require.False(t, actor.IsBot)
}
