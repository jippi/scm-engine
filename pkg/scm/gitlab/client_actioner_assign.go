package gitlab

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	slogctx "github.com/veqryn/slog-context"
)

func (c *Client) AssignReviewers(ctx context.Context, evalContext scm.EvalContext, update *scm.UpdateMergeRequestOptions, step scm.ActionStep) error {
	source, err := step.RequiredStringEnum("source", "codeowners", "backstage", "static")
	if err != nil {
		return err
	}

	desiredLimit, err := step.OptionalInt("limit", 1)
	if err != nil {
		return err
	}

	mode, err := step.OptionalStringEnum("mode", "random", "random", "static")
	if err != nil {
		return err
	}

	// "static" mode assigns an explicitly listed set of reviewers, so it is only
	// meaningful together with an explicit "static" user list.
	if mode == "static" && source != "static" {
		return fmt.Errorf("step field 'mode: static' is only supported with 'source: static'")
	}

	existingReviewers := evalContext.GetReviewers()

	// Look up which reviewers are already assigned. Both modes use this to avoid
	// re-adding people and to preserve the existing reviewers when updating, since
	// GitLab's "reviewer_ids" replaces the whole set on update.
	alreadyAssigned := make(map[int]struct{}, len(existingReviewers))
	for _, reviewer := range existingReviewers {
		if id := reviewer.IntID(); id != 0 {
			alreadyAssigned[id] = struct{}{}
		}
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

	switch mode {
	case "random":
		// Only the eligible reviewers that are not already assigned are candidates for
		// selection, and the ones already assigned count towards the limit. This "tops
		// up" reviewers from the list until the limit is satisfied, rather than skipping
		// entirely when reviewers already exist or endlessly adding on repeat runs.
		var candidates scm.Actors

		satisfied := 0

		for _, actor := range eligibleReviewers {
			if _, ok := alreadyAssigned[actor.IntID()]; ok {
				satisfied++

				continue
			}

			candidates = append(candidates, actor)
		}

		needed := desiredLimit - satisfied
		if needed > len(candidates) {
			needed = len(candidates)
		}

		if needed < 0 {
			needed = 0
		}

		reviewers = make(scm.Actors, needed)

		rand := state.RandomSeed(ctx)
		perm := rand.Perm(len(candidates))

		for i := 0; i < needed; i++ {
			reviewers[i] = candidates[perm[i]]
		}

		break
	case "static":
		// Assign every explicitly listed reviewer; "limit" is ignored in this mode.
		reviewers = append(scm.Actors{}, eligibleReviewers...)

		break
	}

	// Build the final reviewer set. GitLab's "reviewer_ids" replaces the whole set on
	// update, so any existing reviewers must be preserved to avoid removing them. Newly
	// selected reviewers are de-duplicated against the reviewers already present.
	reviewerIDs := make([]int, 0, len(existingReviewers)+len(reviewers))
	seen := make(map[int]struct{}, cap(reviewerIDs))

	// Preserve the reviewers already assigned; selected reviewers are added alongside them.
	for _, reviewer := range existingReviewers {
		id := reviewer.IntID()
		if id == 0 {
			continue
		}

		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		reviewerIDs = append(reviewerIDs, id)
	}

	added := 0

	for _, reviewer := range reviewers {
		id := reviewer.IntID()

		// skip invalid int ids, this should not happen but still safeguard against it
		if id == 0 {
			slogctx.Warn(ctx, "Invalid reviewer ID", slog.String("id", reviewer.ID))

			continue
		}

		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		reviewerIDs = append(reviewerIDs, id)
		added++
	}

	// If there are no new reviewers to add, skip the update to avoid needless MR churn.
	// This makes both modes idempotent across repeated evaluations.
	if added == 0 {
		slogctx.Debug(ctx, "No new reviewers to assign")

		return nil
	}

	if state.IsDryRun(ctx) {
		slogctx.Info(ctx, "(Dry Run) Assigning MR", slog.String("source", source), slog.Int("limit", desiredLimit), slog.String("mode", mode), slog.Any("reviewers", reviewers))

		return nil
	}

	update.AppendReviewerIDs(reviewerIDs)

	return nil
}
