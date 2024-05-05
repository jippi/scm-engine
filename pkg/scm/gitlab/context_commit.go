package gitlab

import "time"

type ContextCommit struct {
	AuthorEmail   string    `graphql:"authorEmail" expr:"author_email"`
	CommittedDate time.Time `graphql:"committedDate" expr:"committed_date"`
}

type ContextCommitsNode struct {
	Nodes []*ContextCommit `graphql:"nodes"`
}
