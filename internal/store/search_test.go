package store

import (
	"testing"
	"time"
)

func TestSearch_MatchesBodySection(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	entry := EmptyEntry(date)
	entry.Sections["notes"] = []string{"worked on oauth flow", "unrelated note"}
	if err := st.Save(entry); err != nil {
		t.Fatal(err)
	}

	results, err := st.Search(SearchOpts{Query: "oauth"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Search() = %d results, want 1", len(results))
	}
	if results[0].Section != "notes" || results[0].Line != "worked on oauth flow" {
		t.Errorf("unexpected result: %+v", results[0])
	}
	if results[0].Date != "2026-06-01" {
		t.Errorf("Date = %q, want 2026-06-01", results[0].Date)
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	entry := EmptyEntry(date)
	entry.Sections["notes"] = []string{"OAuth 2.0 setup"}
	if err := st.Save(entry); err != nil {
		t.Fatal(err)
	}

	results, err := st.Search(SearchOpts{Query: "oauth"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Search() = %d results, want 1", len(results))
	}
}

func TestSearch_SectionsFilter(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	entry := EmptyEntry(date)
	entry.Sections["notes"] = []string{"oauth in notes"}
	entry.Sections["blockers"] = []string{"oauth in blockers"}
	if err := st.Save(entry); err != nil {
		t.Fatal(err)
	}

	results, err := st.Search(SearchOpts{Query: "oauth", Sections: []string{"notes"}})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Search() = %d results, want 1 (only notes)", len(results))
	}
	if results[0].Section != "notes" {
		t.Errorf("Section = %q, want 'notes'", results[0].Section)
	}
}

func TestSearch_DateRangeFilter(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	early := time.Date(2026, 5, 1, 0, 0, 0, 0, time.Local)
	late := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)

	for _, d := range []time.Time{early, late} {
		e := EmptyEntry(d)
		e.Sections["notes"] = []string{"oauth work"}
		if err := st.Save(e); err != nil {
			t.Fatal(err)
		}
	}

	results, err := st.Search(SearchOpts{
		Query: "oauth",
		From:  time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local),
		Until: time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local),
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Search() = %d results, want 1 (only June 1)", len(results))
	}
	if results[0].Date != "2026-06-01" {
		t.Errorf("Date = %q, want 2026-06-01", results[0].Date)
	}
}

func TestSearch_MatchesCommitMessage(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	entry := EmptyEntry(date)
	entry.Commits = []Commit{{SHA: "abc1234", Message: "fix: oauth token refresh", Repo: "api"}}
	if err := st.Save(entry); err != nil {
		t.Fatal(err)
	}

	results, err := st.Search(SearchOpts{Query: "oauth"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Search() = %d results, want 1", len(results))
	}
	if results[0].Section != "commits" {
		t.Errorf("Section = %q, want 'commits'", results[0].Section)
	}
}

func TestSearch_MatchesPRTitle(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	entry := EmptyEntry(date)
	entry.PRs = []PR{{Number: 12, Title: "feat: add oauth login", State: "open", Repo: "api"}}
	if err := st.Save(entry); err != nil {
		t.Fatal(err)
	}

	results, err := st.Search(SearchOpts{Query: "oauth"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Search() = %d results, want 1", len(results))
	}
	if results[0].Section != "prs" {
		t.Errorf("Section = %q, want 'prs'", results[0].Section)
	}
}

func TestSearch_NoMatches_ReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	entry := EmptyEntry(date)
	entry.Sections["notes"] = []string{"nothing relevant here"}
	if err := st.Save(entry); err != nil {
		t.Fatal(err)
	}

	results, err := st.Search(SearchOpts{Query: "oauth"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Search() = %d results, want 0", len(results))
	}
}

func TestSearch_EmptyDir_ReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	results, err := st.Search(SearchOpts{Query: "anything"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Search() = %d results, want 0", len(results))
	}
}

func TestSearch_UnknownSection_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	_, err := st.Search(SearchOpts{Query: "oauth", Sections: []string{"diary"}})
	if err == nil {
		t.Error("Search() with unknown section should return error")
	}
}

func TestSearch_SectionsFilter_CommitsOnly(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	entry := EmptyEntry(date)
	entry.Sections["notes"] = []string{"oauth in notes"}
	entry.Commits = []Commit{{SHA: "abc1234", Message: "oauth commit", Repo: "api"}}
	if err := st.Save(entry); err != nil {
		t.Fatal(err)
	}

	results, err := st.Search(SearchOpts{Query: "oauth", Sections: []string{"commits"}})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Search() = %d results, want 1 (commits only)", len(results))
	}
	if results[0].Section != "commits" {
		t.Errorf("Section = %q, want 'commits'", results[0].Section)
	}
}

func TestSearch_MultipleFiles_ReturnsAll(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)

	for i := 1; i <= 3; i++ {
		d := time.Date(2026, 6, i, 0, 0, 0, 0, time.Local)
		e := EmptyEntry(d)
		e.Sections["notes"] = []string{"oauth work"}
		if err := st.Save(e); err != nil {
			t.Fatal(err)
		}
	}

	results, err := st.Search(SearchOpts{Query: "oauth"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("Search() = %d results, want 3", len(results))
	}
}
