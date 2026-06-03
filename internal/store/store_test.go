package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
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

func TestModify_CreatesEntryIfMissing(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)
	date := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)

	err := st.Modify(date, func(e *DayEntry) error {
		e.Sections["notes"] = append(e.Sections["notes"], "hello")
		return nil
	})
	if err != nil {
		t.Fatalf("Modify: %v", err)
	}

	entry, err := st.Load(date)
	if err != nil || entry == nil {
		t.Fatalf("Load after Modify: %v", err)
	}
	if len(entry.Sections["notes"]) != 1 || entry.Sections["notes"][0] != "hello" {
		t.Fatalf("expected [hello], got %v", entry.Sections["notes"])
	}
}

func TestModify_PreservesExistingData(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)
	date := time.Date(2026, 1, 2, 0, 0, 0, 0, time.Local)

	if err := st.Modify(date, func(e *DayEntry) error {
		e.Sections["notes"] = append(e.Sections["notes"], "first")
		return nil
	}); err != nil {
		t.Fatalf("first Modify: %v", err)
	}
	if err := st.Modify(date, func(e *DayEntry) error {
		e.Sections["notes"] = append(e.Sections["notes"], "second")
		return nil
	}); err != nil {
		t.Fatalf("second Modify: %v", err)
	}

	entry, _ := st.Load(date)
	if len(entry.Sections["notes"]) != 2 {
		t.Fatalf("expected 2 notes, got %v", entry.Sections["notes"])
	}
}

func TestModify_FnErrorAbortsWrite(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)
	date := time.Date(2026, 1, 3, 0, 0, 0, 0, time.Local)

	_ = st.Modify(date, func(e *DayEntry) error {
		e.Sections["notes"] = append(e.Sections["notes"], "initial")
		return nil
	})

	err := st.Modify(date, func(e *DayEntry) error {
		e.Sections["notes"] = append(e.Sections["notes"], "should not persist")
		return fmt.Errorf("abort")
	})
	if err == nil {
		t.Fatal("expected error from Modify when fn returns error")
	}

	entry, _ := st.Load(date)
	if len(entry.Sections["notes"]) != 1 {
		t.Fatalf("expected 1 note (abort should prevent write), got %v", entry.Sections["notes"])
	}
}

func TestModify_ConcurrentWrites(t *testing.T) {
	dir := t.TempDir()
	st, _ := New(dir)
	date := time.Date(2026, 1, 6, 0, 0, 0, 0, time.Local)

	const n = 10
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := st.Modify(date, func(e *DayEntry) error {
				e.Sections["notes"] = append(e.Sections["notes"], "item")
				return nil
			}); err != nil {
				t.Errorf("Modify: %v", err)
			}
		}()
	}
	wg.Wait()

	entry, err := st.Load(date)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(entry.Sections["notes"]) != n {
		t.Errorf("expected %d notes, got %d (lost updates under concurrency)", n, len(entry.Sections["notes"]))
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
