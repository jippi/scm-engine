package gitlab

import "time"

type ContextProject struct {
	ID                string         `expr:"id" graphql:"id"`
	Name              string         `expr:"name" graphql:"name"`
	NameWithNamespace string         `expr:"name_with_namespace"`
	Description       string         `expr:"description"`
	Path              string         `expr:"path"`
	FullPath          string         `expr:"full_path" graphql:"fullPath"`
	Archived          bool           `expr:"archived" graphql:"archived"`
	Topics            []string       `expr:"topics" graphql:"topics"`
	Visibility        string         `expr:"visibility" graphql:"visibility"`
	Labels            []ContextLabel `expr:"labels" graphql:"-"`
	LastActivityAt    time.Time      `expr:"last_activity_at" graphql:"lastActivityAt"`
	CreatedAt         time.Time      `expr:"created_at" graphql:"createdAt"`
	UpdatedAt         time.Time      `expr:"updated_at" graphql:"updatedAt"`

	// Internal state
	MergeRequest   *ContextMergeRequest `expr:"-" graphql:"mergeRequest(iid: $mr_id)"`
	ResponseLabels *ContextLabelNodes   `expr:"-" graphql:"labels(first: 200)" json:"-" yaml:"-"`
	ResponseGroup  *ContextGroup        `expr:"-" graphql:"group" json:"-" yaml:"-"`
}
