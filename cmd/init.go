package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kacheo/devlog/internal/config"
)

const (
	hookBlockStart = "# >>> devlog managed block >>>"
	hookBlockEnd   = "# <<< devlog managed block <<<"
	hookBlockBody  = "devlog sync --quiet"
)

var (
	initNonInteractive bool
	initAddRepos       []string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "First-time setup: create config and optionally install git hooks",
	Args:  cobra.NoArgs,
	RunE:  runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initNonInteractive, "non-interactive", false, "use defaults, no prompts")
	initCmd.Flags().StringArrayVar(&initAddRepos, "add-repo", nil, "add a repo path (repeatable)")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	cfgPath := config.DefaultPath()

	// Load existing config (or start fresh)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if initNonInteractive {
		return runInitNonInteractive(cfg, cfgPath)
	}
	return runInitInteractive(cmd, cfg, cfgPath)
}

func runInitNonInteractive(cfg *config.Config, cfgPath string) error {
	// Defaults
	if cfg.Journal.Dir == "" {
		home, _ := os.UserHomeDir()
		cfg.Journal.Dir = filepath.Join(home, "devlog")
	}

	// Append repos from --add-repo flags
	for _, repoPath := range initAddRepos {
		cfg.Repos = appendRepoIfMissing(cfg, repoPath)
	}

	return cfg.Write(cfgPath)
}

func runInitInteractive(cmd *cobra.Command, cfg *config.Config, cfgPath string) error {
	w := cmd.OutOrStdout()
	r := bufio.NewReader(os.Stdin)

	fmt.Fprintln(w, "devlog setup")
	fmt.Fprintln(w, "------------")

	// Journal dir
	home, _ := os.UserHomeDir()
	defaultDir := cfg.Journal.Dir
	if defaultDir == "" {
		defaultDir = filepath.Join(home, "devlog")
	}
	fmt.Fprintf(w, "Journal directory [%s]: ", defaultDir)
	line, _ := r.ReadString('\n')
	line = strings.TrimSpace(line)
	if line != "" {
		cfg.Journal.Dir = line
	} else {
		cfg.Journal.Dir = defaultDir
	}

	// GitHub token
	fmt.Fprintf(w, "GitHub token (optional, press Enter to skip): ")
	line, _ = r.ReadString('\n')
	line = strings.TrimSpace(line)
	if line != "" {
		cfg.GitHub.Token = line
	}

	// Append repos from --add-repo flags
	for _, repoPath := range initAddRepos {
		cfg.Repos = appendRepoIfMissing(cfg, repoPath)
	}

	// Write config
	if err := cfg.Write(cfgPath); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	fmt.Fprintf(w, "Config written to %s\n", cfgPath)

	// Offer hook installation for each effective repo (explicit + workspace-discovered)
	allRepos, err := cfg.EffectiveRepos()
	if err != nil {
		fmt.Fprintf(w, "warning: discovering workspace repos: %v\n", err)
	}
	for _, repo := range allRepos {
		hookPath := filepath.Join(repo.Path, ".git", "hooks", "post-commit")
		fmt.Fprintf(w, "Install post-commit hook in %s? [y/N]: ", repo.Name)
		line, _ = r.ReadString('\n')
		line = strings.TrimSpace(strings.ToLower(line))
		if line == "y" || line == "yes" {
			if err := installHook(hookPath); err != nil {
				fmt.Fprintf(w, "warning: could not install hook for %s: %v\n", repo.Name, err)
			} else {
				fmt.Fprintf(w, "Hook installed at %s\n", hookPath)
			}
		}
	}

	return nil
}

// appendRepoIfMissing adds a repo entry for repoPath if not already in cfg.Repos.
func appendRepoIfMissing(cfg *config.Config, repoPath string) []config.RepoConfig {
	// Check if already present
	for _, r := range cfg.Repos {
		if r.Path == repoPath {
			return cfg.Repos
		}
	}
	name := filepath.Base(repoPath)
	return append(cfg.Repos, config.RepoConfig{
		Path: repoPath,
		Name: name,
	})
}

// addRepoToConfig loads the config at DefaultPath, appends the repo, and writes it back.
// Used by --add-repo in non-interactive mode and tests.
func addRepoToConfig(repoPath string) error {
	cfgPath := config.DefaultPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	cfg.Repos = appendRepoIfMissing(cfg, repoPath)
	return cfg.Write(cfgPath)
}

// installHook installs (or appends) the devlog managed block in hookPath.
// Creates the file if it doesn't exist. Replaces the block if already present.
func installHook(hookPath string) error {
	managedContent := hookBlockStart + "\n" + hookBlockBody + "\n" + hookBlockEnd + "\n"

	// Read existing content (if any)
	existing := ""
	data, err := os.ReadFile(hookPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading hook: %w", err)
	}
	if err == nil {
		existing = string(data)
	}

	var newContent string
	if existing == "" {
		// New file: shebang + managed block
		newContent = "#!/bin/sh\n" + managedContent
	} else if strings.Contains(existing, hookBlockStart) {
		// Replace existing managed block
		start := strings.Index(existing, hookBlockStart)
		end := strings.Index(existing, hookBlockEnd)
		if end < 0 {
			// Malformed — just append
			newContent = existing + "\n" + managedContent
		} else {
			end += len(hookBlockEnd)
			// Include trailing newline after end marker
			if end < len(existing) && existing[end] == '\n' {
				end++
			}
			newContent = existing[:start] + managedContent + existing[end:]
		}
	} else {
		// Append managed block to existing hook
		if !strings.HasSuffix(existing, "\n") {
			existing += "\n"
		}
		newContent = existing + managedContent
	}

	// Write with executable permission (shell scripts require 0755).
	if err := os.WriteFile(hookPath, []byte(newContent), 0755); err != nil { //nolint:gosec // G306: hook must be executable
		return fmt.Errorf("writing hook: %w", err)
	}
	return nil
}
