## yt issue update

Update a YouTrack issue

### Synopsis

Update fields on a YouTrack issue by executing a command string.
Supports setting state, assignee, priority, type, and subsystem.
Multiple flags can be combined in a single invocation.

Use --field to set arbitrary custom fields by name.

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

  # set subsystem
  yt issue update PROJ-123 --subsystem API

  # set arbitrary custom field
  yt issue update PROJ-123 --field "Severity=Critical"

  # add tags
  yt issue update PROJ-123 --tag tech-debt --tag scheduler

  # remove a tag
  yt issue update PROJ-123 --remove-tag obsolete

  # combine all fields
  yt issue update PROJ-123 -s Open -a john -p Normal -t Task --subsystem API
```

### Options

```
  -a, --assignee string      set assignee (supports 'me')
      --field strings        set custom field as "Name=Value" (repeatable)
  -h, --help                 help for update
  -p, --priority string      set priority
      --remove-tag strings   remove tag (repeatable)
  -s, --state string         set issue state
      --subsystem string     set subsystem
      --tag strings          add tag (repeatable)
  -t, --type string          set issue type
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt issue](yt_issue.md)	 - Show or manage YouTrack issues

