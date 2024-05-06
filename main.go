package main

import (
	"io"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/jippi/scm-engine/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	spew.Config.DisableMethods = true

	app := &cli.App{
		Name:                 "scm-engine",
		Usage:                "GitHub/GitLab automation",
		Copyright:            "Christian Winther",
		EnableBashCompletion: true,
		Suggest:              true,
		Authors: []*cli.Author{
			{
				Name:  "Christian Winther",
				Email: "gitlab-engine@jippi.dev",
			},
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
				Name:     cmd.FlagSCMProject,
				Usage:    "GitLab project (example: 'gitlab-org/gitlab')",
				Required: true,
				EnvVars: []string{
					"GITLAB_PROJECT",
					"CI_PROJECT_PATH",
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
		},
		Commands: []*cli.Command{
			{
				Name:   "evaluate",
				Usage:  "Evaluate a Merge Request",
				Action: cmd.Evaluate,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  cmd.FlagMergeRequestID,
						Usage: "The pull/merge to process, if not provided as a CLI flag",
						Aliases: []string{
							"merge-request-id", // GitLab naming
							"pull-request-id",  // GitHub naming
						},
						EnvVars: []string{
							"CI_MERGE_REQUEST_IID", // GitLab CI
						},
					},
				},
			},
			{
				Name:   "server",
				Usage:  "Start HTTP server for webhook event driven usage",
				Action: cmd.Server,
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
