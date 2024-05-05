package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/jippi/gitlab-labeller/pkg/config"
	"github.com/jippi/gitlab-labeller/pkg/scm"
	"github.com/jippi/gitlab-labeller/pkg/scm/gitlab"
	"github.com/jippi/gitlab-labeller/pkg/state"
	"github.com/urfave/cli/v2"
)

const (
	FlagConfigFile     = "config"
	FlagAPIToken       = "api-token"
	FlagSCMProject     = "project"
	FlagSCMBaseURL     = "base-url"
	FlagMergeRequestID = "id"
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
				Name:      FlagConfigFile,
				Usage:     "Path to the scm-engine config file",
				Value:     ".scm-engine.yml",
				TakesFile: true,
				EnvVars: []string{
					"SCM_ENGINE_CONFIG_FILE",
				},
			},
			&cli.StringFlag{
				Name:     FlagAPIToken,
				Usage:    "GitHub/GitLab API token",
				Required: true,
				EnvVars: []string{
					"GITLAB_TOKEN",
					"SCM_ENGINE_TOKEN",
				},
			},
			&cli.StringFlag{
				Name:     FlagSCMProject,
				Usage:    "GitLab project (example: 'gitlab-org/gitlab')",
				Required: true,
				EnvVars: []string{
					"GITLAB_PROJECT",
					"CI_PROJECT_PATH",
				},
			},
			&cli.StringFlag{
				Name:  FlagSCMBaseURL,
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
				Action: evaluateCmd,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  FlagMergeRequestID,
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
				Action: serverCmd,
			},
		},
	}

	origHelpPrinterCustom := cli.HelpPrinterCustom
	cli.HelpPrinterCustom = func(out io.Writer, templ string, data interface{}, customFuncs map[string]interface{}) {
		origHelpPrinterCustom(out, templ, data, customFuncs)
		if data != app {
			origHelpPrinterCustom(app.Writer, globalOptionsTemplate, app, nil)
		}
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func ProcessMR(ctx context.Context, cCtx *cli.Context, mr string) error {
	ctx = state.ContextWithMergeRequestID(ctx, mr)

	// for mr := 900; mr <= 1000; mr++ {
	fmt.Println("Processing MR", mr)

	cfg, err := config.LoadFile(cCtx.String(FlagConfigFile))
	if err != nil {
		return err
	}

	client, err := gitlab.NewClient(cCtx.String(FlagAPIToken), cCtx.String(FlagSCMBaseURL))
	if err != nil {
		return err
	}

	remoteLabels, err := client.Labels().List(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Creating evaluation context")
	evalContext, err := client.EvalContext(ctx)
	if err != nil {
		return err
	}

	if evalContext == nil {
		return nil
	}

	fmt.Println("Evaluating context")
	matches, err := cfg.Evaluate(evalContext)
	if err != nil {
		panic(err)
	}

	// spew.Dump(matches)

	// for _, label := range matches {
	// 	fmt.Println(label.Name, label.Matched, label.Color)
	// }

	fmt.Println("Sync labels")
	if err := sync(ctx, client, remoteLabels, matches); err != nil {
		panic(err)
	}
	fmt.Println("Done!")

	fmt.Println("Updating MR")
	if err := apply(ctx, client, matches); err != nil {
		panic(err)
	}
	fmt.Println("Done!")

	return nil
}

func apply(ctx context.Context, client scm.Client, remoteLabels []scm.EvaluationResult) error {
	var add scm.LabelOptions
	var remove scm.LabelOptions

	for _, e := range remoteLabels {
		if e.Matched {
			add = append(add, e.Name)
		} else {
			remove = append(remove, e.Name)
		}
	}

	_, err := client.MergeRequests().Update(ctx, &scm.UpdateMergeRequestOptions{
		AddLabels:    &add,
		RemoveLabels: &remove,
	})
	if err != nil {
		return err
	}

	return nil
}

func sync(ctx context.Context, client scm.Client, remote []*scm.Label, required []scm.EvaluationResult) error {
	fmt.Println("Going to sync", len(required), "required labels")

	remoteLabels := map[string]*scm.Label{}
	for _, e := range remote {
		remoteLabels[e.Name] = e
	}

	// Create
	for _, r := range required {
		if _, ok := remoteLabels[r.Name]; ok {
			continue
		}

		fmt.Print("Creating label ", r.Name, ": ")
		_, resp, err := client.Labels().Create(ctx, &scm.CreateLabelOptions{
			Name:        &r.Name,
			Color:       &r.Color,
			Description: &r.Description,
			Priority:    r.Priority,
		})
		if err != nil {
			// Label already exists
			if resp.StatusCode == http.StatusConflict {
				fmt.Println("Already exists!")

				continue
			}

			return err
		}

		fmt.Println("OK")
	}

	// Update
	for _, r := range required {
		e, ok := remoteLabels[r.Name]
		if !ok {
			continue
		}

		if r.EqualLabel(e) {
			continue
		}

		fmt.Print("Updating label ", r.Name, ": ")
		_, _, err := client.Labels().Update(ctx, &scm.UpdateLabelOptions{
			Name:        &r.Name,
			Color:       &r.Color,
			Description: &r.Description,
			Priority:    r.Priority,
		})
		if err != nil {
			return err
		}

		fmt.Println("OK")
	}

	return nil
}
