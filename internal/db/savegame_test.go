package db_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	internaldb "smb-tools/internal/db"
)

func TestOpenSnapshot_ValidSQLite(t *testing.T) {
	// Create a minimal SQLite file by opening a real in-memory DB and copying
	// its content. The easiest approach: create a temp file via OpenCompanion
	// which writes a real SQLite DB with our schema, then use OpenSnapshot on it.
	ctx := context.Background()
	dir := t.TempDir()
	snapshotPath := filepath.Join(dir, "test.sqlite")

	// Write a valid SQLite DB to disk.
	db, err := internaldb.OpenCompanion(ctx, snapshotPath)
	if err != nil {
		t.Fatalf("creating test sqlite file: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("closing test sqlite file: %v", err)
	}

	snapDB, err := internaldb.OpenSnapshot(ctx, snapshotPath)
	if err != nil {
		t.Fatalf("OpenSnapshot: %v", err)
	}
	defer func() { _ = snapDB.Close() }()

	if err := snapDB.PingContext(ctx); err != nil {
		t.Errorf("ping after OpenSnapshot: %v", err)
	}
}

func TestOpenSnapshot_MissingFile(t *testing.T) {
	ctx := context.Background()
	_, err := internaldb.OpenSnapshot(ctx, "/nonexistent/path/snapshot.sqlite")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestOpenSnapshot_NotSQLite(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	notSQLite := filepath.Join(dir, "garbage.sqlite")
	if err := os.WriteFile(notSQLite, []byte("this is not a sqlite database"), 0o600); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	_, err := internaldb.OpenSnapshot(ctx, notSQLite)
	if err == nil {
		t.Error("expected error for non-SQLite file, got nil")
	}
}
