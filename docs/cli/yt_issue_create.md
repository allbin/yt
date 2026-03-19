## yt issue create

Create a new YouTrack issue

### Synopsis

Create a new issue in the specified YouTrack project. Requires a project
short name and summary. Optionally accepts a description.

The created issue is displayed after creation.

```
yt issue create [flags]
```

### Examples

```
  # create a minimal issue
  yt issue create -p PROJ -s "Fix login bug"

  # create with description
  yt issue create -p PROJ -s "Add dark mode" -d "Support system-level dark mode preference"

  # output as JSON
  yt issue create -p PROJ -s "New feature" --json
```

### Options

```
  -d, --description string   issue description
  -h, --help                 help for create
  -p, --project string       project short name (required)
  -s, --summary string       issue summary (required)
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt issue](yt_issue.md)	 - Show or manage YouTrack issues

