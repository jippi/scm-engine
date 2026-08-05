package gitlab_test

import (
	"testing"

	"github.com/jippi/scm-engine/pkg/scm/gitlab"
	"github.com/stretchr/testify/require"
)

func TestParseID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      any
		want    string
		wantErr string
	}{
		{name: "numeric project id", in: 1234, want: "1234"},
		{name: "zero", in: 0, want: "0"},
		{name: "negative", in: -1, want: "-1"},
		{name: "project slug", in: "jippi/scm-engine", want: "jippi/scm-engine"},
		{name: "empty string is passed through", in: "", want: ""},
		{name: "float is rejected", in: 12.5, wantErr: "the ID must be an int or a string"},
		{name: "nil is rejected", in: nil, wantErr: "the ID must be an int or a string"},
		{name: "slice is rejected", in: []string{"a"}, wantErr: "the ID must be an int or a string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := gitlab.ParseID(tt.in)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				require.Empty(t, got)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParseProjectName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      any
		want    string
		wantErr string
	}{
		{name: "group and project", in: "jippi/scm-engine", want: "scm-engine"},
		{name: "nested subgroups use the last segment", in: "group/subgroup/project", want: "project"},
		{name: "deeply nested subgroups", in: "a/b/c/d/project", want: "project"},
		{name: "project name containing dots", in: "group/my.project", want: "my.project"},
		{name: "bare project name is rejected", in: "scm-engine", wantErr: "the ID must be a group/project"},
		{name: "empty string is rejected", in: "", wantErr: "the ID must be a group/project"},
		// A numeric project ID carries no group/project structure to split on
		{name: "numeric project id is rejected", in: 1234, wantErr: "the ID must be a group/project"},
		{name: "invalid id type is propagated", in: 12.5, wantErr: "the ID must be an int or a string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := gitlab.ParseProjectName(tt.in)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				require.Empty(t, got)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
