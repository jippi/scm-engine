package cmd

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	slogctx "github.com/veqryn/slog-context"
)

func ServerPeriodicEvaluation(ctx context.Context, interval time.Duration, filter scm.ProjectListFilter) {
	if interval == 0 {
		slogctx.Warn(ctx, "scm-engine will not be doing periodic evaluation since interval is '0'. Set 'SCM_ENGINE_PERIODIC_EVALUATION_INTERVAL' or '--periodic-evaluation-interval'  to a non-zero duration to activate")

		return
	}

	// Initialize GitLab client
	client, err := getClient(ctx)
	if err != nil {
		panic(err)
	}

	ctx = slogctx.With(ctx,
		slog.String("subsystem", "periodic_evaluation"),
		slog.Duration("periodic_evaluation_interval", interval),
		slog.Any("periodic_evaluation_filters", filter.AsGraphqlVariables()),
	)

	go func() {
		timer := time.NewTicker(interval)

		for {
			slogctx.Info(ctx, "Waiting for next periodic evaluation")
			<-timer.C
			slogctx.Info(ctx, "Starting periodic evaluation cycle")

			results, err := client.FindMergeRequestsForPeriodicEvaluation(ctx, filter)
			if err != nil {
				slogctx.Error(ctx, "Failed to generate merge request list to evaluate", slog.Any("error", err))

				continue
			}

			for _, mergeRequest := range results {
				ctx := state.ContextWithMergeRequestID(ctx, mergeRequest.MergeRequestID)
				ctx = state.WithProjectID(ctx, mergeRequest.Project)
				ctx = state.WithCommitSHA(ctx, mergeRequest.SHA)

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

				ProcessMR(ctx, client, cfg, nil)
			}

			panic("end)")
		}
	}()
}
