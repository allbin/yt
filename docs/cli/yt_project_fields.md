## yt project fields

List custom fields for a project

### Synopsis

List all custom fields configured on a YouTrack project, including
their types and allowed values.

Useful for discovering which fields can be set with --field or --subsystem
on issue create and update commands.

```
yt project fields <project> [flags]
```

### Examples

```
  # list fields for a project
  yt project fields PROJ

  # output as JSON
  yt project fields PROJ --json
```

### Options

```
  -h, --help   help for fields
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt project](yt_project.md)	 - Inspect YouTrack project details

