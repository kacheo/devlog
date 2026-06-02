package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEditCmd_DateAndPositionalMutuallyExclusive(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalDate = "2026-06-01"

	rootCmd.SetArgs([]string{"edit", "2026-06-02"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error when --date and positional date both given")
	}
	globalDate = ""
}

func TestEditCmd_CreatesFileIfMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalDate = ""

	// Use a no-op editor that doesn't block: "true" on Unix
	t.Setenv("DEVLOG_EDITOR", "true")
	today := time.Now()
	expectedPath := filepath.Join(dir, today.Format("2006-01-02")+".md")

	// File should not exist yet
	if _, err := os.Stat(expectedPath); err == nil {
		_ = os.Remove(expectedPath)
	}

	rootCmd.SetArgs([]string{"edit"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("day file was not created before opening editor")
	}
}
