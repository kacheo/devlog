# Contributing to devlog

Thank you for your interest in contributing to devlog.

## Development Setup

**Prerequisites:**
- Go 1.22+ (`go version`)
- [golangci-lint](https://golangci-lint.run/usage/install/) v2.x (`golangci-lint --version`)
- Optional: [gh CLI](https://cli.github.com/) for PR integration testing

**Clone and build:**

```bash
git clone https://github.com/kacheo/devlog
cd devlog
go build -o devlog .
./devlog --help
```

## Before Submitting a PR

Run all checks locally — CI will enforce them:

```bash
make lint    # golangci-lint run ./...
make test    # go test -race -covermode=atomic -coverprofile=coverage.out ./...
```

Coverage must stay at or above **70%** overall. The CI job will fail below this threshold.

## Package Boundaries

- `cmd/` — public command surface; add new commands here
- `internal/` — implementation details; never import from outside this repo
- `cmd/` may import from `internal/`; `internal/` packages must not import `cmd/`

## Commit Messages

Use conventional commit prefixes:

```
feat:     a new feature
fix:      a bug fix
docs:     documentation only
test:     adding or updating tests
refactor: code change that neither fixes a bug nor adds a feature
chore:    build, CI, or tooling changes
```

These prefixes are used to generate the release changelog. `docs:`, `test:`, and `chore:`
entries are excluded from the published changelog.

## Adding a New Command

1. Create `cmd/mycommand.go` with a Cobra command registered in `init()`
2. Add the command to the root in `cmd/root.go`
3. Add `cmd/mycommand_test.go` with coverage for the happy path and error cases
4. Update the command table in `README.md`

## Running a Subset of Tests

```bash
go test ./cmd/... -run TestShow    # run TestShow* in cmd/
go test ./internal/git/...         # run all tests in internal/git
```

## Release Process

Releases are automated via [goreleaser](https://goreleaser.com/). Maintainers tag a release:

```bash
make release VERSION=v0.2.0
```

This creates the git tag and pushes it, which triggers the release GitHub Actions workflow.
