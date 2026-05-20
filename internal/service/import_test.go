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

// newTestImportService builds an ImportService backed by a fresh in-memory companion DB.
// Returns the service, the companion DB, and a SaveGameReader over the test save game DB.
func newTestImportService(t *testing.T) (*service.ImportService, *sql.DB, store.SaveGameReader) {
	t.Helper()
	companionDB := testutil.NewTestDB(t)
	saveGameDB := testutil.NewTestSaveGameDB(t)
	reader := store.NewSqliteSaveGameReader(saveGameDB, "")

	svc := service.NewImportService()
	return svc, companionDB, reader
}

// ── Basic flow ───────────────────────────────────────────────────────────────

func TestImportSeason_BasicFlow(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	result, err := svc.ImportSeason(ctx, companionDB, reader, 100, 1)
	if err != nil {
		t.Fatalf("ImportSeason: %v", err)
	}

	if result.SeasonID != 100 {
		t.Errorf("SeasonID: got %d, want 100", result.SeasonID)
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

	_, err := svc.ImportSeason(ctx, companionDB, reader, 100, 1)
	if err != nil {
		t.Fatalf("ImportSeason: %v", err)
	}

	ss := store.NewSeasonStore(companionDB)
	season, err := ss.GetByID(ctx, 100)
	if err != nil {
		t.Fatalf("GetByID after import: %v", err)
	}
	if season.SeasonNum != 1 {
		t.Errorf("season_num: got %d, want 1", season.SeasonNum)
	}
}

func TestImportSeason_PlayersCreated(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	_, err := svc.ImportSeason(ctx, companionDB, reader, 100, 1)
	if err != nil {
		t.Fatalf("ImportSeason: %v", err)
	}

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

	_, err := svc.ImportSeason(ctx, companionDB, reader, 100, 1)
	if err != nil {
		t.Fatalf("ImportSeason: %v", err)
	}

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

	_, err := svc.ImportSeason(ctx, companionDB, reader, 100, 1)
	if err != nil {
		t.Fatalf("ImportSeason: %v", err)
	}

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

	_, err := svc.ImportSeason(ctx, companionDB, reader, 100, 1)
	if err != nil {
		t.Fatalf("ImportSeason: %v", err)
	}

	// The fixture seeds player AA with 180 AB, 54 H, 12 HR, 40 RBI
	var atBats, hits, homeRuns, rbi int
	err = companionDB.QueryRowContext(ctx, `
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

	_, err := svc.ImportSeason(ctx, companionDB, reader, 100, 1)
	if err != nil {
		t.Fatalf("ImportSeason: %v", err)
	}

	// Fixture: pitcher BB with W=12, L=8, outsPitched=540 (= 180 IP), K=180
	var wins, losses, outsPitched, strikeouts int
	err = companionDB.QueryRowContext(ctx, `
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

	_, err := svc.ImportSeason(ctx, companionDB, reader, 100, 1)
	if err != nil {
		t.Fatalf("ImportSeason: %v", err)
	}

	var count int
	if err := companionDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM team_season_schedules WHERE season_id = 100`,
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

	_, err := svc.ImportSeason(ctx, companionDB, reader, 100, 1)
	if err != nil {
		t.Fatalf("ImportSeason: %v", err)
	}

	// Fixture: player AA has Power=80, Contact=75
	var power, contact int
	err = companionDB.QueryRowContext(ctx, `
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

	for i := range 3 {
		if _, err := svc.ImportSeason(ctx, companionDB, reader, 100, 1); err != nil {
			t.Fatalf("import attempt %d failed: %v", i+1, err)
		}
	}

	// After 3 imports, should still have exactly the right number of records
	var playerCount, seasonCount int
	_ = companionDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM players`).Scan(&playerCount)
	_ = companionDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM seasons`).Scan(&seasonCount)

	if seasonCount != 1 {
		t.Errorf("idempotency: expected 1 season record after 3 imports, got %d", seasonCount)
	}
	// Players created once and reused
	if playerCount < 1 {
		t.Errorf("idempotency: expected players after 3 imports, got %d", playerCount)
	}
}

func TestImportSeason_IdempotentBattingStats(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	_, _ = svc.ImportSeason(ctx, companionDB, reader, 100, 1)
	_, _ = svc.ImportSeason(ctx, companionDB, reader, 100, 1)

	var count int
	_ = companionDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM player_season_batting_stats WHERE is_regular_season = 1`,
	).Scan(&count)

	// Should not have doubled up
	if count > 10 {
		t.Errorf("idempotency: batting stats look doubled — got %d rows (expected ~2 for 2 players)", count)
	}
}

// ── Multi-season tracking ────────────────────────────────────────────────────

func TestImportSeason_PlayerTrackedAcrossSeasons(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	// Import season 100 and 101 — both have the same two players
	_, err := svc.ImportSeason(ctx, companionDB, reader, 100, 1)
	if err != nil {
		t.Fatalf("ImportSeason 100: %v", err)
	}
	_, err = svc.ImportSeason(ctx, companionDB, reader, 101, 2)
	if err != nil {
		t.Fatalf("ImportSeason 101: %v", err)
	}

	// There should be exactly 2 player records (same players, different seasons)
	var playerCount int
	if err := companionDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM players`).Scan(&playerCount); err != nil {
		t.Fatalf("counting players: %v", err)
	}
	if playerCount != 2 {
		t.Errorf("expected 2 unique players across 2 seasons, got %d", playerCount)
	}

	// But 4 player_season records (2 players × 2 seasons)
	var psCount int
	if err := companionDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM player_seasons`).Scan(&psCount); err != nil {
		t.Fatalf("counting player_seasons: %v", err)
	}
	if psCount != 4 {
		t.Errorf("expected 4 player_season records (2 players × 2 seasons), got %d", psCount)
	}
}


// TestImportSeason_MidSeasonThenEndOfSeason verifies that importing mid-season
// (partial stats) followed by an end-of-season import (full stats) correctly
// reflects the final state. This is the core idempotency guarantee: the last
// import wins, so users can sync freely during the season and the final
// end-of-season sync will always produce the definitive record.
func TestImportSeason_MidSeasonThenEndOfSeason(t *testing.T) {
	companionDB := testutil.NewTestDB(t)
	ctx := context.Background()
	svc := service.NewImportService()

	// First import: mid-season (player AA has 10 hits, 35 AB, 2 HR)
	midSeasonDB := testutil.NewTestSaveGameDB_MidSeason(t, testutil.MidSeasonStats{
		Hits: 10, AtBats: 35, HomeRuns: 2,
	})
	midReader := store.NewSqliteSaveGameReader(midSeasonDB, "")
	_, err := svc.ImportSeason(ctx, companionDB, midReader, 100, 1)
	if err != nil {
		t.Fatalf("mid-season import: %v", err)
	}

	// Verify mid-season stats were recorded
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

	// Second import: end of season (player AA has the full 54 hits, 180 AB, 12 HR)
	endSeasonDB := testutil.NewTestSaveGameDB(t)
	endReader := store.NewSqliteSaveGameReader(endSeasonDB, "")
	_, err = svc.ImportSeason(ctx, companionDB, endReader, 100, 1)
	if err != nil {
		t.Fatalf("end-of-season import: %v", err)
	}

	// Final values must reflect the end-of-season snapshot, not the mid-season one.
	// The mid-season 10 hits should be gone; 54 hits is the truth.
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
		t.Errorf("end-of-season hits: got %d, want 54 (mid-season value should be overwritten)", finalHits)
	}
	if finalAtBats != 180 {
		t.Errorf("end-of-season at_bats: got %d, want 180", finalAtBats)
	}
	if finalHR != 12 {
		t.Errorf("end-of-season home_runs: got %d, want 12", finalHR)
	}

	// Only one season record (no duplication from two imports)
	var seasonCount int
	_ = companionDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM seasons WHERE id = 100`).Scan(&seasonCount)
	if seasonCount != 1 {
		t.Errorf("expected 1 season record, got %d", seasonCount)
	}
}
func TestImportSeason_TeamTrackedAcrossSeasons(t *testing.T) {
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	_, _ = svc.ImportSeason(ctx, companionDB, reader, 100, 1)
	_, _ = svc.ImportSeason(ctx, companionDB, reader, 101, 2)

	// 2 unique teams, 4 team_season_history records (2 teams × 2 seasons)
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

	_, _ = svc.ImportSeason(ctx, companionDB, reader, 100, 1)
	_, _ = svc.ImportSeason(ctx, companionDB, reader, 101, 2)

	// Season 101 fixture: player AA has 190 AB, 60 H, 15 HR, 48 RBI
	var atBats, hits, homeRuns int
	err := companionDB.QueryRowContext(ctx, `
		SELECT bs.at_bats, bs.hits, bs.home_runs
		FROM player_season_batting_stats bs
		JOIN player_seasons ps ON ps.id = bs.player_season_id
		JOIN players p ON p.id = ps.player_id
		WHERE p.game_guid = 'AA000000000000000000000000000000'
		  AND ps.season_id = 101
		  AND bs.is_regular_season = 1
	`).Scan(&atBats, &hits, &homeRuns)
	if err != nil {
		t.Fatalf("querying season 101 batting stats: %v", err)
	}
	if atBats != 190 {
		t.Errorf("season 101 at_bats: got %d, want 190", atBats)
	}
	if hits != 60 {
		t.Errorf("season 101 hits: got %d, want 60", hits)
	}
	if homeRuns != 15 {
		t.Errorf("season 101 home_runs: got %d, want 15", homeRuns)
	}
}

