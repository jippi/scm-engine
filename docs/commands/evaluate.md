---
hide:
  - toc
---

# Evaluate

Evaluate the SCM engine rules against a specific Merge Request.

```plain
NAME:
   scm-engine evaluate - Evaluate a Merge Request

USAGE:
   scm-engine evaluate [command options] [mr_id, mr_id, ...]

OPTIONS:
   --project value    GitLab project (example: 'gitlab-org/gitlab') [$GITLAB_PROJECT, $CI_PROJECT_PATH]
   --id value         The pull/merge ID to process, if not provided as a CLI flag [$CI_MERGE_REQUEST_IID]
   --commit value     The git commit sha [$CI_COMMIT_SHA]
   --update-pipeline  Update the CI pipeline status with progress (default: false) [$SCM_ENGINE_UPDATE_PIPELINE]
   --help, -h         show help

GLOBAL OPTIONS:
   --config value     Path to the scm-engine config file (default: ".scm-engine.yml") [$SCM_ENGINE_CONFIG_FILE]
   --api-token value  GitHub/GitLab API token [$SCM_ENGINE_TOKEN]
   --base-url value   Base URL for the SCM instance (default: "https://gitlab.com/") [$GITLAB_BASEURL, $CI_SERVER_URL]
   --dry-run          Dry run, don't actually _do_ actions, just print them (default: false)
   --help, -h         show help
   --version, -v      print the version
```
