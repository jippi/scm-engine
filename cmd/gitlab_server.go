package cmd

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/urfave/cli/v3"
	slogctx "github.com/veqryn/slog-context"
)

func Server(cCtx *cli.Context) error {
	var wg sync.WaitGroup

	// Setup context configuration
	ctx := cCtx.Context
	ctx = state.WithConfigFilePath(ctx, cCtx.String(FlagConfigFile))
	ctx = state.WithUpdatePipeline(ctx, cCtx.Bool(FlagUpdatePipeline), cCtx.String(FlagUpdatePipelineURL))

	// Add logging context key/value pairs
	ctx = slogctx.With(ctx, slog.String("gitlab_url", cCtx.String(FlagSCMBaseURL)))
	ctx = slogctx.With(ctx, slog.Duration("server_timeout", cCtx.Duration(FlagServerTimeout)))

	//
	// Setup periodic evaluation logic
	//

	slogctx.Info(ctx, "Starting periodic evaluation server")

	filter := scm.MergeRequestListFilters{
		IgnoreMergeRequestWithLabels: cCtx.StringSlice(FlagPeriodicEvaluationIgnoreMergeRequestsWithLabel),
		OnlyMergeRequestsWithLabels:  cCtx.StringSlice(FlagPeriodicEvaluationRequireMergeRequestsWithLabel),
		OnlyProjectsWithMembership:   cCtx.Bool(FlagPeriodicEvaluationOnlyProjectsWithMembership),
		OnlyProjectsWithTopics:       cCtx.StringSlice(FlagPeriodicEvaluationOnlyProjectsWithTopics),
		SCMConfigurationFilePath:     cCtx.String(FlagConfigFile),
	}

	evalCtx, stopPeriodicEvaluation := context.WithCancel(ctx)
	startPeriodicEvaluation(evalCtx, cCtx.Duration(FlagPeriodicEvaluationInterval), filter, &wg)

	//
	// Setup HTTP server
	//

	listenAddr := net.JoinHostPort(cCtx.String(FlagServerListenHost), cCtx.String(FlagServerListenPort))
	slogctx.Info(ctx, "Starting HTTP server", slog.String("listen_address", listenAddr))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /_status", GitLabStatusHandler)
	mux.HandleFunc("POST /gitlab", GitLabWebhookHandler(ctx, cCtx.String(FlagWebhookSecret)))

	server := &http.Server{
		Addr:         listenAddr,
		Handler:      mux,
		ReadTimeout:  cCtx.Duration(FlagServerTimeout),
		WriteTimeout: cCtx.Duration(FlagServerTimeout),
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	//
	// Start HTTP server in a Go routine
	//

	wg.Add(1) // +1: HTTP Server

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slogctx.Error(ctx, "HTTP server error", slog.Any("error", err))

			os.Exit(1)
		}

		slogctx.Info(ctx, "Stopped serving new connections.")
	}()

	//
	// Wait for shutdown signals
	//

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	//
	// Graceful shutdown logic
	//

	slogctx.Info(ctx, "Got SIGINT/SIGTERM, starting graceful shutdown.")

	stopPeriodicEvaluation()

	// NOTE: do not use the existing "ctx" since its already cancelled in developer mode if CTRL+C-ing
	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slogctx.Error(ctx, "HTTP shutdown error", slog.Any("error", err))
	}

	wg.Done() // -1: HTTP Server - shutdown complete

	slogctx.Info(ctx, "Graceful HTTP shutdown complete")

	wg.Wait() // Wait for PeriodicEvaluation to complete

	slogctx.Info(ctx, "Graceful shutdown complete")

	return nil
}
