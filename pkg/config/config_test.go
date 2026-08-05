package config_test

import (
	"testing"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/stretchr/testify/require"
)

func TestConfig_Merge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		cfg   *config.Config
		other *config.Config
		want  *config.Config
	}{
		{
			name: "merge when other is nil",
			cfg: &config.Config{
				DryRun: scm.Ptr(false),
			},
			other: nil,
			want: &config.Config{
				DryRun: scm.Ptr(false),
			},
		},
		{
			name: "override dry run",
			cfg: &config.Config{
				DryRun: scm.Ptr(false),
			},
			other: &config.Config{
				DryRun: nil,
			},
			want: &config.Config{
				DryRun: nil,
			},
		},
		{
			name: "merge ignore activity",
			cfg: &config.Config{
				IgnoreActivityFrom: config.IgnoreActivityFrom{
					IsBot:     false,
					Usernames: []string{"user1", "user2"},
					Emails:    []string{"user2@example.com"},
				},
			},
			other: &config.Config{
				IgnoreActivityFrom: config.IgnoreActivityFrom{
					IsBot:     true,
					Usernames: []string{"user3", "user2"},
					Emails:    []string{"user4@example.com"},
				},
			},
			want: &config.Config{
				IgnoreActivityFrom: config.IgnoreActivityFrom{
					IsBot:     true,
					Usernames: []string{"user1", "user2", "user3"},
					Emails:    []string{"user2@example.com", "user4@example.com"},
				},
			},
		},
		{
			name: "merge actions",
			cfg: &config.Config{
				Actions: []config.Action{{Name: "action1"}, {Name: "action2"}},
			},
			other: &config.Config{
				Actions: []config.Action{{Name: "action3"}, {Name: "action2"}},
			},
			want: &config.Config{Actions: []config.Action{{Name: "action1"}, {Name: "action2"}, {Name: "action3"}}},
		},
		{
			name: "merge labels",
			cfg:  &config.Config{Labels: config.Labels{{Name: "label1"}, {Name: "label2"}}},
			other: &config.Config{
				Labels: config.Labels{{Name: "label3"}, {Name: "label2"}},
			},
			want: &config.Config{Labels: config.Labels{{Name: "label1"}, {Name: "label2"}, {Name: "label3"}}},
		},
		{
			name: "merge includes",
			cfg:  &config.Config{Includes: []config.Include{{Project: "project1", Ref: scm.Ptr("ref1"), Files: []string{"file1"}}}},
			other: &config.Config{
				Includes: []config.Include{{Project: "project1", Ref: scm.Ptr("ref1"), Files: []string{"file2"}}},
			},
			want: &config.Config{
				Includes: []config.Include{
					{Project: "project1", Ref: scm.Ptr("ref1"), Files: []string{"file1", "file2"}},
				},
			},
		},
		{
			name: "merge includes de-duplicates files within a project/ref",
			cfg: &config.Config{Includes: []config.Include{
				{Project: "project1", Ref: scm.Ptr("ref1"), Files: []string{"file1", "file2"}},
			}},
			other: &config.Config{Includes: []config.Include{
				{Project: "project1", Ref: scm.Ptr("ref1"), Files: []string{"file2", "file3"}},
			}},
			want: &config.Config{
				Includes: []config.Include{
					{Project: "project1", Ref: scm.Ptr("ref1"), Files: []string{"file1", "file2", "file3"}},
				},
			},
		},
		{
			name: "the same project on a different ref stays a separate include",
			cfg: &config.Config{Includes: []config.Include{
				{Project: "project1", Ref: scm.Ptr("ref1"), Files: []string{"file1"}},
			}},
			other: &config.Config{Includes: []config.Include{
				{Project: "project1", Ref: scm.Ptr("ref2"), Files: []string{"file1"}},
			}},
			want: &config.Config{
				Includes: []config.Include{
					{Project: "project1", Ref: scm.Ptr("ref1"), Files: []string{"file1"}},
					{Project: "project1", Ref: scm.Ptr("ref2"), Files: []string{"file1"}},
				},
			},
		},
		{
			name: "an include without a ref is distinct from one with a ref",
			cfg: &config.Config{Includes: []config.Include{
				{Project: "project1", Files: []string{"file1"}},
			}},
			other: &config.Config{Includes: []config.Include{
				{Project: "project1", Ref: scm.Ptr("ref1"), Files: []string{"file2"}},
			}},
			want: &config.Config{
				Includes: []config.Include{
					{Project: "project1", Files: []string{"file1"}},
					{Project: "project1", Ref: scm.Ptr("ref1"), Files: []string{"file2"}},
				},
			},
		},
		{
			name:  "includes are carried over when only the base config has them",
			cfg:   &config.Config{Includes: []config.Include{{Project: "project1", Files: []string{"file1"}}}},
			other: &config.Config{},
			want: &config.Config{
				Includes: []config.Include{{Project: "project1", Files: []string{"file1"}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.cfg.Merge(tt.other)
			require.Equal(t, tt.want, got)
		})
	}
}

// Merge used to build its result by ranging over a map, so both the order of
// the includes and the order of the files inside them changed from run to run.
// That surfaced as a randomly failing test and, more importantly, as a config
// whose include order could not be relied on.
//
// See https://github.com/jippi/scm-engine/actions/runs/31007145072
func TestConfig_Merge_isDeterministic(t *testing.T) {
	t.Parallel()

	newBase := func() *config.Config {
		return &config.Config{
			Includes: []config.Include{
				{Project: "project-b", Ref: scm.Ptr("ref1"), Files: []string{"b-file1", "b-file2"}},
				{Project: "project-a", Files: []string{"a-file1"}},
			},
		}
	}

	newOther := func() *config.Config {
		return &config.Config{
			Includes: []config.Include{
				{Project: "project-a", Files: []string{"a-file2", "a-file1"}},
				{Project: "project-c", Files: []string{"c-file1"}},
				{Project: "project-b", Ref: scm.Ptr("ref1"), Files: []string{"b-file3"}},
			},
		}
	}

	want := []config.Include{
		{Project: "project-b", Ref: scm.Ptr("ref1"), Files: []string{"b-file1", "b-file2", "b-file3"}},
		{Project: "project-a", Files: []string{"a-file1", "a-file2"}},
		{Project: "project-c", Files: []string{"c-file1"}},
	}

	// Go randomises map iteration order per range statement, so repeating the
	// merge is what makes a map-ordered implementation fail here.
	for i := range 100 {
		got := newBase().Merge(newOther())

		require.Equal(t, want, got.Includes, "include order changed on iteration %d", i)
	}
}
