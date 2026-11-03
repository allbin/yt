## yt article update

Change an article's title, content, place, or tags

### Synopsis

Change a knowledge base article's title, replace or append to its content,
move it under another article, or add and remove tags. Title, content and
parent are saved in one request; tags follow, and are read back after.

--content replaces the whole content. --append adds text after the current
content, separated by a blank line: it reads the article and writes it back,
so an edit someone saves between the two is lost. For notes several people
write at once, give each their own sub-article.

--parent moves the article, with its sub-articles, under another article;
--parent none moves it to the top level of its project.

--tag and --remove-tag match tag names as "yt issue update" does: an unknown
name errors with the closest tags, and --create-tag creates it instead.

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

  # move under a notes page, and tag it
  yt article update PROJ-A-12 --parent PROJ-A-3 --tag meeting

  # move to the top level
  yt article update PROJ-A-12 --parent none
```

### Options

```
      --append string            append to content (@file or - for stdin)
  -c, --content string           replace content (@file or - for stdin)
      --create-tag               create --tag names that do not exist yet
  -h, --help                     help for update
      --parent string            move under this article ("none": to the top level)
      --remove-tag stringArray   remove tag (repeatable)
  -s, --summary string           new title
      --tag stringArray          add tag (repeatable)
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt article](yt_article.md)	 - Show or manage knowledge base articles

