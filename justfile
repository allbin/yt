default:
    @just --list

build:
    go build -o yt .

test:
    go test ./...

lint:
    golangci-lint run

check: lint test

docs:
    go run ./internal/tools/docgen -out ./docs/cli

install:
    go install .
