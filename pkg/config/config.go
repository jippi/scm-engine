package config

import (
	"context"
	"fmt"

	"github.com/jippi/scm-engine/pkg/scm"
	slogctx "github.com/veqryn/slog-context"
)

type Config struct {
	DryRun             *bool              `yaml:"dry_run"`
	Labels             Labels             `yaml:"label"`
	Actions            Actions            `yaml:"actions"`
	IgnoreActivityFrom IgnoreActivityFrom `yaml:"ignore_activity_from"`
}

func (c Config) Evaluate(ctx context.Context, evalContext scm.EvalContext) ([]scm.EvaluationResult, []Action, error) {
	slogctx.Info(ctx, "Evaluating labels")

	labels, err := c.Labels.Evaluate(ctx, evalContext)
	if err != nil {
		return nil, nil, fmt.Errorf("evaluation failed: %w", err)
	}

	slogctx.Info(ctx, "Evaluating Actions")

	actions, err := c.Actions.Evaluate(ctx, evalContext)
	if err != nil {
		return nil, nil, err
	}

	return labels, actions, nil
}
