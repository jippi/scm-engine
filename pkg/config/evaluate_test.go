package config_test

import (
	"testing"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/scm/gitlab"
	"github.com/jippi/scm-engine/pkg/types"
	"github.com/stretchr/testify/require"
)

// evalContext returns a GitLab evaluation context with the given labels, which
// is enough for scripts built on merge_request.has_label.
func evalContext(labels ...string) *gitlab.Context {
	contextLabels := make([]gitlab.ContextLabel, 0, len(labels))
	for _, label := range labels {
		contextLabels = append(contextLabels, gitlab.ContextLabel{Title: label})
	}

	return &gitlab.Context{
		MergeRequest: &gitlab.ContextMergeRequest{Labels: contextLabels},
	}
}

func TestLabel_Evaluate_conditional(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		script      string
		wantMatched bool
	}{
		{name: "a true script matches", script: `true`, wantMatched: true},
		{name: "a false script does not match", script: `false`, wantMatched: false},
		{
			name:        "a script reading the context",
			script:      `merge_request.has_label("bug")`,
			wantMatched: true,
		},
		{
			name:        "a script reading the context negatively",
			script:      `merge_request.has_label("missing")`,
			wantMatched: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			label := &config.Label{Name: "my-label", Script: tt.script, Strategy: config.ConditionalLabel}

			results, err := label.Evaluate(t.Context(), evalContext("bug"))
			require.NoError(t, err)
			require.Len(t, results, 1, "a conditional label always produces exactly one result")

			require.Equal(t, "my-label", results[0].Name)
			require.Equal(t, tt.wantMatched, results[0].Matched)
		})
	}
}

// A conditional label carries its presentation through to the result, which is
// what syncLabels later compares against the remote label.
func TestLabel_Evaluate_carriesLabelSettings(t *testing.T) {
	t.Parallel()

	label := &config.Label{
		Name:        "my-label",
		Script:      `true`,
		Strategy:    config.ConditionalLabel,
		Color:       "#FF0000",
		Description: "a description",
		Priority:    types.ValueFrom(3),
	}

	results, err := label.Evaluate(t.Context(), evalContext())
	require.NoError(t, err)
	require.Len(t, results, 1)

	require.Equal(t, "#FF0000", results[0].Color)
	require.Equal(t, "a description", results[0].Description)
	require.Equal(t, types.ValueFrom(3), results[0].Priority)
}

func TestLabel_Evaluate_generate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		script string
		want   []string
	}{
		{name: "a list of strings", script: `["one", "two"]`, want: []string{"one", "two"}},
		{name: "an empty list produces nothing", script: `[]`, want: nil},
		{
			// map() produces a []any rather than a []string, and both shapes have
			// to be accepted or every generated label config would break.
			name:   "a list of any produced by map",
			script: `["one", "two"] | map({ # })`,
			want:   []string{"one", "two"},
		},
		{
			name:   "a list run through uniq",
			script: `["b", "a", "b"] | uniq()`,
			want:   []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			label := &config.Label{Script: tt.script, Strategy: config.GenerateLabels}

			results, err := label.Evaluate(t.Context(), evalContext())
			require.NoError(t, err)

			names := make([]string, 0, len(results))

			for _, result := range results {
				require.True(t, result.Matched, "generated labels are always matched")

				names = append(names, result.Name)
			}

			if tt.want == nil {
				require.Empty(t, names)

				return
			}

			require.Equal(t, tt.want, names)
		})
	}
}

