package config

import (
	"context"
	"fmt"
	"log/slog"

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

	// Update logger with a friendly tag to differentiate the events within
	ctx = slogctx.With(ctx, slog.String("phase", "remote_include"))

	// For each project, do a read of all the files we need
	for _, include := range c.Includes {
		ctx := slogctx.With(ctx, slog.Any("remote_include_config", include))

		slogctx.Debug(ctx, fmt.Sprintf("Loading remote configuration from project %q", include.Project))

		files, err := client.GetProjectFiles(ctx, include.Project, include.Ref, include.Files)
		if err != nil {
			return fmt.Errorf("failed to load included config files from project [%s]: %w", include.Project, err)
		}

		for fileName, fileContent := range files {
			remoteConfig, err := ParseFileString(fileContent)
			if err != nil {
				return fmt.Errorf("failed to parse remote config file [%s] from project [%s]: %w", fileName, include.Project, err)
			}

			// Disallow nested includes
			if len(remoteConfig.Includes) != 0 {
				slogctx.Warn(ctx, fmt.Sprintf("file [%s] from project [%s] may not have any 'include' settings; Recursive include is not supported", fileName, include.Project))
			}

			// Disallow changing dry run
			if remoteConfig.DryRun != nil {
				slogctx.Warn(ctx, fmt.Sprintf("file [%s] from project [%s] may not have a 'dry_run' setting; Remote include are not allowed to change this setting", fileName, include.Project))
			}

			// Append actions
			if len(remoteConfig.Actions) != 0 {
				slogctx.Debug(ctx, fmt.Sprintf("file [%s] from project [%s] added %d new actions to the config file", fileName, include.Project, len(remoteConfig.Actions)))

				c.Actions = append(c.Actions, remoteConfig.Actions...)
			}

			// Append labels
			if len(remoteConfig.Labels) != 0 {
				slogctx.Debug(ctx, fmt.Sprintf("file [%s] from project [%s] added %d new labels to the config file", fileName, include.Project, len(remoteConfig.Labels)))

				c.Labels = append(c.Labels, remoteConfig.Labels...)
			}
		}
	}

	slogctx.Debug(ctx, "Done loading remote configuration files")

	return nil
}
