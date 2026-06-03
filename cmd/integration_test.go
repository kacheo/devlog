package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kacheo/devlog/internal/store"
)

// TestIntegration_AddShowRoundTrip exercises: add → show terminal → show JSON
func TestIntegration_AddShowRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalJSON = false
	globalDate = ""
	addSection = ""
	addTags = nil
	addDeps = nil
	t.Cleanup(func() {
		globalJSON = false
		globalDate = ""
		addSection = ""
		addTags = nil
		addDeps = nil
	})

	// Add a note
	rootCmd.SetArgs([]string{"add", "completed the auth refactor"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add: %v", err)
	}

	// Add to blockers
	addSection = ""
	rootCmd.SetArgs([]string{"add", "--section", "blockers", "waiting on DB migration"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add blockers: %v", err)
	}

	// Show terminal
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	addSection = ""
	rootCmd.SetArgs([]string{"show", "today"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("show: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "completed the auth refactor") {
		t.Errorf("show terminal missing note:\n%s", out)
	}
	if !strings.Contains(out, "waiting on DB migration") {
		t.Errorf("show terminal missing blocker:\n%s", out)
	}

	// Show JSON
	globalJSON = true
	buf.Reset()
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"show", "today"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("show --json: %v", err)
	}

	var v map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
		t.Fatalf("show JSON invalid: %v\nOutput: %s", err, buf.String())
	}
	sections := v["sections"].(map[string]interface{})
	notes := sections["notes"].([]interface{})
	if len(notes) != 1 || notes[0] != "completed the auth refactor" {
		t.Errorf("JSON notes = %v", notes)
	}
	// Blockers are now objects with id/text/resolved/dependencies
	blockers := sections["blockers"].([]interface{})
	if len(blockers) != 1 {
		t.Fatalf("JSON blockers len = %d, want 1", len(blockers))
	}
	b := blockers[0].(map[string]interface{})
	if b["text"] != "waiting on DB migration" {
		t.Errorf("JSON blocker text = %v, want 'waiting on DB migration'", b["text"])
	}
	if b["resolved"] != false {
		t.Errorf("JSON blocker resolved = %v, want false", b["resolved"])
	}
}

// TestIntegration_SyncDeduplication verifies that running sync twice doesn't duplicate commits
func TestIntegration_SyncDeduplication(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalJSON = false
	globalDate = ""
	t.Cleanup(func() { globalJSON = false; globalDate = "" })

	// Pre-populate with a commit
	st, _ := store.New(dir)
	date := time.Now()
	entry := store.EmptyEntry(date)
	entry.Commits = []store.Commit{{SHA: "abc1234", Message: "fix: existing commit", Repo: "api"}}
	if err := st.Save(entry); err != nil {
		t.Fatal(err)
	}

	// mergeCommits with the same commit twice = still 1
	existing := []store.Commit{{SHA: "abc1234", Message: "fix: existing commit", Repo: "api"}}
	again := []store.Commit{{SHA: "abc1234", Message: "fix: existing commit", Repo: "api"}}
	result := mergeCommits(existing, again)
	if len(result) != 1 {
		t.Errorf("mergeCommits idempotent: len = %d, want 1", len(result))
	}
}

// TestIntegration_TagFlow verifies --tag adds to frontmatter and appears in JSON
func TestIntegration_TagFlow(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalJSON = false
	globalDate = ""
	addSection = ""
	addTags = nil
	addDeps = nil
	t.Cleanup(func() {
		globalJSON = false
		globalDate = ""
		addSection = ""
		addTags = nil
		addDeps = nil
	})

	rootCmd.SetArgs([]string{"add", "--tag", "infra", "--tag", "backend", "deployed new service"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add with tags: %v", err)
	}

	// Add again with same tag — should deduplicate
	addTags = nil
	rootCmd.SetArgs([]string{"add", "--tag", "infra", "another note"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add second with duplicate tag: %v", err)
	}

	globalJSON = true
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"show", "today"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("show --json: %v", err)
	}

	var v map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
		t.Fatalf("JSON invalid: %v", err)
	}
	tags := v["tags"].([]interface{})
	// Should have "infra" and "backend" but NOT a second "infra"
	tagStrs := make([]string, len(tags))
	for i, tag := range tags {
		tagStrs[i] = tag.(string)
	}
	count := 0
	for _, tag := range tagStrs {
		if tag == "infra" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("tag 'infra' appears %d times, want 1 (dedup): %v", count, tagStrs)
	}
}

// TestIntegration_HookInstall verifies installHook creates an executable file
func TestIntegration_HookInstall(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	hookPath := filepath.Join(gitDir, "post-commit")

	if err := installHook(hookPath); err != nil {
		t.Fatalf("installHook: %v", err)
	}

	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "devlog sync --quiet") {
		t.Error("hook missing expected content")
	}
	info, _ := os.Stat(hookPath)
	if info.Mode()&0111 == 0 {
		t.Error("hook not executable")
	}
}
