# devlog

**Local-first engineer journal — git commits and PRs auto-imported, standup-ready output.**

[![License](https://img.shields.io/github/license/kacheo/devlog)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/kacheo/devlog)](https://goreportcard.com/report/github.com/kacheo/devlog)
[![Release](https://img.shields.io/github/v/release/kacheo/devlog)](https://github.com/kacheo/devlog/releases)

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

```bash
go install github.com/kacheo/devlog@latest
```

The binary lands in `~/go/bin`. Make sure that's on your `$PATH`. Prebuilt binaries for macOS and Linux are on the [releases page](https://github.com/kacheo/devlog/releases).

---

## Quick Start

```bash
devlog init                                    # set up config and watched repos
devlog add "Implemented rate limiter middleware"  # log what you're working on
devlog show today                              # review your day
devlog standup                                 # standup-ready output
devlog sync                                    # import today's commits and PRs
```

---

## Commands

| Command | What it does |
|---------|-------------|
| `init [--non-interactive] [--add-repo PATH]` | First-time setup — config, repos, optional post-commit hook |
| `add "text" [--section SECTION] [--tag TAG]` | Append a bullet to today's notes or a named section |
| `edit [DATE]` | Open a day file in `$EDITOR` |
| `show [today\|yesterday\|YYYY-MM-DD\|week]` | Print entries for a day or the last 7 days |
| `sync [--quiet]` | Import today's commits and PRs from configured repos |
| `standup [--since DATE]` | Compile done items + blockers for standup |

**Global flags:** `--json` (structured output on `show`, `standup`, `sync`) · `--date YYYY-MM-DD` (target a specific date on `add`, `edit`, `show`, `sync`)

Unattended setup example:

```bash
devlog init --non-interactive \
  --add-repo ~/workspace/api-server \
  --add-repo ~/workspace/frontend
```

---

## Daily File Format

Each day is stored as `~/devlog/YYYY-MM-DD.md`:

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

`commits` and `prs` in frontmatter are the source of truth (written by `devlog sync`). `## Notes` and `## Blockers` are free prose (`devlog add` appends here). `## Commits` and `## PRs` body sections are rendered from frontmatter — do not edit them manually.

---

## GitHub Integration

`devlog sync` pulls PR activity via whichever method is available:

| Method | How | Requirement |
|--------|-----|-------------|
| `gh` CLI | `gh pr list --author=@me ...` | `gh auth login` |
| GitHub REST API | `/repos/:owner/:repo/pulls` | `DEVLOG_GITHUB_TOKEN` env var |
| No access | Commits still import; PRs skipped silently | — |

---

## Auto-Sync Hook

`devlog init` can install a `post-commit` hook in each watched repo:

```sh
#!/bin/sh
# >>> devlog managed block >>>
devlog sync --quiet
# <<< devlog managed block <<<
```

- Existing hooks are not overwritten — the managed block is appended.
- Multiple syncs on the same day file use advisory locks and write atomically — safe for parallel commits across repos.

---

## Configuration

Config at `~/.config/devlog/config.toml` (created by `devlog init`):

```toml
[journal]
dir    = "~/devlog"
editor = ""           # falls back to $EDITOR, then $VISUAL, then vi

[github]
token = ""            # optional; falls back to gh CLI auth

[[repos]]
path = "~/workspace/api-server"
name = "api-server"

[[repos]]
path = "~/workspace/frontend"
name = "frontend"
```

| Env var | Overrides |
|---------|-----------|
| `DEVLOG_DIR` | `journal.dir` |
| `DEVLOG_EDITOR` | `journal.editor` |
| `DEVLOG_GITHUB_TOKEN` | `github.token` |

---

## AI Agent Workflow

devlog works with any CLI-based agent — no MCP server needed. See [AGENT_INSTRUCTIONS.md](AGENT_INSTRUCTIONS.md) for the full JSON schema and integration patterns.

For the complete command reference, advanced filtering, workspace management, and troubleshooting, see the [User Guide](docs/user-guide.md).

```bash
devlog show today --json          # read today's context before starting work
devlog add "what was done"        # log completed tasks
devlog sync --quiet               # import git activity
devlog standup --json             # produce standup summary
```

---

## License

Apache 2.0 — see [LICENSE](LICENSE).
