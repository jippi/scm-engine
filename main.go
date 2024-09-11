//go:build !generate
// +build !generate

package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/jippi/scm-engine/cmd"
	"github.com/jippi/scm-engine/pkg/state"
	"github.com/jippi/scm-engine/pkg/tui"
	"github.com/urfave/cli/v2"
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

	app := &cli.App{
		Name:                 "scm-engine",
		Usage:                "GitHub/GitLab automation",
		Copyright:            "Christian Winther",
		EnableBashCompletion: true,
		Suggest:              true,
		Version:              fmt.Sprintf("%s (date: %s; commit: %s)", version, date, commit),
		Authors: []*cli.Author{
			{
				Name:  "Christian Winther",
				Email: "scm-engine@jippi.dev",
			},
		},
		Before: func(cCtx *cli.Context) error {
			// Setup global state
			cCtx.Context = tui.NewContext(cCtx.Context, cCtx.App.Writer, cCtx.App.ErrWriter)
			cCtx.Context = slogctx.With(cCtx.Context, "scm_engine_version", version)

			// Write global flags to context
			cCtx.Context = state.WithDryRun(cCtx.Context, cCtx.Bool(cmd.FlagDryRun))

			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      cmd.FlagConfigFile,
				Usage:     "Path to the scm-engine config file",
				Value:     ".scm-engine.yml",
				TakesFile: true,
				EnvVars: []string{
					"SCM_ENGINE_CONFIG_FILE",
				},
			},
			&cli.BoolFlag{
				Name:  cmd.FlagDryRun,
				Usage: "Dry run, don't actually _do_ actions, just print them",
				Value: false,
				EnvVars: []string{
					"SCM_ENGINE_DRY_RUN",
				},
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
				Args:      true,
				ArgsUsage: " [mr_id, mr_id, ...]",
				Action:    cmd.Evaluate,
				Before: func(cCtx *cli.Context) error {
					cCtx.Context = state.WithBaseURL(cCtx.Context, cCtx.String(cmd.FlagSCMBaseURL))
					cCtx.Context = state.WithProvider(cCtx.Context, "gitlab")
					cCtx.Context = state.WithToken(cCtx.Context, cCtx.String(cmd.FlagAPIToken))

					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  cmd.FlagAPIToken,
						Usage: "GitLab API token",
						EnvVars: []string{
							"SCM_ENGINE_TOKEN", // SCM Engine Native
						},
					},
					&cli.StringFlag{
						Name:  cmd.FlagSCMBaseURL,
						Usage: "Base URL for the SCM instance",
						Value: "https://gitlab.com/",
						EnvVars: []string{
							"SCM_ENGINE_BASE_URL", // SCM Engine Native
							"CI_SERVER_URL",       // GitLab CI
						},
					},
					&cli.BoolFlag{
						Name:  cmd.FlagUpdatePipeline,
						Usage: "Update the CI pipeline status with progress",
						Value: false,
						EnvVars: []string{
							"SCM_ENGINE_UPDATE_PIPELINE",
						},
					},
					&cli.StringFlag{
						Name:     cmd.FlagSCMProject,
						Usage:    "GitLab project (example: 'gitlab-org/gitlab')",
						Required: true,
						EnvVars: []string{
							"GITLAB_PROJECT",
							"CI_PROJECT_PATH", // GitLab CI
						},
					},
					&cli.StringFlag{
						Name:  cmd.FlagMergeRequestID,
						Usage: "The pull/merge ID to process, if not provided as a CLI flag",
						EnvVars: []string{
							"CI_MERGE_REQUEST_IID", // GitLab CI
						},
					},
					&cli.StringFlag{
						Name:  cmd.FlagCommitSHA,
						Usage: "The git commit sha",
						EnvVars: []string{
							"CI_COMMIT_SHA", // GitLab CI
						},
					},
				},
			},
		},
	}

	origHelpPrinterCustom := cli.HelpPrinterCustom
	cli.HelpPrinterCustom = func(out io.Writer, templ string, data interface{}, customFuncs map[string]interface{}) {
		origHelpPrinterCustom(out, templ, data, customFuncs)

		if data != app {
			origHelpPrinterCustom(app.Writer, cmd.GlobalOptionsTemplate, app, nil)
		}
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
