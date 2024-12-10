package gitlab

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/datolabs-io/go-backstage/v3"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	slogctx "github.com/veqryn/slog-context"
)

func (c *Client) AssignReviewers(ctx context.Context, evalContext scm.EvalContext, update *scm.UpdateMergeRequestOptions, step scm.ActionStep) error {
	source, err := step.OptionalStringEnum("source", "codeowners", "codeowners", "backstage")
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

		backstageOwners, err := getBackstageOwners(ctx, c.backstage)
		if err != nil {
			return err
		}

		for _, owner := range backstageOwners {
			if evalContext.GetAuthor().ID != owner.ID {
				eligibleReviewers = append(eligibleReviewers, owner)
			}
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

// getBackstageOwners returns a list of GitLab-mapped Backstage owners for the
func getBackstageOwners(ctx context.Context, backstageClient *backstage.Client) ([]scm.Actor, error) {
	if backstageClient == nil {
		return nil, nil
	}

	projectID, err := ParseID(state.ProjectID(ctx))
	if err != nil {
		return nil, err
	}

	if len(projectID) == 0 {
		slogctx.Debug(ctx, "Empty project id")

		return nil, nil
	}

	// de-slugify the project ID
	projectParts := strings.Split(projectID, "/")
	project := projectParts[len(projectParts)-1]

	// search for the project in the Backstage catalog
	systems, entityResponse, err := backstageClient.Catalog.Entities.List(ctx, &backstage.ListEntityOptions{
		Filters: []string{
			"kind=system,metadata.name=" + project,
			"kind=system,metadata.annotations.gitlab.com/project-slug=" + projectID,
			"kind=system,metadata.annotations.gitlab.com/project=" + project,
		},
		Fields: []string{
			"spec.owner", // retrieve only the owner field
			"metadata.name",
		},
	})
	defer entityResponse.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("failed to search Backstage catalog: %w", err)
	}

	if len(systems) == 0 {
		slogctx.Debug(ctx, "No systems found in Backstage catalog")

		return nil, nil
	}

	system := systems[0]

	// search for the group that owns the system in Backstage
	groupRef := systems[0].Spec["owner"]
	if groupRef == nil {
		slogctx.Debug(ctx, "No owner found in Backstage catalog for system", slog.Any("backstage_system", system.Metadata.Name))

		return nil, nil
	}

	groupRefStr, ok := groupRef.(string) // add type assertion check
	if !ok {
		return nil, fmt.Errorf("owner field is not a string: %v", groupRef)
	}

	// search for users in backstage
	users, usersResponse, err := backstageClient.Catalog.Entities.List(ctx, &backstage.ListEntityOptions{
		Filters: []string{
			"kind=user,relations.memberof=" + groupRefStr,
		},
		Fields: []string{
			"metadata.annotations.gitlab.com/user_id",
		},
	})
	defer usersResponse.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("failed to search Backstage catalog for users: %w", err)
	}

	if len(users) == 0 {
		slogctx.Debug(ctx, "No users found in Backstage catalog for group", slog.String("backstage_group", groupRefStr))

		return nil, nil
	}

	var reviewers scm.Actors

	for _, user := range users {
		userID, ok := user.Metadata.Annotations["gitlab.com/user_id"]
		if !ok {
			continue
		}

		reviewers = append(reviewers, scm.Actor{
			ID: userID,
		})
	}

	return reviewers, nil
}
