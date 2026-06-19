package service_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"smb-tools/internal/models"
	"smb-tools/internal/service"
	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

const testLeagueGUID = "TESTLEAGUEGUID00000000000000000000"

// newTestImportService builds an ImportService backed by a fresh in-memory companion DB.
func newTestImportService(t *testing.T) (*service.ImportService, *sql.DB, store.SaveGameReader) {
	t.Helper()
	companionDB := testutil.NewTestDB(t)
	saveGameDB := testutil.NewTestSaveGameDB(t)
	reader := store.NewSqliteSaveGameReader(saveGameDB, "")
	svc := service.NewImportService()
	return svc, companionDB, reader
}

// importSeason1 is a helper that imports save game season 100 as display season 1.
func importSeason1(t *testing.T, svc *service.ImportService, companionDB *sql.DB, reader store.SaveGameReader) service.ImportResult {
	t.Helper()
	result, err := svc.ImportSeason(context.Background(), companionDB, reader, 100, 1, testLeagueGUID, 0)
	if err != nil {
		t.Fatalf("ImportSeason (season 1): %v", err)
	}
	return result
}

// ── Basic flow ───────────────────────────────────────────────────────────────

func TestImportSeason_BasicFlow(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)

	result := importSeason1(t, svc, companionDB, reader)

	if result.SeasonID == 0 {
		t.Error("expected non-zero companion DB season ID")
	}
	if result.SeasonNum != 1 {
		t.Errorf("SeasonNum: got %d, want 1", result.SeasonNum)
	}
	if result.Players < 1 {
		t.Errorf("expected at least 1 player, got %d", result.Players)
	}
	if result.Teams < 1 {
		t.Errorf("expected at least 1 team, got %d", result.Teams)
	}
}

func TestImportSeason_SeasonRecordCreated(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	result := importSeason1(t, svc, companionDB, reader)

	ss := store.NewSeasonStore(companionDB)
	season, err := ss.GetByID(ctx, result.SeasonID)
	if err != nil {
		t.Fatalf("GetByID after import: %v", err)
	}
	if season.SeasonNum != 1 {
		t.Errorf("season_num: got %d, want 1", season.SeasonNum)
	}
	if season.SaveGameSeasonID != 100 {
		t.Errorf("save_game_season_id: got %d, want 100", season.SaveGameSeasonID)
	}
	if season.LeagueGUID != testLeagueGUID {
		t.Errorf("league_guid: got %q, want %q", season.LeagueGUID, testLeagueGUID)
	}
	if season.NumGames != result.Games {
		t.Errorf("num_games: got %d, want %d (result.Games) — qualified-player thresholds break when this is 0", season.NumGames, result.Games)
	}
	if season.NumGames == 0 {
		t.Error("num_games: got 0, want > 0")
	}
}

func TestImportSeason_LeagueAvgAttributesPopulated(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	result := importSeason1(t, svc, companionDB, reader)

	var avgPower float64
	err := companionDB.QueryRowContext(ctx,
		`SELECT avg_power FROM season_attribute_averages WHERE season_id = ?`, result.SeasonID,
	).Scan(&avgPower)
	if err != nil {
		t.Fatalf("season_attribute_averages row missing after import: %v", err)
	}
	if avgPower <= 0 {
		t.Errorf("avg_power = %.4f, want > 0 (fixture has players with non-zero attributes)", avgPower)
	}

	var pctRows int
	if err := companionDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM player_season_attribute_percentiles
		 WHERE player_season_id IN (SELECT id FROM player_seasons WHERE season_id = ?)`,
		result.SeasonID,
	).Scan(&pctRows); err != nil {
		t.Fatalf("querying player_season_attribute_percentiles: %v", err)
	}
	if pctRows == 0 {
		t.Error("player_season_attribute_percentiles has 0 rows after import, want > 0")
	}
}

func TestImportSeason_SeasonOffsetApplied(t *testing.T) {
	svc, companionDB, _ := newTestImportService(t)
	ctx := context.Background()

	// Simulate a fork: save game season 1 with offset 15 → display season 16
	saveGameDB := testutil.NewTestSaveGameDB(t)
	reader := store.NewSqliteSaveGameReader(saveGameDB, "")
	result, err := svc.ImportSeason(ctx, companionDB, reader, 100, 1, "FORKED-LEAGUE-GUID0000000000000000", 15)
	if err != nil {
		t.Fatalf("ImportSeason with offset: %v", err)
	}
	if result.SeasonNum != 16 {
		t.Errorf("season_num with offset: got %d, want 16 (1 + 15)", result.SeasonNum)
	}
}

func TestImportSeason_PlayersCreated(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	importSeason1(t, svc, companionDB, reader)

	var count int
	if err := companionDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM players`).Scan(&count); err != nil {
		t.Fatalf("counting players: %v", err)
	}
	if count < 1 {
		t.Errorf("expected players in DB after import, got %d", count)
	}
}

