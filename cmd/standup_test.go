package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/kacheo/devlog/internal/store"
)

func TestStandupCmd_Terminal_ContainsDoneAndBlockers(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalJSON = false
	standupSince = ""

	// Create yesterday's entry with a commit
	st, _ := store.New(dir)
	yesterday := time.Now().AddDate(0, 0, -1)
	yEntry := store.EmptyEntry(yesterday)
	yEntry.Commits = []store.Commit{{SHA: "abc1234", Message: "fix: auth", Repo: "api"}}
	if err := st.Save(yEntry); err != nil {
		t.Fatal(err)
	}

	// Create today's entry with a blocker
	today := time.Now()
	tEntry := store.EmptyEntry(today)
	tEntry.Sections["blockers"] = []string{"blocked on review"}
	if err := st.Save(tEntry); err != nil {
		t.Fatal(err)
	}

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"standup"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Done") {
		t.Errorf("output missing 'Done':\n%s", out)
	}
	if !strings.Contains(out, "abc1234") {
		t.Errorf("output missing commit SHA:\n%s", out)
	}
	if !strings.Contains(out, "blocked on review") {
		t.Errorf("output missing blocker:\n%s", out)
	}
}

func TestStandupCmd_JSON_Schema(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	globalJSON = true
	standupSince = ""
	t.Cleanup(func() { globalJSON = false })

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"standup"})
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
	if v["generated_at"] == nil || v["generated_at"] == "" {
		t.Error("generated_at is missing or empty")
	}
	period, ok := v["period"].(map[string]interface{})
	if !ok {
		t.Fatal("period missing from output")
	}
	if period["since"] == "" || period["until"] == "" {
		t.Error("period.since or period.until is empty")
	}
	if _, ok := v["done"]; !ok {
		t.Error("done field missing")
	}
	if _, ok := v["blockers"]; !ok {
		t.Error("blockers field missing")
	}
}
