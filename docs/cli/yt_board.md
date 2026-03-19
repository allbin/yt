## yt board

Show board issues or list boards

### Synopsis

Show issues on an agile board's sprint. Looks up the board by name
(case-insensitive). Defaults to the current sprint.

Filters are handled server-side by YouTrack, so assignee supports "me",
login names, and full names.

Use subcommands to list available boards.

```
yt board [name] [flags]
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

  # JSON output
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
* [yt board list](yt_board_list.md)	 - List agile boards
* [yt board view](yt_board_view.md)	 - Open interactive board viewer

