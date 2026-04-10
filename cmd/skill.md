---
name: yt
description: "Interact with YouTrack: fetch issue details, list/filter issues, view boards, create branches, download attachments. Use when referencing issue keys (e.g. PROJ-123), searching issues, or checking board status."
argument-hint: <issue-id>
allowed-tools: Bash(yt *)
---

# YouTrack CLI

Use the `yt` CLI to interact with YouTrack. Always pass `--json` for structured output.

## Fetch a single issue

If `$ARGUMENTS` looks like an issue ID (e.g. PROJ-123):

```bash
yt issue $ARGUMENTS --json
```

With no ID, auto-detects from the current git branch name:

```bash
yt issue --json
```

## List and filter issues

```bash
yt issue list --json [flags]
```

Flags:
- `-p, --project` — filter by project
- `-s, --state` — filter by state
- `-a, --assignee` — filter by assignee (supports "me", login, or full name)
- `-q, --query` — raw YouTrack search query
- `-n, --limit` — max results (default 20)

## Agile boards

List available boards:

```bash
yt board list --json
```

Show issues on a board's current sprint:

```bash
yt board <name> --json [-s state] [-a assignee] [-q query] [--sprint name]
```

Board name matching is case-insensitive. Assignee supports "me", login, or full name.

## Git integration

Create a branch from an issue:

```bash
yt branch <ID>           # e.g. proj-123-slugified-summary
yt branch <ID> --no-slug # e.g. proj-123
```

## Create an issue

```bash
yt issue create --json -p PROJ -s "Summary" [-d "Description"] [-t tag1 -t tag2]
```

Flags:
- `-p, --project` — project short name (required)
- `-s, --summary` — issue summary (required)
- `-d, --description` — issue description
- `-t, --tag` — add tag (repeatable)

Tags are created automatically by YouTrack if they don't exist.

## Update an issue

```bash
yt issue update PROJ-123 [flags]
```

Flags:
- `-s, --state` — set issue state
- `-a, --assignee` — set assignee (supports "me", login, or full name)
- `-p, --priority` — set priority
- `-t, --type` — set issue type
- `--tag` — add tag (repeatable)
- `--remove-tag` — remove tag (repeatable)

Multiple flags can be combined. Tags are added/removed without affecting existing tags.

## Attachments

Issue details (`yt issue <ID> --json`) include an `attachments` array with name and size.

Download an attachment:

```bash
yt attachment download <ID> <filename>
yt attachment download <ID> <filename> --output /tmp/file.csv
```

## Projects

```bash
yt project list --json
```

## Presenting results

For a single issue:
1. Issue ID & Summary as heading
2. State, Priority, Assignee, Type as metadata
3. Subsystem and Tags if present
4. Description if available
5. Attachments if present (offer to download when relevant)

For lists: compact table with ID, state, priority, assignee, summary.

If a command fails, report the error clearly.
