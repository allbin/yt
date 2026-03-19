## yt issue comment

Add a comment to an issue

### Synopsis

Post a new comment on a YouTrack issue. The comment text is provided
via the --message flag.

```
yt issue comment <id> [flags]
```

### Examples

```
  # add a comment
  yt issue comment PROJ-123 -m "Looks good, merging."

  # JSON output of created comment
  yt issue comment PROJ-123 -m "Done" --json
```

### Options

```
  -h, --help             help for comment
  -m, --message string   comment text (required)
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt issue](yt_issue.md)	 - Show or manage YouTrack issues

