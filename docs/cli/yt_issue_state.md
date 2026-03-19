## yt issue state

Interactively set issue state

### Synopsis

Open an interactive picker to change the state of a YouTrack issue.

Shows all available states for the issue's project with the current state
marked. Navigate with arrow keys or j/k, select with Enter, cancel with
Esc or q.

If no ID is given, attempts to detect it from the current git branch name.

```
yt issue state [id] [flags]
```

### Examples

```
  # pick state interactively
  yt issue state PROJ-123

  # auto-detect from current branch
  yt issue state
```

### Options

```
  -h, --help   help for state
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt issue](yt_issue.md)	 - Show or manage YouTrack issues

