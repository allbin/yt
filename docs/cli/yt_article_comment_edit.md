## yt article comment edit

Replace the text of a comment on an article

### Synopsis

Replace the whole text of a comment on an article. The comment ID is the
"id" in "yt article comments --json", or the one "yt article comment" printed.
YouTrack lets you edit only comments you may change, usually your own.

The message accepts "@path" to read from a file or "-" to read from stdin.

```
yt article comment edit <id> <commentID> [flags]
```

### Examples

```
  yt article comment edit PROJ-A-12 4-3431 -m "Corrected: deploy on Friday."

  # from a file
  yt article comment edit PROJ-A-12 4-3431 -m @comment.md --json
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

* [yt article comment](yt_article_comment.md)	 - Add a comment to an article

