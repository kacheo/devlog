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

	err := addWorkspaceToConfig(cfgPath, wsDir)
	if err != nil {
		t.Fatalf("addWorkspaceToConfig: %v", err)
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
}

func TestWorkspaceExclude_AddsToExcludeList(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.toml")
	wsDir := t.TempDir()
	repoPath := filepath.Join(wsDir, "my-repo")
	os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)

	cfg := &config.Config{
		Workspaces: []config.WorkspaceConfig{
			{Path: wsDir, Name: filepath.Base(wsDir)},
		},
	}
	_ = cfg.Write(cfgPath)

	err := excludeRepoFromWorkspace(cfgPath, repoPath)
	if err != nil {
		t.Fatalf("excludeRepoFromWorkspace: %v", err)
	}

	loaded, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if len(loaded.Workspaces[0].Exclude) != 1 || loaded.Workspaces[0].Exclude[0] != repoPath {
		t.Errorf("expected exclude list to contain %q, got %v", repoPath, loaded.Workspaces[0].Exclude)
	}
}

func TestWorkspaceList_PrintsDiscoveredRepos(t *testing.T) {
	wsDir := t.TempDir()
	os.MkdirAll(filepath.Join(wsDir, "svc-a", ".git"), 0755)
	os.MkdirAll(filepath.Join(wsDir, "svc-b", ".git"), 0755)

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
