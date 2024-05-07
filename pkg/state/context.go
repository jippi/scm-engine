package state

import (
	"context"
	"strconv"
)

type contextKey uint

const (
	projectID contextKey = iota
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
	return context.WithValue(ctx, projectID, value)
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

func ContextWithMergeRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, mergeRequestID, id)
}
