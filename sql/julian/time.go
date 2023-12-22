package julian

import (
	"time"
)

const (
	DaySeconds         = 86400
	UnixEpochJulianDay = 2440587.5
)

type Time float64

// AsTime converts a Julian day into a time.Time.
func (t Time) AsTime() time.Time {
	return time.Unix(int64((t-UnixEpochJulianDay)*DaySeconds), 0).UTC()
}

// FromTime converts a time.Time into a Julian day.
func FromTime(t time.Time) Time {
	return Time(float64(t.UTC().Unix())/DaySeconds + UnixEpochJulianDay)
}

func Now() Time {
	return FromTime(time.Now())
}
