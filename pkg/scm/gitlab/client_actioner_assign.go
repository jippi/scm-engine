package gitlab

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	slogctx "github.com/veqryn/slog-context"
)

func (c *Client) AssignReviewers(ctx context.Context, evalContext scm.EvalContext, update *scm.UpdateMergeRequestOptions, step scm.ActionStep) error {
	source, err := step.OptionalStringEnum("source", "codeowners", "codeowners", "backstage", "static")
	if err != nil {
		return err
	}

	desiredLimit, err := step.OptionalInt("limit", 1)
	if err != nil {
		return err
	}

	mode, err := step.OptionalStringEnum("mode", "random", "random")
	if err != nil {
		return err
	}

	// prevents misuse and situations where evaluate will assign reviewers endlessly
	existingReviewers := evalContext.GetReviewers()
	if len(existingReviewers) > 0 {
		slogctx.Debug(ctx, "Reviewers already assigned", slog.Any("reviewers", existingReviewers))

		return nil
	}

	var eligibleReviewers []scm.Actor

	switch source {
	case "codeowners":
		eligibleReviewers = evalContext.GetCodeOwners()

		break
	case "backstage":
		if c.backstage == nil {
			slogctx.Warn(ctx, "Backstage client not initialized and source is backstage, skipping")

			break
		}

		projectName, err := ParseProjectName(state.ProjectID(ctx))
		if err != nil {
			return err
		}

		owners, err := c.backstage.GetOwnersForGitLabProject(ctx, projectName)
		if err != nil {
			return err
		}

		authorID := strconv.Itoa(evalContext.GetAuthor().IntID())
		for _, owner := range owners {
			if authorID != owner.ID {
				eligibleReviewers = append(eligibleReviewers, owner)
			}
		}

		break
	case "static":
		userIDs, err := step.RequiredStringSlice("user_ids")
		if err != nil {
			return err
		}

		for _, id := range userIDs {
			eligibleReviewers = append(eligibleReviewers, scm.Actor{ID: id})
		}

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

		rand := state.RandomSeed(ctx)
		perm := rand.Perm(len(eligibleReviewers))

		for i := 0; i < limit; i++ {
			reviewers[i] = eligibleReviewers[perm[i]]
		}

		break
	}

	reviewerIDs := make([]int, 0, len(reviewers))

	for _, reviewer := range reviewers {
		id := reviewer.IntID()

		// skip invalid int ids, this should not happen but still safeguard against it
		if id == 0 {
			slogctx.Warn(ctx, "Invalid reviewer ID", slog.String("id", reviewer.ID))

			continue
		}

		reviewerIDs = append(reviewerIDs, id)
	}

	if state.IsDryRun(ctx) {
		slogctx.Info(ctx, "(Dry Run) Assigning MR", slog.String("source", source), slog.Int("limit", limit), slog.String("mode", mode), slog.Any("reviewers", reviewers))

		return nil
	}

	update.AppendReviewerIDs(reviewerIDs)

	return nil
}
