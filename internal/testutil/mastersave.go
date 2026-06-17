package testutil

import (
	"database/sql"
	"testing"
)

// NewTestMasterSaveDB creates an in-memory SQLite database seeded with the
// real master.sav schema, verified directly against a live, decompressed
// master.sav during the league-transfer research/validation work — see
// docs/domain/master-save-schema.md for the full schema and how it was
// confirmed. Only t_league_savedatas is seeded with data; the feature reads
// and writes nothing else in master.sav.
//
// Fixture data includes:
//   - 2 leagues already registered (isMissing = 0), with GUIDs stored as
//     real 16-byte blobs (not text) to make any accidental regression to the
//     legacy string-binding bug immediately visible in test failures.
func NewTestMasterSaveDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("testutil.NewTestMasterSaveDB: open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := createMasterSaveSchema(db); err != nil {
		t.Fatalf("testutil.NewTestMasterSaveDB: schema: %v", err)
	}
	if err := seedMasterSaveData(db); err != nil {
		t.Fatalf("testutil.NewTestMasterSaveDB: seed: %v", err)
	}
	return db
}

func createMasterSaveSchema(db *sql.DB) error {
	// Schema mirrors the real master.sav, confirmed via PRAGMA/sqlite_master
	// against a live, decompressed master.sav. Only t_league_savedatas is
	// reproduced in full; other master.sav tables are out of scope for
	// League Transfer and are not needed by any test using this fixture.
	_, err := db.Exec(`
		CREATE TABLE t_league_savedatas (
			GUID      BLOB NOT NULL PRIMARY KEY,
			isMissing BOOL NOT NULL DEFAULT 0
		);
	`)
	return err
}

func seedMasterSaveData(db *sql.DB) error {
	_, err := db.Exec(`
		INSERT INTO t_league_savedatas (GUID, isMissing) VALUES
			(X'99F30082775B4547ADD88C7D2C94FCE5', 0),
			(X'1EE40D82453A474082E50827731C22E0', 0);
	`)
	return err
}