// The strategy and the script return type have to agree. The mismatch is caught
// while compiling the script, because the expected return type is baked into the
// compile options, so it surfaces during `scm-engine lint` rather than half way
// through an evaluation.
func TestLabel_Evaluate_strategyMismatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		label   *config.Label
		wantErr string
	}{
		{
			name:    "a boolean from a generate label",
			label:   &config.Label{Script: `true`, Strategy: config.GenerateLabels},
			wantErr: "expected slice, but got bool",
		},
		{
			name:    "a list from a conditional label",
			label:   &config.Label{Name: "x", Script: `["one"]`, Strategy: config.ConditionalLabel},
			wantErr: "expected bool, but got []interface {}",
		},
		{
			name:    "a list of non-strings",
			label:   &config.Label{Script: `[1, 2] | map({ # })`, Strategy: config.GenerateLabels},
			wantErr: "must return a list of strings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := tt.label.Evaluate(t.Context(), evalContext())
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestLabel_ShouldSkip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		skipIf string
		want   bool
	}{
		{name: "no skip_if never skips", skipIf: "", want: false},
		{name: "a true skip_if skips", skipIf: `true`, want: true},
		{name: "a false skip_if does not skip", skipIf: `false`, want: false},
		{name: "a skip_if reading the context", skipIf: `merge_request.has_label("wip")`, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			label := &config.Label{Name: "x", Script: `true`, Strategy: config.ConditionalLabel, SkipIf: tt.skipIf}

			skip, err := label.ShouldSkip(t.Context(), evalContext("wip"))
			require.NoError(t, err)
			require.Equal(t, tt.want, skip)
		})
	}
}

// A skipped label produces no result at all, which is different from producing
// a result that did not match: the label is left exactly as it is on the MR.
func TestLabel_Evaluate_skippedProducesNoResult(t *testing.T) {
	t.Parallel()

	label := &config.Label{
		Name:     "x",
		Script:   `true`,
		Strategy: config.ConditionalLabel,
		SkipIf:   `true`,
	}

	results, err := label.Evaluate(t.Context(), evalContext())
	require.NoError(t, err)
	require.Nil(t, results)
}

func TestLabels_Evaluate(t *testing.T) {
	t.Parallel()

	labels := config.Labels{
		{Name: "first", Script: `true`, Strategy: config.ConditionalLabel},
		{Name: "second", Script: `false`, Strategy: config.ConditionalLabel},
		{Name: "skipped", Script: `true`, Strategy: config.ConditionalLabel, SkipIf: `true`},
	}

	results, err := labels.Evaluate(t.Context(), evalContext())
	require.NoError(t, err)

	// The skipped label contributes nothing; the negative one still reports so
	// that it can be removed from the Merge Request.
	require.Len(t, results, 2)
	require.Equal(t, "first", results[0].Name)
	require.True(t, results[0].Matched)
	require.Equal(t, "second", results[1].Name)
	require.False(t, results[1].Matched)
}

func TestLabels_Evaluate_rejectsDuplicateNames(t *testing.T) {
	t.Parallel()

	labels := config.Labels{
		{Name: "dupe", Script: `true`, Strategy: config.ConditionalLabel},
		{Name: "dupe", Script: `true`, Strategy: config.ConditionalLabel},
	}

	_, err := labels.Evaluate(t.Context(), evalContext())
	require.ErrorContains(t, err, "was generated multiple times")
}

func TestLabels_Evaluate_rejectsEmptyName(t *testing.T) {
	t.Parallel()

	labels := config.Labels{
		{Script: `[""]`, Strategy: config.GenerateLabels},
	}

	_, err := labels.Evaluate(t.Context(), evalContext())
	require.ErrorContains(t, err, "empty name")
}

func TestLabels_Evaluate_reportsWhichLabelFailed(t *testing.T) {
	t.Parallel()

	labels := config.Labels{
		{Name: "fine", Script: `true`, Strategy: config.ConditionalLabel},
		{Name: "broken", Script: `["a"]`, Strategy: config.ConditionalLabel},
	}

	_, err := labels.Evaluate(t.Context(), evalContext())
	require.ErrorContains(t, err, "label: broken")
}

