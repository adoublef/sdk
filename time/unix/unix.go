package unix

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// Time implements the RFC3339 format.
type Time int64

func (t Time) Time() time.Time { return time.UnixMilli(int64(t)).UTC() }

func (t *Time) Scan(value any) (err error) {
	switch v := value.(type) {
	case int64:
		*t = Time(v)
	case string:
		*t, err = Parse(v)
	default:
		return fmt.Errorf("unix: unsupported type: %T", v)
	}
	return err
}

func (t Time) Value() (driver.Value, error) { return t, nil }

// Parse parses a formatted string and returns the time value it represents using the RFC3339 format.
func Parse(s string) (Time, error) {
	tt, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return 0, err
	}
	return Time(tt.UnixMilli()), nil
}

// FromTime converts a [time.Time] into unix time.
func FromTime(t time.Time) Time { return Time(t.UTC().UnixMilli()) }

// Now returns the current time using the RFC3339 format.
func Now() Time { return FromTime(time.Now()) }
