package state

import (
	"context"
	"log/slog"
	"strconv"

	slogctx "github.com/veqryn/slog-context"
)

type contextKey uint

const (
	projectID contextKey = iota
	dryRun
	mergeRequestID
	commitSha
	updatePipeline
)

func ProjectID(ctx context.Context) string {
	return ctx.Value(projectID).(string) //nolint:forcetypeassert
}

func CommitSHA(ctx context.Context) string {
	return ctx.Value(commitSha).(string) //nolint:forcetypeassert
}

func WithProjectID(ctx context.Context, value string) context.Context {
	ctx = slogctx.With(ctx, slog.String("project_id", value))
	ctx = context.WithValue(ctx, projectID, value)

	return ctx
}

func WithDryRun(ctx context.Context, dry bool) context.Context {
	ctx = slogctx.With(ctx, slog.Bool("dry_run", dry))
	ctx = context.WithValue(ctx, dryRun, dry)

	return ctx
}

func WithUpdatePipeline(ctx context.Context, update bool) context.Context {
	ctx = slogctx.With(ctx, slog.Bool("update_pipeline", update))
	ctx = context.WithValue(ctx, updatePipeline, update)

	return ctx
}

func WithCommitSHA(ctx context.Context, sha string) context.Context {
	ctx = slogctx.With(ctx, slog.String("git_commit_sha", sha))
	ctx = context.WithValue(ctx, commitSha, sha)

	return ctx
}

func ContextWithMergeRequestID(ctx context.Context, id string) context.Context {
	ctx = slogctx.With(ctx, "merge_request_id", id)
	ctx = context.WithValue(ctx, mergeRequestID, id)

	return ctx
}

func IsDryRun(ctx context.Context) bool {
	return ctx.Value(dryRun).(bool) //nolint:forcetypeassert
}

func ShouldUpdatePipeline(ctx context.Context) bool {
	return ctx.Value(updatePipeline).(bool) //nolint:forcetypeassert
}

func MergeRequestID(ctx context.Context) string {
	return ctx.Value(mergeRequestID).(string) //nolint:forcetypeassert
}

func MergeRequestIDInt(ctx context.Context) int {
	val := MergeRequestID(ctx)

	number, err := strconv.Atoi(val)
	if err != nil {
		panic(err)
	}

	return number
}
