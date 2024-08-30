package cmd

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	slogctx "github.com/veqryn/slog-context"
)

func errHandler(ctx context.Context, w http.ResponseWriter, code int, err error) {
	// Treat 404 errors as informational instead of actual errors
	if strings.Contains(err.Error(), "404 Not Found") {
		slogctx.Info(ctx, "Server response", slog.Int("response_code", code), slog.Any("response_message", err))
	} else {
		slogctx.Error(ctx, "Server response", slog.Int("response_code", code), slog.Any("response_message", err))
	}

	w.WriteHeader(code)
	w.Write([]byte(err.Error()))

	return
}
