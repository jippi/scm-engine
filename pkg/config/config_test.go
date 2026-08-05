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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.cfg.Merge(tt.other)
			require.Equal(t, tt.want, got)
		})
	}
}
