package gitlab

import (
	"context"
	"log/slog"

	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	slogctx "github.com/veqryn/slog-context"
)

func (c *Client) AssignReviewers(ctx context.Context, evalContext *Context, update *scm.UpdateMergeRequestOptions, step scm.ActionStep) error {
	source, err := step.RequiredStringEnum("source", "codeowners")
	if err != nil {
		return err
	}

	desiredLimit, err := step.RequiredInt("limit")
	if err != nil {
		return err
	}

	mode, err := step.OptionalStringEnum("mode", "random", "random")
	if err != nil {
		return err
	}

	var eligibleReviewers []scm.Actor

	switch source {
	case "codeowners":
		eligibleReviewers = evalContext.GetCodeOwners()

		break
	}

	if len(eligibleReviewers) == 0 {
		slogctx.Debug(ctx, "No eligible reviewers found")

		return nil
	}

	var reviewers scm.Actors

	limit := desiredLimit
	if limit > len(eligibleReviewers) {
		limit = len(eligibleReviewers)
	}

	switch mode {
	case "random":
		reviewers = make(scm.Actors, limit)
		perm := randSource.Perm(len(eligibleReviewers))

		for i := 0; i < limit; i++ {
			reviewers[i] = eligibleReviewers[perm[i]]
		}

		break
	}

	reviewerIDs := make([]int, 0, len(reviewers))

	for _, reviewer := range reviewers {
		id := reviewer.IntID()
		// skip invalid int ids, this should not happen but still safeguard against it
		if id != -1 {
			slogctx.Warn(ctx, "Invalid reviewer ID", slog.String("id", reviewer.ID))

			continue
		}

		reviewerIDs = append(reviewerIDs, reviewer.IntID())
	}

	if state.IsDryRun(ctx) {
		slogctx.Info(ctx, "(Dry Run) Assigning MR", slog.String("source", source), slog.Int("limit", limit), slog.String("mode", mode), slog.Any("reviewers", reviewers))

		return nil
	}

	update.AppendReviewerIDs(reviewerIDs)

	return nil
}
