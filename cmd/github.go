package cmd

import (
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/urfave/cli/v3"
)

var GitHub = &cli.Command{
	Name:  "github",
	Usage: "GitHub related commands",
	Before: func(ctx *cli.Context) error {
		ctx.Context = state.WithProvider(ctx.Context, "github")

		return nil
	},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  FlagAPIToken,
			Usage: "GitHub API token",
			EnvVars: []string{
				"SCM_ENGINE_TOKEN", // SCM Engine Native
			},
		},
		&cli.StringFlag{
			Name:  FlagSCMBaseURL,
			Usage: "Base URL for the SCM instance",
			Value: "https://api.github.com/",
			EnvVars: []string{
				"SCM_ENGINE_BASE_URL", // SCM Engine Native
			},
		},
	},
	Subcommands: []*cli.Command{
		{
			Name:      "evaluate",
			Usage:     "Evaluate a Pull Request",
			Args:      true,
			ArgsUsage: " [pr_id, pr_id, ...]",
			Action:    Evaluate,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     FlagSCMProject,
					Usage:    "GitHub project (example: 'jippi/scm-engine')",
					Required: true,
					EnvVars: []string{
						"GITHUB_REPOSITORY", // GitHub Actions CI
					},
				},
				&cli.StringFlag{
					Name:  FlagMergeRequestID,
					Usage: "The Pull Request ID to process, if not provided as a CLI flag",
					EnvVars: []string{
						"SCM_ENGINE_PULL_REQUEST_ID", // SCM Engine native
					},
				},
				&cli.StringFlag{
					Name:  FlagCommitSHA,
					Usage: "The git commit sha",
					EnvVars: []string{
						"GITHUB_SHA", // GitHub Actions
					},
				},
			},
		},
	},
}
