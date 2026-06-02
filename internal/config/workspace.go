package config

import (
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
