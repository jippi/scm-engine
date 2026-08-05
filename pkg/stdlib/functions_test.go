package stdlib_test

import (
	"testing"
	"time"

	"github.com/expr-lang/expr"
	"github.com/jippi/scm-engine/pkg/stdlib"
	"github.com/stretchr/testify/require"
)

// evaluate compiles and runs script with the scm-engine standard library
// available, mirroring how user supplied `script:` and `if:` values are run.
func evaluate(t *testing.T, script string, env map[string]any) (any, error) {
	t.Helper()

	opts := make([]expr.Option, 0, len(stdlib.Functions)+2)
	opts = append(opts, expr.Env(env), stdlib.FunctionRenamer)
	opts = append(opts, stdlib.Functions...)

	program, err := expr.Compile(script, opts...)
	if err != nil {
		return nil, err
	}

	return expr.Run(program, env)
}

func TestUniqSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{name: "empty", in: []string{}, want: []string{}},
		{name: "nil", in: nil, want: nil},
		{name: "already unique and sorted", in: []string{"a", "b"}, want: []string{"a", "b"}},
		{name: "sorts the output", in: []string{"c", "a", "b"}, want: []string{"a", "b", "c"}},
		{name: "removes duplicates", in: []string{"b", "a", "b", "a"}, want: []string{"a", "b"}},
		{name: "all identical", in: []string{"x", "x", "x"}, want: []string{"x"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, stdlib.UniqSlice(tt.in))
		})
	}
}

// UniqSlice used to sort its argument in place, which corrupted the caller's
// slice. For script functions that slice is the evaluation context itself.
func TestUniqSlice_doesNotMutateInput(t *testing.T) {
	t.Parallel()

	input := []string{"c", "a", "b", "a"}
	original := []string{"c", "a", "b", "a"}

	got := stdlib.UniqSlice(input)

	require.Equal(t, []string{"a", "b", "c"}, got)
	require.Equal(t, original, input, "UniqSlice must not modify the slice it was given")
}

func TestUniq(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		script  string
		env     map[string]any
		want    any
		wantErr string
	}{
		{
			name:   "documented example",
			script: `["hello", "world", "world"] | uniq()`,
			env:    map[string]any{},
			want:   []string{"hello", "world"},
		},
		{
			name:   "string slice from the environment",
			script: `uniq(labels)`,
			env:    map[string]any{"labels": []string{"b", "a", "b"}},
			want:   []string{"a", "b"},
		},
		{
			name:   "list of any, as produced by map()",
			script: `["b", "a", "b"] | map({ # }) | uniq()`,
			env:    map[string]any{},
			want:   []string{"a", "b"},
		},
		{
			name:   "empty list",
			script: `[] | map({ # }) | uniq()`,
			env:    map[string]any{},
			want:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := evaluate(t, tt.script, tt.env)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

// uniq() only accepts []string and []any, so a non-list argument is rejected by
// the expr type checker at compile time rather than reaching the function body.
// That means a bad `script:` fails during `scm-engine lint`, not mid-evaluation.
func TestUniq_rejectsUnsupportedInput(t *testing.T) {
	t.Parallel()

	_, err := evaluate(t, `uniq(value)`, map[string]any{"value": 42})
	require.ErrorContains(t, err, "cannot use int as argument")
}

func TestUniq_doesNotMutateEvaluationContext(t *testing.T) {
	t.Parallel()

	labels := []string{"c", "a", "b", "a"}
	env := map[string]any{"labels": labels}

	got, err := evaluate(t, `uniq(labels)`, env)
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b", "c"}, got)

	require.Equal(t, []string{"c", "a", "b", "a"}, labels, "uniq() must leave the evaluation context untouched")
}

func TestFilepathDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		script string
		want   any
	}{
		{name: "documented example", script: `filepath_dir("example/directory/file.go")`, want: "example/directory"},
		{name: "empty path returns dot", script: `filepath_dir("")`, want: "."},
		{name: "only separators", script: `filepath_dir("/")`, want: "/"},
		{name: "file at root", script: `filepath_dir("file.go")`, want: "."},
		{name: "trailing separator", script: `filepath_dir("a/b/")`, want: "a/b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := evaluate(t, tt.script, map[string]any{})
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestLimitPathDepthTo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		script string
		want   any
	}{
		// Both examples are taken verbatim from docs/gitlab/script-functions.md
		{
			name:   "documented example, path deeper than the limit",
			script: `limit_path_depth_to("path1/path2/path3/path4", 2)`,
			want:   "path1/path2",
		},
		{
			name:   "documented example, path shallower than the limit",
			script: `limit_path_depth_to("path1/path2", 3)`,
			want:   "path1/path2",
		},
		{
			name:   "path exactly at the limit is returned untouched",
			script: `limit_path_depth_to("path1/path2", 2)`,
			want:   "path1/path2",
		},
		{
			name:   "limit of one keeps the first segment",
			script: `limit_path_depth_to("path1/path2/path3", 1)`,
			want:   "path1",
		},
		{
			name:   "single segment path",
			script: `limit_path_depth_to("path1", 2)`,
			want:   "path1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := evaluate(t, tt.script, map[string]any{})
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		script string
		want   time.Duration
	}{
		{name: "seconds", script: `duration("30s")`, want: 30 * time.Second},
		{name: "hours", script: `duration("2h")`, want: 2 * time.Hour},
		// The days/weeks/months units are why the built-in duration() is overridden
		{name: "days", script: `duration("1d")`, want: 24 * time.Hour},
		{name: "weeks", script: `duration("1w")`, want: 7 * 24 * time.Hour},
		{name: "combined", script: `duration("1d12h")`, want: 36 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := evaluate(t, tt.script, map[string]any{})
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDuration_rejectsGarbage(t *testing.T) {
	t.Parallel()

	_, err := evaluate(t, `duration("not-a-duration")`, map[string]any{})
	require.Error(t, err)
}

func TestSince(t *testing.T) {
	t.Parallel()

	env := map[string]any{"stamp": time.Now().Add(-2 * time.Hour)}

	got, err := evaluate(t, `since(stamp)`, env)
	require.NoError(t, err)

	elapsed, ok := got.(time.Duration)
	require.True(t, ok, "since() should return a time.Duration, got %T", got)
	require.InDelta(t, (2 * time.Hour).Seconds(), elapsed.Seconds(), 60)
}

// The comparison of since() against duration() is the shape used throughout the
// documentation for "has this been stale for N days" style rules.
func TestSince_comparedAgainstDuration(t *testing.T) {
	t.Parallel()

	env := map[string]any{"stamp": time.Now().Add(-10 * 24 * time.Hour)}

	got, err := evaluate(t, `since(stamp) > duration("7d")`, env)
	require.NoError(t, err)
	require.Equal(t, true, got)

	got, err = evaluate(t, `since(stamp) > duration("30d")`, env)
	require.NoError(t, err)
	require.Equal(t, false, got)
}