func TestImportSeason_BattingStatsImported(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	importSeason1(t, svc, companionDB, reader)

	var count int
	if err := companionDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM player_season_batting_stats WHERE is_regular_season = 1`,
	).Scan(&count); err != nil {
		t.Fatalf("counting batting stats: %v", err)
	}
	if count < 1 {
		t.Errorf("expected regular season batting stats, got %d rows", count)
	}
}

func TestImportSeason_PitchingStatsImported(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	importSeason1(t, svc, companionDB, reader)

	var count int
	if err := companionDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM player_season_pitching_stats WHERE is_regular_season = 1`,
	).Scan(&count); err != nil {
		t.Fatalf("counting pitching stats: %v", err)
	}
	if count < 1 {
		t.Errorf("expected regular season pitching stats, got %d rows", count)
	}
}

func TestImportSeason_BattingStatValues(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	importSeason1(t, svc, companionDB, reader)

	// The fixture seeds player AA with 180 AB, 54 H, 12 HR, 40 RBI
	var atBats, hits, homeRuns, rbi int
	err := companionDB.QueryRowContext(ctx, `
		SELECT bs.at_bats, bs.hits, bs.home_runs, bs.rbi
		FROM player_season_batting_stats bs
		JOIN player_seasons ps ON ps.id = bs.player_season_id
		JOIN players p ON p.id = ps.player_id
		WHERE p.game_guid = 'AA000000000000000000000000000000'
		  AND bs.is_regular_season = 1
	`).Scan(&atBats, &hits, &homeRuns, &rbi)
	if err != nil {
		t.Fatalf("querying batter stats: %v", err)
	}
	if atBats != 180 {
		t.Errorf("at_bats: got %d, want 180", atBats)
	}
	if hits != 54 {
		t.Errorf("hits: got %d, want 54", hits)
	}
	if homeRuns != 12 {
		t.Errorf("home_runs: got %d, want 12", homeRuns)
	}
	if rbi != 40 {
		t.Errorf("rbi: got %d, want 40", rbi)
	}
}

func TestImportSeason_PitchingStatValues(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	importSeason1(t, svc, companionDB, reader)

	// Fixture: pitcher BB with W=12, L=8, outsPitched=540 (= 180 IP), K=180
	var wins, losses, outsPitched, strikeouts int
	err := companionDB.QueryRowContext(ctx, `
		SELECT ps2.wins, ps2.losses, ps2.outs_pitched, ps2.strikeouts
		FROM player_season_pitching_stats ps2
		JOIN player_seasons ps ON ps.id = ps2.player_season_id
		JOIN players p ON p.id = ps.player_id
		WHERE p.game_guid = 'BB000000000000000000000000000000'
		  AND ps2.is_regular_season = 1
	`).Scan(&wins, &losses, &outsPitched, &strikeouts)
	if err != nil {
		t.Fatalf("querying pitcher stats: %v", err)
	}
	if wins != 12 {
		t.Errorf("wins: got %d, want 12", wins)
	}
	if losses != 8 {
		t.Errorf("losses: got %d, want 8", losses)
	}
	if outsPitched != 540 {
		t.Errorf("outs_pitched: got %d, want 540 (180 IP)", outsPitched)
	}
	if strikeouts != 180 {
		t.Errorf("strikeouts: got %d, want 180", strikeouts)
	}
}

