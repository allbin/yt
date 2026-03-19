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
# Fetch an issue
yt issue PROJ-123

# JSON output
yt issue PROJ-123 --json

# List issues in a project
yt issue list -p PROJ

# Filter by state and assignee
yt issue list -p PROJ -s "In Progress" -a me

# Raw YouTrack query
yt issue list -q "tag: {Ready for QA} sort by: updated desc"

# Limit results
yt issue list -p PROJ -n 5 --json
```

## Setup

```bash
# Install fish/bash/zsh completions (auto-detects shell)
yt install completion
yt install completion --shell fish

# Install Claude Code skill
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
