package gitlab

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/patcher"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/jippi/scm-engine/pkg/stdlib"
	slogctx "github.com/veqryn/slog-context"
	"github.com/xanzy/go-gitlab"
)

func (c *Client) ApplyStep(ctx context.Context, evalContext scm.EvalContext, update *scm.UpdateMergeRequestOptions, step scm.EvaluationActionStep) error {
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

		replacements, ok := step["replace"]
		if !ok {
			return errors.New("step field 'replace' is required, but missing")
		}

		replacementSlice, ok := replacements.(scm.EvaluationActionStep)
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
		return fmt.Errorf("GitLab client does not know how to apply action %q", step["action"])
	}

	return nil
}
