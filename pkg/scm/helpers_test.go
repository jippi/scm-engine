package scm_test

import (
	"testing"

	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/stretchr/testify/require"
)

func TestFindModifiedFiles(t *testing.T) {
	t.Parallel()

	type args struct {
		files    []string
		patterns []string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "patterns",
			args: args{
				files: []string{
					".dockerignore",
					"docker-compose.ci-override.yaml",
					"docker-compose.ci-override.yml",
					"docker-compose.test.yaml",
					"docker-compose.yaml",
					"docker-compose.yml",
					"docker-sync.yaml",
					"docker-sync.yml",
					"Dockerfile",
					"some-other-file.txt",
				},
				patterns: []string{
					".dockerignore",
					"*.Dockerfile",
					"docker-compose.*.y*ml",
					"docker-compose.y*ml",
					"docker-sync.y*ml",
					"Dockerfile",
				},
			},
			want: []string{
				".dockerignore",
				"docker-compose.ci-override.yaml",
				"docker-compose.ci-override.yml",
				"docker-compose.test.yaml",
				"docker-compose.yaml",
				"docker-compose.yml",
				"docker-sync.yaml",
				"docker-sync.yml",
				"Dockerfile",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, scm.FindModifiedFiles(tt.args.files, tt.args.patterns...))
		})
	}
}
