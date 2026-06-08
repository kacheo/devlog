package store

import (
	"testing"
	"time"
)

func TestAddItem_Blocker(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	item, err := st.AddItem("blocker", "Waiting on DevOps", []string{})
	if err != nil {
		t.Fatalf("AddItem: %v", err)
	}
	if item.Type != "blocker" {
		t.Errorf("Type = %q, want blocker", item.Type)
	}
	if item.Text != "Waiting on DevOps" {
		t.Errorf("Text = %q", item.Text)
	}
	if item.Resolved {
		t.Error("should not be resolved")
	}
	if len(item.ID) < 8 {
		t.Errorf("ID too short: %q", item.ID)
	}
}

func TestAddItem_ActionItem(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	item, err := st.AddItem("action_item", "Write release notes", []string{})
	if err != nil {
		t.Fatalf("AddItem: %v", err)
	}
	if item.Type != "action_item" {
		t.Errorf("Type = %q, want action_item", item.Type)
	}
}

func TestAddItem_InvalidType(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	_, err := st.AddItem("diary", "bad type", nil)
	if err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestAddItem_ActionItemCanDependOnBlocker(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	blocker, _ := st.AddItem("blocker", "staging env not ready", []string{})

	ai, err := st.AddItem("action_item", "deploy service", []string{blocker.ID})
	if err != nil {
		t.Fatalf("action_item depending on blocker: %v", err)
	}
	if len(ai.Dependencies) != 1 || ai.Dependencies[0] != blocker.ID {
		t.Errorf("Dependencies = %v, want [%s]", ai.Dependencies, blocker.ID)
	}
}

func TestAddItem_ActionItemCanDependOnActionItem(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	ai1, _ := st.AddItem("action_item", "first task", []string{})
	_, err := st.AddItem("action_item", "second task", []string{ai1.ID})
	if err != nil {
		t.Fatalf("action_item depending on action_item: %v", err)
	}
}

func TestAddItem_BlockerCanDependOnBlocker(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	b1, _ := st.AddItem("blocker", "first blocker", []string{})
	_, err := st.AddItem("blocker", "second blocker", []string{b1.ID})
	if err != nil {
		t.Fatalf("blocker depending on blocker: %v", err)
	}
}

func TestAddItem_BlockerCannotDependOnActionItem(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	ai, _ := st.AddItem("action_item", "some task", []string{})
	_, err := st.AddItem("blocker", "bad blocker", []string{ai.ID})
	if err == nil {
		t.Error("expected error: blocker cannot depend on action_item")
	}
}

func TestAddItem_ShortIDResolution(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	b, _ := st.AddItem("blocker", "first blocker", []string{})
	shortID := b.ID[:8]

	_, err := st.AddItem("blocker", "second blocker", []string{shortID})
	if err != nil {
		t.Fatalf("short ID resolution failed: %v", err)
	}
}

func TestAddItem_UnknownDepErrors(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	_, err := st.AddItem("blocker", "bad dep", []string{"nonexistent"})
	if err == nil {
		t.Error("expected error for unknown dep")
	}
}

func TestResolveItem(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	item, _ := st.AddItem("blocker", "fix this", []string{})
	resolved, err := st.ResolveItem(item.ID)
	if err != nil {
		t.Fatalf("ResolveItem: %v", err)
	}
	if !resolved.Resolved {
		t.Error("item should be resolved")
	}

	// Verify persisted
	items, _ := st.LoadAllItems()
	if len(items) != 1 || !items[0].Resolved {
		t.Error("resolved state not persisted")
	}
}

func TestResolveItem_SetsResolvedAt(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	before := time.Now().Add(-time.Second)
	item, _ := st.AddItem("blocker", "needs timestamp", []string{})
	resolved, err := st.ResolveItem(item.ID)
	if err != nil {
		t.Fatalf("ResolveItem: %v", err)
	}
	after := time.Now().Add(time.Second)

	if resolved.ResolvedAt == nil {
		t.Fatal("ResolvedAt should be set")
	}
	if resolved.ResolvedAt.Before(before) || resolved.ResolvedAt.After(after) {
		t.Errorf("ResolvedAt %v outside expected window [%v, %v]", resolved.ResolvedAt, before, after)
	}

	// Verify persisted
	items, _ := st.LoadAllItems()
	if items[0].ResolvedAt == nil {
		t.Error("ResolvedAt should be persisted in items.yaml")
	}
}

func TestResolveItem_ByShortID(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	item, _ := st.AddItem("action_item", "write docs", []string{})
	_, err := st.ResolveItem(item.ID[:8])
	if err != nil {
		t.Fatalf("ResolveItem by short ID: %v", err)
	}
}

func TestResolveItem_NotFound(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	_, err := st.ResolveItem("notfound")
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}

func TestFilterUnresolved(t *testing.T) {
	items := []Item{
		{ID: "a", Text: "done", Resolved: true},
		{ID: "b", Text: "todo", Resolved: false},
		{ID: "c", Text: "also done", Resolved: true},
	}
	got := FilterUnresolved(items)
	if len(got) != 1 || got[0].ID != "b" {
		t.Errorf("FilterUnresolved = %v, want [{b todo false}]", got)
	}
}

func TestFilterResolved(t *testing.T) {
	past := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	recent := time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC)

	items := []Item{
		{ID: "a", Text: "old resolved", Resolved: true, ResolvedAt: &past},
		{ID: "b", Text: "recent resolved", Resolved: true, ResolvedAt: &recent},
		{ID: "c", Text: "unresolved", Resolved: false},
	}

	// All resolved, no date filter
	got := FilterResolved(items, time.Time{}, time.Time{})
	if len(got) != 2 {
		t.Errorf("FilterResolved (all) = %d items, want 2", len(got))
	}

	// From filter excludes old entry
	from := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)
	got = FilterResolved(items, from, time.Time{})
	if len(got) != 1 || got[0].ID != "b" {
		t.Errorf("FilterResolved (from) = %v, want only b", got)
	}

	// Until filter excludes recent entry
	until := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)
	got = FilterResolved(items, time.Time{}, until)
	if len(got) != 1 || got[0].ID != "a" {
		t.Errorf("FilterResolved (until) = %v, want only a", got)
	}
}