func TestImportSeason_ScheduleImported(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	result := importSeason1(t, svc, companionDB, reader)

	var count int
	if err := companionDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM team_season_schedules WHERE season_id = ?`, result.SeasonID,
	).Scan(&count); err != nil {
		t.Fatalf("counting schedule: %v", err)
	}
	if count < 1 {
		t.Errorf("expected schedule records, got %d", count)
	}
}

func TestImportSeason_GameStatAttributesImported(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	importSeason1(t, svc, companionDB, reader)

	// Fixture: player AA has Power=80, Contact=75
	var power, contact int
	err := companionDB.QueryRowContext(ctx, `
		SELECT gs.power, gs.contact
		FROM player_season_game_stats gs
		JOIN player_seasons ps ON ps.id = gs.player_season_id
		JOIN players p ON p.id = ps.player_id
		WHERE p.game_guid = 'AA000000000000000000000000000000'
	`).Scan(&power, &contact)
	if err != nil {
		t.Fatalf("querying game stats: %v", err)
	}
	if power != 80 {
		t.Errorf("power: got %d, want 80", power)
	}
	if contact != 75 {
		t.Errorf("contact: got %d, want 75", contact)
	}
}

// ── Idempotency ──────────────────────────────────────────────────────────────

func TestImportSeason_Idempotent(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	var firstID int64
	for i := range 3 {
		result, err := svc.ImportSeason(ctx, companionDB, reader, 100, 1, testLeagueGUID, 0)
		if err != nil {
			t.Fatalf("import attempt %d failed: %v", i+1, err)
		}
		if i == 0 {
			firstID = result.SeasonID
		} else if result.SeasonID != firstID {
			t.Errorf("expected same season ID on re-import, got %d then %d", firstID, result.SeasonID)
		}
	}

	var playerCount, seasonCount int
	_ = companionDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM players`).Scan(&playerCount)
	_ = companionDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM seasons`).Scan(&seasonCount)

	if seasonCount != 1 {
		t.Errorf("idempotency: expected 1 season record after 3 imports, got %d", seasonCount)
	}
	if playerCount < 1 {
		t.Errorf("idempotency: expected players after 3 imports, got %d", playerCount)
	}
}

func TestImportSeason_IdempotentBattingStats(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	_, _ = svc.ImportSeason(ctx, companionDB, reader, 100, 1, testLeagueGUID, 0)
	_, _ = svc.ImportSeason(ctx, companionDB, reader, 100, 1, testLeagueGUID, 0)

	var count int
	_ = companionDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM player_season_batting_stats WHERE is_regular_season = 1`,
	).Scan(&count)

	if count > 10 {
		t.Errorf("idempotency: batting stats look doubled — got %d rows (expected ~2 for 2 players)", count)
	}
}

// ── Multi-season tracking ────────────────────────────────────────────────────

func TestImportSeason_PlayerTrackedAcrossSeasons(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	_, err := svc.ImportSeason(ctx, companionDB, reader, 100, 1, testLeagueGUID, 0)
	if err != nil {
		t.Fatalf("ImportSeason 100: %v", err)
	}
	_, err = svc.ImportSeason(ctx, companionDB, reader, 101, 2, testLeagueGUID, 0)
	if err != nil {
		t.Fatalf("ImportSeason 101: %v", err)
	}

	var playerCount int
	if err := companionDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM players`).Scan(&playerCount); err != nil {
		t.Fatalf("counting players: %v", err)
	}
	if playerCount != 2 {
		t.Errorf("expected 2 unique players across 2 seasons, got %d", playerCount)
	}

	var psCount int
	if err := companionDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM player_seasons`).Scan(&psCount); err != nil {
		t.Fatalf("counting player_seasons: %v", err)
	}
	if psCount != 4 {
		t.Errorf("expected 4 player_season records (2 players × 2 seasons), got %d", psCount)
	}
}

