package sqltest

import (
	"context"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/adoublef/sdk/sql"
)

// RoundTrip will create a database instance to use with testing
func RoundTrip(fsys fs.FS, f func(t testing.TB, db *sql.DB)) func(t *testing.T) {
	return func(t *testing.T) {
		dsn := filepath.Join(t.TempDir(), "test.db")
		db, err := sql.Open(dsn)
		if err != nil {
			t.Fatalf("open connection: %v", err)
		}
		defer db.Close()
		if err := sql.Up(context.Background(), db, fsys); err != nil {
			t.Fatalf("run migrations: %v", err)
		}
		f(t, db)
	}
}
