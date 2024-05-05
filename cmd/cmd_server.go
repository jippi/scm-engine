package cmd

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-playground/webhooks/v6/gitlab"
	"github.com/urfave/cli/v2"
)

func Server(_ *cli.Context) error { //nolint:unparam
	mux := http.NewServeMux()

	mux.HandleFunc("POST /mr", func(writer http.ResponseWriter, reader *http.Request) {
		if reader.Header.Get("Content-Type") != "application/json" {
			slog.Warn("not json")

			writer.WriteHeader(http.StatusInternalServerError)

			return
		}

		var evt gitlab.MergeRequestEventPayload
		if err := json.NewDecoder(reader.Body).Decode(&evt); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)

			return
		}

		writer.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("POST /push", func(writer http.ResponseWriter, reader *http.Request) {
		if reader.Header.Get("Content-Type") != "application/json" {
			slog.Warn("not json")

			writer.WriteHeader(http.StatusInternalServerError)

			return
		}

		var evt gitlab.PushEventPayload
		if err := json.NewDecoder(reader.Body).Decode(&evt); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)

			return
		}

		writer.WriteHeader(http.StatusOK)
	})

	log.Fatal(http.ListenAndServe("0.0.0.0:3000", mux)) //nolint:gosec

	return nil
}
