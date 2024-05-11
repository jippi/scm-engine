# Setup

## Webhook Server

Using `scm-engine` as a webhook server allows for richer feature set compared to [GitLab CI pipeline](#gitlab-ci-pipeline) mode

- `+` Reacting to comments
- `+` Access to webhook event data in scripts via `webhook_event.*` (see [server docs](#server) for more information)
- `+` A single `scm-engine` instance (and single token) for your GitLab project, group, or instance depending on where you configure the webhook.
- `+` Each Project still have their own `.scm-engine.yml` file, it's downloaded via the API when the server is processing a webhook event.
- `+` A single "bot" identity across your projects.
- `+` Turn key once configured; if a project want to use `scm-engine` they just need to create the `.scm-engine.yml` file in their project.
- `+` Real-time reactions to changes
- `-` No intuitive access to [`evaluation` logs](#evaluate) within GitLab (you can see them in the server logs or in the webhook failure log)

**Setup**:

1. Deploy `scm-engine` within your infrastructure in an environment that can communicate egress/ingress with GitLab. ([see `server`](#server))
1. Configure your `webhook` at Project, Group, or Server level to hit the `/gitlab` endpoint on the `scm-engine` server endpoint. ([see `server`](#server))

## GitLab-CI pipeline

Using `scm-engine` within a GitLab CI pipeline is straight forward - every time a CI pipeline runs, `scm-engine` will [evaluate](#evaluate) the Merge Request.

- `+` Simple & quick installation.
- `+` Limited access token permissions.
- `+` Easy access to [`evaluation` logs](#evaluate) within the GitLab CI job.
- `-` Can't react to comments; only works within a CI pipeline.
- `-` Higher latency for reacting to changes depending on how fast CI jobs run (and where in the pipeline it runs).

**Setup**:

1. Add a `.scm-engine.yml` file in the root of your project.
1. Create a [CI/CD Variable](https://docs.gitlab.com/ee/ci/variables/#for-a-group)
    1. Name must be `SCM_ENGINE_TOKEN`
    1. Value must a [Project Access Token](https://docs.gitlab.com/ee/user/project/settings/project_access_tokens.html)
        1. Must have `api` scope.
        1. Must have `developer` or `maintainer` role access so it can edit Merge Requests.
    1. `Mask` **should** be checked.
    1. `Protected` **should NOT** be checked.
    1. `Expand variable reference` **should NOT** be checked.
1. Setup a CI job using the `scm-engine` Docker image that will run when a pipeline is created from a Merge Request Event.

    ```yaml
    scm-engine::evaluate::on-merge-request-event:
      image: ghcr.io/jippi/scm-engine:latest
      rules:
        - if: $CI_PIPELINE_SOURCE == 'merge_request_event'
      script:
        - scm-engine evaluate

    scm-engine::evaluate::on-schedule:
      image: ghcr.io/jippi/scm-engine:latest
      rules:
        - if: $CI_PIPELINE_SOURCE == "schedule"
      script:
        - scm-engine evaluate all
    ```

1. Done! Every Merge Request change should now re-run scm-engine and apply your label rules
