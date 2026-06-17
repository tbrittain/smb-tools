package service_test

import (
	"context"
	"database/sql"
	"math"
	"testing"

	"smb-tools/internal/models"
	"smb-tools/internal/service"
	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func csNewDB(t *testing.T) *sql.DB {
	t.Helper()
	return testutil.NewTestDB(t)
}

func csSeedSeason(t *testing.T, db *sql.DB, sgID, num int) int64 {
	t.Helper()
	ss := store.NewSeasonStore(db)
	id, err := ss.Upsert(context.Background(), store.Season{
		LeagueGUID:       "TEST-GUID-0000",
		SaveGameSeasonID: sgID,
		SeasonNum:        num,
		NumGames:         40,
	})
	if err != nil {
		t.Fatalf("csSeedSeason: %v", err)
	}
	return id
}

func csSeedPlayer(t *testing.T, db *sql.DB, guid string) int64 {
	t.Helper()
	ps := store.NewPlayerSeasonStore(db)
	id, err := ps.UpsertPlayer(context.Background(), store.PlayerIdentity{
		GameGUID:      guid,
		FirstName:     "Test",
		LastName:      "Player",
		BatHand:       "R",
		ThrowHand:     "R",
		ChemistryType: "Competitive",
	})
	if err != nil {
		t.Fatalf("csSeedPlayer: %v", err)
	}
	return id
}

func csSeedPlayerSeason(t *testing.T, db *sql.DB, playerID, seasonID int64) int64 {
	t.Helper()
	ps := store.NewPlayerSeasonStore(db)
	id, err := ps.UpsertSeason(context.Background(), store.PlayerSeason{
		PlayerID:        playerID,
		SeasonID:        seasonID,
		PrimaryPosition: "CF",
		PitcherRole:     "",
		BatHand:         "R",
		ThrowHand:       "R",
		ChemistryType:   "Competitive",
	})
	if err != nil {
		t.Fatalf("csSeedPlayerSeason: %v", err)
	}
	return id
}

// csSeedBatting inserts a batting row via UpsertBattingStats with computed rates.
func csSeedBatting(t *testing.T, db *sql.DB, psID int64, isReg bool, ab, hits, hr, rbi, bb, hbp, sf int) {
	t.Helper()
	tmp := models.CareerBattingStats{
		AtBats: ab, Hits: hits, HomeRuns: hr, Walks: bb,
		HitByPitch: hbp, SacFlies: sf,
	}
	service.ComputeBattingRates(&tmp)
	ps := store.NewPlayerSeasonStore(db)
	if err := ps.UpsertBattingStats(context.Background(), store.PlayerSeasonBattingStats{
		PlayerSeasonID:  psID,
		IsRegularSeason: isReg,
		GamesPlayed:     ab / 4,
		GamesBatting:    ab / 4,
		AtBats:          ab,
		Hits:            hits,
		HomeRuns:        hr,
		RBI:             rbi,
		Walks:           bb,
		HitByPitch:      hbp,
		SacFlies:        sf,
		BA:              tmp.BA,
		OBP:             tmp.OBP,
		SLG:             tmp.SLG,
		OPS:             tmp.OPS,
		ISO:             tmp.ISO,
		BABIP:           tmp.BABIP,
		KPct:            tmp.KPct,
		BBPct:           tmp.BBPct,
		ABPerHR:         tmp.ABPerHR,
	}); err != nil {
		t.Fatalf("csSeedBatting: %v", err)
	}
}

// csSeedPitching inserts a pitching row via UpsertPitchingStats with computed rates.
func csSeedPitching(t *testing.T, db *sql.DB, psID int64, isReg bool, w, l, outs, er, hr, bb, k int) {
	t.Helper()
	tmp := models.CareerPitchingStats{
		Wins: w, Losses: l, OutsPitched: outs,
		EarnedRuns: er, HomeRunsAllowed: hr, Walks: bb, Strikeouts: k,
	}
	service.ComputePitchingRates(&tmp)
	ps := store.NewPlayerSeasonStore(db)
	if err := ps.UpsertPitchingStats(context.Background(), store.PlayerSeasonPitchingStats{
		PlayerSeasonID:  psID,
		IsRegularSeason: isReg,
		Wins:            w,
		Losses:          l,
		Games:           w + l,
		GamesStarted:    w + l,
		OutsPitched:     outs,
		EarnedRuns:      er,
		HomeRunsAllowed: hr,
		Walks:           bb,
		Strikeouts:      k,
		ERA:             tmp.ERA,
		WHIP:            tmp.WHIP,
		K9:              tmp.K9,
		BB9:             tmp.BB9,
		H9:              tmp.H9,
		HR9:             tmp.HR9,
		KPerBB:          tmp.KPerBB,
		KPct:            tmp.KPct,
		WinPct:          tmp.WinPct,
		PPerIP:          tmp.PPerIP,
	}); err != nil {
		t.Fatalf("csSeedPitching: %v", err)
	}
}

// csReadCareerBatting queries player_career_batting_stats for the given stat_type.
func csReadCareerBatting(t *testing.T, db *sql.DB, playerID int64, st models.CareerStatType) *models.CareerBattingStats {
	t.Helper()
	var b models.CareerBattingStats
	var ba, obp sql.NullFloat64
	err := db.QueryRowContext(context.Background(), `
SELECT at_bats, hits, home_runs, rbi, walks, ba, obp, ops_plus, smb_war
FROM player_career_batting_stats
WHERE player_id = ? AND stat_type = ?
`, playerID, st).Scan(
		&b.AtBats, &b.Hits, &b.HomeRuns, &b.RBI, &b.Walks,
		&ba, &obp, &b.OPSPlus, &b.SmbWAR,
	)
	if err != nil {
		return nil
	}
	if ba.Valid {
		b.BA = &ba.Float64
	}
	return &b
}

// csReadCareerPitching queries player_career_pitching_stats for the given stat_type.
func csReadCareerPitching(t *testing.T, db *sql.DB, playerID int64, st models.CareerStatType) *models.CareerPitchingStats {
	t.Helper()
	var p models.CareerPitchingStats
	var era sql.NullFloat64
	err := db.QueryRowContext(context.Background(), `
SELECT wins, losses, outs_pitched, earned_runs, strikeouts, era, era_plus, smb_war
FROM player_career_pitching_stats
WHERE player_id = ? AND stat_type = ?
`, playerID, st).Scan(
		&p.Wins, &p.Losses, &p.OutsPitched, &p.EarnedRuns, &p.Strikeouts,
		&era, &p.ERAPlus, &p.SmbWAR,
	)
	if err != nil {
		return nil
	}
	if era.Valid {
		p.ERA = &era.Float64
	}
	return &p
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestApplyCareerStats_SingleRegularSeason(t *testing.T) {
	db := csNewDB(t)
	ctx := context.Background()

	sid := csSeedSeason(t, db, 1, 1)
	pid := csSeedPlayer(t, db, "g1")
	psid := csSeedPlayerSeason(t, db, pid, sid)
	// 400 AB, 120 H, 20 HR, 80 RBI, 40 BB, 5 HBP, 3 SF
	csSeedBatting(t, db, psid, true, 400, 120, 20, 80, 40, 5, 3)

	if err := service.ApplyCareerStats(ctx, db, []int64{pid}); err != nil {
		t.Fatalf("ApplyCareerStats: %v", err)
	}

	b := csReadCareerBatting(t, db, pid, models.CareerStatTypeRegularSeason)
	if b == nil {
		t.Fatal("expected regular_season career batting row")
		return
	}
	if b.AtBats != 400 {
		t.Errorf("AB: want 400, got %d", b.AtBats)
	}
	if b.Hits != 120 {
		t.Errorf("H: want 120, got %d", b.Hits)
	}
	if b.BA == nil {
		t.Fatal("BA should be non-nil")
	}
	wantBA := 120.0 / 400.0
	if math.Abs(*b.BA-wantBA) > 1e-9 {
		t.Errorf("BA: want %.6f, got %.6f", wantBA, *b.BA)
	}
}

func TestApplyCareerStats_TwoSeasons_BAFromSummedCounts(t *testing.T) {
	db := csNewDB(t)
	ctx := context.Background()

	sid1 := csSeedSeason(t, db, 1, 1)
	sid2 := csSeedSeason(t, db, 2, 2)
	pid := csSeedPlayer(t, db, "g1")
	psid1 := csSeedPlayerSeason(t, db, pid, sid1)
	psid2 := csSeedPlayerSeason(t, db, pid, sid2)

	// Season 1: 400 AB, 120 H → BA=.300
	// Season 2: 500 AB, 125 H → BA=.250
	// Career:   900 AB, 245 H → BA=245/900≈.2722 (NOT average of .300 and .250)
	csSeedBatting(t, db, psid1, true, 400, 120, 10, 50, 30, 0, 2)
	csSeedBatting(t, db, psid2, true, 500, 125, 15, 60, 40, 0, 3)

	if err := service.ApplyCareerStats(ctx, db, []int64{pid}); err != nil {
		t.Fatalf("ApplyCareerStats: %v", err)
	}

	b := csReadCareerBatting(t, db, pid, models.CareerStatTypeRegularSeason)
	if b == nil {
		t.Fatal("expected regular_season career batting row")
		return
	}
	if b.AtBats != 900 {
		t.Errorf("career AB: want 900, got %d", b.AtBats)
	}
	if b.Hits != 245 {
		t.Errorf("career H: want 245, got %d", b.Hits)
	}
	if b.BA == nil {
		t.Fatal("BA should be non-nil")
	}
	wantBA := 245.0 / 900.0
	if math.Abs(*b.BA-wantBA) > 1e-9 {
		t.Errorf("career BA (from summed counts): want %.6f, got %.6f", wantBA, *b.BA)
	}
	// Verify it's NOT the average of per-season BAs (that would be .275)
	avgOfAvgs := (120.0/400.0 + 125.0/500.0) / 2.0
	if math.Abs(*b.BA-avgOfAvgs) < 1e-9 {
		t.Error("career BA appears to be averaging per-season BAs instead of using summed counts")
	}
}

func TestApplyCareerStats_RegularAndPlayoff_AllThreeRows(t *testing.T) {
	db := csNewDB(t)
	ctx := context.Background()

	sid := csSeedSeason(t, db, 1, 1)
	pid := csSeedPlayer(t, db, "g1")
	psid := csSeedPlayerSeason(t, db, pid, sid)

	// Regular season: 400 AB, 120 H, 20 HR
	csSeedBatting(t, db, psid, true, 400, 120, 20, 80, 40, 5, 3)
	// Playoffs: 50 AB, 15 H, 3 HR
	csSeedBatting(t, db, psid, false, 50, 15, 3, 10, 5, 0, 0)

	if err := service.ApplyCareerStats(ctx, db, []int64{pid}); err != nil {
		t.Fatalf("ApplyCareerStats: %v", err)
	}

	reg := csReadCareerBatting(t, db, pid, models.CareerStatTypeRegularSeason)
	if reg == nil {
		t.Fatal("expected regular_season row")
		return
	}
	if reg.AtBats != 400 {
		t.Errorf("reg AB: want 400, got %d", reg.AtBats)
	}

	po := csReadCareerBatting(t, db, pid, models.CareerStatTypePlayoffs)
	if po == nil {
		t.Fatal("expected playoffs row")
		return
	}
	if po.AtBats != 50 {
		t.Errorf("playoff AB: want 50, got %d", po.AtBats)
	}

	tot := csReadCareerBatting(t, db, pid, models.CareerStatTypeTotalCareer)
	if tot == nil {
		t.Fatal("expected total_career row")
		return
	}
	// total_career counts = sum of reg + playoffs
	if tot.AtBats != 450 {
		t.Errorf("total AB: want 450, got %d", tot.AtBats)
	}
	if tot.Hits != 135 {
		t.Errorf("total H: want 135, got %d", tot.Hits)
	}
	// total_career BA from combined counts: 135/450 = .300
	if tot.BA == nil {
		t.Fatal("total_career BA should be non-nil")
	}
	wantTotalBA := 135.0 / 450.0
	if math.Abs(*tot.BA-wantTotalBA) > 1e-9 {
		t.Errorf("total_career BA: want %.6f, got %.6f", wantTotalBA, *tot.BA)
	}
}

func TestApplyCareerStats_PlayerWithNoBattingRows_NoCareerRow(t *testing.T) {
	db := csNewDB(t)
	ctx := context.Background()

	sid := csSeedSeason(t, db, 1, 1)
	pid := csSeedPlayer(t, db, "g1")
	psid := csSeedPlayerSeason(t, db, pid, sid)
	// Only pitching — no batting rows at all.
	csSeedPitching(t, db, psid, true, 10, 5, 270, 30, 5, 40, 100)

	if err := service.ApplyCareerStats(ctx, db, []int64{pid}); err != nil {
		t.Fatalf("ApplyCareerStats: %v", err)
	}

	b := csReadCareerBatting(t, db, pid, models.CareerStatTypeRegularSeason)
	if b != nil {
		t.Error("expected no career batting row for pitcher-only player")
	}

	p := csReadCareerPitching(t, db, pid, models.CareerStatTypeRegularSeason)
	if p == nil {
		t.Fatal("expected career pitching row")
		return
	}
	if p.OutsPitched != 270 {
		t.Errorf("outs pitched: want 270, got %d", p.OutsPitched)
	}
	if p.ERA == nil {
		t.Fatal("ERA should be non-nil")
	}
	// ERA = 30 ER * 27 / 270 outs = 3.00
	wantERA := 30.0 * 27.0 / 270.0
	if math.Abs(*p.ERA-wantERA) > 1e-9 {
		t.Errorf("career ERA: want %.4f, got %.4f", wantERA, *p.ERA)
	}
}

func TestApplyCareerStats_OPSPlusNullWithoutLeagueData(t *testing.T) {
	db := csNewDB(t)
	ctx := context.Background()

	sid := csSeedSeason(t, db, 1, 1)
	pid := csSeedPlayer(t, db, "g1")
	psid := csSeedPlayerSeason(t, db, pid, sid)
	csSeedBatting(t, db, psid, true, 400, 120, 20, 80, 40, 5, 3)

	// No league_season_stats row → OPS+ should be NULL.
	if err := service.ApplyCareerStats(ctx, db, []int64{pid}); err != nil {
		t.Fatalf("ApplyCareerStats: %v", err)
	}

	b := csReadCareerBatting(t, db, pid, models.CareerStatTypeRegularSeason)
	if b == nil {
		t.Fatal("expected career batting row")
		return
	}
	if b.OPSPlus != nil {
		t.Errorf("OPS+ should be nil without league data, got %v", *b.OPSPlus)
	}
	// BA should still be computed even without league data.
	if b.BA == nil {
		t.Error("BA should be non-nil regardless of league data")
	}
}

func TestApplyCareerStats_Idempotent(t *testing.T) {
	db := csNewDB(t)
	ctx := context.Background()

	sid := csSeedSeason(t, db, 1, 1)
	pid := csSeedPlayer(t, db, "g1")
	psid := csSeedPlayerSeason(t, db, pid, sid)
	csSeedBatting(t, db, psid, true, 400, 120, 20, 80, 40, 5, 3)

	// Apply twice — should not create duplicates.
	if err := service.ApplyCareerStats(ctx, db, []int64{pid}); err != nil {
		t.Fatalf("ApplyCareerStats (1st): %v", err)
	}
	if err := service.ApplyCareerStats(ctx, db, []int64{pid}); err != nil {
		t.Fatalf("ApplyCareerStats (2nd): %v", err)
	}

	var count int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM player_career_batting_stats WHERE player_id = ?`, pid,
	).Scan(&count); err != nil {
		t.Fatalf("counting career rows: %v", err)
	}
	// regular_season + total_career (no playoff data seeded → 2 rows, not 3).
	if count != 2 {
		t.Errorf("expected 2 career batting rows after two applies (no duplication), got %d", count)
	}
}

func TestApplyCareerStats_CareerSmbWARSumsSeasonValues(t *testing.T) {
	db := csNewDB(t)
	ctx := context.Background()

	sid1 := csSeedSeason(t, db, 1, 1)
	sid2 := csSeedSeason(t, db, 2, 2)
	pid := csSeedPlayer(t, db, "g1")
	psid1 := csSeedPlayerSeason(t, db, pid, sid1)
	psid2 := csSeedPlayerSeason(t, db, pid, sid2)

	csSeedBatting(t, db, psid1, true, 400, 120, 20, 80, 40, 5, 3)
	csSeedBatting(t, db, psid2, true, 500, 125, 15, 60, 30, 0, 2)

	// Manually set per-season smb_war values (normally set by ApplyContextStats).
	_, err := db.ExecContext(ctx,
		`UPDATE player_season_batting_stats SET smb_war = 2.5 WHERE player_season_id = ? AND is_regular_season = 1`, psid1)
	if err != nil {
		t.Fatalf("setting smb_war season 1: %v", err)
	}
	_, err = db.ExecContext(ctx,
		`UPDATE player_season_batting_stats SET smb_war = 1.8 WHERE player_season_id = ? AND is_regular_season = 1`, psid2)
	if err != nil {
		t.Fatalf("setting smb_war season 2: %v", err)
	}

	if err := service.ApplyCareerStats(ctx, db, []int64{pid}); err != nil {
		t.Fatalf("ApplyCareerStats: %v", err)
	}

	b := csReadCareerBatting(t, db, pid, models.CareerStatTypeRegularSeason)
	if b == nil {
		t.Fatal("expected career batting row")
		return
	}
	if b.SmbWAR == nil {
		t.Fatal("career smbWAR should be non-nil")
	}
	wantWAR := 2.5 + 1.8
	if math.Abs(*b.SmbWAR-wantWAR) > 1e-9 {
		t.Errorf("career smbWAR (sum of seasons): want %.4f, got %.4f", wantWAR, *b.SmbWAR)
	}
}

func TestApplyCareerStats_PitcherSingleSeason(t *testing.T) {
	db := csNewDB(t)
	ctx := context.Background()

	sid := csSeedSeason(t, db, 1, 1)
	pid := csSeedPlayer(t, db, "g1")
	psid := csSeedPlayerSeason(t, db, pid, sid)
	// 15 W, 8 L, 600 outs (200 IP), 50 ER, 10 HR, 60 BB, 200 K
	csSeedPitching(t, db, psid, true, 15, 8, 600, 50, 10, 60, 200)

	if err := service.ApplyCareerStats(ctx, db, []int64{pid}); err != nil {
		t.Fatalf("ApplyCareerStats: %v", err)
	}

	p := csReadCareerPitching(t, db, pid, models.CareerStatTypeRegularSeason)
	if p == nil {
		t.Fatal("expected career pitching row")
		return
	}
	if p.Wins != 15 {
		t.Errorf("Wins: want 15, got %d", p.Wins)
	}
	if p.ERA == nil {
		t.Fatal("ERA should be non-nil")
	}
	// ERA = 50 * 27 / 600 = 2.25
	wantERA := 50.0 * 27.0 / 600.0
	if math.Abs(*p.ERA-wantERA) > 1e-9 {
		t.Errorf("ERA: want %.4f, got %.4f", wantERA, *p.ERA)
	}
}
