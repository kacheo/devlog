package store

import (
	"sort"
	"testing"
	"time"
)

func makeEntry(t *testing.T, st *Store, date time.Time, tags []string) {
	t.Helper()
	e := EmptyEntry(date)
	e.Tags = tags
	if err := st.Save(e); err != nil {
		t.Fatalf("Save(%s): %v", date.Format("2006-01-02"), err)
	}
}

func TestAllDates_EmptyDir(t *testing.T) {
	st, _ := New(t.TempDir())
	dates, err := st.AllDates()
	if err != nil {
		t.Fatalf("AllDates: %v", err)
	}
	if len(dates) != 0 {
		t.Errorf("expected 0 dates, got %d", len(dates))
	}
}

func TestAllDates_ReturnsSortedDates(t *testing.T) {
	st, _ := New(t.TempDir())
	d1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	d2 := time.Date(2026, 1, 3, 0, 0, 0, 0, time.Local)
	d3 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.Local)
	makeEntry(t, st, d1, nil)
	makeEntry(t, st, d3, nil)
	makeEntry(t, st, d2, nil)

	dates, err := st.AllDates()
	if err != nil {
		t.Fatalf("AllDates: %v", err)
	}
	if len(dates) != 3 {
		t.Fatalf("expected 3 dates, got %d", len(dates))
	}
	if !dates[0].Equal(d1) || !dates[1].Equal(d3) || !dates[2].Equal(d2) {
		t.Errorf("dates not sorted oldest-first: %v", dates)
	}
}

func TestListTags_Empty(t *testing.T) {
	st, _ := New(t.TempDir())
	tags, err := st.ListTags()
	if err != nil {
		t.Fatalf("ListTags: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("expected 0 tags, got %d", len(tags))
	}
}

func TestListTags_CountsAndSorts(t *testing.T) {
	st, _ := New(t.TempDir())
	d1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	d2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.Local)
	d3 := time.Date(2026, 1, 3, 0, 0, 0, 0, time.Local)
	makeEntry(t, st, d1, []string{"auth", "backend"})
	makeEntry(t, st, d2, []string{"auth", "frontend"})
	makeEntry(t, st, d3, []string{"auth"})

	tags, err := st.ListTags()
	if err != nil {
		t.Fatalf("ListTags: %v", err)
	}
	if len(tags) != 3 {
		t.Fatalf("expected 3 distinct tags, got %d", len(tags))
	}
	// Sorted by count desc, then alpha
	if tags[0].Tag != "auth" || tags[0].Count != 3 {
		t.Errorf("tags[0] = %+v, want {auth, 3}", tags[0])
	}
	// backend and frontend both have count 1; alphabetical tiebreak
	if tags[1].Tag != "backend" || tags[1].Count != 1 {
		t.Errorf("tags[1] = %+v, want {backend, 1}", tags[1])
	}
	if tags[2].Tag != "frontend" || tags[2].Count != 1 {
		t.Errorf("tags[2] = %+v, want {frontend, 1}", tags[2])
	}
}

func TestListTags_EntriesWithNoTags(t *testing.T) {
	st, _ := New(t.TempDir())
	d1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	d2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.Local)
	makeEntry(t, st, d1, []string{"auth"})
	makeEntry(t, st, d2, nil) // no tags

	tags, err := st.ListTags()
	if err != nil {
		t.Fatalf("ListTags: %v", err)
	}
	if len(tags) != 1 || tags[0].Tag != "auth" {
		t.Errorf("expected [{auth,1}], got %v", tags)
	}
}

func TestRenameTag_Basic(t *testing.T) {
	st, _ := New(t.TempDir())
	d1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	d2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.Local)
	makeEntry(t, st, d1, []string{"auth", "backend"})
	makeEntry(t, st, d2, []string{"auth"})

	n, err := st.RenameTag("auth", "authentication")
	if err != nil {
		t.Fatalf("RenameTag: %v", err)
	}
	if n != 2 {
		t.Errorf("modified %d files, want 2", n)
	}

	// Verify old tag is gone, new tag is present
	e1, _ := st.Load(d1)
	e2, _ := st.Load(d2)
	for _, e := range []*DayEntry{e1, e2} {
		hasNew, hasOld := false, false
		for _, tag := range e.Tags {
			if tag == "authentication" {
				hasNew = true
			}
			if tag == "auth" {
				hasOld = true
			}
		}
		if !hasNew {
			t.Errorf("entry %s missing new tag 'authentication': %v", e.Date.Format("2006-01-02"), e.Tags)
		}
		if hasOld {
			t.Errorf("entry %s still has old tag 'auth': %v", e.Date.Format("2006-01-02"), e.Tags)
		}
	}
	// backend should be preserved
	hasBE := false
	for _, tag := range e1.Tags {
		if tag == "backend" {
			hasBE = true
		}
	}
	if !hasBE {
		t.Errorf("entry d1 lost 'backend' tag: %v", e1.Tags)
	}
}

