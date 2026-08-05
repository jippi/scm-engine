package tui_test

import (
	"log/slog"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/jippi/scm-engine/pkg/tui"
	"github.com/stretchr/testify/require"
)

func TestShadeColor(t *testing.T) {
	t.Parallel()

	// Shading by 0% leaves the colour untouched, by 100% takes it to black.
	require.Equal(t, "#FFFFFF", string(tui.ShadeColor("#FFFFFF", 0)))
	require.Equal(t, "#000000", string(tui.ShadeColor("#FFFFFF", 1)))

	// Anything in between is darker than the input but not yet black.
	mid := string(tui.ShadeColor("#FFFFFF", 0.5))
	require.NotEqual(t, "#FFFFFF", mid)
	require.NotEqual(t, "#000000", mid)
	require.Len(t, mid, 7, "expected a #RRGGBB value, got %q", mid)
}

func TestTintColor(t *testing.T) {
	t.Parallel()

	// Tinting by 0% leaves the colour untouched, by 100% takes it to white.
	require.Equal(t, "#000000", string(tui.TintColor("#000000", 0)))
	require.Equal(t, "#FFFFFF", string(tui.TintColor("#000000", 1)))

	mid := string(tui.TintColor("#000000", 0.5))
	require.NotEqual(t, "#000000", mid)
	require.NotEqual(t, "#FFFFFF", mid)
}

func TestShadeAndTintColor_rejectOutOfRangePercent(t *testing.T) {
	t.Parallel()

	for _, percent := range []float64{-0.1, 1.1, 50} {
		require.Panics(t, func() { tui.ShadeColor("#FFFFFF", percent) })
		require.Panics(t, func() { tui.TintColor("#FFFFFF", percent) })
	}
}

func TestColorToHex(t *testing.T) {
	t.Parallel()

	require.Equal(t, "#AABBCC", tui.ColorToHex(lipgloss.Color("#AABBCC")))
	require.Empty(t, tui.ColorToHex(lipgloss.Color("")))
}

func TestTransformColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		base   string
		filter string
		want   string
	}{
		{name: "shade to black", base: "#FFFFFF", filter: "shade", want: "#000000"},
		{name: "tint to white", base: "#000000", filter: "tint", want: "#FFFFFF"},
		{name: "unknown filter returns the base colour", base: "#123456", filter: "nope", want: "#123456"},
		{name: "empty filter returns the base colour", base: "#123456", filter: "", want: "#123456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, tui.TransformColor(tt.base, tt.filter, 1))
		})
	}
}

func TestTransformColor_mixIsUnsupported(t *testing.T) {
	t.Parallel()

	require.PanicsWithValue(t, "unexpected mix filter", func() {
		tui.TransformColor("#FFFFFF", "mix", 0.5)
	})
}

// Label colours may be written as a named palette entry prefixed with "$".
func TestReplace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "a raw hex colour is passed through", input: "#AABBCC", want: "#AABBCC"},
		{name: "a non-palette name is passed through", input: "red", want: "red"},
		{name: "an empty string is passed through", input: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, tui.Replace(tt.input))
		})
	}
}

// Label colours may reference the Bootstrap palette by name with a "$" prefix,
// which Replace resolves to the underlying hex value.
func TestReplace_resolvesPaletteNames(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"purple-300", "gray-800", "red-500"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			want, ok := tui.AllColors[name]
			require.True(t, ok, "%q is missing from the palette", name)

			got := tui.Replace("$" + name)
			require.Equal(t, string(want), got)
			require.NotEqual(t, "$"+name, got, "the $ prefix should have been resolved")
			require.Regexp(t, `^#[0-9a-fA-F]{6}$`, got)
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		fallback slog.Level
		want     slog.Level
	}{
		{name: "debug", input: "DEBUG", want: slog.LevelDebug},
		{name: "info", input: "INFO", want: slog.LevelInfo},
		{name: "warn", input: "WARN", want: slog.LevelWarn},
		{name: "error", input: "ERROR", want: slog.LevelError},
		{name: "lower case is accepted", input: "debug", want: slog.LevelDebug},
		{name: "mixed case is accepted", input: "InFo", want: slog.LevelInfo},
		{
			name:     "an unknown level uses the fallback",
			input:    "nonsense",
			fallback: slog.LevelWarn,
			want:     slog.LevelWarn,
		},
		{
			name:     "an empty level uses the fallback",
			input:    "",
			fallback: slog.LevelError,
			want:     slog.LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, tui.ParseLogLevel(tt.input, tt.fallback))
		})
	}
}

func TestStringDump(t *testing.T) {
	t.Parallel()

	attr := tui.StringDump("payload", "hello")

	require.Equal(t, "payload", attr.Key)
	require.NotEmpty(t, attr.Value.String())
}
