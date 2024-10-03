package gitlab

import (
	"testing"

	"github.com/jippi/scm-engine/pkg/scm"
	"github.com/stretchr/testify/assert"
)

func TestGetCodeOwners(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mergeRequest   *ContextMergeRequest
		expectedOwners scm.Actors
	}{
		{
			name: "No approval rules",
			mergeRequest: &ContextMergeRequest{
				ApprovalState: &ContextApprovalState{
					Rules: []ContextApprovalRule{},
				},
			},
			expectedOwners: scm.Actors{},
		},
		{
			name: "Approval rules without code owners",
			mergeRequest: &ContextMergeRequest{
				ApprovalState: &ContextApprovalState{
					Rules: []ContextApprovalRule{
						{
							Type: scm.Ptr(ApprovalRuleTypeAnyApprover),
							EligibleApprovers: []ContextUser{
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
			mergeRequest: &ContextMergeRequest{
				ApprovalState: &ContextApprovalState{
					Rules: []ContextApprovalRule{
						{
							Type: scm.Ptr(ApprovalRuleTypeCodeOwner),
							EligibleApprovers: []ContextUser{
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
			mergeRequest: &ContextMergeRequest{
				ApprovalState: &ContextApprovalState{
					Rules: []ContextApprovalRule{
						{
							Type: scm.Ptr(ApprovalRuleTypeCodeOwner),
							EligibleApprovers: []ContextUser{
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

			ctx := &Context{
				MergeRequest: tt.mergeRequest,
			}
			owners := ctx.GetCodeOwners()
			assert.Equal(t, tt.expectedOwners, owners)
		})
	}
}
