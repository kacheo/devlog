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
	os.MkdirAll(filepath.Join(ws, "not-a-repo"), 0755)

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

func TestDiscoverRepos_ErrorOnUnreadableDir(t *testing.T) {
	repos, err := DiscoverRepos("/nonexistent-dir-that-does-not-exist", nil)
	if err == nil {
		t.Fatalf("expected error for non-existent dir, got repos: %v", repos)
	}
}
