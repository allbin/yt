## yt issue update

Update a YouTrack issue

### Synopsis

Update fields on a YouTrack issue by executing a command string.
Supports setting state, assignee, priority, and type. Multiple flags
can be combined in a single invocation.

After a successful update the issue is fetched and displayed.

```
yt issue update <id> [flags]
```

### Examples

```
  # set state
  yt issue update PROJ-123 -s "In Progress"

  # set assignee and priority
  yt issue update PROJ-123 -a me -p Critical

  # set type
  yt issue update PROJ-123 -t Bug

  # combine all fields
  yt issue update PROJ-123 -s Open -a john -p Normal -t Task
```

### Options

```
  -a, --assignee string   set assignee (supports 'me')
  -h, --help              help for update
  -p, --priority string   set priority
  -s, --state string      set issue state
  -t, --type string       set issue type
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt issue](yt_issue.md)	 - Show or manage YouTrack issues

