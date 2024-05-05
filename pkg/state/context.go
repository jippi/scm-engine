package state

import "context"

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

func ContextWithMergeRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, mergeRequestID, id)
}
