## yt branch

Create git branch from issue

### Synopsis

Create and switch to a new git branch named after a YouTrack issue.

Branch name format: <id>-<slugified-summary> (lowercase).
Use --no-slug for just the issue ID.

```
yt branch <id> [flags]
```

### Examples

```
  # creates branch like "proj-123-fix-login-bug"
  yt branch PROJ-123

  # creates branch "proj-123"
  yt branch PROJ-123 --no-slug
```

### Options

```
  -h, --help      help for branch
      --no-slug   omit summary slug from branch name
```

### Options inherited from parent commands

```
      --json   output raw JSON
```

### SEE ALSO

* [yt](yt.md)	 - YouTrack CLI

