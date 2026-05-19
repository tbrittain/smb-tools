package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	_ "modernc.org/sqlite" // registers the "sqlite" driver
)

//go:embed migrations/registry/*.sql
var registryMigrations embed.FS

// OpenRegistry opens (or creates) the registry database at path and runs any
// pending migrations. The caller is responsible for closing the returned DB.
func OpenRegistry(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening registry DB: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pinging registry DB: %w", err)
	}
	if _, err := db.ExecContext(ctx, `PRAGMA journal_mode=WAL`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enabling WAL mode on registry DB: %w", err)
	}
	if err := runMigrations(ctx, db, registryMigrations, "migrations/registry"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("running registry migrations: %w", err)
	}
	return db, nil
}

