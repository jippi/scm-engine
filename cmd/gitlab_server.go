package cmd

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jippi/scm-engine/pkg/state"
	"github.com/urfave/cli/v2"
	slogctx "github.com/veqryn/slog-context"
)

type Commit struct {
	ID string `json:"id"`
}

type MergeRequest struct {
	IID        int    `json:"iid"`
	LastCommit Commit `json:"last_commit"`
}

type Project struct {
	PathWithNamespace string `json:"path_with_namespace"`
}

type Payload struct {
	EventType        string        `json:"event_type"`
	Project          Project       `json:"project"`                     // "project" is sent for all events
	ObjectAttributes *MergeRequest `json:"object_attributes,omitempty"` // "object_attributes" is sent on "merge_request" events
	MergeRequest     *MergeRequest `json:"merge_request,omitempty"`     // "merge_request" is sent on "note" activity
}

func errHandler(ctx context.Context, w http.ResponseWriter, code int, err error) {
	switch code {
	case http.StatusOK:
		slogctx.Info(ctx, "Server response", slog.Int("response_code", code), slog.Any("response_message", err))

	default:
		slogctx.Error(ctx, "Server response", slog.Int("response_code", code), slog.Any("response_message", err))
	}

	w.WriteHeader(code)
	w.Write([]byte(err.Error()))

	return
}

func Server(cCtx *cli.Context) error {
	// Initialize context
	ctx := state.WithUpdatePipeline(cCtx.Context, cCtx.Bool(FlagUpdatePipeline))

	// Add BaseURL env to ctx
	ctx = slogctx.With(ctx, slog.String("gitlab_url", cCtx.String(FlagSCMBaseURL)))
	ctx = slogctx.With(ctx, slog.String("config_file", cCtx.String(FlagConfigFile)))

	listenAddr := net.JoinHostPort(cCtx.String(FlagServerListenHost), cCtx.String(FlagServerListenPort))

	slogctx.Info(ctx, "Starting HTTP server", slog.String("listen_address", listenAddr))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /_status", GitLabStatusHandler)
	mux.HandleFunc("POST /gitlab", GitLabWebhookHandler(ctx, cCtx.String(FlagWebhookSecret), cCtx.String(FlagConfigFile)))

	server := &http.Server{
		Addr:         listenAddr,
		Handler:      http.Handler(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slogctx.Error(ctx, "HTTP server error", slog.Any("error", err))

			os.Exit(1)
		}

		slogctx.Info(ctx, "Stopped serving new connections.")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	slogctx.Info(ctx, "Got SIGINT/SIGTERM, starting graceful shutdown.")

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slogctx.Error(ctx, "HTTP shutdown error", slog.Any("error", err))

		os.Exit(1)
	}

	slogctx.Info(ctx, "Graceful shutdown complete.")

	return nil
}
