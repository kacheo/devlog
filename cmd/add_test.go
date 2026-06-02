package cmd

import (
	"os"
	"path/filepath"
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
	addTags = nil
	globalDate = ""

	rootCmd.SetArgs([]string{"add", "--section", "blocker", "stuck on CI"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	today := time.Now()
	path := filepath.Join(dir, today.Format("2006-01-02")+".md")
	data, _ := os.ReadFile(path)
	entry, _ := store.Parse(data)
	blockers := entry.Sections["blockers"]
	if len(blockers) != 1 || blockers[0] != "stuck on CI" {
		t.Errorf("blockers = %v, want [stuck on CI]", blockers)
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
