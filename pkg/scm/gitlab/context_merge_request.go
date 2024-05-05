package gitlab

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type ContextMergeRequest struct {
	ApprovalsLeft                 int                           `expr:"approvals_left" graphql:"approvalsLeft"`
	ApprovalsRequired             int                           `expr:"approvals_required" graphql:"approvalsRequired"`
	Approved                      bool                          `expr:"approved"`
	AutoMergeEnabled              bool                          `expr:"auto_merge_enabled" graphql:"autoMergeEnabled"`
	AutoMergeStrategy             string                        `expr:"auto_merge_strategy" graphql:"autoMergeStrategy"`
	Conflicts                     bool                          `expr:"conflicts" graphql:"conflicts"`
	CreatedAt                     time.Time                     `expr:"created_at" graphql:"createdAt"`
	Description                   string                        `expr:"description"`
	DiffStats                     []ContextMergeRequestDiffStat `expr:"diff_stats"`
	DivergedFromTargetBranch      bool                          `expr:"diverged_from_target_branch" graphql:"divergedFromTargetBranch"`
	Draft                         bool                          `expr:"draft"`
	FirstCommit                   *ContextCommit                `expr:"first_commit" graphql:"-"`
	ID                            string                        `expr:"id" graphql:"id"`
	IID                           string                        `expr:"iid" graphql:"iid"`
	Labels                        []ContextLabel                `expr:"labels" graphql:"-"`
	LastCommit                    *ContextCommit                `expr:"last_commit" graphql:"-"`
	Mergeable                     bool                          `expr:"mergeable" graphql:"mergeable"`
	MergedAt                      *time.Time                    `expr:"merged_at" graphql:"mergedAt"`
	MergeStatusEnum               string                        `expr:"merge_status_enum" graphql:"mergeStatusEnum"`
	SourceBranch                  string                        `expr:"source_branch" graphql:"sourceBranch"`
	SourceBranchExists            bool                          `expr:"source_branch_exists" graphql:"sourceBranchExists"`
	SourceBranchProtected         bool                          `expr:"source_branch_protected" graphql:"sourceBranchProtected"`
	Squash                        bool                          `expr:"squash" graphql:"squash"`
	SquashOnMerge                 bool                          `expr:"squash_on_merge" graphql:"squashOnMerge"`
	State                         string                        `expr:"state"`
	TargetBranch                  string                        `expr:"target_branch" graphql:"targetBranch"`
	TargetBranchExists            bool                          `expr:"target_branch_exists" graphql:"targetBranchExists"`
	TimeBetweenFirstAndLastCommit time.Duration                 `expr:"time_between_first_and_last_commit" graphql:"-"`
	TimeSinceFirstCommit          time.Duration                 `expr:"time_since_first_commit" graphql:"-"`
	TimeSinceLastCommit           time.Duration                 `expr:"time_since_last_commit" graphql:"-"`
	Title                         string                        `expr:"title"`
	UpdatedAt                     time.Time                     `expr:"updated_at" graphql:"updatedAt"`

	// Internal state
	ResponseLabels       *ContextLabelNodes  `expr:"-" graphql:"labels(first: 200)" json:"-" yaml:"-"`
	ResponseFirstCommits *ContextCommitsNode `expr:"-" graphql:"first_commit: commits(first:1)"`
	ResponseLastCommits  *ContextCommitsNode `expr:"-" graphql:"last_commit: commits(last:1)"`
}

