# scm-engine

> SCM Engine allow for easy Merge Request automation within your GitLab projects.
>
> Automatically add / remove labels depending on files changes, age of the Merge Request, who contributes,
> and pretty much anything else you could want, thanks to the usage of [expr-lang](https://expr-lang.org/).
>
> SCM engine can be run either as a regular CI job in your pipeline, or be triggered through the Webhook system, allowing for versatile and flexible deployments.

- [Installation](#installation)
  - [Docker](#docker)
  - [homebrew tap](#homebrew-tap)
  - [apt](#apt)
  - [yum](#yum)
  - [snapcraft](#snapcraft)
  - [scoop](#scoop)
  - [aur](#aur)
  - [deb, rpm and apk packages](#deb-rpm-and-apk-packages)
  - [go install](#go-install)
- [Usage](#usage)
  - [GitLab-CI pipeline](#gitlab-ci-pipeline)
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
  - [Functions](#functions)

## Installation

### Docker

```shell
docker run --rm ghcr.io/jippi/scm-engine
```

### homebrew tap

```shell
brew install jippi/tap/scm-engine
```

### apt

```shell
echo 'deb [trusted=yes] https://pkg.jippi.dev/apt/ * *' | sudo tee /etc/apt/sources.list.d/scm-engine.list
sudo apt update
sudo apt install scm-engine
```

### yum

```shell
echo '[scm-engine]
name=scm-engine
baseurl=https://pkg.jippi.dev/yum/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/scm-engine.repo
sudo yum install scm-engine
```

### snapcraft

```shell
sudo snap install scm-engine
```

### scoop

```shell
scoop bucket add scm-engine https://github.com/jippi/scoop-bucket.git
scoop install scm-engine
```

### aur

```shell
yay -S scm-engine-bin
```

### deb, rpm and apk packages

Download the `.deb`, `.rpm` or `.apk` packages from the [releases page](https://github.com/jippi/scm-engine/releases) and install them with the appropriate tools.

### go install

```shell
go install github.com/jippi/scm-engine/cmd@latest
```

## Usage

### GitLab-CI pipeline

Using scm-engine within a GitLab CI pipeline is straight forward.

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
    scm-engine:
      image: ghcr.io/jippi/scm-engine:latest
      rules:
        - if: $CI_PIPELINE_SOURCE == 'merge_request_event'
      script:
        - scm-engine evaluate
    ```

1. Done! Every Merge Request change should now re-run scm-engine and apply your label rules

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
   --api-token value  GitHub/GitLab API token [$SCM_ENGINE_TOKEN]
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
    description: "Modified this service directory"
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
    description: "Modified this service directory"
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
> Missing an attribute? The `schema/gitlab.schema.graphqls` file are what is used to query GitLab, adding the missing `field` to the right `type` should make it accessible.
> Please open an issue or Pull Request if something is missing.

> [!IMPORTANT]
> _SCM Engine uses [`snake_case`](https://en.wikipedia.org/wiki/Snake_case) for fields instead of [`camelCase`](https://en.wikipedia.org/wiki/Camel_case)_

The following attributes are available in `script` fields.

They can be accessed exactly as shown in this list.

- `group.description` (string) Description of the namespace
- `group.emails_disabled` (optional bool) Indicates if a group has email notifications disabled
- `group.full_name` (string) Full name of the namespace
- `group.full_path` (string) Full path of the namespace
- `group.id` (string) ID of the namespace
- `group.mentions_disabled` (optional bool) Indicates if a group is disabled from getting mentioned
- `group.name` (string) Name of the namespace
- `group.path` (string) Path of the namespace
- `group.visibility` (optional string) Visibility of the namespace
- `group.web_url` (string) Web URL of the group
- `merge_request.approvals_left` (optional int) Number of approvals left
- `merge_request.approvals_required` (optional int) Number of approvals required
- `merge_request.approved` (bool) Indicates if the merge request has all the required approvals
- `merge_request.auto_merge_enabled` (bool) Indicates if auto merge is enabled for the merge request
- `merge_request.auto_merge_strategy` (optional string) Selected auto merge strategy
- `merge_request.commit_count` (optional int) Number of commits in the merge request
- `merge_request.conflicts` (bool) Indicates if the merge request has conflicts
- `merge_request.created_at` (time) Timestamp of when the merge request was created
- `merge_request.description` (optional string) Description of the merge request (Markdown rendered as HTML for caching)
- `merge_request.diff_stats[].additions` (int) Number of lines added to this file
- `merge_request.diff_stats[].deletions` (int) Number of lines deleted from this file
- `merge_request.diff_stats[].path` (string) File path, relative to repository root
- `merge_request.discussion_locked` (bool) Indicates if comments on the merge request are locked to members only
- `merge_request.diverged_from_target_branch` (bool) Indicates if the source branch is behind the target branch
- `merge_request.downvotes` (int) Number of downvotes for the merge request
- `merge_request.draft` (bool) Indicates if the merge request is a draft
- `merge_request.first_commit.author_email` (optional string) Commit author’s email
- `merge_request.first_commit.author_name` (optional string) Commit authors name
- `merge_request.first_commit.authored_date` (optional time) Timestamp of when the commit was authored
- `merge_request.first_commit.committed_date` (optional time) Timestamp of when the commit was committed
- `merge_request.first_commit.committer_email` (optional string) Email of the committer
- `merge_request.first_commit.committer_name` (optional string) Name of the committer
- `merge_request.first_commit.description` (optional string) Description of the commit message
- `merge_request.first_commit.full_title` (optional string) Full title of the commit message
- `merge_request.first_commit.id` (optional string) ID (global ID) of the commit
- `merge_request.first_commit.message` (optional string) Raw commit message
- `merge_request.first_commit.sha` (string) SHA1 ID of the commit
- `merge_request.first_commit.short_id` (string) Short SHA1 ID of the commit
- `merge_request.first_commit.title` (optional string) Title of the commit message
- `merge_request.first_commit.web_url` (string) Web URL of the commit
- `merge_request.force_remove_source_branch` (optional bool) Indicates if the project settings will lead to source branch deletion after merge
- `merge_request.id` (string) ID of the merge request
- `merge_request.iid` (string) Internal ID of the merge request
- `merge_request.labels[].color` (string) Background color of the label
- `merge_request.labels[].description` (string) Description of the label (Markdown rendered as HTML for caching)
- `merge_request.labels[].id` (string) Label ID
- `merge_request.labels[].title` (string) Content of the label
- `merge_request.last_commit.author_email` (optional string) Commit author’s email
- `merge_request.last_commit.author_name` (optional string) Commit authors name
- `merge_request.last_commit.authored_date` (optional time) Timestamp of when the commit was authored
- `merge_request.last_commit.committed_date` (optional time) Timestamp of when the commit was committed
- `merge_request.last_commit.committer_email` (optional string) Email of the committer
- `merge_request.last_commit.committer_name` (optional string) Name of the committer
- `merge_request.last_commit.description` (optional string) Description of the commit message
- `merge_request.last_commit.full_title` (optional string) Full title of the commit message
- `merge_request.last_commit.id` (optional string) ID (global ID) of the commit
- `merge_request.last_commit.message` (optional string) Raw commit message
- `merge_request.last_commit.sha` (string) SHA1 ID of the commit
- `merge_request.last_commit.short_id` (string) Short SHA1 ID of the commit
- `merge_request.last_commit.title` (optional string) Title of the commit message
- `merge_request.last_commit.web_url` (string) Web URL of the commit
- `merge_request.merge_status_enum` (string) Merge status of the merge request
- `merge_request.merge_when_pipeline_succeeds` (optional bool) Indicates if the merge has been set to auto-merge
- `merge_request.mergeable` (bool) Indicates if the merge request is mergeable
- `merge_request.mergeable_discussions_state` (optional bool) Indicates if all discussions in the merge request have been resolved, allowing the merge request to be merged
- `merge_request.merged_at` (optional time) Timestamp of when the merge request was merged, null if not merged
- `merge_request.prepared_at` (optional time) Timestamp of when the merge request was prepared
- `merge_request.should_be_rebased` (bool) Indicates if the merge request will be rebased
- `merge_request.should_remove_source_branch` (optional bool) Indicates if the source branch of the merge request will be deleted after merge
- `merge_request.source_branch` (string) Source branch of the merge request
- `merge_request.source_branch_exists` (bool) Indicates if the source branch of the merge request exists
- `merge_request.source_branch_protected` (bool) Indicates if the source branch is protected
- `merge_request.squash` (bool) Indicates if the merge request is set to be squashed when merged. Project settings may override this value. Use squash_on_merge instead to take project squash options into account
- `merge_request.squash_on_merge` (bool) Indicates if the merge request will be squashed when merged
- `merge_request.state` (string) State of the merge request
- `merge_request.target_branch` (string) Target branch of the merge request
- `merge_request.target_branch_exists` (bool) Indicates if the target branch of the merge request exists
- `merge_request.time_between_first_and_last_commit` (optional duration)
- `merge_request.time_since_first_commit` (optional duration)
- `merge_request.time_since_last_commit` (optional duration)
- `merge_request.title` (string) Title of the merge request
- `merge_request.updated_at` (time) Timestamp of when the merge request was last updated
- `merge_request.upvotes` (int) Number of upvotes for the merge request.
- `merge_request.user_discussions_count` (optional int) Number of user discussions in the merge request
- `merge_request.user_notes_count` (optional int) User notes count of the merge request
- `project.archived` (bool) Indicates the archived status of the project
- `project.created_at` (time) Timestamp of the project creation
- `project.description` (string) Short description of the project
- `project.full_path` (string) Full path of the project
- `project.id` (string) ID of the project
- `project.issues_enabled` (bool) Indicates if Issues are enabled for the current user
- `project.labels[].color` (string) Background color of the label
- `project.labels[].description` (string) Description of the label (Markdown rendered as HTML for caching)
- `project.labels[].id` (string) Label ID
- `project.labels[].title` (string) Content of the label
- `project.last_activity_at` (time) Timestamp of the project last activity
- `project.name` (string) Name of the project (without namespace)
- `project.name_with_namespace` (string) Full name of the project with its namespace
- `project.path` (string) Path of the project
- `project.topics` ([]string) List of project topics
- `project.visibility` (string) Visibility of the project

### Functions

#### `merge_request.modified_files`

Returns wether any of the provided files patterns have been modified in the Merge Request.

The file patterns use the [`.gitignore` format](https://git-scm.com/docs/gitignore#_pattern_format).

```expr
merge_request.modified_files("*.go", "docs/")
```

#### `merge_request.has_label`

Returns wether any of the provided label exist on the Merge Request.

```expr
merge_request.has_label("my-label-name")
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
filepath_dir("example/directory/file.go") == "example/directory"
```

#### `limit_path_depth_to`

`limit_path_depth_to` takes a path structure, and limits it to the configured maximum depth. Particularly useful when using `generated` labels from a directory structure, and want to to have a label naming scheme that only uses path of the path.

```expr
limit_path_depth_to("path1/path2/path3/path4", 2), == "path1/path2"
limit_path_depth_to("path1/path2", 3), == "path1/path2"
```