func TestRenameTag_CaseInsensitiveMatch(t *testing.T) {
	st, _ := New(t.TempDir())
	d1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	makeEntry(t, st, d1, []string{"Auth"})

	n, err := st.RenameTag("auth", "authentication") // lowercase old, mixed-case stored
	if err != nil {
		t.Fatalf("RenameTag: %v", err)
	}
	if n != 1 {
		t.Errorf("modified %d files, want 1", n)
	}

	e, _ := st.Load(d1)
	found := false
	for _, tag := range e.Tags {
		if tag == "authentication" {
			found = true
		}
	}
	if !found {
		t.Errorf("new tag not found: %v", e.Tags)
	}
}

func TestRenameTag_NonExistent(t *testing.T) {
	st, _ := New(t.TempDir())
	d1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	makeEntry(t, st, d1, []string{"backend"})

	n, err := st.RenameTag("auth", "authentication")
	if err != nil {
		t.Fatalf("RenameTag: %v", err)
	}
	if n != 0 {
		t.Errorf("modified %d files, want 0", n)
	}
}

func TestRenameTag_SameTag(t *testing.T) {
	st, _ := New(t.TempDir())
	_, err := st.RenameTag("auth", "auth")
	if err == nil {
		t.Error("expected error when old == new tag")
	}
}

func TestRenameTag_UnaffectedFilesUntouched(t *testing.T) {
	st, _ := New(t.TempDir())
	d1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	d2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.Local)
	makeEntry(t, st, d1, []string{"auth"})
	makeEntry(t, st, d2, []string{"backend"}) // no auth tag

	n, err := st.RenameTag("auth", "authentication")
	if err != nil {
		t.Fatalf("RenameTag: %v", err)
	}
	if n != 1 {
		t.Errorf("modified %d files, want 1", n)
	}

	e2, _ := st.Load(d2)
	tags := e2.Tags
	if len(tags) != 1 || tags[0] != "backend" {
		t.Errorf("d2 tags changed unexpectedly: %v", tags)
	}
}

func TestRenameTag_PreservesOrder(t *testing.T) {
	st, _ := New(t.TempDir())
	d1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	makeEntry(t, st, d1, []string{"first", "auth", "last"})

	_, err := st.RenameTag("auth", "middle")
	if err != nil {
		t.Fatalf("RenameTag: %v", err)
	}

	e, _ := st.Load(d1)
	want := []string{"first", "middle", "last"}
	if !equalStringSlice(e.Tags, want) {
		t.Errorf("tags = %v, want %v", e.Tags, want)
	}
}

func TestRenameTag_NewTagAlreadyPresent(t *testing.T) {
	// Renaming auth→oauth when entry already has [auth, oauth] should yield [oauth], not [oauth, oauth].
	st, _ := New(t.TempDir())
	d1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	makeEntry(t, st, d1, []string{"auth", "oauth"})

	n, err := st.RenameTag("auth", "oauth")
	if err != nil {
		t.Fatalf("RenameTag: %v", err)
	}
	if n != 1 {
		t.Errorf("modified %d files, want 1", n)
	}

	e, _ := st.Load(d1)
	if len(e.Tags) != 1 {
		t.Errorf("tags after rename = %v, want [oauth] (no duplicate)", e.Tags)
	}
	if e.Tags[0] != "oauth" {
		t.Errorf("tag = %q, want 'oauth'", e.Tags[0])
	}
}

func TestRenameTag_NewTagAlreadyPresent_PreservesFirst(t *testing.T) {
	// When [oauth, auth] is renamed auth→oauth, the first oauth (position 0) survives.
	st, _ := New(t.TempDir())
	d1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	makeEntry(t, st, d1, []string{"oauth", "auth"})

	_, err := st.RenameTag("auth", "oauth")
	if err != nil {
		t.Fatalf("RenameTag: %v", err)
	}

	e, _ := st.Load(d1)
	if len(e.Tags) != 1 || e.Tags[0] != "oauth" {
		t.Errorf("tags = %v, want [oauth]", e.Tags)
	}
}

