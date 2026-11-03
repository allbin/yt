## yt issue history

Show every change to an issue

### Synopsis

List every recorded change to an issue, oldest first: creation, summary and
description edits, custom fields, links, tags, attachments, comments and their
edits, sprints, project moves, visibility, resolution, and commits.

Each change has a kind: created, summary, description, field, link, tag,
attachment, attachment-rename, comment, comment-edit, comment-attachment,
resolved, sprint, project, visibility, commit, pull-request. --kind keeps only
the listed kinds.

Text output cuts long text to its first line; --full prints it whole, with a
description's text before and after. --json always carries the full text.

If no ID is given, attempts to detect it from the current git branch name.
For State changes across many issues, use "yt transitions".

```
yt issue history [id] [flags]
```

### Examples

```
  # everything that happened to an issue
  yt issue history PROJ-123

  # only field changes and comments
  yt issue history PROJ-123 --kind field,comment

  # full text of description edits
  yt issue history PROJ-123 --kind description --full
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

* [yt issue](yt_issue.md)	 - Show or manage YouTrack issues

