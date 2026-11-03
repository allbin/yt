## yt issue comment edit

Replace the text of a comment on an issue

### Synopsis

Replace the whole text of a comment on an issue. The comment ID is the
"id" in "yt issue comments --json", or the one "yt issue comment" printed.
YouTrack lets you edit only comments you may change, usually your own.

The message accepts "@path" to read from a file or "-" to read from stdin.

```
yt issue comment edit <id> <commentID> [flags]
```

### Examples

```
  yt issue comment edit PROJ-123 4-3431 -m "Corrected: deploy on Friday."

  # from a file
  yt issue comment edit PROJ-123 4-3431 -m @comment.md --json
```

### Options

```
  -h, --help             help for edit
  -m, --message string   new comment text (@file or - for stdin) (required)
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt issue comment](yt_issue_comment.md)	 - Add a comment to an issue

