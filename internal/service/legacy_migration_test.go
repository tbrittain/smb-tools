package service_test

import (
	"context"
	"database/sql"
	"testing"

	"smb-tools/internal/service"
	"smb-tools/internal/testutil"
)

const (
	legacyFranchiseA     = 1
	legacyFranchiseB     = 2
	testLeagueGUIDLegacy = "legacy-test-0000-0000-000000000001"
)

func newLegacyMigrationSvc() *service.LegacyMigrationService {
	return service.NewLegacyMigrationService()
}

func assertRowCount(t *testing.T, db *sql.DB, table string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM "+table,
	).Scan(&got); err != nil {
		t.Fatalf("counting %s: %v", table, err)
	}
	if got != want {
		t.Errorf("%s row count = %d, want %d", table, got, want)
	}
}

// TestMigrateLegacy_Full migrates franchise A (2 seasons, 3 players, all stat types,
// traits, pitches, awards, schedules, playoffs) and verifies all record counts.
func TestMigrateLegacy_Full(t *testing.T) {
	ctx := context.Background()
	legacyDB := testutil.NewLegacyCompanionDB(t)
	companionDB := testutil.NewTestDB(t)

	result, err := newLegacyMigrationSvc().Migrate(ctx, legacyDB, legacyFranchiseA, companionDB, testLeagueGUIDLegacy)
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	if result.SeasonsMigrated != 2 {
		t.Errorf("SeasonsMigrated = %d, want 2", result.SeasonsMigrated)
	}
	if result.TeamsMigrated != 2 {
		t.Errorf("TeamsMigrated = %d, want 2", result.TeamsMigrated)
	}
	if result.PlayersMigrated != 3 {
		t.Errorf("PlayersMigrated = %d, want 3", result.PlayersMigrated)
	}
	if result.AwardsMigrated != 2 {
		t.Errorf("AwardsMigrated = %d, want 2", result.AwardsMigrated)
	}

	assertRowCount(t, companionDB, "seasons", 2)
	assertRowCount(t, companionDB, "teams", 2)
	assertRowCount(t, companionDB, "team_season_history", 4) // 2 teams × 2 seasons
	assertRowCount(t, companionDB, "players", 3)
	assertRowCount(t, companionDB, "player_seasons", 6) // 3 players × 2 seasons
	assertRowCount(t, companionDB, "player_season_game_stats", 6)
	// 2 user awards (Alex MVP S10, Alex Greatest Slugger S11) + 2 championship awards
	// (Sam S11 + Riley S11 League Champion — Alpha Squad S11 is the ChampionshipWinners entry;
	// no runner-up because the fixture has no played playoff games for season 11).
	assertRowCount(t, companionDB, "player_season_awards", 4)
	// Season 10: 2 regular + Season 11: 1 regular = 3
	assertRowCount(t, companionDB, "team_season_schedules", 3)
	// Season 10: 1 playoff game
	assertRowCount(t, companionDB, "team_playoff_schedules", 1)
	// Team 1 has 2 GUIDs → 1 alt GUID row
	assertRowCount(t, companionDB, "team_alt_guids", 1)
	// Player 1 has 2 GUIDs → 1 alt GUID row
	assertRowCount(t, companionDB, "player_alt_guids", 1)
	// player_season_teams: Alex S10=1, Sam S10=1, Riley S10=1 (FA slot skipped; prior team kept),
	// Alex S11=2 (traded), Sam S11=1, Riley S11=1 → 7 rows total
	assertRowCount(t, companionDB, "player_season_teams", 7)

	// Verify HoF flag on Alex Power
	var isHoF int
	if err := companionDB.QueryRowContext(ctx,
		`SELECT is_hall_of_famer FROM players WHERE first_name = 'Alex' AND last_name = 'Power'`,
	).Scan(&isHoF); err != nil {
		t.Fatalf("querying Alex HoF: %v", err)
	}
	if isHoF != 1 {
		t.Errorf("Alex Power is_hall_of_famer = %d, want 1", isHoF)
	}
}

