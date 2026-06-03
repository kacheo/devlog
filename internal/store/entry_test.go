package store

import (
	"bytes"
	"testing"
	"time"
)

var sampleFile = []byte(`---
date: 2026-06-01
tags: [auth, backend]
commits:
  - { sha: abc1234, message: "fix: oauth token refresh loop", repo: api-server }
prs:
  - { number: 142, title: "Add rate limiter to user endpoints", state: merged, repo: api-server }
---

## Notes
- Fixed the OAuth token refresh loop
- Reviewed Alice's PR

## Commits
- abc1234 fix: oauth token refresh loop (api-server)

## PRs
- #142 Add rate limiter to user endpoints [merged] (api-server)

## Blockers
- Waiting on DevOps to provision staging DB
`)

func TestParse_Frontmatter(t *testing.T) {
	entry, err := Parse(sampleFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if entry.Date.Format("2006-01-02") != "2026-06-01" {
		t.Errorf("Date = %v, want 2026-06-01", entry.Date)
	}
	if len(entry.Tags) != 2 || entry.Tags[0] != "auth" || entry.Tags[1] != "backend" {
		t.Errorf("Tags = %v, want [auth backend]", entry.Tags)
	}
	if len(entry.Commits) != 1 {
		t.Fatalf("len(Commits) = %d, want 1", len(entry.Commits))
	}
	if entry.Commits[0].SHA != "abc1234" {
		t.Errorf("Commits[0].SHA = %q, want abc1234", entry.Commits[0].SHA)
	}
	if entry.Commits[0].Message != "fix: oauth token refresh loop" {
		t.Errorf("Commits[0].Message = %q", entry.Commits[0].Message)
	}
	if entry.Commits[0].Repo != "api-server" {
		t.Errorf("Commits[0].Repo = %q", entry.Commits[0].Repo)
	}
	if entry.Commits[0].RepoSlug != "" {
		t.Errorf("Commits[0].RepoSlug = %q, want empty (old format has no repo_slug)", entry.Commits[0].RepoSlug)
	}
	if len(entry.PRs) != 1 || entry.PRs[0].Number != 142 {
		t.Errorf("PRs = %+v", entry.PRs)
	}
	if entry.PRs[0].URL != "" {
		t.Errorf("PRs[0].URL = %q, want empty (old format has no url)", entry.PRs[0].URL)
	}
}

func TestParse_ProseSection_Notes(t *testing.T) {
	entry, err := Parse(sampleFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	notes := entry.Sections["notes"]
	if len(notes) != 2 {
		t.Fatalf("notes len = %d, want 2", len(notes))
	}
	if notes[0] != "Fixed the OAuth token refresh loop" {
		t.Errorf("notes[0] = %q", notes[0])
	}
	if notes[1] != "Reviewed Alice's PR" {
		t.Errorf("notes[1] = %q", notes[1])
	}
}

func TestParse_ProseSection_Blockers(t *testing.T) {
	entry, err := Parse(sampleFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	blockers := entry.Sections["blockers"]
	if len(blockers) != 1 || blockers[0] != "Waiting on DevOps to provision staging DB" {
		t.Errorf("blockers = %v", blockers)
	}
}

func TestParse_EmptyFrontmatter(t *testing.T) {
	entry, err := Parse([]byte("---\ndate: 2026-06-01\n---\n"))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if entry.Date.IsZero() {
		t.Error("Date is zero")
	}
	if len(entry.Commits) != 0 {
		t.Errorf("expected no commits, got %v", entry.Commits)
	}
}

func TestParse_FrontmatterNoTrailingNewline(t *testing.T) {
	// Regression: frontmatter closed by "---" at EOF without trailing newline
	// previously caused an out-of-bounds panic.
	input := []byte("---\ndate: 2026-06-01\n---")
	entry, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if entry.Date.Format("2006-01-02") != "2026-06-01" {
		t.Errorf("Date = %v, want 2026-06-01", entry.Date)
	}
}

func TestParse_NoFrontmatter(t *testing.T) {
	entry, err := Parse([]byte("## Notes\n- hello\n"))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !entry.Date.IsZero() {
		t.Errorf("Date should be zero, got %v", entry.Date)
	}
	notes := entry.Sections["notes"]
	if len(notes) != 1 || notes[0] != "hello" {
		t.Errorf("notes = %v", notes)
	}
}

func TestNormalizeSection(t *testing.T) {
	tests := []struct {
		input  string
		want   string
		wantOK bool
	}{
		{"notes", "notes", true},
		{"Notes", "notes", true},
		{"note", "notes", true},
		{"NOTE", "notes", true},
		{"blockers", "blockers", true},
		{"blocker", "blockers", true},
		{"Blockers", "blockers", true},
		{"action_items", "action_items", true},
		{"action_item", "action_items", true},
		{"action", "action_items", true},
		{"unknown", "", false},
		{"diary", "", false},
		{"", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := NormalizeSection(tt.input)
			if ok != tt.wantOK {
				t.Errorf("NormalizeSection(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("NormalizeSection(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEmptyEntry_HasAllSections(t *testing.T) {
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	entry := EmptyEntry(date)
	for _, section := range KnownSections {
		if _, ok := entry.Sections[section]; !ok {
			t.Errorf("EmptyEntry missing section %q", section)
		}
	}
	if entry.Date != date {
		t.Errorf("Date = %v, want %v", entry.Date, date)
	}
}

func TestSerialize_RoundTrip(t *testing.T) {
	entry, err := Parse(sampleFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	out, err := Serialize(entry)
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	entry2, err := Parse(out)
	if err != nil {
		t.Fatalf("Parse(Serialize(entry)) error = %v", err)
	}
	if entry2.Date.Format("2006-01-02") != entry.Date.Format("2006-01-02") {
		t.Errorf("Date mismatch after round-trip")
	}
	if len(entry2.Commits) != len(entry.Commits) {
		t.Errorf("Commits length mismatch: got %d, want %d", len(entry2.Commits), len(entry.Commits))
	}
	if len(entry2.Sections["notes"]) != len(entry.Sections["notes"]) {
		t.Errorf("notes length mismatch: got %d, want %d", len(entry2.Sections["notes"]), len(entry.Sections["notes"]))
	}
}

func TestSerialize_CommitsNotInBody(t *testing.T) {
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	entry := EmptyEntry(date)
	entry.Commits = []Commit{{SHA: "abc1234", Message: "fix: auth bug", Repo: "api-server"}}
	out, err := Serialize(entry)
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	if bytes.Contains(out, []byte("## Commits")) {
		t.Errorf("Serialize should not emit ## Commits body section")
	}
	// Commits still appear in frontmatter
	if !bytes.Contains(out, []byte("sha: abc1234")) {
		t.Errorf("Commits should still appear in frontmatter")
	}
}

func TestSerialize_PRsNotInBody(t *testing.T) {
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	entry := EmptyEntry(date)
	entry.PRs = []PR{{Number: 142, Title: "Add rate limiter", State: "merged", Repo: "api-server"}}
	out, err := Serialize(entry)
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	if bytes.Contains(out, []byte("## PRs")) {
		t.Errorf("Serialize should not emit ## PRs body section")
	}
	// PRs still appear in frontmatter
	if !bytes.Contains(out, []byte("number: 142")) {
		t.Errorf("PRs should still appear in frontmatter")
	}
}

func TestSerialize_FrontmatterDateFormat(t *testing.T) {
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	entry := EmptyEntry(date)
	out, err := Serialize(entry)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out, []byte("date: 2026-06-01")) {
		t.Errorf("Frontmatter missing date:\n%s", out)
	}
}

func TestSerialize_PreservesNotesOrder(t *testing.T) {
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	entry := EmptyEntry(date)
	entry.Sections["notes"] = []string{"first", "second", "third"}
	out, err := Serialize(entry)
	if err != nil {
		t.Fatal(err)
	}
	entry2, _ := Parse(out)
	notes := entry2.Sections["notes"]
	if len(notes) != 3 || notes[0] != "first" || notes[2] != "third" {
		t.Errorf("notes order not preserved: %v", notes)
	}
}

func TestSerialize_CommitsWithRepoSlug(t *testing.T) {
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	entry := EmptyEntry(date)
	entry.Commits = []Commit{{SHA: "abc1234", Message: "fix: auth", Repo: "api-server", RepoSlug: "kacheo/api-server"}}
	out, err := Serialize(entry)
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	if !bytes.Contains(out, []byte("repo_slug: kacheo/api-server")) {
		t.Errorf("frontmatter missing repo_slug field.\nGot:\n%s", out)
	}
}

func TestSerialize_PRWithURL(t *testing.T) {
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	entry := EmptyEntry(date)
	entry.PRs = []PR{{Number: 142, Title: "Add rate limiter", State: "merged", Repo: "api-server", URL: "https://github.com/kacheo/api-server/pull/142"}}
	out, err := Serialize(entry)
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	if !bytes.Contains(out, []byte("url: https://github.com/kacheo/api-server/pull/142")) {
		t.Errorf("frontmatter missing url field.\nGot:\n%s", out)
	}
}

func TestSerialize_PROMitsEmptyURL(t *testing.T) {
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	entry := EmptyEntry(date)
	entry.PRs = []PR{{Number: 142, Title: "Add rate limiter", State: "merged", Repo: "api-server", URL: ""}}
	out, err := Serialize(entry)
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	if bytes.Contains(out, []byte("url:")) {
		t.Errorf("frontmatter should omit empty url field.\nGot:\n%s", out)
	}
}

func TestKnownSections_NoCommitsPRs(t *testing.T) {
	for _, s := range KnownSections {
		if s == "commits" || s == "prs" {
			t.Errorf("KnownSections still contains %q — should be removed", s)
		}
	}
}

func TestKnownSections_HasActionItems(t *testing.T) {
	found := false
	for _, s := range KnownSections {
		if s == "action_items" {
			found = true
		}
	}
	if !found {
		t.Error("KnownSections missing action_items")
	}
}

func TestActionItems_RoundTrip(t *testing.T) {
	e := EmptyEntry(time.Date(2026, 1, 4, 0, 0, 0, 0, time.Local))
	e.Sections["action_items"] = []string{"review PR", "update docs"}

	data, err := Serialize(e)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte("## Action Items")) {
		t.Error("expected ## Action Items heading in serialized output")
	}

	parsed, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.Sections["action_items"]) != 2 {
		t.Fatalf("expected 2 action_items, got %v", parsed.Sections["action_items"])
	}
	if parsed.Sections["action_items"][0] != "review PR" {
		t.Errorf("expected 'review PR', got %q", parsed.Sections["action_items"][0])
	}
}

func TestNormalizeSection_ActionItems(t *testing.T) {
	cases := []string{"action", "action_item", "action_items", "action items", "Action Items"}
	for _, c := range cases {
		got, ok := NormalizeSection(c)
		if !ok || got != "action_items" {
			t.Errorf("NormalizeSection(%q) = %q, %v; want action_items, true", c, got, ok)
		}
	}
}

func TestNormalizeSection_CommitsPRsRemoved(t *testing.T) {
	for _, s := range []string{"commit", "commits", "pr", "prs"} {
		_, ok := NormalizeSection(s)
		if ok {
			t.Errorf("NormalizeSection(%q) should return false after removing commits/prs", s)
		}
	}
}

func TestParse_ActionItems(t *testing.T) {
	raw := []byte(`---
date: "2026-01-07"
tags: []
commits: []
prs: []
---

## Notes

## Action Items
- review PR
- update docs

## Blockers
`)
	entry, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	items := entry.Sections["action_items"]
	if len(items) != 2 {
		t.Fatalf("expected 2 action_items, got %v", items)
	}
	if items[0] != "review PR" {
		t.Errorf("items[0] = %q, want %q", items[0], "review PR")
	}
	if items[1] != "update docs" {
		t.Errorf("items[1] = %q, want %q", items[1], "update docs")
	}
}
