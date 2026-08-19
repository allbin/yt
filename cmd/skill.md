---
name: yt
description: "Read and change YouTrack issues, boards, and sprints with the yt CLI. Use when an issue key appears (e.g. PROJ-123), or the user wants to search issues, create or update an issue, make a subtask, comment, link issues, put an issue on a board or sprint, check board status, download an attachment, or branch from an issue."
allowed-tools: Bash(yt *)
---

# YouTrack CLI

Pass `--json` to every `yt` command. It is the machine-readable form, and it
guarantees the command prints data rather than opening an interactive viewer.

`yt <command> --help` lists flags and arguments. This document covers what
`--help` cannot tell you: the conventions, the gotchas, and the orderings.

## Authentication

A "url not configured" or "token not configured" error means the user must
authenticate. Have them run `yt login`, which prompts for URL and token,
validates the token before saving, and writes `~/.config/yt/config.yaml`.
`YOUTRACK_URL` / `YOUTRACK_TOKEN` override the saved config at runtime.

## Reading issues

```bash
yt issue <ID> --json          # omit <ID> to detect it from the git branch name
yt issue list --json          # -p project, -s state, -a assignee, -q query, -n limit
yt issue comments <ID> --json
```

`yt issue <ID>` includes the issue's links, board/sprint membership, and
attachments, so it usually answers a question in one call.

Every `--assignee` accepts `me`, a login, or a full name. `-q/--query` takes raw
YouTrack search syntax.

## Boards and sprints

```bash
yt board list --json
yt board <name> --json        # current sprint; --sprint <name> for another
yt sprint list <board> --json
```

Board names match case-insensitively.

An issue appears on a sprint-based board only if it is assigned to one of that
board's sprints — setting a project or state is not enough. Manage membership
explicitly:

```bash
yt board add <board> <issue>... [--sprint <name>]      # default: current sprint
yt board remove <board> <issue>... [--sprint <name>]
```

Both are idempotent: `add` reports `(already on board)` and `remove` reports
`(not on board)` rather than failing.

## Creating and updating issues

```bash
yt issue create --json -p PROJ -s "Summary" [-d "Description"]
yt issue update <ID> --json [-s state] [-a assignee] [--tag x] [--field "Name=Value"]
yt issue comment <ID> -m "text"
```

`yt projects --json` lists the project short names that `-p` expects.

Before setting `--subsystem`, `--state`, `--priority`, `--type`, or any
`--field "Name=Value"`, discover the valid values:

```bash
yt project fields PROJ --json
```

Tags are the exception — YouTrack creates them on demand, so `--tag` accepts a
tag that does not exist yet.

For long or multi-line text, `-d/--description` and `-m/--message` accept `@path`
to read a file and `-` to read stdin, which avoids shell mangling:

```bash
yt issue create -p PROJ -s "Writeup" -d @notes.md
git log -1 --format=%B | yt issue comment PROJ-123 -m -
```

To create a subtask that lands on its parent's board in one call:

```bash
yt issue create -p AX -s "Subtask summary" --parent AX-332 --json
```

`--parent` adds the `subtask of` link *and* places the issue on the parent's
board and sprint. `--like <ID>` mirrors another issue's board and sprint without
linking. Both are overridden by an explicit `--board`/`--sprint`.

`yt issue state` and `yt issue view` open an interactive picker and viewer for a
human. To change a state non-interactively, use `yt issue update -s`.

## Linking issues

Link types are instance-specific and directed, so discover them before linking:

```bash
yt link types --json
```

Each directed type has an outward and an inward phrase — Subtask is `parent for`
outward and `subtask of` inward, which makes "A is a subtask of B" and "B is
parent for A" the same link stated from either end.

```bash
yt link <ID> <relation> <target-ID>... [--json]
yt unlink <ID> <relation> <target-ID> [--json]
yt links <ID> --json
```

The relation accepts kebab, spaced, or squashed forms (`subtask-of`,
`"subtask of"`, `subtaskof`). Linking is idempotent — an existing link reports
`(already linked)`. An unknown relation errors and prints the valid ones.

## Attachments

`yt issue <ID> --json` lists attachments with name and size. Fetch one by name:

```bash
yt attachment download <ID> <filename> [--output /tmp/file.csv]
```

## Git branches

```bash
yt branch <ID>              # proj-123-slugified-summary
yt branch <ID> --no-slug    # proj-123
```

Branch names are what `yt issue` reads back when no ID is given.

## Presenting results

For one issue, lead with ID and summary as a heading, then state, priority,
assignee, and type; then subsystem, tags, links, and board membership where they
exist; then the description; then attachments, offering to download when they
look relevant.

For several issues, use a compact table of ID, state, priority, assignee, and
summary.
