package gitlab

import (
	"context"
	"fmt"
	"time"

	"github.com/jippi/scm-engine/pkg/config"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/stdlib"
)

func (e ContextMergeRequest) HasLabel(in string) bool {
	for _, label := range e.Labels {
		if label.Title == in {
			return true
		}
	}

	return false
}

func (e ContextMergeRequest) HasNoLabel(in string) bool {
	return !e.HasLabel(in)
}

func (e ContextMergeRequest) StateIs(anyOf ...string) bool {
	for _, state := range anyOf {
		if !MergeRequestState(state).IsValid() {
			panic(fmt.Errorf("unknown state value: %q", state))
		}

		if state == e.State {
			return true
		}
	}

	return false
}

func (e ContextMergeRequest) StateIsNot(anyOf ...string) bool {
	for _, state := range anyOf {
		if !MergeRequestState(state).IsValid() {
			panic(fmt.Errorf("unknown state value: %q", state))
		}

		if state == e.State {
			return false
		}
	}

	return true
}

// has_no_activity_within
func (e ContextMergeRequest) HasNoActivityWithin(ctx context.Context, input any) bool {
	return !e.HasAnyActivityWithin(ctx, input)
}

// has_activity_within (alias)
func (e ContextMergeRequest) HasActivityWithin(ctx context.Context, input any) bool {
	return e.HasAnyActivityWithin(ctx, input)
}

// updatedWithinDuration checks if the MR has been updated within the provided duration
func (e ContextMergeRequest) updatedWithinDuration(now time.Time, dur time.Duration) bool {
	return now.Sub(e.UpdatedAt) < dur
}

// has_any_activity_within
func (e ContextMergeRequest) HasAnyActivityWithin(ctx context.Context, input any) bool {
	dur := stdlib.ToDuration(input)
	now := time.Now()
	cfg := config.FromContext(ctx)

	// If the MR UpdatedAt has been updated within the duration, then we got some kind of activity
	if e.updatedWithinDuration(now, dur) {
		return true
	}

	// If we have a recent commit, check if its within the duration
	if e.LastCommit != nil && now.Sub(*e.LastCommit.CommittedDate) < dur {
		return true
	}

	for _, note := range e.Notes {
		// Check if we should ignore the actor (user) activity
		if cfg.IgnoreActivityFrom.Matches(note.Author.ToActor()) {
			continue
		}

		// Check is within the configured duration
		if now.Sub(note.UpdatedAt) < dur {
			return true
		}
	}

	// No positive matches, so we conclude there was no activity
	return false
}

// has_no_user_activity_within
func (e ContextMergeRequest) HasNoUserActivityWithin(ctx context.Context, input any) bool {
	return !e.HasUserActivityWithin(ctx, input)
}

// has_user_activity_within
func (e ContextMergeRequest) HasUserActivityWithin(ctx context.Context, input any) bool {
	dur := stdlib.ToDuration(input)
	now := time.Now()
	cfg := config.FromContext(ctx)

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
			return true
		}
	}

	// NOTE: we can't use the "UpdatedAt" timestamp on the MergeRequest because we
	//       can't guarantee it was a user activity change that bumped the timestamp;
	//       use the "has_any_activity_within" function instead for that

	// If we have a recent commit, check if its within the duration
	return e.LastCommit != nil && now.Sub(*e.LastCommit.CommittedDate) < dur
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
