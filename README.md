# scm-engine

> SCM Engine allow for easy Merge Request automation within your GitLab projects.
>
> Automatically add / remove labels depending on files changes, age of the Merge Request, who contributes,
> and pretty much anything else you could want, thanks to the usage of [expr-lang](https://expr-lang.org/).
>
> SCM engine can be run either as a regular CI job in your pipeline, or be triggered through the Webhook system, allowing for versatile and flexible deployments.

- [Installation](#installation)
- [Commands](#commands)
  - [`evaluate`](#evaluate)
- [Configuration file](#configuration-file)
  - [Examples](#examples)
  - [`label` (list)](#label-list)
    - [`label.name`](#labelname)
    - [`label.script` (required)](#labelscript-required)
    - [`label.strategy` (optional)](#labelstrategy-optional)
      - [`label.strategy: conditional` use-cases](#labelstrategy-conditional-use-cases)
      - [`label.strategy: conditional` examples](#labelstrategy-conditional-examples)
      - [`label.strategy: generate` use-cases](#labelstrategy-generate-use-cases)
      - [`label.strategy: generate` examples](#labelstrategy-generate-examples)
    - [`label.color` (required)](#labelcolor-required)
    - [`label.description` (optional)](#labeldescription-optional)
    - [`label.priority` (optional)](#labelpriority-optional)
    - [`label.skip_if` (optional)](#labelskip_if-optional)
- [Expr-lang information](#expr-lang-information)
  - [Attributes](#attributes)
    - [project](#project)
    - [project.labels](#projectlabels)
    - [group](#group)
    - [merge_request](#merge_request)
    - [merge_request.diff_stats](#merge_requestdiff_stats)
    - [merge_request.first_commit](#merge_requestfirst_commit)
    - [merge_request.last_commit](#merge_requestlast_commit)
    - [merge_request.labels](#merge_requestlabels)
  - [Functions](#functions)
    - [`merge_request.modified_files`](#merge_requestmodified_files)
    - [`duration`](#duration)
    - [`uniq`](#uniq)
    - [`filepath_dir`](#filepath_dir)

## Installation

TODO

## Commands

### `evaluate`

Evaluate the SCM engine rules against a specific Merge Request.

```plain
NAME:
   scm-engine evaluate - Evaluate a Merge Request

USAGE:
   scm-engine evaluate [command options]

OPTIONS:
   --id value, --merge-request-id value, --pull-request-id value  The pull/merge to process, if not provided as a CLI flag [$CI_MERGE_REQUEST_IID]

GLOBAL OPTIONS:
   --config value     Path to the scm-engine config file (default: ".scm-engine.yml") [$SCM_ENGINE_CONFIG_FILE]
   --api-token value  GitHub/GitLab API token [$GITLAB_TOKEN, $SCM_ENGINE_TOKEN]
   --project value    GitLab project (example: 'gitlab-org/gitlab') [$GITLAB_PROJECT, $CI_PROJECT_PATH]
   --base-url value   Base URL for the SCM instance (default: "https://gitlab.com/") [$GITLAB_BASEURL, $CI_SERVER_URL]
   --help, -h         show help
```

## Configuration file

The default configuration filename is `.scm-engine.yml`, either in current working directory, or if you are in a Git repository, the root of the project.

The file path can be changed via `--config` CLI flag and `$SCM_ENGINE_CONFIG_FILE` environment variable.

### Examples

> [!NOTE]
> A quick demo of what SCM Engine can do. More details documentation further down the document.
>
> The `script` field is a [expr-lang](https://expr-lang.org/) expression, a safe, fast, and intuitive expression evaluator.

```yaml
label:
    # Add a label named "lang/go"
  - name: lang/go
    # using the "conditional" strategy
    strategy: conditional
    # and a description (optional)
    description: "Modified Go files"
    # and the color $indigo
    color: "$indigo"
    # if files matching "*.go" was modified
    script: merge_request.modified_files("*.go")

    # Generate list of labels via script
  - strategy: generate
    # With a description (optional)
    description: "Modified this letsgo service directory"
    # With the color $pink
    color: "$pink"
    # From this script, returning a list of labels
    script: >
      map(merge_request.diff_stats, { .path })   // Generate a list of all file paths that was changed in the Merge Request
      | filter({ hasPrefix(#, "pkg/service/") }) // Remove all paths that doesn't start with "pkg/service/"
      | map({ filepath_dir(#) })                 // Remove the filename from the path "pkg/service/example/file.go" => "pkg/service/example"
      | map({ trimPrefix(#, "pkg/") })           // Remove the prefix "pkg/" from the path "pkg/service/example" => "service/example"
      | uniq()                                   // Remove duplicate values from the output
```

### `label` (list)

The `label` key is a list of the labels you want to manage.

These keys are shared between the `conditional` and `generate` label strategy. (more above these below!)

#### `label.name`

- When using `label.strategy: conditional`

    **REQUIRED** The `name` of the label to create.

- When using `label.strategy: generate`

    **OMITTED** The `name` field must not be set when using the `generate` strategy.

#### `label.script` (required)

> [!TIP]
> See the [SCM engine expr-lang documentation](#expr-lang-information) for more information about [functions](#functions) and [attributes](#attributes) available.

The `script` field is an [expr-lang](https://expr-lang.org/) expression, a safe, fast, and intuitive expression evaluator.

Depending on the `label.strategy` used, the behavior of the script changes, read more about this below.

#### `label.strategy` (optional)

SCM Engine supports two strategies for managing labels, each changes the behavior of the `script`.

- `conditional` (default, if `type` key is omitted), where you provide the `name` of the label, and a `script` that returns a boolean for wether the label should be added to the Merge Request.

    The `script` must return a `boolean` value, where `true` mean `add the label` and `false` mean `remove the label`.

- `generate`, where your `script` generates the list of labels that should be added to the Merge Request.

    The `script` must return a `list of strings`, where each label returned will be added to the Merge Request.

##### `label.strategy: conditional` use-cases

Use the `conditional` strategy when you want to add/remove a label on a Merge Request depending on _something_. It's the default strategy, and the most simple one to use.

##### `label.strategy: conditional` examples

> [!NOTE]
> The `script` field is a [expr-lang](https://expr-lang.org/) expression, a safe, fast, and intuitive expression evaluator.

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

##### `label.strategy: generate` use-cases

Use the `generate` strategy if you want to manage dynamic labels, for example, depending on the file structure within your project.

##### `label.strategy: generate` examples

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
    description: "Modified this letsgo service directory"
    color: "$pink"
    script: >
      map(merge_request.diff_stats, { .path })   // Generate a list of all file paths that was changed in the Merge Request
      | filter({ hasPrefix(#, "pkg/service/") }) // Remove all paths that doesn't start with "pkg/service/"
      | map({ filepath_dir(#) })                 // Remove the filename from the path "pkg/service/example/file.go" => "pkg/service/example"
      | map({ trimPrefix(#, "pkg/") })           // Remove the prefix "pkg/" from the path "pkg/service/example" => "service/example"
      | uniq()                                   // Remove duplicate values from the output
```

#### `label.color` (required)

> [!NOTE]
> When used on `strategy: generate` labels, all generated labels will have the same color.

`color` is a mandatory field, controlling the background color of the label when viewed in the User Interface.

You can either provide your own `#hex` value or use the [Twitter Bootstrap color variables](https://getbootstrap.com/docs/5.3/customize/color/#all-colors), for example `$blue-500` and `$teal`.

#### `label.description` (optional)

> [!NOTE]
> When used on `strategy: generate` labels, all generated labels will have the same description.

An optional key that control the `description` field for the label within GitLab.

Descriptions are shown in the User Interface when you hover any label.

#### `label.priority` (optional)

> [!NOTE]
> When used on `strategy: generate` labels, all generated labels will have the same priority.

An optional key that controls the [label `priority`](https://docs.gitlab.com/ee/user/project/labels.html#set-label-priority).

#### `label.skip_if` (optional)

An optional key controlling if the label should be skipped (meaning no removal or adding of labels).

The `skip_if` field must be a valid [Expr-lang](https://expr-lang.org/) expression returning a boolean, where `true` means `skip` and `false` means `process`.

## Expr-lang information

> [!TIP]
> The [Expr Language Definition](https://expr-lang.org/docs/language-definition) is a great resource to learn more about the language. This guide will only cover SCM Engine specific extensions and information.

### Attributes

> [!NOTE]
> Missing an attribute? The `pkg/scm/gitlab/context_*` files are what is used to query GitLab, adding the missing `field` to the right `struct` should make it accessible.
> Please open an issue or Pull Request if something is missing.

> [!IMPORTANT]
> _SCM Engine uses [`snake_case`](https://en.wikipedia.org/wiki/Snake_case) for fields instead of [`camelCase`](https://en.wikipedia.org/wiki/Camel_case)_

The following attributes are available in `script` fields.

They can be accessed exactly as shown in this list.

#### project

> [!NOTE]
> See the [GitLab GraphQL `Project` GraphQL resource](https://docs.gitlab.com/ee/api/graphql/reference/#project) for more details about the fields.

- `project.archived` (boolean)
- `project.created_at` (time)
- `project.description` (string)
- `project.full_path` (string)
- `project.id` (string)
- `project.last_activity_at` (time)
- `project.name_with_namespace` (string)
- `project.name` (string)
- `project.path` (string)
- `project.topics[]` (array of string)
- `project.visibility` (string)

#### project.labels

> [!NOTE]
> See the [GitLab GraphQL `Label` GraphQL resource](https://docs.gitlab.com/ee/api/graphql/reference/#label) for more details about the fields.

- `project.labels[].color` (string)
- `project.labels[].description` (string)
- `project.labels[].id` (string)
- `project.labels[].title` (string)

#### group

> See the [GitLab GraphQL `Group` GraphQL resource](https://docs.gitlab.com/ee/api/graphql/reference/#group) for more details about the fields.

- `group.description` (string)
- `group.id` (string)
- `group.name` (string)

#### merge_request

> See the [GitLab GraphQL `MergeRequest` GraphQL resource](https://docs.gitlab.com/ee/api/graphql/reference/#mergerequest) for more details about the fields.

- `merge_request.approvals_left` (int)
- `merge_request.approvals_required` (int)
- `merge_request.approved` (boolean)
- `merge_request.auto_merge_enabled` (int)
- `merge_request.auto_merge_strategy` (string)
- `merge_request.conflicts` (bool)
- `merge_request.created_at` (time)
- `merge_request.description` (string)
- `merge_request.diverged_from_target_branch` (bool)
- `merge_request.draft` (boolean)
- `merge_request.id` (string)
- `merge_request.iid` (string)
- `merge_request.merge_status_enum` (string)
- `merge_request.mergeable` (boolean)
- `merge_request.merged_at` (optional, time)
- `merge_request.source_branch_exists` (boolean)
- `merge_request.source_branch_protected` (boolean)
- `merge_request.source_branch` (string)
- `merge_request.squash_on_merge` (boolean)
- `merge_request.squash` (boolean)
- `merge_request.state` (string)
- `merge_request.target_branch_exists` (string)
- `merge_request.target_branch` (string)
- `merge_request.time_between_first_and_last_commit` (duration) - SCM Engine - The `duration()` between the first and last commit in the Merge Request.
- `merge_request.time_since_first_commit` (duration) - SCM Engine - The `duration()` between `now()` and the first commit in the Merge Request.
- `merge_request.time_since_last_commit` (duration) - SCM Engine - The `duration()` between `now()` and the last commit in the Merge Request.
- `merge_request.title` (string)
- `merge_request.updated_at` (time)

#### merge_request.diff_stats

> See the [GitLab GraphQL `DiffStats` GraphQL resource](https://docs.gitlab.com/ee/api/graphql/reference/#diffstats) for more details about the fields.

- `merge_request.diff_stats[].additions` (int)
- `merge_request.diff_stats[].deletions` (int)
- `merge_request.diff_stats[].path` (string)

#### merge_request.first_commit

> See the [GitLab GraphQL `Commit` GraphQL resource](https://docs.gitlab.com/ee/api/graphql/reference/#commit) for more details about the fields.

- `merge_request.first_commit.author_email` (string)
- `merge_request.first_commit.committed_date` (string)

#### merge_request.last_commit

> See the [GitLab GraphQL `Commit` GraphQL resource](https://docs.gitlab.com/ee/api/graphql/reference/#commit) for more details about the fields.

- `merge_request.last_commit.author_email` (string)
- `merge_request.last_commit.committed_date` (string)

#### merge_request.labels

> See the [GitLab GraphQL `Label` GraphQL resource](https://docs.gitlab.com/ee/api/graphql/reference/#label) for more details about the fields.

- `merge_request.labels[].color` (string)
- `merge_request.labels[].description` (string)
- `merge_request.labels[].id` (string)
- `merge_request.labels[].title` (string)

### Functions

#### `merge_request.modified_files`

Returns wether any of the provided files patterns have been modified in the Merge Request.

The file patterns use the [`.gitignore` format](https://git-scm.com/docs/gitignore#_pattern_format).

```expr
merge_request.modified_files("*.go", "docs/")
```

#### `duration`

Returns the [`time.Duration`](https://pkg.go.dev/time#Duration) value of the given string str.

Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d" and "w".

```expr
duration("1h").Seconds() == 3600
```

#### `uniq`

Returns a new array where all duplicate values has been removed.

```expr
(["hello", "world", "world"] | uniq) == ["hello", "world"]
```

#### `filepath_dir`

`filepath_dir` returns all but the last element of path, typically the path's directory. After dropping the final element,

Dir calls [Clean](https://pkg.go.dev/path/filepath#Clean) on the path and trailing slashes are removed.

If the path is empty, `filepath_dir` returns ".". If the path consists entirely of separators, `filepath_dir` returns a single separator.

The returned path does not end in a separator unless it is the root directory.

```expr
filepath_dir("/example/directory/file.go") == "/example/directory"
```
