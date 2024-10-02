package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	slogctx "github.com/veqryn/slog-context"
)

func startPeriodicEvaluation(ctx context.Context, interval time.Duration, filter scm.MergeRequestListFilters, wg *sync.WaitGroup) {
	// Empty interval means disabling
	if interval == 0 {
		slogctx.Warn(ctx, "scm-engine will not be doing periodic evaluation since interval is '0'. Set 'SCM_ENGINE_PERIODIC_EVALUATION_INTERVAL' or '--periodic-evaluation-interval'  to a non-zero duration to activate")

		return
	}

	wg.Add(1) // +1: Periodic Evaluation

	// I can't think of a good reason why anyone would want to have it running more frequently than every 15m, so enforcing a floor value.
	//
	// If you are reading this code and need more frequent periodic evaluations then please open a PR or issue with use-case
	// and I will happily re-evaluate this logic.
	if interval < 15*time.Minute {
		slogctx.Warn(ctx, "'SCM_ENGINE_PERIODIC_EVALUATION_INTERVAL' / '--periodic-evaluation-interval' is set to a value less than '15 minutes'; changing the value to '15m'")

		interval = 15 * time.Minute
	}

	// Initialize the SCM-Engine client
	client, err := getClient(ctx)
	if err != nil {
		panic(err)
	}

	// Configure logger and custom fields
	ctx = slogctx.With(ctx,
		slog.Any("periodic_evaluation_filters", filter.AsGraphqlVariables()),
		slog.Duration("periodic_evaluation_interval", interval),
		slog.String("event_type", "periodic_evaluation"),
	)

	go func(wg *sync.WaitGroup) {
		defer wg.Done() // -1: Periodic Evaluation

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		slogctx.Info(ctx, "Waiting for first evaluation cycle tick to happen")

		for {
			select {
			case <-ctx.Done():
				slogctx.Info(ctx, "Stopping periodic evaluation as scm-engine is shutting down")

				return

			case <-ticker.C:
				// Track all log output back to a periodic evaluation cycle
				ctx := slogctx.With(ctx, slog.String("periodic_eval_id", sid.MustGenerate()))

				slogctx.Info(ctx, "Starting periodic evaluation cycle")

				results, err := client.FindMergeRequestsForPeriodicEvaluation(ctx, filter)
				if err != nil {
					slogctx.Error(ctx, "Failed to generate merge request list to evaluate", slog.Any("error", err))

					continue
				}

				slogctx.Info(ctx, fmt.Sprintf("Found %d Merge Requests to evaluate", len(results)), slog.Int("number_of_projects", len(results)))

				for _, mergeRequest := range results {
					ctx := ctx // make sure we define a fresh GC-able context per merge request so we don't append to the existing forever
					ctx = state.WithCommitSHA(ctx, mergeRequest.SHA)
					ctx = state.WithMergeRequestID(ctx, mergeRequest.MergeRequestID)
					ctx = state.WithProjectID(ctx, mergeRequest.Project)

					if !mergeRequest.UpdatePipeline {
						slogctx.Info(ctx, "Disabling CI pipeline commit status updating since the MR HEAD CI pipeline is in a failed state")

						ctx = state.WithUpdatePipeline(ctx, false, "")
					}

					if len(mergeRequest.ConfigBlob) == 0 {
						slogctx.Warn(ctx, "Could not find the scm-engine configuration file in the repository, skipping...")

						continue
					}

					// Parse the file
					cfg, err := config.ParseFile(strings.NewReader(mergeRequest.ConfigBlob))
					if err != nil {
						slogctx.Error(ctx, "could not parse config file", slog.Any("error", err))

						continue
					}

					// Process the Merge Request
					if err := ProcessMR(ctx, client, cfg, nil); err != nil {
						slogctx.Error(ctx, "failed to process MR", slog.Any("error", err))

						continue
					}
				} // end loop results

				slogctx.Info(ctx, "Completed periodic evaluation cycle")
			} // end select
		} // end loop
	}(wg)
}
