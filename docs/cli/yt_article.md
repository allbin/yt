## yt article

Show or manage knowledge base articles

### Synopsis

Fetch a knowledge base article by its readable ID (e.g. PROJ-A-12) and display
its summary, project, parent and sub-articles, tags, the issues its content
mentions, its content, and its attachments.

Use subcommands to search, create, update, comment on, and delete articles.
Attachments on an article go through "yt attachment" with the article ID.

```
yt article <id> [flags]
```

### Examples

```
  # show an article
  yt article PROJ-A-12

  # show an article as JSON
  yt article PROJ-A-12 --json
```

### Options

```
  -h, --help   help for article
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt](yt.md)	 - YouTrack CLI
* [yt article comment](yt_article_comment.md)	 - Add a comment to an article
* [yt article comments](yt_article_comments.md)	 - List comments on an article
* [yt article create](yt_article_create.md)	 - Create a knowledge base article
* [yt article delete](yt_article_delete.md)	 - Delete a knowledge base article
* [yt article list](yt_article_list.md)	 - Search knowledge base articles
* [yt article update](yt_article_update.md)	 - Change an article's title or content

