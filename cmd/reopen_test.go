package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kacheo/devlog/internal/store"
)

func TestReopenCmd_Basic(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)

	st, _ := store.New(dir)
	item, _ := st.AddItem("blocker", "waiting on infra", []string{})
	if _, err := st.ResolveItem(item.ID); err != nil {
		t.Fatalf("resolve setup: %v", err)
	}

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"reopen", item.ID})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("reopen: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "reopened: blocker") {
		t.Errorf("expected 'reopened: blocker' in output: %q", out)
	}
	if !strings.Contains(out, "waiting on infra") {
		t.Errorf("expected item text in output: %q", out)
	}

	// Verify persisted: unresolved and timestamp cleared.
	items, _ := st.LoadAllItems()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Resolved {
		t.Error("item should be unresolved after reopen")
	}
	if items[0].ResolvedAt != nil {
		t.Error("ResolvedAt should be nil after reopen")
	}
}

func TestReopenCmd_ByShortID(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)

	st, _ := store.New(dir)
	item, _ := st.AddItem("action_item", "write docs", []string{})
	if _, err := st.ResolveItem(item.ID); err != nil {
		t.Fatalf("resolve setup: %v", err)
	}
	shortID := item.ID[:8]

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"reopen", shortID})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("reopen by short ID: %v", err)
	}

	items, _ := st.LoadAllItems()
	if len(items) != 1 || items[0].Resolved || items[0].ResolvedAt != nil {
		t.Error("item should be reopened (unresolved, no ResolvedAt)")
	}
}

func TestReopenCmd_NotFound(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)

	rootCmd.SetArgs([]string{"reopen", "nonexistent"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}
