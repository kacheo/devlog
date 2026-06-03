package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func resetShowFlags() {
	showTag = ""
	showRepo = ""
	showSections = nil
	showFrom = ""
	showUntil = ""
}

func TestShowCmd_TagFilter_Week(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalDate = ""
	globalJSON = false
	resetShowFlags()

	st, _ := store.New(dir)

	tagged := store.EmptyEntry(time.Now())
	tagged.Tags = []string{"auth"}
	tagged.Sections["notes"] = []string{"auth work today"}
	if err := st.Save(tagged); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	showTag = "auth"
	rootCmd.SetArgs([]string{"show", "week"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	resetShowFlags()

	out := buf.String()
	if !strings.Contains(out, "auth work today") {
		t.Errorf("expected tagged entry in output, got: %q", out)
	}
}

func TestShowCmd_TagFilter_ExcludesUntagged(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalDate = ""
	globalJSON = false
	resetShowFlags()

	st, _ := store.New(dir)

	today := store.EmptyEntry(time.Now())
	today.Tags = []string{"backend"}
	today.Sections["notes"] = []string{"backend only"}
	if err := st.Save(today); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	showTag = "auth"
	rootCmd.SetArgs([]string{"show", "week"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	resetShowFlags()

	out := buf.String()
	if strings.Contains(out, "backend only") {
		t.Errorf("untagged entry should be excluded, got: %q", out)
	}
}

func TestShowCmd_RepoFilter_Week(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalDate = ""
	globalJSON = false
	resetShowFlags()

	st, _ := store.New(dir)

	entry := store.EmptyEntry(time.Now())
	entry.Commits = []store.Commit{{SHA: "abc1234", Message: "fix: bug", Repo: "myapp"}}
	entry.Sections["notes"] = []string{"myapp work"}
	if err := st.Save(entry); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	showRepo = "myapp"
	rootCmd.SetArgs([]string{"show", "week"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	resetShowFlags()

	out := buf.String()
	if !strings.Contains(out, "myapp work") {
		t.Errorf("expected repo-matching entry in output, got: %q", out)
	}
}

func TestShowCmd_SectionFilter_Today(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalDate = ""
	globalJSON = false
	resetShowFlags()

	st, _ := store.New(dir)

	entry := store.EmptyEntry(time.Now())
	entry.Sections["notes"] = []string{"a note"}
	entry.Sections["blockers"] = []string{"a blocker"}
	if err := st.Save(entry); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	showSections = []string{"notes"}
	rootCmd.SetArgs([]string{"show", "today"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	resetShowFlags()

	out := buf.String()
	if !strings.Contains(out, "a note") {
		t.Errorf("expected notes section in output, got: %q", out)
	}
	if strings.Contains(out, "a blocker") {
		t.Errorf("blockers should be filtered out, got: %q", out)
	}
}

func TestShowCmd_From_MultiDay(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalDate = ""
	globalJSON = false
	resetShowFlags()
	t.Cleanup(resetShowFlags)

	st, _ := store.New(dir)
	yesterday := time.Now().AddDate(0, 0, -1)
	yEntry := store.EmptyEntry(yesterday)
	yEntry.Sections["notes"] = []string{"yesterday note"}
	if err := st.Save(yEntry); err != nil {
		t.Fatal(err)
	}
	today := time.Now()
	tEntry := store.EmptyEntry(today)
	tEntry.Sections["notes"] = []string{"today note"}
	if err := st.Save(tEntry); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	showFrom = yesterday.Format("2006-01-02")
	rootCmd.SetArgs([]string{"show"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "yesterday note") {
		t.Errorf("output missing yesterday entry:\n%s", out)
	}
	if !strings.Contains(out, "today note") {
		t.Errorf("output missing today entry:\n%s", out)
	}
}

func TestShowCmd_FromUntil_Range(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalDate = ""
	globalJSON = false
	resetShowFlags()
	t.Cleanup(resetShowFlags)

	st, _ := store.New(dir)
	base := time.Now().AddDate(0, 0, -5)
	for i := 0; i <= 4; i++ {
		d := base.AddDate(0, 0, i)
		e := store.EmptyEntry(d)
		e.Sections["notes"] = []string{fmt.Sprintf("note day %d", i)}
		if err := st.Save(e); err != nil {
			t.Fatal(err)
		}
	}

	// Range covers only days 1-2 (not day 0 or days 3-4)
	from := base.AddDate(0, 0, 1)
	until := base.AddDate(0, 0, 2)

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	showFrom = from.Format("2006-01-02")
	showUntil = until.Format("2006-01-02")
	rootCmd.SetArgs([]string{"show"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "note day 1") {
		t.Errorf("output missing day 1:\n%s", out)
	}
	if !strings.Contains(out, "note day 2") {
		t.Errorf("output missing day 2:\n%s", out)
	}
	if strings.Contains(out, "note day 0") {
		t.Errorf("output should not include day 0:\n%s", out)
	}
	if strings.Contains(out, "note day 3") {
		t.Errorf("output should not include day 3:\n%s", out)
	}
}

func TestShowCmd_From_JSON_Array(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalDate = ""
	globalJSON = true
	resetShowFlags()
	t.Cleanup(func() {
		globalJSON = false
		resetShowFlags()
	})

	st, _ := store.New(dir)
	yesterday := time.Now().AddDate(0, 0, -1)
	e := store.EmptyEntry(yesterday)
	e.Sections["notes"] = []string{"test"}
	if err := st.Save(e); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	showFrom = yesterday.Format("2006-01-02")
	rootCmd.SetArgs([]string{"show"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var v []interface{}
	if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
		t.Fatalf("output is not a JSON array: %v\nOutput: %s", err, buf.String())
	}
	if len(v) == 0 {
		t.Error("expected at least one entry in JSON array")
	}
}

func TestShowCmd_Until_WithoutFrom_Error(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalDate = ""
	globalJSON = false
	resetShowFlags()
	t.Cleanup(resetShowFlags)

	showUntil = "2026-06-01"
	rootCmd.SetArgs([]string{"show"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error when --until given without --from")
	}
}

func TestShowCmd_From_WithPositionalArg_Error(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalDate = ""
	globalJSON = false
	resetShowFlags()
	t.Cleanup(resetShowFlags)

	showFrom = "2026-06-01"
	rootCmd.SetArgs([]string{"show", "week"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error when --from and positional arg both given")
	}
}

func TestShowCmd_From_WithDate_Error(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalDate = "2026-06-01"
	globalJSON = false
	resetShowFlags()
	t.Cleanup(func() {
		globalDate = ""
		resetShowFlags()
	})

	showFrom = "2026-05-28"
	rootCmd.SetArgs([]string{"show"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error when --from and --date both given")
	}
}

func TestShowCmd_UnknownSection_Error(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalDate = ""
	globalJSON = false
	resetShowFlags()

	showSections = []string{"diary"}
	rootCmd.SetArgs([]string{"show", "today"})
	err := rootCmd.Execute()
	resetShowFlags()
	if err == nil {
		t.Error("expected error for unknown section name, got nil")
	}
}
