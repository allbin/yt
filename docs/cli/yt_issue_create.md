## yt issue create

Create a new YouTrack issue

### Synopsis

Create a new issue in the specified YouTrack project. Requires a project
short name and summary. Optionally accepts a description.

The created issue is displayed after creation.

Use --subsystem or --field to set custom fields on the new issue.

```
yt issue create [flags]
```

### Examples

```
  # create a minimal issue
  yt issue create -p PROJ -s "Fix login bug"

  # create with description
  yt issue create -p PROJ -s "Add dark mode" -d "Support system-level dark mode preference"

  # create with subsystem
  yt issue create -p PROJ -s "Fix API auth" --subsystem API

  # create with custom field
  yt issue create -p PROJ -s "Critical outage" --field "Severity=Critical"

  # create with tags
  yt issue create -p PROJ -s "Fix stale state" -t tech-debt -t scheduler

  # output as JSON
  yt issue create -p PROJ -s "New feature" --json
```

### Options

```
  -d, --description string   issue description
      --field strings        set custom field as "Name=Value" (repeatable)
  -h, --help                 help for create
  -p, --project string       project short name (required)
      --subsystem string     set subsystem
  -s, --summary string       issue summary (required)
  -t, --tag strings          add tag (repeatable)
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt issue](yt_issue.md)	 - Show or manage YouTrack issues

