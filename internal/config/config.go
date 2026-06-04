package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config is the top-level configuration structure.
type Config struct {
	Journal    JournalConfig     `toml:"journal"`
	GitHub     GitHubConfig      `toml:"github"`
	Repos      []RepoConfig      `toml:"repos"`
	Workspaces []WorkspaceConfig `toml:"workspaces"`
}

// JournalConfig holds journal directory and editor settings.
type JournalConfig struct {
	Dir    string `toml:"dir"`
	Editor string `toml:"editor"`
}

// GitHubConfig holds the optional GitHub token.
type GitHubConfig struct {
	Token string `toml:"token"`
}

// RepoConfig represents a watched repository.
type RepoConfig struct {
	Path       string `toml:"path"`
	Name       string `toml:"name"`
	GitHubSlug string `toml:"github_slug"`
}

// WorkspaceConfig represents a directory auto-scanned for git repos.
type WorkspaceConfig struct {
	Path    string   `toml:"path"`
	Name    string   `toml:"name"`
	Exclude []string `toml:"exclude"`
}

// DefaultPath returns the canonical config file path (~/.config/devlog/config.toml).
func DefaultPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "devlog", "config.toml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "devlog", "config.toml")
}

// Load reads config from path, applying environment variable overrides.
// If the file does not exist, a default (zero-value) config is returned — not an error.
func Load(path string) (*Config, error) {
	cfg := &Config{}

	data, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	} else {
		if _, err := toml.Decode(string(data), cfg); err != nil {
			return nil, fmt.Errorf("parsing config: %w", err)
		}
	}

	// Tilde expansion
	cfg.Journal.Dir = ExpandHome(cfg.Journal.Dir)
	for i := range cfg.Repos {
		cfg.Repos[i].Path = ExpandHome(cfg.Repos[i].Path)
	}
	for i := range cfg.Workspaces {
		cfg.Workspaces[i].Path = ExpandHome(cfg.Workspaces[i].Path)
		for j := range cfg.Workspaces[i].Exclude {
			cfg.Workspaces[i].Exclude[j] = ExpandHome(cfg.Workspaces[i].Exclude[j])
		}
	}

	// Environment variable overrides (applied after tilde expansion)
	if v := os.Getenv("DEVLOG_DIR"); v != "" {
		cfg.Journal.Dir = v
	}
	if v := os.Getenv("DEVLOG_EDITOR"); v != "" {
		cfg.Journal.Editor = v
	}
	if v := os.Getenv("DEVLOG_GITHUB_TOKEN"); v != "" {
		cfg.GitHub.Token = v
	}

	return cfg, nil
}

// ResolvedEditor returns the editor to use, in priority order:
// DEVLOG_EDITOR env (already applied by Load into Journal.Editor) > journal.editor config > $EDITOR > $VISUAL > "vi"
func (c *Config) ResolvedEditor() string {
	if c.Journal.Editor != "" {
		return c.Journal.Editor
	}
	if v := os.Getenv("EDITOR"); v != "" {
		return v
	}
	if v := os.Getenv("VISUAL"); v != "" {
		return v
	}
	return "vi"
}

// Write serializes the config to TOML at path, creating parent directories as needed.
// Uses atomic write (write to temp file, then rename) to avoid partial writes.
func (c *Config) Write(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("creating temp config: %w", err)
	}
	if err := toml.NewEncoder(f).Encode(c); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("encoding config: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("closing temp config: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("renaming temp config: %w", err)
	}
	return nil
}

// ExpandHome replaces a leading ~ with the user home directory.
func ExpandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, path[1:])
}
