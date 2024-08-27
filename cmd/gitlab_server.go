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

func Server(cCtx *cli.Context) error {
	// Setup context configuration
	ctx := state.WithUpdatePipeline(cCtx.Context, cCtx.Bool(FlagUpdatePipeline), cCtx.String(FlagUpdatePipelineURL))

	// Add logging context key/value pairs
	ctx = slogctx.With(ctx, slog.String("gitlab_url", cCtx.String(FlagSCMBaseURL)))
	ctx = slogctx.With(ctx, slog.String("config_file", cCtx.String(FlagConfigFile)))
	ctx = slogctx.With(ctx, slog.Duration("server_timeout", cCtx.Duration(FlagServerTimeout)))

	listenAddr := net.JoinHostPort(cCtx.String(FlagServerListenHost), cCtx.String(FlagServerListenPort))

	slogctx.Info(ctx, "Starting HTTP server", slog.String("listen_address", listenAddr))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /_status", GitLabStatusHandler)
	mux.HandleFunc("POST /gitlab", GitLabWebhookHandler(ctx, cCtx.String(FlagWebhookSecret), cCtx.String(FlagConfigFile)))

	server := &http.Server{
		Addr:         listenAddr,
		Handler:      mux,
		ReadTimeout:  cCtx.Duration(FlagServerTimeout),
		WriteTimeout: cCtx.Duration(FlagServerTimeout),
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
