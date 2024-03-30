// Package date implements types a time-zone-independent
// representation of time.
//
// Use [time.Time] for that purpose if location information is required.
package date

import (
	"fmt"
	"time"
)

type Month = time.Month

const (
	January Month = iota + 1
	February
	March
	April
	May
	June
	July
	August
	September
	October
	November
	December
)

// A Date represents a date (year, month, day).
type Date struct {
	Year  int
	Month Month
	Day   int
}

func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

// Add returns the [Date] that is n days in the future.
// n can also be negative to go into the past.
func (d Date) Add(n int) Date {
	return DateOf(d.In(time.UTC).AddDate(0, 0, n))
}

// IsValid reports whether the [Date] is valid.
func (d Date) IsValid() bool {
	return DateOf(d.In(time.UTC)) == d
}

// In returns the [time.Time] corresponding to time 00:00:00 of the [Date] in the location.
func (d Date) In(loc *time.Location) time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, loc)
}

// Since returns the signed number of days between the date and s, not including the end day.
// This is the inverse operation to [Time.Add].
func (d Date) Since(s Date) (days int) {
	// We convert to Unix time so we do not have to worry about leap seconds:
	// Unix time increases by exactly 86400 seconds per day.
	deltaUnix := d.In(time.UTC).Unix() - s.In(time.UTC).Unix()
	return int(deltaUnix / 86400)
}

// Before reports whether d occurs before d2.
func (d Date) Before(d2 Date) bool {
	if d.Year != d2.Year {
		return d.Year < d2.Year
	}
	if d.Month != d2.Month {
		return d.Month < d2.Month
	}
	return d.Day < d2.Day
}

// After reports whether d occurs after d2.
func (d Date) After(d2 Date) bool {
	return d2.Before(d)
}

// IsZero reports whether date fields are set to their default value.
func (d Date) IsZero() bool {
	return (d.Year == 0) && (int(d.Month) == 0) && (d.Day == 0)
}

func (d *Date) UnmarshalText(p []byte) (err error) {
	s, _, _ := cut(string(p), 10) // informal
	*d, err = Parse(s)
	return err
}

func cut(s string, idx int) (before, after string, found bool) {
	if len(s) > idx && (s[idx] == 'T' || s[idx] == 't') {
		return s[:idx], s[idx+1:], true
	}
	return s, "", false
}

func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// Scan

// Value

// Parse parses a string in RFC3339 full-date formate and returns the value as a [Date].
func Parse(s string) (Date, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return Date{}, err
	}
	return DateOf(t), nil
}

// DateOf
func DateOf(t time.Time) (d Date) {
	d.Year, d.Month, d.Day = t.Date()
	return
}
