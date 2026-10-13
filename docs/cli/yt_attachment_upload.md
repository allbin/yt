## yt attachment upload

Attach files to an issue

### Synopsis

Upload one or more local files as attachments on a YouTrack issue. Each
attachment is named after its file's base name; --name stores a single file
under another name. All files are opened before anything is sent, so a
missing file uploads nothing.

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

  # store /tmp/out.csv as report.csv
  yt attachment upload PROJ-123 /tmp/out.csv --name report.csv

  # JSON output (id, name and size of each created attachment)
  yt attachment upload PROJ-123 screenshot.png --json
```

### Options

```
  -h, --help          help for upload
      --name string   store the file under this name (one file only)
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt attachment](yt_attachment.md)	 - Manage issue attachments

