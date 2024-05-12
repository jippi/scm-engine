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
   --update-pipeline  Update the CI pipeline status with progress (default: false) [$SCM_ENGINE_UPDATE_PIPELINE]
   --project value    GitLab project (example: 'gitlab-org/gitlab') [$GITLAB_PROJECT, $CI_PROJECT_PATH, $GITHUB_REPOSITORY]
   --id value         The pull/merge ID to process, if not provided as a CLI flag [$CI_MERGE_REQUEST_IID]
   --commit value     The git commit sha [$CI_COMMIT_SHA, $GITHUB_SHA]
   --help, -h         show help

GLOBAL OPTIONS:
   --config value     Path to the scm-engine config file (default: ".scm-engine.yml") [$SCM_ENGINE_CONFIG_FILE]
   --provider value   SCM provider to use. Must be either "github" or "gitlab". SCM Engine will automatically detect "github" if "GITHUB_ACTIONS" environment variable is set (e.g., inside GitHub Actions) and detect "gitlab" if "GITLAB_CI" environment variable is set (e.g., inside GitLab CI). [$SCM_ENGINE_PROVIDER]
   --api-token value  GitHub/GitLab API token [$SCM_ENGINE_TOKEN, $GITHUB_TOKEN]
   --base-url value   Base URL for the SCM instance (default: "https://gitlab.com/") [$GITLAB_BASEURL, $CI_SERVER_URL, $GITHUB_API_URL]
   --dry-run          Dry run, don't actually _do_ actions, just print them (default: false)
   --help, -h         show help
   --version, -v      print the version
```
