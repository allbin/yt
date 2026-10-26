## yt board

Show board issues or list boards

### Synopsis

Show the issues on an agile board. The board is a name (case-insensitive),
an id (124-66) or a board URL (https://<host>/agiles/124-66/current). A board
private to its owner is missing from the board list, so its name is not found;
its id or URL still works.

A board is one of four kinds, by what puts an issue on it: "sprints" (the
issue is assigned to one of the board's sprints; shown for the current sprint
unless --sprint names another), "query" (the issue matches the board's query),
"projects" (every issue of the board's projects) or "manual" (added by hand).
The last three have no sprints. --json prints the kind, what membership means
for it, the sprint shown, and the issues.

Filters are handled server-side by YouTrack, so assignee supports "me",
login names, and full names.

Use subcommands to list available boards.

```
yt board [board] [flags]
```

### Examples

```
  # show current sprint issues
  yt board HållKoll

  # specific sprint
  yt board HållKoll --sprint 2025-02

  # filter by state
  yt board HållKoll -s "In Progress"

  # issues assigned to me
  yt board HållKoll -a me

  # combine filters with extra query
  yt board HållKoll -a me -q "sort by: Priority"

  # by id or URL, for a board not in "yt board list"
  yt board 124-66
  yt board https://example.youtrack.cloud/agiles/124-66/current

  # JSON output: kind, membership, sprint and issues
  yt board HållKoll --json
```

### Options

```
  -a, --assignee string   filter by assignee (supports 'me')
  -h, --help              help for board
  -q, --query string      additional YouTrack query
      --sprint string     sprint name (default: current)
  -s, --state string      filter by state
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt](yt.md)	 - YouTrack CLI
* [yt board add](yt_board_add.md)	 - Add issue(s) to a board's sprint
* [yt board list](yt_board_list.md)	 - List agile boards
* [yt board remove](yt_board_remove.md)	 - Remove issue(s) from a board's sprint
* [yt board view](yt_board_view.md)	 - Open interactive board viewer

