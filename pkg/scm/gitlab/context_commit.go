package gitlab

import "time"

type ContextCommit struct {
	AuthorEmail   string    `expr:"author_email"   graphql:"authorEmail"`
	CommittedDate time.Time `expr:"committed_date" graphql:"committedDate"`
}

type ContextCommitsNode struct {
	Nodes []*ContextCommit `graphql:"nodes"`
}
