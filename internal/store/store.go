package store

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// Store manages day files in the journal directory.
type Store struct {
	dir string
}

// New creates a Store rooted at dir. Returns an error if dir is empty.
func New(dir string) (*Store, error) {
	if dir == "" {
		return nil, fmt.Errorf("journal directory is empty; run 'devlog init' to configure")
	}
	return &Store{dir: dir}, nil
}

// DayFilePath returns the absolute path for a date's day file.
func (s *Store) DayFilePath(date time.Time) string {
	return filepath.Join(s.dir, date.Format("2006-01-02")+".md")
}

// Load reads and parses the day file for date. Returns (nil, nil) if the file does not exist.
func (s *Store) Load(date time.Time) (*DayEntry, error) {
	data, err := os.ReadFile(s.DayFilePath(date))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return Parse(data)
}

// LoadOrCreate loads the day file or returns an EmptyEntry if the file is missing.
func (s *Store) LoadOrCreate(date time.Time) (*DayEntry, error) {
	entry, err := s.Load(date)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		entry = EmptyEntry(date)
	}
	return entry, nil
}

// Save atomically writes entry to its day file using a per-file advisory lock.
// Creates the journal directory if needed.
func (s *Store) Save(entry *DayEntry) error {
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return fmt.Errorf("creating journal directory: %w", err)
	}

	path := s.DayFilePath(entry.Date)

	// Advisory lock via a sidecar .lock file
	lockPath := path + ".lock"
	lf, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("opening lock file: %w", err)
	}
	defer func() {
		_ = lf.Close()
		_ = os.Remove(lockPath)
	}()

	if err := syscall.Flock(int(lf.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("acquiring lock: %w", err)
	}
	defer syscall.Flock(int(lf.Fd()), syscall.LOCK_UN) //nolint:errcheck

	content, err := Serialize(entry)
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, content, 0644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}

// ParseDate parses "today", "yesterday", or "YYYY-MM-DD" into a local time.Time at midnight.
func ParseDate(s string) (time.Time, error) {
	now := time.Now()
	switch s {
	case "today":
		return truncateToDay(now), nil
	case "yesterday":
		return truncateToDay(now.AddDate(0, 0, -1)), nil
	case "":
		return time.Time{}, fmt.Errorf("date is required")
	default:
		t, err := time.ParseInLocation("2006-01-02", s, time.Local)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid date %q: expected YYYY-MM-DD", s)
		}
		return t, nil
	}
}

// truncateToDay returns midnight of t in local zone.
func truncateToDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}
