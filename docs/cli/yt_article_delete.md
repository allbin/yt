## yt article delete

Delete a knowledge base article

### Synopsis

Delete a knowledge base article permanently. It cannot be undone, so --yes is
required. An article with sub-articles is refused: delete or move them
first, so a whole tree never goes in one call.

```
yt article delete <id> [flags]
```

### Examples

```
  yt article delete PROJ-A-12 --yes
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

* [yt article](yt_article.md)	 - Show or manage knowledge base articles

