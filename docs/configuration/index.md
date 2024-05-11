# Options

The default configuration filename is `.scm-engine.yml`, either in current working directory, or if you are in a Git repository, the root of the project.

The file path can be changed via `--config` CLI flag and `$SCM_ENGINE_CONFIG_FILE` environment variable.

## `actions[]` {#actions data-toc-label="actions"}

The `actions` key is a list of actions that can be taken on a Merge Request.

### `actions[].name` {#actions.name data-toc-label="name"}

The name of the action, this is purely for debugging and your convenience. It's encouraged to be descriptive of the actions.

### `actions[].if` {#actions.if data-toc-label="if"}

A key controlling if the action should executed or not.

The `if` field must be a valid [Expr-lang](https://expr-lang.org/) expression returning a boolean.

### `actions[].if.then[]` {#actions.if.then data-toc-label="then"}

The list of operations to take if the `action.if` returned `true`.

#### `actions[].if.then[].action` {#actions.if.then.action data-toc-label="action"}

This key controls what kind of action that should be taken.

- `close` to close the Merge Request.
- `reopen` to reopen the Merge Request.
- `lock_discussion` to prevent further discussions on the Merge Request.
- `unlock_discussion` to allow discussions on the Merge Request.
- `approve` to approve the Merge Request.
- `unapprove` to approve the Merge Request.
- `comment` to add a comment to the Merge Request

      *Additional fields:*

      - (required) `message` The message that will be commented on the Merge Request.

      ```{.yaml title="'comment' example"}
      - action: comment
        message: |
          Hello world
      ```

- `add_label` to add *an existing* label to the Merge Request

      *Additional fields:*

      - (required) `label` The label name to add.

      ```{.yaml title="add_label example"}
      - action: add_label
        label: example
      ```

- `remove_label` to remove a label from the Merge Request

      *Additional fields:*

      - (required) `label` The label name to add.

      ```{.yaml title="remove_label example"}
      - action: remove_label
        label: example
      ```

## `label[]` {#label data-toc-label="label"}

The `label` key is a list of the labels you want to manage.

These keys are shared between the `conditional` and `generate` label strategy. (more above these below!)

### `label[].name` {#label.name data-toc-label="name"}

- When using `label.strategy: conditional`

    **REQUIRED** The `name` of the label to create.

- When using `label.strategy: generate`

    **OMITTED** The `name` field must not be set when using the `generate` strategy.

### `label[].script` {#label.script data-toc-label="script"}

!!! tip

    See the [SCM engine expr-lang documentation](#expr-lang-information) for more information about [functions](#functions) and [attributes](#attributes) available.

The `script` field is an [expr-lang](https://expr-lang.org/) expression, a safe, fast, and intuitive expression evaluator.

Depending on the `label.strategy` used, the behavior of the script changes, read more about this below.

### `label[].strategy` {#label.strategy data-toc-label="strategy"}

SCM Engine supports two strategies for managing labels, each changes the behavior of the `script`.

- `conditional` (default, if `type` key is omitted), where you provide the `name` of the label, and a `script` that returns a boolean for wether the label should be added to the Merge Request.

    The `script` must return a `boolean` value, where `true` mean `add the label` and `false` mean `remove the label`.

- `generate`, where your `script` generates the list of labels that should be added to the Merge Request.

    The `script` must return a `list of strings`, where each label returned will be added to the Merge Request.

#### `label[].strategy = conditional` {#label.strategy-conditional data-toc-label="conditional"}

Use the `conditional` strategy when you want to add/remove a label on a Merge Request depending on *something*. It's the default strategy, and the most simple one to use.

!!! note

    The `script` field is a [expr-lang](https://expr-lang.org/) expression, a safe, fast, and intuitive expression evaluator.

```yaml
label:
    # Add a "lang/go" label if any "*.go" files was changed
  - name: lang/go
    color: "$indigo"
    script: merge_request.modified_files("*.go")

    # Add a "lang/markdown" label if any "*.md" files was changed
  - name: lang/markdown
    color: "$indigo"
    script: merge_request.modified_files("*.md")

    # Add a "type/documentation" label if any files was changed within the "docs/" folder
  - name: type/documentation
    color: "$green"
    script: merge_request.modified_files("docs/")

    # Add a "go::tests" scoped & prioritized label with value "missing" if no "*_test.go" files was changed
  - name: go::tests::missing
    color: "$red"
    priority: 999
    script: not merge_request.modified_files("*_test.go")

    # Add a "go::tests" scoped & prioritized label with value "OK" if any "*_test.go" files was changed
  - name: go::tests::ok
    color: "$green"
    priority: 999
    script: merge_request.modified_files("*_test.go")
```

#### `label[].strategy = generate` {#label.strategy-generate data-toc-label="generate"}

Use the `generate` strategy if you want to manage dynamic labels, for example, depending on the file structure within your project.

> The `script` field is a [expr-lang](https://expr-lang.org/) expression, a safe, fast, and intuitive expression evaluator.

Thanks to the dynamic nature of the `generate` strategy, it has fantastic flexibility, at the cost of greater flexibility.

```yaml
label:
    # Generate list of labels via script.
    #
    # Image you have a project where you have multiple "service" directories
    #
    # * pkg/service/example/file.go
    # * pkg/service/scm/gitlab/file.go
    # * pkg/service/scm/github/file.go
    #
    # and you want to generate a labels like this
    #
    # * service/example
    # * service/scm/gitlab
    # * service/scm/github
    #
    # depending on what directories are having files changed in a Merge Request.
  - strategy: generate
    description: "Modified this service directory"
    color: "$pink"
    script: >
      // Generate a list of all file paths that was changed in the Merge Request inside pkg/service/
      merge_request.modified_files_list("pkg/service/")

      // Remove the filename from the path "pkg/service/example/file.go" => "pkg/service/example"
      | map({ filepath_dir(#) })

      // Remove the prefix "pkg/" from the path "pkg/service/example" => "service/example"
      | map({ trimPrefix(#, "pkg/") })

      // Remove duplicate values from the output
      | uniq()
```

### `label[].color` {#label.color data-toc-label="color"}

!!! note

    When used on `strategy: generate` labels, all generated labels will have the same color.

`color` is a mandatory field, controlling the background color of the label when viewed in the User Interface.

You can either provide your own `#hex` value or use the [Twitter Bootstrap color variables](https://getbootstrap.com/docs/5.3/customize/color/#all-colors), for example `$blue-500` and `$teal`.

### `label[].description` {#label.description data-toc-label="description"}

!!! note

    When used on `strategy: generate` labels, all generated labels will have the same description.

An optional key that control the `description` field for the label within GitLab.

Descriptions are shown in the User Interface when you hover any label.

### `label[].priority` {#label.priority data-toc-label="priority"}

!!! note

    When used on `strategy: generate` labels, all generated labels will have the same priority.

An optional key that controls the [label `priority`](https://docs.gitlab.com/ee/user/project/labels.html#set-label-priority).

### `label[].skip_if` {#label.skip_if data-toc-label="skip_if"}

An optional key controlling if the label should be skipped (meaning no removal or adding of labels).

The `skip_if` field must be a valid [Expr-lang](https://expr-lang.org/) expression returning a boolean, where `true` means `skip` and `false` means `process`.
