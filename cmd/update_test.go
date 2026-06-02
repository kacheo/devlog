package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kacheo/devlog/internal/store"
)

func TestUpdateCmd_ReplacesBullet(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	updateSection = "notes"
	updateID = 0
	globalDate = ""

	// Seed a note via add
	addSection = ""
	addTags = nil
	rootCmd.SetArgs([]string{"add", "original text"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add: %v", err)
	}

	// Update bullet #1
	addSection = ""
	addTags = nil
	updateSection = "notes"
	updateID = 1
	rootCmd.SetArgs([]string{"update", "--section", "notes", "--id", "1", "updated text"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("update: %v", err)
	}

	today := time.Now()
	path := filepath.Join(dir, today.Format("2006-01-02")+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("day file not found: %v", err)
	}
	entry, err := store.Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(entry.Sections["notes"]) != 1 || entry.Sections["notes"][0] != "updated text" {
		t.Errorf("notes = %v, want [updated text]", entry.Sections["notes"])
	}
}

func TestUpdateCmd_OutOfRange_Errors(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	updateSection = "notes"
	updateID = 0
	globalDate = ""

	// Seed one note
	addSection = ""
	addTags = nil
	rootCmd.SetArgs([]string{"add", "only note"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add: %v", err)
	}

	// Try to update out-of-range id
	addSection = ""
	addTags = nil
	updateSection = "notes"
	updateID = 99
	rootCmd.SetArgs([]string{"update", "--section", "notes", "--id", "99", "new text"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for out-of-range id, got nil")
	}
}

func TestUpdateCmd_UnknownSection_Errors(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	updateSection = "notes"
	updateID = 1
	globalDate = ""

	rootCmd.SetArgs([]string{"update", "--section", "diary", "--id", "1", "text"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for unknown section, got nil")
	}
}
