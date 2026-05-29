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
yt issue create --json -p PROJ -s "Summary" [-d "Description"] [-t tag1 -t tag2] [--subsystem API] [--field "Name=Value"]
```

Flags:
- `-p, --project` — project short name (required)
- `-s, --summary` — issue summary (required)
- `-d, --description` — issue description
- `-t, --tag` — add tag (repeatable)
- `--subsystem` — set subsystem
- `--field` — set custom field as "Name=Value" (repeatable)

Tags are created automatically by YouTrack if they don't exist.

## Update an issue

```bash
yt issue update PROJ-123 [flags]
```

Flags:
- `-S, --summary` — set issue summary (uses REST API)
- `-d, --description` — set issue description (uses REST API)
- `-s, --state` — set issue state (uses REST API)
- `-a, --assignee` — set assignee (supports "me", login, or full name)
- `-p, --priority` — set priority
- `-t, --type` — set issue type
- `--subsystem` — set subsystem
- `--tag` — add tag (repeatable)
- `--remove-tag` — remove tag (repeatable)
- `--field` — set custom field as "Name=Value" (repeatable)

Multiple flags can be combined. Summary, description, and state use the REST API; other fields use the command API. Both can be used in a single invocation.

## Links between issues

Link types are instance-specific and directed. Discover them with:

```bash
yt link types --json
```

Each type has an outward and (if directed) inward phrase, e.g. Subtask is
`parent for` (outward) / `subtask of` (inward). "Make A a subtask of B" and
"make A parent for B" are the same type in opposite directions.

Create link(s) — the relation accepts kebab, spaced, or squashed forms:

```bash
yt link <ID> <relation> <target-ID>... [--json]

yt link AX-804 subtask-of AX-332     # AX-804 becomes a subtask of AX-332
yt link AX-1 relates AX-2            # symmetric
yt link AX-1 depends-on AX-3
yt link AX-1 duplicates AX-4
yt link AX-1 relates AX-2 AX-3       # multiple targets in one call
```

Linking is idempotent: an existing link is reported as `(already linked)` and
left unchanged. Unknown relations error and print the valid relations.

Remove a link:

```bash
yt unlink <ID> <relation> <target-ID> [--json]

yt unlink AX-804 subtask-of AX-332
```

List an issue's links, grouped by relation:

```bash
yt links <ID> [--json]
```

The same links also appear in `yt issue <ID>` output (text and `--json`).

For `link`/`unlink`, `--json` returns the source issue's links after the change;
for `links` it returns the links array; for `link types` it returns the types.

## Attachments

Issue details (`yt issue <ID> --json`) include an `attachments` array with name and size.

Download an attachment:

```bash
yt attachment download <ID> <filename>
yt attachment download <ID> <filename> --output /tmp/file.csv
```

## Projects

List all projects:

```bash
yt projects --json
```

## Project custom fields

Before setting custom fields with `--field` or `--subsystem`, discover what
fields are available and their allowed values:

```bash
yt project fields PROJ --json
```

Returns an array of fields with name, type, and allowed values. Use this to
determine valid values for `--subsystem`, `--state`, `--priority`, `--type`,
or any `--field "Name=Value"` flag.

## Presenting results

For a single issue:
1. Issue ID & Summary as heading
2. State, Priority, Assignee, Type as metadata
3. Subsystem and Tags if present
4. Links if present (relation + target IDs)
5. Description if available
6. Attachments if present (offer to download when relevant)

For lists: compact table with ID, state, priority, assignee, summary.

If a command fails, report the error clearly.
