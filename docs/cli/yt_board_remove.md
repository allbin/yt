## yt board remove

Remove issue(s) from a board's sprint

### Synopsis

Remove one or more issues from an agile board's sprint.

The board is a name (case-insensitive), id or URL, as for "yt board". Without --sprint the board's
current sprint is used.

Removing is idempotent: an issue not on the sprint is reported as such. A
board whose query or projects decide membership takes no removals.

```
yt board remove <board> <issue>... [flags]
```

### Examples

```
  # remove an issue from the current sprint
  yt board remove AllTix AX-812

  # remove from a specific sprint
  yt board remove AllTix AX-812 --sprint 2025-06

  # JSON output
  yt board remove AllTix AX-812 --json
```

### Options

```
  -h, --help            help for remove
      --sprint string   sprint name (default: current)
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt board](yt_board.md)	 - Show board issues or list boards