// Partially lifted from https://github.com/hmarr/codeowners/blob/main/match.go
func (e ContextMergeRequest) ModifiedFiles(patterns ...string) bool {
	leftAnchoredLiteral := false

	for _, pattern := range patterns {
		if !strings.ContainsAny(pattern, "*?\\") && pattern[0] == '/' {
			leftAnchoredLiteral = true
		}

		regex, err := buildPatternRegex(pattern)
		if err != nil {
			panic(err)
		}

		for _, changedFile := range e.DiffStats {
			// Normalize Windows-style path separators to forward slashes
			testPath := filepath.ToSlash(changedFile.Path)

			if leftAnchoredLiteral {
				prefix := pattern

				// Strip the leading slash as we're anchored to the root already
				if prefix[0] == '/' {
					prefix = prefix[1:]
				}

				// If the pattern ends with a slash we can do a simple prefix match
				if prefix[len(prefix)-1] == '/' && strings.HasPrefix(testPath, prefix) {
					return true
				}

				// If the strings are the same length, check for an exact match
				if len(testPath) == len(prefix) && testPath == prefix {
					return true
				}

				// Otherwise check if the test path is a subdirectory of the pattern
				if len(testPath) > len(prefix) && testPath[len(prefix)] == '/' && testPath[:len(prefix)] == prefix {
					return true
				}
			}

			if regex.MatchString(testPath) {
				return true
			}
		}
	}

	return false
}

// buildPatternRegex compiles a new regexp object from a gitignore-style pattern string
func buildPatternRegex(pattern string) (*regexp.Regexp, error) {
	// Handle specific edge cases first
	switch {
	case strings.Contains(pattern, "***"):
		return nil, fmt.Errorf("pattern cannot contain three consecutive asterisks")

	case pattern == "":
		return nil, fmt.Errorf("empty pattern")

	// "/" doesn't match anything
	case pattern == "/":
		return regexp.Compile(`\A\z`)
	}

	segs := strings.Split(pattern, "/")

	if segs[0] == "" {
		// Leading slash: match is relative to root
		segs = segs[1:]
	} else {
		// No leading slash - check for a single segment pattern, which matches
		// relative to any descendent path (equivalent to a leading **/)
		if len(segs) == 1 || (len(segs) == 2 && segs[1] == "") {
			if segs[0] != "**" {
				segs = append([]string{"**"}, segs...)
			}
		}
	}

	if len(segs) > 1 && segs[len(segs)-1] == "" {
		// Trailing slash is equivalent to "/**"
		segs[len(segs)-1] = "**"
	}

	sep := "/"

	lastSegIndex := len(segs) - 1
	needSlash := false

	var re strings.Builder
	re.WriteString(`\A`)

	for i, seg := range segs {
		switch seg {
		case "**":
			switch {
			// If the pattern is just "**" we match everything
			case i == 0 && i == lastSegIndex:
				re.WriteString(`.+`)

			// If the pattern starts with "**" we match any leading path segment
			case i == 0:
				re.WriteString(`(?:.+` + sep + `)?`)
				needSlash = false

			// If the pattern ends with "**" we match any trailing path segment
			case i == lastSegIndex:
				re.WriteString(sep + `.*`)

			// If the pattern contains "**" we match zero or more path segments
			default:
				re.WriteString(`(?:` + sep + `.+)?`)
				needSlash = true
			}

		case "*":
			if needSlash {
				re.WriteString(sep)
			}

			// Regular wildcard - match any characters except the separator
			re.WriteString(`[^` + sep + `]+`)
			needSlash = true

		default:
			if needSlash {
				re.WriteString(sep)
			}

			escape := false
			for _, ch := range seg {
				if escape {
					escape = false
					re.WriteString(regexp.QuoteMeta(string(ch)))

					continue
				}

				// Other pathspec implementations handle character classes here (e.g.
				// [AaBb]), but CODEOWNERS doesn't support that so we don't need to
				switch ch {
				case '\\':
					escape = true

				// Multi-character wildcard
				case '*':
					re.WriteString(`[^` + sep + `]*`)

				// Single-character wildcard
				case '?':
					re.WriteString(`[^` + sep + `]`)

				// Regular character
				default:
					re.WriteString(regexp.QuoteMeta(string(ch)))
				}
			}

			if i == lastSegIndex {
				// As there's no trailing slash (that'd hit the '**' case), we
				// need to match descendent paths
				re.WriteString(`(?:` + sep + `.*)?`)
			}

			needSlash = true
		}
	}

	re.WriteString(`\z`)

	return regexp.Compile(re.String())
}
