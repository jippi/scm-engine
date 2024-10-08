package tui

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/golang-cz/devslog"
	"github.com/lmittmann/tint"
)

const pkgPrefix = "github.com/jippi/dottie"

func ParseLogLevel(name string, fallback slog.Level) slog.Level {
	switch strings.ToUpper(name) {
	case "DEBUG":
		return slog.LevelDebug

	case "INFO":
		return slog.LevelInfo

	case "WARN":
		return slog.LevelWarn

	case "ERROR":
		return slog.LevelError

	default:
		return fallback
	}
}

func logHandler(out io.Writer) slog.Handler {
	logLevel := ParseLogLevel(os.Getenv("LOG_LEVEL"), slog.LevelInfo)

	if val := os.Getenv("LOG_FORMAT"); val == "json" {
		return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     logLevel,
			AddSource: logLevel == slog.LevelDebug,
		})
	}

	if _, ok := os.LookupEnv("CI"); ok {
		return tint.NewHandler(
			out,
			&tint.Options{
				Level:     logLevel,
				AddSource: logLevel == slog.LevelDebug,
			},
		)
	}

	return devslog.NewHandler(
		out,
		&devslog.Options{
			SortKeys:          true,
			MaxSlicePrintSize: 999,
			HandlerOptions: &slog.HandlerOptions{
				Level:     logLevel,
				AddSource: logLevel == slog.LevelDebug,
			},
		},
	)
}

func StringDump(key, value string) slog.Attr {
	return slog.Group(
		key,
		slog.String("Raw", value),
		slog.String("Glyph", fmt.Sprintf("%q", value)),
		slog.String("UTF-8", fmt.Sprintf("% x", []rune(value))),
		slog.String("Unicode", fmt.Sprintf("%U", []rune(value))),
		slog.String("[]rune", fmt.Sprintf("%v", []rune(value))),
		slog.String("[]byte", fmt.Sprintf("%v", []byte(value))),
	)
}
