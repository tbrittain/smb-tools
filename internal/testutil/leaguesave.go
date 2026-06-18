package testutil

import (
	"database/sql"
	"os"
	"testing"

	internaldb "smb-tools/internal/db"
)

// NewTestLeagueSaveDB creates an in-memory SQLite database seeded with the
// real per-league save schema columns that League Transfer reads and
// rewrites: the 6 GUID-bearing columns across 5 tables confirmed via
// PRAGMA foreign_key_list against a real league save (see
// docs/league-transfer/validation-results.md), plus enough surrounding
// structure (teams, divisions) to exercise league-overview introspection.
//
// This is deliberately a separate, narrower fixture from
// NewTestSaveGameDB — that fixture is tuned for franchise-tracking
// queries and its t_conferences table omits the real leagueGUID column
// (not needed there, since it only ever seeds one league). League
// Transfer's GUID rewrite must touch t_conferences.leagueGUID, so this
// fixture includes it rather than risk destabilizing the shared
// franchise-tracking fixture.
//
// Fixture data includes two independent leagues, so rewrite tests can
// assert the other league's rows are left untouched:
//   - League A (GUID AA…): one conference with one division (2 teams),
//     and a second conference with zero divisions (covers the "divisions
//     are optional" rule from docs/league-transfer/ux-flow.md), one
//     franchise, one season, one t_league_local_ids row.
//   - League B (GUID BB…): minimal — one conference, one franchise, one
//     season, one t_league_local_ids row — exists purely to prove
//     rewrites scoped to League A don't bleed into League B.
func NewTestLeagueSaveDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("testutil.NewTestLeagueSaveDB: open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := createLeagueSaveSchema(db); err != nil {
		t.Fatalf("testutil.NewTestLeagueSaveDB: schema: %v", err)
	}
	if err := seedLeagueSaveData(db); err != nil {
		t.Fatalf("testutil.NewTestLeagueSaveDB: seed: %v", err)
	}
	return db
}

