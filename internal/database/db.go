package database

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

// New opens a PostgreSQL connection using the given DSN and pings it.
func New(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Reasonable pool defaults; adjust for your workload.
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(1 * time.Hour)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
