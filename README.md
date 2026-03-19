# yt

CLI for JetBrains YouTrack.

## Install

```bash
go install github.com/allbin/yt@latest
```

## Configuration

Set environment variables:

```bash
export YOUTRACK_URL=https://youtrack.example.com
export YOUTRACK_TOKEN=perm:your-token-here
```

Or create `~/.config/yt/config.yaml`:

```yaml
url: https://youtrack.example.com
token: perm:your-token-here
```

## Usage

```bash
# Fetch an issue (or auto-detect from git branch)
yt issue PROJ-123
yt issue

# Create, update, comment
yt issue create -p PROJ -s "Fix login bug"
yt issue update PROJ-123 -s "In Progress" -a me
yt issue comment PROJ-123 -m "Started working on this"

# List and filter issues
yt issue list -p PROJ -s "In Progress" -a me
yt issue list -q "tag: {Ready for QA} sort by: updated desc"

# Agile boards
yt board list
yt board HållKoll -a me
yt board HållKoll -s "In Review" --json

# Git integration
yt branch PROJ-123         # creates proj-123-slugified-summary
yt branch PROJ-123 --no-slug

# Projects
yt project list

# Read comments
yt issue comments PROJ-123
```

All commands support `--json` for structured output.

## Setup

```bash
# Shell completions (auto-detects shell)
yt install completion

# Claude Code skill
yt install skill
```

## Development

Requires Go 1.26+ and [just](https://github.com/casey/just).

```bash
just build    # build binary
just test     # run tests
just lint     # golangci-lint
just check    # lint + test
just docs     # regenerate CLI reference docs
just install  # go install
```
