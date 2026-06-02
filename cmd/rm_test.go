package cmd_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kacheo/devlog/internal/store"
)

func TestRmBullet(t *testing.T) {
	dir := t.TempDir()
	st, _ := store.New(dir)
	date := time.Date(2026, 1, 5, 0, 0, 0, 0, time.Local)

	if err := st.Modify(date, func(e *store.DayEntry) error {
		e.Sections["notes"] = []string{"keep this", "remove this", "also keep"}
		return nil
	}); err != nil {
		t.Fatalf("setup Modify: %v", err)
	}

	// simulate rm --section notes --id 2
	err := st.Modify(date, func(e *store.DayEntry) error {
		bullets := e.Sections["notes"]
		id := 2
		if id < 1 || id > len(bullets) {
			return fmt.Errorf("id %d out of range: section %q has %d item(s)", id, "notes", len(bullets))
		}
		e.Sections["notes"] = append(bullets[:id-1], bullets[id:]...)
		return nil
	})
	if err != nil {
		t.Fatalf("rm: %v", err)
	}

	entry, _ := st.Load(date)
	if len(entry.Sections["notes"]) != 2 {
		t.Fatalf("expected 2 notes after rm, got %v", entry.Sections["notes"])
	}
	if entry.Sections["notes"][0] != "keep this" || entry.Sections["notes"][1] != "also keep" {
		t.Fatalf("unexpected notes: %v", entry.Sections["notes"])
	}
}

func TestRmBullet_OutOfRange(t *testing.T) {
	dir := t.TempDir()
	st, _ := store.New(dir)
	date := time.Date(2026, 1, 6, 0, 0, 0, 0, time.Local)

	if err := st.Modify(date, func(e *store.DayEntry) error {
		e.Sections["notes"] = []string{"only one"}
		return nil
	}); err != nil {
		t.Fatalf("setup Modify: %v", err)
	}

	err := st.Modify(date, func(e *store.DayEntry) error {
		bullets := e.Sections["notes"]
		id := 5
		if id < 1 || id > len(bullets) {
			return fmt.Errorf("id %d out of range: section %q has %d item(s)", id, "notes", len(bullets))
		}
		e.Sections["notes"] = append(bullets[:id-1], bullets[id:]...)
		return nil
	})
	if err == nil {
		t.Fatal("expected error for out-of-range id")
	}
}
