package sql

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/fs"

	"github.com/maragudk/migrate"
	_ "github.com/mattn/go-sqlite3"
)

const (
	driverName = "sqlite3"
	args       = "_fk=true&_journal_mod=wal&_synchronous=normal&_timeout=5000"
)

var _ ReadWriter = (*DB)(nil)

type DB struct {
	rwc *sql.DB
}

func (db *DB) Begin() (*Tx, error) {
	tx, err := db.rwc.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transactions: %w", err)
	}
	return &Tx{tx}, nil
}

// Query implements DB.
func (db *DB) Query(ctx context.Context, query string, args ...any) (ScanIterator, error) {
	rs, err := db.rwc.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	return rs, nil
}

// Exec executes a query without returning any rows.
func (db *DB) Exec(ctx context.Context, query string, args ...any) (int64, error) {
	rs, err := db.rwc.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("execute: %w", err)
	}
	n, err := rs.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}
	return n, nil
}

// QueryRow executes a query that is expected to return at most one row.
func (db *DB) QueryRow(ctx context.Context, query string, args ...any) Scanner {
	return db.rwc.QueryRowContext(ctx, query, args...)
}

// Close closes the database and prevents new queries from starting. C
func (db *DB) Close() error {
	return db.rwc.Close()
}

// Open opens a database connection for the given sqlite file.
func Open(dsn string) (*DB, error) {
	db, err := sql.Open(driverName, dsn+"?"+args)
	if err != nil {
		return nil, fmt.Errorf("open sqlite3: %w", err)
	}
	return &DB{db}, nil
}

type Scanner interface {
	Err() error
	Scan(dest ...any) error
}

type ScanIterator interface {
	io.Closer
	Scanner
	Next() bool
}

// Up from the current version.
func Up(ctx context.Context, db *DB, fsys fs.FS) error {
	err := migrate.Up(ctx, db.rwc, fsys)
	if err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

var _ ReadWriter = (*Tx)(nil)

type Tx struct {
	rwc *sql.Tx
}

// Rollback aborts the transaction.
func (tx *Tx) Rollback() error {
	return tx.rwc.Rollback()
}

// Commit commits the transaction.
func (tx *Tx) Commit() error {
	return tx.rwc.Commit()
}

// Exec executes a query without returning any rows.
func (tx *Tx) Exec(ctx context.Context, query string, args ...any) (int64, error) {
	rs, err := tx.rwc.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("execute: %w", err)
	}
	n, err := rs.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}
	return n, nil
}

// Query implements Tx.
func (tx *Tx) Query(ctx context.Context, query string, args ...any) (ScanIterator, error) {
	rs, err := tx.rwc.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	return rs, nil
}

// QueryRow executes a query that is expected to return at most one row.
func (tx *Tx) QueryRow(ctx context.Context, query string, args ...any) Scanner {
	return tx.rwc.QueryRowContext(ctx, query, args...)
}

type ReadWriter interface {
	Reader
	Writer
}

type Writer interface {
	Exec(ctx context.Context, query string, args ...any) (rowsAffected int64, err error)
}

type Reader interface {
	Query(ctx context.Context, query string, args ...any) (ScanIterator, error)
	QueryRow(ctx context.Context, query string, args ...any) Scanner
}

var (
	ErrNoRows = sql.ErrNoRows
)
