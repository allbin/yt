## yt issue update

Update a YouTrack issue

### Synopsis

Update fields on a YouTrack issue.

Summary and description use the REST API; other fields use the command API.
Both can be combined in a single invocation.

Use --field to set arbitrary custom fields by name.

After a successful update the issue is fetched and displayed.

```
yt issue update <id> [flags]
```

### Examples

```
  # set state
  yt issue update PROJ-123 -s "In Progress"

  # update summary
  yt issue update PROJ-123 -S "New title"

  # update description
  yt issue update PROJ-123 -d "Updated description"

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

  # combine REST and command fields
  yt issue update PROJ-123 -S "New title" -s "In Progress" -a me
```

### Options

```
  -a, --assignee string      set assignee (supports 'me')
  -d, --description string   set issue description
      --field strings        set custom field as "Name=Value" (repeatable)
  -h, --help                 help for update
  -p, --priority string      set priority
      --remove-tag strings   remove tag (repeatable)
  -s, --state string         set issue state
      --subsystem string     set subsystem
  -S, --summary string       set issue summary
      --tag strings          add tag (repeatable)
  -t, --type string          set issue type
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt issue](yt_issue.md)	 - Show or manage YouTrack issues

