package gitlab

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/stdlib"
	slogctx "github.com/veqryn/slog-context"
)

func withFunction(in string) slog.Attr {
	return slog.String("function_name", in)
}

func withResult(in bool) slog.Attr {
	return slog.Bool("function_result", in)
}

func withInput(in any) slog.Attr {
	return slog.Any("function_argument", in)
}

func withComparisonValue(in any) slog.Attr {
	return slog.Any("function_comparison_value", in)
}

func withSubCondition(in string) slog.Attr {
	return slog.String("function_sub_condition_that_matched", in)
}

const defaultScriptEvalResult = "script function eval result"

func (e ContextMergeRequest) HasLabel(ctx context.Context, input string) bool {
	ctx = slogctx.With(ctx, withFunction("merge_request.has_label"), withInput(input))

	for _, label := range e.Labels {
		if label.Title == input {
			slogctx.Debug(ctx, defaultScriptEvalResult, withResult(true))

			return true
		}
	}

	slogctx.Debug(ctx, defaultScriptEvalResult, withResult(false))

	return false
}

func (e ContextMergeRequest) HasNoLabel(ctx context.Context, input string) bool {
	val := !e.HasLabel(ctx, input)

	slogctx.Debug(ctx, defaultScriptEvalResult,
		withFunction("merge_request.has_no_label"),
		withResult(val),
		withInput(input),
	)

	return val
}

func (e ContextMergeRequest) StateIs(ctx context.Context, anyOf ...string) bool {
	ctx = slogctx.With(ctx, withFunction("merge_request.state_is"))

	for _, state := range anyOf {
		if !MergeRequestState(state).IsValid() {
			panic(fmt.Errorf("unknown state value: %q", state))
		}

		if state == e.State {
			slogctx.Debug(ctx, defaultScriptEvalResult, withResult(true))

			return true
		}
	}

	slogctx.Debug(ctx, defaultScriptEvalResult, withResult(false))

	return false
}

func (e ContextMergeRequest) StateIsNot(ctx context.Context, anyOf ...string) bool {
	ctx = slogctx.With(ctx, withFunction("merge_request.state_is_not"))

	for _, state := range anyOf {
		if !MergeRequestState(state).IsValid() {
			panic(fmt.Errorf("unknown state value: %q", state))
		}

		if state == e.State {
			slogctx.Debug(ctx, defaultScriptEvalResult, withResult(false))

			return false
		}
	}

	slogctx.Debug(ctx, defaultScriptEvalResult, withResult(true))

	return true
}

// has_no_activity_within
func (e ContextMergeRequest) HasNoActivityWithin(ctx context.Context, input any) bool {
	val := !e.HasAnyActivityWithin(ctx, input)

	slogctx.Debug(ctx, defaultScriptEvalResult,
		withFunction("merge_request.has_no_activity_within"),
		withResult(val),
	)

	return val
}

// has_activity_within (alias)
func (e ContextMergeRequest) HasActivityWithin(ctx context.Context, input any) bool {
	val := e.HasAnyActivityWithin(ctx, input)

	slogctx.Debug(ctx, defaultScriptEvalResult,
		withFunction("merge_request.has_activity_within"),
		withResult(val),
	)

	return val
}

