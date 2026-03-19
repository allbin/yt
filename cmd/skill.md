---
name: yt
description: "Interact with YouTrack: fetch issue details, list/filter issues, view agile board sprints. Use when referencing issue keys (e.g. PROJ-123), searching issues, or checking board status."
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

## Presenting results

For a single issue:
1. Issue ID & Summary as heading
2. State, Priority, Assignee, Type as metadata
3. Subsystem and Tags if present
4. Description if available

For lists: compact table with ID, state, priority, assignee, summary.

If a command fails, report the error clearly.
