package gitlab_test

import (
	"fmt"
	"testing"

	"github.com/hasura/go-graphql-client/pkg/jsonutil"
	"github.com/jippi/scm-engine/pkg/scm/gitlab"
	"github.com/stretchr/testify/require"
)

// gqlgen generates an UnmarshalGQL that rejects any enum value not declared in
// schema/gitlab.schema.graphqls, and the GraphQL client calls it while decoding
// the response. A value GitLab has added upstream but we have not declared
// therefore fails the decode of the *entire* merge request, not just the field.
//
// This guards the enums we read back from GitLab against drifting behind
// upstream. When it fails, add the missing value to the .graphqls file.
func TestEnumsDecodeEveryUpstreamValue(t *testing.T) {
	t.Parallel()

	// Values as of GitLab 18.x. Compare against
	// https://docs.gitlab.com/api/graphql/reference/#detailedmergestatus
	detailedMergeStatus := []string{
		"UNCHECKED", "CHECKING", "MERGEABLE", "COMMITS_STATUS",
		"CI_MUST_PASS", "CI_STILL_RUNNING", "DISCUSSIONS_NOT_RESOLVED",
		"DRAFT_STATUS", "NOT_OPEN", "NOT_APPROVED", "BLOCKED_STATUS",
		"EXTERNAL_STATUS_CHECKS", "PREPARING", "JIRA_ASSOCIATION",
		"CONFLICT", "NEED_REBASE", "REQUESTED_CHANGES", "APPROVALS_SYNCING",
		"LOCKED_PATHS", "LOCKED_LFS_FILES", "MERGE_TIME",
		"SECURITY_POLICIES_VIOLATIONS", "TITLE_NOT_MATCHING",
		"SECURITY_POLICY_PIPELINE_CHECK",
	}

	for _, value := range detailedMergeStatus {
		t.Run(value, func(t *testing.T) {
			t.Parallel()

			var query struct {
				DetailedMergeStatus *gitlab.DetailedMergeStatus `graphql:"detailedMergeStatus"`
			}

			payload := fmt.Sprintf(`{"detailedMergeStatus":%q}`, value)

			require.NoError(t, jsonutil.UnmarshalGraphQL([]byte(payload), &query))
			require.NotNil(t, query.DetailedMergeStatus)
			require.Equal(t, value, string(*query.DetailedMergeStatus))
		})
	}
}
