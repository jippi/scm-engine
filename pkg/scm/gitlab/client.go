package gitlab

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/aquilax/truncate"
	"github.com/hasura/go-graphql-client"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
	slogctx "github.com/veqryn/slog-context"
	go_gitlab "github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

var pipelineName = scm.Ptr("scm-engine")

// Skip the "update external pipeline" step if the HEAD pipeline is any
// of the configured options
//
// Disable updating CI pipelines when running periodic evaluations in the background
// to avoid sending "CI pipeline failed" notification emails to users
//
// If the HEAD CI pipeline for a MR is in "failed" mode and we update the external
// pipeline, GitLab will send the user (if configured in their profile) a 'your pipeline failed'
// email every time the background evaluation process runs.
//
// Possible values:
//
// - CANCELED
// - CANCELING
// - CREATED
// - FAILED
// - MANUAL
// - PENDING
// - PREPARING
// - RUNNING
// - SCHEDULED
// - SKIPPED
// - SUCCESS
// - WAITING_FOR_CALLBACK
// - WAITING_FOR_RESOURCE
var SkipPipelineUpdateIfPeriodicAndPipelineStatusIs = []string{
	"FAILED",
	"SKIPPED",
}

// Ensure the GitLab client implements the [scm.Client]
var _ scm.Client = (*Client)(nil)

// Client is a wrapper around the GitLab specific implementation of [scm.Client] interface
type Client struct {
	wrapped *go_gitlab.Client

	labels        *LabelClient
	mergeRequests *MergeRequestClient
}

// NewClient creates a new GitLab client
func NewClient(ctx context.Context) (*Client, error) {
	client, err := go_gitlab.NewClient(state.Token(ctx), go_gitlab.WithBaseURL(state.BaseURL(ctx)))
	if err != nil {
		return nil, err
	}

	return &Client{wrapped: client}, nil
}

// Labels returns a client target at managing labels/tags
func (client *Client) Labels() scm.LabelClient {
	if client.labels == nil {
		client.labels = NewLabelClient(client)
	}

	return client.labels
}

// MergeRequests returns a client target at managing merge/pull requests
func (client *Client) MergeRequests() scm.MergeRequestClient {
	if client.mergeRequests == nil {
		client.mergeRequests = NewMergeRequestClient(client)
	}

	return client.mergeRequests
}

// FindMergeRequestsForPeriodicEvaluation will find all Merge Requests legible for
// periodic re-evaluation.
func (client *Client) FindMergeRequestsForPeriodicEvaluation(ctx context.Context, filters scm.MergeRequestListFilters) ([]scm.PeriodicEvaluationMergeRequest, error) {
	var response PeriodicEvaluationResult

	if err := client.newGraphQLClient(ctx).Query(ctx, &response, filters.AsGraphqlVariables()); err != nil {
		return nil, err
	}

	slogctx.Debug(ctx, fmt.Sprintf("Found %d projects", len(response.Projects.Nodes)))

	var result []scm.PeriodicEvaluationMergeRequest

	// Check if the evaluation should update CI pipelines by default
	updatePipeline, _ := state.ShouldUpdatePipeline(ctx)

	for _, project := range response.Projects.Nodes {
		slogctx.Debug(ctx, fmt.Sprintf("Project %s has %d Merge Requests", project.FullPath, len(project.MergeRequests.Nodes)))

		for _, mr := range project.MergeRequests.Nodes {
			item := scm.PeriodicEvaluationMergeRequest{
				Project:        project.FullPath,
				MergeRequestID: mr.IID,
				SHA:            mr.SHA,
				UpdatePipeline: updatePipeline,
			}

			// If periodic evaluation are updating CI pipelines, check if the status of the HEAD pipeline
			// is in a state where re-triggering the external CI pipeline would potentially send the MR creator
			// a "Your CI pipeline failed" e-mail every time we evaluate the MR in the background (spammy!)
			if item.UpdatePipeline && mr.HeadPipeline != nil && slices.Contains(SkipPipelineUpdateIfPeriodicAndPipelineStatusIs, mr.HeadPipeline.Status) {
				item.UpdatePipeline = false
			}

			// Only set the ConfigBlob struct if the config file exists in the repository
			if len(project.Repository.Blobs.Nodes) == 1 {
				item.ConfigBlob = project.Repository.Blobs.Nodes[0].Blob
			}

			result = append(result, item)
		}
	}

	return result, nil
}

// EvalContext creates a new evaluation context for GitLab specific usage
func (client *Client) EvalContext(ctx context.Context) (scm.EvalContext, error) {
	return NewContext(ctx, graphqlBaseURL(client.wrapped.BaseURL()), state.Token(ctx))
}

