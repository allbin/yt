## yt attachment download

Download an attachment from an issue

### Synopsis

Download an attachment by filename from a YouTrack issue.
If multiple attachments share the same name, the first match is downloaded.

```
yt attachment download <issueID> <filename> [flags]
```

### Examples

```
  # download to current directory
  yt attachment download PROJ-123 report.csv

  # download to specific path
  yt attachment download PROJ-123 photo.png --output /tmp/photo.png
```

### Options

```
  -h, --help            help for download
  -o, --output string   output file path (default: filename in current directory)
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt attachment](yt_attachment.md)	 - Manage issue attachments

