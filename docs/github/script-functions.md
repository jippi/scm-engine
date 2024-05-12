# Script Functions

!!! tip "The [Expr Language Definition](https://expr-lang.org/docs/language-definition) is a great resource to learn more about the language"

## pull_request

### `pull_request.state_is(string...) -> boolean` {: #pull_request.state_is data-toc-label="state_is"}

Check if the `pull_request` state is any of the provided states

**Valid options**:

- `CLOSED` - In closed state
- `MERGED` - Pull Request has been merged
- `OPEN` - Opened Pull Request

```css
pull_request.state_is("MERGED")
pull_request.state_is("CLOSED", "MERGED")
```

### `pull_request.modified_files(string...) -> boolean` {: #pull_request.modified_files data-toc-label="modified_files"}

Returns wether any of the provided files patterns have been modified in the Pull Request.

The file patterns use the [`.gitignore` format](https://git-scm.com/docs/gitignore#_pattern_format).

```css
pull_request.modified_files("*.go", "docs/") == true
```

### `pull_request.modified_files_list(string...) -> []string` {: #pull_request.modified_files_list data-toc-label="modified_files_list"}

Returns an array of files matching the provided (optional) pattern thas has been modified in the Pull Request.

The file patterns use the [`.gitignore` format](https://git-scm.com/docs/gitignore#_pattern_format).

```css
pull_request.modified_files_list("*.go", "docs/") == ["example/file.go", "docs/index.md"]
```

### `pull_request.has_label(string) -> boolean` {: #pull_request.has_label data-toc-label="has_label"}

Returns wether any of the provided label exist on the Pull Request.

```css
pull_request.labels = ["hello"]
pull_request.has_label("hello") == true
pull_request.has_label("world") == false
```

### `pull_request.has_no_label(string) -> boolean` {: #pull_request.has_no_label data-toc-label="has_no_label"}

Returns wether the Pull Request has the provided label or not.

```css
pull_request.labels = ["hello"]
pull_request.has_no_label("hello") == false
pull_request.has_no_label("world") == true
```

## Global

### `duration(string) -> duration` {: #duration data-toc-label="duration"}

Returns the [`time.Duration`](https://pkg.go.dev/time#Duration) value of the given string str.

Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d" and "w".

```css
duration("1h").Seconds() == 3600
```

### `uniq([]string) -> []string` {: #uniq data-toc-label="uniq"}

Returns a new array where all duplicate values has been removed.

```css
(["hello", "world", "world"] | uniq) == ["hello", "world"]
```

### `filepath_dir` {: #filepath_dir data-toc-label="filepath_dir"}

`filepath_dir` returns all but the last element of path, typically the path's directory. After dropping the final element,

Dir calls [Clean](https://pkg.go.dev/path/filepath#Clean) on the path and trailing slashes are removed.

If the path is empty, `filepath_dir` returns ".". If the path consists entirely of separators, `filepath_dir` returns a single separator.

The returned path does not end in a separator unless it is the root directory.

```css
filepath_dir("example/directory/file.go") == "example/directory"
```

### `limit_path_depth_to` {: #limit_path_depth_to data-toc-label="limit_path_depth_to"}

`limit_path_depth_to` takes a path structure, and limits it to the configured maximum depth. Particularly useful when using `generated` labels from a directory structure, and want to to have a label naming scheme that only uses path of the path.

```css
limit_path_depth_to("path1/path2/path3/path4", 2), == "path1/path2"
limit_path_depth_to("path1/path2", 3), == "path1/path2"
```
