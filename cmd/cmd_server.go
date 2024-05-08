package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm/gitlab"
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/urfave/cli/v2"
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

func errHandler(w http.ResponseWriter, code int, err error) {
	slog.Error(err.Error())

	w.WriteHeader(code)
	w.Write([]byte(err.Error()))

	return
}

func Server(cCtx *cli.Context) error { //nolint:unparam
	mux := http.NewServeMux()

	ourSecret := cCtx.String(FlagWebhookSecret)

	// Initialize GitLab client
	client, err := gitlab.NewClient(cCtx.String(FlagAPIToken), cCtx.String(FlagSCMBaseURL))
	if err != nil {
		return err
	}

	mux.HandleFunc("POST /gitlab", func(writer http.ResponseWriter, reader *http.Request) {
		if len(ourSecret) > 0 {
			theirSecret := reader.Header.Get("X-Gitlab-Token")
			if ourSecret != theirSecret {
				errHandler(writer, http.StatusForbidden, errors.New("Missing or invalid X-Gitlab-Token header"))

				return
			}
		}

		// Validate headers
		if reader.Header.Get("Content-Type") != "application/json" {
			errHandler(writer, http.StatusInternalServerError, errors.New("not json"))

			return
		}

		body, err := io.ReadAll(reader.Body)
		if err != nil {
			errHandler(writer, http.StatusInternalServerError, err)

			return
		}

		// Decode request payload
		var payload Payload
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&payload); err != nil {
			errHandler(writer, http.StatusInternalServerError, err)

			return
		}

		// Initialize context
		ctx := state.ContextWithProjectID(reader.Context(), payload.Project.PathWithNamespace)

		// Grab event specific information
		var (
			id  string
			ref string
		)

		switch payload.EventType {
		case "merge_request":
			id = strconv.Itoa(payload.ObjectAttributes.IID)
			ref = payload.ObjectAttributes.LastCommit.ID

		case "note":
			id = strconv.Itoa(payload.MergeRequest.IID)
			ref = payload.MergeRequest.LastCommit.ID

		default:
			errHandler(writer, http.StatusInternalServerError, fmt.Errorf("unknown event: %s", payload.EventType))
		}

		// Get the remote config file
		file, err := client.MergeRequests().GetRemoteConfig(ctx, cCtx.String(FlagConfigFile), ref)
		if err != nil {
			errHandler(writer, http.StatusOK, err)

			return
		}

		// Parse the file
		cfg, err := config.ParseFile(file)
		if err != nil {
			errHandler(writer, http.StatusOK, err)

			return
		}

		// Decode request payload
		var full any
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&full); err != nil {
			errHandler(writer, http.StatusInternalServerError, err)

			return
		}

		// Process the MR
		if err := ProcessMR(ctx, client, cfg, id, full); err != nil {
			errHandler(writer, http.StatusOK, err)

			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("OK"))
	})

	log.Fatal(http.ListenAndServe("0.0.0.0:3000", mux)) //nolint:gosec

	return nil
}
