package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kacheo/devlog/internal/store"
)

func resetAddFlags() {
	addSection = ""
	addTags = nil
	addDeps = nil
	addDue = ""
	addETA = ""
	globalDate = ""
}

func TestAddCmd_AppendsToNotes(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetAddFlags()

	rootCmd.SetArgs([]string{"add", "hello world"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Read back the file
	today := time.Now()
	path := filepath.Join(dir, today.Format("2006-01-02")+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("day file not found: %v", err)
	}

	entry, err := store.Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	notes := entry.Sections["notes"]
	if len(notes) != 1 || notes[0] != "hello world" {
		t.Errorf("notes = %v, want [hello world]", notes)
	}
}

func TestAddCmd_SectionFlag_Blockers(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetAddFlags()

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"add", "--section", "blocker", "stuck on CI"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Blocker should be in items.yaml, not the day file
	st, _ := store.New(dir)
	items, err := st.LoadAllItems()
	if err != nil {
		t.Fatalf("LoadAllItems: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Type != "blocker" || items[0].Text != "stuck on CI" {
		t.Errorf("item = %+v, want type=blocker text='stuck on CI'", items[0])
	}
	// Day file should NOT contain blockers
	today := time.Now()
	path := filepath.Join(dir, today.Format("2006-01-02")+".md")
	data, _ := os.ReadFile(path)
	if len(data) > 0 {
		entry, _ := store.Parse(data)
		if _, ok := entry.Sections["blockers"]; ok {
			t.Error("day file should not contain blockers section")
		}
	}
	// Output should contain the short ID
	output := buf.String()
	if !strings.Contains(output, "added blocker:") {
		t.Errorf("expected 'added blocker:' in output: %q", output)
	}
}

func TestAddCmd_UnknownSection_Errors(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetAddFlags()

	rootCmd.SetArgs([]string{"add", "--section", "diary", "text"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for unknown section, got nil")
	}
}

func TestAddCmd_RejectsUppercaseTag(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetAddFlags()

	rootCmd.SetArgs([]string{"add", "--tag", "Auth", "some note"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for uppercase tag, got nil")
	}
}

func TestAddCmd_RejectsHyphenatedTag(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetAddFlags()

	rootCmd.SetArgs([]string{"add", "--tag", "auth-backend", "some note"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for hyphenated tag, got nil")
	}
}

func TestAddCmd_RejectsEmptyTag(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetAddFlags()

	rootCmd.SetArgs([]string{"add", "--tag", "", "some note"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for empty tag, got nil")
	}
}

func TestAddCmd_DueFlag_ActionItem(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetAddFlags()

	rootCmd.SetArgs([]string{"add", "--section", "action_item", "--due", "2099-12-31", "finish report"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	st, _ := store.New(dir)
	items, _ := st.LoadAllItems()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Due == nil {
		t.Fatal("Due should be set")
	}
	if items[0].Due.String() != "2099-12-31" {
		t.Errorf("Due = %q, want 2099-12-31", items[0].Due.String())
	}
}

func TestAddCmd_DueFlag_OnBlocker_Errors(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetAddFlags()

	rootCmd.SetArgs([]string{"add", "--section", "blocker", "--due", "2099-12-31", "external block"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error: --due on blocker should fail")
	}
}

func TestAddCmd_ETAFlag_Blocker(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetAddFlags()

	rootCmd.SetArgs([]string{"add", "--section", "blocker", "--eta", "2026-09-01", "waiting on vendor"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	st, _ := store.New(dir)
	items, _ := st.LoadAllItems()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ETA == nil {
		t.Fatal("ETA should be set")
	}
	if items[0].ETA.String() != "2026-09-01" {
		t.Errorf("ETA = %q, want 2026-09-01", items[0].ETA.String())
	}
}

func TestAddCmd_ETAFlag_OnActionItem_Errors(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetAddFlags()

	rootCmd.SetArgs([]string{"add", "--section", "action_item", "--eta", "2026-09-01", "write tests"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error: --eta on action_item should fail")
	}
}

func TestAddCmd_DueFlag_InvalidDate_Errors(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetAddFlags()

	rootCmd.SetArgs([]string{"add", "--section", "action_item", "--due", "not-a-date", "task"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for invalid --due date")
	}
}

func TestAddCmd_DueFlag_TodayRelative(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetAddFlags()

	rootCmd.SetArgs([]string{"add", "--section", "action_item", "--due", "today", "due today task"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v; --due today should be accepted", err)
	}

	st, _ := store.New(dir)
	items, _ := st.LoadAllItems()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Due == nil {
		t.Fatal("Due should be set")
	}
	today := store.DateOf(time.Now())
	if items[0].Due.String() != today.String() {
		t.Errorf("Due = %q, want %q (today)", items[0].Due.String(), today.String())
	}
}

func TestAddCmd_DueFlag_OnNotesSection_Errors(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetAddFlags()

	rootCmd.SetArgs([]string{"add", "--due", "2099-12-31", "a note"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error: --due is not valid on notes section")
	}
}

func TestAddCmd_TagFlag(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetAddFlags()

	rootCmd.SetArgs([]string{"add", "--tag", "auth", "--tag", "backend", "tagged note"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	today := time.Now()
	path := filepath.Join(dir, today.Format("2006-01-02")+".md")
	data, _ := os.ReadFile(path)
	entry, _ := store.Parse(data)

	hasAuth := false
	hasBackend := false
	for _, tag := range entry.Tags {
		if tag == "auth" {
			hasAuth = true
		}
		if tag == "backend" {
			hasBackend = true
		}
	}
	if !hasAuth || !hasBackend {
		t.Errorf("Tags = %v, want [auth backend]", entry.Tags)
	}
}
