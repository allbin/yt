## yt issue update

Update a YouTrack issue

### Synopsis

Update fields on a YouTrack issue.

Use --field to set any custom field by name, as "Name=Value". Values with
spaces need no escaping beyond shell quoting. Repeat --field with the same name
to set several values on a multi-value field; the list replaces the current
values. An empty value clears the field. To keep the current values of a
multi-value field, use --add-field and --remove-field instead.

--state, --assignee, --priority, --type and --subsystem are shorthands for
--field on those fields, so --subsystem also replaces. Assignee accepts "me",
a login or a name.

Values are checked against the field's allowed values before anything is
written ("yt project fields PROJ" lists them), and summary, description and
fields are saved in one request: either all of them change or none do.

--tag and --remove-tag name tags by their whole name: the exact name first,
then a case-insensitive match when only one tag has it. A name with no tag is
an error that lists the closest tags; --create-tag creates it instead. Tags
are written by ID and read back afterwards, so a tag that did not stick is an
error. Nothing goes through YouTrack's command language, so multi-word values
need only shell quoting; braces are refused.

The description accepts "@path" to read from a file or "-" to read from stdin,
which avoids shell mangling of multi-line text.

Use --board (with optional --sprint) to also place the issue on an agile board.

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

  # set a multi-word value
  yt issue update PROJ-123 -t "User Story"

  # set subsystem
  yt issue update PROJ-123 --subsystem "Management UI"

  # set several values on a multi-value field
  yt issue update PROJ-123 --field "Subsystem=API" --field "Subsystem=Management UI"

  # add to / remove from a multi-value field, keeping other values
  yt issue update PROJ-123 --add-field "Subsystem=API" --remove-field "Subsystem=Mobile"

  # clear a field
  yt issue update PROJ-123 --field "Subsystem="

  # set arbitrary custom field
  yt issue update PROJ-123 --field "Severity=Critical"

  # add tags, including a multi-word one
  yt issue update PROJ-123 --tag tech-debt --tag "To be discussed"

  # create a tag that does not exist yet and apply it
  yt issue update PROJ-123 --tag "needs triage" --create-tag

  # remove a tag
  yt issue update PROJ-123 --remove-tag obsolete

  # read the description from a file
  yt issue update PROJ-123 -d @notes.md

  # place the issue on a board's current sprint
  yt issue update AX-812 --board AllTix

  # several changes at once
  yt issue update PROJ-123 -S "New title" -s "In Progress" -a me
```

### Options

```
      --add-field stringArray      add a value to a multi-value field, as "Name=Value" (repeatable)
  -a, --assignee string            set assignee (supports 'me')
      --board string               add the issue to this agile board (name, id or URL)
      --create-tag                 create --tag names that do not exist yet
  -d, --description string         set issue description
      --field stringArray          set custom field as "Name=Value" (repeatable)
  -h, --help                       help for update
  -p, --priority string            set priority
      --remove-field stringArray   remove a value from a multi-value field, as "Name=Value" (repeatable)
      --remove-tag stringArray     remove tag (repeatable)
      --sprint string              sprint for --board (default: current)
  -s, --state string               set issue state
      --subsystem string           set subsystem, replacing current values (see --add-field)
  -S, --summary string             set issue summary
      --tag stringArray            add tag (repeatable)
  -t, --type string                set issue type
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt issue](yt_issue.md)	 - Show or manage YouTrack issues

