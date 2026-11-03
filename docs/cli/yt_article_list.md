## yt article list

Search knowledge base articles

### Synopsis

List knowledge base articles, without their content. --project and --query
combine into one YouTrack search; the query takes free text and attributes
such as "mentions: PROJ-123". --parent lists the direct sub-articles of an
article instead.

Text output is a tree: each article under its parent when the parent is in
the result, in the order the knowledge base shows them. --json gives each
article's id, summary, project, parent, hasChildren, and timestamps; read one
with "yt article <id>" for its content.

```
yt article list [flags]
```

### Examples

```
  # every article in a project
  yt article list -p PROJ

  # full-text search
  yt article list -q "retro"

  # articles mentioning an issue
  yt article list -q "mentions: PROJ-123"

  # sub-articles of a page, e.g. a meeting notes folder
  yt article list --parent PROJ-A-3 --json
```

### Options

```
  -h, --help             help for list
  -n, --limit int        max results (default 100)
      --parent string    list the sub-articles of this article
  -p, --project string   filter by project
  -q, --query string     search text or YouTrack query
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt article](yt_article.md)	 - Show or manage knowledge base articles

