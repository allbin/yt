## yt issue

Show or manage YouTrack issues

### Synopsis

Fetch a single YouTrack issue by its readable ID (e.g. PROJ-123) and display
its summary, state, assignee, priority, type, subsystems, tags, and description.

If no ID is given, attempts to detect it from the current git branch name.

Use subcommands to list and filter issues.

```
yt issue [id] [flags]
```

### Examples

```
  # show an issue as formatted text
  yt issue PROJ-123

  # show an issue as JSON
  yt issue PROJ-123 --json

  # auto-detect from current branch (e.g. proj-123-some-slug)
  yt issue
```

### Options

```
  -h, --help   help for issue
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt](yt.md)	 - YouTrack CLI
* [yt issue list](yt_issue_list.md)	 - List YouTrack issues

