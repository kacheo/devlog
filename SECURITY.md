# Security Policy

## Supported Versions

devlog is currently pre-v1. The latest release on the `main` branch receives security fixes.

## Reporting a Vulnerability

Please **do not** open a public GitHub issue for security vulnerabilities. Instead, email
the maintainer directly at the address in the git log, or open a
[GitHub Security Advisory](https://github.com/kacheo/devlog/security/advisories/new).

You'll receive a response within 72 hours. If the vulnerability is confirmed, a fix will be
released as soon as possible — typically within 7 days for critical issues.

## GitHub Token Handling

devlog can optionally use a GitHub token for PR import. Important notes:

- **Storage**: The token is stored in plaintext in `~/.config/devlog/config.toml`. This file
  is created with permissions that restrict access to your user account. Do not commit this
  file to version control.

- **Safer alternative — environment variable**: Set `DEVLOG_GITHUB_TOKEN` in your shell
  instead of writing the token to config. Environment variables are not persisted to disk.

  ```bash
  export DEVLOG_GITHUB_TOKEN="ghp_..."
  ```

- **Safest alternative — `gh` CLI**: Install the [GitHub CLI](https://cli.github.com/) and
  run `gh auth login`. devlog will use its credential store automatically and no token needs
  to be stored in the config file at all.

- **Minimal scope**: If you do store a token, generate a fine-grained token with read-only
  access to the repositories you want to import PRs from. devlog never writes to GitHub.

## Data Storage

devlog stores all journal data locally in `~/devlog/` (or your configured `journal.dir`).
No data is transmitted to any remote server unless you explicitly run `devlog sync`, which
reads from GitHub APIs only and never writes to them.

## Dependencies

devlog has a minimal dependency surface: cobra (CLI), BurntSushi/toml (config), yaml.v3
(frontmatter), and stdlib. Dependency updates are managed via
[Dependabot](https://github.com/kacheo/devlog/blob/main/.github/dependabot.yml).
