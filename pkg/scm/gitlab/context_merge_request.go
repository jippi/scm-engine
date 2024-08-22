package gitlab

import (
	"fmt"
	"time"

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
func (e ContextMergeRequest) HasNoActivityWithin(input any) bool {
	return !e.HasAnyActivityWithin(input)
}

// has_any_activity_within
func (e ContextMergeRequest) HasAnyActivityWithin(input any) bool {
	dur := stdlib.ToDuration(input)
	now := time.Now()

	for _, note := range e.Notes {
		if now.Sub(note.UpdatedAt) < dur {
			return true
		}
	}

	if e.LastCommit != nil {
		if now.Sub(*e.LastCommit.CommittedDate) < dur {
			return true
		}
	}

	return false
}

// has_no_user_activity_within
func (e ContextMergeRequest) HasNoUserActivityWithin(input any) bool {
	return !e.HasUserActivityWithin(input)
}

// has_user_activity_within
func (e ContextMergeRequest) HasUserActivityWithin(input any) bool {
	dur := stdlib.ToDuration(input)
	now := time.Now()

	for _, note := range e.Notes {
		// Ignore "my" activity
		if e.CurrentUser.Username == note.Author.Username {
			continue
		}

		// Ignore bots
		if e.Author.Bot {
			continue
		}

		if now.Sub(note.UpdatedAt) < dur {
			return true
		}
	}

	if e.LastCommit != nil {
		if now.Sub(*e.LastCommit.CommittedDate) < dur {
			return true
		}
	}

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
