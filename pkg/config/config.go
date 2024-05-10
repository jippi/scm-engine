package config

import (
	"context"
	"fmt"

	"github.com/jippi/scm-engine/pkg/scm"
	slogctx "github.com/veqryn/slog-context"
)

type Config struct {
	Labels  Labels  `yaml:"label"`
	Actions Actions `yaml:"actions"`
}

func (c Config) Evaluate(ctx context.Context, evalContext scm.EvalContext) ([]scm.EvaluationResult, []Action, error) {
	slogctx.Info(ctx, "Evaluating labels")

	labels, err := c.Labels.Evaluate(evalContext)
	if err != nil {
		return nil, nil, fmt.Errorf("evaluation failed: %w", err)
	}

	slogctx.Info(ctx, "Evaluating Actions")

	actions, err := c.Actions.Evaluate(evalContext)
	if err != nil {
		return nil, nil, err
	}

	return labels, actions, nil
}
