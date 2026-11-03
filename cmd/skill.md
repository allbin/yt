---
name: yt
description: "Read and change YouTrack issues, boards, and sprints with the yt CLI. Use when an issue key appears (e.g. PROJ-123), or the user wants to search issues, create or update an issue, set custom fields, make a subtask, comment, link issues, put an issue on a board or sprint, check board status, see who moved which cards between states (e.g. for a standup), attach a file or download an attachment, add a value to a project field, or branch from an issue. Also for the knowledge base: when an article key appears (e.g. PROJ-A-12) or the user wants to read, search, write, or append to articles, wiki pages, or meeting notes."
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

`yt issue <ID>` includes the issue's links, board/sprint membership,
attachments, and the knowledge base articles that mention it, so it usually
answers a question in one call.

A full `--json` list carries every description, so beyond about 50 issues it
is too big to read in one go. Use `--json --compact` for a scan (one small
object per issue: id, summary, state, type, priority, assignee, tags,
updated), or write the full form to a file and query it with jq:

```bash
yt issue list -p AX -n 500 --json --compact
yt issue list -p AX -n 500 --json > /tmp/ax.json && jq '.[] | select(.fields.State=="Blocked") | .idReadable' /tmp/ax.json
```

Read custom fields from the `fields` map, not from `customFields`: `fields`
holds one value per field name, a string for a single-value field and a list
for a multi-value one (`{"State":"Planned","Subsystem":["Webticket"]}`), and
`null` when the field is empty. `customFields` is the raw API shape, whose
`value` is an object for State and an array for Subsystem, so a jq path
written for one breaks on the other.

Every `--assignee` accepts `me`, a login, or a full name. `-q/--query` takes raw
YouTrack search syntax.

## Boards and sprints

```bash
yt board list --json
yt board <board> --json       # current sprint; --sprint <name> for another
yt sprint list <board> --json
```

`<board>` (and every `--board`) is a name, matched case-insensitively, a board
id (`124-66`), or a board URL (`https://<host>/agiles/124-66/current`).
`yt board list` shows only boards shared with the user: a board private to its
owner is missing from it, so its name is not found, but its id or URL works.
When given a board link, pass the link.

A board's `kind` says what puts an issue on it, and `yt board <board> --json`
states it with a `membership` sentence:

- `sprints`: the issue is assigned to one of the board's sprints; a project or
  state alone is not enough. `yt board add` places it on the current sprint,
  or `--sprint <name>` on another.
- `query`: the issue matches the board's `query` (e.g. `Subsystem: Webticket`).
  Nothing is added or removed by hand; set the field the query names.
- `projects`: every issue of the board's projects is on it.
- `manual`: no sprints, cards added by hand with `yt board add`.

```bash
yt board add <board> <issue>... [--sprint <name>]      # default: current sprint
yt board remove <board> <issue>... [--sprint <name>]
```

Both are idempotent: `add` reports `already on board` and `remove` reports
`not on board` rather than failing. On a `query` or `projects` board, `add`
reports `on board by its query, nothing to add` for an issue the board shows
and errors for one it does not, naming the query to match. `yt issue <ID>
--json` lists every board the issue is on as `{board, kind, sprint}`, with
`sprint` only on a `sprints` board; it scans only the boards `yt board list`
shows.

## Activity history

To see which cards someone actually moved — for a standup or a weekly
summary — read the activity history, not the current assignee or `updated`
time:

```bash
yt transitions --json                                  # my State changes, last 7 days
yt transitions -u alice --since 2026-09-28 --until 2026-10-04 --json
yt transitions -u all --board AllTix --since 14d --json
```

Each entry has `issue`, `summary`, `author`, `time`, `from`, and `to`, oldest
first. `--user` defaults to `me`; `all` drops the user filter. `--since` and
`--until` take a date, an RFC 3339 time, or a duration like `7d`/`2w`; a
date-only `--until` includes that day. `--board` switches to the field the
board's columns are bound to, adds `fromColumn`/`toColumn`, and keeps only
issues on the board's sprints. `--field` lists another field's changes, and
`-p`/`-q` narrow the issues. An issue moved twice appears twice; group by
`issue` when summarising.

For everything that happened to one issue or article — who edited the
description, added a link, deleted a comment — read its history:

```bash
yt issue history <ID> --json                        # oldest first
yt issue history <ID> --kind field,comment --json   # only some kinds
yt article history <ID> --json
```

Each entry has `time`, `author`, `kind`, `field`, `removed`, and `added`.
`kind` is one of created, summary, description (content on an article),
field, link, tag, attachment, attachment-rename, comment, comment-edit,
comment-attachment, resolved, sprint, project, visibility, commit,
pull-request. `field` names the custom field, link type, or board, and for
comment kinds the comment ID. `--json` carries the whole text of every edit,
so a long description history is large: filter with `--kind`. YouTrack does
not record moving an article.

## Creating and updating issues

