# About

> [!NOTE]
> SCM Engine allow for easy Merge Request automation within your GitLab projects.
>
> Automatically add / remove labels depending on files changes, age of the Merge Request, who contributes,
> and pretty much anything else you could want.
>
> SCM engine can be run either as a regular CI job in your pipeline, or be triggered through the Webhook system, allowing for versatile and flexible deployments.

## Documentation

> [!TIP]
> Please see [the documentation site](https://jippi.github.io/scm-engine/) for in-depth information
>
> * [Installation](https://jippi.github.io/scm-engine/install/)
> * [Configuration](https://jippi.github.io/scm-engine/configuration/)
> * [Configuration Examples](https://jippi.github.io/scm-engine/configuration/examples/)
>
> **Commands:**
>
> * [evaluate](https://jippi.github.io/scm-engine/commands/evaluate/)
> * [server](https://jippi.github.io/scm-engine/commands/server/)
>
> **GitLab:**
>
> * [Getting started](https://jippi.github.io/scm-engine/gitlab/setup/)
> * [Script attributes](https://jippi.github.io/scm-engine/gitlab/script-attributes/)
> * [Script functions](https://jippi.github.io/scm-engine/gitlab/script-functions/)

## Example

```yaml
# See: https://getbootstrap.com/docs/5.3/customize/color/#all-colors

actions:
  - name: Warn if the Merge Request haven't had commit activity for 21 days and will be closed
    if: |1
         merge_request.state != "closed"
      && merge_request.time_since_last_commit > duration("21d")
      && merge_request.time_since_last_commit < duration("28d")
      && not merge_request.has_label("do-not-close")
    then:
      - action: comment
        message: |
          :wave: Hello!

          This Merge Request has not seen any commit activity for 21 days.
          We will automatically close the Merge request after 28 days to keep our project clean.

          To disable this behavior, add the `do-not-close` label to the Merge Request in the right menu or add a comment with `/label ~"do-not-close"`.

  - name: Close the Merge Request if it haven't had commit activity for 28 days
    if: |1
         merge_request.state != "closed"
      && merge_request.time_since_last_commit > duration("28d")
      && not merge_request.has_label("do-not-close")
    then:
      - action: close
      - action: comment
        message: |
          :wave: Hello!

          This Merge Request has not seen any commit activity for 28 days.
          To keep our project clean, we will close the Merge request now.

          To disable this behavior, add the `do-not-close` label to the Merge Request in the right menu or add a comment with `/label ~"do-not-close"`.

  - name: Approve MR if the 'break-glass-approve' label is configured
    if: |1
         merge_request.state != "closed"
      && not merge_request.approved
      && merge_request.has_label("break-glass-approve")
    then:
      - action: approve
      - action: comment
        message: "Approving the MR since it has the 'break-glass-approve' label. Talk to ITGC about this!"

label:
  - name: lang/go
    color: "$indigo"
    script: merge_request.modified_files("*.go")

  - name: lang/markdown
    color: "$indigo"
    description: "Modified MarkDown files"
    script: merge_request.modified_files("*.md")

  - name: dependencies/go
    color: "$orange"
    description: "Updated Go dependency files like go.mod and go.sum"
    script: merge_request.modified_files("go.mod", "go.sum")

  - name: type/ci
    color: "$green"
    description: "Modified CI files"
    script: merge_request.modified_files(".gitlab-ci.yml") || merge_request.modified_files("build/")

  - name: type/deployment
    color: "$green"
    description: "Modified Deployment files"
    script: merge_request.modified_files("_infrastructure/", "scripts/", "configs/")

  - name: type/documentation
    color: "$green"
    description: "Modified Documentation files"
    script: merge_request.modified_files("docs/")

  - name: type/services
    color: "$green"
    description: "Modified pkg/services files"
    script: merge_request.modified_files("internal/pkg/services")

  - name: go::tests::missing
    color: "$red"
    description: "The Merge Request did NOT modify Go test files"
    priority: 999
    script: not merge_request.modified_files("*_test.go") && merge_request.modified_files("*.go")

  - name: go::tests::OK
    color: "$green"
    description: "The Merge Request modified Go test files"
    priority: 999
    script: merge_request.modified_files("*_test.go") && merge_request.modified_files("*.go")

  - name: status::age::abandoned
    color: "$red"
    description: "The most recent commit is older than 45 days"
    priority: 999
    script: merge_request.time_since_last_commit > duration("45d")
    skip_if: merge_request.state in ["merged", "closed", "locked"]

  - name: status::age::stale
    color: "$red"
    description: "The most recent commit is older than 30 days"
    priority: 999
    script: duration("30d") < merge_request.time_since_last_commit < duration("45d")
    skip_if: merge_request.state in ["merged", "closed", "locked"]

  - name: status::age::old
    color: "$red"
    description: "The most recent commit is older than 14 days"
    priority: 999
    script: duration("14d") < merge_request.time_since_last_commit < duration("30d")
    skip_if: merge_request.state in ["merged", "closed", "locked"]

  # generate labels for services
  #
  # internal/service/vault/client.go
  # =>
  # service/vault
  - strategy: generate
    description: "Modified this a service directory"
    color: "$pink"
    script: >
      merge_request.modified_files_list("internal/service/")
      | map({ filepath_dir(#) })
      | map({ trimPrefix(#, "internal/") })
      | uniq()

  # generate labels for commands
  #
  # internal/app/my-command/subcommands/aws/login/login.go
  # =>
  # command/aws/login
  - strategy: generate
    description: "Modified this my-command command"
    color: "$purple"
    script: >
      merge_request.modified_files_list("internal/app/my-command/subcommands/")
      | map({ filepath_dir(#) })
      | map({ trimPrefix(#, "internal/app/my-command/subcommands/") })
      | map({ string("command/" + #) })
      | uniq()
```
