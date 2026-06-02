package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNew_EmptyDir_Errors(t *testing.T) {
	_, err := New("")
	if err == nil {
		t.Error("New(\"\") should return error")
	}
}

func TestDayFilePath(t *testing.T) {
	s, err := New("/tmp/devlog-test")
	if err != nil {
		t.Fatal(err)
	}
	date := time.Date(2026, 6, 1, 12, 0, 0, 0, time.Local)
	got := s.DayFilePath(date)
	want := "/tmp/devlog-test/2026-06-01.md"
	if got != want {
		t.Errorf("DayFilePath() = %q, want %q", got, want)
	}
}

func TestLoad_MissingFile_ReturnsNil(t *testing.T) {
	s, _ := New(t.TempDir())
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	entry, err := s.Load(date)
	if err != nil {
		t.Fatalf("Load() on missing file error = %v", err)
	}
	if entry != nil {
		t.Errorf("Load() on missing file = %v, want nil", entry)
	}
}

func TestSave_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(dir)
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	entry := EmptyEntry(date)
	entry.Sections["notes"] = append(entry.Sections["notes"], "hello world")

	if err := s.Save(entry); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	path := filepath.Join(dir, "2026-06-01.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file at %s: %v", path, err)
	}
}

func TestLoadOrCreate_NewFile(t *testing.T) {
	s, _ := New(t.TempDir())
	date := time.Date(2026, 6, 2, 0, 0, 0, 0, time.Local)
	entry, err := s.LoadOrCreate(date)
	if err != nil {
		t.Fatalf("LoadOrCreate() error = %v", err)
	}
	if entry == nil {
		t.Fatal("LoadOrCreate() returned nil")
	}
	if entry.Date != date {
		t.Errorf("Date = %v, want %v", entry.Date, date)
	}
	for _, sec := range KnownSections {
		if _, ok := entry.Sections[sec]; !ok {
			t.Errorf("missing section %q in new entry", sec)
		}
	}
}

func TestSave_RoundTrip(t *testing.T) {
	s, _ := New(t.TempDir())
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	orig := EmptyEntry(date)
	orig.Tags = []string{"auth", "backend"}
	orig.Commits = []Commit{{SHA: "abc1234", Message: "fix: auth bug", Repo: "api"}}
	orig.PRs = []PR{{Number: 10, Title: "Fix auth", State: "merged", Repo: "api"}}
	orig.Sections["notes"] = []string{"Did some work"}
	orig.Sections["blockers"] = []string{"Blocked on review"}

	if err := s.Save(orig); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := s.Load(date)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Date.Format("2006-01-02") != "2026-06-01" {
		t.Errorf("Date mismatch: %v", got.Date)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "auth" {
		t.Errorf("Tags = %v", got.Tags)
	}
	if len(got.Commits) != 1 || got.Commits[0].SHA != "abc1234" {
		t.Errorf("Commits = %v", got.Commits)
	}
	if got.Sections["notes"][0] != "Did some work" {
		t.Errorf("notes = %v", got.Sections["notes"])
	}
}

func TestSave_AtomicWrite_NoTempFileLeft(t *testing.T) {
	dir := t.TempDir()
	s, _ := New(dir)
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local)
	entry := EmptyEntry(date)

	if err := s.Save(entry); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
		check   func(t *testing.T, got time.Time)
	}{
		{
			input: "today",
			check: func(t *testing.T, got time.Time) {
				today := time.Now()
				if got.Year() != today.Year() || got.Month() != today.Month() || got.Day() != today.Day() {
					t.Errorf("'today' resolved to %v, want today %v", got, today)
				}
			},
		},
		{
			input: "yesterday",
			check: func(t *testing.T, got time.Time) {
				yesterday := time.Now().AddDate(0, 0, -1)
				if got.Year() != yesterday.Year() || got.Month() != yesterday.Month() || got.Day() != yesterday.Day() {
					t.Errorf("'yesterday' resolved to %v, want %v", got, yesterday)
				}
			},
		},
		{
			input: "2026-06-01",
			check: func(t *testing.T, got time.Time) {
				if got.Year() != 2026 || got.Month() != 6 || got.Day() != 1 {
					t.Errorf("'2026-06-01' = %v", got)
				}
			},
		},
		{input: "not-a-date", wantErr: true},
		{input: "2026/06/01", wantErr: true},
		{input: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseDate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if err == nil && tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