```bash
yt issue create --json -p PROJ -s "Summary" [-d "Description"]
yt issue update <ID> --json [-s state] [-a assignee] [--tag x] [--field "Name=Value"]
yt issue comment <ID> -m "text"
yt issue comment edit <ID> <commentID> -m "text"   # replaces the whole text
yt issue comment delete <ID> <commentID> --yes
```

Comment IDs are the `id` in `yt issue comments --json` and in the reply of
`yt issue comment --json`. Deleting a comment cannot be undone; without
`--yes` the command quotes the comment and changes nothing. Edit or delete
only comments you wrote in this session unless the user asks.

`yt projects --json` lists the project short names that `-p` expects.

Before setting `--subsystem`, `--state`, `--priority`, `--type`, or any
`--field "Name=Value"`, discover the valid values:

```bash
yt project fields PROJ --json
```

Field values are checked against that list before anything is written, and a
typo errors with the allowed values. Every value and tag goes through the REST
API, never YouTrack's command language, so a multi-word value needs only shell
quoting (`--field "Type=User Story"`, `--tag "To be discussed"`); braces are
not syntax here and are refused. On a multi-value field (`owned[]`, `enum[]`), repeat
`--field` with the same name to set several values — the list *replaces* the
current values, and so do `--subsystem` and the other shorthands. To keep
existing values, use `--add-field "Name=Value"` / `--remove-field` on `update`.
`--field "Name="` clears a field. Summary, description, state and fields are
saved in one request, so a rejected value changes nothing. On `create`, `-t`
is `--tag`; the type is `--type`.

A value missing from a field's bundle must be added before it can be set.
Bundles are often shared across projects and need admin rights, so ask before
adding:

```bash
yt project fields add PROJ Customer LTVB
```

`--tag` and `--remove-tag` match the whole tag name: the exact name first,
then a case-insensitive match if only one tag has it. An unknown name errors
with the closest existing tags; add `--create-tag` only when a new tag is
really meant. Tags are read back after writing, so a zero exit means they are
on the issue. `--remove-tag` of a tag the issue does not carry is a no-op.

Custom fields and existing tags go with the create request itself, so a bad
value creates nothing, not even a `--create-tag` tag. If `create` prints an
issue but exits non-zero, the issue exists and a follow-up step (a
`--create-tag` tag, links, board) failed; the error names each one. Finish it
with `yt issue update <ID>` or `yt link`; running `create` again makes a
duplicate.

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
board and sprint. Other links can be added on creation with repeatable
`--link relation=ID` (e.g. `--link depends-on=AX-12`). `--like <ID>` mirrors another issue's board and sprint without
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

`yt issue <ID> --json` lists attachments with name and size. Fetch one by name,
or attach local files:

```bash
yt attachment download <ID> <filename> [--output /tmp/file.csv]
yt attachment upload <ID> <file>... [--name stored.csv] --json
```

Upload answers with each attachment's id, name and size; `download` finds it
by that name. An uploaded image can be shown in a description or comment as
`![](name.png)`.

## Knowledge base articles

Articles are markdown pages in a project's knowledge base, nested in a tree.
Their IDs carry `-A-` (`PROJ-A-12`); `yt article` refuses an issue ID.

```bash
yt article list -p PROJ --json             # metadata only, no content; -q searches, -n limit
yt article list --parent PROJ-A-3 --json   # direct sub-articles
yt article <ID> --json                     # content, parent, childArticles, mentionedIssues, tags, visibility
yt article create -p PROJ -s "Title" -c @notes.md [--parent PROJ-A-3] [--tag x] --json
yt article update <ID> [-s title] [-c @file | --append @file] [--parent ID|none] [--tag x] [--remove-tag y] --json
yt article comment <ID> -m "text"            # also: comment edit / comment delete, as on issues
yt article comments <ID> --json
yt article history <ID> --json
yt article delete <ID> --yes
```

`-q` takes free text or `mentions: PROJ-123`. With `--parent`, `create` puts
the article in the parent's project, so `-p` is optional. `--content` replaces
the whole content; `--append` adds after it with a blank line between.
Append reads then writes, so a save by someone else in between is lost: give
concurrent note-takers their own sub-article. `--parent` on `update` moves
the article with its sub-articles; `none` moves it to the top level. Tags
match as on issues (`--create-tag` for a new one).

`visibility` is `{"limited": false}` for an article everyone in the project
can read; when `limited`, only the listed `users` and `groups` can.

An issue ID written in an article's content becomes a mention, which is the
only link between articles and issues. The article lists it under
`mentionedIssues`, and `yt issue <ID> --json` lists the article under
`articles`. To tie meeting notes to issues, write the IDs in the notes; to
find the notes behind an issue, read its `articles`.

Deleting cannot be undone: `delete` needs `--yes` and refuses an article that
has sub-articles. Ask the user before deleting an article you did not create
in this session.

Attachments on articles use the same commands as issues, with the article ID:
`yt attachment upload PROJ-A-12 photo.jpg`, `yt attachment download PROJ-A-12
photo.jpg`.

When presenting an article, lead with ID and title, then parent and project,
then the content; list mentioned issues with their summaries when they
matter to the question.

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
