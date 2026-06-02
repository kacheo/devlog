package render

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/kacheo/devlog/internal/store"
)

func TestShowTerminal_ContainsEntries(t *testing.T) {
	entry := makeEntry("2026-06-01")
	var buf bytes.Buffer
	ShowTerminal(entry, &buf)
	out := buf.String()
	if !strings.Contains(out, "abc1234") {
		t.Errorf("output missing commit SHA:\n%s", out)
	}
	if !strings.Contains(out, "Fix auth") {
		t.Errorf("output missing PR title:\n%s", out)
	}
	if !strings.Contains(out, "Did some work") {
		t.Errorf("output missing note:\n%s", out)
	}
}

func TestStandupTerminal_ContainsSection(t *testing.T) {
	since := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	until := time.Date(2026, 6, 2, 0, 0, 0, 0, time.Local)
	done := makeEntry("2026-06-01")
	today := makeEntry("2026-06-02")

	var buf bytes.Buffer
	StandupTerminal(since, until, []*store.DayEntry{done}, today, &buf)
	out := buf.String()
	if !strings.Contains(out, "Done") {
		t.Errorf("output missing 'Done' section:\n%s", out)
	}
	if !strings.Contains(out, "Blocker") {
		t.Errorf("output missing 'Blocker' section:\n%s", out)
	}
	if !strings.Contains(out, "abc1234") {
		t.Errorf("output missing commit SHA:\n%s", out)
	}
}
