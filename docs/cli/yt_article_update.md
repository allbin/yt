## yt article update

Change an article's title or content

### Synopsis

Change a knowledge base article's title, replace its content, or append to it.
Title and content are saved in one request.

--content replaces the whole content. --append adds text after the current
content, separated by a blank line: it reads the article and writes it back,
so an edit someone saves between the two is lost. For notes several people
write at once, give each their own sub-article.

--content and --append accept "@path" to read a file or "-" to read stdin.

```
yt article update <id> [flags]
```

### Examples

```
  # rename
  yt article update PROJ-A-12 -s "Planning 2026-11-03"

  # replace the content from a file
  yt article update PROJ-A-12 -c @notes.md

  # append a section
  yt article update PROJ-A-12 --append "## Decisions
- ship PROJ-123 this sprint"
```

### Options

```
      --append string    append to content (@file or - for stdin)
  -c, --content string   replace content (@file or - for stdin)
  -h, --help             help for update
  -s, --summary string   new title
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt article](yt_article.md)	 - Show or manage knowledge base articles

