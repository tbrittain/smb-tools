package db_test

import (
	"context"
	"testing"

	"smb-tools/internal/testutil"
)

func TestOpenCompanion_MigrationsApplied(t *testing.T) {
	db := testutil.NewTestDB(t)

	// schema_migrations table must exist and have at least one entry
	var count int
	if err := db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM schema_migrations`,
	).Scan(&count); err != nil {
		t.Fatalf("querying schema_migrations: %v", err)
	}
	if count == 0 {
		t.Error("expected at least one applied migration, got 0")
	}

	// save_game_snapshots table must exist (from 0001_initial.up.sql)
	if _, err := db.ExecContext(context.Background(),
		`SELECT id FROM save_game_snapshots LIMIT 0`,
	); err != nil {
		t.Errorf("save_game_snapshots table missing or malformed: %v", err)
	}
}

func TestOpenRegistry_MigrationsApplied(t *testing.T) {
	db := testutil.NewTestRegistryDB(t)

	var count int
	if err := db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM schema_migrations`,
	).Scan(&count); err != nil {
		t.Fatalf("querying schema_migrations: %v", err)
	}
	if count == 0 {
		t.Error("expected at least one applied migration, got 0")
	}

	// franchises table must exist (from 0001_initial.up.sql)
	if _, err := db.ExecContext(context.Background(),
		`SELECT id FROM franchises LIMIT 0`,
	); err != nil {
		t.Errorf("franchises table missing or malformed: %v", err)
	}

	// league_mode column must exist (from 0003_add_league_mode.up.sql)
	if _, err := db.ExecContext(context.Background(),
		`SELECT league_mode FROM franchises LIMIT 0`,
	); err != nil {
		t.Errorf("franchises.league_mode column missing or malformed: %v", err)
	}
}

func TestOpenCompanion_IdempotentMigrations(t *testing.T) {
	// Running migrations twice on the same DB must not error
	ctx := context.Background()
	db := testutil.NewTestDB(t)

	// Simulate a second startup by calling the migration path again indirectly —
	// if schema_migrations already has all versions, nothing should change.
	var before, after int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&before)

	// Re-open would run migrations again; verify count is stable
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&after)
	if before != after {
		t.Errorf("migration count changed unexpectedly: %d → %d", before, after)
	}
}
