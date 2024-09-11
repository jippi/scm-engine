package cmd

import (
	"time"

	"github.com/jippi/scm-engine/pkg/state"
	"github.com/urfave/cli/v2"
)

var GitLab = &cli.Command{
	Name:  "gitlab",
	Usage: "GitLab related commands",
	Before: func(cCtx *cli.Context) error {
		cCtx.Context = state.WithBaseURL(cCtx.Context, cCtx.String(FlagSCMBaseURL))
		cCtx.Context = state.WithProvider(cCtx.Context, "gitlab")
		cCtx.Context = state.WithToken(cCtx.Context, cCtx.String(FlagAPIToken))

		return nil
	},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  FlagAPIToken,
			Usage: "GitLab API token",
			EnvVars: []string{
				"SCM_ENGINE_TOKEN", // SCM Engine Native
			},
		},
		&cli.StringFlag{
			Name:  FlagSCMBaseURL,
			Usage: "Base URL for the SCM instance",
			Value: "https://gitlab.com/",
			EnvVars: []string{
				"SCM_ENGINE_BASE_URL", // SCM Engine Native
				"CI_SERVER_URL",       // GitLab CI
			},
		},
	},
	Subcommands: []*cli.Command{
		{
			Name:   "lint",
			Usage:  "lint a configuration file",
			Args:   false,
			Action: Lint,
		},
		{
			Name:      "evaluate",
			Usage:     "Evaluate a Merge Request",
			Args:      true,
			ArgsUsage: " [mr_id, mr_id, ...]",
			Action:    Evaluate,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  FlagUpdatePipeline,
					Usage: "Update the CI pipeline status with progress",
					Value: true,
					EnvVars: []string{
						"SCM_ENGINE_UPDATE_PIPELINE",
					},
				},
				&cli.StringFlag{
					Name:  FlagUpdatePipelineURL,
					Usage: "(Optional) URL to where logs can be found for the pipeline",
					EnvVars: []string{
						"SCM_ENGINE_UPDATE_PIPELINE_URL",
					},
				},
				&cli.StringFlag{
					Name:  FlagSCMProject,
					Usage: "GitLab project (example: 'gitlab-org/gitlab')",
					EnvVars: []string{
						"GITLAB_PROJECT",
						"CI_PROJECT_PATH", // GitLab CI
					},
				},
				&cli.StringFlag{
					Name:  FlagMergeRequestID,
					Usage: "The Merge Request ID to process, if not provided as a CLI flag",
					EnvVars: []string{
						"CI_MERGE_REQUEST_IID", // GitLab CI
					},
				},
				&cli.StringFlag{
					Name:  FlagCommitSHA,
					Usage: "The git commit sha",
					EnvVars: []string{
						"CI_COMMIT_SHA", // GitLab CI
					},
				},
			},
		},
		{
			Name:   "server",
			Usage:  "Start HTTP server for webhook event driven usage",
			Hidden: true, // DEPRECATED
			Action: Server,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  FlagWebhookSecret,
					Usage: "Used to validate received payloads. Sent with the request in the X-Gitlab-Token HTTP header",
					EnvVars: []string{
						"SCM_ENGINE_WEBHOOK_SECRET",
					},
				},
				&cli.StringFlag{
					Name:  FlagServerListenHost,
					Usage: "IP that the HTTP server should listen on",
					Value: "0.0.0.0",
					EnvVars: []string{
						"SCM_ENGINE_LISTEN_ADDR",
					},
				},
				&cli.IntFlag{
					Name:  FlagServerListenPort,
					Usage: "Port that the HTTP server should listen on",
					Value: 3000,
					EnvVars: []string{
						"SCM_ENGINE_LISTEN_PORT",
						"PORT",
					},
				},
				&cli.DurationFlag{
					Name:  FlagServerTimeout,
					Usage: "Timeout for webhook requests",
					Value: 5 * time.Second,
					EnvVars: []string{
						"SCM_ENGINE_TIMEOUT",
					},
				},
				&cli.BoolFlag{
					Name:  FlagUpdatePipeline,
					Usage: "Update the CI pipeline status with progress",
					Value: true,
					EnvVars: []string{
						"SCM_ENGINE_UPDATE_PIPELINE",
					},
				},
				&cli.StringFlag{
					Name:  FlagUpdatePipelineURL,
					Usage: "(Optional) URL to where logs can be found for the pipeline",
					EnvVars: []string{
						"SCM_ENGINE_UPDATE_PIPELINE_URL",
					},
				},
				&cli.DurationFlag{
					Name:  FlagPeriodicEvaluationInterval,
					Usage: "(Optional) Frequency of which to evaluate all Merge Requests regardless of user activity",
					EnvVars: []string{
						"SCM_ENGINE_PERIODIC_EVALUATION_INTERVAL",
					},
				},
				&cli.StringSliceFlag{
					Name:  FlagPeriodicEvaluationIgnoreMergeRequestsWithLabel,
					Usage: "(Optional) Ignore MR with these labels",
					EnvVars: []string{
						"SCM_ENGINE_PERIODIC_EVALUATION_IGNORE_MR_WITH_LABELS",
					},
				},
				&cli.StringSliceFlag{
					Name:  FlagPeriodicEvaluationRequireMergeRequestsWithLabel,
					Usage: "(Optional) Only process MR with these labels",
					EnvVars: []string{
						"SCM_ENGINE_PERIODIC_EVALUATION_REQUIRE_MR_WITH_LABELS",
					},
				},
				&cli.StringSliceFlag{
					Name:  FlagPeriodicEvaluationOnlyProjectsWithTopics,
					Usage: "(Optional) Only evaluate projects with these topics",
					EnvVars: []string{
						"SCM_ENGINE_PERIODIC_EVALUATION_REQUIRE_PROJECT_TOPICS",
					},
				},
				&cli.BoolFlag{
					Name:  FlagPeriodicEvaluationOnlyProjectsWithMembership,
					Usage: "(Optional) Only evaluate projects with membership",
					Value: true,
					EnvVars: []string{
						"SCM_ENGINE_PERIODIC_EVALUATION_ONLY_PROJECTS_WITH_MEMBERSHIP",
					},
				},
			},
		},
	},
}