func TestImportSeason_MidSeasonThenEndOfSeason(t *testing.T) {
	companionDB := testutil.NewTestDB(t)
	ctx := context.Background()
	svc := service.NewImportService()

	midSeasonDB := testutil.NewTestSaveGameDB_MidSeason(t, testutil.MidSeasonStats{
		Hits: 10, AtBats: 35, HomeRuns: 2,
	})
	midReader := store.NewSqliteSaveGameReader(midSeasonDB, "")
	_, err := svc.ImportSeason(ctx, companionDB, midReader, 100, 1, testLeagueGUID, 0)
	if err != nil {
		t.Fatalf("mid-season import: %v", err)
	}

	var midHits int
	_ = companionDB.QueryRowContext(ctx, `
		SELECT bs.hits
		FROM player_season_batting_stats bs
		JOIN player_seasons ps ON ps.id = bs.player_season_id
		JOIN players p ON p.id = ps.player_id
		WHERE p.game_guid = 'AA000000000000000000000000000000'
		  AND bs.is_regular_season = 1`).Scan(&midHits)
	if midHits != 10 {
		t.Errorf("mid-season hits: got %d, want 10", midHits)
	}

	endSeasonDB := testutil.NewTestSaveGameDB(t)
	endReader := store.NewSqliteSaveGameReader(endSeasonDB, "")
	_, err = svc.ImportSeason(ctx, companionDB, endReader, 100, 1, testLeagueGUID, 0)
	if err != nil {
		t.Fatalf("end-of-season import: %v", err)
	}

	var finalHits, finalAtBats, finalHR int
	err = companionDB.QueryRowContext(ctx, `
		SELECT bs.hits, bs.at_bats, bs.home_runs
		FROM player_season_batting_stats bs
		JOIN player_seasons ps ON ps.id = bs.player_season_id
		JOIN players p ON p.id = ps.player_id
		WHERE p.game_guid = 'AA000000000000000000000000000000'
		  AND bs.is_regular_season = 1`).Scan(&finalHits, &finalAtBats, &finalHR)
	if err != nil {
		t.Fatalf("querying end-of-season stats: %v", err)
	}
	if finalHits != 54 {
		t.Errorf("end-of-season hits: got %d, want 54", finalHits)
	}
	if finalAtBats != 180 {
		t.Errorf("end-of-season at_bats: got %d, want 180", finalAtBats)
	}
	if finalHR != 12 {
		t.Errorf("end-of-season home_runs: got %d, want 12", finalHR)
	}

	var seasonCount int
	_ = companionDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM seasons WHERE league_guid = ? AND save_game_season_id = 100`,
		testLeagueGUID,
	).Scan(&seasonCount)
	if seasonCount != 1 {
		t.Errorf("expected 1 season record, got %d", seasonCount)
	}
}

func TestImportSeason_TeamTrackedAcrossSeasons(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	_, _ = svc.ImportSeason(ctx, companionDB, reader, 100, 1, testLeagueGUID, 0)
	_, _ = svc.ImportSeason(ctx, companionDB, reader, 101, 2, testLeagueGUID, 0)

	var teamCount, histCount int
	_ = companionDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM teams`).Scan(&teamCount)
	_ = companionDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM team_season_history`).Scan(&histCount)

	if teamCount != 2 {
		t.Errorf("expected 2 teams, got %d", teamCount)
	}
	if histCount != 4 {
		t.Errorf("expected 4 team_season_history records, got %d", histCount)
	}
}

func TestImportSeason_MultiSeasonStatValues(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	_, _ = svc.ImportSeason(ctx, companionDB, reader, 100, 1, testLeagueGUID, 0)
	r2, _ := svc.ImportSeason(ctx, companionDB, reader, 101, 2, testLeagueGUID, 0)

	// Season 101 fixture: player AA has 190 AB, 60 H, 15 HR
	var atBats, hits, homeRuns int
	err := companionDB.QueryRowContext(ctx, `
		SELECT bs.at_bats, bs.hits, bs.home_runs
		FROM player_season_batting_stats bs
		JOIN player_seasons ps ON ps.id = bs.player_season_id
		JOIN players p ON p.id = ps.player_id
		WHERE p.game_guid = 'AA000000000000000000000000000000'
		  AND ps.season_id = ?
		  AND bs.is_regular_season = 1
	`, r2.SeasonID).Scan(&atBats, &hits, &homeRuns)
	if err != nil {
		t.Fatalf("querying season 2 batting stats: %v", err)
	}
	if atBats != 190 {
		t.Errorf("season 2 at_bats: got %d, want 190", atBats)
	}
	if hits != 60 {
		t.Errorf("season 2 hits: got %d, want 60", hits)
	}
	if homeRuns != 15 {
		t.Errorf("season 2 home_runs: got %d, want 15", homeRuns)
	}
}

// ── Context stats (Phase 8.5) ────────────────────────────────────────────────

func TestImportSeason_ContextStatsPopulated(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	result := importSeason1(t, svc, companionDB, reader)

	// league_season_stats should have one row for the regular season.
	var lssCount int
	if err := companionDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM league_season_stats WHERE season_id = ? AND is_regular_season = 1`,
		result.SeasonID,
	).Scan(&lssCount); err != nil {
		t.Fatalf("querying league_season_stats: %v", err)
	}
	if lssCount != 1 {
		t.Errorf("expected 1 league_season_stats row, got %d", lssCount)
	}

	// Batter AA (AB=180 > 0) should have non-NULL ops_plus and smb_war.
	var opsPlus, smbWARBat *float64
	if err := companionDB.QueryRowContext(ctx, `
		SELECT bs.ops_plus, bs.smb_war
		FROM player_season_batting_stats bs
		JOIN player_seasons ps ON ps.id = bs.player_season_id
		JOIN players p ON p.id = ps.player_id
		WHERE p.game_guid = 'AA000000000000000000000000000000'
		  AND bs.is_regular_season = 1
	`).Scan(&opsPlus, &smbWARBat); err != nil {
		t.Fatalf("querying batter context stats: %v", err)
	}
	if opsPlus == nil {
		t.Error("expected non-NULL ops_plus for batter AA (AB=180)")
	}
	if smbWARBat == nil {
		t.Error("expected non-NULL smb_war for batter AA")
	}

	// Pitcher BB (outs_pitched=540 > 0) should have non-NULL era_plus, fip, fip_minus, smb_war.
	var eraPlus, fip, fipMinus, smbWARPit *float64
	if err := companionDB.QueryRowContext(ctx, `
		SELECT ps2.era_plus, ps2.fip, ps2.fip_minus, ps2.smb_war
		FROM player_season_pitching_stats ps2
		JOIN player_seasons ps ON ps.id = ps2.player_season_id
		JOIN players p ON p.id = ps.player_id
		WHERE p.game_guid = 'BB000000000000000000000000000000'
		  AND ps2.is_regular_season = 1
	`).Scan(&eraPlus, &fip, &fipMinus, &smbWARPit); err != nil {
		t.Fatalf("querying pitcher context stats: %v", err)
	}
	if eraPlus == nil {
		t.Error("expected non-NULL era_plus for pitcher BB (outs_pitched=540)")
	}
	if fip == nil {
		t.Error("expected non-NULL fip for pitcher BB")
	}
	if fipMinus == nil {
		t.Error("expected non-NULL fip_minus for pitcher BB")
	}
	if smbWARPit == nil {
		t.Error("expected non-NULL smb_war for pitcher BB")
	}
}

