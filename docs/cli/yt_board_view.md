## yt board view

Open interactive board viewer

### Synopsis

Open a full-screen interactive viewer for an agile board.

Shows board columns, swimlanes, and issue cards in a navigable grid.
Supports changing issue state, opening issue details, and refreshing.

Defaults to the current sprint unless --sprint is specified.

```
yt board view [name] [flags]
```

### Examples

```
  # open board viewer
  yt board view HållKoll

  # specific sprint
  yt board view HållKoll --sprint 2025-02
```

### Options

```
  -h, --help            help for view
      --sprint string   sprint name (default: current)
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt board](yt_board.md)	 - Show board issues or list boards

