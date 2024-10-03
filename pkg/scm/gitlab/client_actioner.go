package gitlab

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/patcher"
	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/jippi/scm-engine/pkg/stdlib"
	slogctx "github.com/veqryn/slog-context"
	"github.com/xanzy/go-gitlab"
)

var randSource = rand.New(rand.NewSource(time.Now().UnixNano()))

func (c *Client) ApplyStep(ctx context.Context, evalContext scm.EvalContext, update *scm.UpdateMergeRequestOptions, step scm.ActionStep) error {
	action, err := step.RequiredString("action")
	if err != nil {
		return err
	}

	switch action {
	case "update_description":
		// Use the raw MR description
		body := evalContext.GetDescription()

		// Unless something else already updated the description in the Update struct
		if update.Description != nil {
			body = *update.Description
		}

		replacements, err := step.Get("replace")
		if err != nil {
			return err
		}

		replacementSlice, ok := replacements.(config.ActionStep)
		if !ok {
			return fmt.Errorf(`step field 'replace' must be a dictionary with string key and string values ("key": "value"), got: %T`, replacements)
		}

		replacedAnything := false

		for key, script := range replacementSlice {
			// If the replacement key do not exist; we can skip the replacement logic entirely!
			if !strings.Contains(body, key) {
				continue
			}

			replacedAnything = true

			// Build the ExprLang VM program
			// TODO(jippi): make this something generic/shared somewhere more central so we keep settings in sync
			opts := []expr.Option{}
			opts = append(opts, expr.AsKind(reflect.TypeFor[string]().Kind()))
			opts = append(opts, expr.Env(evalContext))
			opts = append(opts, stdlib.FunctionRenamer)
			opts = append(opts, stdlib.Functions...)
			opts = append(opts, expr.Patch(patcher.WithContext{Name: "ctx"}))

			program, err := expr.Compile(fmt.Sprintf("%s", script), opts...)
			if err != nil {
				return fmt.Errorf("could not evaluate value for 'replace' key '%s': %w", key, err)
			}

			output, err := expr.Run(program, evalContext)
			if err != nil {
				return err
			}

			switch val := output.(type) {
			case string:
				body = strings.ReplaceAll(body, key, val)

			default:
				return fmt.Errorf("'replace' value for key '%s' did not return a string: %w", key, err)
			}
		}

		// Don't update the body if there were no replacements
		if !replacedAnything {
			return nil
		}

		update.Description = &body

	case "add_label":
		name, err := step.RequiredString("name")
		if err != nil {
			return err
		}

		labels := update.AddLabels
		if labels == nil {
			labels = &scm.LabelOptions{}
		}

		tmp := append(*labels, name)

		update.AddLabels = &tmp

	case "remove_label":
		name, err := step.RequiredString("name")
		if err != nil {
			return err
		}

		labels := update.RemoveLabels
		if labels == nil {
			labels = &scm.LabelOptions{}
		}

		tmp := append(*labels, name)

		update.AddLabels = &tmp

	case "close":
		update.StateEvent = scm.Ptr("close")

	case "reopen":
		update.StateEvent = scm.Ptr("reopen")

	case "lock_discussion":
		update.DiscussionLocked = scm.Ptr(true)

	case "unlock_discussion":
		update.DiscussionLocked = scm.Ptr(false)

	case "approve":
		if state.IsDryRun(ctx) {
			slogctx.Info(ctx, "(Dry Run) Approving MR")

			return nil
		}

		_, _, err := c.wrapped.MergeRequestApprovals.ApproveMergeRequest(state.ProjectID(ctx), state.MergeRequestIDInt(ctx), &gitlab.ApproveMergeRequestOptions{})

		return err

	case "unapprove":
		if state.IsDryRun(ctx) {
			slogctx.Info(ctx, "(Dry Run) Unapproving MR")

			return nil
		}

		_, err := c.wrapped.MergeRequestApprovals.UnapproveMergeRequest(state.ProjectID(ctx), state.MergeRequestIDInt(ctx))

		return err

	case "assign_reviewers":
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

		var reviewers []scm.Actor

		limit = desiredLimit
		if limit > len(eligibleReviewers) {
			limit = len(eligibleReviewers)
		} 

		switch mode {
		case "linear":
			reviewers = eligibleReviewers[:limit]

			break
		case "random":
			reviewers = make([]scm.Actor, limit)
			perm := randSource.Perm(len(eligibleReviewers))

			for i := 0; i < limit; i++ {
				reviewers[i] = eligibleReviewers[perm[i]]
			}

			break
		}

		var reviewerIDs []int

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

		update := &scm.UpdateMergeRequestOptions{
			ReviewerIDs: &reviewerIDs,
		}

		_, err = c.MergeRequests().Update(ctx, update)

		return err

	case "comment":
		message, err := step.RequiredString("message")
		if err != nil {
			return err
		}

		if len(message) == 0 {
			return errors.New("step field 'message' must not be an empty string")
		}

		if state.IsDryRun(ctx) {
			slogctx.Info(ctx, "(Dry Run) Commenting on MR", slog.String("message", message))

			return nil
		}

		_, _, err = c.wrapped.Notes.CreateMergeRequestNote(state.ProjectID(ctx), state.MergeRequestIDInt(ctx), &gitlab.CreateMergeRequestNoteOptions{
			Body: scm.Ptr(message),
		})

		return err

	default:
		return fmt.Errorf("GitLab client does not know how to apply action %q", action)
	}

	return nil
}
