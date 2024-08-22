# Configuration file

The default configuration filename is `.scm-engine.yml`, either in current working directory, or if you are in a Git repository, the root of the project.

The file path can be changed via `--config` CLI flag and `#!css $SCM_ENGINE_CONFIG_FILE` environment variable.

## `ignore_activity_from` {#actions data-toc-label="ignore_activity_from"}

!!! question "What are activity?"

  SCM-Engine defines activity as comments, reviews, commits, adding/removing labels and similar.

  *Generally*, `activity` is what you see in the Merge/Pull Request `timeline` in the browser UI.

Configure what users that should be ignored when considering activity on a Merge Request

### `ignore_activity_from.bots` {#actions data-toc-label="bots"}

Should `bot` users be ignored when considering activity? Default: `false`

### `ignore_activity_from.usernames[]` {#actions data-toc-label="usernames"}

A list of usernames that should be ignored when considering user activity. Default: `[]`

### `ignore_activity_from.emails[]` {#actions data-toc-label="emails"}

A list of emails that should be ignored when considering user activity. Default: `[]`

**NOTE:** If a user do not have a public email configured on their profile, that users activity will never match this rule.

## `actions[]` {#actions data-toc-label="actions"}

!!! question "What are actions?"

    Actions can [modify a Merge Request](#actions.if.then.action) in various ways, for example, adding a comment or closing the Merge Request.

    Due to actions powerful and flexible capabilities, they can be a bit harder to get *right* than [adding and removing labels](#label).

    Please see the [examples page](examples.md) for use-cases.

The `#!css actions` key is a list of actions that can be taken on a Merge Request.

### `actions[].name` {#actions.name data-toc-label="name"}

The name of the action, this is purely for debugging and your convenience.

It's encouraged to be descriptive of the actions.

### `actions[].if` {#actions.if data-toc-label="if"}

--8<-- "docs/_partials/expr-lang-info.md"

!!! tip "The script must return a `#!css boolean`"

A key controlling if the action should executed or not.

### `actions[].if.then[]` {#actions.if.then data-toc-label="then"}

The list of operations to take if the [`#!css action.if`](#actions.if) returned `true`.

#### `actions[].if.then[].action` {#actions.if.then.action data-toc-label="action"}

This key controls what kind of action that should be taken.

- `#!yaml approve` to approve the Merge Request.
- `#!yaml unapprove` to approve the Merge Request.
- `#!yaml close` to close the Merge Request.
- `#!yaml reopen` to reopen the Merge Request.
- `#!yaml comment` to add a comment to the Merge Request

      *Additional fields:*

      - (required) `#!css message` The message that will be commented on the Merge Request.

      ```{.yaml title="'comment' example"}
      - action: comment
        message: |
          Hello world
      ```

- `#!yaml lock_discussion` to prevent further discussions on the Merge Request.
- `#!yaml unlock_discussion` to allow discussions on the Merge Request.
- `#!yaml add_label` to add *an existing* label to the Merge Request

      *Additional fields:*

      - (required) `#!css label` The label name to add.

      ```{.yaml title="add_label example"}
      - action: add_label
        label: example
      ```

- `#!yaml remove_label` to remove a label from the Merge Request

      *Additional fields:*

      - (required) `#!css label` The label name to add.

      ```{.yaml title="remove_label example"}
      - action: remove_label
        label: example
      ```

- `#!yaml update_description` updates the Merge Request Description

      *Additional fields:*

      - (required) `#!css replace` A list of key/value pairs to replace in the description. The `key` is the raw string to replace in the Merge Request description. The `value` is an Expr Lang expression returning a `string` that `key` will be replaced with - all Script Attributes and Script Functions are available within the script.

      ```{.yaml title="update_description example"}
      - action: update_description
        replace:
          "${{CI_MERGE_REQUEST_IID}}": "merge_request.iid"
      ```

## `label[]` {#label data-toc-label="label"}

!!! question "What are labels?"

    Labels are a way to categorize and filter issues, merge requests, and epics in GitLab.
      -- *[GitLab documentation](https://docs.gitlab.com/ee/user/project/labels.html){target="_blank"}*

The `#!css label` key is a list of the labels you want to manage.

These keys are shared between the [`#!yaml conditional`](#label.strategy-conditional) and [`#!yaml generate`](#label.strategy-generate) label strategy. (more above these below!)

### `label[].strategy` {#label.strategy data-toc-label="strategy"}

SCM Engine supports two strategies for managing labels, each changes the behavior of the [`#!css script`](#label.script).

- `#!yaml conditional` (default, if `#!css strategy` key is omitted), where you provide the `#!css name` of the label, and a [`#!css script`](#label.script) that returns a boolean for wether the label should be added to the Merge Request.

    The [`#!css script`](#label.script) must return a `#!yaml boolean` value, where `#!yaml true` mean `add the label` and `#!yaml false` mean `remove the label`.

- `#!yaml generate`, where your `#!css script` generates the list of labels that should be added to the Merge Request.

    The [`#!css script`](#label.script) must return a `list of strings`, where each label returned will be added to the Merge Request.

#### `label[].strategy = conditional` {#label.strategy-conditional data-toc-label="conditional"}

Use the `#!yaml conditional` strategy when you want to add/remove a label on a Merge Request depending on *something*. It's the default strategy, and the most simple one to use.

!!! example "Please see the [*Add label if a file extension is modified*](./gitlab/examples.md#add-label-if-a-file-extension-is-modified) example for how to use this"

#### `label[].strategy = generate` {#label.strategy-generate data-toc-label="generate"}

Use the [`#!yaml generate`](#label.strategy) strategy if you want to create dynamic labels, for example, depending labels based on the file structure within your project.

Thanks to the dynamic nature of the `#!yaml generate` strategy, it has fantastic flexibility, at the cost of greater complexity.

!!! example "Please see the [*generate labels from directory layout*](./gitlab/examples.md#generate-labels-via-script) example for how to use this"

### `label[].name` {#label.name data-toc-label="name"}

- When using `#!yaml label.strategy: conditional`

    **REQUIRED** The `#!css name` of the label to create.

- When using `#!yaml label.strategy: generate`

    **OMITTED** The `#!css name` field must not be set when using the `#!yaml generate` strategy.

### `label[].script` {#label.script data-toc-label="script"}

--8<-- "docs/_partials/expr-lang-info.md"

Depending on the `#!yaml label.strategy:` used, the behavior of the script changes, read more about this below.

### `label[].color` {#label.color data-toc-label="color"}

!!! note

    When used on `#!yaml strategy: generate` labels, all generated labels will have the same color.

`#!css color` is a mandatory field, controlling the background color of the label when viewed in the User Interface.

You can either provide your own `#!yaml hex` value (e.g, `#FFFF00`) or use the [Twitter Bootstrap color variables](https://getbootstrap.com/docs/5.3/customize/color/#all-colors), for example`#!yaml $blue-500` and `#!yaml $teal`.

### `label[].description` {#label.description data-toc-label="description"}

!!! info "When used on [`#!yaml strategy: generate`](#label.strategy-generate) labels, all generated labels will have the same description."

An *optional* key that sets the description field for the label within GitLab.

Descriptions are shown in the User Interface when you hover any label.

### `label[].priority` {#label.priority data-toc-label="priority"}

!!! info "When used on [`#!yaml strategy: generate`](#label.strategy-generate) labels, all generated labels will have the same priority."

An *optional* key that controls the [GitLab Label Priority](https://docs.gitlab.com/ee/user/project/labels.html#set-label-priority){target="_blank"}.

### `label[].skip_if` {#label.skip_if data-toc-label="skip_if"}

--8<-- "docs/_partials/expr-lang-info.md"

!!! tip "The script must return a `boolean` value"

An optional key controlling if the label should be skipped (meaning no removal or adding of labels).
