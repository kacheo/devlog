package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/kacheo/devlog/internal/store"
)

func resetTagsFlags() {
	// no persistent flags on tags subcommands yet
}

func makeTagEntry(t *testing.T, st *store.Store, date time.Time, tags []string) {
	t.Helper()
	e := store.EmptyEntry(date)
	e.Tags = tags
	if err := st.Save(e); err != nil {
		t.Fatalf("Save: %v", err)
	}
}

func TestTagsListCmd_Empty(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetTagsFlags()

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"tags", "list"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("tags list: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "(no tags)") {
		t.Errorf("expected '(no tags)' in output: %q", out)
	}
}

func TestTagsListCmd_ShowsTagsWithCounts(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetTagsFlags()

	st, _ := store.New(dir)
	makeTagEntry(t, st, time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local), []string{"auth", "backend"})
	makeTagEntry(t, st, time.Date(2026, 1, 2, 0, 0, 0, 0, time.Local), []string{"auth"})

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"tags", "list"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("tags list: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "auth") {
		t.Errorf("expected 'auth' in output: %q", out)
	}
	if !strings.Contains(out, "2") {
		t.Errorf("expected count '2' for auth in output: %q", out)
	}
	if !strings.Contains(out, "backend") {
		t.Errorf("expected 'backend' in output: %q", out)
	}
}

// Invoking `devlog tags` (no subcommand) should default to list behavior.
func TestTagsCmd_DefaultIsList(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetTagsFlags()

	st, _ := store.New(dir)
	makeTagEntry(t, st, time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local), []string{"infra"})

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"tags"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("tags: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "infra") {
		t.Errorf("expected 'infra' in default tags output: %q", out)
	}
}

func TestTagsListCmd_JSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetTagsFlags()
	globalJSON = false

	st, _ := store.New(dir)
	makeTagEntry(t, st, time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local), []string{"auth", "backend"})
	makeTagEntry(t, st, time.Date(2026, 1, 2, 0, 0, 0, 0, time.Local), []string{"auth"})

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"tags", "list", "--json"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("tags list --json: %v", err)
	}
	globalJSON = false

	var out map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &out); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", err, buf.String())
	}
	if out["version"] != "1" {
		t.Errorf("version = %v, want '1'", out["version"])
	}
	tagsList, ok := out["tags"].([]interface{})
	if !ok {
		t.Fatalf("tags field missing or wrong type: %v", out["tags"])
	}
	if len(tagsList) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tagsList))
	}
	first := tagsList[0].(map[string]interface{})
	if first["tag"] != "auth" {
		t.Errorf("first tag = %v, want 'auth'", first["tag"])
	}
	if first["count"] != float64(2) {
		t.Errorf("first count = %v, want 2", first["count"])
	}
}

func TestTagsRenameCmd_Basic(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetTagsFlags()

	st, _ := store.New(dir)
	makeTagEntry(t, st, time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local), []string{"auth"})
	makeTagEntry(t, st, time.Date(2026, 1, 2, 0, 0, 0, 0, time.Local), []string{"auth", "backend"})

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"tags", "rename", "auth", "authentication"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("tags rename: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "2") {
		t.Errorf("expected file count '2' in output: %q", out)
	}
	if !strings.Contains(out, "authentication") {
		t.Errorf("expected new tag name in output: %q", out)
	}

	// Verify data on disk
	e1, _ := st.Load(time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))
	found := false
	for _, tag := range e1.Tags {
		if tag == "authentication" {
			found = true
		}
	}
	if !found {
		t.Errorf("d1 missing 'authentication' tag: %v", e1.Tags)
	}
}

func TestTagsRenameCmd_NonExistent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetTagsFlags()

	st, _ := store.New(dir)
	makeTagEntry(t, st, time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local), []string{"backend"})

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"tags", "rename", "auth", "authentication"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("tags rename: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "No entries") {
		t.Errorf("expected 'No entries' message: %q", out)
	}
}

func TestTagsRenameCmd_SameTag(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetTagsFlags()

	rootCmd.SetArgs([]string{"tags", "rename", "auth", "auth"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error renaming tag to itself")
	}
}

func TestTagsRenameCmd_RequiresTwoArgs(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetTagsFlags()

	rootCmd.SetArgs([]string{"tags", "rename", "only-one-arg"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error with wrong argument count")
	}
}

func TestTagsRenameCmd_DeduplicatesWhenNewTagAlreadyPresent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetTagsFlags()

	st, _ := store.New(dir)
	// Entry has both auth and oauth; renaming auth→oauth should yield just [oauth].
	makeTagEntry(t, st, time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local), []string{"auth", "oauth"})

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"tags", "rename", "auth", "oauth"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("tags rename: %v", err)
	}

	e, _ := st.Load(time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))
	if len(e.Tags) != 1 || e.Tags[0] != "oauth" {
		t.Errorf("tags after rename = %v, want [oauth]", e.Tags)
	}
}

func TestTagsRenameCmd_RejectsInvalidNewTag(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetTagsFlags()

	st, _ := store.New(dir)
	makeTagEntry(t, st, time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local), []string{"auth"})

	for _, badTag := range []string{"Auth", "new-tag", "New Tag", ""} {
		rootCmd.SetArgs([]string{"tags", "rename", "auth", badTag})
		err := rootCmd.Execute()
		if err == nil {
			t.Errorf("expected error for invalid new tag %q, got nil", badTag)
		}
	}
}

func TestTagsRenameCmd_CaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetTagsFlags()

	st, _ := store.New(dir)
	makeTagEntry(t, st, time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local), []string{"Auth"})

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"tags", "rename", "auth", "authentication"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("tags rename: %v", err)
	}

	e, _ := st.Load(time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))
	if len(e.Tags) != 1 || e.Tags[0] != "authentication" {
		t.Errorf("tags = %v, want [authentication]", e.Tags)
	}
}
