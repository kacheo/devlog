package cmd_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kacheo/devlog/internal/store"
)

func TestUpdateBullet(t *testing.T) {
	dir := t.TempDir()
	st, _ := store.New(dir)
	date := time.Date(2026, 1, 7, 0, 0, 0, 0, time.Local)

	if err := st.Modify(date, func(e *store.DayEntry) error {
		e.Sections["notes"] = []string{"original text", "other note"}
		return nil
	}); err != nil {
		t.Fatalf("setup Modify: %v", err)
	}

	err := st.Modify(date, func(e *store.DayEntry) error {
		bullets := e.Sections["notes"]
		id := 1
		if id < 1 || id > len(bullets) {
			return fmt.Errorf("id %d out of range: %d item(s)", id, len(bullets))
		}
		e.Sections["notes"][id-1] = "updated text"
		return nil
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	entry, _ := st.Load(date)
	if entry.Sections["notes"][0] != "updated text" {
		t.Fatalf("expected 'updated text', got %q", entry.Sections["notes"][0])
	}
	if entry.Sections["notes"][1] != "other note" {
		t.Fatalf("expected 'other note' unchanged, got %q", entry.Sections["notes"][1])
	}
}

func TestUpdateBullet_OutOfRange(t *testing.T) {
	dir := t.TempDir()
	st, _ := store.New(dir)
	date := time.Date(2026, 1, 8, 0, 0, 0, 0, time.Local)

	if err := st.Modify(date, func(e *store.DayEntry) error {
		e.Sections["notes"] = []string{"only one"}
		return nil
	}); err != nil {
		t.Fatalf("setup Modify: %v", err)
	}

	err := st.Modify(date, func(e *store.DayEntry) error {
		bullets := e.Sections["notes"]
		id := 99
		if id < 1 || id > len(bullets) {
			return fmt.Errorf("id %d out of range: section %q has %d item(s)", id, "notes", len(bullets))
		}
		e.Sections["notes"][id-1] = "new text"
		return nil
	})
	if err == nil {
		t.Fatal("expected error for out-of-range id")
	}
}
