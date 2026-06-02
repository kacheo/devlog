package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// DiscoverRepos scans dir one level deep for git repositories.
// Any entry whose absolute path appears in exclude is skipped.
func DiscoverRepos(dir string, exclude []string) ([]RepoConfig, error) {
	excludeSet := make(map[string]bool, len(exclude))
	for _, e := range exclude {
		excludeSet[e] = true
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil // workspace dir gone; caller decides whether to warn
		}
		return nil, fmt.Errorf("reading workspace dir %q: %w", dir, err)
	}

	var repos []RepoConfig
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		absPath := filepath.Join(dir, entry.Name())
		if excludeSet[absPath] {
			continue
		}
		if _, err := os.Stat(filepath.Join(absPath, ".git")); err == nil {
			repos = append(repos, RepoConfig{
				Path: absPath,
				Name: entry.Name(),
			})
		}
	}
	return repos, nil
}

// EffectiveRepos returns the union of explicitly-configured repos and
// workspace-discovered repos. Explicit entries take precedence on path collision.
func (c *Config) EffectiveRepos() ([]RepoConfig, error) {
	seen := make(map[string]bool, len(c.Repos))
	result := make([]RepoConfig, len(c.Repos))
	copy(result, c.Repos)
	for _, r := range c.Repos {
		seen[r.Path] = true
	}

	for _, ws := range c.Workspaces {
		discovered, err := DiscoverRepos(ws.Path, ws.Exclude)
		if err != nil {
			return nil, fmt.Errorf("workspace %q: %w", ws.Name, err)
		}
		for _, r := range discovered {
			if !seen[r.Path] {
				seen[r.Path] = true
				result = append(result, r)
			}
		}
	}
	return result, nil
}
