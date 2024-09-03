package config

import (
	"context"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/jippi/scm-engine/pkg/scm"
	slogctx "github.com/veqryn/slog-context"
)

type Config struct {
	DryRun             *bool              `yaml:"dry_run"`
	Actions            Actions            `yaml:"actions"`
	IgnoreActivityFrom IgnoreActivityFrom `yaml:"ignore_activity_from"`
	Includes           []Include          `yaml:"include"`
	Labels             Labels             `yaml:"label"`
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

func (c *Config) LoadIncludes(ctx context.Context, client scm.Client) error {
	// No files to include
	if len(c.Includes) == 0 {
		return nil
	}

	for _, include := range c.Includes {
		blobs, err := client.GetProjectFiles(ctx, include.Project, include.Ref, include.Files)
		if err != nil {
			return fmt.Errorf("failed to load included config files from project [%s]: %w", include.Project, err)
		}

		spew.Dump(blobs)
	}

	return nil
}
