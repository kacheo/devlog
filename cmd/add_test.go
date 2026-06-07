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

func TestAddCmd_AppendsToNotes(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	// Reset flags between tests
	addSection = ""
	addTags = nil
	globalDate = ""

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
	addSection = "blocker"
	addDeps = nil
	addTags = nil
	globalDate = ""

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
	addSection = ""
	addTags = nil
	globalDate = ""

	rootCmd.SetArgs([]string{"add", "--section", "diary", "text"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for unknown section, got nil")
	}
}

func TestAddCmd_RejectsUppercaseTag(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	addSection = ""
	addTags = nil
	globalDate = ""

	rootCmd.SetArgs([]string{"add", "--tag", "Auth", "some note"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for uppercase tag, got nil")
	}
}

func TestAddCmd_RejectsHyphenatedTag(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	addSection = ""
	addTags = nil
	globalDate = ""

	rootCmd.SetArgs([]string{"add", "--tag", "auth-backend", "some note"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for hyphenated tag, got nil")
	}
}

func TestAddCmd_RejectsEmptyTag(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	addSection = ""
	addTags = nil
	globalDate = ""

	rootCmd.SetArgs([]string{"add", "--tag", "", "some note"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for empty tag, got nil")
	}
}

func TestAddCmd_TagFlag(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	addSection = ""
	addTags = nil
	globalDate = ""

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