// WriteCompressedLeagueSaveFixture writes a real, zlib-compressed
// league .sav file (the same schema/seed data as NewTestLeagueSaveDB) to
// destPath — for tests exercising code that reads actual .sav files from
// disk (decompression, zip packaging) rather than an in-memory connection.
func WriteCompressedLeagueSaveFixture(t *testing.T, destPath string) {
	t.Helper()
	tmpSqlitePath := destPath + ".tmp-source.sqlite"
	defer func() { _ = os.Remove(tmpSqlitePath) }()

	db, err := sql.Open("sqlite", tmpSqlitePath)
	if err != nil {
		t.Fatalf("testutil.WriteCompressedLeagueSaveFixture: open: %v", err)
	}
	if err := createLeagueSaveSchema(db); err != nil {
		t.Fatalf("testutil.WriteCompressedLeagueSaveFixture: schema: %v", err)
	}
	if err := seedLeagueSaveData(db); err != nil {
		t.Fatalf("testutil.WriteCompressedLeagueSaveFixture: seed: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("testutil.WriteCompressedLeagueSaveFixture: close: %v", err)
	}

	if err := internaldb.CompressFileAtomically(tmpSqlitePath, destPath); err != nil {
		t.Fatalf("testutil.WriteCompressedLeagueSaveFixture: compress: %v", err)
	}
}

func createLeagueSaveSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE t_leagues (
			GUID            BLOB PRIMARY KEY NOT NULL,
			originalGUID    BLOB,
			name            TEXT NOT NULL,
			allowedTeamType INTEGER NOT NULL DEFAULT 1
		);
		CREATE TABLE t_team_types (
			teamType INTEGER PRIMARY KEY NOT NULL,
			typeName TEXT NOT NULL
		);
		CREATE TABLE t_teams (
			GUID     BLOB PRIMARY KEY NOT NULL,
			teamName TEXT NOT NULL
		);
		CREATE TABLE t_conferences (
			GUID                  BLOB PRIMARY KEY NOT NULL,
			leagueGUID            BLOB NOT NULL REFERENCES t_leagues(GUID),
			name                  TEXT,
			usesDesignatedHitter  BOOLEAN NOT NULL DEFAULT 0
		);
		CREATE TABLE t_divisions (
			GUID           BLOB PRIMARY KEY NOT NULL,
			conferenceGUID BLOB NOT NULL REFERENCES t_conferences(GUID),
			name           TEXT
		);
		CREATE TABLE t_division_teams (
			divisionGUID BLOB NOT NULL REFERENCES t_divisions(GUID),
			teamGUID     BLOB NOT NULL REFERENCES t_teams(GUID)
		);
		CREATE TABLE t_franchise (
			GUID           BLOB PRIMARY KEY NOT NULL,
			leagueGUID     BLOB NOT NULL REFERENCES t_leagues(GUID),
			playerTeamGUID BLOB
		);
		CREATE TABLE t_seasons (
			id                   INTEGER PRIMARY KEY NOT NULL,
			GUID                 BLOB,
			historicalLeagueGUID BLOB NOT NULL REFERENCES t_leagues(GUID),
			elimination          INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE t_league_local_ids (
			localID INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			GUID    BLOB NOT NULL REFERENCES t_leagues(GUID)
		);
	`)
	return err
}

func seedLeagueSaveData(db *sql.DB) error {
	_, err := db.Exec(`
		INSERT INTO t_team_types (teamType, typeName) VALUES (1, 'Custom');

		-- League A: full structure, including a conference with zero divisions.
		INSERT INTO t_leagues (GUID, name, allowedTeamType) VALUES (X'AA000000000000000000000000000000', 'League A', 1);

		INSERT INTO t_teams (GUID, teamName) VALUES (X'01000000000000000000000000000000', 'A Team One');
		INSERT INTO t_teams (GUID, teamName) VALUES (X'02000000000000000000000000000000', 'A Team Two');

		INSERT INTO t_conferences (GUID, leagueGUID, name) VALUES (X'A1000000000000000000000000000000', X'AA000000000000000000000000000000', 'East');
		INSERT INTO t_divisions (GUID, conferenceGUID, name) VALUES (X'B1000000000000000000000000000000', X'A1000000000000000000000000000000', 'North');
		INSERT INTO t_division_teams (divisionGUID, teamGUID) VALUES (X'B1000000000000000000000000000000', X'01000000000000000000000000000000');
		INSERT INTO t_division_teams (divisionGUID, teamGUID) VALUES (X'B1000000000000000000000000000000', X'02000000000000000000000000000000');

		-- Second conference under League A with zero divisions — divisions are optional.
		INSERT INTO t_conferences (GUID, leagueGUID, name) VALUES (X'A2000000000000000000000000000000', X'AA000000000000000000000000000000', 'West');

		INSERT INTO t_franchise (GUID, leagueGUID, playerTeamGUID) VALUES (X'FA000000000000000000000000000000', X'AA000000000000000000000000000000', X'01000000000000000000000000000000');
		INSERT INTO t_seasons (id, GUID, historicalLeagueGUID) VALUES (100, X'DA000000000000000000000000000000', X'AA000000000000000000000000000000');
		INSERT INTO t_league_local_ids (GUID) VALUES (X'AA000000000000000000000000000000');

		-- League B: minimal, independent — proves rewrites scoped to League A don't bleed over.
		INSERT INTO t_leagues (GUID, name, allowedTeamType) VALUES (X'BB000000000000000000000000000000', 'League B', 1);
		INSERT INTO t_conferences (GUID, leagueGUID, name) VALUES (X'B2000000000000000000000000000000', X'BB000000000000000000000000000000', 'Only Conference');
		INSERT INTO t_franchise (GUID, leagueGUID) VALUES (X'FB000000000000000000000000000000', X'BB000000000000000000000000000000');
		INSERT INTO t_seasons (id, GUID, historicalLeagueGUID) VALUES (200, X'DB000000000000000000000000000000', X'BB000000000000000000000000000000');
		INSERT INTO t_league_local_ids (GUID) VALUES (X'BB000000000000000000000000000000');
	`)
	return err
}
