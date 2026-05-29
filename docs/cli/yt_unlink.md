## yt unlink

Remove a link between issues

### Synopsis

Remove an existing link from an issue to a target issue.

The relation is matched the same way as "yt link" (kebab, spaced, or squashed
form). The matching link must already exist or the command errors.

```
yt unlink <id> <relation> <target-id> [flags]
```

### Examples

```
  # remove a subtask link
  yt unlink AX-804 subtask-of AX-332

  # remove a relation
  yt unlink AX-1 relates AX-2

  # JSON output (the source issue's remaining links)
  yt unlink AX-804 subtask-of AX-332 --json
```

### Options

```
  -h, --help   help for unlink
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt](yt.md)	 - YouTrack CLI

