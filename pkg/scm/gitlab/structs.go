package gitlab

// PeriodicEvaluationResult structs maps to the GraphQL query used to find Merge Requests
// that should be periodically evaluated.
//
// GraphQL query:
//
//    query (
//      $project_topics: [String!],
//      $config_file: String!,
//      $project_membership: Boolean,
//      $mr_ignore_labels: [String!],
//      $mr_require_labels: [String!]
//    ) {
//      projects(
//        first: 100
//        membership: $project_membership
//        withMergeRequestsEnabled: true
//        topics: $project_topics
//      ) {
//        nodes {
//          fullPath
//          repository {
//            blobs(paths: [$config_file]) {
//              nodes {
//                rawBlob
//              }
//            }
//          }
//          mergeRequests(
//            first: 100,
//            state: opened,
//            not: {labels: $mr_ignore_labels},
//            labels: $mr_require_labels,
//            sort: UPDATED_ASC
//          ) {
//            nodes {
//              iid
//              diffHeadSha
//            }
//          }
//        }
//      }
//    }
//
// Query Variables
//
//    {
//      "config_file": ".scm-engine.yml",
//      "project_topics": ["scm-engine"],
//      "project_membership": true,
//      "mr_ignore_labels": ["security", "do-not-close"],
//      "mr_ignore_labels": []
//    }

type PeriodicEvaluationResult struct {
	// Projects contains first 100 projects that matches the filtering conditions
	Projects graphqlNodesOf[PeriodicEvaluationProjectNode] `graphql:"projects(first: 100, membership: $project_membership, withMergeRequestsEnabled: true, topics: $project_topics)"`
}

type PeriodicEvaluationProjectNode struct {
	// FullPath is the complete group + project slug / project identifier for a Project in GitLab
	FullPath string `graphql:"fullPath"`

	// MergeRequests contains up to 100 merge requests, sorted by oldest update/last change first
	MergeRequests graphqlNodesOf[PeriodicEvaluationMergeRequestNode] `graphql:"mergeRequests(first: 100, state: opened, not: {labels: $mr_ignore_labels}, labels: $mr_require_labels, sort: UPDATED_ASC)"`

	// Repository contains information about the git repository
	Repository PeriodicEvaluationRepository `graphql:"repository"`
}

type PeriodicEvaluationRepository struct {
	// Blobs contains a single (optional) node with the content of the ".scm-config.yml" file
	// read from the projects default branch at the time of reading
	Blobs graphqlNodesOf[PeriodicEvaluationBlobNode] `graphql:"blobs(paths: [$scm_config_file_path])"`
}

type PeriodicEvaluationMergeRequestNode struct {
	IID string `graphql:"iid"`
	SHA string `graphql:"diffHeadSha"`
}

type PeriodicEvaluationBlobNode struct {
	Blob string `graphql:"rawBlob"`
}

type graphqlNodesOf[T any] struct {
	Nodes []T `graphql:"nodes"`
}
