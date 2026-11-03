## yt article create

Create a knowledge base article

### Synopsis

Create a knowledge base article in a project, optionally as a sub-article of
another. With --parent the project defaults to the parent's.

Content is markdown. Issue IDs written in it (PROJ-123) become mentions: the
issue shows the article under "Articles" in "yt issue", and the article lists
the issue under "Mentions".

--content accepts "@path" to read a file or "-" to read stdin.

```
yt article create [flags]
```

### Examples

```
  # a top-level article
  yt article create -p PROJ -s "Onboarding"

  # meeting notes under a notes page, content from a file
  yt article create --parent PROJ-A-3 -s "Planning 2026-11-03" -c @notes.md

  # content from stdin
  cat notes.md | yt article create -p PROJ -s "Retro" -c - --json
```

### Options

```
  -c, --content string   markdown content (@file or - for stdin)
  -h, --help             help for create
      --parent string    create as a sub-article of this article
  -p, --project string   project short name (default: the parent's)
  -s, --summary string   article title (required)
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt article](yt_article.md)	 - Show or manage knowledge base articles

