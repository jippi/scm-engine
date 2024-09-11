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

func (c *Client) ApplyStep(ctx context.Context, evalContext scm.EvalContext, update *scm.UpdateMergeRequestOptions, step scm.ActionStep) error {
	owner, repo := ownerAndRepo(ctx)

	action, err := step.RequiredString("action")
	if err != nil {
		return err
	}

	switch action {
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
		msg, err := step.RequiredString("message")
		if err != nil {
			return err
		}

		if len(msg) == 0 {
			return errors.New("step field 'message' must not be an empty string")
		}

		if state.IsDryRun(ctx) {
			slogctx.Info(ctx, "Commenting on MR", slog.String("message", msg))

			return nil
		}

		_, _, err = c.wrapped.PullRequests.CreateComment(ctx, owner, repo, state.MergeRequestIDInt(ctx), &go_github.PullRequestComment{
			Body: scm.Ptr(msg),
		})

		return err

	default:
		return fmt.Errorf("GitLab client does not know how to apply action %q", action)
	}

	return nil
}