func TestImportSeason_ContextStatsLeagueAverage(t *testing.T) {
	// With only one batter and one pitcher, each player IS the league average.
	// So OPS+ = 100 and ERA+ = 100 (they equal the league).
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	importSeason1(t, svc, companionDB, reader)

	var opsPlus, eraPlus float64
	if err := companionDB.QueryRowContext(ctx, `
		SELECT bs.ops_plus
		FROM player_season_batting_stats bs
		JOIN player_seasons ps ON ps.id = bs.player_season_id
		JOIN players p ON p.id = ps.player_id
		WHERE p.game_guid = 'AA000000000000000000000000000000'
		  AND bs.is_regular_season = 1
	`).Scan(&opsPlus); err != nil {
		t.Fatalf("querying ops_plus: %v", err)
	}
	// Single-player league: lgOBP == playerOBP, lgSLG == playerSLG → OPS+ = 100
	if opsPlus < 99.9 || opsPlus > 100.1 {
		t.Errorf("expected OPS+ ≈ 100 for sole batter, got %.2f", opsPlus)
	}

	if err := companionDB.QueryRowContext(ctx, `
		SELECT ps2.era_plus
		FROM player_season_pitching_stats ps2
		JOIN player_seasons ps ON ps.id = ps2.player_season_id
		JOIN players p ON p.id = ps.player_id
		WHERE p.game_guid = 'BB000000000000000000000000000000'
		  AND ps2.is_regular_season = 1
	`).Scan(&eraPlus); err != nil {
		t.Fatalf("querying era_plus: %v", err)
	}
	// Single-pitcher league: lgERA == pitcher ERA → ERA+ = 100
	if eraPlus < 99.9 || eraPlus > 100.1 {
		t.Errorf("expected ERA+ ≈ 100 for sole pitcher, got %.2f", eraPlus)
	}
}