// TestMigrateLegacy_Minimal migrates franchise B (1 season, 2 players, regular season only).
func TestMigrateLegacy_Minimal(t *testing.T) {
	legacyDB := testutil.NewLegacyCompanionDB(t)
	companionDB := testutil.NewTestDB(t)

	result, err := newLegacyMigrationSvc().Migrate(context.Background(),
		legacyDB, legacyFranchiseB, companionDB, "legacy-test-0000-0000-000000000002")
	if err != nil {
		t.Fatalf("Migrate franchise B: %v", err)
	}

	if result.SeasonsMigrated != 1 {
		t.Errorf("SeasonsMigrated = %d, want 1", result.SeasonsMigrated)
	}
	if result.PlayersMigrated != 2 {
		t.Errorf("PlayersMigrated = %d, want 2", result.PlayersMigrated)
	}
	assertRowCount(t, companionDB, "seasons", 1)
	assertRowCount(t, companionDB, "players", 2)
	assertRowCount(t, companionDB, "player_seasons", 2)
	assertRowCount(t, companionDB, "team_playoff_schedules", 0)
}

// TestMigrateLegacy_IPConversion verifies the InningsPitched→outs_pitched transform.
func TestMigrateLegacy_IPConversion(t *testing.T) {
	legacyDB := testutil.NewLegacyCompanionDB(t)
	companionDB := testutil.NewTestDB(t)

	if _, err := newLegacyMigrationSvc().Migrate(context.Background(),
		legacyDB, legacyFranchiseA, companionDB, testLeagueGUIDLegacy); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	cases := []struct {
		firstName  string
		seasonNum  int
		wantOuts   int
		desc       string
	}{
		{"Sam", 1, 540, "180.0 IP → 540 outs"},
		{"Riley", 1, 137, "45.2 IP → 137 outs"},
		{"Sam", 2, 556, "185.1 IP → 556 outs"},
	}
	for _, tc := range cases {
		var outs int
		err := companionDB.QueryRowContext(context.Background(), `
			SELECT pp.outs_pitched
			FROM player_season_pitching_stats pp
			JOIN player_seasons ps ON ps.id = pp.player_season_id
			JOIN players p ON p.id = ps.player_id
			JOIN seasons s ON s.id = ps.season_id
			WHERE p.first_name = ? AND s.season_num = ? AND pp.is_regular_season = 1
		`, tc.firstName, tc.seasonNum).Scan(&outs)
		if err != nil {
			t.Errorf("querying %s S%d: %v", tc.firstName, tc.seasonNum, err)
			continue
		}
		if outs != tc.wantOuts {
			t.Errorf("%s S%d outs_pitched = %d, want %d (%s)", tc.firstName, tc.seasonNum, outs, tc.wantOuts, tc.desc)
		}
	}
}

// TestMigrateLegacy_TraitsAndPitches verifies JSON arrays in player_seasons.
func TestMigrateLegacy_TraitsAndPitches(t *testing.T) {
	legacyDB := testutil.NewLegacyCompanionDB(t)
	companionDB := testutil.NewTestDB(t)

	if _, err := newLegacyMigrationSvc().Migrate(context.Background(),
		legacyDB, legacyFranchiseA, companionDB, testLeagueGUIDLegacy); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	var traits string
	if err := companionDB.QueryRowContext(context.Background(), `
		SELECT ps.traits_json
		FROM player_seasons ps
		JOIN players p ON p.id = ps.player_id
		JOIN seasons s ON s.id = ps.season_id
		WHERE p.first_name = 'Alex' AND s.season_num = 1
	`).Scan(&traits); err != nil {
		t.Fatalf("querying Alex S1 traits: %v", err)
	}
	if traits != `["Clutch","Tough Out"]` {
		t.Errorf("Alex S1 traits_json = %q, want [\"Clutch\",\"Tough Out\"]", traits)
	}

	var pitches string
	if err := companionDB.QueryRowContext(context.Background(), `
		SELECT ps.pitches_json
		FROM player_seasons ps
		JOIN players p ON p.id = ps.player_id
		JOIN seasons s ON s.id = ps.season_id
		WHERE p.first_name = 'Sam' AND s.season_num = 1
	`).Scan(&pitches); err != nil {
		t.Fatalf("querying Sam S1 pitches: %v", err)
	}
	if pitches != `["4F","SL"]` {
		t.Errorf("Sam S1 pitches_json = %q, want [\"4F\",\"SL\"]", pitches)
	}
}