func (client *Client) GetProjectFiles(ctx context.Context, project string, ref *string, files []string) (map[string]string, error) {
	if len(project) == 0 {
		return nil, errors.New("Missing required 'project' value for include")
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("Missing list of files to include from project [%s]", project)
	}

	var (
		response  IncludeConfigurationResult
		variables = map[string]any{
			"project": graphql.ID(project),
			"files":   files,
			"ref":     ref,
		}
	)

	if err := client.newGraphQLClient(ctx).Query(ctx, &response, variables); err != nil {
		return nil, fmt.Errorf("GraphQL query failed while trying to read remote configuration files [%v] for project [%s]: %w", files, project, err)
	}

	fileContents := map[string]string{}

	// Convert the GraphQL response into a simple map
	for _, blob := range response.Project.Repository.Blobs.Nodes {
		fileContents[blob.Path] = blob.Blob
	}

	// Check if the files provided as input all exist in the file content and is not empty
	for _, file := range files {
		val, ok := fileContents[file]
		if !ok {
			return nil, fmt.Errorf("configuration file [%s] in project [%s] does not exist (or could not be read)", file, project)
		}

		if len(val) == 0 {
			return nil, fmt.Errorf("configuration file [%s] in project [%s] is empty", file, project)
		}
	}

	return fileContents, nil
}

// Start pipeline
func (client *Client) Start(ctx context.Context) error {
	ok, pattern := state.ShouldUpdatePipeline(ctx)
	if !ok {
		return nil
	}

	var targetURL *string

	if len(pattern) != 0 {
		link := pattern
		link = strings.ReplaceAll(link, "__ID__", state.EvaluationID(ctx))
		link = strings.ReplaceAll(link, "__MR_ID__", state.MergeRequestID(ctx))
		link = strings.ReplaceAll(link, "__PROJECT_ID__", state.ProjectID(ctx))
		link = strings.ReplaceAll(link, "__START_TS_MS__", strconv.FormatInt(state.StartTime(ctx).UnixMilli(), 10))
		link = strings.ReplaceAll(link, "__STOP_TS_MS__", "")

		targetURL = &link
	}

	_, response, err := client.wrapped.Commits.SetCommitStatus(state.ProjectID(ctx), state.CommitSHA(ctx), &go_gitlab.SetCommitStatusOptions{
		State:       go_gitlab.Running,
		Context:     pipelineName,
		Description: scm.Ptr("Currently evaluating MR"),
		TargetURL:   targetURL,
	})

	switch response.StatusCode {
	// GitLab returns '400 {message: {name: [has already been taken]}}' if the context/pipeline name we use
	// are already used by a regular CI job. We treat that as a non-failure and continue on after logging
	case http.StatusBadRequest:
		slogctx.Warn(ctx, "could not update commit pipeline status", slog.Any("err", err))

		return nil

	default:
		return err
	}
}

// Stop pipeline
func (client *Client) Stop(ctx context.Context, evalError error, allowPipelineFailure bool) error {
	ok, pattern := state.ShouldUpdatePipeline(ctx)
	if !ok {
		return nil
	}

	var targetURL *string

	if len(pattern) != 0 {
		link := pattern
		link = strings.ReplaceAll(link, "__ID__", state.EvaluationID(ctx))
		link = strings.ReplaceAll(link, "__MR_ID__", state.MergeRequestID(ctx))
		link = strings.ReplaceAll(link, "__PROJECT_ID__", state.ProjectID(ctx))
		link = strings.ReplaceAll(link, "__START_TS_MS__", strconv.FormatInt(state.StartTime(ctx).UnixMilli(), 10))
		link = strings.ReplaceAll(link, "__STOP_TS_MS__", strconv.FormatInt(time.Now().UnixMilli(), 10))

		targetURL = &link
	}

	var (
		status      = go_gitlab.Success
		description = "OK"
	)

	if evalError != nil {
		if allowPipelineFailure {
			status = go_gitlab.Failed
		}

		description = truncate.Truncate(evalError.Error(), 250, "...", truncate.PositionEnd)
	}

	_, response, err := client.wrapped.Commits.SetCommitStatus(state.ProjectID(ctx), state.CommitSHA(ctx), &go_gitlab.SetCommitStatusOptions{
		State:       status,
		Context:     pipelineName,
		Description: scm.Ptr(description),
		TargetURL:   targetURL,
	})

	switch response.StatusCode {
	// GitLab returns '400 {message: {name: [has already been taken]}}' if the context/pipeline name we use
	// are already used by a regular CI job. We treat that as a non-failure and continue on after logging
	case http.StatusBadRequest:
		slogctx.Warn(ctx, "could not update commit pipeline status", slog.Any("err", err))

		return nil

	default:
		return err
	}
}

func graphqlBaseURL(inputURL *url.URL) string {
	var buf strings.Builder
	if inputURL.Scheme != "" {
		buf.WriteString(inputURL.Scheme)
		buf.WriteByte(':')
	}

	if inputURL.Scheme != "" || inputURL.Host != "" || inputURL.User != nil {
		if inputURL.OmitHost && inputURL.Host == "" && inputURL.User == nil {
			// omit empty host
		} else {
			if inputURL.Host != "" || inputURL.Path != "" || inputURL.User != nil {
				buf.WriteString("//")
			}

			if ui := inputURL.User; ui != nil {
				buf.WriteString(ui.String())
				buf.WriteByte('@')
			}

			if h := inputURL.Host; h != "" {
				buf.WriteString(inputURL.Host)
			}
		}
	}

	return buf.String()
}

func (client Client) newGraphQLClient(ctx context.Context) *graphql.Client {
	httpClient := oauth2.NewClient(
		ctx,
		oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: state.Token(ctx),
			},
		),
	)

	return graphql.NewClient(
		graphqlBaseURL(client.wrapped.BaseURL())+"/api/graphql",
		httpClient,
	)
}
