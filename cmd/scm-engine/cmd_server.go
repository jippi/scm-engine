package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-playground/webhooks/v6/gitlab"
	"github.com/urfave/cli/v2"
)

func serverCmd(cCtx *cli.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /mr", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			slog.Warn("not json")

			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		var evt gitlab.MergeRequestEventPayload
		if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
			slog.Error("failed to decode json request body: %w", err)

			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("POST /push", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			slog.Warn("not json")

			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		var evt gitlab.PushEventPayload
		if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
			slog.Error("failed to decode json request body: %w", err)

			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		w.WriteHeader(http.StatusOK)
	})

	log.Fatal(http.ListenAndServe("0.0.0.0:3000", mux))

	return nil
}
