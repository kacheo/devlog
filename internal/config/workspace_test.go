package config

import (
	"os"
	"path/filepath"
	"testing"
)

func makeGitRepo(t *testing.T, parent, name string) string {
	t.Helper()
	dir := filepath.Join(parent, name)
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestDiscoverRepos_FindsGitDirs(t *testing.T) {
	ws := t.TempDir()
	repoA := makeGitRepo(t, ws, "repo-a")
	repoB := makeGitRepo(t, ws, "repo-b")
	// non-git dir
	if err := os.MkdirAll(filepath.Join(ws, "not-a-repo"), 0755); err != nil {
		t.Fatal(err)
	}

	repos, err := DiscoverRepos(ws, nil)
	if err != nil {
		t.Fatalf("DiscoverRepos: %v", err)
	}
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}
	paths := map[string]bool{repos[0].Path: true, repos[1].Path: true}
	if !paths[repoA] || !paths[repoB] {
		t.Errorf("unexpected repo paths: %v", repos)
	}
}

func TestDiscoverRepos_RespectsExclude(t *testing.T) {
	ws := t.TempDir()
	makeGitRepo(t, ws, "keep")
	excluded := makeGitRepo(t, ws, "skip")

	repos, err := DiscoverRepos(ws, []string{excluded})
	if err != nil {
		t.Fatalf("DiscoverRepos: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d: %v", len(repos), repos)
	}
	if repos[0].Path != filepath.Join(ws, "keep") {
		t.Errorf("wrong repo returned: %v", repos[0])
	}
}

func TestDiscoverRepos_NameIsBasename(t *testing.T) {
	ws := t.TempDir()
	makeGitRepo(t, ws, "my-service")

	repos, err := DiscoverRepos(ws, nil)
	if err != nil {
		t.Fatalf("DiscoverRepos: %v", err)
	}
	if len(repos) != 1 || repos[0].Name != "my-service" {
		t.Errorf("expected name 'my-service', got %+v", repos)
	}
}

func TestDiscoverRepos_EmptyForMissingDir(t *testing.T) {
	repos, err := DiscoverRepos("/nonexistent-dir-that-does-not-exist", nil)
	if err != nil {
		t.Fatalf("expected no error for missing dir (graceful skip), got: %v", err)
	}
	if len(repos) != 0 {
		t.Fatalf("expected empty repos for missing dir, got: %v", repos)
	}
}

func TestEffectiveRepos_MergesWorkspaceRepos(t *testing.T) {
	ws := t.TempDir()
	makeGitRepo(t, ws, "discovered")

	cfg := &Config{
		Workspaces: []WorkspaceConfig{
			{Path: ws, Name: "test-ws"},
		},
	}

	repos, err := cfg.EffectiveRepos()
	if err != nil {
		t.Fatalf("EffectiveRepos: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1, got %d: %v", len(repos), repos)
	}
	if repos[0].Name != "discovered" {
		t.Errorf("expected name 'discovered', got %q", repos[0].Name)
	}
}

func TestEffectiveRepos_DeduplicatesByPath(t *testing.T) {
	ws := t.TempDir()
	repoPath := makeGitRepo(t, ws, "shared")

	cfg := &Config{
		Repos: []RepoConfig{
			{Path: repoPath, Name: "explicit-name", GitHubSlug: "org/shared"},
		},
		Workspaces: []WorkspaceConfig{
			{Path: ws, Name: "test-ws"},
		},
	}

	repos, err := cfg.EffectiveRepos()
	if err != nil {
		t.Fatalf("EffectiveRepos: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1 (deduped), got %d: %v", len(repos), repos)
	}
	// Explicit entry wins — keeps its name and github_slug
	if repos[0].Name != "explicit-name" {
		t.Errorf("explicit entry should win, got name %q", repos[0].Name)
	}
	if repos[0].GitHubSlug != "org/shared" {
		t.Errorf("explicit entry should win, got slug %q", repos[0].GitHubSlug)
	}
}

func TestEffectiveRepos_ExcludesHonored(t *testing.T) {
	ws := t.TempDir()
	makeGitRepo(t, ws, "keep")
	excludedPath := makeGitRepo(t, ws, "skip")

	cfg := &Config{
		Workspaces: []WorkspaceConfig{
			{Path: ws, Name: "test-ws", Exclude: []string{excludedPath}},
		},
	}

	repos, err := cfg.EffectiveRepos()
	if err != nil {
		t.Fatalf("EffectiveRepos: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1, got %d: %v", len(repos), repos)
	}
	if repos[0].Name != "keep" {
		t.Errorf("wrong repo kept: %+v", repos[0])
	}
}

func TestEffectiveRepos_SkipsMissingWorkspaceDir(t *testing.T) {
	cfg := &Config{
		Workspaces: []WorkspaceConfig{
			{Path: "/nonexistent-workspace-path", Name: "gone-ws"},
		},
	}
	repos, err := cfg.EffectiveRepos()
	if err != nil {
		t.Fatalf("expected no error for missing workspace dir (graceful skip), got: %v", err)
	}
	if len(repos) != 0 {
		t.Fatalf("expected empty repos for missing workspace, got: %v", repos)
	}
}
