package testutil

import (
	"database/sql"
	"testing"
)

// NewTestSaveGameDB creates an in-memory SQLite database seeded with the core
// SMB save game schema and a representative set of synthetic data. Use this as
// the backing store for SqliteSaveGameReader in unit and integration tests.
//
// Fixture data includes:
//   - 1 league / franchise
//   - 2 seasons (IDs 100 and 101)
//   - 2 teams (Home Squad, Away Crew)
//   - 2 players: a position player (batter, GUID AA…) and a pitcher (BB…)
//   - Batting and pitching stats for both seasons, regular season and playoffs
//   - A regular season schedule and a playoff game
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

		-- Two seasons for multi-season tracking tests
		INSERT INTO t_franchise_seasons (seasonID, franchiseId) VALUES (100, 1);
		INSERT INTO t_franchise_seasons (seasonID, franchiseId) VALUES (101, 1);

		-- Teams
		INSERT INTO t_teams (GUID, teamName) VALUES (X'01000000000000000000000000000000', 'Home Squad');
		INSERT INTO t_teams (GUID, teamName) VALUES (X'02000000000000000000000000000000', 'Away Crew');
		INSERT INTO t_team_local_ids (GUID) VALUES (X'01000000000000000000000000000000'); -- localID 1
		INSERT INTO t_team_local_ids (GUID) VALUES (X'02000000000000000000000000000000'); -- localID 2

		-- Conferences and divisions
		INSERT INTO t_conferences (conferenceName) VALUES ('East Conference');
		INSERT INTO t_divisions (divisionName, conferenceId) VALUES ('East Division', 1);
		INSERT INTO t_division_teams (teamLocalId, divisionId) VALUES (1, 1);
		INSERT INTO t_division_teams (teamLocalId, divisionId) VALUES (2, 1);

		-- Player AA: outfielder (batter)
		INSERT INTO t_baseball_players (GUID, power, contact, speed, fielding, arm, velocity, junk, accuracy, age)
		VALUES (X'AA000000000000000000000000000000', 80, 75, 60, 70, 65, 50, 50, 50, 27);
		INSERT INTO t_baseball_player_local_ids (GUID) VALUES (X'AA000000000000000000000000000000'); -- localID 1
		INSERT INTO t_baseball_player_traits (baseballPlayerGUID, traits) VALUES (X'AA000000000000000000000000000000', '[]');
		INSERT INTO t_salary (baseballPlayerGUID, salary) VALUES (X'AA000000000000000000000000000000', 250);

		-- Player BB: pitcher
		INSERT INTO t_baseball_players (GUID, power, contact, speed, fielding, arm, velocity, junk, accuracy, age)
		VALUES (X'BB000000000000000000000000000000', 40, 40, 40, 50, 55, 88, 78, 82, 30);
		INSERT INTO t_baseball_player_local_ids (GUID) VALUES (X'BB000000000000000000000000000000'); -- localID 2
		INSERT INTO t_baseball_player_traits (baseballPlayerGUID, traits) VALUES (X'BB000000000000000000000000000000', '[]');
		INSERT INTO t_salary (baseballPlayerGUID, salary) VALUES (X'BB000000000000000000000000000000', 300);

		-- ── Season 100 stats ─────────────────────────────────────────────────

		-- Batter (AA) — regular season
		INSERT INTO t_stats (aggregatorID, currentTeamName) VALUES (1, 'Home Squad');
		INSERT INTO t_stats_players (aggregatorID, baseballPlayerGUIDIfKnown, firstName, lastName, primaryPosition, age)
		VALUES (1, X'AA000000000000000000000000000000', 'Test', 'Batter', 'CF', 27);
		INSERT INTO t_stats_batting (aggregatorID, gamesPlayed, gamesBatting, atBats, runs, hits, doubles, triples, homeruns, rbi, baseOnBalls, strikeOuts)
		VALUES (1, 50, 50, 180, 30, 54, 10, 2, 12, 40, 20, 35);
		INSERT INTO t_season_stats (aggregatorID, seasonID) VALUES (1, 100);

		-- Pitcher (BB) — regular season
		INSERT INTO t_stats (aggregatorID, currentTeamName) VALUES (2, 'Home Squad');
		INSERT INTO t_stats_players (aggregatorID, baseballPlayerGUIDIfKnown, firstName, lastName, primaryPosition, pitcherRole, age)
		VALUES (2, X'BB000000000000000000000000000000', 'Test', 'Pitcher', 'P', 'SP', 30);
		INSERT INTO t_stats_pitching (aggregatorID, wins, losses, games, gamesStarted, outsPitched, hits, earnedRuns, homeRuns, baseOnBalls, strikeOuts, battersFaced, totalPitches)
		VALUES (2, 12, 8, 25, 25, 540, 140, 55, 15, 40, 180, 740, 3200);
		INSERT INTO t_season_stats (aggregatorID, seasonID) VALUES (2, 100);

		-- Batter (AA) — playoff
		INSERT INTO t_stats (aggregatorID, currentTeamName) VALUES (3, 'Home Squad');
		INSERT INTO t_stats_players (aggregatorID, baseballPlayerGUIDIfKnown, firstName, lastName, primaryPosition, age)
		VALUES (3, X'AA000000000000000000000000000000', 'Test', 'Batter', 'CF', 27);
		INSERT INTO t_stats_batting (aggregatorID, gamesPlayed, gamesBatting, atBats, hits, homeruns, rbi)
		VALUES (3, 5, 5, 18, 6, 2, 5);

		-- Pitcher (BB) — playoff
		INSERT INTO t_stats (aggregatorID, currentTeamName) VALUES (4, 'Home Squad');
		INSERT INTO t_stats_players (aggregatorID, baseballPlayerGUIDIfKnown, firstName, lastName, primaryPosition, pitcherRole, age)
		VALUES (4, X'BB000000000000000000000000000000', 'Test', 'Pitcher', 'P', 'SP', 30);
		INSERT INTO t_stats_pitching (aggregatorID, wins, losses, games, gamesStarted, outsPitched, hits, earnedRuns, strikeOuts)
		VALUES (4, 2, 0, 2, 2, 54, 10, 3, 18);

		-- Career stats (aggregators 1 and 2)
		INSERT INTO t_career_season_stats (aggregatorID) VALUES (1);
		INSERT INTO t_career_season_stats (aggregatorID) VALUES (2);

		-- ── Season 100 schedule ──────────────────────────────────────────────

		INSERT INTO t_season_schedule (gameNumber, day, homeTeamID, awayTeamID) VALUES (1, 1, 1, 2);
		INSERT INTO t_game_results (gameNumber, homeRunsScored, awayRunsScored, homePitcherLocalID, awayPitcherLocalID)
		VALUES (1, 5, 3, 2, 1);
		INSERT INTO t_season_games (gameNumber, seasonID) VALUES (1, 100);

		INSERT INTO t_season_schedule (gameNumber, day, homeTeamID, awayTeamID) VALUES (2, 2, 2, 1);
		INSERT INTO t_game_results (gameNumber, homeRunsScored, awayRunsScored, homePitcherLocalID, awayPitcherLocalID)
		VALUES (2, 1, 4, 1, 2);
		INSERT INTO t_season_games (gameNumber, seasonID) VALUES (2, 100);

		-- ── Season 100 playoffs ───────────────────────────────────────────────

		INSERT INTO t_playoffs (GUID, seasonGUID) VALUES (X'CC000000000000000000000000000000', X'DD000000000000000000000000000000');
		INSERT INTO t_playoff_series (playoffGUID, seriesNumber, team1GUID, team2GUID, team1Standing, team2Standing)
		VALUES (X'CC000000000000000000000000000000', 1, X'01000000000000000000000000000000', X'02000000000000000000000000000000', 1, 2);

		INSERT INTO t_season_schedule (gameNumber, day, homeTeamID, awayTeamID) VALUES (100, 1, 1, 2);
		INSERT INTO t_game_results (gameNumber, homeRunsScored, awayRunsScored, homePitcherLocalID, awayPitcherLocalID)
		VALUES (100, 3, 1, 2, 1);
		INSERT INTO t_season_games (gameNumber, seasonID) VALUES (100, 100);

		-- ── Season 101 stats (same players, second season) ───────────────────

		INSERT INTO t_stats (aggregatorID, currentTeamName) VALUES (5, 'Home Squad');
		INSERT INTO t_stats_players (aggregatorID, baseballPlayerGUIDIfKnown, firstName, lastName, primaryPosition, age)
		VALUES (5, X'AA000000000000000000000000000000', 'Test', 'Batter', 'CF', 28);
		INSERT INTO t_stats_batting (aggregatorID, gamesPlayed, gamesBatting, atBats, runs, hits, homeruns, rbi)
		VALUES (5, 52, 52, 190, 35, 60, 15, 48);
		INSERT INTO t_season_stats (aggregatorID, seasonID) VALUES (5, 101);

		INSERT INTO t_stats (aggregatorID, currentTeamName) VALUES (6, 'Home Squad');
		INSERT INTO t_stats_players (aggregatorID, baseballPlayerGUIDIfKnown, firstName, lastName, primaryPosition, pitcherRole, age)
		VALUES (6, X'BB000000000000000000000000000000', 'Test', 'Pitcher', 'P', 'SP', 31);
		INSERT INTO t_stats_pitching (aggregatorID, wins, losses, games, gamesStarted, outsPitched, hits, earnedRuns, strikeOuts)
		VALUES (6, 14, 7, 25, 25, 570, 130, 50, 195);
		INSERT INTO t_season_stats (aggregatorID, seasonID) VALUES (6, 101);

		INSERT INTO t_season_schedule (gameNumber, day, homeTeamID, awayTeamID) VALUES (3, 1, 1, 2);
		INSERT INTO t_game_results (gameNumber, homeRunsScored, awayRunsScored) VALUES (3, 6, 2);
		INSERT INTO t_season_games (gameNumber, seasonID) VALUES (3, 101);
	`)
	return err
}
