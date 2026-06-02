package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitHook_Install_NewFile(t *testing.T) {
	dir := t.TempDir()
	hookPath := filepath.Join(dir, "post-commit")

	if err := installHook(hookPath); err != nil {
		t.Fatalf("installHook() error = %v", err)
	}

	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("hook file not created: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "devlog sync --quiet") {
		t.Errorf("hook missing 'devlog sync --quiet':\n%s", content)
	}
	if !strings.Contains(content, hookBlockStart) {
		t.Errorf("hook missing start marker:\n%s", content)
	}
	if !strings.Contains(content, hookBlockEnd) {
		t.Errorf("hook missing end marker:\n%s", content)
	}
	// Must be executable
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("hook file is not executable")
	}
}

func TestInitHook_Install_ExistingHook_Appends(t *testing.T) {
	dir := t.TempDir()
	hookPath := filepath.Join(dir, "post-commit")

	// Write existing hook content
	existing := "#!/bin/sh\necho 'existing hook'\n"
	if err := os.WriteFile(hookPath, []byte(existing), 0755); err != nil {
		t.Fatal(err)
	}

	if err := installHook(hookPath); err != nil {
		t.Fatalf("installHook() error = %v", err)
	}

	data, _ := os.ReadFile(hookPath)
	content := string(data)
	if !strings.Contains(content, "existing hook") {
		t.Error("existing hook content was lost")
	}
	if !strings.Contains(content, "devlog sync --quiet") {
		t.Error("devlog block not appended")
	}
}

func TestInitHook_Install_Idempotent(t *testing.T) {
	dir := t.TempDir()
	hookPath := filepath.Join(dir, "post-commit")

	// Install twice
	if err := installHook(hookPath); err != nil {
		t.Fatalf("first installHook() error = %v", err)
	}
	if err := installHook(hookPath); err != nil {
		t.Fatalf("second installHook() error = %v", err)
	}

	data, _ := os.ReadFile(hookPath)
	content := string(data)
	// Should only have one copy of the managed block
	count := strings.Count(content, hookBlockStart)
	if count != 1 {
		t.Errorf("managed block appears %d times, want 1", count)
	}
}

func TestInitHook_AddRepo(t *testing.T) {
	dir := t.TempDir()
	// addRepoToConfig uses DefaultPath — override via env
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Write an existing config first
	cfgPath := filepath.Join(dir, "devlog", "config.toml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfgPath, []byte("[journal]\ndir = \"/tmp/devlog\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	repoPath := "/home/user/workspace/myproject"
	if err := addRepoToConfig(repoPath); err != nil {
		t.Fatalf("addRepoToConfig() error = %v", err)
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "myproject") {
		t.Errorf("config missing repo name 'myproject':\n%s", content)
	}
	if !strings.Contains(content, repoPath) {
		t.Errorf("config missing repo path:\n%s", content)
	}
}
