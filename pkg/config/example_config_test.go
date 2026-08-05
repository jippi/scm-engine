package config_test

import (
	"path/filepath"
	"testing"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm/gitlab"
	"github.com/stretchr/testify/require"
)

// The example configuration shipped in the repository is the closest thing to an
// end-to-end check that every documented script function still compiles against
// the real evaluation context.
//
// This exists because a patch release of expr-lang once stopped binding variadic
// methods on value receivers, which silently broke every
// merge_request.modified_files(...) script. Nothing in the unit tests noticed,
// because nothing compiled a realistic configuration.
func TestExampleConfigCompiles(t *testing.T) {
	t.Parallel()

	path, err := filepath.Abs(filepath.Join("..", "..", ".scm-engine.gitlab.example.yml"))
	require.NoError(t, err)

	cfg, err := config.LoadFile(path)
	require.NoError(t, err, "the shipped example config must parse")
	require.NotNil(t, cfg)

	require.NotEmpty(t, cfg.Labels, "example config is expected to define labels")

	// Lint compiles every label and action script against the GitLab evaluation
	// context, which is where a broken script function surfaces.
	require.NoError(t, cfg.Lint(t.Context(), &gitlab.Context{}))
}

// Each documented script function is compiled against the real GitLab context so
// that a change to the context shape or to expr-lang cannot silently break the
// documented usage.
func TestDocumentedScriptFunctionsCompile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		script string
	}{
		{
			name:   "modified_files with a single pattern",
			script: `merge_request.modified_files("docs/")`,
		},
		{
			name:   "modified_files with several patterns",
			script: `merge_request.modified_files("*.go", "*_test.go")`,
		},
		{
			name:   "modified_files_list",
			script: `len(merge_request.modified_files_list("*.go")) > 0`,
		},
		{
			name:   "modified_files combined with a boolean operator",
			script: `not merge_request.modified_files("*_test.go") && merge_request.modified_files("*.go")`,
		},
		{
			name:   "state_is",
			script: `merge_request.state_is("opened")`,
		},
		{
			name:   "filepath_dir over a mapped list",
			script: `merge_request.modified_files_list("*") | map({ filepath_dir(#) }) | uniq() | len() > 0`,
		},
		{
			name:   "limit_path_depth_to",
			script: `limit_path_depth_to("a/b/c/d", 2) == "a/b"`,
		},
		{
			name:   "duration compared against a timestamp",
			script: `duration("7d") > duration("1h")`,
		},
		{
			name:   "first_commit and last_commit are reachable",
			script: `merge_request.first_commit != nil && merge_request.last_commit != nil`,
		},
		{
			name:   "diff stat totals added by the upstream merge",
			script: `merge_request.total_lines_added() >= 0 && merge_request.total_lines_deleted() >= 0`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{
				Labels: config.Labels{
					{
						Name:     "test-label",
						Script:   tt.script,
						Strategy: "conditional",
					},
				},
			}

			require.NoError(t, cfg.Lint(t.Context(), &gitlab.Context{}), "script failed to compile: %s", tt.script)
		})
	}
}
