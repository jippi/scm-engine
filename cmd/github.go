package cmd

import (
	"context"

	"github.com/jippi/scm-engine/pkg/state"
	"github.com/urfave/cli/v3"
)

var GitHub = &cli.Command{
	Name:  "github",
	Usage: "GitHub related commands",
	Before: func(ctx context.Context, cCtx *cli.Command) (context.Context, error) {
		return state.WithProvider(ctx, "github"), nil
	},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  FlagAPIToken,
			Usage: "GitHub API token",
			Sources: cli.EnvVars(
				"SCM_ENGINE_TOKEN", // SCM Engine Native
			),
		},
		&cli.StringFlag{
			Name:  FlagSCMBaseURL,
			Usage: "Base URL for the SCM instance",
			Value: "https://api.github.com/",
			Sources: cli.EnvVars(
				"SCM_ENGINE_BASE_URL", // SCM Engine Native
			),
		},
	},
	Commands: []*cli.Command{
		{
			Name:      "evaluate",
			Usage:     "Evaluate a Pull Request",
			ArgsUsage: " [pr_id, pr_id, ...]",
			Action:    Evaluate,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     FlagSCMProject,
					Usage:    "GitHub project (example: 'jippi/scm-engine')",
					Required: true,
					Sources: cli.EnvVars(
						"GITHUB_REPOSITORY", // GitHub Actions CI
					),
				},
				&cli.StringFlag{
					Name:  FlagMergeRequestID,
					Usage: "The Pull Request ID to process, if not provided as a CLI flag",
					Sources: cli.EnvVars(
						"SCM_ENGINE_PULL_REQUEST_ID", // SCM Engine native
					),
				},
				&cli.StringFlag{
					Name:  FlagCommitSHA,
					Usage: "The git commit sha",
					Sources: cli.EnvVars(
						"GITHUB_SHA", // GitHub Actions
					),
				},
			},
		},
	},
}
