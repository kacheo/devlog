package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kacheo/devlog/internal/store"
)

func TestResolveCmd_Basic(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)

	// Add a blocker via the store directly
	st, _ := store.New(dir)
	item, _ := st.AddItem("blocker", "waiting on infra", []string{})

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"resolve", item.ID})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "resolved: blocker") {
		t.Errorf("expected 'resolved: blocker' in output: %q", out)
	}
	if !strings.Contains(out, "waiting on infra") {
		t.Errorf("expected item text in output: %q", out)
	}

	// Verify persisted
	items, _ := st.LoadAllItems()
	if len(items) != 1 || !items[0].Resolved {
		t.Error("item should be marked resolved in store")
	}
}

func TestResolveCmd_ByShortID(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)

	st, _ := store.New(dir)
	item, _ := st.AddItem("action_item", "write docs", []string{})
	shortID := item.ID[:8]

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"resolve", shortID})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("resolve by short ID: %v", err)
	}

	items, _ := st.LoadAllItems()
	if len(items) != 1 || !items[0].Resolved {
		t.Error("item should be resolved")
	}
}

func TestResolveCmd_NotFound(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLOG_DIR", dir)

	rootCmd.SetArgs([]string{"resolve", "nonexistent"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}
