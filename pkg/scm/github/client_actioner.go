package github

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	go_github "github.com/google/go-github/v64/github"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	slogctx "github.com/veqryn/slog-context"
)

func (c *Client) ApplyStep(ctx context.Context, evalContext scm.EvalContext, update *scm.UpdateMergeRequestOptions, step scm.EvaluationActionStep) error {
	owner, repo := ownerAndRepo(ctx)

	action, ok := step["action"]
	if !ok {
		return errors.New("step is missing an 'action' key")
	}

	actionString, ok := action.(string)
	if !ok {
		return fmt.Errorf("step field 'action' must be of type string, got %T", action)
	}

	switch actionString {
	case "add_label":
		name, ok := step["name"]
		if !ok {
			return errors.New("step field 'name' is required, but missing")
		}

		nameVal, ok := name.(string)
		if !ok {
			return errors.New("step field 'name' must be a string")
		}

		labels := update.AddLabels
		if labels == nil {
			labels = &scm.LabelOptions{}
		}

		tmp := append(*labels, nameVal)

		update.AddLabels = &tmp

	case "remove_label":
		name, ok := step["name"]
		if !ok {
			return errors.New("step field 'name' is required, but missing")
		}

		nameVal, ok := name.(string)
		if !ok {
			return errors.New("step field 'name' must be a string")
		}

		labels := update.RemoveLabels
		if labels == nil {
			labels = &scm.LabelOptions{}
		}

		tmp := append(*labels, nameVal)

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
			slogctx.Info(ctx, "Approving MR")

			return nil
		}

		_, _, err := c.wrapped.PullRequests.CreateReview(ctx, owner, repo, state.MergeRequestIDInt(ctx), &go_github.PullRequestReviewRequest{
			Event: scm.Ptr("APPROVE"),
		})

		return err

	case "unapprove":
		if state.IsDryRun(ctx) {
			slogctx.Info(ctx, "Unapproving MR")

			return nil
		}

		_, _, err := c.wrapped.PullRequests.CreateReview(ctx, owner, repo, state.MergeRequestIDInt(ctx), &go_github.PullRequestReviewRequest{})

		return err

	case "comment":
		msg, ok := step["message"]
		if !ok {
			return errors.New("step field 'message' is required, but missing")
		}

		msgString, ok := msg.(string)
		if !ok {
			return fmt.Errorf("step field 'message' must be a string, got %T", msg)
		}

		if len(msgString) == 0 {
			return errors.New("step field 'message' must not be an empty string")
		}

		if state.IsDryRun(ctx) {
			slogctx.Info(ctx, "Commenting on MR", slog.String("message", msgString))

			return nil
		}

		_, _, err := c.wrapped.PullRequests.CreateComment(ctx, owner, repo, state.MergeRequestIDInt(ctx), &go_github.PullRequestComment{
			Body: scm.Ptr(msgString),
		})

		return err

	default:
		return fmt.Errorf("GitLab client does not know how to apply action %q", step["action"])
	}

	return nil
}
