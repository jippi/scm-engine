package gitlab

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"
)

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
		return nil, errors.New("pattern cannot contain three consecutive asterisks")

	case pattern == "":
		return nil, errors.New("empty pattern")

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

	var regexString strings.Builder

	regexString.WriteString(`\A`)

	for i, seg := range segs {
		switch seg {
		case "**":
			switch {
			// If the pattern is just "**" we match everything
			case i == 0 && i == lastSegIndex:
				regexString.WriteString(`.+`)

			// If the pattern starts with "**" we match any leading path segment
			case i == 0:
				regexString.WriteString(`(?:.+` + sep + `)?`)

				needSlash = false

			// If the pattern ends with "**" we match any trailing path segment
			case i == lastSegIndex:
				regexString.WriteString(sep + `.*`)

			// If the pattern contains "**" we match zero or more path segments
			default:
				regexString.WriteString(`(?:` + sep + `.+)?`)

				needSlash = true
			}

		case "*":
			if needSlash {
				regexString.WriteString(sep)
			}

			// Regular wildcard - match any characters except the separator
			regexString.WriteString(`[^` + sep + `]+`)

			needSlash = true

		default:
			if needSlash {
				regexString.WriteString(sep)
			}

			escape := false

			for _, char := range seg {
				if escape {
					escape = false

					regexString.WriteString(regexp.QuoteMeta(string(char)))

					continue
				}

				// Other pathspec implementations handle character classes here (e.g.
				// [AaBb]), but CODEOWNERS doesn't support that so we don't need to
				switch char {
				case '\\':
					escape = true

				// Multi-character wildcard
				case '*':
					regexString.WriteString(`[^` + sep + `]*`)

				// Single-character wildcard
				case '?':
					regexString.WriteString(`[^` + sep + `]`)

				// Regular character
				default:
					regexString.WriteString(regexp.QuoteMeta(string(char)))
				}
			}

			if i == lastSegIndex {
				// As there's no trailing slash (that'd hit the '**' case), we
				// need to match descendent paths
				regexString.WriteString(`(?:` + sep + `.*)?`)
			}

			needSlash = true
		}
	}

	regexString.WriteString(`\z`)

	return regexp.Compile(regexString.String())
}
