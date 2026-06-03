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
	ShowTerminal(entry, nil, nil, &buf)
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

func TestShowTerminal_SectionOrder(t *testing.T) {
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	e := store.EmptyEntry(date)
	e.Sections["notes"] = []string{"a note"}
	e.Commits = []store.Commit{{SHA: "abc", Message: "fix", Repo: "r"}}
	e.PRs = []store.PR{{Number: 1, Title: "T", State: "open", Repo: "r"}}

	blockers := []store.Item{{ID: "aaaabbbb-0000-4000-8000-ccccddddeeee", Type: "blocker", Text: "a blocker", Dependencies: []string{}}}
	actionItems := []store.Item{{ID: "11112222-3333-4444-5555-666677778888", Type: "action_item", Text: "an action", Dependencies: []string{}}}

	var buf bytes.Buffer
	ShowTerminal(e, blockers, actionItems, &buf)
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

	actionItems := []store.Item{
		{ID: "aabbccdd-0000-4000-8000-112233445566", Type: "action_item", Text: "do the thing", Dependencies: []string{}},
		{ID: "eeff0011-2233-4455-6677-889900aabbcc", Type: "action_item", Text: "another task", Dependencies: []string{}},
	}

	var buf bytes.Buffer
	ShowTerminal(e, nil, actionItems, &buf)
	out := buf.String()

	if !strings.Contains(out, "Action Items:") {
		t.Errorf("missing 'Action Items:' header:\n%s", out)
	}
	if !strings.Contains(out, "do the thing") {
		t.Errorf("missing 'do the thing':\n%s", out)
	}
	if !strings.Contains(out, "another task") {
		t.Errorf("missing 'another task':\n%s", out)
	}
}

func TestShowTerminal_ItemsShowShortID(t *testing.T) {
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	e := store.EmptyEntry(date)

	blockers := []store.Item{
		{ID: "deadbeef-cafe-4000-8000-112233445566", Type: "blocker", Text: "waiting on infra", Dependencies: []string{}},
	}

	var buf bytes.Buffer
	ShowTerminal(e, blockers, nil, &buf)
	out := buf.String()

	if !strings.Contains(out, "[deadbeef]") {
		t.Errorf("expected short ID [deadbeef] in output:\n%s", out)
	}
}

func TestShowTerminal_ItemsShowDeps(t *testing.T) {
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	e := store.EmptyEntry(date)

	depID := "aaaabbbb-0000-4000-8000-ccccddddeeee"
	actionItems := []store.Item{
		{ID: "11112222-3333-4444-5555-666677778888", Type: "action_item", Text: "write docs", Dependencies: []string{depID}},
	}

	var buf bytes.Buffer
	ShowTerminal(e, nil, actionItems, &buf)
	out := buf.String()

	if !strings.Contains(out, "needs: aaaabbbb") {
		t.Errorf("expected dep short ID in output:\n%s", out)
	}
}
