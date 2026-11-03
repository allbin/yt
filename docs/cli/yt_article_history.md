## yt article history

Show every change to an article

### Synopsis

List every recorded change to a knowledge base article, oldest first: creation,
title and content edits, comments, tags, and attachments. Moving an article
is not recorded by YouTrack.

Each change has a kind: created, summary, content, comment, tag, attachment,
comment-attachment. --kind keeps only the listed kinds.

Text output cuts long text to its first line; --full prints it whole, with
the content before and after each edit. --json always carries the full text.

```
yt article history <id> [flags]
```

### Examples

```
  yt article history PROJ-A-12

  # who changed the content, and how
  yt article history PROJ-A-12 --kind content --full
```

### Options

```
      --full           print long text whole
  -h, --help           help for history
      --kind strings   only these kinds of change (repeatable or comma-separated)
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt article](yt_article.md)	 - Show or manage knowledge base articles

