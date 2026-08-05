package gitlab_test

import (
	"context"
	"testing"
	"time"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm/gitlab"
	"github.com/stretchr/testify/require"
)

// activityContext returns a context carrying the configuration the activity
// helpers read to decide whose activity should be ignored.
func activityContext(t *testing.T, ignore config.IgnoreActivityFrom) context.Context {
	t.Helper()

	return config.WithConfig(t.Context(), &config.Config{IgnoreActivityFrom: ignore})
}

func TestHasLabel(t *testing.T) {
	t.Parallel()

	mr := gitlab.ContextMergeRequest{
		Labels: []gitlab.ContextLabel{
			{Title: "bug"},
			{Title: "type/ci"},
		},
	}

	tests := []struct {
		name  string
		label string
		want  bool
	}{
		{name: "first label matches", label: "bug", want: true},
		{name: "later label matches", label: "type/ci", want: true},
		{name: "unknown label", label: "feature", want: false},
		{name: "matching is exact, not a prefix", label: "type", want: false},
		{name: "matching is case sensitive", label: "Bug", want: false},
		{name: "empty string never matches", label: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, mr.HasLabel(t.Context(), tt.label))
			require.Equal(t, !tt.want, mr.HasNoLabel(t.Context(), tt.label), "has_no_label must be the inverse of has_label")
		})
	}
}

func TestHasLabel_noLabels(t *testing.T) {
	t.Parallel()

	mr := gitlab.ContextMergeRequest{}

	require.False(t, mr.HasLabel(t.Context(), "bug"))
	require.True(t, mr.HasNoLabel(t.Context(), "bug"))
}

func TestStateIs(t *testing.T) {
	t.Parallel()

	mr := gitlab.ContextMergeRequest{State: "opened"}

	tests := []struct {
		name  string
		anyOf []string
		want  bool
	}{
		{name: "single matching state", anyOf: []string{"opened"}, want: true},
		{name: "single non-matching state", anyOf: []string{"closed"}, want: false},
		{name: "matches one of several", anyOf: []string{"closed", "merged", "opened"}, want: true},
		{name: "matches none of several", anyOf: []string{"closed", "merged"}, want: false},
		{name: "no states given", anyOf: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, mr.StateIs(t.Context(), tt.anyOf...))
		})
	}
}

func TestStateIsNot(t *testing.T) {
	t.Parallel()

	mr := gitlab.ContextMergeRequest{State: "opened"}

	require.False(t, mr.StateIsNot(t.Context(), "opened"))
	require.True(t, mr.StateIsNot(t.Context(), "closed"))
	require.False(t, mr.StateIsNot(t.Context(), "closed", "opened"))
	require.True(t, mr.StateIsNot(t.Context(), "closed", "merged"))

	// With nothing to compare against the Merge Request is "not" every state
	require.True(t, mr.StateIsNot(t.Context()))
}

// A typo in a state name is a configuration mistake, so it panics loudly rather
// than silently evaluating to false forever.
func TestStateIs_panicsOnUnknownState(t *testing.T) {
	t.Parallel()

	mr := gitlab.ContextMergeRequest{State: "opened"}

	require.PanicsWithError(t, `unknown state value: "nonsense"`, func() {
		mr.StateIs(t.Context(), "nonsense")
	})

	require.PanicsWithError(t, `unknown state value: "nonsense"`, func() {
		mr.StateIsNot(t.Context(), "nonsense")
	})
}

func TestStateIs_acceptsEveryDocumentedState(t *testing.T) {
	t.Parallel()

	for _, state := range gitlab.AllMergeRequestState {
		t.Run(string(state), func(t *testing.T) {
			t.Parallel()

			mr := gitlab.ContextMergeRequest{State: string(state)}
			require.True(t, mr.StateIs(t.Context(), string(state)))
		})
	}
}

func TestModifiedFiles(t *testing.T) {
	t.Parallel()

	mr := gitlab.ContextMergeRequest{
		DiffStats: []gitlab.ContextDiffStat{
			{Path: "main.go"},
			{Path: "pkg/config/config.go"},
			{Path: "pkg/config/config_test.go"},
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
		{
			name:     "extension glob matches at any depth",
			patterns: []string{"*_test.go"},
			want:     []string{"pkg/config/config_test.go"},
		},
		{
			name:     "several patterns are unioned",
			patterns: []string{"docs/", "main.go"},
			want:     []string{"docs/index.md", "main.go"},
		},
		{name: "no match", patterns: []string{"nope/"}, want: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := mr.ModifiedFilesList(tt.patterns...)
			require.Equal(t, tt.want, got)
			require.Equal(t, len(tt.want) > 0, mr.ModifiedFiles(tt.patterns...))
		})
	}
}

func TestModifiedFiles_noDiff(t *testing.T) {
	t.Parallel()

	mr := gitlab.ContextMergeRequest{}

	require.Empty(t, mr.ModifiedFilesList("*"))
	require.False(t, mr.ModifiedFiles("*"))
}

