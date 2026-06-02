package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DefaultValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("[journal]\ndir = \"/tmp/devlog-test\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Journal.Dir != "/tmp/devlog-test" {
		t.Errorf("Journal.Dir = %q, want /tmp/devlog-test", cfg.Journal.Dir)
	}
	if cfg.Journal.Editor != "" {
		t.Errorf("Journal.Editor = %q, want empty", cfg.Journal.Editor)
	}
	if cfg.GitHub.Token != "" {
		t.Errorf("GitHub.Token = %q, want empty", cfg.GitHub.Token)
	}
}

func TestLoad_Repos(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[journal]
dir = "~/devlog"

[[repos]]
path = "~/workspace/api-server"
name = "api-server"

[[repos]]
path = "~/workspace/frontend"
name = "frontend"
github_slug = "acme/frontend"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Repos) != 2 {
		t.Fatalf("len(Repos) = %d, want 2", len(cfg.Repos))
	}
	if cfg.Repos[0].Name != "api-server" {
		t.Errorf("Repos[0].Name = %q, want api-server", cfg.Repos[0].Name)
	}
	if cfg.Repos[1].GitHubSlug != "acme/frontend" {
		t.Errorf("Repos[1].GitHubSlug = %q, want acme/frontend", cfg.Repos[1].GitHubSlug)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("[journal]\ndir = \"~/devlog\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DEVLOG_DIR", "/custom/dir")
	t.Setenv("DEVLOG_EDITOR", "nano")
	t.Setenv("DEVLOG_GITHUB_TOKEN", "ghp_test")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Journal.Dir != "/custom/dir" {
		t.Errorf("Journal.Dir = %q, want /custom/dir", cfg.Journal.Dir)
	}
	if cfg.Journal.Editor != "nano" {
		t.Errorf("Journal.Editor = %q, want nano", cfg.Journal.Editor)
	}
	if cfg.GitHub.Token != "ghp_test" {
		t.Errorf("GitHub.Token = %q, want ghp_test", cfg.GitHub.Token)
	}
}

func TestLoad_FileNotFound_ReturnsDefault(t *testing.T) {
	// Clear env overrides to get a clean default
	t.Setenv("DEVLOG_DIR", "")
	t.Setenv("DEVLOG_EDITOR", "")
	t.Setenv("DEVLOG_GITHUB_TOKEN", "")
	cfg, err := Load("/nonexistent/path/config.toml")
	if err != nil {
		t.Fatalf("Load() on missing file should not error, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultPath()
	if path == "" {
		t.Error("DefaultPath() returned empty string")
	}
	if filepath.Base(path) != "config.toml" {
		t.Errorf("DefaultPath() base = %q, want config.toml", filepath.Base(path))
	}
}

func TestConfig_JournalDirExpanded(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte("[journal]\ndir = \"~/devlog\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Clear DEVLOG_DIR so tilde expansion is tested on the config value
	t.Setenv("DEVLOG_DIR", "")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Journal.Dir) > 0 && cfg.Journal.Dir[0] == '~' {
		t.Errorf("Journal.Dir still contains tilde: %q", cfg.Journal.Dir)
	}
}

func TestWrite_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	cfg := &Config{
		Journal: JournalConfig{Dir: "~/devlog"},
		GitHub:  GitHubConfig{Token: "ghp_abc"},
		Repos: []RepoConfig{
			{Path: "~/workspace/api", Name: "api"},
		},
	}
	if err := cfg.Write(path); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() after Write() error = %v", err)
	}
	// Dir is tilde-expanded on Load, so compare expanded form
	home, _ := os.UserHomeDir()
	wantDir := filepath.Join(home, "devlog")
	if got.Journal.Dir != wantDir {
		t.Errorf("Journal.Dir = %q, want %q", got.Journal.Dir, wantDir)
	}
	if len(got.Repos) != 1 || got.Repos[0].Name != "api" {
		t.Errorf("Repos mismatch: %+v", got.Repos)
	}
}

func TestWrite_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.toml")
	cfg := &Config{Journal: JournalConfig{Dir: "/tmp"}}
	if err := cfg.Write(path); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("config file not created: %v", err)
	}
}

func TestWrite_AtomicNoPanicOnError(t *testing.T) {
	// Writing to a path whose parent doesn't exist should still error cleanly
	cfg := &Config{Journal: JournalConfig{Dir: "/tmp"}}
	// Write creates parent dirs, so use a read-only parent to force failure
	// (Skip on systems where root can always write)
	dir := t.TempDir()
	if err := os.Chmod(dir, 0500); err != nil {
		t.Skip("cannot set dir read-only")
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0700) })
	path := filepath.Join(dir, "sub", "config.toml")
	err := cfg.Write(path)
	if err == nil {
		t.Error("Write() to unwritable dir should error")
	}
}

func TestResolvedEditor(t *testing.T) {
	tests := []struct {
		name         string
		configEditor string
		devlogEditor string
		envEditor    string
		envVisual    string
		wantEditor   string
	}{
		{"config takes priority", "vim", "", "", "", "vim"},
		{"DEVLOG_EDITOR overrides config", "vim", "emacs", "", "", "emacs"},
		{"falls back to EDITOR", "", "", "nano", "", "nano"},
		{"falls back to VISUAL", "", "", "", "code", "code"},
		{"ultimate fallback is vi", "", "", "", "", "vi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set all env vars explicitly so tests are independent
			t.Setenv("DEVLOG_EDITOR", tt.devlogEditor)
			t.Setenv("EDITOR", tt.envEditor)
			t.Setenv("VISUAL", tt.envVisual)

			dir := t.TempDir()
			path := filepath.Join(dir, "config.toml")
			content := "[journal]\ndir=\"/tmp\"\neditor=\"" + tt.configEditor + "\"\n"
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				t.Fatal(err)
			}

			cfg, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			got := cfg.ResolvedEditor()
			if got != tt.wantEditor {
				t.Errorf("ResolvedEditor() = %q, want %q", got, tt.wantEditor)
			}
		})
	}
}

func TestLoad_WorkspaceConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	content := `
[[workspaces]]
path = "~/projects"
name  = "main"
exclude = ["~/projects/dotfiles"]
`
	if err := os.WriteFile(cfgPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Workspaces) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(cfg.Workspaces))
	}
	ws := cfg.Workspaces[0]
	home, _ := os.UserHomeDir()
	if ws.Path != filepath.Join(home, "projects") {
		t.Errorf("expected expanded path, got %q", ws.Path)
	}
	if ws.Name != "main" {
		t.Errorf("expected name 'main', got %q", ws.Name)
	}
	if len(ws.Exclude) != 1 || ws.Exclude[0] != filepath.Join(home, "projects/dotfiles") {
		t.Errorf("unexpected exclude: %v", ws.Exclude)
	}
}
