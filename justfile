# short git sha, plus -dirty when the working tree has uncommitted changes
version := `git rev-parse --short HEAD 2>/dev/null || echo unknown` + `test -z "$(git status --porcelain 2>/dev/null)" || echo -dirty`
ldflags := "-X github.com/allbin/yt/internal/version.version=" + version

default:
    @just --list

build:
    go build -ldflags "{{ldflags}}" -o yt .

test:
    go test ./...

lint:
    golangci-lint run

check: lint test

docs:
    go run ./internal/tools/docgen -out ./docs/cli

install:
    go install -ldflags "{{ldflags}}" .
    yt install skill
    yt install completion
