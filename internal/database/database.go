package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "modernc.org/sqlite"
)

type Options struct {
	Path           string
	BusyTimeout    time.Duration
	MaxConnections int
	ReadOnly       bool
}

func Open(ctx context.Context, opts Options) (*sql.DB, error) {
	q := url.Values{}
	q.Add("_pragma", "journal_mode(WAL)")
	q.Add("_pragma", "synchronous(NORMAL)")
	q.Add("_pragma", "foreign_keys(1)")
	q.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", opts.BusyTimeout.Milliseconds()))
	q.Set("_txlock", "immediate") // avoids SQLITE_BUSY on deferred->write upgrades

	dsn := "file:" + opts.Path + "?" + q.Encode()
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(opts.MaxConnections)
	db.SetMaxIdleConns(opts.MaxConnections)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(0)

	if err = db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to connect to sqlite database: %w", err)
	}
	return db, nil
}
