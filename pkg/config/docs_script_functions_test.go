package config_test

import (
	"reflect"
	"strings"
	"testing"
	"unicode"

	"github.com/jippi/scm-engine/pkg/scm/github"
	"github.com/jippi/scm-engine/pkg/scm/gitlab"
	"github.com/stretchr/testify/require"
)

// scriptName converts an exported Go method name to the snake_case name the
// expr-lang renamer exposes to user scripts.
func scriptName(method string) string {
	var out strings.Builder

	for i, char := range method {
		if unicode.IsUpper(char) {
			if i > 0 {
				out.WriteRune('_')
			}

			out.WriteRune(unicode.ToLower(char))

			continue
		}

		out.WriteRune(char)
	}

	return out.String()
}

// exportedMethods returns the script-callable method names on a context type.
func exportedMethods(target any) []string {
	typ := reflect.TypeOf(target)

	names := make([]string, 0, typ.NumMethod())

	for i := range typ.NumMethod() {
		names = append(names, typ.Method(i).Name)
	}

	return names
}

// Every script function callable on a Merge Request must appear in the GitLab
// script function reference. has_any_activity_within, total_lines_added and
// total_lines_deleted all shipped undocumented before this test existed.
func TestGitLabScriptFunctionsAreDocumented(t *testing.T) {
	t.Parallel()

	docs := readDoc(t, "docs", "gitlab", "script-functions.md")

	for _, method := range exportedMethods(gitlab.ContextMergeRequest{}) {
		name := scriptName(method)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.Contains(t, docs, "merge_request."+name,
				"merge_request.%s is callable from scripts but missing from docs/gitlab/script-functions.md", name)
		})
	}
}

// Same contract for the GitHub provider.
func TestGitHubScriptFunctionsAreDocumented(t *testing.T) {
	t.Parallel()

	docs := readDoc(t, "docs", "github", "script-functions.md")

	for _, method := range exportedMethods(github.ContextPullRequest{}) {
		name := scriptName(method)

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.Contains(t, docs, "pull_request."+name,
				"pull_request.%s is callable from scripts but missing from docs/github/script-functions.md", name)
		})
	}
}

func TestScriptName(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"HasLabel":             "has_label",
		"ModifiedFilesList":    "modified_files_list",
		"StateIs":              "state_is",
		"TotalLinesAdded":      "total_lines_added",
		"HasAnyActivityWithin": "has_any_activity_within",
		"IsApproved":           "is_approved",
	}

	for method, want := range tests {
		t.Run(method, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, want, scriptName(method))
		})
	}
}
