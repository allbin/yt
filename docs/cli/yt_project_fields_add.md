## yt project fields add

Add allowed values to a project field

### Synopsis

Add values to the bundle behind a project's custom field, so they can be
set with --field. Works for enum, owned, version, build and state fields.

A value the field already has is reported and left unchanged. Bundles can be
shared by several projects: a value added here appears in every project that
uses the same bundle, which the output names. Requires permission to edit the
bundle.

```
yt project fields add <project> <field> <value>... [flags]
```

### Examples

```
  # add a customer to the Customer field
  yt project fields add HK Customer LTVB

  # add several values at once
  yt project fields add HK Subsystem "Admin UI" Billing
```

### Options

```
  -h, --help   help for add
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt project fields](yt_project_fields.md)	 - List custom fields for a project

