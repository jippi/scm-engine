package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/scm/github"
	"github.com/jippi/scm-engine/pkg/scm/gitlab"
	"github.com/jippi/scm-engine/pkg/state"
	slogctx "github.com/veqryn/slog-context"
)

func getClient(ctx context.Context) (scm.Client, error) {
	switch state.Provider(ctx) {
	case "github":
		return github.NewClient(ctx), nil

	case "gitlab":
		return gitlab.NewClient(ctx)

	default:
		return nil, fmt.Errorf("unknown provider %q - we only support 'github' and 'gitlab'", state.Provider(ctx))
	}
}

func ProcessMR(ctx context.Context, client scm.Client, cfg *config.Config, event any) (err error) {
	// Stop the pipeline when we leave this func
	defer func() {
		if stopErr := client.Stop(ctx, err); stopErr != nil {
			slogctx.Error(ctx, "Failed to update pipeline", slog.Any("error", stopErr))
		}
	}()

	// Start the pipeline
	if err = client.Start(ctx); err != nil {
		return fmt.Errorf("failed to update pipeline monitor: %w", err)
	}

	slogctx.Info(ctx, "Processing MR")

	remoteLabels, err := client.Labels().List(ctx)
	if err != nil {
		return err
	}

	slogctx.Info(ctx, "Creating evaluation context")

	evalContext, err := client.EvalContext(ctx)
	if err != nil {
		return err
	}

	if evalContext == nil || !evalContext.IsValid() {
		slogctx.Warn(ctx, "Evaluating context is empty, does the Merge Request exists?")

		return nil
	}

	evalContext.SetWebhookEvent(event)

	slogctx.Info(ctx, "Evaluating context")

	labels, actions, err := cfg.Evaluate(ctx, evalContext)
	if err != nil {
		return err
	}

	slogctx.Info(ctx, "Sync labels")

	if err := syncLabels(ctx, client, remoteLabels, labels); err != nil {
		return err
	}

	var (
		add    scm.LabelOptions
		remove scm.LabelOptions
	)

	for _, e := range labels {
		if e.Matched {
			add = append(add, e.Name)
		} else {
			remove = append(remove, e.Name)
		}
	}

	update := &scm.UpdateMergeRequestOptions{
		AddLabels:    &add,
		RemoveLabels: &remove,
	}

	slogctx.Info(ctx, "Applying actions")

	if err := runActions(ctx, client, update, actions); err != nil {
		return err
	}

	slogctx.Info(ctx, "Updating MR")

	if err := updateMergeRequest(ctx, client, update); err != nil {
		return err
	}

	return nil
}

func updateMergeRequest(ctx context.Context, client scm.Client, update *scm.UpdateMergeRequestOptions) error {
	if state.IsDryRun(ctx) {
		slogctx.Info(ctx, "In dry-run, dumping the update struct we would send to GitLab", slog.Any("changes", update))

		return nil
	}

	_, err := client.MergeRequests().Update(ctx, update)

	return err
}

func runActions(ctx context.Context, client scm.Client, update *scm.UpdateMergeRequestOptions, actions []config.Action) error {
	for _, action := range actions {
		for _, task := range action.Then {
			if err := client.ApplyStep(ctx, update, task); err != nil {
				return err
			}
		}
	}

	return nil
}

func syncLabels(ctx context.Context, client scm.Client, remote []*scm.Label, required []scm.EvaluationResult) error {
	slogctx.Info(ctx, "Going to sync required labels", slog.Int("number_of_labels", len(required)))

	remoteLabels := map[string]*scm.Label{}
	for _, e := range remote {
		remoteLabels[e.Name] = e
	}

	// Create
	for _, label := range required {
		if _, ok := remoteLabels[label.Name]; ok {
			continue
		}

		slogctx.Info(ctx, "Creating label", slog.String("label", label.Name))

		if state.IsDryRun(ctx) {
			continue
		}

		_, resp, err := client.Labels().Create(ctx, &scm.CreateLabelOptions{
			Name:        &label.Name,        //nolint:gosec
			Color:       &label.Color,       //nolint:gosec
			Description: &label.Description, //nolint:gosec
			Priority:    label.Priority,
		})
		if err != nil {
			// Label already exists
			if resp.StatusCode == http.StatusConflict {
				slogctx.Warn(ctx, "Label already exists", slog.String("label", label.Name))

				continue
			}

			return err
		}
	}

	// Update
	for _, label := range required {
		remote, ok := remoteLabels[label.Name]
		if !ok {
			continue
		}

		if label.IsEqual(ctx, remote) {
			continue
		}

		slogctx.Info(ctx, "Updating label", slog.String("label", label.Name))

		if state.IsDryRun(ctx) {
			continue
		}

		_, _, err := client.Labels().Update(ctx, &scm.UpdateLabelOptions{
			Name:        &label.Name,        //nolint:gosec
			Color:       &label.Color,       //nolint:gosec
			Description: &label.Description, //nolint:gosec
			Priority:    label.Priority,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