func TestSplitByType(t *testing.T) {
	items := []Item{
		{ID: "1", Type: "blocker", Text: "b1"},
		{ID: "2", Type: "action_item", Text: "a1"},
		{ID: "3", Type: "blocker", Text: "b2"},
	}
	blockers, actionItems := SplitByType(items)
	if len(blockers) != 2 {
		t.Errorf("blockers len = %d, want 2", len(blockers))
	}
	if len(actionItems) != 1 {
		t.Errorf("actionItems len = %d, want 1", len(actionItems))
	}
}

func TestNormalizeItemType(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantOK  bool
	}{
		{"blocker", "blocker", true},
		{"blockers", "blocker", true},
		{"action_item", "action_item", true},
		{"action_items", "action_item", true},
		{"notes", "", false},
		{"diary", "", false},
	}
	for _, c := range cases {
		got, ok := NormalizeItemType(c.in)
		if ok != c.wantOK || got != c.want {
			t.Errorf("NormalizeItemType(%q) = %q, %v; want %q, %v", c.in, got, ok, c.want, c.wantOK)
		}
	}
}

func TestShortID(t *testing.T) {
	id := "deadbeef-cafe-4000-8000-112233445566"
	if got := ShortID(id); got != "deadbeef" {
		t.Errorf("ShortID = %q, want deadbeef", got)
	}
}

func TestLoadAllItems_MissingFile(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	items, err := st.LoadAllItems()
	if err != nil {
		t.Fatalf("LoadAllItems on missing file: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty slice, got %v", items)
	}
}

func TestIsGlobalSection(t *testing.T) {
	if !IsGlobalSection("blockers") {
		t.Error("blockers should be global")
	}
	if !IsGlobalSection("action_items") {
		t.Error("action_items should be global")
	}
	if IsGlobalSection("notes") {
		t.Error("notes should NOT be global")
	}
	if IsGlobalSection("diary") {
		t.Error("diary should NOT be global")
	}
}

func TestMatchesID_EmptyString(t *testing.T) {
	if matchesID("deadbeef-cafe-4000-8000-112233445566", "") {
		t.Error("empty id must not match any candidate")
	}
	if matchesID("", "") {
		t.Error("empty id must not match empty candidate")
	}
}

func TestAddItem_EmptyText(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	_, err := st.AddItem("blocker", "", nil)
	if err == nil {
		t.Error("expected error for empty text")
	}
	_, err = st.AddItem("blocker", "   ", nil)
	if err == nil {
		t.Error("expected error for whitespace-only text")
	}
}

