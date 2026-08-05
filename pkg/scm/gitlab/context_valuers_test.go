package gitlab_test

import (
	"testing"

	"github.com/jippi/scm-engine/pkg/scm/gitlab"
	"github.com/stretchr/testify/require"
)

// The AsString valuers are hand-written glue that lets expr-lang compare a
// generated enum against a plain string in a user script, e.g.
// merge_request.merge_status == "can_be_merged". Without them the comparison
// silently fails to match.
func TestEnumValuers(t *testing.T) {
	t.Parallel()

	require.Equal(t, "opened", gitlab.MergeRequestState("opened").AsString())
	require.Equal(t, "active", gitlab.UserState("active").AsString())
	require.Equal(t, "can_be_merged", gitlab.MergeStatus("can_be_merged").AsString())
	require.Equal(t, "mergeable", gitlab.DetailedMergeStatus("mergeable").AsString())
	require.Equal(t, "success", gitlab.PipelineStatusEnum("success").AsString())
}
