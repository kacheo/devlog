package store

import (
	"testing"
	"time"
)

func TestDateOf_TruncatesTime(t *testing.T) {
	noon := time.Date(2026, 6, 15, 12, 30, 45, 999, time.Local)
	d := DateOf(noon)
	got := d.Time()
	if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
		t.Errorf("DateOf should truncate to midnight, got %v", got)
	}
	y, m, day := got.Date()
	if y != 2026 || m != time.June || day != 15 {
		t.Errorf("DateOf changed the calendar date: got %04d-%02d-%02d, want 2026-06-15", y, m, day)
	}
}

func TestDateOf_PreservesCalendarDate(t *testing.T) {
	in := time.Date(2026, 6, 15, 12, 30, 0, 0, time.UTC)
	if got := DateOf(in).String(); got != "2026-06-15" {
		t.Errorf("DateOf().String() = %q, want \"2026-06-15\"", got)
	}
}

func TestDate_String(t *testing.T) {
	d := DateOf(time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local))
	if got := d.String(); got != "2000-01-01" {
		t.Errorf("String() = %q, want \"2000-01-01\"", got)
	}
}

func TestDate_Time(t *testing.T) {
	in := time.Date(2026, 9, 1, 14, 0, 0, 0, time.Local)
	d := DateOf(in)
	got := d.Time()
	if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 {
		t.Errorf("Time() should be midnight, got %v", got)
	}
	y, m, day := got.Date()
	if y != 2026 || m != time.September || day != 1 {
		t.Errorf("Time() changed the calendar date: got %04d-%02d-%02d", y, m, day)
	}
}

func TestDate_MarshalYAML(t *testing.T) {
	d := DateOf(time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local))
	v, err := d.MarshalYAML()
	if err != nil {
		t.Fatalf("MarshalYAML: %v", err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("MarshalYAML returned %T, want string", v)
	}
	if s != "2026-03-15" {
		t.Errorf("MarshalYAML = %q, want \"2026-03-15\"", s)
	}
}

func TestDate_UnmarshalYAML_Valid(t *testing.T) {
	var d Date
	err := d.UnmarshalYAML(func(v interface{}) error {
		*v.(*string) = "2026-09-01"
		return nil
	})
	if err != nil {
		t.Fatalf("UnmarshalYAML: %v", err)
	}
	if got := d.String(); got != "2026-09-01" {
		t.Errorf("String() = %q, want \"2026-09-01\"", got)
	}
}

func TestDate_UnmarshalYAML_Invalid(t *testing.T) {
	var d Date
	err := d.UnmarshalYAML(func(v interface{}) error {
		*v.(*string) = "not-a-date"
		return nil
	})
	if err == nil {
		t.Error("expected error for invalid date string")
	}
}

func TestDate_MarshalUnmarshalYAML(t *testing.T) {
	original := DateOf(time.Date(2026, 6, 15, 12, 30, 0, 0, time.UTC))
	b, err := original.MarshalYAML()
	if err != nil {
		t.Fatalf("MarshalYAML: %v", err)
	}
	s, ok := b.(string)
	if !ok {
		t.Fatalf("MarshalYAML returned %T, want string", b)
	}
	if s != "2026-06-15" {
		t.Errorf("MarshalYAML = %q, want \"2026-06-15\"", s)
	}

	var got Date
	if err := got.UnmarshalYAML(func(v interface{}) error {
		sp := v.(*string)
		*sp = s
		return nil
	}); err != nil {
		t.Fatalf("UnmarshalYAML: %v", err)
	}
	if got.String() != "2026-06-15" {
		t.Errorf("round-trip = %q, want \"2026-06-15\"", got.String())
	}
}
