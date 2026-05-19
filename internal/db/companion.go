package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
)

//go:embed migrations/companion/*.sql
var companionMigrations embed.FS

// OpenCompanion opens (or creates) the per-franchise companion database at path
// and runs any pending migrations. The caller is responsible for closing the
// returned DB.
func OpenCompanion(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening companion DB: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pinging companion DB: %w", err)
	}
	if _, err := db.ExecContext(ctx, `PRAGMA journal_mode=WAL`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enabling WAL mode on companion DB: %w", err)
	}
	if err := runMigrations(ctx, db, companionMigrations, "migrations/companion"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("running companion migrations: %w", err)
	}
	return db, nil
}

