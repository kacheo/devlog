# Changelog

All notable changes to devlog are documented here.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)  
Versioning: [Semantic Versioning](https://semver.org/)

---

## [Unreleased]

---

## [0.2.0] — 2026-06-08

### Added
- `devlog items` — list blockers and action items across all day files; flags: `--resolved`, `--all`, `--type`, `--from`, `--until`, `--json`
- `devlog tags [list]` — list all tags across all day files with usage counts, sorted by frequency then alphabetically; `--json` supported
- `devlog tags rename <old> <new>` — rename a tag (case-insensitive match) across all day files; preserves tag order, leaves unaffected files untouched
- `--due YYYY-MM-DD` flag on `devlog add` for action items — displayed in `show` and `items` output, surfaced via `--overdue` filter
- `--eta YYYY-MM-DD` flag on `devlog add` for blockers
- `--overdue` flag on `devlog items` — filters to unresolved action items past their due date
- Global UUIDs and `ResolvedAt` timestamps on blockers and action items (backward-compatible YAML)
- DAG dependency support between items (cross-file references via UUID)
- Sync progress status — prints per-repo status during `devlog sync`
- Tag format enforcement — tags must contain only lowercase letters, digits, and underscores

---

## [0.1.0] — 2026-06-04

### Added
- `devlog show --from DATE [--until DATE]` — date-range view replacing the removed standup command
- `devlog search QUERY` — full-text search across all journal entries
- `devlog rm SECTION INDEX` — remove a section bullet by 1-based index
- `devlog update SECTION INDEX "text"` — replace a section bullet by 1-based index
- `devlog workspace add/list/exclude` — manage workspace directories for auto-discovery
- `scripts/install.sh` — curl-downloadable install script with checksum verification
- Homebrew formula via goreleaser `brews:` block
- `SECURITY.md` — token handling guidance and vulnerability reporting
- `CONTRIBUTING.md` — development setup, commit conventions, and release process
- CI coverage enforcement at 70% threshold
- macOS runner added to CI matrix
- Dependabot for Go modules and GitHub Actions
- `sync.Mutex` in Store to serialize same-process goroutines
- `post-commit` hook idempotent managed-block install/update
- GitHub PR import via `gh` CLI (preferred) or REST API with token
- `devlog init --non-interactive` for agent/CI bootstrap
- `--json` flag on `show` and `sync` with versioned schema (`"version": "1"`)
- Apache 2.0 license
- golangci-lint with errcheck, staticcheck, govet, nolintlint
- goreleaser config for macOS (arm64/amd64) and Linux (amd64) binaries

### Removed
- `devlog standup` — superseded by `devlog show --from/--until` date range

---

[Unreleased]: https://github.com/kacheo/devlog/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/kacheo/devlog/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/kacheo/devlog/releases/tag/v0.1.0
