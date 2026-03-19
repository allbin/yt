# yt

YouTrack CLI. Module: `github.com/allbin/yt`. Go 1.26+.

## Commands

```
just build    # go build -o yt .
just test     # go test ./...
just lint     # golangci-lint run
just check    # lint + test
just docs     # regenerate docs/cli/
just install  # go install
```

## Env

- `YOUTRACK_URL` -- base URL
- `YOUTRACK_TOKEN` -- permanent token
- Also reads `~/.config/yt/config.yaml` via viper

## Structure

- `cmd/` -- cobra commands (root, issue, board, project, branch, install)
- `internal/youtrack/` -- API interface + HTTP client + types
- `internal/format/` -- text/JSON formatters
- `internal/git/` -- git helpers (branch naming)
- `internal/tools/` -- docgen

## Patterns

- `youtrack.API` interface is the abstraction; `newClient()` in `cmd/root.go` returns it
- errcheck is strict -- handle all errors, use `errWriter` pattern in formatters
- All cobra commands need `Long` + `Example` fields (used for LLM skill docs)

## After Changes

1. `just check` -- must pass
2. `just build` -- verify binary works
3. `just docs` -- regenerate CLI reference
4. `./yt install skill` -- reinstall Claude Code skill
