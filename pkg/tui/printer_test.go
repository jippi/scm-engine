package tui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/jippi/scm-engine/pkg/tui"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/require"
)

// newPrinter builds a printer that writes into buf with colour disabled, so the
// assertions are about the text the user sees rather than the escape sequences
// lipgloss happens to emit for a given terminal.
func newPrinter(buf *bytes.Buffer, options ...tui.PrinterOption) tui.Printer {
	renderer := lipgloss.NewRenderer(buf)
	renderer.SetColorProfile(termenv.Ascii)

	options = append([]tui.PrinterOption{tui.WithWriter(buf)}, options...)

	return tui.NewPrinter(tui.NewStyleWithoutColor(), renderer, options...)
}

func TestPrinter_Sprint(t *testing.T) {
	t.Parallel()

	printer := newPrinter(&bytes.Buffer{})

	require.Equal(t, "hello", printer.Sprint("hello"))
	require.Equal(t, "hello world", printer.Sprint("hello ", "world"))
	// Sprint mirrors fmt.Sprint, which inserts a space between operands when
	// neither side is a string
	require.Equal(t, "1 2", printer.Sprint(1, 2))
	require.Equal(t, "a1", printer.Sprint("a", 1))
	require.Empty(t, printer.Sprint())
}

func TestPrinter_Sprintf(t *testing.T) {
	t.Parallel()

	printer := newPrinter(&bytes.Buffer{})

	require.Equal(t, "hello world", printer.Sprintf("hello %s", "world"))
	require.Equal(t, "count: 42", printer.Sprintf("count: %d", 42))
}

func TestPrinter_Sprintln(t *testing.T) {
	t.Parallel()

	printer := newPrinter(&bytes.Buffer{})

	require.Equal(t, "hello\n", printer.Sprintln("hello"))
	require.True(t, strings.HasSuffix(printer.Sprintln("a", "b"), "\n"))
}

func TestPrinter_Sprintfln(t *testing.T) {
	t.Parallel()

	printer := newPrinter(&bytes.Buffer{})

	require.Equal(t, "hello world\n", printer.Sprintfln("hello %s", "world"))
}

// The Print family writes to the printer's configured writer rather than
// os.Stdout, which is what lets the CLI direct output at stderr.
func TestPrinter_PrintFamilyUsesConfiguredWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		print func(tui.Printer)
		want  string
	}{
		{
			name:  "Print",
			print: func(p tui.Printer) { p.Print("hello") },
			want:  "hello",
		},
		{
			name:  "Printf",
			print: func(p tui.Printer) { p.Printf("hello %s", "world") },
			want:  "hello world",
		},
		{
			name:  "Println",
			print: func(p tui.Printer) { p.Println("hello") },
			want:  "hello\n",
		},
		{
			name:  "Printfln",
			print: func(p tui.Printer) { p.Printfln("hello %s", "world") },
			want:  "hello world\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			tt.print(newPrinter(&buf))

			require.Equal(t, tt.want, buf.String())
		})
	}
}

func TestPrinter_FprintFamilyUsesGivenWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		print func(tui.Printer, *bytes.Buffer)
		want  string
	}{
		{
			name:  "Fprint",
			print: func(p tui.Printer, w *bytes.Buffer) { p.Fprint(w, "hello") },
			want:  "hello",
		},
		{
			name:  "Fprintf",
			print: func(p tui.Printer, w *bytes.Buffer) { p.Fprintf(w, "hello %s", "world") },
			want:  "hello world",
		},
		{
			name:  "Fprintln",
			print: func(p tui.Printer, w *bytes.Buffer) { p.Fprintln(w, "hello") },
			want:  "hello\n",
		},
		{
			name:  "Fprintfln",
			print: func(p tui.Printer, w *bytes.Buffer) { p.Fprintfln(w, "hello %s", "world") },
			want:  "hello world\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				configured bytes.Buffer
				target     bytes.Buffer
			)

			tt.print(newPrinter(&configured), &target)

			require.Equal(t, tt.want, target.String())
			require.Empty(t, configured.String(), "Fprint* must not touch the configured writer")
		})
	}
}

func TestPrinter_ReportsBytesWritten(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	written, err := newPrinter(&buf).Print("hello")
	require.NoError(t, err)
	require.Equal(t, len("hello"), written)
}

func TestPrinter_Box(t *testing.T) {
	t.Parallel()

	t.Run("a header only box", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		newPrinter(&buf).Box("the header")

		out := buf.String()
		require.Contains(t, out, "the header")
		require.NotEmpty(t, strings.TrimSpace(out))
	})

	t.Run("a header and body box", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		newPrinter(&buf).Box("the header", "the body")

		out := buf.String()
		require.Contains(t, out, "the header")
		require.Contains(t, out, "the body")
	})

	t.Run("several body parts are joined", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		newPrinter(&buf).Box("header", "first", "second")

		require.Contains(t, buf.String(), "first second")
	})
}

// The box is laid out to the configured width, which is what keeps the header
// and body boxes aligned with each other.
func TestPrinter_BoxWidth(t *testing.T) {
	t.Parallel()

	widths := map[int]int{}

	for _, width := range []int{40, 100} {
		var buf bytes.Buffer

		newPrinter(&buf, tui.WitBoxWidth(width)).Box("header", "body")

		longest := 0

		for _, line := range strings.Split(buf.String(), "\n") {
			if len(line) > longest {
				longest = len(line)
			}
		}

		widths[width] = longest
	}

	require.Less(t, widths[40], widths[100], "a wider box must produce wider output")
}

func TestNewStyleWithoutColor(t *testing.T) {
	t.Parallel()

	style := tui.NewStyleWithoutColor()

	// The style is usable for every box part without a colour being configured
	require.NotNil(t, style.TextStyle())
	require.NotNil(t, style.TextEmphasisStyle())
	require.NotNil(t, style.BoxHeader())
	require.NotNil(t, style.BoxBody())
}

func TestNewStyle(t *testing.T) {
	t.Parallel()

	style := tui.NewStyle(lipgloss.Color("#FF0000"))

	require.NotNil(t, style.TextStyle())
	require.NotNil(t, style.BoxHeader())
	require.NotNil(t, style.BoxBody())
}

func TestNewTheme(t *testing.T) {
	t.Parallel()

	theme := tui.NewTheme()

	var buf bytes.Buffer

	renderer := lipgloss.NewRenderer(&buf)
	renderer.SetColorProfile(termenv.Ascii)

	writer := theme.Writer(renderer)

	require.NotNil(t, writer)
}
