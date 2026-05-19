package testutil

import (
	"context"
	"database/sql"
	"testing"

	"smb-tools/internal/db"
)

// NewTestDB creates an in-memory companion SQLite database with all migrations
// applied. The database is automatically closed when the test ends.
//
// Use this for integration tests against the companion schema.
func NewTestDB(t *testing.T) *sql.DB {
	t.Helper()
	database, err := db.OpenCompanion(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("testutil.NewTestDB: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("testutil.NewTestDB cleanup: %v", err)
		}
	})
	return database
}

// NewTestRegistryDB creates an in-memory registry SQLite database with all
// migrations applied. The database is automatically closed when the test ends.
func NewTestRegistryDB(t *testing.T) *sql.DB {
	t.Helper()
	database, err := db.OpenRegistry(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("testutil.NewTestRegistryDB: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("testutil.NewTestRegistryDB cleanup: %v", err)
		}
	})
	return database
}
