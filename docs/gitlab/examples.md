# Examples

!!! note

    A quick demo of what SCM Engine can do.

    The `script` field is a [expr-lang](https://expr-lang.org/) expression, a safe, fast, and intuitive expression evaluator.

## Close Merge Requests without recent activity

This example will close a Merge Request if no activity has happened for 28 days.

The script will warn at 21 days mark that the Merge Request will be closed, with instructions on how to prevent it.

```{.yaml linenums=1}
# yaml-language-server: $schema=https://jippi.github.io/scm-engine/scm-engine.schema.json

label:
  - name: "stale" # (1)!
    color: $red # (11)!
    script: |1 # (2)!
        --8<-- "docs/gitlab/snippets/close-merge-request/label-script.expr"

actions:
  - name: "warn" #(5)!
    if: |1 # (3)!
        --8<-- "docs/gitlab/snippets/close-merge-request/warn-if.expr"
    then:
      - action: add_label # (6)!
        name: stale

      - action: comment # (9)!
        message: |
          :wave: Hello!

          This MR has not seen any commit activity for 21 days.
          We will automatically close the MR after 28 days.

          To disable this behavior, add the `do-not-close` label to the
          MR in the right menu or add comment with `/label ~"do-not-close"`

  - name: "close" # (10)!
    if: |1 # (4)!
        --8<-- "docs/gitlab/snippets/close-merge-request/close-if.expr"
    then:
      - action: close # (8)!

      - action: comment # (7)!
        message: |
          :wave: Hello!

          This MR has not seen any commit activity for 28 days.
          To keep our project clean, we will close the Merge request now.

          To disable this behavior, add the `do-not-close` label to the
          MR in the right menu or add comment with `/label ~"do-not-close"`
```

1. Add the label `stale` to MRs without activity in the last 21 days.

    The `stale` label will automatically be removed if any activity happens on the MR.

2. Syntax highlighted `script`

    ```css
    --8<-- "docs/gitlab/snippets/close-merge-request/label-script.expr"
    ```

3. Syntax highlighted `if`

    ```css
    --8<-- "docs/gitlab/snippets/close-merge-request/warn-if.expr"
    ```

4. Syntax highlighted `if`

    ```css
    --8<-- "docs/gitlab/snippets/close-merge-request/close-if.expr"
    ```

5. Send "warning" about the MR being inactive
6. Add the `stale` label to the MR (if it doesn't exists)
7. Add a comment to the MR
8. Close the MR
9. Add a comment to the MR
10. Close the MR if no activity has happened after 7 days.

    !!! question "Why 7 days?"

        The `merge_request.updated_at` updated when we commented and added the `stale` label at the 21 day mark.

        So instead we count 7 days from *that* point in time for the `close` step.

11. You can use [Twitter Bootstrap color variables](https://getbootstrap.com/docs/5.3/customize/color/#all-colors){target="_blank"} instead of HEX values.

## Add label if a file extension is modified

=== "Config"

    ```yaml
    # yaml-language-server: $schema=https://jippi.github.io/scm-engine/scm-engine.schema.json

    label:
      - name: lang/go
        color: $indigo
        script: merge_request.modified_files("*.go")

      - name: lang/markdown
        color: $indigo
        script: merge_request.modified_files("*.md")

      - name: type/documentation
        color: $green
        script: merge_request.modified_files("docs/")

      - name: go::tests::missing
        color: $red
        priority: 999
        script: |1
              merge_request.modified_files("*.go")
          && NOT merge_request.modified_files("*_test.go")

      - name: go::tests::ok
        color: $green
        priority: 999
        script: |1
              merge_request.modified_files("*.go")
          && merge_request.modified_files("*_test.go")
    ```

=== "Script with highlight"

    ```css
    merge_request.modified_files("*.go")
    ```

## Generate labels via script

=== "Config"

    ```yaml
    # yaml-language-server: $schema=https://jippi.github.io/scm-engine/scm-engine.schema.json

    label:
      # Generate list of labels via script
      - strategy: generate
        # With a description (optional)
        description: "Modified this service directory"
        # With the color $pink
        color: "$pink"
        # From this script, returning a list of labels
        script: >
          /* Generate a list of files changed in the MR inside pkg/service/ */
          merge_request.modified_files_list("pkg/service/")

          /* Remove the filename from the path
            pkg/service/example/file.go => pkg/service/example */
          | map({ filepath_dir(#) })

          /* Remove the prefix "pkg/" from the path
             pkg/service/example => service/example */
          | map({ trimPrefix(#, "pkg/") })

          /* Remove duplicate values from the output */
          | uniq()
    ```

=== "Script with highlight"

      ```css
      /* Generate a list of files changed in the MR inside pkg/service/ */
      merge_request.modified_files_list("pkg/service/")

      /* Remove the filename from the path
       *
       * pkg/service/example/file.go => pkg/service/example
       */
      | map({ filepath_dir(#) })

      /* Remove the prefix "pkg/" from the path
       *
       * pkg/service/example => service/example
       */
      | map({ trimPrefix(#, "pkg/") })

      /* Remove duplicate values from the output */
      | uniq()
      ```

## Assign reviewers from CODEOWNERS

The `codeowners` source relies on the GitLab [CODE OWNERS](https://docs.gitlab.com/ee/user/project/codeowners/) feature.

For a reviewer to be assigned to a Merge Request, the project must have a CODEOWNERS file and the groups/users mentioned in the file must have [direct membership](https://gitlab.com/gitlab-org/gitlab/-/issues/288851/) to the project.

```yaml
# yaml-language-server: $schema=https://jippi.github.io/scm-engine/scm-engine.schema.json

actions:
  - name: "assign"
    if: |1
        --8<-- "docs/gitlab/snippets/assign-merge-request/assign-if.expr"
    then:
      - action: assign_reviewers
        source: codeowners
        limit: 1
        mode: random
```

## Assign reviewers from Backstage

The `backstage` source relies on the [Backstage Catalog](https://backstage.io/docs/features/software-catalog/) to derive ownership information.

This assumes [System](https://backstage.io/docs/features/software-catalog/descriptor-format#kind-system) and [User](https://backstage.io/docs/features/software-catalog/descriptor-format/#kind-user) in Backstage can be mapped directly to a GitLab project and user.

For a Backstage System entity to be mapped to a GitLab project, the system must  have same name as the GitLab project or the `gitlab.com/project` annotation is set to the GitLab project name.

For a Backstage User entity to be mapped to a GitLab user, the user must have the `gitlab.com/user_id` annotation set to the numeric ID of the user. A plugin that can be used to automatically set this annotation is [@seatgeek/backstage-plugin-gitlab-catalog-backend](https://github.com/seatgeek/backstage-plugins/tree/main/plugins/gitlab-catalog-backend).


```yaml
# yaml-language-server: $schema=https://jippi.github.io/scm-engine/scm-engine.schema.json

actions:
  - name: "assign"
    if: |1
        --8<-- "docs/gitlab/snippets/assign-merge-request/assign-if.expr"
    then:
      - action: assign_reviewers
        source: backstage
        limit: 1
        mode: random
```

## Assign reviewers from static user IDs

The `static` source allows you to specify a fixed list of GitLab user IDs to assign as reviewers. This is useful when you want to assign specific users regardless of CODEOWNERS or Backstage configuration.

You can find a user's numeric ID by visiting their GitLab profile page - the ID is shown in the URL or on the profile.

```yaml
# yaml-language-server: $schema=https://jippi.github.io/scm-engine/scm-engine.schema.json

actions:
  - name: "assign"
    if: |1
        --8<-- "docs/gitlab/snippets/assign-merge-request/assign-if.expr"
    then:
      - action: assign_reviewers
        source: static
        user_ids:
          - "123"
          - "456"
          - "789"
        limit: 2
        mode: random
```
