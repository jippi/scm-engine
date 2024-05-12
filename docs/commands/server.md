---
hide:
  - toc
---

# Server

Point your GitLab webhook at the `/gitlab` endpoint.

Support the following events, and they will both trigger an Merge Request `evaluation`

- [`Comments`](https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html#comment-events) - A comment is made or edited on an issue or merge request.
- [`Merge request events`](https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html#merge-request-events) - A merge request is created, updated, or merged.

!!! tip

    You have access to the raw webhook event payload via `webhook_event.*` fields in Expr script fields when using `server` mode. See the [GitLab Webhook Events documentation](https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html) for available fields.

```plain
NAME:
   scm-engine server - Start HTTP server for webhook event driven usage

USAGE:
   scm-engine server [command options]

OPTIONS:
   --webhook-secret value  Used to validate received payloads. Sent with the request in the X-Gitlab-Token HTTP header [$SCM_ENGINE_WEBHOOK_SECRET]
   --listen value          IP + Port that the HTTP server should listen on (default: "0.0.0.0:3000") [$SCM_ENGINE_LISTEN]
   --update-pipeline       Update the CI pipeline status with progress (default: true) [$SCM_ENGINE_UPDATE_PIPELINE]
   --help, -h              show help

GLOBAL OPTIONS:
   --config value     Path to the scm-engine config file (default: ".scm-engine.yml") [$SCM_ENGINE_CONFIG_FILE]
   --provider value   SCM provider to use. Must be either "github" or "gitlab". SCM Engine will automatically detect "github" if "GITHUB_ACTIONS" environment variable is set (e.g., inside GitHub Actions) and detect "gitlab" if "GITLAB_CI" environment variable is set (e.g., inside GitLab CI). [$SCM_ENGINE_PROVIDER]
   --api-token value  GitHub/GitLab API token [$SCM_ENGINE_TOKEN, $GITHUB_TOKEN]
   --base-url value   Base URL for the SCM instance (default: "https://gitlab.com/") [$GITLAB_BASEURL, $CI_SERVER_URL, $GITHUB_API_URL]
   --dry-run          Dry run, don't actually _do_ actions, just print them (default: false)
   --help, -h         show help
   --version, -v      print the version
```
