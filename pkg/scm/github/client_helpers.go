package github

import (
	"context"
	"strings"

	go_github "github.com/google/go-github/v67/github"
	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/state"
)

// Convert a GitLab native response to a SCM agnostic one
func convertResponse(upstream *go_github.Response) *scm.Response {
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

func ownerAndRepo(ctx context.Context) (string, string) {
	project := state.ProjectID(ctx)
	chunks := strings.Split(project, "/")

	return chunks[0], chunks[1]
}
