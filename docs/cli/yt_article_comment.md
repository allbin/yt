## yt article comment

Add a comment to an article

### Synopsis

Post a new comment on a knowledge base article. The comment text is provided
via the --message flag.

The message accepts "@path" to read from a file or "-" to read from stdin.

```
yt article comment <id> [flags]
```

### Examples

```
  # add a comment
  yt article comment PROJ-A-12 -m "Action items moved to PROJ-123."

  # JSON output of created comment
  yt article comment PROJ-A-12 -m @comment.md --json
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

* [yt article](yt_article.md)	 - Show or manage knowledge base articles

