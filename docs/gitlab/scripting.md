# Scripting information (Expr)

!!! TIP

    The [Expr Language Definition](https://expr-lang.org/docs/language-definition) is a great resource to learn more about the language. This guide will only cover SCM Engine specific extensions and information.

## Attributes

!!! note

    Missing an attribute? The `schema/gitlab.schema.graphqls` file are what is used to query GitLab, adding the missing `field` to the right `type` should make it accessible. Please open an issue or Pull Request if something is missing.

!!! important

    _SCM Engine uses [`snake_case`](https://en.wikipedia.org/wiki/Snake_case) for fields instead of [`camelCase`](https://en.wikipedia.org/wiki/Camel_case)_

!!! tip

    You have access to the raw webhook event payload via `webhook_event.*` attributes (not listed below) in Expr script fields when using [`server`](#server) mode. See the [GitLab Webhook Events documentation](https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html) for available fields. The attributes are named _exactly_ as documented in the GitLab documentation.

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

## Functions

### `merge_request.modified_files`

Returns wether any of the provided files patterns have been modified in the Merge Request.

The file patterns use the [`.gitignore` format](https://git-scm.com/docs/gitignore#_pattern_format).

```expr
merge_request.modified_files("*.go", "docs/") == true
```

### `merge_request.modified_files_list`

Returns an array of files matching the provided (optional) pattern thas has been modified in the Merge Request.

The file patterns use the [`.gitignore` format](https://git-scm.com/docs/gitignore#_pattern_format).

```expr
merge_request.modified_files_list("*.go", "docs/") == ["example/file.go", "docs/index.md"]
```

### `merge_request.has_label`

Returns wether any of the provided label exist on the Merge Request.

```expr
merge_request.has_label("my-label-name")
```

### `duration`

Returns the [`time.Duration`](https://pkg.go.dev/time#Duration) value of the given string str.

Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d" and "w".

```expr
duration("1h").Seconds() == 3600
```

### `uniq`

Returns a new array where all duplicate values has been removed.

```expr
(["hello", "world", "world"] | uniq) == ["hello", "world"]
```

### `filepath_dir`

`filepath_dir` returns all but the last element of path, typically the path's directory. After dropping the final element,

Dir calls [Clean](https://pkg.go.dev/path/filepath#Clean) on the path and trailing slashes are removed.

If the path is empty, `filepath_dir` returns ".". If the path consists entirely of separators, `filepath_dir` returns a single separator.

The returned path does not end in a separator unless it is the root directory.

```expr
filepath_dir("example/directory/file.go") == "example/directory"
```

### `limit_path_depth_to`

`limit_path_depth_to` takes a path structure, and limits it to the configured maximum depth. Particularly useful when using `generated` labels from a directory structure, and want to to have a label naming scheme that only uses path of the path.

```expr
limit_path_depth_to("path1/path2/path3/path4", 2), == "path1/path2"
limit_path_depth_to("path1/path2", 3), == "path1/path2"
```
