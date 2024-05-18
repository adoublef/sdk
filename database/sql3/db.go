package sql3

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/maragudk/migrate"
	_ "github.com/mattn/go-sqlite3"
)

type Executor interface {
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type Querier interface {
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
}

const driverName = "sqlite3"

// DB
type DB struct {
	wc *sql.DB
	rc *sql.DB
}

// Close closes the database and prevents new queries from starting.
func (db *DB) Close() error {
	return errors.Join(db.wc.Close(), db.rc.Close())
}

// Exec executes a query without returning any rows. The args are for any placeholder parameters in the query.
func (db *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.wc.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows, typically a SELECT. The args are for any placeholder parameters in the query.
func (db *DB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.rc.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
func (db *DB) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return db.rc.QueryRowContext(ctx, query, args...)
}

// Tx
func (db *DB) Tx(ctx context.Context) (*Tx, error) {
	tx, err := db.wc.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	return &Tx{tx: tx}, nil
}

// DoTx begins a transaction.
//
// FIXME issue with transaction already been committed
func (db *DB) DoTx(ctx context.Context, f func(context.Context, *Tx) error) error {
	tx, err := db.Tx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = f(ctx, tx); err != nil {
		return fmt.Errorf("run transaction: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit read/write transactions: %w", err)
	}
	return nil
}

// Open a new [DB] connection.
func Open(filename string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return nil, fmt.Errorf("create directory for database files: %w", err)
	}
	dsn := fmt.Sprintf("file:%s?%s", filename, pragma())
	wc, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("open write pool: %w", err)
	}
	wc.SetMaxOpenConns(1)
	rc, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("open read pool: %w", err)
	}
	rc.SetMaxOpenConns(max(4, runtime.NumCPU()))
	return &DB{wc: wc, rc: rc}, nil
}

// Tx
type Tx struct {
	tx *sql.Tx
}

// Commit the transaction.
func (tx *Tx) Commit() error { return tx.tx.Commit() }

// Exec executes a query without returning any rows. The args are for any placeholder parameters in the query.
func (tx *Tx) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return tx.tx.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows, typically a SELECT. The args are for any placeholder parameters in the query.
func (tx *Tx) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return tx.tx.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
func (tx *Tx) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return tx.tx.QueryRowContext(ctx, query, args...)
}

// Rollback the transaction.
func (tx *Tx) Rollback() error { return tx.tx.Rollback() }

// Up from the current version.
func Up(ctx context.Context, filename string, fsys fs.FS) (*DB, error) {
	db, err := Open(filename)
	if err != nil {
		return nil, err
	}
	if err := migrate.Up(ctx, db.wc, fsys); err != nil {
		return nil, fmt.Errorf("running up migrations: %w", err)
	}
	return db, nil
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// FS
type FS struct {
	fsys fs.FS
}

// Up from the current version.
func (fs *FS) Up(ctx context.Context, filename string) (*DB, error) {
	return Up(ctx, filename, fs.fsys)
}

// NewFS returns a [FS]
func NewFS(fsys fs.FS, dir string) (*FS, error) {
	if dir == "" {
		return &FS{fsys: fsys}, nil
	}
	fsys, err := fs.Sub(fsys, dir)
	return &FS{fsys: fsys}, err
}

func pragma() string {
	v := map[string]string{
		"_journal_mode": "wal",
		"_busy_timeout": "5000",
		"_synchronous":  "normal",
		"_cache_size":   "1000000000",
		"_foreign_keys": "true",
		"_txlock":       "immediate",
		"_temp_store":   "memory",
		"_mmap_size":    "3000000000",
	}

	var buf strings.Builder
	for k, val := range v {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(val)
	}
	return buf.String()
}
