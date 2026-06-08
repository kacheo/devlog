package store

import (
	"fmt"
	"time"
)

// Date is a date-only value that marshals to/from YAML as YYYY-MM-DD.
// It deliberately strips time-of-day to avoid timezone drift when storing
// user-facing due dates and ETAs.
type Date time.Time

// DateOf returns a Date from a time.Time, truncating to day boundary in local time.
func DateOf(t time.Time) Date {
	return Date(truncateToDay(t))
}

// String returns the date as YYYY-MM-DD.
func (d Date) String() string {
	return time.Time(d).Format("2006-01-02")
}

// Time returns the underlying time.Time.
func (d Date) Time() time.Time {
	return time.Time(d)
}

// MarshalYAML implements yaml.Marshaler.
func (d Date) MarshalYAML() (interface{}, error) {
	return d.String(), nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (d *Date) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return fmt.Errorf("invalid date %q: expected YYYY-MM-DD", s)
	}
	*d = DateOf(t)
	return nil
}
