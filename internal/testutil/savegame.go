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
		-- Schema mirrors the real SMB4 save game, confirmed against SMB3Explorer SQL.
		CREATE TABLE t_leagues (
			GUID              BLOB PRIMARY KEY NOT NULL,
			name              TEXT NOT NULL,
			allowedTeamType   INTEGER NOT NULL,
			originalGUID      BLOB
		);
		CREATE TABLE t_team_types (
			teamType  INTEGER PRIMARY KEY NOT NULL,
			typeName  TEXT NOT NULL
		);
		CREATE TABLE t_franchise (
			GUID            BLOB PRIMARY KEY NOT NULL,
			leagueGUID      BLOB NOT NULL REFERENCES t_leagues(GUID),
			playerTeamGUID  BLOB
		);
		CREATE TABLE t_seasons (
			id                    INTEGER PRIMARY KEY NOT NULL,
			GUID                  BLOB,
			historicalLeagueGUID  BLOB NOT NULL REFERENCES t_leagues(GUID),
			elimination           INTEGER NOT NULL DEFAULT 0
		);
		-- t_franchise_seasons retained for structural completeness; no longer
		-- used by any reader query (GetCurrentSeason uses t_seasons directly).
		CREATE TABLE t_franchise_seasons (
			seasonID     INTEGER NOT NULL,
			franchiseGUID BLOB NOT NULL
		);
		CREATE TABLE t_teams (
			GUID      BLOB PRIMARY KEY NOT NULL,
			teamName  TEXT NOT NULL
		);
		CREATE TABLE t_team_local_ids (
			localID  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			GUID     BLOB NOT NULL REFERENCES t_teams(GUID)
		);
		-- Real game: t_conferences/t_divisions use GUID blobs as PKs, name TEXT.
		CREATE TABLE t_conferences (
			GUID  BLOB PRIMARY KEY NOT NULL,
			name  TEXT NOT NULL
		);
		CREATE TABLE t_divisions (
			GUID           BLOB PRIMARY KEY NOT NULL,
			name           TEXT NOT NULL,
			conferenceGUID BLOB NOT NULL REFERENCES t_conferences(GUID)
		);
		-- Real game: t_division_teams joins by team GUID, not local ID.
		CREATE TABLE t_division_teams (
			teamGUID     BLOB NOT NULL REFERENCES t_teams(GUID),
			divisionGUID BLOB NOT NULL REFERENCES t_divisions(GUID)
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
		-- SMB4-only player attributes stored as key-value pairs.
		-- optionKey 4 = throw hand, 5 = bat hand, 107 = chemistry type.
		CREATE TABLE t_baseball_player_options (
			baseballPlayerLocalID INTEGER NOT NULL REFERENCES t_baseball_player_local_ids(localID),
			optionKey             INTEGER NOT NULL,
			optionValue           INTEGER NOT NULL
		);
		-- Real game: one row per trait; trait and subType are integer IDs.
		-- The JSON representation is assembled at query time via json_group_array.
		CREATE TABLE t_baseball_player_traits (
			baseballPlayerLocalID INTEGER NOT NULL REFERENCES t_baseball_player_local_ids(localID),
			trait                 INTEGER NOT NULL,
			subType               INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE t_salary (
			baseballPlayerGUID BLOB NOT NULL REFERENCES t_baseball_players(GUID),
			salary             INTEGER NOT NULL DEFAULT 0
		);
		CREATE TABLE t_stats (
			aggregatorID                       INTEGER PRIMARY KEY NOT NULL,
			statsPlayerID                      INTEGER,
			currentTeamLocalID                 INTEGER,
			mostRecentlyPlayedTeamLocalID      INTEGER,
			previousRecentlyPlayedTeamLocalID  INTEGER
		);
		-- Real game: t_stats_players links via baseballPlayerLocalID (integer).
		-- baseballPlayerGUIDIfKnown is only an output alias in SMB3Explorer SQL,
		-- not a real column name.
		CREATE TABLE t_stats_players (
			statsPlayerID         INTEGER PRIMARY KEY NOT NULL,
			baseballPlayerLocalID INTEGER REFERENCES t_baseball_player_local_ids(localID),
			firstName             TEXT,
			lastName              TEXT,
			primaryPos            TEXT,
			secondaryPos          TEXT,
			pitcherRole           TEXT,
			age                   INTEGER,
			retirementSeason      INTEGER
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
			seasonID     INTEGER NOT NULL
		);
		CREATE TABLE t_career_season_stats (
			aggregatorID INTEGER NOT NULL REFERENCES t_stats(aggregatorID)
		);
		CREATE TABLE t_playoff_stats (
			aggregatorID INTEGER NOT NULL REFERENCES t_stats(aggregatorID),
			seasonID     INTEGER NOT NULL
		);
		-- Real game: t_season_schedule has no gameNumber or day — those are
		-- computed via RANK/ROW_NUMBER in queries.
		CREATE TABLE t_season_schedule (
			seasonID    INTEGER NOT NULL,
			homeTeamID  INTEGER NOT NULL,
			awayTeamID  INTEGER NOT NULL
		);
		-- Real game: t_game_results.ID is the integer PK; teams linked by local ID.
		CREATE TABLE t_game_results (
			ID                 INTEGER PRIMARY KEY NOT NULL,
			homeTeamLocalID    INTEGER NOT NULL,
			awayTeamLocalID    INTEGER NOT NULL,
			homeRunsScored     INTEGER,
			awayRunsScored     INTEGER,
			homePitcherLocalID INTEGER,
			awayPitcherLocalID INTEGER
		);
		-- Real game: t_season_games links seasons to game results via gameID.
		CREATE TABLE t_season_games (
			seasonID INTEGER NOT NULL,
			gameID   INTEGER NOT NULL REFERENCES t_game_results(ID)
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
		-- Real game: t_playoff_games links directly to t_game_results.
		CREATE TABLE t_playoff_games (
			playoffGUID  BLOB NOT NULL REFERENCES t_playoffs(GUID),
			seriesNumber INTEGER NOT NULL,
			gameID       INTEGER NOT NULL REFERENCES t_game_results(ID)
		);
		-- v_season_standings: teamGUID (blob) and seasonGUID (blob) per real schema.
		CREATE VIEW v_season_standings AS
		SELECT teamGUID, seasonGUID,
		       SUM(won) AS gamesWon, SUM(lost) AS gamesLost,
		       0.0 AS gamesBack, SUM(runsFor) AS runsFor, SUM(runsAgainst) AS runsAgainst
		FROM (
			SELECT htli.GUID AS teamGUID, ts.GUID AS seasonGUID,
			       CASE WHEN gr.homeRunsScored > gr.awayRunsScored THEN 1 ELSE 0 END AS won,
			       CASE WHEN gr.homeRunsScored < gr.awayRunsScored THEN 1 ELSE 0 END AS lost,
			       COALESCE(gr.homeRunsScored, 0) AS runsFor,
			       COALESCE(gr.awayRunsScored, 0) AS runsAgainst
			FROM t_season_games sg
			JOIN t_game_results gr ON gr.ID = sg.gameID
			JOIN t_team_local_ids htli ON htli.localID = gr.homeTeamLocalID
			JOIN t_seasons ts ON ts.id = sg.seasonID
			UNION ALL
			SELECT atli.GUID AS teamGUID, ts.GUID AS seasonGUID,
			       CASE WHEN gr.awayRunsScored > gr.homeRunsScored THEN 1 ELSE 0 END AS won,
			       CASE WHEN gr.awayRunsScored < gr.homeRunsScored THEN 1 ELSE 0 END AS lost,
			       COALESCE(gr.awayRunsScored, 0) AS runsFor,
			       COALESCE(gr.homeRunsScored, 0) AS runsAgainst
			FROM t_season_games sg
			JOIN t_game_results gr ON gr.ID = sg.gameID
			JOIN t_team_local_ids atli ON atli.localID = gr.awayTeamLocalID
			JOIN t_seasons ts ON ts.id = sg.seasonID
		) GROUP BY teamGUID, seasonGUID;
	`)
	return err
}

func seedSaveGameData(db *sql.DB) error {
	_, err := db.Exec(`
		INSERT INTO t_team_types (teamType, typeName) VALUES (1, 'franchise');

		INSERT INTO t_leagues (GUID, name, allowedTeamType)
		VALUES (X'EE000000000000000000000000000000', 'Test Franchise League', 1);

		INSERT INTO t_franchise (GUID, leagueGUID, playerTeamGUID)
		VALUES (X'FF000000000000000000000000000000', X'EE000000000000000000000000000000', X'01000000000000000000000000000000');

		-- t_seasons.GUID is a blob linked by t_playoffs.seasonGUID.
		-- Season 100 GUID matches the playoff entry below.
		INSERT INTO t_seasons (id, GUID, historicalLeagueGUID, elimination)
		VALUES (100, X'DD000000000000000000000000000000', X'EE000000000000000000000000000000', 0);
		INSERT INTO t_seasons (id, GUID, historicalLeagueGUID, elimination)
		VALUES (101, X'DE000000000000000000000000000000', X'EE000000000000000000000000000000', 0);

		INSERT INTO t_franchise_seasons (seasonID, franchiseGUID) VALUES (100, X'FF000000000000000000000000000000');
		INSERT INTO t_franchise_seasons (seasonID, franchiseGUID) VALUES (101, X'FF000000000000000000000000000000');

		-- Teams
		INSERT INTO t_teams (GUID, teamName) VALUES (X'01000000000000000000000000000000', 'Home Squad');
		INSERT INTO t_teams (GUID, teamName) VALUES (X'02000000000000000000000000000000', 'Away Crew');
		INSERT INTO t_team_local_ids (GUID) VALUES (X'01000000000000000000000000000000'); -- localID 1
		INSERT INTO t_team_local_ids (GUID) VALUES (X'02000000000000000000000000000000'); -- localID 2

		-- Conferences and divisions (GUID-based, not rowid-based)
		INSERT INTO t_conferences (GUID, name)
		VALUES (X'A1000000000000000000000000000000', 'East Conference');
		INSERT INTO t_divisions (GUID, name, conferenceGUID)
		VALUES (X'B1000000000000000000000000000000', 'East Division', X'A1000000000000000000000000000000');
		INSERT INTO t_division_teams (teamGUID, divisionGUID)
		VALUES (X'01000000000000000000000000000000', X'B1000000000000000000000000000000');
		INSERT INTO t_division_teams (teamGUID, divisionGUID)
		VALUES (X'02000000000000000000000000000000', X'B1000000000000000000000000000000');

		-- Player AA: outfielder (batter) — localID 1
		INSERT INTO t_baseball_players (GUID, power, contact, speed, fielding, arm, velocity, junk, accuracy, age)
		VALUES (X'AA000000000000000000000000000000', 80, 75, 60, 70, 65, 50, 50, 50, 27);
		INSERT INTO t_baseball_player_local_ids (GUID) VALUES (X'AA000000000000000000000000000000');
		-- v_baseball_player_info: simplified fixture view that provides player
		-- biographical data keyed by baseballPlayerGUID (same GUID as t_baseball_player_local_ids).
		-- In the real game this is a more complex view; for tests, names come from
		-- t_stats_players directly since the fixture populates them there.
		CREATE VIEW v_baseball_player_info AS
		SELECT
			bpli.GUID        AS baseballPlayerGUID,
			sp.firstName,
			sp.lastName,
			sp.primaryPos    AS primaryPosition,
			sp.pitcherRole,
			sp.statsPlayerID
		FROM t_baseball_player_local_ids bpli
		JOIN t_stats_players sp ON sp.baseballPlayerLocalID = bpli.localID;

		-- Player options (SMB4-only): throw hand, bat hand, chemistry
		INSERT INTO t_baseball_player_options (baseballPlayerLocalID, optionKey, optionValue) VALUES (1, 4, 1);   -- batter: throw R
		INSERT INTO t_baseball_player_options (baseballPlayerLocalID, optionKey, optionValue) VALUES (1, 5, 1);   -- batter: bat R
		INSERT INTO t_baseball_player_options (baseballPlayerLocalID, optionKey, optionValue) VALUES (1, 107, 0); -- batter: Competitive
		INSERT INTO t_baseball_player_options (baseballPlayerLocalID, optionKey, optionValue) VALUES (2, 4, 1);   -- pitcher: throw R
		INSERT INTO t_baseball_player_options (baseballPlayerLocalID, optionKey, optionValue) VALUES (2, 107, 1); -- pitcher: Spirited

		-- no trait rows → query subquery returns NULL → COALESCE gives '[]'
		INSERT INTO t_salary (baseballPlayerGUID, salary) VALUES (X'AA000000000000000000000000000000', 250);

		-- Player BB: pitcher — localID 2
		INSERT INTO t_baseball_players (GUID, power, contact, speed, fielding, arm, velocity, junk, accuracy, age)
		VALUES (X'BB000000000000000000000000000000', 40, 40, 40, 50, 55, 88, 78, 82, 30);
		INSERT INTO t_baseball_player_local_ids (GUID) VALUES (X'BB000000000000000000000000000000');
		INSERT INTO t_salary (baseballPlayerGUID, salary) VALUES (X'BB000000000000000000000000000000', 300);

		-- ── Season 100 stats ──────────────────────────────────────────────────

		-- Batter (AA) — regular season; baseballPlayerLocalID = 1
		INSERT INTO t_stats (aggregatorID, statsPlayerID, currentTeamLocalID) VALUES (1, 1, 1);
		INSERT INTO t_stats_players (statsPlayerID, baseballPlayerLocalID, firstName, lastName, primaryPos, age)
		VALUES (1, 1, 'Test', 'Batter', 'CF', 27);
		INSERT INTO t_stats_batting (aggregatorID, gamesPlayed, gamesBatting, atBats, runs, hits, doubles, triples, homeruns, rbi, baseOnBalls, strikeOuts)
		VALUES (1, 50, 50, 180, 30, 54, 10, 2, 12, 40, 20, 35);
		INSERT INTO t_season_stats (aggregatorID, seasonID) VALUES (1, 100);

		-- Pitcher (BB) — regular season; baseballPlayerLocalID = 2
		INSERT INTO t_stats (aggregatorID, statsPlayerID, currentTeamLocalID) VALUES (2, 2, 1);
		INSERT INTO t_stats_players (statsPlayerID, baseballPlayerLocalID, firstName, lastName, primaryPos, pitcherRole, age)
		VALUES (2, 2, 'Test', 'Pitcher', 'P', 'SP', 30);
		INSERT INTO t_stats_pitching (aggregatorID, wins, losses, games, gamesStarted, outsPitched, hits, earnedRuns, homeRuns, baseOnBalls, strikeOuts, battersFaced, totalPitches)
		VALUES (2, 12, 8, 25, 25, 540, 140, 55, 15, 40, 180, 740, 3200);
		INSERT INTO t_season_stats (aggregatorID, seasonID) VALUES (2, 100);

		-- Batter (AA) — playoff career stats aggregator
		INSERT INTO t_stats (aggregatorID, statsPlayerID, currentTeamLocalID) VALUES (3, 3, 1);
		INSERT INTO t_stats_players (statsPlayerID, baseballPlayerLocalID, firstName, lastName, primaryPos, age)
		VALUES (3, 1, 'Test', 'Batter', 'CF', 27);
		INSERT INTO t_stats_batting (aggregatorID, gamesPlayed, gamesBatting, atBats, hits, homeruns, rbi)
		VALUES (3, 5, 5, 18, 6, 2, 5);

		-- Pitcher (BB) — playoff career stats aggregator
		INSERT INTO t_stats (aggregatorID, statsPlayerID, currentTeamLocalID) VALUES (4, 4, 1);
		INSERT INTO t_stats_players (statsPlayerID, baseballPlayerLocalID, firstName, lastName, primaryPos, pitcherRole, age)
		VALUES (4, 2, 'Test', 'Pitcher', 'P', 'SP', 30);
		INSERT INTO t_stats_pitching (aggregatorID, wins, losses, games, gamesStarted, outsPitched, hits, earnedRuns, strikeOuts)
		VALUES (4, 2, 0, 2, 2, 54, 10, 3, 18);

		-- Playoff stats for season 100 (aggregators 3 and 4)
		INSERT INTO t_playoff_stats (aggregatorID, seasonID) VALUES (3, 100);
		INSERT INTO t_playoff_stats (aggregatorID, seasonID) VALUES (4, 100);

		-- Career stats (aggregators 1 and 2)
		INSERT INTO t_career_season_stats (aggregatorID) VALUES (1);
		INSERT INTO t_career_season_stats (aggregatorID) VALUES (2);

		-- ── Season 100 regular season games ───────────────────────────────────
		-- t_season_schedule: just homeTeamID + awayTeamID (no gameNumber/day)
		INSERT INTO t_season_schedule (seasonID, homeTeamID, awayTeamID) VALUES (100, 1, 2);
		INSERT INTO t_season_schedule (seasonID, homeTeamID, awayTeamID) VALUES (100, 2, 1);

		-- t_game_results.ID is the integer PK
		INSERT INTO t_game_results (ID, homeTeamLocalID, awayTeamLocalID, homeRunsScored, awayRunsScored, homePitcherLocalID, awayPitcherLocalID)
		VALUES (1, 1, 2, 5, 3, 2, 1);
		INSERT INTO t_game_results (ID, homeTeamLocalID, awayTeamLocalID, homeRunsScored, awayRunsScored, homePitcherLocalID, awayPitcherLocalID)
		VALUES (2, 2, 1, 1, 4, 1, 2);
		INSERT INTO t_season_games (seasonID, gameID) VALUES (100, 1);
		INSERT INTO t_season_games (seasonID, gameID) VALUES (100, 2);

		-- ── Season 100 playoffs ────────────────────────────────────────────────
		-- t_playoffs.seasonGUID links to t_seasons.GUID for season 100
		INSERT INTO t_playoffs (GUID, seasonGUID)
		VALUES (X'CC000000000000000000000000000000', X'DD000000000000000000000000000000');
		INSERT INTO t_playoff_series (playoffGUID, seriesNumber, team1GUID, team2GUID, team1Standing, team2Standing)
		VALUES (X'CC000000000000000000000000000000', 1, X'01000000000000000000000000000000', X'02000000000000000000000000000000', 1, 2);

		-- Playoff game result (ID=3) linked via t_playoff_games, not t_season_games
		INSERT INTO t_game_results (ID, homeTeamLocalID, awayTeamLocalID, homeRunsScored, awayRunsScored, homePitcherLocalID, awayPitcherLocalID)
		VALUES (3, 1, 2, 3, 1, 2, 1);
		INSERT INTO t_playoff_games (playoffGUID, seriesNumber, gameID)
		VALUES (X'CC000000000000000000000000000000', 1, 3);

		-- ── Season 101 stats (same players, second season) ────────────────────

		INSERT INTO t_stats (aggregatorID, statsPlayerID, currentTeamLocalID) VALUES (5, 5, 1);
		INSERT INTO t_stats_players (statsPlayerID, baseballPlayerLocalID, firstName, lastName, primaryPos, age)
		VALUES (5, 1, 'Test', 'Batter', 'CF', 28);
		INSERT INTO t_stats_batting (aggregatorID, gamesPlayed, gamesBatting, atBats, runs, hits, homeruns, rbi)
		VALUES (5, 52, 52, 190, 35, 60, 15, 48);
		INSERT INTO t_season_stats (aggregatorID, seasonID) VALUES (5, 101);

		INSERT INTO t_stats (aggregatorID, statsPlayerID, currentTeamLocalID) VALUES (6, 6, 1);
		INSERT INTO t_stats_players (statsPlayerID, baseballPlayerLocalID, firstName, lastName, primaryPos, pitcherRole, age)
		VALUES (6, 2, 'Test', 'Pitcher', 'P', 'SP', 31);
		INSERT INTO t_stats_pitching (aggregatorID, wins, losses, games, gamesStarted, outsPitched, hits, earnedRuns, strikeOuts)
		VALUES (6, 14, 7, 25, 25, 570, 130, 50, 195);
		INSERT INTO t_season_stats (aggregatorID, seasonID) VALUES (6, 101);

		INSERT INTO t_season_schedule (seasonID, homeTeamID, awayTeamID) VALUES (101, 1, 2);
		INSERT INTO t_game_results (ID, homeTeamLocalID, awayTeamLocalID, homeRunsScored, awayRunsScored)
		VALUES (4, 1, 2, 6, 2);
		INSERT INTO t_season_games (seasonID, gameID) VALUES (101, 4);
	`)
	return err
}

// MidSeasonStats are overrides for creating a save game DB that simulates
// a mid-season snapshot (partial stats) for player AA in season 100.
type MidSeasonStats struct {
	Hits   int
	AtBats int
	HomeRuns int
}

// NewTestSaveGameDB_MidSeason creates a save game DB identical to
// NewTestSaveGameDB except that player AA's batting stats for season 100
// are replaced with the provided partial-season values.
// Use this to test idempotent re-imports (mid-season then end-of-season).
func NewTestSaveGameDB_MidSeason(t *testing.T, stats MidSeasonStats) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("testutil.NewTestSaveGameDB_MidSeason: open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := createSaveGameSchema(db); err != nil {
		t.Fatalf("testutil.NewTestSaveGameDB_MidSeason: schema: %v", err)
	}
	if err := seedSaveGameData(db); err != nil {
		t.Fatalf("testutil.NewTestSaveGameDB_MidSeason: seed: %v", err)
	}
	// Override batting stats for player AA, season 100 (aggregatorID=1)
	_, err = db.Exec(`
		UPDATE t_stats_batting
		SET hits = ?, atBats = ?, homeruns = ?
		WHERE aggregatorID = 1
	`, stats.Hits, stats.AtBats, stats.HomeRuns)
	if err != nil {
		t.Fatalf("testutil.NewTestSaveGameDB_MidSeason: overriding stats: %v", err)
	}
	return db
}
