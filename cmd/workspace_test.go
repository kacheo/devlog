package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kacheo/devlog/internal/config"
)

func TestWorkspaceAdd_AppendsToConfig(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.toml")
	wsDir := t.TempDir()
	cfg := &config.Config{}
	_ = cfg.Write(cfgPath)

	added, err := addWorkspaceToConfig(cfgPath, wsDir)
	if err != nil {
		t.Fatalf("addWorkspaceToConfig: %v", err)
	}
	if !added {
		t.Error("expected added=true on first registration")
	}

	loaded, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if len(loaded.Workspaces) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(loaded.Workspaces))
	}
	if loaded.Workspaces[0].Path != wsDir {
		t.Errorf("unexpected path: %q", loaded.Workspaces[0].Path)
	}

	// Idempotent: second call returns added=false, no duplicate
	added2, err := addWorkspaceToConfig(cfgPath, wsDir)
	if err != nil {
		t.Fatalf("second addWorkspaceToConfig: %v", err)
	}
	if added2 {
		t.Error("expected added=false on duplicate registration")
	}
	loaded2, _ := config.Load(cfgPath)
	if len(loaded2.Workspaces) != 1 {
		t.Errorf("expected still 1 workspace, got %d", len(loaded2.Workspaces))
	}
}

func TestWorkspaceExclude_AddsToExcludeList(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.toml")
	wsDir := t.TempDir()
	repoPath := filepath.Join(wsDir, "my-repo")

	cfg := &config.Config{
		Workspaces: []config.WorkspaceConfig{
			{Path: wsDir, Name: filepath.Base(wsDir)},
		},
	}
	_ = cfg.Write(cfgPath)

	excluded, err := excludeRepoFromWorkspace(cfgPath, repoPath)
	if err != nil {
		t.Fatalf("excludeRepoFromWorkspace: %v", err)
	}
	if !excluded {
		t.Error("expected excluded=true on first call")
	}

	loaded, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if len(loaded.Workspaces[0].Exclude) != 1 || loaded.Workspaces[0].Exclude[0] != repoPath {
		t.Errorf("expected exclude list to contain %q, got %v", repoPath, loaded.Workspaces[0].Exclude)
	}

	// Idempotent: second call returns excluded=false
	excluded2, err := excludeRepoFromWorkspace(cfgPath, repoPath)
	if err != nil {
		t.Fatalf("second excludeRepoFromWorkspace: %v", err)
	}
	if excluded2 {
		t.Error("expected excluded=false on duplicate call")
	}
}

func TestWorkspaceExclude_ErrorWhenNoWorkspaceFound(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.toml")
	cfg := &config.Config{}
	_ = cfg.Write(cfgPath)

	_, err := excludeRepoFromWorkspace(cfgPath, "/some/path/repo")
	if err == nil {
		t.Fatal("expected error when no workspace registered")
	}
	if !strings.Contains(err.Error(), "no workspace found") {
		t.Errorf("error should mention 'no workspace found', got: %v", err)
	}
}

func TestWorkspaceList_PrintsDiscoveredRepos(t *testing.T) {
	wsDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(wsDir, "svc-a", ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(wsDir, "svc-b", ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Journal: config.JournalConfig{Dir: t.TempDir()},
		Workspaces: []config.WorkspaceConfig{
			{Path: wsDir, Name: "myws"},
		},
	}

	var sb strings.Builder
	err := printWorkspaceList(cfg, &sb)
	if err != nil {
		t.Fatalf("printWorkspaceList: %v", err)
	}
	out := sb.String()
	if !strings.Contains(out, "svc-a") || !strings.Contains(out, "svc-b") {
		t.Errorf("expected repo names in output, got:\n%s", out)
	}
}

func TestWorkspaceList_ShowsExcludeListWhenAllReposExcluded(t *testing.T) {
	wsDir := t.TempDir()
	repoPath := filepath.Join(wsDir, "the-repo")
	if err := os.MkdirAll(filepath.Join(repoPath, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Workspaces: []config.WorkspaceConfig{
			{Path: wsDir, Name: "myws", Exclude: []string{repoPath}},
		},
	}

	var sb strings.Builder
	if err := printWorkspaceList(cfg, &sb); err != nil {
		t.Fatalf("printWorkspaceList: %v", err)
	}
	out := sb.String()
	if !strings.Contains(out, "excluded:") {
		t.Errorf("expected exclude list in output, got:\n%s", out)
	}
	if !strings.Contains(out, repoPath) {
		t.Errorf("expected excluded repo path in output, got:\n%s", out)
	}
}
