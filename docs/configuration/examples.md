# Examples

!!! note

    A quick demo of what SCM Engine can do.

    The `script` field is a [expr-lang](https://expr-lang.org/) expression, a safe, fast, and intuitive expression evaluator.

## Close Merge Request without recent commit activity

!!! warning "When adopting this script on existing projects"

    The script will *NOT* wait 7 days between warning and closing

    * On the first run, all MRs with commits older than 21 days will be warned
    * On the second run, all MRs with commits older than 28 days will be closed.

This example will close a Merge Request if no commits has been made for 28 days.

The script will warn at 21 days mark that this will happen.

```{.yaml linenums=1}
label:
  - name: "mark MR as stale" # (1)!
    color: $red
    script: |1 # (2)!
        --8<-- "docs/configuration/snippets/close-merge-request/label-script.expr"

actions:
  - name: "warn" #(5)!
    if: |1 # (3)!
        --8<-- "docs/configuration/snippets/close-merge-request/warn-if.expr"
    then:
      - action: add_label
        name: stale

      - action: comment
        message: |
          :wave: Hello!

          This MR has not seen any commit activity for 21 days.
          We will automatically close the MR after 28 days.

          To disable this behavior, add the `do-not-close` label to the
          MR in the right menu or add comment with `/label ~"do-not-close"`

  - name: "close"
    if: |1 # (4)!
        --8<-- "docs/configuration/snippets/close-merge-request/close-if.expr"
    then:
      - action: close
      - action: comment
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
    --8<-- "docs/configuration/snippets/close-merge-request/label-script.expr"
    ```

3. Syntax highlighted `if`

    ```css
    --8<-- "docs/configuration/snippets/close-merge-request/warn-if.expr"
    ```

4. Syntax highlighted `if`

    ```css
    --8<-- "docs/configuration/snippets/close-merge-request/close-if.expr"
    ```

5. Send "warning" about the MR being inactive

## Add label if a file extension is modified

=== "Config"

    ```yaml
    label:
        # Add a label named "lang/go"
      - name: lang/go
        # and a label description (optional)
        description: "Modified Go files"
        # and the color $indigo
        color: "$indigo"
        # if files matching "*.go" was modified
        script: merge_request.modified_files("*.go")
    ```

=== "Script with highlight"

    ```css
    merge_request.modified_files("*.go")
    ```

## Generate labels via script

=== "Config"

    ```yaml
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
