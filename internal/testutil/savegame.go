package testutil

import (
	"database/sql"
	"testing"
)

// NewTestSaveGameDB creates an in-memory SQLite database seeded with the core
// SMB save game schema and a minimal set of synthetic data. Use this as the
// backing store for SqliteSaveGameReader in unit and integration tests.
//
// The schema mirrors the actual SMB save game structure documented in
// docs/domain/save-game-schema.md. Synthetic data uses clearly fake values
// (player names like "Test Player", GUIDs like X'0000...') so tests are
// self-contained and never depend on real game files.
func NewTestSaveGameDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("testutil.NewTestSaveGameDB: open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := createSaveGameSchema(db); err != nil {
		t.Fatalf("testutil.NewTestSaveGameDB: schema: %v", err)
	}
	if err := seedSaveGameData(db); err != nil {
		t.Fatalf("testutil.NewTestSaveGameDB: seed: %v", err)
	}
	return db
}

func createSaveGameSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE t_leagues (
			leagueId          INTEGER PRIMARY KEY NOT NULL,
			leagueName        TEXT NOT NULL,
			leagueTeamTypeId  INTEGER NOT NULL
		);
		CREATE TABLE t_team_types (
			teamType  INTEGER PRIMARY KEY NOT NULL,
			typeName  TEXT NOT NULL
		);
		CREATE TABLE t_franchise (
			franchiseId   INTEGER PRIMARY KEY NOT NULL,
			leagueId      INTEGER NOT NULL REFERENCES t_leagues(leagueId)
		);
		CREATE TABLE t_franchise_seasons (
			seasonID     INTEGER PRIMARY KEY NOT NULL,
			franchiseId  INTEGER NOT NULL REFERENCES t_franchise(franchiseId)
		);
		CREATE TABLE t_teams (
			GUID      BLOB PRIMARY KEY NOT NULL,
			teamName  TEXT NOT NULL
		);
		CREATE TABLE t_team_local_ids (
			localID  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			GUID     BLOB NOT NULL REFERENCES t_teams(GUID)
		);
		CREATE TABLE t_conferences (
			rowid          INTEGER PRIMARY KEY AUTOINCREMENT,
			conferenceName TEXT NOT NULL
		);
		CREATE TABLE t_divisions (
			rowid        INTEGER PRIMARY KEY AUTOINCREMENT,
			divisionName TEXT NOT NULL,
			conferenceId INTEGER NOT NULL REFERENCES t_conferences(rowid)
		);
		CREATE TABLE t_division_teams (
			teamLocalId INTEGER NOT NULL,
			divisionId  INTEGER NOT NULL
		);
		CREATE TABLE t_baseball_players (
			GUID      BLOB PRIMARY KEY NOT NULL,
			power     INTEGER NOT NULL DEFAULT 50,
			contact   INTEGER NOT NULL DEFAULT 50,
			speed     INTEGER NOT NULL DEFAULT 50,
			fielding  INTEGER NOT NULL DEFAULT 50,
			arm       INTEGER NOT NULL DEFAULT 50,
			velocity  INTEGER NOT NULL DEFAULT 50,
			junk      INTEGER NOT NULL DEFAULT 50,
			accuracy  INTEGER NOT NULL DEFAULT 50,
			age       INTEGER NOT NULL DEFAULT 25
		);
		CREATE TABLE t_baseball_player_local_ids (
			localID  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			GUID     BLOB NOT NULL REFERENCES t_baseball_players(GUID)
		);
		CREATE TABLE t_baseball_player_traits (
			baseballPlayerGUID BLOB NOT NULL REFERENCES t_baseball_players(GUID),
			traits             TEXT NOT NULL DEFAULT '[]'
		);
		CREATE TABLE t_salary (
			baseballPlayerGUID BLOB NOT NULL REFERENCES t_baseball_players(GUID),
			salary             INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE t_stats (
			aggregatorID              INTEGER PRIMARY KEY NOT NULL,
			currentTeamName           TEXT,
			mostRecentTeamName        TEXT,
			secondMostRecentTeamName  TEXT
		);
		CREATE TABLE t_stats_players (
			aggregatorID                INTEGER PRIMARY KEY REFERENCES t_stats(aggregatorID),
			baseballPlayerGUIDIfKnown  BLOB REFERENCES t_baseball_players(GUID),
			firstName                  TEXT,
			lastName                   TEXT,
			primaryPosition            TEXT,
			secondaryPosition          TEXT,
			pitcherRole                TEXT,
			age                        INTEGER,
			retirementSeason           INTEGER
		);
		CREATE TABLE t_stats_batting (
			aggregatorID   INTEGER PRIMARY KEY REFERENCES t_stats(aggregatorID),
			gamesPlayed    INTEGER NOT NULL DEFAULT 0,
			gamesBatting   INTEGER NOT NULL DEFAULT 0,
			atBats         INTEGER NOT NULL DEFAULT 0,
			runs           INTEGER NOT NULL DEFAULT 0,
			hits           INTEGER NOT NULL DEFAULT 0,
			doubles        INTEGER NOT NULL DEFAULT 0,
			triples        INTEGER NOT NULL DEFAULT 0,
			homeruns       INTEGER NOT NULL DEFAULT 0,
			rbi            INTEGER NOT NULL DEFAULT 0,
			stolenBases    INTEGER NOT NULL DEFAULT 0,
			caughtStealing INTEGER NOT NULL DEFAULT 0,
			baseOnBalls    INTEGER NOT NULL DEFAULT 0,
			strikeOuts     INTEGER NOT NULL DEFAULT 0,
			hitByPitch     INTEGER NOT NULL DEFAULT 0,
			sacrificeHits  INTEGER NOT NULL DEFAULT 0,
			sacrificeFlies INTEGER NOT NULL DEFAULT 0,
			errors         INTEGER NOT NULL DEFAULT 0,
			passedBalls    INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE t_stats_pitching (
			aggregatorID       INTEGER PRIMARY KEY REFERENCES t_stats(aggregatorID),
			wins               INTEGER NOT NULL DEFAULT 0,
			losses             INTEGER NOT NULL DEFAULT 0,
			games              INTEGER NOT NULL DEFAULT 0,
			gamesStarted       INTEGER NOT NULL DEFAULT 0,
			completeGames      INTEGER NOT NULL DEFAULT 0,
			totalPitches       INTEGER NOT NULL DEFAULT 0,
			shutouts           INTEGER NOT NULL DEFAULT 0,
			saves              INTEGER NOT NULL DEFAULT 0,
			outsPitched        INTEGER NOT NULL DEFAULT 0,
			hits               INTEGER NOT NULL DEFAULT 0,
			earnedRuns         INTEGER NOT NULL DEFAULT 0,
			homeRuns           INTEGER NOT NULL DEFAULT 0,
			baseOnBalls        INTEGER NOT NULL DEFAULT 0,
			strikeOuts         INTEGER NOT NULL DEFAULT 0,
			battersHitByPitch  INTEGER NOT NULL DEFAULT 0,
			battersFaced       INTEGER NOT NULL DEFAULT 0,
			gamesFinished      INTEGER NOT NULL DEFAULT 0,
			runsAllowed        INTEGER NOT NULL DEFAULT 0,
			wildPitches        INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE t_season_stats (
			aggregatorID INTEGER NOT NULL REFERENCES t_stats(aggregatorID),
			seasonID     INTEGER NOT NULL REFERENCES t_franchise_seasons(seasonID)
		);
		CREATE TABLE t_career_season_stats (
			aggregatorID INTEGER NOT NULL REFERENCES t_stats(aggregatorID)
		);
		CREATE TABLE t_season_schedule (
			gameNumber  INTEGER PRIMARY KEY NOT NULL,
			day         INTEGER NOT NULL DEFAULT 1,
			homeTeamID  INTEGER NOT NULL,
			awayTeamID  INTEGER NOT NULL
		);
		CREATE TABLE t_game_results (
			gameNumber         INTEGER PRIMARY KEY NOT NULL,
			homeRunsScored     INTEGER,
			awayRunsScored     INTEGER,
			homePitcherLocalID INTEGER,
			awayPitcherLocalID INTEGER
		);
		CREATE TABLE t_season_games (
			gameNumber INTEGER NOT NULL,
			seasonID   INTEGER NOT NULL
		);
		CREATE TABLE t_playoffs (
			GUID       BLOB PRIMARY KEY NOT NULL,
			seasonGUID BLOB NOT NULL
		);
		CREATE TABLE t_playoff_series (
			playoffGUID   BLOB NOT NULL REFERENCES t_playoffs(GUID),
			seriesNumber  INTEGER NOT NULL,
			team1GUID     BLOB NOT NULL,
			team2GUID     BLOB NOT NULL,
			team1Standing INTEGER NOT NULL DEFAULT 1,
			team2Standing INTEGER NOT NULL DEFAULT 2
		);
		CREATE VIEW v_season_standings AS
			SELECT
				homeTeamID   AS teamLocalId,
				seasonID,
				SUM(CASE WHEN gr.homeRunsScored > gr.awayRunsScored THEN 1 ELSE 0 END) AS gamesWon,
				SUM(CASE WHEN gr.homeRunsScored < gr.awayRunsScored THEN 1 ELSE 0 END) AS gamesLost,
				0.0 AS gamesBack,
				SUM(COALESCE(gr.homeRunsScored, 0)) AS runsFor,
				SUM(COALESCE(gr.awayRunsScored, 0)) AS runsAgainst
			FROM t_season_games sg
			JOIN t_season_schedule sc ON sc.gameNumber = sg.gameNumber
			LEFT JOIN t_game_results gr ON gr.gameNumber = sc.gameNumber
			GROUP BY homeTeamID, seasonID;
	`)
	return err
}

func seedSaveGameData(db *sql.DB) error {
	_, err := db.Exec(`
		INSERT INTO t_team_types (teamType, typeName) VALUES (1, 'franchise');

		INSERT INTO t_leagues (leagueId, leagueName, leagueTeamTypeId)
		VALUES (1, 'Test Franchise League', 1);

		INSERT INTO t_franchise (franchiseId, leagueId) VALUES (1, 1);

		INSERT INTO t_franchise_seasons (seasonID, franchiseId) VALUES (100, 1);

		-- Teams
		INSERT INTO t_teams (GUID, teamName) VALUES (X'01000000000000000000000000000000', 'Home Squad');
		INSERT INTO t_teams (GUID, teamName) VALUES (X'02000000000000000000000000000000', 'Away Crew');
		INSERT INTO t_team_local_ids (GUID) VALUES (X'01000000000000000000000000000000');
		INSERT INTO t_team_local_ids (GUID) VALUES (X'02000000000000000000000000000000');

		-- Conferences and divisions
		INSERT INTO t_conferences (conferenceName) VALUES ('East Conference');
		INSERT INTO t_divisions (divisionName, conferenceId) VALUES ('East Division', 1);
		INSERT INTO t_division_teams (teamLocalId, divisionId) VALUES (1, 1);
		INSERT INTO t_division_teams (teamLocalId, divisionId) VALUES (2, 1);

		-- Players
		INSERT INTO t_baseball_players (GUID, power, contact, speed, fielding, arm, velocity, junk, accuracy, age)
		VALUES (X'AA000000000000000000000000000000', 80, 75, 60, 70, 65, 85, 70, 80, 27);
		INSERT INTO t_baseball_player_local_ids (GUID) VALUES (X'AA000000000000000000000000000000');
		INSERT INTO t_baseball_player_traits (baseballPlayerGUID, traits) VALUES (X'AA000000000000000000000000000000', '[]');
		INSERT INTO t_salary (baseballPlayerGUID, salary) VALUES (X'AA000000000000000000000000000000', 250);

		-- Stats
		INSERT INTO t_stats (aggregatorID, currentTeamName) VALUES (1, 'Home Squad');
		INSERT INTO t_stats_players (aggregatorID, baseballPlayerGUIDIfKnown, firstName, lastName, primaryPosition, age)
		VALUES (1, X'AA000000000000000000000000000000', 'Test', 'Player', 'CF', 27);
		INSERT INTO t_stats_batting (aggregatorID, gamesPlayed, atBats, hits, homeruns, rbi)
		VALUES (1, 50, 180, 54, 12, 40);
		INSERT INTO t_season_stats (aggregatorID, seasonID) VALUES (1, 100);
		INSERT INTO t_career_season_stats (aggregatorID) VALUES (1);

		-- Schedule
		INSERT INTO t_season_schedule (gameNumber, day, homeTeamID, awayTeamID) VALUES (1, 1, 1, 2);
		INSERT INTO t_game_results (gameNumber, homeRunsScored, awayRunsScored) VALUES (1, 5, 3);
		INSERT INTO t_season_games (gameNumber, seasonID) VALUES (1, 100);
	`)
	return err
}
