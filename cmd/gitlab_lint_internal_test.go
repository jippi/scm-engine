//nolint:testpackage,paralleltest // Lint is driven through the cli package; urfave/cli mutates a package level HelpFlag inside App.Run, so these cannot run in parallel
package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

// runLint invokes the lint command exactly as the CLI does, against a config
// file written for the test.
//
// These tests deliberately do not call t.Parallel: urfave/cli mutates its
// package level HelpFlag inside App.Run, so running several apps concurrently
// is a data race in the library rather than in scm-engine.
func runLint(t *testing.T, contents string) error {
	t.Helper()

	path := filepath.Join(t.TempDir(), ".scm-engine.yml")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: FlagConfigFile, Value: path},
			&cli.StringFlag{Name: "schema", Value: "embed://"},
		},
		Action: Lint,
	}

	return app.Run([]string{"scm-engine"})
}

func TestLint_acceptsAValidConfig(t *testing.T) {
	require.NoError(t, runLint(t, `
label:
  - name: bug
    script: merge_request.has_label("bug")
    color: "$red-500"

actions:
  - name: close-stale
    if: merge_request.has_no_activity_within("30d")
    then:
      - action: close
`))
}

func TestLint_acceptsAnEmptyConfig(t *testing.T) {
	require.NoError(t, runLint(t, "{}\n"))
}

func TestLint_rejectsAMissingFile(t *testing.T) {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: FlagConfigFile, Value: filepath.Join(t.TempDir(), "does-not-exist.yml")},
			&cli.StringFlag{Name: "schema", Value: "embed://"},
		},
		Action: Lint,
	}

	require.Error(t, app.Run([]string{"scm-engine"}))
}

func TestLint_rejectsMalformedYAML(t *testing.T) {
	require.Error(t, runLint(t, "label: [unclosed\n"))
}

// The JSON schema is what catches a key the user misspelled, before any script
// is compiled.
func TestLint_rejectsUnknownKeys(t *testing.T) {
	err := runLint(t, `
labels:
  - name: typo-in-the-top-level-key
`)
	require.Error(t, err)
}

func TestLint_rejectsAnInvalidStrategy(t *testing.T) {
	err := runLint(t, `
label:
  - name: bug
    strategy: not-a-strategy
    script: "true"
`)
	require.Error(t, err)
}

// A script that does not compile has to fail lint, which is the whole point of
// running the command in CI before merging a configuration change.
func TestLint_rejectsAnUncompilableScript(t *testing.T) {
	err := runLint(t, `
label:
  - name: broken
    script: this is not )( valid expr
`)
	require.ErrorContains(t, err, "broken")
}

func TestLint_rejectsAnUncompilableAction(t *testing.T) {
	err := runLint(t, `
actions:
  - name: broken
    if: this is not )( valid expr
    then:
      - action: close
`)
	require.ErrorContains(t, err, "broken")
}

// A generate label returning a boolean is a strategy mismatch, and lint is where
// the user should find out.
func TestLint_rejectsAStrategyMismatch(t *testing.T) {
	err := runLint(t, `
label:
  - strategy: generate
    script: "true"
`)
	require.Error(t, err)
}

func TestLint_rejectsAnUnknownAction(t *testing.T) {
	err := runLint(t, `
actions:
  - name: bad-action
    if: "true"
    then:
      - action: no_such_action
`)
	require.Error(t, err)
}
