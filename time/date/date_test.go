package date_test

import (
	"encoding/json"
	"testing"

	"go.adoublef.dev/is"
	. "go.adoublef.dev/sdk/time/date"
)

func Test_Parse(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		is := is.NewRelaxed(t)

		d, err := Parse("2016-10-18")
		is.NoErr(err) // date.Parse

		is.Equal(d.Year, 2016)
		is.Equal(d.Month, October)
		is.Equal(d.Day, 18)
	})
}

func Test_Date_UnmarshalJSON(t *testing.T) {
	t.Run("Civil", func(t *testing.T) {
		is := is.NewRelaxed(t)

		var s = `"2016-10-18"`

		var d Date
		is.NoErr(json.Unmarshal([]byte(s), &d))
	})
	t.Run("RFC3339", func(t *testing.T) {
		is := is.NewRelaxed(t)

		var s = `"2016-10-18T00:00:00Z"`

		var d Date
		is.NoErr(json.Unmarshal([]byte(s), &d))
	})

	t.Run("Informal", func(t *testing.T) {
		is := is.NewRelaxed(t)

		var s = `"2016-10-18t00:00:00z"`

		var d Date
		is.NoErr(json.Unmarshal([]byte(s), &d))
	})
}

func Test_Date_MarshalJSON(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		is := is.NewRelaxed(t)

		a := Date{Year: 2016, Month: October, Day: 18}

		p, err := json.Marshal(a)
		is.NoErr(err) // json.Marshal

		var b Date
		is.NoErr(json.Unmarshal(p, &b))

		is.Equal(a, b)
	})
}
