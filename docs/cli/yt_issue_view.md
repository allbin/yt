## yt issue view

Open interactive issue viewer

### Synopsis

Open a full-screen interactive viewer for a YouTrack issue.

Shows issue summary, metadata, description, and comments in a scrollable
viewport. Supports changing issue state via an embedded state picker.

If no ID is given, attempts to detect it from the current git branch name.

```
yt issue view [id] [flags]
```

### Examples

```
  # open viewer for a specific issue
  yt issue view PROJ-123

  # auto-detect from current branch
  yt issue view
```

### Options

```
  -h, --help   help for view
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt issue](yt_issue.md)	 - Show or manage YouTrack issues

