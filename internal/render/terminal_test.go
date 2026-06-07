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

func TestItemsTerminal_Unresolved(t *testing.T) {
	items := []store.Item{
		{ID: "deadbeef-cafe-4000-8000-112233445566", Type: "blocker", Text: "waiting on infra", Dependencies: []string{}},
	}
	var buf bytes.Buffer
	ItemsTerminal(items, &buf)
	out := buf.String()
	if !strings.Contains(out, "[deadbeef]") {
		t.Errorf("expected short ID in output: %q", out)
	}
	if !strings.Contains(out, "waiting on infra") {
		t.Errorf("expected item text in output: %q", out)
	}
	if strings.Contains(out, "resolved") {
		t.Errorf("unresolved item should not show 'resolved': %q", out)
	}
}

func TestItemsTerminal_Resolved(t *testing.T) {
	ts := time.Date(2026, 6, 3, 14, 0, 0, 0, time.UTC)
	items := []store.Item{
		{ID: "aabbccdd-0000-4000-8000-112233445566", Type: "action_item", Text: "write docs", Resolved: true, ResolvedAt: &ts, Dependencies: []string{}},
	}
	var buf bytes.Buffer
	ItemsTerminal(items, &buf)
	out := buf.String()
	if !strings.Contains(out, "resolved 2026-06-03") {
		t.Errorf("expected resolved date in output: %q", out)
	}
}

func TestItemsTerminal_Empty(t *testing.T) {
	var buf bytes.Buffer
	ItemsTerminal([]store.Item{}, &buf)
	if !strings.Contains(buf.String(), "(none)") {
		t.Errorf("expected (none) for empty list: %q", buf.String())
	}
}

func TestItemsTerminal_ShowsDueDate(t *testing.T) {
	due := store.DateOf(time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC))
	items := []store.Item{
		{ID: "deadbeef-cafe-4000-8000-112233445566", Type: "action_item", Text: "finish report", Due: &due, Dependencies: []string{}},
	}
	var buf bytes.Buffer
	ItemsTerminal(items, &buf)
	out := buf.String()
	if !strings.Contains(out, "due 2099-12-31") {
		t.Errorf("expected due date in output: %q", out)
	}
	if strings.Contains(out, "OVERDUE") {
		t.Errorf("future due date should not show OVERDUE: %q", out)
	}
}

func TestItemsTerminal_ShowsOverdue(t *testing.T) {
	past := store.DateOf(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	items := []store.Item{
		{ID: "deadbeef-cafe-4000-8000-112233445566", Type: "action_item", Text: "overdue task", Due: &past, Dependencies: []string{}},
	}
	var buf bytes.Buffer
	ItemsTerminal(items, &buf)
	out := buf.String()
	if !strings.Contains(out, "OVERDUE") {
		t.Errorf("expected OVERDUE marker in output: %q", out)
	}
}

func TestItemsTerminal_ShowsETA(t *testing.T) {
	eta := store.DateOf(time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC))
	items := []store.Item{
		{ID: "deadbeef-cafe-4000-8000-112233445566", Type: "blocker", Text: "waiting on vendor", ETA: &eta, Dependencies: []string{}},
	}
	var buf bytes.Buffer
	ItemsTerminal(items, &buf)
	out := buf.String()
	if !strings.Contains(out, "eta 2026-08-15") {
		t.Errorf("expected eta in output: %q", out)
	}
}

func TestTagsTerminal_Empty(t *testing.T) {
	var buf bytes.Buffer
	TagsTerminal([]store.TagCount{}, &buf)
	if !strings.Contains(buf.String(), "(no tags)") {
		t.Errorf("expected '(no tags)': %q", buf.String())
	}
}

func TestTagsTerminal_ShowsTagsAndCounts(t *testing.T) {
	tags := []store.TagCount{
		{Tag: "auth", Count: 5},
		{Tag: "backend", Count: 1},
	}
	var buf bytes.Buffer
	TagsTerminal(tags, &buf)
	out := buf.String()
	if !strings.Contains(out, "auth") {
		t.Errorf("missing 'auth': %q", out)
	}
	if !strings.Contains(out, "5") {
		t.Errorf("missing count '5': %q", out)
	}
	if !strings.Contains(out, "backend") {
		t.Errorf("missing 'backend': %q", out)
	}
	if !strings.Contains(out, "1") {
		t.Errorf("missing count '1': %q", out)
	}
	// auth should appear before backend
	if strings.Index(out, "auth") > strings.Index(out, "backend") {
		t.Errorf("auth should appear before backend:\n%s", out)
	}
}