func TestResolveItem_AmbiguousPrefix(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	// Write two items with a shared 8-char prefix directly via modifyItems.
	err := st.modifyItems(func(f *itemFile) error {
		f.Items = []Item{
			{ID: "aaaabbbb-0000-4000-8000-111122223333", Type: "blocker", Text: "first", Dependencies: []string{}},
			{ID: "aaaabbbb-1111-4000-8000-444455556666", Type: "blocker", Text: "second", Dependencies: []string{}},
		}
		return nil
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, err = st.ResolveItem("aaaabbbb")
	if err == nil {
		t.Error("expected ambiguous id error")
	}
}

func TestAddItem_WithDue(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	due := DateOf(time.Now().Add(48 * time.Hour))
	item, err := st.AddItem("action_item", "finish report", []string{}, ItemOptions{Due: &due})
	if err != nil {
		t.Fatalf("AddItem with due: %v", err)
	}
	if item.Due == nil {
		t.Fatal("Due should be set")
	}
	if item.Due.String() != due.String() {
		t.Errorf("Due = %q, want %q", item.Due.String(), due.String())
	}
	if item.ETA != nil {
		t.Error("ETA should be nil on action_item")
	}

	// Verify persisted
	items, _ := st.LoadAllItems()
	if items[0].Due == nil {
		t.Error("Due not persisted")
	}
}

func TestAddItem_DueOnBlocker_Errors(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	due := DateOf(time.Now().Add(48 * time.Hour))
	_, err := st.AddItem("blocker", "external constraint", []string{}, ItemOptions{Due: &due})
	if err == nil {
		t.Error("expected error: --due is only valid on action_item, not blocker")
	}
}

func TestAddItem_WithETA(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	eta := DateOf(time.Now().Add(72 * time.Hour))
	item, err := st.AddItem("blocker", "waiting for infra", []string{}, ItemOptions{ETA: &eta})
	if err != nil {
		t.Fatalf("AddItem with eta: %v", err)
	}
	if item.ETA == nil {
		t.Fatal("ETA should be set")
	}
	if item.ETA.String() != eta.String() {
		t.Errorf("ETA = %q, want %q", item.ETA.String(), eta.String())
	}
	if item.Due != nil {
		t.Error("Due should be nil on blocker")
	}
}

func TestAddItem_ETAOnActionItem_Errors(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	eta := DateOf(time.Now().Add(48 * time.Hour))
	_, err := st.AddItem("action_item", "write tests", []string{}, ItemOptions{ETA: &eta})
	if err == nil {
		t.Error("expected error: --eta is only valid on blocker, not action_item")
	}
}

func TestItem_IsOverdue(t *testing.T) {
	past := DateOf(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	future := DateOf(time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name     string
		item     Item
		wantOver bool
	}{
		{"no due date", Item{Type: "action_item"}, false},
		{"future due date", Item{Type: "action_item", Due: &future}, false},
		{"past due date", Item{Type: "action_item", Due: &past}, true},
		{"past due but resolved", Item{Type: "action_item", Due: &past, Resolved: true}, false},
		{"blocker with past eta", Item{Type: "blocker", ETA: &past}, false}, // ETA doesn't drive overdue
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.IsOverdue(); got != tt.wantOver {
				t.Errorf("IsOverdue() = %v, want %v", got, tt.wantOver)
			}
		})
	}
}

func TestItem_IsETAExceeded(t *testing.T) {
	past := DateOf(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	future := DateOf(time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name    string
		item    Item
		wantExc bool
	}{
		{"no eta", Item{Type: "blocker"}, false},
		{"future eta", Item{Type: "blocker", ETA: &future}, false},
		{"past eta", Item{Type: "blocker", ETA: &past}, true},
		{"past eta but resolved", Item{Type: "blocker", ETA: &past, Resolved: true}, false},
		{"action_item with past due", Item{Type: "action_item", Due: &past}, false}, // Due doesn't drive IsETAExceeded
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.IsETAExceeded(); got != tt.wantExc {
				t.Errorf("IsETAExceeded() = %v, want %v", got, tt.wantExc)
			}
		})
	}
}

func TestAddItem_BackwardCompat_NoDue(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	// Existing callers pass no opts — must still work.
	item, err := st.AddItem("blocker", "legacy call", []string{})
	if err != nil {
		t.Fatalf("AddItem without opts: %v", err)
	}
	if item.Due != nil || item.ETA != nil {
		t.Error("Due and ETA should be nil when not provided")
	}
}
