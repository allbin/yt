## yt issue create

Create a new YouTrack issue

### Synopsis

Create a new issue in the specified YouTrack project. Requires a project
short name and summary. Optionally accepts a description.

The created issue is displayed after creation.

Use --subsystem or --field to set custom fields on the new issue.

The description accepts "@path" to read from a file or "-" to read from stdin,
which avoids shell mangling of multi-line text.

Use --board (with optional --sprint) to place the new issue on an agile board,
or --like to mirror another issue's board and sprint.

Use --parent <id> to create the issue as a subtask of another issue: it adds
a "subtask of" link to the parent and places the new issue on the parent's
board and sprint in one step -- the common "subtask on the parent's board"
workflow. When --board is also given it overrides the parent's board.

```
yt issue create [flags]
```

### Examples

```
  # create a minimal issue
  yt issue create -p PROJ -s "Fix login bug"

  # create with description
  yt issue create -p PROJ -s "Add dark mode" -d "Support system-level dark mode preference"

  # read the description from a file
  yt issue create -p PROJ -s "Big writeup" -d @notes.md

  # read the description from stdin
  cat notes.md | yt issue create -p PROJ -s "Big writeup" -d -

  # create with subsystem
  yt issue create -p PROJ -s "Fix API auth" --subsystem API

  # create with custom field
  yt issue create -p PROJ -s "Critical outage" --field "Severity=Critical"

  # create with tags
  yt issue create -p PROJ -s "Fix stale state" -t tech-debt -t scheduler

  # place the new issue on a board's current sprint
  yt issue create -p AX -s "Subtask" --board AllTix

  # mirror another issue's board+sprint without linking
  yt issue create -p AX -s "Subtask" --like AX-332

  # create a subtask: link to the parent AND share its board+sprint
  yt issue create -p AX -s "Subtask" --parent AX-332

  # output as JSON
  yt issue create -p PROJ -s "New feature" --json
```

### Options

```
      --board string         add the issue to this agile board
  -d, --description string   issue description (@file or - for stdin)
      --field strings        set custom field as "Name=Value" (repeatable)
  -h, --help                 help for create
      --like string          mirror another issue's board and sprint
      --parent string        make the issue a subtask of this issue and share its board
  -p, --project string       project short name (required)
      --sprint string        sprint for --board (default: current)
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

