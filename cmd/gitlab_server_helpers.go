package cmd

import (
	"context"
	"log/slog"
	"net/http"

	slogctx "github.com/veqryn/slog-context"
)

func errHandler(ctx context.Context, w http.ResponseWriter, code int, err error) {
	slogctx.Error(ctx, "Server response", slog.Int("response_code", code), slog.Any("response_message", err))

	w.WriteHeader(code)
	w.Write([]byte(err.Error()))

	return
}
