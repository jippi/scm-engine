package cmd

import "github.com/urfave/cli/v2"

const (
	FlagAPIToken                                        = "api-token"
	FlagBackstageURL                                    = "backstage-url"
	FlagBackstageNamespace                              = "backstage-namespace"
	FlagBackstageToken                                  = "backstage-token"
	FlagCommitSHA                                       = "commit"
	FlagConfigFile                                      = "config"
	FlagDryRun                                          = "dry-run"
	FlagMergeRequestID                                  = "id"
	FlagSCMBaseURL                                      = "base-url"
	FlagSCMProject                                      = "project"
	FlagServerListenHost                                = "listen-host"
	FlagServerListenPort                                = "listen-port"
	FlagServerTimeout                                   = "timeout"
	FlagUpdatePipeline                                  = "update-pipeline"
	FlagUpdatePipelineURL                               = "update-pipeline-url"
	FlagPeriodicEvaluationInterval                      = "periodic-evaluation-interval"
	FlagPeriodicEvaluationIgnoreMergeRequestsWithLabel  = "periodic-evaluation-ignore-mr-labels"
	FlagPeriodicEvaluationRequireMergeRequestsWithLabel = "periodic-evaluation-require-mr-labels"
	FlagPeriodicEvaluationOnlyProjectsWithTopics        = "periodic-evaluation-project-topics"
	FlagPeriodicEvaluationOnlyProjectsWithMembership    = "periodic-evaluation-only-project-membership"
	FlagWebhookSecret                                   = "webhook-secret"
)

var (
	StringFlagBackstageURL = &cli.StringFlag{
		Name:  FlagBackstageURL,
		Usage: "The Backstage base URL",
		EnvVars: []string{
			"BACKSTAGE_URL", // Backstage catalog integration
		},
	}
	StringFlagBackstageNamespace = &cli.StringFlag{
		Name:  FlagBackstageNamespace,
		Usage: "The Backstage namespace",
		EnvVars: []string{
			"BACKSTAGE_NAMESPACE", // Backstage catalog integration
		},
		Value: "default",
	}
	StringFlagBackstageToken = &cli.StringFlag{
		Name:  FlagBackstageToken,
		Usage: "The Backstage static token with access to the catalog plugin", // https://backstage.io/docs/auth/service-to-service-auth/#static-tokens
		EnvVars: []string{
			"BACKSTAGE_TOKEN", // Backstage catalog integration
		},
	}
)
