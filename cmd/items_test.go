package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/kacheo/devlog/internal/store"
)

func resetItemsFlags() {
	itemsResolved = false
	itemsAll = false
	itemsOverdue = false
	itemsType = ""
	itemsFrom = ""
	itemsUntil = ""
}

func TestItemsCmd_Default_ShowsUnresolved(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetItemsFlags()

	st, _ := store.New(dir)
	_, _ = st.AddItem("blocker", "open blocker", []string{})
	item2, _ := st.AddItem("action_item", "done task", []string{})
	_, _ = st.ResolveItem(item2.ID)

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"items"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("items: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "open blocker") {
		t.Errorf("expected unresolved item in output: %q", out)
	}
	if strings.Contains(out, "done task") {
		t.Errorf("resolved item should not appear by default: %q", out)
	}
}

func TestItemsCmd_Resolved(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetItemsFlags()

	st, _ := store.New(dir)
	_, _ = st.AddItem("blocker", "open blocker", []string{})
	item2, _ := st.AddItem("action_item", "done task", []string{})
	_, _ = st.ResolveItem(item2.ID)

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"items", "--resolved"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("items --resolved: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "done task") {
		t.Errorf("expected resolved item in output: %q", out)
	}
	if strings.Contains(out, "open blocker") {
		t.Errorf("unresolved item should not appear with --resolved: %q", out)
	}
}

func TestItemsCmd_All(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetItemsFlags()

	st, _ := store.New(dir)
	_, _ = st.AddItem("blocker", "still open", []string{})
	item2, _ := st.AddItem("action_item", "already done", []string{})
	_, _ = st.ResolveItem(item2.ID)

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"items", "--all"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("items --all: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "still open") {
		t.Errorf("expected unresolved item in --all output: %q", out)
	}
	if !strings.Contains(out, "already done") {
		t.Errorf("expected resolved item in --all output: %q", out)
	}
}

func TestItemsCmd_TypeFilter(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetItemsFlags()

	st, _ := store.New(dir)
	_, _ = st.AddItem("blocker", "the blocker", []string{})
	_, _ = st.AddItem("action_item", "the task", []string{})

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"items", "--type", "blockers"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("items --type blockers: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "the blocker") {
		t.Errorf("expected blocker in output: %q", out)
	}
	if strings.Contains(out, "the task") {
		t.Errorf("action_item should not appear with --type blockers: %q", out)
	}
}

func TestItemsCmd_JSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetItemsFlags()
	globalJSON = false

	st, _ := store.New(dir)
	item, _ := st.AddItem("blocker", "json blocker", []string{})
	_, _ = st.ResolveItem(item.ID)

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"items", "--resolved", "--json"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("items --resolved --json: %v", err)
	}

	globalJSON = false // reset after test

	var out []map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &out); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", err, buf.String())
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 item, got %d", len(out))
	}
	if out[0]["text"] != "json blocker" {
		t.Errorf("text = %v", out[0]["text"])
	}
	if out[0]["resolved"] != true {
		t.Errorf("resolved = %v, want true", out[0]["resolved"])
	}
	if _, ok := out[0]["resolved_at"]; !ok {
		t.Error("expected resolved_at field in JSON output")
	}
}

func TestItemsCmd_ResolvedAndAllMutuallyExclusive(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetItemsFlags()

	rootCmd.SetArgs([]string{"items", "--resolved", "--all"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for --resolved --all")
	}
}

func TestItemsCmd_InvalidType(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetItemsFlags()

	rootCmd.SetArgs([]string{"items", "--type", "diary"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for unknown --type")
	}
}

func TestItemsCmd_Overdue(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetItemsFlags()

	st, _ := store.New(dir)
	past := store.DateOf(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	future := store.DateOf(time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC))
	_, _ = st.AddItem("action_item", "overdue task", []string{}, store.ItemOptions{Due: &past})
	_, _ = st.AddItem("action_item", "future task", []string{}, store.ItemOptions{Due: &future})
	_, _ = st.AddItem("blocker", "plain blocker", []string{})

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"items", "--overdue"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "overdue task") {
		t.Errorf("expected overdue task in output: %q", out)
	}
	if strings.Contains(out, "future task") {
		t.Errorf("future task should not appear in overdue list: %q", out)
	}
	if strings.Contains(out, "plain blocker") {
		t.Errorf("blocker should not appear in overdue list: %q", out)
	}
}

func TestItemsCmd_OverdueAndResolvedMutuallyExclusive(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)
	resetItemsFlags()

	rootCmd.SetArgs([]string{"items", "--overdue", "--resolved"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for --overdue --resolved")
	}
}
