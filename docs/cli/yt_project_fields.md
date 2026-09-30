## yt project fields

List custom fields for a project

### Synopsis

List all custom fields configured on a YouTrack project, including
their types and allowed values.

Useful for discovering which fields can be set with --field or --subsystem
on issue create and update commands. A type ending in [] takes several values.

Use "yt project fields add" to add a value to a field.

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
* [yt project fields add](yt_project_fields_add.md)	 - Add allowed values to a project field

