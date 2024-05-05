package scm

import (
	"net/http"

	"github.com/jippi/gitlab-labeller/pkg/types"
)

// Label represents a GitLab label.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/labels.html
type Label struct {
	ID                     int              `json:"id"`
	Name                   string           `json:"name"`
	Color                  string           `json:"color"`
	TextColor              string           `json:"text_color"`
	Description            string           `json:"description"`
	OpenIssuesCount        int              `json:"open_issues_count"`
	ClosedIssuesCount      int              `json:"closed_issues_count"`
	OpenMergeRequestsCount int              `json:"open_merge_requests_count"`
	Subscribed             bool             `json:"subscribed"`
	Priority               types.Value[int] `json:"priority"`
	IsProjectLabel         bool             `json:"is_project_label"`
}

// CreateLabelOptions represents the available CreateLabel() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/labels.html#create-a-new-label
type CreateLabelOptions struct {
	Name        *string          `url:"name,omitempty" json:"name,omitempty"`
	Color       *string          `url:"color,omitempty" json:"color,omitempty"`
	Description *string          `url:"description,omitempty" json:"description,omitempty"`
	Priority    types.Value[int] `url:"priority,omitempty" json:"priority"`
}

type UpdateLabelOptions struct {
	Name        *string          `url:"name,omitempty" json:"name,omitempty"`
	NewName     *string          `url:"new_name,omitempty" json:"new_name,omitempty"`
	Color       *string          `url:"color,omitempty" json:"color,omitempty"`
	Description *string          `url:"description,omitempty" json:"description,omitempty"`
	Priority    types.Value[int] `url:"priority,omitempty" json:"priority"`
}

// LabelOptions is a custom type with specific marshaling characteristics.
type LabelOptions []string

// UpdateMergeRequestOptions represents the available UpdateMergeRequest()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#update-mr
type UpdateMergeRequestOptions struct {
	Title              *string       `url:"title,omitempty" json:"title,omitempty"`
	Description        *string       `url:"description,omitempty" json:"description,omitempty"`
	TargetBranch       *string       `url:"target_branch,omitempty" json:"target_branch,omitempty"`
	AssigneeID         *int          `url:"assignee_id,omitempty" json:"assignee_id,omitempty"`
	AssigneeIDs        *[]int        `url:"assignee_ids,omitempty" json:"assignee_ids,omitempty"`
	ReviewerIDs        *[]int        `url:"reviewer_ids,omitempty" json:"reviewer_ids,omitempty"`
	Labels             *LabelOptions `url:"labels,comma,omitempty" json:"labels,omitempty"`
	AddLabels          *LabelOptions `url:"add_labels,comma,omitempty" json:"add_labels,omitempty"`
	RemoveLabels       *LabelOptions `url:"remove_labels,comma,omitempty" json:"remove_labels,omitempty"`
	MilestoneID        *int          `url:"milestone_id,omitempty" json:"milestone_id,omitempty"`
	StateEvent         *string       `url:"state_event,omitempty" json:"state_event,omitempty"`
	RemoveSourceBranch *bool         `url:"remove_source_branch,omitempty" json:"remove_source_branch,omitempty"`
	Squash             *bool         `url:"squash,omitempty" json:"squash,omitempty"`
	DiscussionLocked   *bool         `url:"discussion_locked,omitempty" json:"discussion_locked,omitempty"`
	AllowCollaboration *bool         `url:"allow_collaboration,omitempty" json:"allow_collaboration,omitempty"`
}

// ListLabelsOptions represents the available ListLabels() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/labels.html#list-labels
type ListLabelsOptions struct {
	ListOptions
	WithCounts            *bool   `url:"with_counts,omitempty" json:"with_counts,omitempty"`
	IncludeAncestorGroups *bool   `url:"include_ancestor_groups,omitempty" json:"include_ancestor_groups,omitempty"`
	Search                *string `url:"search,omitempty" json:"search,omitempty"`
}

// ListOptions specifies the optional parameters to various List methods that
// support pagination.
type ListOptions struct {
	// For offset-based paginated result sets, page of results to retrieve.
	Page int `url:"page,omitempty" json:"page,omitempty"`
	// For offset-based and keyset-based paginated result sets, the number of results to include per page.
	PerPage int `url:"per_page,omitempty" json:"per_page,omitempty"`

	// For keyset-based paginated result sets, name of the column by which to order
	OrderBy string `url:"order_by,omitempty" json:"order_by,omitempty"`
	// For keyset-based paginated result sets, the value must be `"keyset"`
	Pagination string `url:"pagination,omitempty" json:"pagination,omitempty"`
	// For keyset-based paginated result sets, sort order (`"asc"`` or `"desc"`)
	Sort string `url:"sort,omitempty" json:"sort,omitempty"`
}

// Response is a GitLab API response. This wraps the standard http.Response
// returned from GitLab and provides convenient access to things like
// pagination links.
type Response struct {
	*http.Response

	// Fields used for offset-based pagination.
	// TotalItems   int
	// TotalPages   int
	// ItemsPerPage int
	// CurrentPage  int
	NextPage int
	// PreviousPage int

	// Fields used for keyset-based pagination.
	// PreviousLink string
	// NextLink     string
	// FirstLink    string
	// LastLink     string
}

type EvaluationResult struct {
	// Name of the label being generated.
	//
	// May only be used with [conditional] labelling type
	Name string

	// Description for the label being generated.
	//
	// Optional; will be an empty string if omitted
	Description string

	// The HEX color code to use for the label.
	//
	// May use the color variables (e.g., $purple-300) defined in Twitter Bootstrap
	// https://getbootstrap.com/docs/5.3/customize/color/#all-colors
	Color string

	// Priority controls wether the label should be a priority label or regular one.
	//
	// This controls if the label is prioritized (sorted first) in the list.
	Priority types.Value[int]

	//
	Matched       bool
	CreateInGroup string
}

func (local EvaluationResult) EqualLabel(remote *Label) bool {
	if local.Name != remote.Name {
		return false
	}

	if local.Description != remote.Description {
		return false
	}

	if local.Color != remote.Color {
		return false
	}

	// Priority must agree on being NULL or not
	if local.Priority.Valid != remote.Priority.Valid {
		return false
	}

	// Priority must agree on their value
	if local.Priority.ValueOrZero() != remote.Priority.ValueOrZero() {
		return false
	}

	return true
}