func TestRenameTag_EmptyOldTag(t *testing.T) {
	st, _ := New(t.TempDir())
	_, err := st.RenameTag("", "new")
	if err == nil {
		t.Error("expected error for empty old tag")
	}
}

func TestRenameTag_EmptyNewTag(t *testing.T) {
	st, _ := New(t.TempDir())
	_, err := st.RenameTag("auth", "")
	if err == nil {
		t.Error("expected error for empty new tag")
	}
}

func TestValidateTag_Valid(t *testing.T) {
	valid := []string{
		"auth",
		"backend",
		"api_server",
		"oauth2",
		"v2",
		"my_tag_123",
		"a",
		"z",
		"0",
	}
	for _, tag := range valid {
		if err := ValidateTag(tag); err != nil {
			t.Errorf("ValidateTag(%q) = %v, want nil", tag, err)
		}
	}
}

func TestValidateTag_Invalid(t *testing.T) {
	invalid := []string{
		"Auth",          // uppercase
		"BACKEND",       // uppercase
		"auth-backend",  // hyphen
		"auth backend",  // space
		"auth.backend",  // dot
		"",              // empty
		"CamelCase",     // mixed case
		"snake_Case",    // mixed case
	}
	for _, tag := range invalid {
		if err := ValidateTag(tag); err == nil {
			t.Errorf("ValidateTag(%q) = nil, want error", tag)
		}
	}
}

func TestAllDates_NonExistentDir(t *testing.T) {
	// A store pointing at a dir that doesn't exist yet should return empty, not error.
	st, _ := New(t.TempDir() + "/does-not-exist")
	dates, err := st.AllDates()
	if err != nil {
		t.Fatalf("AllDates on non-existent dir: %v", err)
	}
	if len(dates) != 0 {
		t.Errorf("expected 0 dates, got %d", len(dates))
	}
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestListTags_DeduplicatesWithinDay(t *testing.T) {
	// A single day file should not inflate counts if it somehow has duplicate tags
	st, _ := New(t.TempDir())
	d1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	e := EmptyEntry(d1)
	// Write raw to simulate a file with duplicate tags (not reachable via normal add but defensively correct)
	e.Tags = []string{"auth", "auth"}
	if err := st.Save(e); err != nil {
		t.Fatal(err)
	}

	tags, err := st.ListTags()
	if err != nil {
		t.Fatalf("ListTags: %v", err)
	}
	// Count should be 1 (one day, not 2 occurrences within the same day)
	if len(tags) != 1 || tags[0].Count != 1 {
		t.Errorf("tags = %v, want [{auth,1}]", tags)
	}
}

// Verify that ListTags is case-preserving (first-seen casing wins).
func TestListTags_CasePreserving(t *testing.T) {
	st, _ := New(t.TempDir())
	d1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	d2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.Local)
	// First-seen is "Auth" (d1 is older)
	makeEntry(t, st, d1, []string{"Auth"})
	makeEntry(t, st, d2, []string{"auth"})

	tags, err := st.ListTags()
	if err != nil {
		t.Fatalf("ListTags: %v", err)
	}
	if len(tags) != 1 {
		t.Fatalf("expected 1 distinct tag, got %d: %v", len(tags), tags)
	}
	if tags[0].Count != 2 {
		t.Errorf("count = %d, want 2", tags[0].Count)
	}
	// Display name is the first-seen casing
	if tags[0].Tag != "Auth" {
		t.Errorf("tag display = %q, want %q (first-seen casing)", tags[0].Tag, "Auth")
	}
}

// Make sure AllDates uses sort.Slice (this is implicitly tested by TestAllDates_ReturnsSortedDates,
// but this test makes the sort direction explicit with many dates).
func TestAllDates_SortedOldestFirst(t *testing.T) {
	st, _ := New(t.TempDir())
	dates := []time.Time{
		time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local),
		time.Date(2026, 1, 15, 0, 0, 0, 0, time.Local),
		time.Date(2026, 2, 10, 0, 0, 0, 0, time.Local),
	}
	for _, d := range dates {
		makeEntry(t, st, d, nil)
	}

	got, err := st.AllDates()
	if err != nil {
		t.Fatalf("AllDates: %v", err)
	}
	if !sort.SliceIsSorted(got, func(i, j int) bool { return got[i].Before(got[j]) }) {
		t.Errorf("AllDates not sorted: %v", got)
	}
}
