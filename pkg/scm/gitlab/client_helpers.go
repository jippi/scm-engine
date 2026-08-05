package gitlab

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jippi/scm-engine/pkg/scm"
	go_gitlab "gitlab.com/gitlab-org/api/client-go"
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

// ParseProjectName returns the project name from a project slug.
//
// The project slug must be in the format of group/project.
func ParseProjectName(id interface{}) (string, error) {
	projectID, err := ParseID(id)
	if err != nil {
		return "", err
	}

	projectParts := strings.Split(projectID, "/")
	if len(projectParts) < 2 {
		return "", fmt.Errorf("invalid project ID %s, the ID must be a group/project", projectID)
	}

	// the project name is the last part of the project ID.
	projectName := projectParts[len(projectParts)-1]

	return projectName, nil
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
