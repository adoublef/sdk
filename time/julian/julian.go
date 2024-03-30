package julian

import (
	"database/sql/driver"
	"fmt"
	"time"
)

const (
	DaySeconds         = 86400
	UnixEpochJulianDay = 2440587.5
)

// Time
type Time float64

// Time returns t as [time.Time]
func (t Time) Time() time.Time {
	return time.Unix(int64((t-UnixEpochJulianDay)*DaySeconds), 0).UTC()
}

func (t *Time) Scan(value any) (err error) {
	switch v := value.(type) {
	case int64:
		*t = Time(v)
	default:
		return fmt.Errorf("unix: unsupported type: %T", v)
	}
	return err
}

func (t Time) Value() (driver.Value, error) { return t, nil }

// FromTime converts a [time.Time] into a Julian day.
func FromTime(t time.Time) Time {
	return Time(float64(t.UTC().Unix())/DaySeconds + UnixEpochJulianDay)
}

// Now returns the current time. Location is set to UTC.
func Now() Time { return FromTime(time.Now()) }
