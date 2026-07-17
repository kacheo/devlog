package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/kacheo/devlog/internal/store"
)

func resetSearchFlags() {
	searchSections = nil
	searchFrom = ""
	searchUntil = ""
}

func TestSearchCmd_BasicMatch(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetSearchFlags()

	st, _ := store.New(dir)
	entry := store.EmptyEntry(time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local))
	entry.Sections["notes"] = []string{"worked on oauth flow"}
	if err := st.Save(entry); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"search", "oauth"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "oauth") {
		t.Errorf("expected 'oauth' in output, got: %q", out)
	}
	if !strings.Contains(out, "notes") {
		t.Errorf("expected section 'notes' in output, got: %q", out)
	}
}

func TestSearchCmd_NoMatches_ExitsZero(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetSearchFlags()

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"search", "nonexistentquery12345"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected exit 0 when no matches, got error: %v", err)
	}
	if out := buf.String(); out != "" {
		t.Errorf("expected empty stdout on no match, got: %q", out)
	}
}

func TestSearchCmd_NoMatches_JSON_EmptyArray(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetSearchFlags()
	globalJSON = true
	t.Cleanup(func() { globalJSON = false })

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"search", "--json", "nonexistentquery12345"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected exit 0 when no matches, got error: %v", err)
	}

	if got := strings.TrimSpace(buf.String()); got != "[]" {
		t.Errorf("expected empty JSON array %q, got: %q", "[]", got)
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("output is not valid JSON array: %v\nOutput: %s", err, buf.String())
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearchCmd_JSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetSearchFlags()
	globalJSON = true
	t.Cleanup(func() { globalJSON = false })

	st, _ := store.New(dir)
	entry := store.EmptyEntry(time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local))
	entry.Sections["notes"] = []string{"oauth flow"}
	if err := st.Save(entry); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"search", "--json", "oauth"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("output is not valid JSON array: %v\nOutput: %s", err, buf.String())
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["section"] != "notes" {
		t.Errorf("section = %v, want 'notes'", results[0]["section"])
	}
}

func TestSearchCmd_SectionFilter(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetSearchFlags()

	st, _ := store.New(dir)
	entry := store.EmptyEntry(time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local))
	entry.Sections["notes"] = []string{"oauth in notes"}
	entry.Sections["blockers"] = []string{"oauth blocker"}
	if err := st.Save(entry); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"search", "--section", "notes", "oauth"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "blockers") {
		t.Errorf("blockers section should not appear when --section notes, got: %q", out)
	}
	if !strings.Contains(out, "notes") {
		t.Errorf("expected notes section in output, got: %q", out)
	}
}

func TestSearchCmd_UnknownSection_Error(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetSearchFlags()

	rootCmd.SetArgs([]string{"search", "--section", "diary", "oauth"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for unknown section name, got nil")
	}
}

func TestSearchCmd_DateRange(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetSearchFlags()

	st, _ := store.New(dir)
	for _, d := range []time.Time{
		time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local),
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local),
	} {
		e := store.EmptyEntry(d)
		e.Sections["notes"] = []string{"oauth work"}
		if err := st.Save(e); err != nil {
			t.Fatal(err)
		}
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"search", "--from", "2026-06-01", "--until", "2026-06-01", "oauth"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "2026-05-01") {
		t.Errorf("May result should be filtered out, got: %q", out)
	}
	if !strings.Contains(out, "2026-06-01") {
		t.Errorf("expected June 1 result, got: %q", out)
	}
}
