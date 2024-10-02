package gitlab

// PeriodicEvaluationResult structs maps to the GraphQL query used to find Merge Requests
// that should be periodically evaluated.
//
// GraphQL query:
//
//	query (
//	  $project_topics: [String!],
//	  $config_file: String!,
//	  $project_membership: Boolean,
//	  $mr_ignore_labels: [String!],
//	  $mr_require_labels: [String!]
//	) {
//	  projects(
//	    first: 100
//	    membership: $project_membership
//	    withMergeRequestsEnabled: true
//	    topics: $project_topics
//	  ) {
//	    nodes {
//	      fullPath
//	      repository {
//	        blobs(paths: [$config_file]) {
//	          nodes {
//	            rawBlob
//	          }
//	        }
//	      }
//	      mergeRequests(
//	        first: 100,
//	        state: opened,
//	        not: {labels: $mr_ignore_labels},
//	        labels: $mr_require_labels,
//	        sort: UPDATED_ASC
//	      ) {
//	        nodes {
//	          iid
//	          diffHeadSha
//
//	          headPipeline {
//	            status
//	          }
//	        }
//	      }
//	    }
//	  }
//	}
//
// Query Variables
//
//	{
//	  "config_file": ".scm-engine.yml",
//	  "project_topics": ["scm-engine"],
//	  "project_membership": true,
//	  "mr_ignore_labels": ["security", "do-not-close"],
//	  "mr_require_labels": null
//	}
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
	Blobs graphqlNodesOf[BlobNode] `graphql:"blobs(paths: [$scm_config_file_path])"`
}

type PeriodicEvaluationMergeRequestNode struct {
	IID          string        `graphql:"iid"`
	SHA          string        `graphql:"diffHeadSha"`
	HeadPipeline *PipelineNode `graphql:"headPipeline"`
}

type PipelineNode struct {
	Status string `graphql:"status"`
}

type BlobNode struct {
	Path string `graphql:"path"`
	Blob string `graphql:"rawBlob"`
}

type graphqlNodesOf[T any] struct {
	Nodes []T `graphql:"nodes"`
}

// IncludeConfigurationResult is the GraphQL response for downloading
// a list of configuration files from a project repository within GitLab
//
// GraphQL query:
//
//	query ($project: ID!, $ref: String ="HEAD", $files: [String!]!) {
//	  project(fullPath: $project) {
//	    repository {
//	      blobs(paths:$files, ref: $ref, first: 100) {
//	        nodes {
//	          path
//	          rawBlob
//	        }
//	      }
//	    }
//	  }
//	}
//
// Query Variables
//
//	{
//	   "project": "platform/scm-engine-library",
//	   "files": ["label/change-type.yml", "label/last-commit-age.yml", "label/need-rebase.yml", "life-cycle/close-merge-request-3-weeks.yml"]
//	}
type IncludeConfigurationResult struct {
	Project IncludeConfigurationProject `graphql:"project(fullPath: $project)"`
}

type IncludeConfigurationProject struct {
	Repository IncludeConfigurationRepository `graphql:"repository"`
}

type IncludeConfigurationRepository struct {
	// Blobs contains a single (optional) node with the content of the ".scm-config.yml" file
	// read from the projects default branch at the time of reading
	Blobs graphqlNodesOf[BlobNode] `graphql:"blobs(paths: $files, ref: $ref, first: 100)"`
}
