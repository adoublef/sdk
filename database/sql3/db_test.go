package sql3_test

import (
	"context"
	"embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"go.adoublef.dev/is"
	. "go.adoublef.dev/sdk/database/sql3"
)

//go:embed all:testdata/*.up.sql
var embedFS embed.FS
var sqlFS, _ = NewFS(embedFS, "testdata")

func Test_Up(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		is := is.NewRelaxed(t)

		_, err := sqlFS.Up(context.TODO(), testFilename(t, "test.db"))
		is.NoErr(err) // (sql3.FS).Up
	})
}

func Test_DB_Tx(t *testing.T) {
	if testing.Short() {
		t.Skip("this is a long test")
	}

	t.Run("OK", testRoundTrip(func(db *DB) {
		is := is.NewRelaxed(t)

		tx, err := db.Tx(context.TODO())
		is.NoErr(err) // (sql3.DB).Tx

		t.Cleanup(func() { tx.Rollback() })

		for i := range 5_000_000 {
			rid := uuid.Must(uuid.NewV7())
			_, err = tx.Exec(context.TODO(), `insert into tests (id, counter) values (?, ?)`, rid, i)
			is.NoErr(err) // (sql3.Tx).Exec

			if i%500_000 == 0 {
				err = tx.Commit()
				is.NoErr(err) // (sql3.Tx).Commit
				tx, err = db.Tx(context.TODO())
				is.NoErr(err) // (sql3.DB).Tx
			}
		}

		is.NoErr(tx.Commit()) // (sql3.Tx).Commit

		// find
		var rid uuid.UUID
		err = db.QueryRow(context.TODO(), `select id from tests order by id desc limit 1`).Scan(&rid)
		is.NoErr(err) // (sql3.DB).QueryRow
	}))
}

func testRoundTrip(f func(*DB)) func(*testing.T) {
	return func(t *testing.T) {
		db, err := sqlFS.Up(context.TODO(), t.TempDir()+"/test.db")
		if err != nil {
			t.Fatalf("sql3.Up: %v", err)
		}
		t.Cleanup(func() { db.Close() })
		f(db)
	}
}

func testFilename(t testing.TB, filename string) string {
	t.Helper()
	if os.Getenv("DEBUG") != "1" {
		return filepath.Join(t.TempDir(), filename)
	}
	_ = os.Remove(filename)
	return filepath.Join(filename)
}