// has_any_activity_within
func (e ContextMergeRequest) HasAnyActivityWithin(ctx context.Context, input any) bool {
	dur := stdlib.ToDuration(input)
	now := time.Now()
	cfg := config.FromContext(ctx)

	ctx = slogctx.With(ctx,
		withFunction("merge_request.has_any_activity_within"),
		withInput(dur),
	)

	// If the MR UpdatedAt has been updated within the duration, then we got some kind of activity
	if e.updatedWithinDuration(now, dur) {
		slogctx.Debug(ctx, defaultScriptEvalResult,
			withResult(true),
			withSubCondition("updated_with_duration"),
			withComparisonValue(e.UpdatedAt),
		)

		return true
	}

	// If we have a recent commit, check if its within the duration
	if e.LastCommit != nil && now.Sub(*e.LastCommit.CommittedDate) < dur {
		slogctx.Debug(ctx, defaultScriptEvalResult,
			withResult(true),
			withSubCondition("last_commit_created_at"),
			withComparisonValue(*e.LastCommit.CommittedDate),
		)

		return true
	}

	for _, note := range e.Notes {
		// Check if we should ignore the actor (user) activity
		if cfg.IgnoreActivityFrom.Matches(note.Author.ToActor()) {
			continue
		}

		// Check is within the configured duration
		if now.Sub(note.UpdatedAt) < dur {
			slogctx.Debug(ctx, defaultScriptEvalResult,
				withResult(true),
				withSubCondition("note_updated_at"),
				withComparisonValue(note.UpdatedAt),
			)

			return true
		}
	}

	slogctx.Debug(ctx, defaultScriptEvalResult,
		withResult(false),
		withSubCondition("default"),
	)

	// No positive matches, so we conclude there was no activity
	return false
}

// has_no_user_activity_within
func (e ContextMergeRequest) HasNoUserActivityWithin(ctx context.Context, input any) bool {
	val := !e.HasUserActivityWithin(ctx, input)

	slogctx.Debug(ctx, defaultScriptEvalResult,
		withFunction("merge_request.has_no_activity_within"),
		withResult(val),
	)

	return val
}

// has_user_activity_within
func (e ContextMergeRequest) HasUserActivityWithin(ctx context.Context, input any) bool {
	dur := stdlib.ToDuration(input)
	now := time.Now()
	cfg := config.FromContext(ctx)

	ctx = slogctx.With(ctx,
		withFunction("merge_request.has_user_activity_within"),
		withInput(dur),
	)

	for _, note := range e.Notes {
		// Check if we should ignore the actor (user) activity
		if cfg.IgnoreActivityFrom.Matches(note.Author.ToActor()) {
			continue
		}

		// Ignore all bots when considering 'user' activity
		if note.Author.Bot {
			continue
		}

		// Ignore "scm-engine" activity since we shouldn't consider ourself a user
		if e.CurrentUser.Username == note.Author.Username {
			continue
		}

		// Check if the note is within the duration
		if now.Sub(note.UpdatedAt) < dur {
			slogctx.Debug(ctx, defaultScriptEvalResult,
				withResult(true),
				withSubCondition("note_updated_at"),
				withComparisonValue(note.UpdatedAt),
			)

			return true
		}
	}

	// NOTE: we can't use the "UpdatedAt" timestamp on the MergeRequest because we
	//       can't guarantee it was a user activity change that bumped the timestamp;
	//       use the "has_any_activity_within" function instead for that

	// If we have a recent commit, check if its within the duration
	if e.LastCommit != nil && now.Sub(*e.LastCommit.CommittedDate) < dur {
		slogctx.Debug(ctx, defaultScriptEvalResult,
			withResult(true),
			withSubCondition("last_commit_created_at"),
			withComparisonValue(*e.LastCommit.CommittedDate),
		)

		return true
	}

	slogctx.Debug(ctx, defaultScriptEvalResult,
		withResult(false),
		withSubCondition("default"),
	)

	// No positive matches, so we conclude there was no activity
	return false
}

func (e ContextMergeRequest) ModifiedFilesList(patterns ...string) []string {
	return e.findModifiedFiles(patterns...)
}

// Partially lifted from https://github.com/hmarr/codeowners/blob/main/match.go
func (e ContextMergeRequest) ModifiedFiles(patterns ...string) bool {
	return len(e.findModifiedFiles(patterns...)) > 0
}

func (e ContextMergeRequest) findModifiedFiles(patterns ...string) []string {
	files := []string{}
	for _, f := range e.DiffStats {
		files = append(files, f.Path)
	}

	return scm.FindModifiedFiles(files, patterns...)
}

// updatedWithinDuration checks if the MR has been updated within the provided duration
func (e ContextMergeRequest) updatedWithinDuration(now time.Time, dur time.Duration) bool {
	return now.Sub(e.UpdatedAt) < dur
}