// TestMigrateLegacy_UserDefinedAward verifies a custom award (IsBuiltIn=0)
// is inserted into awards with is_built_in=0 and assigned correctly.
func TestMigrateLegacy_UserDefinedAward(t *testing.T) {
	legacyDB := testutil.NewLegacyCompanionDB(t)
	companionDB := testutil.NewTestDB(t)

	if _, err := newLegacyMigrationSvc().Migrate(context.Background(),
		legacyDB, legacyFranchiseA, companionDB, testLeagueGUIDLegacy); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	var isBuiltIn int
	if err := companionDB.QueryRowContext(context.Background(),
		`SELECT is_built_in FROM awards WHERE original_name = 'Greatest Slugger'`,
	).Scan(&isBuiltIn); err != nil {
		t.Fatalf("querying custom award: %v", err)
	}
	if isBuiltIn != 0 {
		t.Errorf("Greatest Slugger is_built_in = %d, want 0", isBuiltIn)
	}

	var count int
	if err := companionDB.QueryRowContext(context.Background(), `
		SELECT COUNT(*) FROM player_season_awards psa
		JOIN awards a ON a.id = psa.award_id
		WHERE a.original_name = 'Greatest Slugger'
	`).Scan(&count); err != nil {
		t.Fatalf("counting custom award assignments: %v", err)
	}
	if count != 1 {
		t.Errorf("Greatest Slugger assignment count = %d, want 1", count)
	}
}

// TestMigrateLegacy_Idempotent verifies running Migrate twice produces no duplicates.
func TestMigrateLegacy_Idempotent(t *testing.T) {
	legacyDB := testutil.NewLegacyCompanionDB(t)
	companionDB := testutil.NewTestDB(t)
	svc := newLegacyMigrationSvc()

	for i := range 2 {
		if _, err := svc.Migrate(context.Background(),
			legacyDB, legacyFranchiseA, companionDB, testLeagueGUIDLegacy); err != nil {
			t.Fatalf("Migrate run %d: %v", i+1, err)
		}
	}

	assertRowCount(t, companionDB, "seasons", 2)
	assertRowCount(t, companionDB, "players", 3)
	assertRowCount(t, companionDB, "player_seasons", 6)
	assertRowCount(t, companionDB, "teams", 2)
	assertRowCount(t, companionDB, "team_season_history", 4)
}

