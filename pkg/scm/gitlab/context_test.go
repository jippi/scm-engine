package gitlab_test

import (
	"testing"

	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/jippi/scm-engine/pkg/scm/gitlab"
	"github.com/stretchr/testify/assert"
)

func TestGetCodeOwners(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mergeRequest   *gitlab.ContextMergeRequest
		expectedOwners scm.Actors
	}{
		{
			name: "No approval rules",
			mergeRequest: &gitlab.ContextMergeRequest{
				ApprovalState: &gitlab.ContextApprovalState{
					Rules: []gitlab.ContextApprovalRule{},
				},
			},
			expectedOwners: scm.Actors{},
		},
		{
			name: "Approval rules without code owners",
			mergeRequest: &gitlab.ContextMergeRequest{
				ApprovalState: &gitlab.ContextApprovalState{
					Rules: []gitlab.ContextApprovalRule{
						{
							Type: scm.Ptr(gitlab.ApprovalRuleTypeAnyApprover),
							EligibleApprovers: []gitlab.ContextUser{
								{Username: "user1"},
							},
						},
					},
				},
			},
			expectedOwners: scm.Actors{},
		},
		{
			name: "Approval rules with code owners",
			mergeRequest: &gitlab.ContextMergeRequest{
				ApprovalState: &gitlab.ContextApprovalState{
					Rules: []gitlab.ContextApprovalRule{
						{
							Type: scm.Ptr(gitlab.ApprovalRuleTypeCodeOwner),
							EligibleApprovers: []gitlab.ContextUser{
								{Username: "user1"},
								{Username: "user2", Bot: true}, // Should be ignored
								{Username: "user3"},
							},
						},
					},
				},
			},
			expectedOwners: scm.Actors{
				{Username: "user1"},
				{Username: "user3"},
			},
		},
		{
			name: "Duplicate code owners",
			mergeRequest: &gitlab.ContextMergeRequest{
				ApprovalState: &gitlab.ContextApprovalState{
					Rules: []gitlab.ContextApprovalRule{
						{
							Type: scm.Ptr(gitlab.ApprovalRuleTypeCodeOwner),
							EligibleApprovers: []gitlab.ContextUser{
								{Username: "user1"},
								{Username: "user1"}, // Duplicate, should be ignored
							},
						},
					},
				},
			},
			expectedOwners: scm.Actors{
				{Username: "user1"},
			},
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := &gitlab.Context{
				MergeRequest: tt.mergeRequest,
			}
			owners := ctx.GetCodeOwners()
			assert.Equal(t, tt.expectedOwners, owners)
		})
	}
}
