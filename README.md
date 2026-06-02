# devlog

**Local-first engineer journal — git commits and PRs auto-imported, standup-ready output.**

```
$ devlog add "Fixed the OAuth token refresh loop"
$ devlog standup
--- Standup: Monday, June 1 ---

Done (since 2026-06-01):
  • fix: oauth token refresh loop  (api-server · abc1234)
  • feat: add rate limiter to /v2/users  (api-server · def5678)
  • PR #142 merged: Add rate limiter to user endpoints  (api-server)

Blockers:
  • (none)
```

devlog keeps a structured daily journal as plain markdown files you can always open in your editor. Every command supports `--json` for AI agent integration — no MCP server required.

---

## Installation

### Go

```bash
go install github.com/kacheo/devlog@latest
```

The binary is placed in `~/go/bin` (default; run `go env GOPATH` if unsure). Make sure that directory is on your PATH:

```bash
# Add to ~/.zshrc (macOS/zsh) or ~/.bashrc (Linux/bash)
export PATH="$PATH:$HOME/go/bin"
```

Then reload your shell (`source ~/.zshrc`) or open a new terminal.

### Homebrew

Not yet available. Use `go install` or download a binary below.

### Download binary

Prebuilt binaries for macOS (arm64, amd64) and Linux (amd64) are available on the [releases page](https://github.com/kacheo/devlog/releases).

After downloading, extract and move the binary to a directory on your PATH:

```bash
tar xzf devlog_*_*.tar.gz

# macOS / Linux (system-wide, requires sudo)
sudo mv devlog /usr/local/bin/

# Or into a user-local bin dir (no sudo needed)
mkdir -p ~/.local/bin && mv devlog ~/.local/bin/
```

If using `~/.local/bin/`, add it to your shell config if it isn't already:

```bash
export PATH="$PATH:$HOME/.local/bin"
```

---

## Quick Start

```bash
# 1. Set up config and watched repos
devlog init

# 2. Log what you're working on
devlog add "Implemented rate limiter middleware"

# 3. Review your day
devlog show today

# 4. Get standup-ready output
devlog standup

# 5. Import today's git commits and PRs
devlog sync
```

---

## Commands

### `devlog init [--non-interactive]`

First-time setup. Creates `~/.config/devlog/config.toml`, registers repos, and offers to install a `post-commit` git hook in each watched repo.

| Flag | Description |
|------|-------------|
| `--add-repo <path>` | Append a repo to config without interactive prompts (repeatable) |
| `--non-interactive` | Use defaults / flags only — no TTY prompts (for CI/agents) |

```bash
# Interactive setup
devlog init

# Unattended setup for a new machine
devlog init --non-interactive \
  --add-repo ~/workspace/api-server \
  --add-repo ~/workspace/frontend
```

---

### `devlog add "text" [--section SECTION] [--tag TAG ...]`

Append a bullet to today's journal file. Default section is `notes`.

| Flag | Description |
|------|-------------|
| `--section` | Target section: `notes` (default), `blockers`, or others |
| `--tag` | Add a tag to today's frontmatter (repeatable) |

```bash
# Add to notes (default)
devlog add "Reviewed Alice's PR on rate limiter"

# Add a blocker
devlog add "Waiting on DevOps to provision staging DB" --section blockers

# Tag the entry
devlog add "Shipped v2 API" --tag backend --tag api
```

---

### `devlog edit [DATE]`

Open a day's file in `$EDITOR`. Default date is today.

```bash
devlog edit 2026-05-28
devlog edit yesterday
devlog edit today
```

---

### `devlog show [today|yesterday|YYYY-MM-DD|week]`

Print entries for a day or range. Use `--date` flag for the same effect.

| Argument | Description |
|----------|-------------|
| `today` | Today (default) |
| `yesterday` | Previous calendar day |
| `YYYY-MM-DD` | Specific date |
| `week` | Last 7 days |

```bash
devlog show today
devlog show yesterday
devlog show 2026-05-28
devlog show week
```

---

### `devlog sync [--quiet]`

Import today's commits and PRs from configured repos. Safe to run multiple times — deduplicates by SHA/number.

| Flag | Description |
|------|-------------|
| `--quiet` | Suppress all output — used by the post-commit hook; also suppresses `--json` output |

```bash
devlog sync
```

---

### `devlog standup [--since DATE|yesterday]`

Compile a standup view: yesterday's commits + PRs as "done", today's blockers.

| Flag | Description |
|------|-------------|
| `--since` | Start date for done items (default: yesterday) |

```bash
devlog standup
devlog standup --since 2026-05-28
```

---

## Global Flags

| Flag | Applies to | Description |
|------|------------|-------------|
| `--json` | `show`, `standup`, `sync` | Emit structured JSON instead of human-readable output |
| `--date YYYY-MM-DD` | `add`, `edit`, `show`, `sync` | Target a specific date instead of today |

> **Note:** If a command also takes a positional date (e.g. `show yesterday`), `--date` and the positional are mutually exclusive — passing both is an error.

---

## Daily File Format

Each day is stored as `~/devlog/YYYY-MM-DD.md`. New files are scaffolded with all sections present.

```markdown
---
date: 2026-06-01
tags: [auth, backend]
commits:
  - { sha: abc1234, message: "fix: oauth token refresh loop", repo: api-server }
  - { sha: def5678, message: "feat: add rate limiter to /v2/users", repo: api-server }
prs:
  - { number: 142, title: "Add rate limiter to user endpoints", state: merged, repo: api-server }
---

## Notes
- Fixed the OAuth token refresh loop
- Reviewed Alice's PR on the API rate limiter

## Commits
- abc1234 fix: oauth token refresh loop (api-server)
- def5678 feat: add rate limiter to /v2/users (api-server)

## PRs
- #142 Add rate limiter to user endpoints [merged] (api-server)

## Blockers
- Waiting on DevOps to provision staging DB
```

### Source of truth

- **`commits` and `prs` in frontmatter** are the source of truth — populated by `devlog sync`.
- **`## Notes` and `## Blockers`** are free prose — `devlog add` appends to them.
- **`## Commits` and `## PRs`** body sections are rendered from frontmatter — regenerated by `devlog sync`, do not edit manually.

---

## GitHub Integration

`devlog sync` pulls activity from configured repos:

| Method | How | Requirement |
|--------|-----|-------------|
| `gh` CLI | `gh pr list --repo <owner/repo> --author=@me --state=all --json ...` | `gh auth login` |
| GitHub REST API | `/repos/:owner/:repo/pulls` with `DEVLOG_GITHUB_TOKEN` | `DEVLOG_GITHUB_TOKEN` env var |
| No access | Commits still import; PRs skipped silently | — |

Set the token for CI/agent contexts:

```bash
export DEVLOG_GITHUB_TOKEN=ghp_...
devlog sync
```

---

## Auto-Sync Hook

After each `git commit`, devlog can sync that day's commits and PRs to your journal. During `devlog init`, you can opt to install a `post-commit` hook in each watched repo:

```sh
#!/bin/sh
# >>> devlog managed block >>>
devlog sync --quiet
# <<< devlog managed block <<<
```

- **Existing hooks** are not overwritten — the managed block is appended, and on re-install only the block content is replaced.
- **Non-shell hooks** or unwritable files: a warning is printed, sync is skipped.
- **Concurrency**: multiple syncs writing the same day file take advisory locks and write atomically (temp file + rename) — safe for parallel commits across repos.

To install manually:

```bash
# In each watched repo
echo '#!/bin/sh
devlog sync --quiet' > .git/hooks/post-commit
chmod +x .git/hooks/post-commit
```

---

## JSON Output

All output commands (`show`, `standup`, `sync`) support `--json` with a stable, versioned schema.

```bash
devlog show today --json
devlog standup --json
```

Schema version policy: **additive, backward-compatible changes keep the same version**; breaking changes increment it. Consumers should check `version` to detect breaking changes and ignore unknown fields.

---

## Configuration

Config file: `~/.config/devlog/config.toml` (created by `devlog init`). Journal files go in `~/devlog/` by default.

```toml
[journal]
dir    = "~/devlog"      # where day files are stored
editor = ""               # falls back to $EDITOR, then $VISUAL, then vi

[github]
token = ""               # optional; falls back to gh CLI auth

[[repos]]
path = "~/workspace/api-server"
name = "api-server"
# github_slug = "acme/api-server"   # optional; override owner/repo from git remote

[[repos]]
path = "~/workspace/frontend"
name = "frontend"
```

### Environment Variable Overrides

| Variable | Overrides | Use case |
|----------|----------|----------|
| `DEVLOG_DIR` | `journal.dir` | CI/agent contexts |
| `DEVLOG_EDITOR` | `journal.editor` | Force a specific editor |
| `DEVLOG_GITHUB_TOKEN` | `github.token` | Pass token without writing to disk |

---

## AI Agent Workflow

devlog is designed for CLI-based agents (Claude Code, shell scripts) with no special integration:

```bash
# After completing work, log what was done
devlog add "Refactored auth middleware to use short-lived tokens"

# Read today's entries to understand context before starting new work
devlog show today --json | jq '.sections.notes'

# Generate standup summary for the human
devlog standup --json

# Sync git activity before standup
devlog sync && devlog standup
```

---

## Using with Claude Code

devlog works out-of-the-box with Claude Code and other CLI agents. Add this to your project's `CLAUDE.md` to give your agent instructions:

```markdown
## Daily Journal

Use devlog to track your work. After completing each task, log it:

- `devlog add "description of what was done"` — appends to today's notes
- `devlog add "blocker description" --section blockers` — logs an impediment
- `devlog show today --json` — read today's entries to understand current context
- `devlog standup` — generate a standup summary for review
- `devlog sync --quiet && devlog standup` — sync git activity then produce standup

The journal lives in `~/devlog/YYYY-MM-DD.md` as plain markdown — you can also open it directly in your editor.
```

### Example prompts for your agent

Copy any of these into your conversation to direct Claude Code:

**Log work after a task:**
> After you finish the refactor, run `devlog add "Refactored auth middleware to use short-lived tokens"` so I can see what was done in my standup.

**Check context before starting:**
> Before you start work today, run `devlog show today --json` and tell me what the human was working on yesterday and what blockers exist.

**Get a standup summary:**
> Generate a standup summary for me with `devlog standup --json`, then print a human-readable version.

**Sync and report:**
> Sync today's git commits and PRs, then give me a summary of what was done: `devlog sync && devlog standup`.

**Block a task:**
> I can't proceed with the database migration until DevOps provisions the staging instance. Log this as a blocker: `devlog add "Waiting on DevOps to provision staging DB" --section blockers`.

---

## License

Apache 2.0 — see [LICENSE](LICENSE).