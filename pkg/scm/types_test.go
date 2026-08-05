package scm_test

import (
	"testing"

	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/jippi/scm-engine/pkg/types"
	"github.com/stretchr/testify/require"
)

// IntID has to cope with GitLab returning user IDs in two shapes: a bare number
// from the REST API and a GID from GraphQL. Getting this wrong means reviewers
// are assigned to user 0.
func TestActor_IntID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   string
		want int
	}{
		{name: "plain numeric id", id: "1234", want: 1234},
		{name: "graphql GID", id: "gid://gitlab/User/1234", want: 1234},
		{name: "GID for the lowest possible user", id: "gid://gitlab/User/1", want: 1},
		// GitLab user IDs start at 1, so 0 is the documented "could not parse"
		{name: "empty id", id: "", want: 0},
		{name: "non numeric id", id: "not-a-number", want: 0},
		{name: "GID of another type is not unwrapped", id: "gid://gitlab/Project/1234", want: 0},
		{name: "username instead of an id", id: "jippi", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, scm.Actor{ID: tt.id}.IntID())
		})
	}
}

// Has is what stops a reviewer being assigned twice across sources.
func TestActors_Has(t *testing.T) {
	t.Parallel()

	actors := scm.Actors{
		{ID: "1", Username: "one"},
		{ID: "2", Username: "two"},
	}

	tests := []struct {
		name  string
		actor scm.Actor
		want  bool
	}{
		{name: "present", actor: scm.Actor{ID: "1", Username: "one"}, want: true},
		{name: "absent", actor: scm.Actor{ID: "3", Username: "three"}, want: false},
		// Both the ID and the username have to match
		{name: "matching id but different username", actor: scm.Actor{ID: "1", Username: "other"}, want: false},
		{name: "matching username but different id", actor: scm.Actor{ID: "9", Username: "one"}, want: false},
		{name: "empty actor", actor: scm.Actor{}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, actors.Has(tt.actor))
		})
	}
}

func TestActors_Has_emptyList(t *testing.T) {
	t.Parallel()

	require.False(t, scm.Actors{}.Has(scm.Actor{ID: "1", Username: "one"}))
}

func TestActors_Add(t *testing.T) {
	t.Parallel()

	var actors scm.Actors

	actors.Add(scm.Actor{ID: "1", Username: "one"})
	actors.Add(scm.Actor{ID: "2", Username: "two"})

	require.Len(t, actors, 2)
	require.True(t, actors.Has(scm.Actor{ID: "1", Username: "one"}))

	// Add does not de-duplicate; callers are expected to check Has first
	actors.Add(scm.Actor{ID: "1", Username: "one"})
	require.Len(t, actors, 3)
}

func TestAppendReviewerIDs(t *testing.T) {
	t.Parallel()

	t.Run("sets the list when empty", func(t *testing.T) {
		t.Parallel()

		opts := &scm.UpdateMergeRequestOptions{}
		opts.AppendReviewerIDs([]int{1, 2})

		require.NotNil(t, opts.ReviewerIDs)
		require.Equal(t, []int{1, 2}, *opts.ReviewerIDs)
	})

	t.Run("appends to an existing list", func(t *testing.T) {
		t.Parallel()

		opts := &scm.UpdateMergeRequestOptions{}
		opts.AppendReviewerIDs([]int{1})
		opts.AppendReviewerIDs([]int{2, 3})

		require.Equal(t, []int{1, 2, 3}, *opts.ReviewerIDs)
	})

	t.Run("appending nothing keeps the list intact", func(t *testing.T) {
		t.Parallel()

		opts := &scm.UpdateMergeRequestOptions{}
		opts.AppendReviewerIDs([]int{1})
		opts.AppendReviewerIDs(nil)

		require.Equal(t, []int{1}, *opts.ReviewerIDs)
	})
}

// IsEqual decides whether a label needs updating at all. A false negative sends
// a pointless API call on every run; a false positive means a label change is
// never applied.
func TestEvaluationResult_IsEqual(t *testing.T) {
	t.Parallel()

	base := scm.EvaluationResult{Name: "bug", Description: "a bug", Color: "#FF0000"}

	tests := []struct {
		name   string
		local  scm.EvaluationResult
		remote *scm.Label
		want   bool
	}{
		{
			name:   "identical",
			local:  base,
			remote: &scm.Label{Name: "bug", Description: "a bug", Color: "#FF0000"},
			want:   true,
		},
		{
			name:   "different name",
			local:  base,
			remote: &scm.Label{Name: "other", Description: "a bug", Color: "#FF0000"},
			want:   false,
		},
		{
			name:   "different description",
			local:  base,
			remote: &scm.Label{Name: "bug", Description: "changed", Color: "#FF0000"},
			want:   false,
		},
		{
			name:   "different color",
			local:  base,
			remote: &scm.Label{Name: "bug", Description: "a bug", Color: "#00FF00"},
			want:   false,
		},
		{
			// GitHub returns colours without the leading "#", so the comparison
			// has to ignore it or every label would look changed on every run.
			name:   "color differing only by the leading hash",
			local:  base,
			remote: &scm.Label{Name: "bug", Description: "a bug", Color: "FF0000"},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := state.WithProvider(t.Context(), "github")
			require.Equal(t, tt.want, tt.local.IsEqual(ctx, tt.remote))
		})
	}
}

// Priority only exists on GitLab, so it must be compared there and ignored
// everywhere else.
func TestEvaluationResult_IsEqual_priorityIsGitLabOnly(t *testing.T) {
	t.Parallel()

	local := scm.EvaluationResult{Name: "bug", Priority: types.ValueFrom(1)}
	remote := &scm.Label{Name: "bug", Priority: types.ValueFrom(2)}

	require.False(t, local.IsEqual(state.WithProvider(t.Context(), "gitlab"), remote),
		"differing priorities must not be equal on GitLab")

	require.True(t, local.IsEqual(state.WithProvider(t.Context(), "github"), remote),
		"priority must be ignored on providers that do not support it")
}

func TestEvaluationResult_IsEqual_priorityNullness(t *testing.T) {
	t.Parallel()

	ctx := state.WithProvider(t.Context(), "gitlab")

	set := scm.EvaluationResult{Name: "bug", Priority: types.ValueFrom(1)}
	unset := scm.EvaluationResult{Name: "bug"}

	require.False(t, set.IsEqual(ctx, &scm.Label{Name: "bug"}),
		"a set priority differs from an unset one")

	require.False(t, unset.IsEqual(ctx, &scm.Label{Name: "bug", Priority: types.ValueFrom(1)}),
		"an unset priority differs from a set one")

	require.True(t, unset.IsEqual(ctx, &scm.Label{Name: "bug"}),
		"two unset priorities are equal")
}
