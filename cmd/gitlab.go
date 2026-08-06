package cmd

import (
	"context"
	"time"

	"github.com/jippi/scm-engine/pkg/state"
	"github.com/urfave/cli/v3"
)

var GitLab = &cli.Command{
	Name:  "gitlab",
	Usage: "GitLab related commands",
	Before: func(ctx context.Context, cCtx *cli.Command) (context.Context, error) {
		ctx = state.WithBaseURL(ctx, cCtx.String(FlagSCMBaseURL))
		ctx = state.WithProvider(ctx, "gitlab")
		ctx = state.WithToken(ctx, cCtx.String(FlagAPIToken))
		ctx = state.WithGlobalConfigFilePath(ctx, cCtx.String(FlagGlobalConfigFile))

		return ctx, nil
	},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  FlagAPIToken,
			Usage: "GitLab API token",
			Sources: cli.EnvVars(
				"SCM_ENGINE_TOKEN", // SCM Engine Native
			),
		},
		&cli.StringFlag{
			Name:  FlagSCMBaseURL,
			Usage: "Base URL for the SCM instance",
			Value: "https://gitlab.com/",
			Sources: cli.EnvVars(
				"SCM_ENGINE_BASE_URL", // SCM Engine Native
				"CI_SERVER_URL",       // GitLab CI
			),
		},
		&cli.StringFlag{
			Name:    FlagGlobalConfigFile,
			Usage:   "Path to a global configuration file. Any repository specific configuration will be merged on top of the global configuration",
			Value:   "",
			Sources: cli.EnvVars("SCM_ENGINE_GLOBAL_CONFIG_FILE"),
		},
	},
	Commands: []*cli.Command{
		{
			Name:   "lint",
			Usage:  "lint a configuration file",
			Action: Lint,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "schema",
					Usage: "Where to find the JSON Schema file. Can load the file from either the embedded version (default), http://, https://, or a file:// URI",
					Value: "embed://",
				},
			},
		},
		{
			Name:      "evaluate",
			Usage:     "Evaluate a Merge Request",
			ArgsUsage: " [mr_id, mr_id, ...]",
			Action:    Evaluate,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    FlagUpdatePipeline,
					Usage:   "Update the CI pipeline status with progress",
					Value:   true,
					Sources: cli.EnvVars("SCM_ENGINE_UPDATE_PIPELINE"),
				},
				&cli.StringFlag{
					Name:    FlagUpdatePipelineURL,
					Usage:   "(Optional) URL to where logs can be found for the pipeline",
					Sources: cli.EnvVars("SCM_ENGINE_UPDATE_PIPELINE_URL"),
				},
				&cli.StringFlag{
					Name:  FlagSCMProject,
					Usage: "GitLab project (example: 'gitlab-org/gitlab')",
					Sources: cli.EnvVars(
						"GITLAB_PROJECT",
						"CI_PROJECT_PATH", // GitLab CI
					),
				},
				&cli.StringFlag{
					Name:  FlagMergeRequestID,
					Usage: "The Merge Request ID to process, if not provided as a CLI flag",
					Sources: cli.EnvVars(
						"CI_MERGE_REQUEST_IID", // GitLab CI
					),
				},
				&cli.StringFlag{
					Name:  FlagCommitSHA,
					Usage: "The git commit sha",
					Sources: cli.EnvVars(
						"CI_COMMIT_SHA", // GitLab CI
					),
				},
				StringFlagBackstageURL,
				StringFlagBackstageToken,
			},
		},
		{
			Name:   "server",
			Usage:  "Start HTTP server for webhook event driven usage",
			Hidden: true, // DEPRECATED
			Action: Server,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    FlagWebhookSecret,
					Usage:   "Used to validate received payloads. Sent with the request in the X-Gitlab-Token HTTP header",
					Sources: cli.EnvVars("SCM_ENGINE_WEBHOOK_SECRET"),
				},
				&cli.StringFlag{
					Name:    FlagServerListenHost,
					Usage:   "IP that the HTTP server should listen on",
					Value:   "0.0.0.0",
					Sources: cli.EnvVars("SCM_ENGINE_LISTEN_ADDR"),
				},
				&cli.IntFlag{
					Name:  FlagServerListenPort,
					Usage: "Port that the HTTP server should listen on",
					Value: 3000,
					Sources: cli.EnvVars(
						"SCM_ENGINE_LISTEN_PORT",
						"PORT",
					),
				},
				&cli.DurationFlag{
					Name:    FlagServerTimeout,
					Usage:   "Timeout for webhook requests",
					Value:   5 * time.Second,
					Sources: cli.EnvVars("SCM_ENGINE_TIMEOUT"),
				},
				&cli.BoolFlag{
					Name:    FlagUpdatePipeline,
					Usage:   "Update the CI pipeline status with progress",
					Value:   true,
					Sources: cli.EnvVars("SCM_ENGINE_UPDATE_PIPELINE"),
				},
				&cli.StringFlag{
					Name:    FlagUpdatePipelineURL,
					Usage:   "(Optional) URL to where logs can be found for the pipeline",
					Sources: cli.EnvVars("SCM_ENGINE_UPDATE_PIPELINE_URL"),
				},
				&cli.DurationFlag{
					Name:    FlagPeriodicEvaluationInterval,
					Usage:   "(Optional) Frequency of which to evaluate all Merge Requests regardless of user activity",
					Sources: cli.EnvVars("SCM_ENGINE_PERIODIC_EVALUATION_INTERVAL"),
				},
				&cli.StringSliceFlag{
					Name:    FlagPeriodicEvaluationIgnoreMergeRequestsWithLabel,
					Usage:   "(Optional) Ignore MR with these labels",
					Sources: cli.EnvVars("SCM_ENGINE_PERIODIC_EVALUATION_IGNORE_MR_WITH_LABELS"),
				},
				&cli.StringSliceFlag{
					Name:    FlagPeriodicEvaluationRequireMergeRequestsWithLabel,
					Usage:   "(Optional) Only process MR with these labels",
					Sources: cli.EnvVars("SCM_ENGINE_PERIODIC_EVALUATION_REQUIRE_MR_WITH_LABELS"),
				},
				&cli.StringSliceFlag{
					Name:    FlagPeriodicEvaluationOnlyProjectsWithTopics,
					Usage:   "(Optional) Only evaluate projects with these topics",
					Sources: cli.EnvVars("SCM_ENGINE_PERIODIC_EVALUATION_REQUIRE_PROJECT_TOPICS"),
				},
				&cli.BoolFlag{
					Name:    FlagPeriodicEvaluationOnlyProjectsWithMembership,
					Usage:   "(Optional) Only evaluate projects with membership",
					Value:   true,
					Sources: cli.EnvVars("SCM_ENGINE_PERIODIC_EVALUATION_ONLY_PROJECTS_WITH_MEMBERSHIP"),
				},
				StringFlagBackstageURL,
				StringFlagBackstageToken,
			},
		},
	},
}
