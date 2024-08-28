package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/state"
	slogctx "github.com/veqryn/slog-context"
)

func GitLabStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	slogctx.Debug(ctx, "GET /_status")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("scm-engine status: OK\n\nNOTE: this is a static 'OK', no actual checks are being made"))
}

func GitLabWebhookHandler(ctx context.Context, ourSecret, configFilePath string) http.HandlerFunc {
	// Initialize GitLab client
	client, err := getClient(ctx)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
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
		var payload GitlabWebhookPayload
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
		ctx = state.WithMergeRequestID(ctx, id)
		ctx = slogctx.With(ctx, slog.String("event_type", payload.EventType))

		// Get the remote config file
		file, err := client.MergeRequests().GetRemoteConfig(ctx, configFilePath, gitSha)
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
	}
}
