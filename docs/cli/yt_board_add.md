## yt board add

Add issue(s) to a board's sprint

### Synopsis

Add one or more issues to an agile board by placing them on a sprint.

The board is matched by name (case-insensitive). Without --sprint the board's
current sprint is used.

Adding is idempotent: an issue already on the sprint is reported as such and
left unchanged.

```
yt board add <board> <issue>... [flags]
```

### Examples

```
  # add an issue to the current sprint
  yt board add AllTix AX-812

  # add several issues to a specific sprint
  yt board add AllTix AX-812 AX-813 --sprint 2025-06

  # JSON output
  yt board add AllTix AX-812 --json
```

### Options

```
  -h, --help            help for add
      --sprint string   sprint name (default: current)
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt board](yt_board.md)	 - Show board issues or list boards

