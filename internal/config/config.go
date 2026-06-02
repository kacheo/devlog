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
	Journal JournalConfig `toml:"journal"`
	GitHub  GitHubConfig  `toml:"github"`
	Repos   []RepoConfig  `toml:"repos"`
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
	cfg.Journal.Dir = expandHome(cfg.Journal.Dir)
	for i := range cfg.Repos {
		cfg.Repos[i].Path = expandHome(cfg.Repos[i].Path)
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
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("creating temp config: %w", err)
	}
	if err := toml.NewEncoder(f).Encode(c); err != nil {
		f.Close()
		os.Remove(tmp)
		return fmt.Errorf("encoding config: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("closing temp config: %w", err)
	}
	return os.Rename(tmp, path)
}

// expandHome replaces a leading ~ with the user home directory.
func expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, path[1:])
}
