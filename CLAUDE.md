# Claude Code Entry Point for devlog

devlog is a local-first CLI journal that keeps daily work entries as plain markdown
and auto-imports git commits and GitHub PRs via `devlog sync`.

## Read First

- **Workflow and agent patterns**: [AGENT_INSTRUCTIONS.md](AGENT_INSTRUCTIONS.md)

## Current Ground Rules

- Run `devlog show today --json` before starting work in any devlog-tracked project.
- Run `go test ./...` before submitting any change; no exceptions.
- Run `make lint` (`golangci-lint run ./...`) before submitting; fix findings, don't suppress them.
- Build with `go build -o devlog .` and smoke-test with `./devlog --help`.
- `cmd/` is the public surface — do not import `internal/` packages from outside this repo.
- If this file conflicts with [AGENT_INSTRUCTIONS.md](AGENT_INSTRUCTIONS.md), trust
  AGENT_INSTRUCTIONS.md and remove the duplicate here.
