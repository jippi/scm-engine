package gitlab

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	slogctx "github.com/veqryn/slog-context"
	"github.com/xanzy/go-gitlab"
)

func (c *Client) ApplyStep(ctx context.Context, update *scm.UpdateMergeRequestOptions, step scm.EvaluationActionStep) error {
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
		update.StateEvent = gitlab.Ptr("close")

	case "reopen":
		update.StateEvent = gitlab.Ptr("reopen")

	case "lock_discussion":
		update.DiscussionLocked = gitlab.Ptr(true)

	case "unlock_discussion":
		update.DiscussionLocked = gitlab.Ptr(false)

	case "approve":
		if state.IsDryRun(ctx) {
			slogctx.Info(ctx, "Approving MR")

			return nil
		}

		_, _, err := c.wrapped.MergeRequestApprovals.ApproveMergeRequest(state.ProjectID(ctx), state.MergeRequestIDInt(ctx), &gitlab.ApproveMergeRequestOptions{})

		return err

	case "unapprove":
		if state.IsDryRun(ctx) {
			slogctx.Info(ctx, "Unapproving MR")

			return nil
		}

		_, err := c.wrapped.MergeRequestApprovals.UnapproveMergeRequest(state.ProjectID(ctx), state.MergeRequestIDInt(ctx))

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

		_, _, err := c.wrapped.Notes.CreateMergeRequestNote(state.ProjectID(ctx), state.MergeRequestIDInt(ctx), &gitlab.CreateMergeRequestNoteOptions{
			Body: gitlab.Ptr(msgString),
		})

		return err

	default:
		return fmt.Errorf("GitLab client does not know how to apply action %q", step["action"])
	}

	return nil
}
