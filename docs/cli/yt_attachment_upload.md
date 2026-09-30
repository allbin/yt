## yt attachment upload

Attach files to an issue

### Synopsis

Upload one or more local files as attachments on a YouTrack issue. Each
attachment is named after its file's base name. All files are sent in one
request; nothing is uploaded if a file cannot be opened.

Reference an uploaded image in a description or comment as ![](name.png).

```
yt attachment upload <issueID> <file>... [flags]
```

### Examples

```
  # attach a screenshot
  yt attachment upload PROJ-123 screenshot.png

  # attach several files
  yt attachment upload PROJ-123 log.txt trace.json

  # JSON output (the created attachments)
  yt attachment upload PROJ-123 screenshot.png --json
```

### Options

```
  -h, --help   help for upload
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt attachment](yt_attachment.md)	 - Manage issue attachments

