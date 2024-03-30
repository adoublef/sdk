package unix_test

import (
	"context"
	"testing"

	"go.adoublef.dev/is"
	"go.adoublef.dev/sdk/database/sql3"
	. "go.adoublef.dev/sdk/time/unix"
)

func Test_Time_Scan(t *testing.T) {
	t.Run("UnixMS", testRoundTrip(func(db *sql3.DB) {
		is := is.NewRelaxed(t)

		var ut Time
		err := db.QueryRow(context.TODO(), `select strftime('%s', '2016-10-18') * 1000`).Scan(&ut)
		is.NoErr(err) // (sql3.DB).QueryRow

		is.Equal(1476748800000, int(ut))
		t.Log(ut.Time()) // 2016-10-18
	}))

	t.Run("OK", testRoundTrip(func(db *sql3.DB) {
		is := is.NewRelaxed(t)

		var ut Time
		err := db.QueryRow(context.TODO(), `select datetime('2016-10-18T12:34:56Z')`).Scan(&ut)
		is.NoErr(err) // (sql3.DB).QueryRow

		is.Equal(1476794096000, int(ut))
		t.Log(ut.Time()) // 2016-10-18
	}))
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

// 1476745200
