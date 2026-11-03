## yt article comment delete

Delete a comment on an article

### Synopsis

Delete a comment on an article permanently. It cannot be undone, so --yes is
required; without it the command names the comment and changes nothing.

```
yt article comment delete <id> <commentID> [flags]
```

### Examples

```
  yt article comment delete PROJ-A-12 4-3431 --yes
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

* [yt article comment](yt_article_comment.md)	 - Add a comment to an article

