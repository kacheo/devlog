package store

import (
	"testing"
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
