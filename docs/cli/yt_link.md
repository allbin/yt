## yt link

Create links between issues

### Synopsis

Link an issue to one or more target issues using a directed relation phrase.

The relation is matched against the instance's link types and is accepted in
kebab, spaced, or squashed form (e.g. "subtask-of", "subtask of", "subtaskof").
Run "yt link types" to list the available relations.

Linking is idempotent: a link that already exists is reported as such and left
unchanged.

```
yt link <id> <relation> <target-id>... [flags]
```

### Examples

```
  # make AX-804 a subtask of AX-332
  yt link AX-804 subtask-of AX-332

  # relate two issues
  yt link AX-1 relates AX-2

  # declare a dependency
  yt link AX-1 depends-on AX-3

  # mark a duplicate
  yt link AX-1 duplicates AX-4

  # link to several targets at once
  yt link AX-1 relates AX-2 AX-3 AX-4

  # JSON output (the source issue's links after the change)
  yt link AX-804 subtask-of AX-332 --json
```

### Options

```
  -h, --help   help for link
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt](yt.md)	 - YouTrack CLI
* [yt link types](yt_link_types.md)	 - List available issue link types

