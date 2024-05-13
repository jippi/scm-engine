# Commands

## `scm-engine`

```plain
--8<-- "docs/gitlab/_partials/cmd-root.md"
```

## `scm-engine gitlab`

```plain
--8<-- "docs/gitlab/_partials/cmd-gitlab.md"
```

## `scm-engine gitlab evaluate`

```plain
--8<-- "docs/gitlab/_partials/cmd-gitlab-evaluate.md"
```

## `scm-engine gitlab server`

Point your GitLab webhook at the `/gitlab` endpoint.

Support the following events, and they will both trigger an Merge Request `evaluation`

- [`Comments`](https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html#comment-events) - A comment is made or edited on an issue or merge request.
- [`Merge request events`](https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html#merge-request-events) - A merge request is created, updated, or merged.

!!! tip

    You have access to the raw webhook event payload via `webhook_event.*` fields in Expr script fields when using `server` mode. See the [GitLab Webhook Events documentation](https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html) for available fields.

```plain
--8<-- "docs/gitlab/_partials/cmd-gitlab-server.md"
```