// TestMigrateLegacy_TwoFranchises migrates both franchises from a single legacy DB
// into separate companion DBs and asserts no cross-contamination between them.
func TestMigrateLegacy_TwoFranchises(t *testing.T) {
	legacyDB := testutil.NewLegacyCompanionDB(t)
	companionDBA := testutil.NewTestDB(t)
	companionDBB := testutil.NewTestDB(t)
	svc := newLegacyMigrationSvc()

	resultA, err := svc.Migrate(context.Background(), legacyDB, legacyFranchiseA,
		companionDBA, "legacy-test-0000-0000-aaaaaaaaaaaa")
	if err != nil {
		t.Fatalf("Migrate franchise A: %v", err)
	}
	resultB, err := svc.Migrate(context.Background(), legacyDB, legacyFranchiseB,
		companionDBB, "legacy-test-0000-0000-bbbbbbbbbbbb")
	if err != nil {
		t.Fatalf("Migrate franchise B: %v", err)
	}

	// Franchise A: 3 players, 2 seasons
	if resultA.PlayersMigrated != 3 {
		t.Errorf("A: PlayersMigrated = %d, want 3", resultA.PlayersMigrated)
	}
	assertRowCount(t, companionDBA, "players", 3)
	assertRowCount(t, companionDBA, "seasons", 2)

	// Franchise B: 2 players, 1 season
	if resultB.PlayersMigrated != 2 {
		t.Errorf("B: PlayersMigrated = %d, want 2", resultB.PlayersMigrated)
	}
	assertRowCount(t, companionDBB, "players", 2)
	assertRowCount(t, companionDBB, "seasons", 1)

	// No franchise A player in franchise B's DB
	var dummy string
	if err := companionDBB.QueryRowContext(context.Background(),
		`SELECT first_name FROM players WHERE first_name = 'Alex'`,
	).Scan(&dummy); err == nil {
		t.Errorf("franchise B DB contains franchise A player 'Alex' — cross-contamination")
	}

	// No franchise B player in franchise A's DB
	if err := companionDBA.QueryRowContext(context.Background(),
		`SELECT first_name FROM players WHERE first_name = 'Chris'`,
	).Scan(&dummy); err == nil {
		t.Errorf("franchise A DB contains franchise B player 'Chris' — cross-contamination")
	}
}

// TestMigrateLegacy_TeamAssociations verifies the three team-history scenarios that
// are present in the legacy seed data end-to-end through the full migration pipeline:
//
//   - Single-team player (Sam S10): one player_season_teams row at sort_order=0
//   - FA player (Riley S10): FA slot (NULL SeasonTeamHistoryId) is skipped; prior team
//     is stored at sort_order=1; no sort_order=0 row (correct — absence means FA)
//   - Traded player (Alex S11): two rows — sort_order=0 (Beta Ballers, current) and
//     sort_order=1 (Alpha Squad, prior)
func TestMigrateLegacy_TeamAssociations(t *testing.T) {
	ctx := context.Background()
	legacyDB := testutil.NewLegacyCompanionDB(t)
	companionDB := testutil.NewTestDB(t)

	if _, err := newLegacyMigrationSvc().Migrate(ctx, legacyDB, legacyFranchiseA, companionDB, testLeagueGUIDLegacy); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	// Helper: return (sort_order, team_name) pairs for a given player+season.
	type teamRow struct {
		sortOrder int
		teamName  string
	}
	queryTeams := func(firstName string, seasonNum int) []teamRow {
		t.Helper()
		rows, err := companionDB.QueryContext(ctx, `
			SELECT pst.sort_order, tsh.team_name
			FROM player_season_teams pst
			JOIN team_season_history tsh ON tsh.id = pst.team_history_id
			JOIN player_seasons ps ON ps.id = pst.player_season_id
			JOIN players p ON p.id = ps.player_id
			JOIN seasons s ON s.id = ps.season_id
			WHERE p.first_name = ? AND s.season_num = ?
			ORDER BY pst.sort_order ASC
		`, firstName, seasonNum)
		if err != nil {
			t.Fatalf("querying teams for %s S%d: %v", firstName, seasonNum, err)
		}
		defer func() { _ = rows.Close() }()
		var out []teamRow
		for rows.Next() {
			var r teamRow
			if err := rows.Scan(&r.sortOrder, &r.teamName); err != nil {
				t.Fatalf("scanning team row: %v", err)
			}
			out = append(out, r)
		}
		return out
	}

	// Sam S10: single team
	samS10 := queryTeams("Sam", 1)
	if len(samS10) != 1 {
		t.Fatalf("Sam S1: expected 1 team, got %d", len(samS10))
	}
	if samS10[0].sortOrder != 0 || samS10[0].teamName != "Alpha Squad" {
		t.Errorf("Sam S1 team: got {sort_order:%d name:%q}, want {0 Alpha Squad}", samS10[0].sortOrder, samS10[0].teamName)
	}

	// Riley S10: FA — no sort_order=0 row; prior team at sort_order=1
	rileyS10 := queryTeams("Riley", 1)
	if len(rileyS10) != 1 {
		t.Fatalf("Riley S1: expected 1 team (FA slot skipped), got %d", len(rileyS10))
	}
	if rileyS10[0].sortOrder != 1 || rileyS10[0].teamName != "Alpha Squad" {
		t.Errorf("Riley S1 prior team: got {sort_order:%d name:%q}, want {1 Alpha Squad}", rileyS10[0].sortOrder, rileyS10[0].teamName)
	}

	// Alex S11: traded — sort_order=0 Beta Ballers (landed), sort_order=1 Alpha Squad (came from)
	alexS11 := queryTeams("Alex", 2)
	if len(alexS11) != 2 {
		t.Fatalf("Alex S2: expected 2 teams (traded), got %d", len(alexS11))
	}
	if alexS11[0].sortOrder != 0 || alexS11[0].teamName != "Beta Ballers" {
		t.Errorf("Alex S2 current team: got {sort_order:%d name:%q}, want {0 Beta Ballers}", alexS11[0].sortOrder, alexS11[0].teamName)
	}
	if alexS11[1].sortOrder != 1 || alexS11[1].teamName != "Alpha Squad" {
		t.Errorf("Alex S2 prior team: got {sort_order:%d name:%q}, want {1 Alpha Squad}", alexS11[1].sortOrder, alexS11[1].teamName)
	}
}

