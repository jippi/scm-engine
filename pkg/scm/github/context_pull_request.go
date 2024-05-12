package github

import (
	"fmt"

	"github.com/jippi/scm-engine/pkg/scm"
)

func (e ContextPullRequest) IsApproved() bool {
	return e.ReviewDecision == PullRequestReviewDecisionApproved
}

func (e ContextPullRequest) StateIs(anyOf ...string) bool {
	for _, state := range anyOf {
		if !PullRequestState(state).IsValid() {
			panic(fmt.Errorf("unknown state value: %q", state))
		}

		if state == e.State.String() {
			return true
		}
	}

	return false
}

func (e ContextPullRequest) HasLabel(in string) bool {
	for _, label := range e.Labels {
		if label.Name == in {
			return true
		}
	}

	return false
}

func (e ContextPullRequest) HasNoLabel(in string) bool {
	return !e.HasLabel(in)
}

func (e ContextPullRequest) ModifiedFilesList(patterns ...string) []string {
	return e.findModifiedFiles(patterns...)
}

// Partially lifted from https://github.com/hmarr/codeowners/blob/main/match.go
func (e ContextPullRequest) ModifiedFiles(patterns ...string) bool {
	return len(e.findModifiedFiles(patterns...)) > 0
}

func (e ContextPullRequest) findModifiedFiles(patterns ...string) []string {
	files := []string{}
	for _, f := range e.Files {
		files = append(files, f.Path)
	}

	return scm.FindModifiedFiles(files, patterns...)
}
