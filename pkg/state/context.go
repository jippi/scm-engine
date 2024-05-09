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
)

func NewContext(project, mr string) context.Context {
	ctx := context.Background()
	ctx = ContextWithProjectID(ctx, project)
	ctx = ContextWithMergeRequestID(ctx, mr)

	return ctx
}

func ProjectIDFromContext(ctx context.Context) string {
	return ctx.Value(projectID).(string) //nolint:forcetypeassert
}

func ContextWithProjectID(ctx context.Context, value string) context.Context {
	ctx = slogctx.With(ctx, slog.String("project_id", value))
	ctx = context.WithValue(ctx, projectID, value)

	return ctx
}

func ContextWithDryRun(ctx context.Context, dry bool) context.Context {
	ctx = slogctx.With(ctx, slog.Bool("dry_run", dry))
	ctx = context.WithValue(ctx, dryRun, dry)

	return ctx
}

func IsDryRun(ctx context.Context) bool {
	return ctx.Value(dryRun).(bool) //nolint:forcetypeassert
}

func ContextWithMergeRequestID(ctx context.Context, id string) context.Context {
	ctx = slogctx.With(ctx, "merge_request_id", id)
	ctx = context.WithValue(ctx, mergeRequestID, id)

	return ctx
}

func MergeRequestIDFromContext(ctx context.Context) string {
	return ctx.Value(mergeRequestID).(string) //nolint:forcetypeassert
}

func MergeRequestIDFromContextInt(ctx context.Context) int {
	val := MergeRequestIDFromContext(ctx)

	number, err := strconv.Atoi(val)
	if err != nil {
		panic(err)
	}

	return number
}