// TestMigrateLegacy_OrphanedPlayerNoGUID verifies that a player with no GameGUIDs is
// silently skipped: their player season and batting stats must NOT be migrated, and
// the migration must still succeed without error.
//
// This exercises the phase-4 guard (player skipped) → phase-5 guard (player season
// skipped) → phase-7 guard (batting stat skipped) chain in migrateInTx.
func TestMigrateLegacy_OrphanedPlayerNoGUID(t *testing.T) {
	ctx := context.Background()
	legacyDB := testutil.NewLegacyCompanionDB(t)

	// Insert a player with no PlayerGameIdHistory entry for franchise A.
	// BatHandednessId=1 (R), ThrowHandednessId=1 (R), PrimaryPositionId=8 (CF), ChemistryId=1.
	if _, err := legacyDB.ExecContext(ctx, `
		INSERT INTO Players (Id, FirstName, LastName, IsHallOfFamer,
		                     BatHandednessId, ThrowHandednessId, PrimaryPositionId,
		                     PitcherRoleId, ChemistryId, FranchiseId)
		VALUES (99, 'Ghost', 'Player', 0, 1, 1, 8, NULL, 1, 1)
	`); err != nil {
		t.Fatalf("inserting ghost player: %v", err)
	}
	// Player season for ghost player in season 10.
	if _, err := legacyDB.ExecContext(ctx, `
		INSERT INTO PlayerSeasons (Id, PlayerId, SeasonId, Age, Salary)
		VALUES (99, 99, 10, 25, 100)
	`); err != nil {
		t.Fatalf("inserting ghost player season: %v", err)
	}
	// Batting stats for the ghost player season.
	if _, err := legacyDB.ExecContext(ctx, `
		INSERT INTO PlayerSeasonBattingStats
		    (PlayerSeasonId, GamesPlayed, GamesBatting, AtBats, PlateAppearances, IsRegularSeason)
		VALUES (99, 30, 30, 90, 100, 1)
	`); err != nil {
		t.Fatalf("inserting ghost batting stats: %v", err)
	}

	companionDB := testutil.NewTestDB(t)
	result, err := newLegacyMigrationSvc().Migrate(ctx, legacyDB, legacyFranchiseA, companionDB, testLeagueGUIDLegacy)
	if err != nil {
		t.Fatalf("Migrate with orphaned player: %v", err)
	}

	// The ghost player must not be counted or migrated.
	if result.PlayersMigrated != 3 {
		t.Errorf("PlayersMigrated = %d, want 3 (ghost player must be skipped)", result.PlayersMigrated)
	}
	assertRowCount(t, companionDB, "players", 3)
	assertRowCount(t, companionDB, "player_seasons", 6)
}

