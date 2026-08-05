package github_test

import (
	"testing"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/scm/github"
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/stretchr/testify/require"
)

// applyStep runs one action step against a fresh update struct.
//
// ApplyStep resolves the owner/repo from the project ID before looking at the
// action, so the project must always be present on the context.
func applyStep(t *testing.T, step config.ActionStep) (*scm.UpdateMergeRequestOptions, error) {
	t.Helper()

	ctx := state.WithProjectID(t.Context(), "jippi/scm-engine")
	ctx = state.WithDryRun(ctx, false)

	update := &scm.UpdateMergeRequestOptions{}

	return update, (&github.Client{}).ApplyStep(ctx, nil, update, step)
}

func TestApplyStep_requiresAnAction(t *testing.T) {
	t.Parallel()

	_, err := applyStep(t, config.ActionStep{})
	require.ErrorContains(t, err, "Required 'step' key 'action' is missing")
}

func TestApplyStep_rejectsUnknownAction(t *testing.T) {
	t.Parallel()

	_, err := applyStep(t, config.ActionStep{"action": "no_such_action"})
	require.ErrorContains(t, err, `does not know how to apply action "no_such_action"`)
}

func TestApplyStep_stateActions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		action string
		assert func(*testing.T, *scm.UpdateMergeRequestOptions)
	}{
		{
			name:   "close",
			action: "close",
			assert: func(t *testing.T, u *scm.UpdateMergeRequestOptions) {
				t.Helper()
				require.Equal(t, scm.Ptr("close"), u.StateEvent)
			},
		},
		{
			name:   "reopen",
			action: "reopen",
			assert: func(t *testing.T, u *scm.UpdateMergeRequestOptions) {
				t.Helper()
				require.Equal(t, scm.Ptr("reopen"), u.StateEvent)
			},
		},
		{
			name:   "lock_discussion",
			action: "lock_discussion",
			assert: func(t *testing.T, u *scm.UpdateMergeRequestOptions) {
				t.Helper()
				require.Equal(t, scm.Ptr(true), u.DiscussionLocked)
			},
		},
		{
			name:   "unlock_discussion",
			action: "unlock_discussion",
			assert: func(t *testing.T, u *scm.UpdateMergeRequestOptions) {
				t.Helper()
				require.Equal(t, scm.Ptr(false), u.DiscussionLocked)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			update, err := applyStep(t, config.ActionStep{"action": tt.action})
			require.NoError(t, err)
			tt.assert(t, update)
		})
	}
}

func TestApplyStep_labels(t *testing.T) {
	t.Parallel()

	added, err := applyStep(t, config.ActionStep{"action": "add_label", "label": "bug"})
	require.NoError(t, err)
	require.Equal(t, scm.LabelOptions{"bug"}, *added.AddLabels)
	require.Nil(t, added.RemoveLabels)

	removed, err := applyStep(t, config.ActionStep{"action": "remove_label", "label": "bug"})
	require.NoError(t, err)
	require.Equal(t, scm.LabelOptions{"bug"}, *removed.RemoveLabels)
	require.Nil(t, removed.AddLabels)
}

func TestApplyStep_labelActionsRequireALabel(t *testing.T) {
	t.Parallel()

	for _, action := range []string{"add_label", "remove_label"} {
		t.Run(action, func(t *testing.T) {
			t.Parallel()

			_, err := applyStep(t, config.ActionStep{"action": action})
			require.ErrorContains(t, err, "Required 'step' key 'label' is missing")
		})
	}
}

func TestApplyStep_labelsAccumulate(t *testing.T) {
	t.Parallel()

	ctx := state.WithProjectID(t.Context(), "jippi/scm-engine")
	ctx = state.WithDryRun(ctx, false)

	client := &github.Client{}
	update := &scm.UpdateMergeRequestOptions{}

	require.NoError(t, client.ApplyStep(ctx, nil, update, config.ActionStep{"action": "add_label", "label": "one"}))
	require.NoError(t, client.ApplyStep(ctx, nil, update, config.ActionStep{"action": "add_label", "label": "two"}))

	require.Equal(t, scm.LabelOptions{"one", "two"}, *update.AddLabels)
}

func TestApplyStep_comment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		step    config.ActionStep
		wantErr string
	}{
		{
			name:    "message is required",
			step:    config.ActionStep{"action": "comment"},
			wantErr: "Required 'step' key 'message' is missing",
		},
		{
			name:    "message may not be empty",
			step:    config.ActionStep{"action": "comment", "message": ""},
			wantErr: "step field 'message' must not be an empty string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := applyStep(t, tt.step)
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

// In dry-run the API-backed actions must return before touching the GitHub
// client, which is nil here — a regression would panic rather than fail softly.
func TestApplyStep_dryRunSkipsAPICalls(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		step config.ActionStep
	}{
		{name: "approve", step: config.ActionStep{"action": "approve"}},
		{name: "unapprove", step: config.ActionStep{"action": "unapprove"}},
		{name: "comment", step: config.ActionStep{"action": "comment", "message": "hello"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := state.WithProjectID(t.Context(), "jippi/scm-engine")
			ctx = state.WithDryRun(ctx, true)

			require.NoError(t, (&github.Client{}).ApplyStep(ctx, nil, &scm.UpdateMergeRequestOptions{}, tt.step))
		})
	}
}

// update_description and assign_reviewers are implemented for GitLab only. They
// are documented without a provider caveat, so this records the actual GitHub
// behaviour: the action is rejected rather than silently ignored.
func TestApplyStep_gitLabOnlyActionsAreRejected(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		step config.ActionStep
	}{
		{
			name: "update_description",
			step: config.ActionStep{"action": "update_description", "replace": config.ActionStep{"a": `"b"`}},
		},
		{
			name: "assign_reviewers",
			step: config.ActionStep{"action": "assign_reviewers", "source": "codeowners"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := applyStep(t, tt.step)
			require.ErrorContains(t, err, "GitHub client does not know how to apply action")
		})
	}
}
