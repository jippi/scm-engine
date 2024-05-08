package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
)

func ProcessMR(ctx context.Context, client scm.Client, cfg *config.Config, mr string, event any) error {
	ctx = state.ContextWithMergeRequestID(ctx, mr)

	// for mr := 900; mr <= 1000; mr++ {
	fmt.Println("Processing MR", mr)

	remoteLabels, err := client.Labels().List(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Creating evaluation context")

	evalContext, err := client.EvalContext(ctx)
	if err != nil {
		return err
	}

	if evalContext == nil || !evalContext.IsValid() {
		fmt.Println("Evaluating context is empty, does the Merge Request exists?")

		return nil
	}

	evalContext.SetWebhookEvent(event)

	fmt.Println("Evaluating context")

	labels, actions, err := cfg.Evaluate(evalContext)
	if err != nil {
		return err
	}

	fmt.Println("Sync labels")

	if err := syncLabels(ctx, client, remoteLabels, labels); err != nil {
		return err
	}

	fmt.Println("Done!")

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

	fmt.Println("Applying actions")

	if err := runActions(ctx, client, update, actions); err != nil {
		return err
	}

	fmt.Println("Done!")

	fmt.Println("Updating MR")

	if err := updateMergeRequest(ctx, client, update); err != nil {
		return err
	}

	fmt.Println("Done!")

	return nil
}

func updateMergeRequest(ctx context.Context, client scm.Client, update *scm.UpdateMergeRequestOptions) error {
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
	fmt.Println("Going to sync", len(required), "required labels")

	remoteLabels := map[string]*scm.Label{}
	for _, e := range remote {
		remoteLabels[e.Name] = e
	}

	// Create
	for _, label := range required {
		if _, ok := remoteLabels[label.Name]; ok {
			continue
		}

		fmt.Print("Creating label ", label.Name, ": ")

		_, resp, err := client.Labels().Create(ctx, &scm.CreateLabelOptions{
			Name:        &label.Name,        //nolint:gosec
			Color:       &label.Color,       //nolint:gosec
			Description: &label.Description, //nolint:gosec
			Priority:    label.Priority,
		})
		if err != nil {
			// Label already exists
			if resp.StatusCode == http.StatusConflict {
				fmt.Println("Already exists!")

				continue
			}

			return err
		}

		fmt.Println("OK")
	}

	// Update
	for _, label := range required {
		e, ok := remoteLabels[label.Name]
		if !ok {
			continue
		}

		if label.IsEqual(e) {
			continue
		}

		fmt.Print("Updating label ", label.Name, ": ")

		_, _, err := client.Labels().Update(ctx, &scm.UpdateLabelOptions{
			Name:        &label.Name,        //nolint:gosec
			Color:       &label.Color,       //nolint:gosec
			Description: &label.Description, //nolint:gosec
			Priority:    label.Priority,
		})
		if err != nil {
			return err
		}

		fmt.Println("OK")
	}

	return nil
}
