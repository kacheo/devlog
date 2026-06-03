# Agent Instructions

See [AGENT_INSTRUCTIONS.md](AGENT_INSTRUCTIONS.md) for full instructions.

This file exists for compatibility with tools that look for AGENTS.md.

## Quick Start

```bash
devlog show today --json   # read current work context before taking action
devlog add "what was done" # log completed work
devlog sync --quiet        # import git commits and PRs
```

## Key Sections in AGENT_INSTRUCTIONS.md

- **Usage Patterns** — JSON output, add/show/standup/sync commands
- **JSON Schema** — versioned `show --json` and `standup --json` output format
- **Environment Variables** — `DEVLOG_DIR`, `DEVLOG_GITHUB_TOKEN`, `DEVLOG_EDITOR`
- **Project Structure** — `cmd/` (public surface), `internal/` (private), import boundary
- **Adding a Command** — cobra pattern with `RunE` and `rootCmd.AddCommand`
- **Testing / Building** — `go test ./...`, `go build -o devlog .`
- **Key Behaviors** — date parsing, write safety, idempotent sync, graceful degradation
