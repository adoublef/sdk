package julian

import (
	"database/sql/driver"
	"fmt"
	"math"
	"time"
)

const (
	DaySeconds         = 86400
	UnixEpochJulianDay = 2440587.5
)

// Time
type Time float64

func (t Time) String() string {
	return t.Time().Format(time.RFC3339)
}

func (t Time) Equal(t2 Time) bool {
	epsilon := math.Nextafter(1, 2) - 1
	// https://floating-point-gui.de/errors/comparison/#look-out-for-edge-cases
	a := math.Abs(float64(t))
	b := math.Abs(float64(t2))
	diff := math.Abs(float64(t - t2))
	if t == t2 {
		return true
	} else if t == 0 || t2 == 0 || (a+b < math.SmallestNonzeroFloat64) {
		return diff < epsilon*math.SmallestNonzeroFloat64
	}
	return diff/math.Min(a+b, math.MaxFloat64) < epsilon
}

func (t *Time) UnmarshalText(p []byte) (err error) {
	*t, err = Parse(time.RFC3339, string(p))
	return
}

func (t Time) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// Time returns t as [time.Time]
func (t Time) Time() time.Time {
	return time.Unix(int64((t-UnixEpochJulianDay)*DaySeconds), 0).UTC()
}

func (t *Time) Scan(value any) (err error) {
	switch v := value.(type) {
	case float64:
		*t = Time(v)
	case string:
		// civil time should also work
		switch len(v) {
		case 10: // yyyy-mm-dd
			*t, err = Parse(time.DateOnly, v)
		default:
			*t, err = Parse(time.RFC3339, v)
		}
	default:
		return fmt.Errorf("julian: unsupported type: %T", v)
	}
	return err
}

func (t Time) Value() (driver.Value, error) { return float64(t), nil }

// FromTime converts a [time.Time] into a Julian day.
func FromTime(t time.Time) Time {
	return Time(float64(t.UTC().Unix())/DaySeconds + UnixEpochJulianDay)
}

// Now returns the current time. Location is set to UTC.
func Now() Time { return FromTime(time.Now()) }

// Parse
func Parse(layout, s string) (Time, error) {
	t, err := time.Parse(layout, s)
	if err != nil {
		return 0, err
	}
	return FromTime(t), nil
}
