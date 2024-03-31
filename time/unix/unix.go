package unix

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// Time implements the RFC3339 format.
type Time int64

func (t Time) Time() time.Time { return time.UnixMilli(int64(t)).UTC() }

func (t Time) Equal(t2 Time) bool { return t == t2 }

func (t *Time) Scan(value any) (err error) {
	switch v := value.(type) {
	case int64:
		*t = Time(v)
	case string:
		layout := time.RFC3339
		if v[10] == ' ' {
			layout = time.DateTime
		}
		*t, err = Parse(layout, v)
	default:
		return fmt.Errorf("unix: unsupported type: %T", v)
	}
	return err
}

func (t Time) Value() (driver.Value, error) { return int64(t), nil }

// Parse parses a formatted string and returns the time value it represents using the RFC3339 format.
func Parse(layout, s string) (Time, error) {
	tt, err := time.ParseInLocation(layout, s, time.UTC)
	if err != nil {
		return 0, err
	}
	return Time(tt.UTC().UnixMilli()), nil
}

// FromTime converts a [time.Time] into unix time.
func FromTime(t time.Time) Time { return Time(t.UTC().UnixMilli()) }

// Now returns the current time using the RFC3339 format.
func Now() Time { return FromTime(time.Now()) }
