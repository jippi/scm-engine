package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/jippi/scm-engine/cmd"
	"github.com/jippi/scm-engine/pkg/tui"
	"github.com/urfave/cli/v2"
	slogctx "github.com/veqryn/slog-context"
)

// nolint: gochecknoglobals
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
			cCtx.Context = tui.NewContext(cCtx.Context, cCtx.App.Writer, cCtx.App.ErrWriter)
			cCtx.Context = slogctx.With(cCtx.Context, "scm_engine_version", version)

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
			&cli.StringFlag{
				Name:     cmd.FlagAPIToken,
				Usage:    "GitHub/GitLab API token",
				Required: true,
				EnvVars: []string{
					"SCM_ENGINE_TOKEN",
				},
			},
			&cli.StringFlag{
				Name:  cmd.FlagSCMBaseURL,
				Usage: "Base URL for the SCM instance",
				Value: "https://gitlab.com/",
				EnvVars: []string{
					"GITLAB_BASEURL",
					"CI_SERVER_URL",
				},
			},
			&cli.BoolFlag{
				Name:  cmd.FlagDryRun,
				Usage: "Dry run, don't actually _do_ actions, just print them",
				Value: false,
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "evaluate",
				Usage:     "Evaluate a Merge Request",
				Args:      true,
				ArgsUsage: " [mr_id, mr_id, ...]",
				Action:    cmd.Evaluate,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     cmd.FlagSCMProject,
						Usage:    "GitLab project (example: 'gitlab-org/gitlab')",
						Required: true,
						EnvVars: []string{
							"GITLAB_PROJECT",
							"CI_PROJECT_PATH",
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
					&cli.BoolFlag{
						Name:  cmd.FlagUpdatePipeline,
						Usage: "Update the CI pipeline status with progress",
						Value: false,
						EnvVars: []string{
							"SCM_ENGINE_UPDATE_PIPELINE",
						},
					},
				},
			},
			{
				Name:   "server",
				Usage:  "Start HTTP server for webhook event driven usage",
				Action: cmd.Server,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  cmd.FlagWebhookSecret,
						Usage: "Used to validate received payloads. Sent with the request in the X-Gitlab-Token HTTP header",
						EnvVars: []string{
							"SCM_ENGINE_WEBHOOK_SECRET",
						},
					},
					&cli.StringFlag{
						Name:  cmd.FlagServerListen,
						Usage: "IP + Port that the HTTP server should listen on",
						Value: "0.0.0.0:3000",
						EnvVars: []string{
							"SCM_ENGINE_LISTEN",
						},
					},
					&cli.BoolFlag{
						Name:  cmd.FlagUpdatePipeline,
						Usage: "Update the CI pipeline status with progress",
						Value: true,
						EnvVars: []string{
							"SCM_ENGINE_UPDATE_PIPELINE",
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
