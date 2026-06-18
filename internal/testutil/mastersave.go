package testutil

import (
	"database/sql"
	"os"
	"testing"

	internaldb "smb-tools/internal/db"
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

// WriteCompressedMasterSaveFixture writes a real, zlib-compressed
// master.sav file (the same schema/seed data as NewTestMasterSaveDB) to
// destPath. If extraGUIDs are given, each is inserted as an additional
// registered league (isMissing = 0) beyond the two the fixture seeds by
// default — useful for setting up collision-detection test cases.
func WriteCompressedMasterSaveFixture(t *testing.T, destPath string, extraGUIDs ...[16]byte) {
	t.Helper()
	tmpSqlitePath := destPath + ".tmp-source.sqlite"
	defer func() { _ = os.Remove(tmpSqlitePath) }()

	db, err := sql.Open("sqlite", tmpSqlitePath)
	if err != nil {
		t.Fatalf("testutil.WriteCompressedMasterSaveFixture: open: %v", err)
	}
	if err := createMasterSaveSchema(db); err != nil {
		t.Fatalf("testutil.WriteCompressedMasterSaveFixture: schema: %v", err)
	}
	if err := seedMasterSaveData(db); err != nil {
		t.Fatalf("testutil.WriteCompressedMasterSaveFixture: seed: %v", err)
	}
	for _, guid := range extraGUIDs {
		if _, err := db.Exec(`INSERT INTO t_league_savedatas (GUID, isMissing) VALUES (?, 0)`, guid[:]); err != nil {
			t.Fatalf("testutil.WriteCompressedMasterSaveFixture: inserting extra GUID: %v", err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatalf("testutil.WriteCompressedMasterSaveFixture: close: %v", err)
	}

	if err := internaldb.CompressFileAtomically(tmpSqlitePath, destPath); err != nil {
		t.Fatalf("testutil.WriteCompressedMasterSaveFixture: compress: %v", err)
	}
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
