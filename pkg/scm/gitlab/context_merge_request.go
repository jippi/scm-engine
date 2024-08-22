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

// has_any_activity_within
func (e ContextMergeRequest) HasAnyActivityWithin(ctx context.Context, input any) bool {
	dur := stdlib.ToDuration(input)
	now := time.Now()
	cfg := config.FromContext(ctx)

	for _, note := range e.Notes {
		// Check if we should ignore the actor (user) activity
		if cfg.IgnoreActivityFrom.Matches(note.Author.ToActorMatcher()) {
			continue
		}

		// Check is within the configured duration
		if now.Sub(note.UpdatedAt) < dur {
			return true
		}
	}

	// If we have a recent commit, check if its within the duration
	return e.LastCommit != nil && now.Sub(*e.LastCommit.CommittedDate) < dur
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
		if cfg.IgnoreActivityFrom.Matches(note.Author.ToActorMatcher()) {
			continue
		}

		// Ignore "scm-engine" activity since we shouldn't consider ourself a user
		if e.CurrentUser.Username == note.Author.Username {
			continue
		}

		if now.Sub(note.UpdatedAt) < dur {
			return true
		}
	}

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
