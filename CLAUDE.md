# CLAUDE.md — For Claude Code working on devlog

This file instructs Claude Code on how to use devlog while working on this codebase.

---

## Working on this project

### Read AGENTS.md first

Run `devlog show today --json` to understand the current development context before starting work.

### Log your work

After completing any task, log it:

```bash
devlog add "description of what was done"
devlog add "blocker description" --section blockers
```

### Generate standup

```bash
devlog standup
```

### Sync git activity

```bash
devlog sync
```

### Build and test

```bash
go build -o devlog .
go test ./...
```

---

## Project conventions

- Commands live in `cmd/`, libraries in `internal/`
- Add tests co-located with source (`*_test.go`)
- Import boundary: external packages (agents, tests) access devlog through `cmd/` only — do not import `internal/` directly
- Run `go test ./...` and verify build (`go build -o devlog .`) before committing