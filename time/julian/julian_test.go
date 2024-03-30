package julian_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go.adoublef.dev/is"
	"go.adoublef.dev/sdk/database/sql3"
	. "go.adoublef.dev/sdk/time/julian"
)

func Test_Time_Scan(t *testing.T) {
	t.Run("OK", testRoundTrip(func(db *sql3.DB) {
		is := is.NewRelaxed(t)

		var jt Time

		err := db.QueryRow(context.TODO(), `select julianday('2016-10-18')`).Scan(&jt)
		is.NoErr(err) // (sql3.DB).QueryRow

		is.Equal(2457679.5, float64(jt))

		if testing.Verbose() {
			t.Log(jt)
		}
	}))
}

func Test_Parse(t *testing.T) {
	t.Run("RFC3339", func(t *testing.T) {
		is := is.NewRelaxed(t)

		jt, err := Parse(time.RFC3339, `2016-10-18T00:00:00Z`)
		is.NoErr(err) // julian.Parse

		if testing.Verbose() {
			t.Log(jt)
		}
	})
}

func Test_Time_MarshalJSON(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		is := is.NewRelaxed(t)

		a, err := Parse(time.RFC3339, `2016-10-18T12:34:56Z`)
		is.NoErr(err) // julian.Parse

		p, err := json.Marshal(a)
		is.NoErr(err) // json.Marshal

		var b Time
		is.NoErr(json.Unmarshal(p, &b))

		is.True(a.Equal(b))
	})
}

func testRoundTrip(f func(*sql3.DB)) func(*testing.T) {
	return func(t *testing.T) {
		db, err := sql3.Open(t.TempDir() + "/test.db")
		if err != nil {
			t.Fatalf("sql3.Up: %v", err)
		}
		t.Cleanup(func() { db.Close() })
		f(db)
	}
}
