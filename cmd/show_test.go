package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/kacheo/devlog/internal/store"
)

func TestShowCmd_Today_NoFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalDate = ""
	globalJSON = false

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"show", "today"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "no entry") {
		t.Errorf("expected 'no entry' for missing day, got: %q", out)
	}
}

func TestShowCmd_DateAndPositionalMutuallyExclusive(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalDate = "2026-06-01"

	rootCmd.SetArgs([]string{"show", "today"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error when both --date and positional arg given")
	}
	globalDate = ""
}

func TestShowCmd_JSON_SingleDay(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalDate = ""
	globalJSON = true
	t.Cleanup(func() { globalJSON = false })

	// Create a day file
	st, _ := store.New(dir)
	date := time.Now()
	entry := store.EmptyEntry(date)
	entry.Sections["notes"] = []string{"test note"}
	if err := st.Save(entry); err != nil {
		t.Fatal(err)
	}

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"show", "today"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var v map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, buf.String())
	}
	if v["version"] != "1" {
		t.Errorf("version = %v, want '1'", v["version"])
	}
}

func TestShowCmd_Week_NoFiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalDate = ""
	globalJSON = false

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"show", "week"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	// Should not error — just print nothing or "(none)"
}