// TestMigrateLegacy_OrphanedSeasonTeamHistory verifies that a season team history entry
// for a team with no GameGUIDs is silently skipped.
//
// This exercises the phase-2 guard (team skipped because no GUIDs) → phase-3 guard
// (season team history for that team is skipped) chain in migrateInTx.
func TestMigrateLegacy_OrphanedSeasonTeamHistory(t *testing.T) {
	ctx := context.Background()
	legacyDB := testutil.NewLegacyCompanionDB(t)

	// Insert a team for franchise A with no TeamGameIdHistory entries — it will be
	// skipped in phase 2 (guard: `if len(t.GameGUIDs) == 0 { continue }`).
	if _, err := legacyDB.ExecContext(ctx, `
		INSERT INTO Teams (Id, FranchiseId) VALUES (99, 1)
	`); err != nil {
		t.Fatalf("inserting team without GUIDs: %v", err)
	}
	// Season team history for the no-GUID team in season 10.
	// Phase-3 guard fires: team 99 is absent from legacyTeamIDToNew.
	if _, err := legacyDB.ExecContext(ctx, `
		INSERT INTO SeasonTeamHistory
		    (Id, SeasonId, TeamId, DivisionId, TeamNameHistoryId,
		     Budget, Payroll, Surplus, SurplusPerGame, Wins, Losses,
		     GamesBehind, WinPercentage, PythagoreanWinPercentage,
		     ExpectedWins, ExpectedLosses, RunsScored, RunsAllowed)
		VALUES (99, 10, 99, 1, 1, 0,0,0,0, 0,0, 0,0,0, 0,0,0,0)
	`); err != nil {
		t.Fatalf("inserting orphaned season team history: %v", err)
	}

	companionDB := testutil.NewTestDB(t)
	_, err := newLegacyMigrationSvc().Migrate(ctx, legacyDB, legacyFranchiseA, companionDB, testLeagueGUIDLegacy)
	if err != nil {
		t.Fatalf("Migrate with orphaned STH: %v", err)
	}
	// The orphaned team and its STH must not appear in the companion DB.
	assertRowCount(t, companionDB, "teams", 2)
	assertRowCount(t, companionDB, "team_season_history", 4) // 2 real teams × 2 seasons
}

// TestMigrateLegacy_GameStatsNullCoalesce verifies NULL Arm/Velocity/Junk/Accuracy
// in legacy are stored as 0 in the new schema.
func TestMigrateLegacy_GameStatsNullCoalesce(t *testing.T) {
	legacyDB := testutil.NewLegacyCompanionDB(t)
	companionDB := testutil.NewTestDB(t)

	if _, err := newLegacyMigrationSvc().Migrate(context.Background(),
		legacyDB, legacyFranchiseA, companionDB, testLeagueGUIDLegacy); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	// Alex Power (batter) has NULL Velocity/Junk/Accuracy in legacy
	var velocity, junk, accuracy int
	if err := companionDB.QueryRowContext(context.Background(), `
		SELECT gs.velocity, gs.junk, gs.accuracy
		FROM player_season_game_stats gs
		JOIN player_seasons ps ON ps.id = gs.player_season_id
		JOIN players p ON p.id = ps.player_id
		JOIN seasons s ON s.id = ps.season_id
		WHERE p.first_name = 'Alex' AND s.season_num = 1
	`).Scan(&velocity, &junk, &accuracy); err != nil {
		t.Fatalf("querying Alex game stats: %v", err)
	}
	if velocity != 0 || junk != 0 || accuracy != 0 {
		t.Errorf("Alex velocity=%d junk=%d accuracy=%d, want all 0", velocity, junk, accuracy)
	}
}
