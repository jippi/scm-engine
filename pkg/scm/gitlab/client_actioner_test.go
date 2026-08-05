package gitlab_test

import (
	"testing"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/scm/gitlab"
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/stretchr/testify/require"
)

// applyStep runs a single action step against a fresh update struct. Only the
// actions that talk to the GitLab API need a real client, and those are covered
// separately through their dry-run path.
func applyStep(t *testing.T, step config.ActionStep, opts ...func(*evalContextMock)) (*scm.UpdateMergeRequestOptions, error) {
	t.Helper()

	evalContext := new(evalContextMock)
	for _, opt := range opts {
		opt(evalContext)
	}

	update := &scm.UpdateMergeRequestOptions{}

	ctx := state.WithDryRun(t.Context(), false)
	ctx = state.WithRandomSeed(ctx, 1)

	err := (&gitlab.Client{}).ApplyStep(ctx, evalContext, update, step)

	return update, err
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

// The state changing actions only populate the update struct; nothing is sent
// to GitLab until the whole action list has been applied.
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

func TestApplyStep_addLabel(t *testing.T) {
	t.Parallel()

	update, err := applyStep(t, config.ActionStep{"action": "add_label", "label": "bug"})
	require.NoError(t, err)
	require.NotNil(t, update.AddLabels)
	require.Equal(t, scm.LabelOptions{"bug"}, *update.AddLabels)
	require.Nil(t, update.RemoveLabels)
}

func TestApplyStep_removeLabel(t *testing.T) {
	t.Parallel()

	update, err := applyStep(t, config.ActionStep{"action": "remove_label", "label": "bug"})
	require.NoError(t, err)
	require.NotNil(t, update.RemoveLabels)
	require.Equal(t, scm.LabelOptions{"bug"}, *update.RemoveLabels)
	require.Nil(t, update.AddLabels)
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

// Several add_label steps accumulate rather than overwrite, which is what lets a
// single action group apply more than one label.
func TestApplyStep_labelsAccumulate(t *testing.T) {
	t.Parallel()

	evalContext := new(evalContextMock)
	update := &scm.UpdateMergeRequestOptions{}
	client := &gitlab.Client{}

	ctx := state.WithDryRun(t.Context(), false)

	require.NoError(t, client.ApplyStep(ctx, evalContext, update, config.ActionStep{"action": "add_label", "label": "one"}))
	require.NoError(t, client.ApplyStep(ctx, evalContext, update, config.ActionStep{"action": "add_label", "label": "two"}))
	require.NoError(t, client.ApplyStep(ctx, evalContext, update, config.ActionStep{"action": "remove_label", "label": "three"}))

	require.Equal(t, scm.LabelOptions{"one", "two"}, *update.AddLabels)
	require.Equal(t, scm.LabelOptions{"three"}, *update.RemoveLabels)
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

// In dry-run the API-backed actions must return before touching the GitLab
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

			ctx := state.WithDryRun(t.Context(), true)

			err := (&gitlab.Client{}).ApplyStep(ctx, new(evalContextMock), &scm.UpdateMergeRequestOptions{}, tt.step)
			require.NoError(t, err)
		})
	}
}

func TestApplyStep_updateDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		description string
		replace     any
		want        *string
		wantErr     string
	}{
		{
			name:        "replaces a placeholder with a script result",
			description: "Merge request ${{IID}} is ready",
			replace:     config.ActionStep{"${{IID}}": `"42"`},
			want:        scm.Ptr("Merge request 42 is ready"),
		},
		{
			name:        "replaces every occurrence of the key",
			description: "${{X}} and ${{X}}",
			replace:     config.ActionStep{"${{X}}": `"y"`},
			want:        scm.Ptr("y and y"),
		},
		{
			name:        "description is left alone when no key matches",
			description: "nothing to replace here",
			replace:     config.ActionStep{"${{MISSING}}": `"value"`},
			want:        nil,
		},
		{
			name:        "replace must be a dictionary",
			description: "anything",
			replace:     "not-a-dictionary",
			wantErr:     "step field 'replace' must be a dictionary",
		},
		{
			name:        "a script that does not compile is reported with its key",
			description: "${{BAD}}",
			replace:     config.ActionStep{"${{BAD}}": `this is not valid expr(`},
			wantErr:     "could not evaluate value for 'replace' key '${{BAD}}'",
		},
		{
			// The program is compiled with expr.AsKind(string), so a script with
			// the wrong return type is caught while compiling rather than by the
			// type switch on the result.
			name:        "a script returning a non-string is rejected",
			description: "${{NUM}}",
			replace:     config.ActionStep{"${{NUM}}": `1234`},
			wantErr:     "expected string, but got int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			update, err := applyStep(t,
				config.ActionStep{"action": "update_description", "replace": tt.replace},
				func(m *evalContextMock) { m.On("GetDescription").Return(tt.description) },
			)

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, update.Description)
		})
	}
}

func TestApplyStep_updateDescription_requiresReplace(t *testing.T) {
	t.Parallel()

	_, err := applyStep(t,
		config.ActionStep{"action": "update_description"},
		func(m *evalContextMock) { m.On("GetDescription").Return("body") },
	)
	require.Error(t, err)
}

// A later update_description step must build on the description produced by an
// earlier one instead of starting over from the Merge Request body.
func TestApplyStep_updateDescription_buildsOnPreviousStep(t *testing.T) {
	t.Parallel()

	evalContext := new(evalContextMock)
	evalContext.On("GetDescription").Return("${{ONE}} ${{TWO}}")

	update := &scm.UpdateMergeRequestOptions{}
	client := &gitlab.Client{}
	ctx := state.WithDryRun(t.Context(), false)

	require.NoError(t, client.ApplyStep(ctx, evalContext, update, config.ActionStep{
		"action":  "update_description",
		"replace": config.ActionStep{"${{ONE}}": `"first"`},
	}))

	require.NoError(t, client.ApplyStep(ctx, evalContext, update, config.ActionStep{
		"action":  "update_description",
		"replace": config.ActionStep{"${{TWO}}": `"second"`},
	}))

	require.Equal(t, scm.Ptr("first second"), update.Description)
}