func TestHasAnyActivityWithin(t *testing.T) {
	t.Parallel()

	var (
		now    = time.Now()
		recent = now.Add(-1 * time.Hour)
		old    = now.Add(-30 * 24 * time.Hour)
	)

	tests := []struct {
		name string
		mr   gitlab.ContextMergeRequest
		want bool
	}{
		{
			name: "merge request itself was updated recently",
			mr:   gitlab.ContextMergeRequest{UpdatedAt: recent},
			want: true,
		},
		{
			name: "nothing happened at all",
			mr:   gitlab.ContextMergeRequest{UpdatedAt: old},
			want: false,
		},
		{
			name: "a recent commit counts as activity",
			mr: gitlab.ContextMergeRequest{
				UpdatedAt:  old,
				LastCommit: &gitlab.ContextCommit{CommittedDate: &recent},
			},
			want: true,
		},
		{
			name: "an old commit does not",
			mr: gitlab.ContextMergeRequest{
				UpdatedAt:  old,
				LastCommit: &gitlab.ContextCommit{CommittedDate: &old},
			},
			want: false,
		},
		{
			name: "a recent note counts as activity",
			mr: gitlab.ContextMergeRequest{
				UpdatedAt: old,
				Notes: []gitlab.ContextNote{
					{UpdatedAt: recent, Author: &gitlab.ContextUser{Username: "someone"}},
				},
			},
			want: true,
		},
		{
			name: "a bot note still counts as *any* activity",
			mr: gitlab.ContextMergeRequest{
				UpdatedAt: old,
				Notes: []gitlab.ContextNote{
					{UpdatedAt: recent, Author: &gitlab.ContextUser{Username: "bot", Bot: true}},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := activityContext(t, config.IgnoreActivityFrom{})

			require.Equal(t, tt.want, tt.mr.HasAnyActivityWithin(ctx, "7d"))
			require.Equal(t, tt.want, tt.mr.HasActivityWithin(ctx, "7d"), "has_activity_within is an alias")
			require.Equal(t, !tt.want, tt.mr.HasNoActivityWithin(ctx, "7d"), "has_no_activity_within is the inverse")
		})
	}
}

// ignore_activity_from is applied to notes, so an ignored author's comment must
// not keep a Merge Request looking alive.
func TestHasAnyActivityWithin_respectsIgnoreActivityFrom(t *testing.T) {
	t.Parallel()

	var (
		now    = time.Now()
		recent = now.Add(-1 * time.Hour)
		old    = now.Add(-30 * 24 * time.Hour)
	)

	mr := gitlab.ContextMergeRequest{
		UpdatedAt: old,
		Notes: []gitlab.ContextNote{
			{UpdatedAt: recent, Author: &gitlab.ContextUser{Username: "renovate"}},
		},
	}

	require.True(t, mr.HasAnyActivityWithin(activityContext(t, config.IgnoreActivityFrom{}), "7d"))

	ignored := activityContext(t, config.IgnoreActivityFrom{Usernames: []string{"renovate"}})
	require.False(t, mr.HasAnyActivityWithin(ignored, "7d"))
}

func TestHasUserActivityWithin(t *testing.T) {
	t.Parallel()

	var (
		now    = time.Now()
		recent = now.Add(-1 * time.Hour)
		old    = now.Add(-30 * 24 * time.Hour)
	)

	tests := []struct {
		name string
		mr   gitlab.ContextMergeRequest
		want bool
	}{
		{
			name: "a recent note from a human",
			mr: gitlab.ContextMergeRequest{
				UpdatedAt: old,
				Notes: []gitlab.ContextNote{
					{UpdatedAt: recent, Author: &gitlab.ContextUser{Username: "human"}},
				},
			},
			want: true,
		},
		{
			name: "bot notes are not user activity",
			mr: gitlab.ContextMergeRequest{
				UpdatedAt: old,
				Notes: []gitlab.ContextNote{
					{UpdatedAt: recent, Author: &gitlab.ContextUser{Username: "bot", Bot: true}},
				},
			},
			want: false,
		},
		{
			name: "scm-engine's own notes are not user activity",
			mr: gitlab.ContextMergeRequest{
				UpdatedAt:   old,
				CurrentUser: &gitlab.ContextUser{Username: "scm-engine"},
				Notes: []gitlab.ContextNote{
					{UpdatedAt: recent, Author: &gitlab.ContextUser{Username: "scm-engine"}},
				},
			},
			want: false,
		},
		{
			// Unlike has_any_activity_within, the Merge Request timestamp is
			// deliberately ignored because it cannot be attributed to a user.
			name: "a recently updated merge request alone is not user activity",
			mr:   gitlab.ContextMergeRequest{UpdatedAt: recent},
			want: false,
		},
		{
			name: "a recent commit is user activity",
			mr: gitlab.ContextMergeRequest{
				UpdatedAt:  old,
				LastCommit: &gitlab.ContextCommit{CommittedDate: &recent},
			},
			want: true,
		},
		{
			name: "an old note is not",
			mr: gitlab.ContextMergeRequest{
				UpdatedAt: old,
				Notes: []gitlab.ContextNote{
					{UpdatedAt: old, Author: &gitlab.ContextUser{Username: "human"}},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := activityContext(t, config.IgnoreActivityFrom{})

			require.Equal(t, tt.want, tt.mr.HasUserActivityWithin(ctx, "7d"))
			require.Equal(t, !tt.want, tt.mr.HasNoUserActivityWithin(ctx, "7d"))
		})
	}
}

// The activity helpers accept the same duration strings as the duration()
// script function, including the extended day and week units.
func TestActivityHelpers_acceptDurationForms(t *testing.T) {
	t.Parallel()

	mr := gitlab.ContextMergeRequest{UpdatedAt: time.Now().Add(-36 * time.Hour)}
	ctx := activityContext(t, config.IgnoreActivityFrom{})

	require.True(t, mr.HasAnyActivityWithin(ctx, "7d"))
	require.True(t, mr.HasAnyActivityWithin(ctx, "1w"))
	require.True(t, mr.HasAnyActivityWithin(ctx, 48*time.Hour))
	require.False(t, mr.HasAnyActivityWithin(ctx, "1h"))
	require.False(t, mr.HasAnyActivityWithin(ctx, 1*time.Hour))
}
