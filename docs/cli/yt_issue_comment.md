## yt issue comment

Add a comment to an issue

### Synopsis

Post a new comment on a YouTrack issue. The comment text is provided
via the --message flag.

The message accepts "@path" to read from a file or "-" to read from stdin,
which avoids shell mangling of multi-line text.

```
yt issue comment <id> [flags]
```

### Examples

```
  # add a comment
  yt issue comment PROJ-123 -m "Looks good, merging."

  # read the comment from a file
  yt issue comment PROJ-123 -m @review.md

  # read the comment from stdin
  cat review.md | yt issue comment PROJ-123 -m -

  # JSON output of created comment
  yt issue comment PROJ-123 -m "Done" --json
```

### Options

```
  -h, --help             help for comment
  -m, --message string   comment text (@file or - for stdin) (required)
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt issue](yt_issue.md)	 - Show or manage YouTrack issues