// ── Transaction atomicity ────────────────────────────────────────────────────

func TestImportSeason_TransactionRollbackOnError(t *testing.T) {
	// Use a closed/broken companion DB to trigger a mid-import failure.
	// All partial data should be rolled back.
	svc, companionDB, reader := newTestImportService(t)
	ctx := context.Background()

	// First import succeeds
	_, err := svc.ImportSeason(ctx, companionDB, reader, 100, 1)
	if err != nil {
		t.Fatalf("baseline import failed: %v", err)
	}

	var before int
	_ = companionDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM players`).Scan(&before)

	// Simulate a broken reader that errors partway through
	_, err = svc.ImportSeason(ctx, companionDB, &erroringReader{inner: reader}, 101, 2)
	if err == nil {
		t.Error("expected error from erroring reader, got nil")
	}

	// Player count should be unchanged (transaction rolled back)
	var after int
	_ = companionDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM players`).Scan(&after)
	// After rollback, no NEW season records should exist for season 101
	var s101count int
	_ = companionDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM seasons WHERE id = 101`,
	).Scan(&s101count)
	if s101count != 0 {
		t.Errorf("expected season 101 to not exist after rollback, found %d records", s101count)
	}
	// Players from season 100 should still be intact
	if after != before {
		t.Errorf("player count changed after failed import: before=%d after=%d", before, after)
	}
}

// erroringReader wraps a real reader but injects an error on GetCurrentSeasonTeams
// to simulate a mid-import failure for rollback testing.
type erroringReader struct {
	inner store.SaveGameReader
}

func (r *erroringReader) Close() error { return nil }
func (r *erroringReader) GetLeagues(ctx context.Context) ([]models.SaveGameLeague, error) {
	return r.inner.GetLeagues(ctx)
}
func (r *erroringReader) GetFranchiseSeasons(ctx context.Context, id int) ([]models.SaveGameFranchiseSeason, error) {
	return r.inner.GetFranchiseSeasons(ctx, id)
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
func (r *erroringReader) GetCareerBattingStats(ctx context.Context) ([]models.SaveGameBattingStat, error) {
	return r.inner.GetCareerBattingStats(ctx)
}
func (r *erroringReader) GetCareerPitchingStats(ctx context.Context) ([]models.SaveGamePitchingStat, error) {
	return r.inner.GetCareerPitchingStats(ctx)
}
func (r *erroringReader) GetCurrentSeason(ctx context.Context, leagueGUID string) (models.SaveGameSeasonInfo, error) {
	return r.inner.GetCurrentSeason(ctx, leagueGUID)
}
