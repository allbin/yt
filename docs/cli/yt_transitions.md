## yt transitions

List state changes from issue activity history

### Synopsis

List changes of an issue field (State by default) as recorded in YouTrack's
activity history: who changed which issue, when, and from which value to which.

This reads the actual history, so it shows what a user did — not who the
issue is assigned to now or when it was last updated. All pages of the
history are fetched.

--user takes "me" (default), a login, or a full name; "all" shows every user.
--since and --until take a date (2026-09-30), an RFC 3339 timestamp, or a
duration back from now (36h, 7d, 2w). A date-only --until includes that whole
day. --since defaults to 7d, --until to now.

--board reads the changes against an agile board: the field becomes the one
the board's columns are bound to (State, Stage, ...), each change gets the
board column it left and entered, and only issues currently on one of the
board's sprints are kept.

```
yt transitions [flags]
```

### Examples

```
  # my state changes over the last week
  yt transitions

  # same, as JSON for a standup summary
  yt transitions --json

  # a teammate's changes in a date range
  yt transitions -u alice --since 2026-09-28 --until 2026-10-04

  # everyone's moves on a board, with column names
  yt transitions -u all --board AllTix --since 14d

  # changes to another field, limited to one project
  yt transitions --field Priority -p AX
```

### Options

```
      --board string     read changes against this board's columns
      --field string     field whose changes to list (default "State")
  -h, --help             help for transitions
  -p, --project string   only issues in this project
  -q, --query string     only issues matching this YouTrack query
      --since string     start of range (date, RFC 3339, or duration ago) (default "7d")
      --until string     end of range (default: now)
  -u, --user string      who made the change ("me", login, full name, or "all") (default "me")
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt](yt.md)	 - YouTrack CLI