func TestImportSeason_ContextStatsIdempotent(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	importSeason1(t, svc, companionDB, reader)
	importSeason1(t, svc, companionDB, reader) // second import should overwrite cleanly

	var lssCount int
	if err := companionDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM league_season_stats WHERE is_regular_season = 1`,
	).Scan(&lssCount); err != nil {
		t.Fatalf("querying league_season_stats: %v", err)
	}
	if lssCount != 1 {
		t.Errorf("idempotency: expected 1 league_season_stats row after 2 imports, got %d", lssCount)
	}
}

// ── Transaction atomicity ────────────────────────────────────────────────────

func TestImportSeason_TransactionRollbackOnError(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	_, err := svc.ImportSeason(ctx, companionDB, reader, 100, 1, testLeagueGUID, 0)
	if err != nil {
		t.Fatalf("baseline import failed: %v", err)
	}

	var before int
	_ = companionDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM players`).Scan(&before)

	_, err = svc.ImportSeason(ctx, companionDB, &erroringReader{inner: reader}, 101, 2, testLeagueGUID, 0)
	if err == nil {
		t.Error("expected error from erroring reader, got nil")
	}

	var after int
	_ = companionDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM players`).Scan(&after)

	var s101count int
	_ = companionDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM seasons WHERE league_guid = ? AND save_game_season_id = 101`,
		testLeagueGUID,
	).Scan(&s101count)
	if s101count != 0 {
		t.Errorf("expected season 101 to not exist after rollback, found %d records", s101count)
	}
	if after != before {
		t.Errorf("player count changed after failed import: before=%d after=%d", before, after)
	}
}

// ── Traits ───────────────────────────────────────────────────────────────────

func TestImportSeason_TraitsImported(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()
	importSeason1(t, svc, companionDB, reader)

	// Fixture seeds player AA (batter) with Clutch (32,6) and Tough Out (4,6).
	var traitsJSON string
	if err := companionDB.QueryRowContext(ctx, `
		SELECT ps.traits_json
		FROM player_seasons ps
		JOIN players p ON p.id = ps.player_id
		WHERE p.first_name = 'Test' AND p.last_name = 'Batter'
	`).Scan(&traitsJSON); err != nil {
		t.Fatalf("querying batter traits_json: %v", err)
	}
	if traitsJSON != `["Clutch","Tough Out"]` {
		t.Errorf("batter traits_json = %q, want [\"Clutch\",\"Tough Out\"]", traitsJSON)
	}
}

func TestImportSeason_PlayoffSeedsStored(t *testing.T) {
	// team1Standing=0 (0-indexed #1 seed) must be stored as playoff_seed=1.
	// team2Standing=1 (0-indexed #2 seed) must be stored as playoff_seed=2.
	// The > 0 guard that previously skipped 0-indexed #1 seeds is the regression
	// this test is protecting against.
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()
	importSeason1(t, svc, companionDB, reader)

	rows, err := companionDB.QueryContext(ctx,
		`SELECT t.game_guid, tsh.playoff_seed
		 FROM team_season_history tsh
		 JOIN teams t ON t.id = tsh.team_id
		 WHERE tsh.playoff_seed IS NOT NULL
		 ORDER BY tsh.playoff_seed`,
	)
	if err != nil {
		t.Fatalf("querying playoff seeds: %v", err)
	}
	defer func() { _ = rows.Close() }()

	type row struct {
		guid string
		seed int
	}
	var got []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.guid, &r.seed); err != nil {
			t.Fatalf("scanning row: %v", err)
		}
		got = append(got, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows error: %v", err)
	}

	want := []row{
		{"01000000000000000000000000000000", 1},
		{"02000000000000000000000000000000", 2},
	}
	if len(got) != len(want) {
		t.Fatalf("playoff seed rows: got %d, want %d; rows=%v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].guid != w.guid || got[i].seed != w.seed {
			t.Errorf("row %d: got {%s, %d}, want {%s, %d}", i, got[i].guid, got[i].seed, w.guid, w.seed)
		}
	}
}

