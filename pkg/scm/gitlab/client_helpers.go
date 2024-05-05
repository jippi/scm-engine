package gitlab

import (
	"fmt"
	"strconv"

	"github.com/jippi/gitlab-labeller/pkg/scm"
	go_gitlab "github.com/xanzy/go-gitlab"
)

// Helper function to accept and format both the project ID or name as project
// identifier for all API calls.
func ParseID(id interface{}) (string, error) { //nolint:varnamelen
	switch v := id.(type) {
	case int:
		return strconv.Itoa(v), nil

	case string:
		return v, nil

	default:
		return "", fmt.Errorf("invalid ID type %#v, the ID must be an int or a string", id)
	}
}

// Convert a GitLab native response to a SCM agnostic one
func convertResponse(upstream *go_gitlab.Response) *scm.Response {
	if upstream == nil {
		return nil
	}

	return &scm.Response{
		Response: upstream.Response,
		// Fields used for offset-based pagination.
		// TotalItems:   upstream.TotalItems,
		// TotalPages:   upstream.TotalPages,
		// ItemsPerPage: upstream.ItemsPerPage,
		// CurrentPage:  upstream.CurrentPage,
		NextPage: upstream.NextPage,
		// PreviousPage: upstream.PreviousPage,

		// Fields used for keyset-based pagination.
		// PreviousLink: upstream.PreviousLink,
		// NextLink:     upstream.NextLink,
		// FirstLink:    upstream.FirstLink,
		// LastLink:     upstream.LastLink,
	}
}
