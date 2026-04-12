package db

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// Open creates a SQLite connection. DATABASE_URL must be a SQLite URI, e.g.
// file:./data/call_booking.db?_foreign_keys=on&_journal_mode=WAL
// If unset, defaults to file:./data/call_booking.db with foreign keys and WAL.
func Open(ctx context.Context) (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "file:./data/call_booking.db?_foreign_keys=on&_journal_mode=WAL"
	}
	if err := ensureSQLiteDir(dsn); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite: single writer; pool size 1 avoids locking issues
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func ensureSQLiteDir(dsn string) error {
	if !strings.HasPrefix(dsn, "file:") {
		return nil
	}
	rest := strings.TrimPrefix(dsn, "file:")
	pathPart := rest
	if i := strings.IndexAny(rest, "?"); i >= 0 {
		pathPart = rest[:i]
	}
	// ":memory:" and similar have no parent dir
	if pathPart == "" || strings.HasPrefix(pathPart, ":memory:") {
		return nil
	}
	dir := filepath.Dir(pathPart)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}
