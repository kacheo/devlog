package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kacheo/devlog/internal/store"
)

func TestRmCmd_RemovesBullet(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	rmSection = "notes"
	rmID = 0
	globalDate = ""

	// Seed two notes via add
	rootCmd.SetArgs([]string{"add", "keep this"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add keep: %v", err)
	}
	addSection = ""
	addTags = nil
	rootCmd.SetArgs([]string{"add", "remove this"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add remove: %v", err)
	}

	// Remove bullet #2
	addSection = ""
	addTags = nil
	rmSection = "notes"
	rmID = 2
	rootCmd.SetArgs([]string{"rm", "--section", "notes", "--id", "2"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rm: %v", err)
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
	if len(entry.Sections["notes"]) != 1 || entry.Sections["notes"][0] != "keep this" {
		t.Errorf("notes = %v, want [keep this]", entry.Sections["notes"])
	}
}

func TestRmCmd_OutOfRange_Errors(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	rmSection = "notes"
	rmID = 0
	globalDate = ""

	// Seed one note
	addSection = ""
	addTags = nil
	rootCmd.SetArgs([]string{"add", "only note"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add: %v", err)
	}

	// Try to remove out-of-range id
	addSection = ""
	addTags = nil
	rmSection = "notes"
	rmID = 5
	rootCmd.SetArgs([]string{"rm", "--section", "notes", "--id", "5"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for out-of-range id, got nil")
	}
}

func TestRmCmd_UnknownSection_Errors(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	rmSection = "notes"
	rmID = 1
	globalDate = ""

	rootCmd.SetArgs([]string{"rm", "--section", "diary", "--id", "1"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for unknown section, got nil")
	}
}
