package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/stretchr/testify/require"
)

func readDoc(t *testing.T, elem ...string) string {
	t.Helper()

	path, err := filepath.Abs(filepath.Join(append([]string{"..", ".."}, elem...)...))
	require.NoError(t, err)

	raw, err := os.ReadFile(path)
	require.NoError(t, err, "could not read %s", path)

	return string(raw)
}

// Every action that can be used in a configuration file must appear in the
// reference documentation. assign_reviewers shipped undocumented for several
// releases because nothing tied the two together.
func TestAllActionsAreDocumented(t *testing.T) {
	t.Parallel()

	docs := readDoc(t, "docs", "configuration.md")

	actions := config.ActionNames()
	require.NotEmpty(t, actions, "no actions were registered")

	for _, name := range actions {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.Contains(t, docs, "`#!yaml "+name+"`",
				"action %q is not documented in docs/configuration.md", name)
		})
	}
}

// The JSON schema is what `scm-engine lint` validates user configuration
// against, so an action missing from it cannot be used at all.
func TestAllActionsAreInTheJSONSchema(t *testing.T) {
	t.Parallel()

	schema := config.ActionStep{}.JSONSchema()
	require.NotNil(t, schema)

	// The `action` property lists every accepted value...
	actionProperty, ok := schema.Properties.Get("action")
	require.True(t, ok, "the schema has no `action` property")

	enum := map[string]bool{}

	for _, value := range actionProperty.Enum {
		name, ok := value.(string)
		require.True(t, ok, "enum value %v is not a string", value)

		enum[name] = true
	}

	// ...and each action contributes an if/then branch carrying its own fields.
	branch := map[string]bool{}

	for _, sub := range schema.AllOf {
		if sub == nil || sub.If == nil || sub.If.Properties == nil {
			continue
		}

		match, ok := sub.If.Properties.Get("action")
		if !ok || match == nil {
			continue
		}

		if name, ok := match.Const.(string); ok {
			branch[name] = true
			require.NotNil(t, sub.Then, "action %q has no `then` schema", name)
		}
	}

	for _, name := range config.ActionNames() {
		require.True(t, enum[name], "action %q is missing from the `action` enum in the JSON schema", name)
		require.True(t, branch[name], "action %q has no if/then branch in the JSON schema", name)
	}

	require.Len(t, enum, len(config.ActionNames()), "the JSON schema enum lists actions that are not registered")
}
