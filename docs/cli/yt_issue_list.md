## yt issue list

List YouTrack issues

### Synopsis

List YouTrack issues with optional filters. Filters are combined into a
YouTrack search query. Use --query for arbitrary YouTrack query syntax.

```
yt issue list [flags]
```

### Examples

```
  # list issues in a project
  yt issue list -p PROJ

  # list open issues assigned to me
  yt issue list -p PROJ -s Open -a me

  # arbitrary YouTrack query
  yt issue list -q "tag: {Ready for QA} sort by: updated desc"

  # combine filters with raw query
  yt issue list -p PROJ -q "created: today"

  # output as JSON, limit to 5 results
  yt issue list -p PROJ -n 5 --json
```

### Options

```
  -a, --assignee string   filter by assignee
  -h, --help              help for list
  -n, --limit int         max results (default 20)
  -p, --project string    filter by project
  -q, --query string      raw YouTrack query
  -s, --state string      filter by state
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt issue](yt_issue.md)	 - Show or manage YouTrack issues