func TestAction_Evaluate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		script string
		want   bool
	}{
		{name: "true", script: `true`, want: true},
		{name: "false", script: `false`, want: false},
		{name: "reads the context", script: `merge_request.has_label("bug")`, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			action := &config.Action{Name: "x", If: tt.script}

			got, err := action.Evaluate(t.Context(), evalContext("bug"))
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestAction_Evaluate_rejectsNonBoolean(t *testing.T) {
	t.Parallel()

	action := &config.Action{Name: "x", If: `"a string"`}

	_, err := action.Evaluate(t.Context(), evalContext())
	require.Error(t, err)
}

func TestActions_Evaluate_keepsOnlyMatching(t *testing.T) {
	t.Parallel()

	actions := config.Actions{
		{Name: "yes", If: `true`},
		{Name: "no", If: `false`},
		{Name: "also-yes", If: `merge_request.has_label("bug")`},
	}

	results, err := actions.Evaluate(t.Context(), evalContext("bug"))
	require.NoError(t, err)
	require.Len(t, results, 2)
	require.Equal(t, "yes", results[0].Name)
	require.Equal(t, "also-yes", results[1].Name)
}

func TestActions_Evaluate_noneMatching(t *testing.T) {
	t.Parallel()

	results, err := config.Actions{{Name: "no", If: `false`}}.Evaluate(t.Context(), evalContext())
	require.NoError(t, err)
	require.Empty(t, results)
}

func TestActions_Evaluate_propagatesError(t *testing.T) {
	t.Parallel()

	_, err := config.Actions{{Name: "broken", If: `this is not valid expr(`}}.Evaluate(t.Context(), evalContext())
	require.Error(t, err)
}

// Config.Evaluate is the entry point ProcessMR calls, and it runs the labels
// before the actions so an action can react to a label that was just computed.
func TestConfig_Evaluate(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Labels:  config.Labels{{Name: "a-label", Script: `true`, Strategy: config.ConditionalLabel}},
		Actions: config.Actions{{Name: "an-action", If: `true`}},
	}

	labels, actions, err := cfg.Evaluate(t.Context(), evalContext())
	require.NoError(t, err)

	require.Len(t, labels, 1)
	require.Equal(t, "a-label", labels[0].Name)

	require.Len(t, actions, 1)
	require.Equal(t, "an-action", actions[0].Name)
}

func TestConfig_Evaluate_reportsLabelErrors(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Labels: config.Labels{{Name: "x", Script: `["a"]`, Strategy: config.ConditionalLabel}},
	}

	_, _, err := cfg.Evaluate(t.Context(), evalContext())
	require.ErrorContains(t, err, "evaluation failed")
}

func TestConfig_Lint(t *testing.T) {
	t.Parallel()

	t.Run("a valid config lints clean", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			Labels:  config.Labels{{Name: "x", Script: `true`, Strategy: config.ConditionalLabel}},
			Actions: config.Actions{{Name: "y", If: `true`}},
		}

		require.NoError(t, cfg.Lint(t.Context(), evalContext()))
	})

	t.Run("a broken label is reported by name", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			Labels: config.Labels{{Name: "broken", Script: `nope(`, Strategy: config.ConditionalLabel}},
		}

		require.ErrorContains(t, cfg.Lint(t.Context(), evalContext()), `Label "broken" failed validation`)
	})

	t.Run("a broken action is reported by name", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{Actions: config.Actions{{Name: "broken", If: `nope(`}}}

		require.ErrorContains(t, cfg.Lint(t.Context(), evalContext()), `Action "broken" failed validation`)
	})

	// Lint collects every problem so a user fixes them in one pass.
	t.Run("every problem is reported", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			Labels: config.Labels{
				{Name: "first", Script: `nope(`, Strategy: config.ConditionalLabel},
				{Name: "second", Script: `also_nope(`, Strategy: config.ConditionalLabel},
			},
		}

		err := cfg.Lint(t.Context(), evalContext())
		require.ErrorContains(t, err, "first")
		require.ErrorContains(t, err, "second")
	})
}

func TestConfigContext(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{DryRun: scm.Ptr(true)}
	require.Equal(t, cfg, config.FromContext(config.WithConfig(t.Context(), cfg)))

	global := &config.Config{DryRun: scm.Ptr(false)}
	require.Equal(t, global, config.GlobalConfigFromContext(config.WithGlobalConfig(t.Context(), global)))
}

// The two configurations live under separate keys so the repository config can
// be merged on top of the global one without either clobbering the other.
func TestConfigContext_areIndependent(t *testing.T) {
	t.Parallel()

	local := &config.Config{DryRun: scm.Ptr(true)}
	global := &config.Config{DryRun: scm.Ptr(false)}

	ctx := config.WithConfig(t.Context(), local)
	ctx = config.WithGlobalConfig(ctx, global)

	require.Equal(t, local, config.FromContext(ctx))
	require.Equal(t, global, config.GlobalConfigFromContext(ctx))
}
