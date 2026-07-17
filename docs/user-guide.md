# devlog User Guide

This guide covers the full command set, advanced filtering, workspace management, editing entries, and AI agent integration patterns. The [README](../README.md) is the quick-start reference; this guide fills in everything else.

---

## Table of Contents

- [Full Command Reference](#full-command-reference)
- [Daily File Format](#daily-file-format)
- [Workspace Management](#workspace-management)
- [Filtering and Searching](#filtering-and-searching)
- [Editing Entries](#editing-entries)
- [AI Agent Integration](#ai-agent-integration)
- [Troubleshooting](#troubleshooting)

**Commands:** `init` · `add` · `edit` · `show` · `sync` · `search` · `rm` · `update` · `resolve` · `reopen` · `items` · `workspace` · `tags`


---

## Full Command Reference

### Global Flags

These flags work on every command:

| Flag | Description |
|------|-------------|
| `--json` | Emit structured JSON output (all output commands) |
| `--date YYYY-MM-DD` | Target a specific date instead of today |

---

### `init`

First-time setup: create config and optionally install git hooks.

```
devlog init [flags]
```

| Flag | Description |
|------|-------------|
| `--non-interactive` | Use defaults without prompting |
| `--add-repo <path>` | Register a repo path (repeatable) |

`devlog init` creates `~/.config/devlog/config.toml`, prompts for repo paths to watch, and offers to install a `post-commit` hook in each. Existing hooks are not overwritten — the managed block is appended.

```bash
# Interactive setup
devlog init

# Unattended setup
devlog init --non-interactive \
  --add-repo ~/workspace/api-server \
  --add-repo ~/workspace/frontend
```

---

### `add`

Append a bullet to a section of today's journal.

```
devlog add "text" [flags]
```

| Flag | Description |
|------|-------------|
| `--section <name>` | Section to append to: `notes` (default), `blockers`, `action_items` |
| `--tag <tag>` | Add a tag to today's frontmatter (repeatable); must be lowercase letters, digits, and underscores only |

Tags must match the pattern `[a-z0-9_]+` — lowercase letters, digits, and underscores. Examples: `auth`, `api_v2`, `oauth2`.

```bash
devlog add "Shipped the rate limiter PR"
devlog add "Blocked on staging DB provisioning" --section blockers
devlog add "Write integration tests for /v2/users" --section action_items
devlog add "Deployed new auth flow" --tag backend --tag auth
devlog add "API work" --tag api_v2          # underscores ok
devlog add "API work" --tag "Auth"          # error: uppercase not allowed
```

---

### `edit`

Open a day file in `$EDITOR`.

```
devlog edit [DATE]
```

`DATE` accepts `today`, `yesterday`, or `YYYY-MM-DD`. Defaults to today.

```bash
devlog edit
devlog edit yesterday
devlog edit 2026-05-28
```

---

### `show`

Print journal entries for a day or date range.

```
devlog show [today|yesterday|YYYY-MM-DD|week] [flags]
```

| Flag | Description |
|------|-------------|
| `--from <date>` | Start of date range (inclusive) |
| `--until <date>` | End of date range (inclusive), defaults to today when `--from` is set |
| `--section <name>` | Render only these sections: `notes`, `blockers`, `action_items` (repeatable or comma-separated) |
| `--tag <tag>` | Include only entries that have this tag |
| `--repo <name>` | Include only entries with commits or PRs from this repo |
| `--json` | Emit structured JSON |

```bash
devlog show today
devlog show week
devlog show 2026-05-28

# Date range (--until defaults to today)
devlog show --from yesterday
devlog show --from 2026-05-28 --until 2026-06-03

# Show only blockers for this week
devlog show week --section blockers

# Show only entries tagged "backend" in a range
devlog show --from 2026-05-01 --tag backend

# Combine filters and JSON output
devlog show --from yesterday --section notes --section action_items --json
```

`week` is a shorthand for the last 7 days (equivalent to `--from <6 days ago>`). `--date` is not compatible with `--from`/`--until` or with `week` — use a specific date (e.g. `devlog show 2026-05-28`) to target a single past day.

When `--from` is used, `--json` output is a JSON array (one object per day). For a single day, `--json` returns a single object.

---

### `sync`

Import today's commits and PRs from configured repos.

```
devlog sync [flags]
```

| Flag | Description |
|------|-------------|
| `--quiet` | Suppress all output |

```bash
devlog sync
devlog sync --quiet        # used in git hooks
devlog sync --date 2026-05-30   # back-fill a specific date
```

`sync` is idempotent — running it multiple times deduplicates by commit SHA and PR number. The post-commit hook runs `devlog sync --quiet` automatically when installed.

---

### `search`

Full-text search across all journal entries.

```
devlog search <query> [flags]
```

| Flag | Description |
|------|-------------|
| `--section <name>` | Search only these sections: `notes`, `blockers`, `action_items`, `commits`, `prs` (repeatable or comma-separated) |
| `--from <YYYY-MM-DD>` | Earliest date to search |
| `--until <YYYY-MM-DD>` | Latest date to search |
| `--json` | Emit structured JSON |

```bash
devlog search "rate limiter"
devlog search "staging" --section blockers
devlog search "auth" --from 2026-05-01 --until 2026-05-31
devlog search "deploy" --section commits --section prs --json
```

Without `--from`/`--until`, search covers all available entries.

`search --json` returns an array of match objects:

```json
[
  { "date": "2026-05-28", "section": "notes", "line": "Fixed the OAuth token refresh loop" },
  { "date": "2026-05-29", "section": "commits", "line": "abc1234 fix: oauth token refresh (api-server)" }
]
```

---

### `rm`

Remove a bullet from a section by its 1-based index.

```
devlog rm --id <n> [flags]
```

| Flag | Description |
|------|-------------|
| `--id <n>` | 1-based index of the bullet to remove (required) |
| `--section <name>` | Section to remove from (default: `notes`) |

```bash
# Show today's notes with indices
devlog show today

# Remove the 2nd bullet from notes
devlog rm --id 2

# Remove the 1st blocker
devlog rm --id 1 --section blockers

# Target a past date
devlog rm --id 3 --date 2026-05-28
```

Use `devlog show today` first to confirm bullet indices before removing.

---

### `update`

Replace a bullet in a section by its 1-based index.

```
devlog update --id <n> "new text" [flags]
```

| Flag | Description |
|------|-------------|
| `--id <n>` | 1-based index of the bullet to replace (required) |
| `--section <name>` | Section to update (default: `notes`) |

```bash
# Replace the 1st note
devlog update --id 1 "Shipped rate limiter — also fixed edge case with empty token"

# Replace the 2nd action item
devlog update --id 2 "Write integration tests (done)" --section action_items
```

---

### `resolve`

Mark a blocker or action item as resolved.

```
devlog resolve <id>
```

`<id>` is the full UUID or 8-character prefix shown by `devlog add` or `devlog items`.
Resolving records the exact time of resolution so you can query what was done and when.

```bash
# Add a blocker (prints: added blocker: a1b2c3d4)
devlog add "Waiting on staging DB" --section blockers

# Resolve it when it's unblocked
devlog resolve a1b2c3d4

# See the resolved item with its timestamp
devlog items --resolved
```

---

### `reopen`

Reopen a resolved blocker or action item — the inverse of `resolve`. Clears the
resolution timestamp and flips the item back to unresolved, so it reappears in the
default `devlog items` view.

```
devlog reopen <id>
```

`<id>` is the full UUID or 8-character prefix, matched like `resolve`. Reopening an
item that is already open is a harmless no-op.

```bash
# Resolved something prematurely? Reopen it.
devlog reopen a1b2c3d4
```

---

### `items`

List blockers and action items.

```
devlog items [flags]
```

| Flag | Description |
|------|-------------|
| *(none)* | Show only open (unresolved) items |
| `--resolved` | Show only resolved items |
| `--all` | Show both resolved and unresolved |
| `--type <name>` | Filter by type: `blockers` or `action_items` |
| `--from <date>` | For resolved items: only those resolved on or after this date |
| `--until <date>` | For resolved items: only those resolved on or before this date |
| `--json` | Emit structured JSON |

```bash
devlog items                              # open items
devlog items --resolved                   # what got resolved
devlog items --resolved --from yesterday  # resolved since yesterday
devlog items --resolved --from 2026-06-01 --until 2026-06-07
devlog items --all                        # everything (open + resolved)
devlog items --all --type blockers        # every blocker ever
devlog items --json                       # machine-readable output
```

`items --json` returns an array:

```json
[
  {
    "id": "a1b2c3d4-e5f6-4789-b012-c3d4e5f67890",
    "short_id": "a1b2c3d4",
    "type": "blocker",
    "text": "Waiting on staging DB",
    "resolved": true,
    "resolved_at": "2026-06-03T14:22:00Z",
    "dependencies": []
  }
]
```

`resolved_at` is omitted for unresolved items.

---

### `workspace`

Manage workspace directories for automatic repo discovery.

```
devlog workspace <subcommand>
```

See [Workspace Management](#workspace-management) for full details.

| Subcommand | Description |
|------------|-------------|
| `workspace add <path>` | Register a workspace directory |
| `workspace list` | List workspaces and their discovered repos |
| `workspace exclude <repo-path>` | Exclude a specific repo path from its parent workspace |

---

### `tags`

List all tags you've used across your journal, or rename a tag across all entries.

```
devlog tags [list] [--json]
devlog tags rename <old> <new>
```

**`tags list`** scans every day file and counts how many days carry each tag. Results are sorted by usage count (descending), then alphabetically. Calling `devlog tags` with no subcommand defaults to `list`.

```bash
devlog tags                      # same as tags list
devlog tags list                 # all tags with counts
devlog tags list --json          # machine-readable
```

Terminal output (two-column):

```
auth                           12
backend                         5
frontend                        2
```

`tags list --json` schema:

```json
{
  "version": "1",
  "tags": [
    { "tag": "auth",     "count": 12 },
    { "tag": "backend",  "count":  5 },
    { "tag": "frontend", "count":  2 }
  ]
}
```

**`tags rename`** replaces every occurrence of `<old>` with `<new>` across all day files. Matching on `<old>` is case-insensitive; `<new>` must conform to the tag format. Tag order within each entry is preserved. Files that don't contain the tag are not written.

```bash
devlog tags rename auth oauth            # rename across all entries
devlog tags rename old_name new_name     # underscores allowed
```

Output confirms the number of files changed:

```
Renamed tag "auth" → "oauth" in 12 file(s).
```

If no entries use the tag, you'll see:

```
No entries use tag "auth".
```

**Notes:**
- Renaming a tag to itself (case-insensitively) is an error.
- Use `devlog tags list` to discover available tags before filtering with `devlog show --tag`.

---

## Daily File Format

Each day is stored as `~/devlog/YYYY-MM-DD.md`. The full schema including all sections:

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

## Blockers
- Waiting on DevOps to provision staging DB

## Action Items
- Write integration tests for /v2/users endpoint

## Commits
- abc1234 fix: oauth token refresh loop (api-server)
- def5678 feat: add rate limiter to /v2/users (api-server)

## PRs
- #142 Add rate limiter to user endpoints [merged] (api-server)
```

**Section ownership:**

| Section | Written by | Edit how |
|---------|-----------|---------|
| Frontmatter (`commits`, `prs`) | `devlog sync` | Do not edit manually |
| `## Notes` | `devlog add` / `devlog edit` | `devlog add`, `devlog rm`, `devlog update` |
| `## Blockers` | `devlog add --section blockers` | `devlog rm/update --section blockers` |
| `## Action Items` | `devlog add --section action_items` | `devlog rm/update --section action_items` |
| `## Commits` / `## PRs` | Rendered from frontmatter | Do not edit manually |

The `## Commits` and `## PRs` body sections are re-rendered from frontmatter on every `sync` — manual edits there will be overwritten. Use the `--json` output to read structured commit and PR data programmatically.

---

## Workspace Management

Workspaces let devlog auto-discover repos within a directory tree, so you don't have to list every repo individually in `config.toml`.

### Register a workspace

```bash
devlog workspace add ~/workspace
```

devlog will scan `~/workspace` for git repos and add them to sync automatically.

### List discovered repos

```bash
devlog workspace list
```

Shows each registered workspace and the repos devlog found within it.

### Exclude a repo

If a workspace contains repos you don't want tracked:

```bash
devlog workspace exclude ~/workspace/vendor/some-lib
```

Excluded repos are remembered across syncs — they won't reappear.

### Workspaces vs. explicit repos

Both can coexist in `config.toml`. Explicit `[[repos]]` entries always sync. Workspace-discovered repos sync unless excluded. For a monorepo or tightly scoped setup, explicit repos are simpler. For a workspace with many projects, workspace auto-discovery reduces maintenance.

> **Note:** `--json` has no effect on `workspace` commands. Their output is always plain text.

---

## Filtering and Searching

### Discover available tags

Use `devlog tags list` to see every tag you've used and how often, before filtering with `--tag`:

```bash
devlog tags list                 # see what tags exist
devlog show week --tag backend   # then filter
```

### Filter `show` output

```bash
# Show only notes and action items for this week
devlog show week --section notes --section action_items

# Comma-separated is equivalent
devlog show week --section notes,action_items

# All entries tagged "backend" in the last week
devlog show week --tag backend

# Only days with activity in the "api-server" repo
devlog show week --repo api-server

# Combine: notes from days with api-server activity
devlog show week --section notes --repo api-server
```

### Search across all history

```bash
# Find any entry mentioning "rate limiter"
devlog search "rate limiter"

# Limit to a date range
devlog search "auth" --from 2026-05-01 --until 2026-05-31

# Search only in commits and PRs
devlog search "deploy" --section commits --section prs

# Search blockers for a specific issue
devlog search "DB provisioning" --section blockers

# Output as JSON for scripting
devlog search "migration" --json
```

### Target a specific date with `--date`

The global `--date` flag is honored by: `show`, `add`, `edit`, `rm`, `update`, and `sync`.
It is **not** used by `search` (use `--from`/`--until` for date-bounded search).

```bash
devlog show --date 2026-05-28
devlog add "Back-filled note" --date 2026-05-28
devlog sync --date 2026-05-28
```

---

## AI Agent Integration

devlog is designed for CLI-based agent workflows. Every output command emits `--json`. No MCP server required.

### Core pattern

```bash
# 1. Read current context before starting work
devlog show today --json

# 2. Log completed tasks
devlog add "Refactored auth middleware"

# 3. Import git activity
devlog sync --quiet

# 4. Review recent activity as JSON
devlog show --from yesterday --json
```

### `show --json` schema

```json
{
  "version": "1",
  "date": "2026-06-01",
  "tags": ["auth", "backend"],
  "sections": {
    "notes":        ["Refactored auth middleware..."],
    "blockers":     ["Waiting on DB provisioning"],
    "action_items": ["Write integration tests"],
    "commits":      [{ "sha": "abc1234", "message": "fix: oauth token refresh", "repo": "api-server" }],
    "prs":          [{ "number": 142, "title": "Add rate limiter", "state": "merged", "repo": "api-server" }]
  }
}
```

### `show --from DATE --json` schema (date range)

Returns a JSON array — one element per matching day, each with the same shape as the single-day schema:

```json
[
  {
    "version": "1",
    "date": "2026-06-01",
    "tags": ["auth"],
    "sections": {
      "notes":        ["Refactored auth middleware..."],
      "blockers":     [],
      "action_items": [],
      "commits":      [{ "sha": "abc1234", "message": "fix: oauth token refresh", "repo": "api-server" }],
      "prs":          [{ "number": 142, "title": "Add rate limiter", "state": "merged", "repo": "api-server" }]
    }
  }
]
```

> Check `version` before parsing. Additive fields keep version `"1"`, breaking changes increment it.

### `tags list --json` schema

```json
{
  "version": "1",
  "tags": [
    { "tag": "auth",     "count": 12 },
    { "tag": "backend",  "count":  5 }
  ]
}
```

```bash
# List tags with count > 3
devlog tags list --json | jq '[.tags[] | select(.count > 3)]'

# Extract just tag names sorted by count
devlog tags list --json | jq '[.tags[].tag]'
```

### jq patterns

```bash
# Extract today's blocker list
devlog show today --json | jq '.sections.blockers[]'

# Check if any PRs were merged today
devlog show today --json | jq '[.sections.prs[] | select(.state == "merged")] | length'

# List all repos with commits today
devlog show today --json | jq '[.sections.commits[].repo] | unique'

# Flatten commits across a date range into a single list
devlog show --from yesterday --json | jq '[.[].sections.commits[]]'

# All merged PRs from the past week
devlog show --from yesterday --json | jq '[.[].sections.prs[] | select(.state == "MERGED")]'

# Search returns [{date, section, line}] — count matches per date
devlog search "auth" --json | jq 'group_by(.date) | map({date: .[0].date, count: length})'

# Extract only blocker matches from search
devlog search "staging" --json | jq '[.[] | select(.section == "blockers") | .line]'
```

### Environment variables

| Variable | Effect |
|----------|--------|
| `DEVLOG_DIR` | Journal directory (default: `~/devlog`) |
| `DEVLOG_EDITOR` | Editor for `devlog edit` |
| `DEVLOG_GITHUB_TOKEN` | GitHub token for REST-based PR import |

```bash
export DEVLOG_DIR=~/my-journal
export DEVLOG_GITHUB_TOKEN=ghp_...
devlog sync
```

---

## Troubleshooting

### `gh` CLI not found or not authenticated

devlog falls back gracefully — commits still import, PRs are skipped silently. To enable PR import:

```bash
brew install gh          # or your platform's package manager
gh auth login
```

### GitHub PRs not importing via REST API

Set `DEVLOG_GITHUB_TOKEN` or add `token` to `config.toml` under `[github]`. The token needs `repo` scope for private repos, `public_repo` for public repos only.

### Post-commit hook not running

Check that the hook exists and is executable:

```bash
cat ~/workspace/your-repo/.git/hooks/post-commit
chmod +x ~/workspace/your-repo/.git/hooks/post-commit
```

If it exists but `devlog` isn't on `$PATH` inside the hook environment, add the full path:

```sh
/Users/you/go/bin/devlog sync --quiet
```

### Sync writes duplicate entries

`devlog sync` deduplicates by commit SHA and PR number within a day file. If you see duplicates, the entries likely have different SHAs (e.g., rebased commits) or the day file was manually edited in a way that broke frontmatter parsing. Open the file with `devlog edit` and remove the duplicate frontmatter entries.

### `devlog show week` missing some days

`week` shows the last 7 days. Days without a journal file (no `devlog add` or `devlog sync` ran that day) are skipped silently. To check if a file exists:

```bash
ls ~/devlog/2026-05-*.md
```

### Config file location

```bash
cat ~/.config/devlog/config.toml
```

If the file is missing, run `devlog init` to create it.

### `devlog` not found after `go install`

Make sure `~/go/bin` is on your `$PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Add that line to your shell profile (`~/.zshrc`, `~/.bashrc`, etc.) to persist it.