func TestImportSeason_PersistsPlayoffConfig(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()
	result := importSeason1(t, svc, companionDB, reader)

	var rounds, seriesLength int64
	if err := companionDB.QueryRowContext(ctx,
		`SELECT playoff_rounds, playoff_series_length FROM seasons WHERE id = ?`,
		result.SeasonID,
	).Scan(&rounds, &seriesLength); err != nil {
		t.Fatalf("querying playoff config: %v", err)
	}
	// Fixture seeds t_playoffs with rounds=1, seriesLength=5 for season 100.
	if rounds != 1 {
		t.Errorf("playoff_rounds: want 1, got %d", rounds)
	}
	if seriesLength != 5 {
		t.Errorf("playoff_series_length: want 5, got %d", seriesLength)
	}
}

// erroringReader wraps a real reader but injects an error on GetCurrentSeasonTeams.
type erroringReader struct {
	inner store.SaveGameReader
}

func (r *erroringReader) Close() error { return nil }
func (r *erroringReader) GetLeagues(ctx context.Context) ([]models.SaveGameLeague, error) {
	return r.inner.GetLeagues(ctx)
}
func (r *erroringReader) GetFranchiseSeasons(ctx context.Context, leagueGUID string) ([]models.SaveGameFranchiseSeason, error) {
	return r.inner.GetFranchiseSeasons(ctx, leagueGUID)
}
func (r *erroringReader) GetCurrentSeasonPlayers(ctx context.Context, id int) ([]models.SaveGamePlayer, error) {
	return r.inner.GetCurrentSeasonPlayers(ctx, id)
}
func (r *erroringReader) GetCurrentSeasonTeams(ctx context.Context, _ int) ([]models.SaveGameTeam, error) {
	return nil, fmt.Errorf("injected error: GetCurrentSeasonTeams failed")
}
func (r *erroringReader) GetSeasonSchedule(ctx context.Context, id int) ([]models.SaveGameGame, error) {
	return r.inner.GetSeasonSchedule(ctx, id)
}
func (r *erroringReader) GetPlayoffSchedule(ctx context.Context, id int) ([]models.SaveGamePlayoffGame, error) {
	return r.inner.GetPlayoffSchedule(ctx, id)
}
func (r *erroringReader) GetSeasonBattingStats(ctx context.Context, id int) ([]models.SaveGameBattingStat, error) {
	return r.inner.GetSeasonBattingStats(ctx, id)
}
func (r *erroringReader) GetSeasonPitchingStats(ctx context.Context, id int) ([]models.SaveGamePitchingStat, error) {
	return r.inner.GetSeasonPitchingStats(ctx, id)
}
func (r *erroringReader) GetPlayoffBattingStats(ctx context.Context, id int) ([]models.SaveGameBattingStat, error) {
	return r.inner.GetPlayoffBattingStats(ctx, id)
}
func (r *erroringReader) GetPlayoffPitchingStats(ctx context.Context, id int) ([]models.SaveGamePitchingStat, error) {
	return r.inner.GetPlayoffPitchingStats(ctx, id)
}
func (r *erroringReader) GetCareerBattingStats(ctx context.Context) ([]models.SaveGameBattingStat, error) {
	return r.inner.GetCareerBattingStats(ctx)
}
func (r *erroringReader) GetCareerPitchingStats(ctx context.Context) ([]models.SaveGamePitchingStat, error) {
	return r.inner.GetCareerPitchingStats(ctx)
}
func (r *erroringReader) GetCurrentSeason(ctx context.Context, leagueGUID string) (models.SaveGameSeasonInfo, error) {
	return r.inner.GetCurrentSeason(ctx, leagueGUID)
}
func (r *erroringReader) GetSeasonPlayoffConfig(ctx context.Context, seasonID int) (*models.SaveGamePlayoffConfig, error) {
	return r.inner.GetSeasonPlayoffConfig(ctx, seasonID)
}
