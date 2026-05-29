## yt links

List an issue's links

### Synopsis

List the links on a YouTrack issue, grouped by relation, showing the
directed phrase and each linked issue's ID and summary.

If no ID is given, attempts to detect it from the current git branch name.

```
yt links [id] [flags]
```

### Examples

```
  # list links for an issue
  yt links AX-804

  # auto-detect from current branch
  yt links

  # JSON output
  yt links AX-804 --json
```

### Options

```
  -h, --help   help for links
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt](yt.md)	 - YouTrack CLI

