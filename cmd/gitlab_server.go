package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jippi/scm-engine/pkg/config"
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

	listenAddr := net.JoinHostPort(cCtx.String(FlagServerListenHost), cCtx.String(FlagServerListenPort))

	slogctx.Info(ctx, "Starting HTTP server", slog.String("listen_address", listenAddr))

	mux := http.NewServeMux()

	ourSecret := cCtx.String(FlagWebhookSecret)

	// Initialize client
	client, err := getClient(cCtx.Context)
	if err != nil {
		return err
	}

	mux.HandleFunc("GET /_status", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		slogctx.Debug(ctx, "GET /_status")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("scm-engine status: OK\n\nNOTE: this is a static 'OK', no actual checks are being made"))
	})

	mux.HandleFunc("POST /gitlab", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		slogctx.Info(ctx, "GET /gitlab request")

		// Check if the webhook secret is set (and if its matching)
		if len(ourSecret) > 0 {
			theirSecret := r.Header.Get("X-Gitlab-Token")
			if ourSecret != theirSecret {
				errHandler(ctx, w, http.StatusForbidden, errors.New("Missing or invalid X-Gitlab-Token header"))

				return
			}
		}

		// Validate content type
		if r.Header.Get("Content-Type") != "application/json" {
			errHandler(ctx, w, http.StatusNotAcceptable, errors.New("The request is not using Content-Type: application/json"))

			return
		}

		// Read the POST body of the request
		body, err := io.ReadAll(r.Body)
		if err != nil {
			errHandler(ctx, w, http.StatusBadRequest, err)

			return
		}

		// Ensure we have content in the POST body
		if len(body) == 0 {
			errHandler(ctx, w, http.StatusBadRequest, errors.New("The POST body is empty; expected a JSON payload"))
		}

		// Decode request payload
		var payload Payload
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&payload); err != nil {
			errHandler(ctx, w, http.StatusBadRequest, fmt.Errorf("could not decode POST body into Payload struct: %w", err))

			return
		}

		// Initialize context
		ctx = state.WithProjectID(ctx, payload.Project.PathWithNamespace)

		// Grab event specific information
		var (
			id     string
			gitSha string
		)

		switch payload.EventType {
		case "merge_request":
			id = strconv.Itoa(payload.ObjectAttributes.IID)
			gitSha = payload.ObjectAttributes.LastCommit.ID

		case "note":
			id = strconv.Itoa(payload.MergeRequest.IID)
			gitSha = payload.MergeRequest.LastCommit.ID

		default:
			errHandler(ctx, w, http.StatusInternalServerError, fmt.Errorf("unknown event type: %s", payload.EventType))

			return
		}

		// Build context for rest of the pipeline
		ctx = state.WithCommitSHA(ctx, gitSha)
		ctx = state.ContextWithMergeRequestID(ctx, id)
		ctx = slogctx.With(ctx, slog.String("event_type", payload.EventType))
		ctx = slogctx.With(ctx, slog.String("config_file", cCtx.String(FlagConfigFile)))

		// Get the remote config file
		file, err := client.MergeRequests().GetRemoteConfig(ctx, cCtx.String(FlagConfigFile), gitSha)
		if err != nil {
			errHandler(ctx, w, http.StatusOK, fmt.Errorf("could not read remote config file: %w", err))

			return
		}

		// Parse the file
		cfg, err := config.ParseFile(file)
		if err != nil {
			errHandler(ctx, w, http.StatusOK, fmt.Errorf("could not parse config file: %w", err))

			return
		}

		// Decode request payload into 'any' so we have all the details
		var fullEventPayload any
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&fullEventPayload); err != nil {
			errHandler(ctx, w, http.StatusInternalServerError, err)

			return
		}

		// Process the MR
		if err := ProcessMR(ctx, client, cfg, fullEventPayload); err != nil {
			errHandler(ctx, w, http.StatusOK, err)

			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

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
