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

func TestShowTerminal_SectionOrder(t *testing.T) {
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	e := store.EmptyEntry(date)
	e.Sections["notes"] = []string{"a note"}
	e.Sections["action_items"] = []string{"an action"}
	e.Sections["blockers"] = []string{"a blocker"}
	e.Commits = []store.Commit{{SHA: "abc", Message: "fix", Repo: "r"}}
	e.PRs = []store.PR{{Number: 1, Title: "T", State: "open", Repo: "r"}}

	var buf bytes.Buffer
	ShowTerminal(e, &buf)
	out := buf.String()

	pos := func(section string) int { return strings.Index(out, section+":") }
	for _, s := range []string{"Notes", "Action Items", "Commits", "PRs", "Blockers"} {
		if pos(s) == -1 {
			t.Errorf("section %q missing from output:\n%s", s, out)
		}
	}
	if pos("Action Items") > pos("Commits") {
		t.Errorf("Action Items should appear before Commits:\n%s", out)
	}
	if pos("Commits") > pos("PRs") {
		t.Errorf("Commits should appear before PRs:\n%s", out)
	}
	if pos("PRs") > pos("Blockers") {
		t.Errorf("PRs should appear before Blockers:\n%s", out)
	}
}

func TestShowTerminal_ActionItems(t *testing.T) {
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	e := store.EmptyEntry(date)
	e.Sections["action_items"] = []string{"do the thing", "another task"}

	var buf bytes.Buffer
	ShowTerminal(e, &buf)
	out := buf.String()

	if !strings.Contains(out, "Action Items:") {
		t.Errorf("missing 'Action Items:' header:\n%s", out)
	}
	if !strings.Contains(out, "[1] do the thing") {
		t.Errorf("missing '[1] do the thing':\n%s", out)
	}
	if !strings.Contains(out, "[2] another task") {
		t.Errorf("missing '[2] another task':\n%s", out)
	}
}

func TestStandupTerminal_NilTodayEntry(t *testing.T) {
	since := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	until := time.Date(2026, 6, 2, 0, 0, 0, 0, time.Local)

	var buf bytes.Buffer
	StandupTerminal(since, until, nil, nil, &buf)
	out := buf.String()

	if !strings.Contains(out, "(none)") {
		t.Errorf("expected '(none)' placeholder with nil entries:\n%s", out)
	}
	if !strings.Contains(out, "Blockers:") {
		t.Errorf("missing Blockers section:\n%s", out)
	}
}
