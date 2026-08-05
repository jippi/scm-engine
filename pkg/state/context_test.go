package state_test

import (
	"context"
	"testing"
	"time"

	"github.com/jippi/scm-engine/pkg/state"
	"github.com/stretchr/testify/require"
)

// The state package stores everything on the context with unexported keys, so
// these tests double as documentation of which accessor pairs with which setter.
func TestStringValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		set  func(context.Context, string) context.Context
		get  func(context.Context) string
	}{
		{name: "base url", set: state.WithBaseURL, get: state.BaseURL},
		{name: "token", set: state.WithToken, get: state.Token},
		{name: "provider", set: state.WithProvider, get: state.Provider},
		{name: "project id", set: state.WithProjectID, get: state.ProjectID},
		{name: "config file path", set: state.WithConfigFilePath, get: state.ConfigFilePath},
		{name: "commit sha", set: state.WithCommitSHA, get: state.CommitSHA},
		{name: "merge request id", set: state.WithMergeRequestID, get: state.MergeRequestID},
		{name: "evaluation id", set: state.WithEvaluationID, get: state.EvaluationID},
		{name: "backstage url", set: state.WithBackstageURL, get: state.BackstageURL},
		{name: "backstage token", set: state.WithBackstageToken, get: state.BackstageToken},
		{name: "global config file path", set: state.WithGlobalConfigFilePath, get: state.GlobalConfigFilePath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := tt.set(t.Context(), "some-value")
			require.Equal(t, "some-value", tt.get(ctx))

			// The most recent value wins, which is what lets the server handlers
			// layer per-request state on top of the process wide state.
			ctx = tt.set(ctx, "replacement")
			require.Equal(t, "replacement", tt.get(ctx))
		})
	}
}

func TestStringValues_emptyIsPreserved(t *testing.T) {
	t.Parallel()

	ctx := state.WithBackstageURL(t.Context(), "")
	require.Empty(t, state.BackstageURL(ctx))
}

func TestWithDryRun(t *testing.T) {
	t.Parallel()

	require.True(t, state.IsDryRun(state.WithDryRun(t.Context(), true)))
	require.False(t, state.IsDryRun(state.WithDryRun(t.Context(), false)))
}

func TestWithUpdatePipeline(t *testing.T) {
	t.Parallel()

	ctx := state.WithUpdatePipeline(t.Context(), true, "https://example.com/logs")

	should, url := state.ShouldUpdatePipeline(ctx)
	require.True(t, should)
	require.Equal(t, "https://example.com/logs", url)

	ctx = state.WithUpdatePipeline(t.Context(), false, "")

	should, url = state.ShouldUpdatePipeline(ctx)
	require.False(t, should)
	require.Empty(t, url)
}

func TestWithStartTime(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, time.May, 5, 12, 0, 0, 0, time.UTC)

	require.Equal(t, now, state.StartTime(state.WithStartTime(t.Context(), now)))
}

func TestMergeRequestIDConversions(t *testing.T) {
	t.Parallel()

	ctx := state.WithMergeRequestID(t.Context(), "1234")

	require.Equal(t, "1234", state.MergeRequestID(ctx))
	require.Equal(t, 1234, state.MergeRequestIDInt(ctx))
	require.Equal(t, uint64(1234), state.MergeRequestIDUint(ctx))
}

func TestMergeRequestIDConversions_panicOnNonNumeric(t *testing.T) {
	t.Parallel()

	ctx := state.WithMergeRequestID(t.Context(), "not-a-number")

	require.Panics(t, func() { state.MergeRequestIDInt(ctx) })
	require.Panics(t, func() { state.MergeRequestIDUint(ctx) })
}

// A negative ID parses as an int but not as an unsigned one.
func TestMergeRequestIDConversions_negative(t *testing.T) {
	t.Parallel()

	ctx := state.WithMergeRequestID(t.Context(), "-1")

	require.Equal(t, -1, state.MergeRequestIDInt(ctx))
	require.Panics(t, func() { state.MergeRequestIDUint(ctx) })
}

// The seed is what makes reviewer selection reproducible for a given evaluation,
// so the same seed must produce the same sequence.
func TestWithRandomSeed(t *testing.T) {
	t.Parallel()

	first := state.RandomSeed(state.WithRandomSeed(t.Context(), 42))
	second := state.RandomSeed(state.WithRandomSeed(t.Context(), 42))

	require.NotNil(t, first)

	for range 10 {
		require.Equal(t, first.Int(), second.Int())
	}
}

func TestWithRandomSeed_differentSeedsDiverge(t *testing.T) {
	t.Parallel()

	first := state.RandomSeed(state.WithRandomSeed(t.Context(), 1))
	second := state.RandomSeed(state.WithRandomSeed(t.Context(), 2))

	diverged := false

	for range 10 {
		if first.Int() != second.Int() {
			diverged = true

			break
		}
	}

	require.True(t, diverged, "two different seeds produced the same sequence")
}

// Every accessor type-asserts the context value, so reading state that was never
// set panics rather than returning a zero value. Callers are expected to have
// populated the context in cmd/ before evaluation starts.
func TestAccessors_panicWhenUnset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		read func(context.Context)
	}{
		{name: "token", read: func(ctx context.Context) { state.Token(ctx) }},
		{name: "project id", read: func(ctx context.Context) { state.ProjectID(ctx) }},
		{name: "is dry run", read: func(ctx context.Context) { state.IsDryRun(ctx) }},
		{name: "merge request id", read: func(ctx context.Context) { state.MergeRequestID(ctx) }},
		{name: "random seed", read: func(ctx context.Context) { state.RandomSeed(ctx) }},
		{name: "backstage url", read: func(ctx context.Context) { state.BackstageURL(ctx) }},
		{name: "global config file path", read: func(ctx context.Context) { state.GlobalConfigFilePath(ctx) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Panics(t, func() { tt.read(t.Context()) })
		})
	}
}
