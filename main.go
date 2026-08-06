//go:build !generate
// +build !generate

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/jippi/scm-engine/cmd"
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/jippi/scm-engine/pkg/tui"
	"github.com/urfave/cli/v3"
	slogctx "github.com/veqryn/slog-context"
)

//nolint:gochecknoglobals
var (
	commit  = "unknown"
	date    = "unknown"
	version = "dev"
)

func main() {
	spew.Config.DisableMethods = true

	app := &cli.Command{
		Name:                  "scm-engine",
		Usage:                 "GitHub/GitLab automation",
		Copyright:             "Christian Winther",
		EnableShellCompletion: true,
		Suggest:               true,
		Version:               fmt.Sprintf("%s (date: %s; commit: %s)", version, date, commit),
		// NOTE: cli/v3 renders authors with fmt, and mail.Address quotes any name
		// containing a space. A plain string keeps the v2 rendering.
		Authors: []any{
			"Christian Winther <scm-engine@jippi.dev>",
		},
		Before: func(ctx context.Context, cCtx *cli.Command) (context.Context, error) {
			// Setup global state
			ctx = tui.NewContext(ctx, cCtx.Root().Writer, cCtx.Root().ErrWriter)
			ctx = slogctx.With(ctx, "scm_engine_version", version)

			// Write global flags to context
			ctx = state.WithDryRun(ctx, cCtx.Bool(cmd.FlagDryRun))
			ctx = state.WithRandomSeed(ctx, time.Now().UnixNano()) // weak seed since only used for codeowner selection

			return ctx, nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      cmd.FlagConfigFile,
				Usage:     "Path to the scm-engine config file",
				Value:     ".scm-engine.yml",
				TakesFile: true,
				Sources:   cli.EnvVars("SCM_ENGINE_CONFIG_FILE"),
			},
			&cli.BoolFlag{
				Name:    cmd.FlagDryRun,
				Usage:   "Dry run, don't actually _do_ actions, just print them",
				Value:   false,
				Sources: cli.EnvVars("SCM_ENGINE_DRY_RUN"),
			},
		},
		Commands: []*cli.Command{
			cmd.GitLab,
			cmd.GitHub,

			// DEPRECATED COMMANDS
			{
				Name:      "evaluate",
				Usage:     "Evaluate a Merge Request",
				Hidden:    true, // DEPRECATED
				ArgsUsage: " [mr_id, mr_id, ...]",
				Action:    cmd.Evaluate,
				Before: func(ctx context.Context, cCtx *cli.Command) (context.Context, error) {
					ctx = state.WithBaseURL(ctx, cCtx.String(cmd.FlagSCMBaseURL))
					ctx = state.WithProvider(ctx, "gitlab")
					ctx = state.WithToken(ctx, cCtx.String(cmd.FlagAPIToken))

					return ctx, nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  cmd.FlagAPIToken,
						Usage: "GitLab API token",
						Sources: cli.EnvVars(
							"SCM_ENGINE_TOKEN", // SCM Engine Native
						),
					},
					&cli.StringFlag{
						Name:  cmd.FlagSCMBaseURL,
						Usage: "Base URL for the SCM instance",
						Value: "https://gitlab.com/",
						Sources: cli.EnvVars(
							"SCM_ENGINE_BASE_URL", // SCM Engine Native
							"CI_SERVER_URL",       // GitLab CI
						),
					},
					&cli.BoolFlag{
						Name:    cmd.FlagUpdatePipeline,
						Usage:   "Update the CI pipeline status with progress",
						Value:   false,
						Sources: cli.EnvVars("SCM_ENGINE_UPDATE_PIPELINE"),
					},
					&cli.StringFlag{
						Name:     cmd.FlagSCMProject,
						Usage:    "GitLab project (example: 'gitlab-org/gitlab')",
						Required: true,
						Sources: cli.EnvVars(
							"GITLAB_PROJECT",
							"CI_PROJECT_PATH", // GitLab CI
						),
					},
					&cli.StringFlag{
						Name:  cmd.FlagMergeRequestID,
						Usage: "The pull/merge ID to process, if not provided as a CLI flag",
						Sources: cli.EnvVars(
							"CI_MERGE_REQUEST_IID", // GitLab CI
						),
					},
					&cli.StringFlag{
						Name:  cmd.FlagCommitSHA,
						Usage: "The git commit sha",
						Sources: cli.EnvVars(
							"CI_COMMIT_SHA", // GitLab CI
						),
					},
				},
			},
		},
	}

	// NOTE: cli/v2 did not print global options on subcommand help, so we used
	// to wrap HelpPrinterCustom to append them. cli/v3 does this natively.

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
