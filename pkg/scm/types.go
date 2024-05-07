package scm

import (
	"net/http"

	"github.com/jippi/scm-engine/pkg/types"
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
	Name        *string          `json:"name,omitempty"        url:"name,omitempty"`
	Color       *string          `json:"color,omitempty"       url:"color,omitempty"`
	Description *string          `json:"description,omitempty" url:"description,omitempty"`
	Priority    types.Value[int] `json:"priority"              url:"priority,omitempty"`
}

type UpdateLabelOptions struct {
	Name        *string          `json:"name,omitempty"        url:"name,omitempty"`
	NewName     *string          `json:"new_name,omitempty"    url:"new_name,omitempty"`
	Color       *string          `json:"color,omitempty"       url:"color,omitempty"`
	Description *string          `json:"description,omitempty" url:"description,omitempty"`
	Priority    types.Value[int] `json:"priority"              url:"priority,omitempty"`
}

// LabelOptions is a custom type with specific marshaling characteristics.
type LabelOptions []string

// UpdateMergeRequestOptions represents the available UpdateMergeRequest()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#update-mr
type UpdateMergeRequestOptions struct {
	Title              *string       `json:"title,omitempty"                url:"title,omitempty"`
	Description        *string       `json:"description,omitempty"          url:"description,omitempty"`
	TargetBranch       *string       `json:"target_branch,omitempty"        url:"target_branch,omitempty"`
	AssigneeID         *int          `json:"assignee_id,omitempty"          url:"assignee_id,omitempty"`
	AssigneeIDs        *[]int        `json:"assignee_ids,omitempty"         url:"assignee_ids,omitempty"`
	ReviewerIDs        *[]int        `json:"reviewer_ids,omitempty"         url:"reviewer_ids,omitempty"`
	Labels             *LabelOptions `json:"labels,omitempty"               url:"labels,comma,omitempty"`
	AddLabels          *LabelOptions `json:"add_labels,omitempty"           url:"add_labels,comma,omitempty"`
	RemoveLabels       *LabelOptions `json:"remove_labels,omitempty"        url:"remove_labels,comma,omitempty"`
	MilestoneID        *int          `json:"milestone_id,omitempty"         url:"milestone_id,omitempty"`
	StateEvent         *string       `json:"state_event,omitempty"          url:"state_event,omitempty"`
	RemoveSourceBranch *bool         `json:"remove_source_branch,omitempty" url:"remove_source_branch,omitempty"`
	Squash             *bool         `json:"squash,omitempty"               url:"squash,omitempty"`
	DiscussionLocked   *bool         `json:"discussion_locked,omitempty"    url:"discussion_locked,omitempty"`
	AllowCollaboration *bool         `json:"allow_collaboration,omitempty"  url:"allow_collaboration,omitempty"`
}

// ListLabelsOptions represents the available ListLabels() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/labels.html#list-labels
type ListLabelsOptions struct {
	ListOptions

	WithCounts            *bool   `json:"with_counts,omitempty"             url:"with_counts,omitempty"`
	IncludeAncestorGroups *bool   `json:"include_ancestor_groups,omitempty" url:"include_ancestor_groups,omitempty"`
	Search                *string `json:"search,omitempty"                  url:"search,omitempty"`
}

// ListOptions specifies the optional parameters to various List methods that
// support pagination.
type ListOptions struct {
	// For offset-based paginated result sets, page of results to retrieve.
	Page int `json:"page,omitempty" url:"page,omitempty"`
	// For offset-based and keyset-based paginated result sets, the number of results to include per page.
	PerPage int `json:"per_page,omitempty" url:"per_page,omitempty"`

	// For keyset-based paginated result sets, name of the column by which to order
	OrderBy string `json:"order_by,omitempty" url:"order_by,omitempty"`
	// For keyset-based paginated result sets, the value must be `"keyset"`
	Pagination string `json:"pagination,omitempty" url:"pagination,omitempty"`
	// For keyset-based paginated result sets, sort order (`"asc"`` or `"desc"`)
	Sort string `json:"sort,omitempty" url:"sort,omitempty"`
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

type EvaluationLabelResult struct {
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

	// Wether the evaluation rule matched positive (add label) or negative (remove label)
	Matched bool
}

type EvaluationActionStep map[string]any

type EvaluationActionResult struct {
	Name string `yaml:"name"`
	If   string `yaml:"if"`
	Then []EvaluationActionStep
}

func (local EvaluationLabelResult) IsEqual(remote *Label) bool {
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
