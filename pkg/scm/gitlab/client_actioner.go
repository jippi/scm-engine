package gitlab

import (
	"context"
	"errors"
	"fmt"

	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
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
	case "close":
		update.StateEvent = gitlab.Ptr("close")

	case "reopen":
		update.StateEvent = gitlab.Ptr("reopen")

	case "lock_discussion":
		update.DiscussionLocked = gitlab.Ptr(true)

	case "unlock_discussion":
		update.DiscussionLocked = gitlab.Ptr(false)

	case "approve":
		_, _, err := c.wrapped.MergeRequestApprovals.ApproveMergeRequest(state.ProjectIDFromContext(ctx), state.MergeRequestIDFromContextInt(ctx), &gitlab.ApproveMergeRequestOptions{})

		return err

	case "unapprove":
		_, err := c.wrapped.MergeRequestApprovals.UnapproveMergeRequest(state.ProjectIDFromContext(ctx), state.MergeRequestIDFromContextInt(ctx))

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

		_, _, err := c.wrapped.Notes.CreateMergeRequestNote(state.ProjectIDFromContext(ctx), state.MergeRequestIDFromContextInt(ctx), &gitlab.CreateMergeRequestNoteOptions{
			Body: gitlab.Ptr(msgString),
		})

		return err

	default:
		return fmt.Errorf("GitLab client does not know how to apply action %q", step["action"])
	}

	return nil
}
