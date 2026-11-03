## yt issue comment delete

Delete a comment on an issue

### Synopsis

Delete a comment on an issue permanently. It cannot be undone, so --yes is
required; without it the command names the comment and changes nothing.

```
yt issue comment delete <id> <commentID> [flags]
```

### Examples

```
  yt issue comment delete PROJ-123 4-3431 --yes
```

### Options

```
  -h, --help   help for delete
      --yes    confirm the permanent delete
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt issue comment](yt_issue_comment.md)	 - Add a comment to an issue

