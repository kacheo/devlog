# AGENTS.md — For AI Agents Using devlog

This file helps AI agents (Claude Code, shell scripts, or any automated client) use devlog correctly.

---

## Purpose

devlog is a local-first CLI journal. It keeps daily work entries as plain markdown files and auto-imports git commits and GitHub PRs. The primary integration surface for agents is the `--json` flag — no MCP server or special SDK required.

**Agents: read this file first.** When you start work in a project with an `AGENTS.md` that references devlog, run `devlog show today --json` to understand the human's current work context before taking actions. After completing tasks, use `devlog add` to record what was done.

---

## Philosophy

- **Plain markdown files** you can open in any editor — devlog never hides your data.
- **`--json` for machines** — every output command emits structured JSON with a versioned schema.
- **No network required** for core use — GitHub access is optional and graceful.
- **Source of truth is frontmatter** — do not parse the rendered markdown body sections (Commits, PRs) to recover structured data; use the `--json` output instead.

---

## Usage Patterns

### Log work after completing a task

```bash
devlog add "Refactored auth middleware to use short-lived tokens"
devlog add "Fixed race condition in the job queue" --section blockers
devlog add "Shipped v2 API" --tag backend --tag api
```

### Read today's entries for context

```bash
devlog show today --json
# Returns:
# {
#   "version": "1",
#   "date": "2026-06-01",
#   "tags": ["auth", "backend"],
#   "sections": {
#     "notes":    ["Refactored auth middleware..."],
#     "commits":  [{ "sha": "abc1234", "message": "fix: oauth token refresh loop", "repo": "api-server" }],
#     "prs":      [{ "number": 142, "title": "Add rate limiter...", "state": "merged", "repo": "api-server" }],
#     "blockers": []
#   }
# }
```

### Review recent activity across a date range

```bash
devlog show --from yesterday --json
# Returns a JSON array — one object per day:
# [
#   {
#     "version": "1",
#     "date": "2026-06-01",
#     "tags": ["auth"],
#     "sections": {
#       "notes":    ["Refactored auth middleware..."],
#       "commits":  [{ "sha": "abc1234", "message": "fix: oauth token refresh loop", "repo": "api-server" }],
#       "prs":      [{ "number": 142, "title": "Add rate limiter...", "state": "merged", "repo": "api-server" }],
#       "blockers": []
#     }
#   }
# ]
```

### Sync git activity before reviewing recent work

```bash
devlog sync --quiet && devlog show --from yesterday --json
```

### Target a specific date

```bash
# Using flag (applies to add, edit, show, sync)
devlog show --date 2026-05-28 --json

# Using positional argument (show, edit only)
devlog show 2026-05-28
devlog edit yesterday
```

---

## JSON Schema

### `devlog show --json`

```json
{
  "version": "1",
  "date": "YYYY-MM-DD",
  "tags": ["tag1", "tag2"],
  "sections": {
    "notes":    ["string", "string"],
    "commits":  [{ "sha": "string", "message": "string", "repo": "string" }],
    "prs":      [{ "number": 142, "title": "string", "state": "merged|open|closed", "repo": "string" }],
    "blockers": ["string"]
  }
}
```

### `devlog show --from DATE --json` (date range)

Returns a JSON array where each element has the same shape as the single-day schema above:

```json
[
  { "version": "1", "date": "YYYY-MM-DD", "tags": [], "sections": { ... } },
  { "version": "1", "date": "YYYY-MM-DD", "tags": [], "sections": { ... } }
]
```

### Version field semantics

- All `--json` output includes `"version": "1"`.
- **Additive changes** (new optional fields) keep the same version.
- **Breaking changes** (removals, semantic changes to existing fields) increment the version.
- Consumers should ignore unknown fields and check `version` before parsing.

---

## Environment Variables for Agents

Set these to configure devlog without interactive setup or config files:

| Variable | Effect |
|----------|--------|
| `DEVLOG_DIR` | Journal directory (default: `~/devlog`) |
| `DEVLOG_EDITOR` | Editor to use for `devlog edit` |
| `DEVLOG_GITHUB_TOKEN` | GitHub API token for REST-based PR import |

```bash
export DEVLOG_DIR=~/my-journal
export DEVLOG_GITHUB_TOKEN=ghp_...
devlog sync
```

---

## Project Structure

```
devlog/
├── main.go
├── go.mod
├── cmd/                         # CLI commands (cobra)
│   ├── root.go                  # Root command, global flags (--json, --date)
│   ├── init.go                  # First-time setup, config creation, hook install
│   ├── add.go                   # Append bullet to a section
│   ├── edit.go                  # Open day file in $EDITOR
│   ├── show.go                  # Read and render day files
│   ├── sync.go                  # Import commits and PRs from git/GitHub
│   └── standup.go               # Compile and render standup view
└── internal/
    ├── config/config.go         # Load/write config.toml; env var overrides
    ├── store/store.go           # Open/create day files; path resolution; date parsing
    ├── store/entry.go           # DayEntry struct (frontmatter + sections)
    ├── git/scanner.go           # exec git log, parse commits
    ├── git/github.go            # gh CLI wrapper + GitHub REST client
    └── render/
        ├── markdown.go          # Serialize DayEntry to .md
        ├── terminal.go          # Human-readable colored output
        └── json.go              # --json output structs
```

**Import boundary:** External packages (agents, tests, other tools) must access devlog only through the `cmd/` surface. Do not import `internal/` packages directly.

---

## Adding a Command

devlog uses [cobra](https://github.com/spf13/cobra). To add a new command:

1. Create `cmd/<name>.go` with a `var <name>Cmd = &cobra.Command{ ... }`.
2. Implement `RunE: func(cmd *cobra.Command, args []string) error`.
3. Register in `root.go` or the appropriate parent command via `rootCmd.AddCommand(<name>Cmd)`.

Pattern from `cmd/add.go`:

```go
var addCmd = &cobra.Command{
    Use:   "add \"text\"",
    Short: "Append a bullet to today's journal",
    Args:  cobra.ExactArgs(1),
    RunE:  runAdd,
}

func init() {
    rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
    // implementation
}
```

---

## Testing

```bash
go test ./...
```

Unit tests are co-located with source files (`*_test.go`). Run all tests before submitting changes.

---

## Building

```bash
go build -o devlog .
./devlog --help
```

For cross-platform builds:

```bash
GOOS=darwin GOARCH=arm64 go build -o devlog-darwin-arm64 .
GOOS=linux GOARCH=amd64 go build -o devlog-linux-amd64 .
```

---

## Key Behaviors

- **Date parsing:** `today`, `yesterday`, `YYYY-MM-DD`. Unknown formats return a clear error.
- **Section normalization:** `--section blockers` maps to `## Blockers` regardless of case or singular/plural.
- **Write safety:** `devlog add` and `devlog sync` take per-file advisory locks and write atomically.
- **Idempotent sync:** Running `devlog sync` multiple times deduplicates by commit SHA and PR number.
- **Graceful degradation:** Missing `gh` CLI, no GitHub token, or repo without a GitHub remote is not an error — that source is skipped silently.
- **Version:** `devlog --version` prints the build version (e.g. `v0.1.0`; `dev` when built from source).