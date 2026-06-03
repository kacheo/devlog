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

// TestIntegration_GlobalItemsFlow exercises: add global item → show surfaces it → resolve → show omits it
func TestIntegration_GlobalItemsFlow(t *testing.T) {
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

	// Add a note first so a day file exists for show to render
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"add", "daily note"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add note: %v", err)
	}

	// Add a global blocker; capture short ID from output
	buf.Reset()
	addSection = ""
	rootCmd.SetArgs([]string{"add", "--section", "blocker", "waiting on infra"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add blocker: %v", err)
	}
	// output: "added blocker: <shortID>"
	parts := strings.Fields(buf.String())
	if len(parts) < 3 {
		t.Fatalf("unexpected add output: %q", buf.String())
	}
	shortID := parts[len(parts)-1]

	// show today: blocker should be visible alongside the note
	buf.Reset()
	addSection = ""
	rootCmd.SetArgs([]string{"show", "today"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("show: %v", err)
	}
	if !strings.Contains(buf.String(), "waiting on infra") {
		t.Errorf("show missing blocker: %q", buf.String())
	}

	// resolve the blocker
	buf.Reset()
	rootCmd.SetArgs([]string{"resolve", shortID})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !strings.Contains(buf.String(), "resolved: blocker") {
		t.Errorf("expected 'resolved: blocker' in output: %q", buf.String())
	}

	// show today: resolved blocker should not appear
	buf.Reset()
	rootCmd.SetArgs([]string{"show", "today"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("show after resolve: %v", err)
	}
	if strings.Contains(buf.String(), "waiting on infra") {
		t.Errorf("resolved blocker should not appear in show: %q", buf.String())
	}
}

// TestIntegration_ShowRange_IncludesGlobalItems verifies unresolved global items appear in --from/--until range output
func TestIntegration_ShowRange_IncludesGlobalItems(t *testing.T) {
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
		showFrom = ""
		showUntil = ""
	})

	// Create a day file entry so the range query has at least one entry to render
	st, _ := store.New(dir)
	e := store.EmptyEntry(time.Now())
	e.Sections["notes"] = []string{"day note"}
	if err := st.Save(e); err != nil {
		t.Fatal(err)
	}

	// Add a global blocker; capture short ID
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"add", "--section", "blocker", "range test blocker"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add blocker: %v", err)
	}
	parts := strings.Fields(buf.String())
	if len(parts) < 3 {
		t.Fatalf("unexpected add output: %q", buf.String())
	}
	shortID := parts[len(parts)-1]

	today := time.Now().Format("2006-01-02")

	// show --from today --until today: global blocker should appear
	buf.Reset()
	addSection = ""
	rootCmd.SetArgs([]string{"show", "--from", today, "--until", today})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("show range: %v", err)
	}
	if !strings.Contains(buf.String(), "range test blocker") {
		t.Errorf("range show missing blocker: %q", buf.String())
	}

	// resolve, then range query should omit it
	buf.Reset()
	rootCmd.SetArgs([]string{"resolve", shortID})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	buf.Reset()
	showFrom = ""
	showUntil = ""
	rootCmd.SetArgs([]string{"show", "--from", today, "--until", today})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("show range after resolve: %v", err)
	}
	if strings.Contains(buf.String(), "range test blocker") {
		t.Errorf("resolved blocker should not appear in range show: %q", buf.String())
	}
}
