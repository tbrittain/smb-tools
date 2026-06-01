package store_test

import (
	"context"
	"database/sql"
	"slices"
	"testing"

	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

// seedFullBatting inserts a batting row with all visible counting stat fields set.
func seedFullBatting(t *testing.T, db *sql.DB, playerSeasonID int64, isReg bool,
	gp, ab, h, d, tr, hr, rbi, sb, bb, k int,
) {
	t.Helper()
	isRegInt := 0
	if isReg {
		isRegInt = 1
	}
	_, err := db.ExecContext(context.Background(), `
INSERT INTO player_season_batting_stats
    (player_season_id, is_regular_season, games_played, games_batting,
     at_bats, runs, hits, doubles, triples, home_runs, rbi,
     stolen_bases, caught_stealing, walks, strikeouts, hit_by_pitch,
     sac_hits, sac_flies, errors, passed_balls)
VALUES (?,?,?,?,?,0,?,?,?,?,?,?,0,?,?,0,0,0,0,0)
`, playerSeasonID, isRegInt, gp, gp, ab, h, d, tr, hr, rbi, sb, bb, k)
	if err != nil {
		t.Fatalf("seedFullBatting: %v", err)
	}
}

// seedFullPitching inserts a pitching row with all visible counting stat fields set.
func seedFullPitching(t *testing.T, db *sql.DB, playerSeasonID int64, isReg bool,
	g, gs, w, l, sv, outs, k, bb, h, er int,
) {
	t.Helper()
	isRegInt := 0
	if isReg {
		isRegInt = 1
	}
	_, err := db.ExecContext(context.Background(), `
INSERT INTO player_season_pitching_stats
    (player_season_id, is_regular_season, wins, losses, games, games_started,
     complete_games, shutouts, saves, outs_pitched, hits_allowed, earned_runs,
     home_runs_allowed, walks, strikeouts, hit_batters, batters_faced,
     games_finished, runs_allowed, wild_pitches, total_pitches)
VALUES (?,?,?,?,?,?,0,0,?,?,?,?,0,?,?,0,?,0,0,0,0)
`, playerSeasonID, isRegInt, w, l, g, gs, sv, outs, h, er, bb, k, g)
	if err != nil {
		t.Fatalf("seedFullPitching: %v", err)
	}
}

func newStatRecordStore(db *sql.DB) *store.StatRecordQueryStore {
	return store.NewStatRecordQueryStore(db)
}

func TestGetBattingCountRows_RegularSeasonOnly(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team A", "E", "AL", 20, 20)

	pA := seedPlayer(t, db, "gA", "Alice", "Alpha")
	pB := seedPlayer(t, db, "gB", "Bob", "Beta")

	psA := seedPlayerSeason(t, db, pA, 1, &h1)
	psB := seedPlayerSeason(t, db, pB, 1, &h1)

	seedFullBatting(t, db, psA, true, 150, 500, 140, 30, 5, 40, 120, 10, 60, 80)
	seedFullBatting(t, db, psB, false, 30, 80, 22, 4, 1, 5, 18, 2, 10, 15) // playoff only

	rsRows, err := newStatRecordStore(db).GetBattingCountRows(ctx, true)
	if err != nil {
		t.Fatalf("GetBattingCountRows RS: %v", err)
	}
	if len(rsRows) != 1 {
		t.Fatalf("want 1 RS row, got %d", len(rsRows))
	}
	if rsRows[0].PlayerID != pA {
		t.Errorf("want playerID %d, got %d", pA, rsRows[0].PlayerID)
	}
	if rsRows[0].HomeRuns != 40 {
		t.Errorf("want HR=40, got %d", rsRows[0].HomeRuns)
	}

	poRows, err := newStatRecordStore(db).GetBattingCountRows(ctx, false)
	if err != nil {
		t.Fatalf("GetBattingCountRows PO: %v", err)
	}
	if len(poRows) != 1 {
		t.Fatalf("want 1 PO row, got %d", len(poRows))
	}
	if poRows[0].PlayerID != pB {
		t.Errorf("want playerID %d (Bob, PO row), got %d", pB, poRows[0].PlayerID)
	}
}

func TestGetBattingCountRows_AllFieldsScanned(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team A", "E", "AL", 20, 20)
	p1 := seedPlayer(t, db, "g1", "One", "Player")
	ps1 := seedPlayerSeason(t, db, p1, 1, &h1)
	seedFullBatting(t, db, ps1, true, 150, 550, 160, 35, 8, 45, 130, 20, 70, 95)

	rows, err := newStatRecordStore(db).GetBattingCountRows(ctx, true)
	if err != nil {
		t.Fatalf("GetBattingCountRows: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	r := rows[0]
	if r.SeasonNum != 1 {
		t.Errorf("SeasonNum: want 1, got %d", r.SeasonNum)
	}
	if r.GamesPlayed != 150 {
		t.Errorf("GamesPlayed: want 150, got %d", r.GamesPlayed)
	}
	if r.Doubles != 35 {
		t.Errorf("Doubles: want 35, got %d", r.Doubles)
	}
	if r.Triples != 8 {
		t.Errorf("Triples: want 8, got %d", r.Triples)
	}
	if r.StolenBases != 20 {
		t.Errorf("StolenBases: want 20, got %d", r.StolenBases)
	}
	if r.Walks != 70 {
		t.Errorf("Walks: want 70, got %d", r.Walks)
	}
	if r.Strikeouts != 95 {
		t.Errorf("Strikeouts: want 95, got %d", r.Strikeouts)
	}
}

func TestGetBattingCountRows_TiedPlayersReturnedAsSeparateRows(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team A", "E", "AL", 20, 20)

	pA := seedPlayer(t, db, "gA", "Alice", "Alpha")
	pB := seedPlayer(t, db, "gB", "Bob", "Beta")

	psA := seedPlayerSeason(t, db, pA, 1, &h1)
	psB := seedPlayerSeason(t, db, pB, 1, &h1)

	// Both hit 35 HR — tied for the league lead
	seedFullBatting(t, db, psA, true, 150, 500, 140, 30, 5, 35, 110, 10, 60, 80)
	seedFullBatting(t, db, psB, true, 155, 520, 148, 28, 3, 35, 105, 8, 55, 75)

	rows, err := newStatRecordStore(db).GetBattingCountRows(ctx, true)
	if err != nil {
		t.Fatalf("GetBattingCountRows: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows (both tied players), got %d", len(rows))
	}
	ids := []int64{rows[0].PlayerID, rows[1].PlayerID}
	slices.Sort(ids)
	if ids[0] != pA || ids[1] != pB {
		t.Errorf("want both playerIDs present, got %v", ids)
	}
}

func TestGetBattingCountRows_MultiSeason(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	seedSeason(t, db, 2, 2, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team A", "E", "AL", 20, 20)
	h2 := seedTeamHistory(t, db, t1, 2, "Team A", "E", "AL", 25, 15)

	p1 := seedPlayer(t, db, "g1", "One", "Player")
	ps1 := seedPlayerSeason(t, db, p1, 1, &h1)
	ps2 := seedPlayerSeason(t, db, p1, 2, &h2)
	seedFullBatting(t, db, ps1, true, 150, 500, 140, 30, 5, 40, 120, 10, 60, 80)
	seedFullBatting(t, db, ps2, true, 152, 510, 145, 32, 4, 38, 115, 12, 62, 85)

	rows, err := newStatRecordStore(db).GetBattingCountRows(ctx, true)
	if err != nil {
		t.Fatalf("GetBattingCountRows: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows (one per season), got %d", len(rows))
	}
	seasonNums := map[int]bool{rows[0].SeasonNum: true, rows[1].SeasonNum: true}
	if !seasonNums[1] || !seasonNums[2] {
		t.Errorf("want season nums 1 and 2, got %v", seasonNums)
	}
}

func TestGetPitchingCountRows_BasicAndRSPOSplit(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team A", "E", "AL", 20, 20)
	p1 := seedPlayer(t, db, "g1", "One", "Pitcher")
	ps1 := seedPlayerSeason(t, db, p1, 1, &h1)

	seedFullPitching(t, db, ps1, true, 32, 32, 18, 8, 0, 720, 210, 65, 195, 72)
	seedFullPitching(t, db, ps1, false, 3, 3, 2, 1, 0, 81, 25, 7, 20, 8) // playoff

	rsRows, err := newStatRecordStore(db).GetPitchingCountRows(ctx, true)
	if err != nil {
		t.Fatalf("GetPitchingCountRows RS: %v", err)
	}
	if len(rsRows) != 1 {
		t.Fatalf("want 1 RS row, got %d", len(rsRows))
	}
	r := rsRows[0]
	if r.Strikeouts != 210 {
		t.Errorf("Strikeouts: want 210, got %d", r.Strikeouts)
	}
	if r.OutsPitched != 720 {
		t.Errorf("OutsPitched: want 720, got %d", r.OutsPitched)
	}
	if r.Saves != 0 {
		t.Errorf("Saves: want 0, got %d", r.Saves)
	}
	if r.HitsAllowed != 195 {
		t.Errorf("HitsAllowed: want 195, got %d", r.HitsAllowed)
	}
	if r.EarnedRuns != 72 {
		t.Errorf("EarnedRuns: want 72, got %d", r.EarnedRuns)
	}

	poRows, err := newStatRecordStore(db).GetPitchingCountRows(ctx, false)
	if err != nil {
		t.Fatalf("GetPitchingCountRows PO: %v", err)
	}
	if len(poRows) != 1 {
		t.Fatalf("want 1 PO row, got %d", len(poRows))
	}
	if poRows[0].Strikeouts != 25 {
		t.Errorf("PO Strikeouts: want 25, got %d", poRows[0].Strikeouts)
	}
}

// ── Rate stat query tests ─────────────────────────────────────────────────────

// seedBattingRate inserts a batting row with rate stats and plate_appearances set.
func seedBattingRate(t *testing.T, db *sql.DB, playerSeasonID int64, isReg bool,
	pa, ab, hits int, ba, obp *float64,
) {
	t.Helper()
	isRegInt := 0
	if isReg {
		isRegInt = 1
	}
	_, err := db.ExecContext(context.Background(), `
INSERT INTO player_season_batting_stats
    (player_season_id, is_regular_season, games_played, games_batting,
     at_bats, runs, hits, doubles, triples, home_runs, rbi,
     stolen_bases, caught_stealing, walks, strikeouts, hit_by_pitch,
     sac_hits, sac_flies, errors, passed_balls, plate_appearances, ba, obp)
VALUES (?,?,?,?,?,0,?,0,0,0,0,0,0,0,0,0,0,0,0,0,?,?,?)
`, playerSeasonID, isRegInt, 1, 1, ab, hits, pa, ba, obp)
	if err != nil {
		t.Fatalf("seedBattingRate: %v", err)
	}
}

// seedPitchingRate inserts a pitching row with rate stats and outs_pitched set.
func seedPitchingRate(t *testing.T, db *sql.DB, playerSeasonID int64, isReg bool,
	outs int, era, whip *float64,
) {
	t.Helper()
	isRegInt := 0
	if isReg {
		isRegInt = 1
	}
	_, err := db.ExecContext(context.Background(), `
INSERT INTO player_season_pitching_stats
    (player_season_id, is_regular_season, wins, losses, games, games_started,
     complete_games, shutouts, saves, outs_pitched, hits_allowed, earned_runs,
     home_runs_allowed, walks, strikeouts, hit_batters, batters_faced,
     games_finished, runs_allowed, wild_pitches, total_pitches, era, whip)
VALUES (?,?,0,0,1,0,0,0,0,?,0,0,0,0,0,0,0,0,0,0,0,?,?)
`, playerSeasonID, isRegInt, outs, era, whip)
	if err != nil {
		t.Fatalf("seedPitchingRate: %v", err)
	}
}

func TestGetBattingRateRows_RSPOSplit(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team A", "E", "AL", 20, 20)
	pA := seedPlayer(t, db, "gA", "Alice", "Alpha")
	pB := seedPlayer(t, db, "gB", "Bob", "Beta")
	psA := seedPlayerSeason(t, db, pA, 1, &h1)
	psB := seedPlayerSeason(t, db, pB, 1, &h1)

	ba := 0.320
	obp := 0.400
	seedBattingRate(t, db, psA, true, 200, 180, 58, &ba, &obp)
	seedBattingRate(t, db, psB, false, 30, 28, 8, nil, nil) // playoff only

	rsRows, err := newStatRecordStore(db).GetBattingRateRows(ctx, true)
	if err != nil {
		t.Fatalf("GetBattingRateRows RS: %v", err)
	}
	if len(rsRows) != 1 {
		t.Fatalf("want 1 RS row, got %d", len(rsRows))
	}
	r := rsRows[0]
	if r.PlayerID != pA {
		t.Errorf("want playerID %d, got %d", pA, r.PlayerID)
	}
	if r.BA == nil || *r.BA != ba {
		t.Errorf("BA: want %v, got %v", ba, r.BA)
	}
	if r.PlateAppearances != 200 {
		t.Errorf("PlateAppearances: want 200, got %d", r.PlateAppearances)
	}
	if r.NumGames != 40 {
		t.Errorf("NumGames: want 40, got %d", r.NumGames)
	}

	poRows, err := newStatRecordStore(db).GetBattingRateRows(ctx, false)
	if err != nil {
		t.Fatalf("GetBattingRateRows PO: %v", err)
	}
	if len(poRows) != 1 || poRows[0].PlayerID != pB {
		t.Fatalf("want 1 PO row for player B, got %v", poRows)
	}
}

func TestGetPitchingRateRows_OutsPitchedAndNumGamesPresent(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team A", "E", "AL", 20, 20)
	p1 := seedPlayer(t, db, "g1", "One", "Pitcher")
	ps1 := seedPlayerSeason(t, db, p1, 1, &h1)

	era := 3.00
	whip := 1.20
	seedPitchingRate(t, db, ps1, true, 270, &era, &whip)

	rows, err := newStatRecordStore(db).GetPitchingRateRows(ctx, true)
	if err != nil {
		t.Fatalf("GetPitchingRateRows: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	r := rows[0]
	if r.OutsPitched != 270 {
		t.Errorf("OutsPitched: want 270, got %d", r.OutsPitched)
	}
	if r.NumGames != 40 {
		t.Errorf("NumGames: want 40, got %d", r.NumGames)
	}
	if r.ERA == nil || *r.ERA != era {
		t.Errorf("ERA: want %v, got %v", era, r.ERA)
	}
}

func TestGetFranchiseSeasonLength_ReturnsFirstSeasonNumGames(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	seedSeason(t, db, 2, 2, 50) // second season with different length

	got, err := newStatRecordStore(db).GetFranchiseSeasonLength(ctx)
	if err != nil {
		t.Fatalf("GetFranchiseSeasonLength: %v", err)
	}
	if got != 40 {
		t.Errorf("want 40 (first season), got %d", got)
	}
}

func TestGetFranchiseSeasonLength_NoSeasons(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	got, err := newStatRecordStore(db).GetFranchiseSeasonLength(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 0 {
		t.Errorf("want 0 when no seasons, got %d", got)
	}
}

func TestGetBattingRateRows_OPSPlusScanned(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 100)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team A", "E", "AL", 50, 50)
	p1 := seedPlayer(t, db, "g1", "One", "Player")
	ps1 := seedPlayerSeason(t, db, p1, 1, &h1)

	opsPlus := 145.0
	_, err := db.ExecContext(ctx, `
INSERT INTO player_season_batting_stats
    (player_season_id, is_regular_season, games_played, games_batting,
     at_bats, runs, hits, doubles, triples, home_runs, rbi,
     stolen_bases, caught_stealing, walks, strikeouts, hit_by_pitch,
     sac_hits, sac_flies, errors, passed_balls, plate_appearances, ops_plus)
VALUES (?,1,100,100,400,0,130,0,0,0,0,0,0,0,0,0,0,0,0,0,430,?)
`, ps1, opsPlus)
	if err != nil {
		t.Fatalf("seed batting with ops_plus: %v", err)
	}

	rows, err := newStatRecordStore(db).GetBattingRateRows(ctx, true)
	if err != nil {
		t.Fatalf("GetBattingRateRows: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	if rows[0].OPSPlus == nil || *rows[0].OPSPlus != opsPlus {
		t.Errorf("OPSPlus: want %v, got %v", opsPlus, rows[0].OPSPlus)
	}
	if rows[0].PlateAppearances != 430 {
		t.Errorf("PlateAppearances: want 430, got %d", rows[0].PlateAppearances)
	}
}

func TestGetPitchingRateRows_ERAAndFIPMinusScanned(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 100)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team A", "E", "AL", 50, 50)
	p1 := seedPlayer(t, db, "g1", "One", "Pitcher")
	ps1 := seedPlayerSeason(t, db, p1, 1, &h1)

	eraPlus := 142.0
	fipMinus := 78.0
	_, err := db.ExecContext(ctx, `
INSERT INTO player_season_pitching_stats
    (player_season_id, is_regular_season, wins, losses, games, games_started,
     complete_games, shutouts, saves, outs_pitched, hits_allowed, earned_runs,
     home_runs_allowed, walks, strikeouts, hit_batters, batters_faced,
     games_finished, runs_allowed, wild_pitches, total_pitches, era_plus, fip_minus)
VALUES (?,1,15,5,30,30,0,0,0,720,180,60,0,50,200,0,0,0,0,0,0,?,?)
`, ps1, eraPlus, fipMinus)
	if err != nil {
		t.Fatalf("seed pitching with era_plus/fip_minus: %v", err)
	}

	rows, err := newStatRecordStore(db).GetPitchingRateRows(ctx, true)
	if err != nil {
		t.Fatalf("GetPitchingRateRows: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	r := rows[0]
	if r.ERAPlus == nil || *r.ERAPlus != eraPlus {
		t.Errorf("ERAPlus: want %v, got %v", eraPlus, r.ERAPlus)
	}
	if r.FIPMinus == nil || *r.FIPMinus != fipMinus {
		t.Errorf("FIPMinus: want %v, got %v", fipMinus, r.FIPMinus)
	}
	if r.OutsPitched != 720 {
		t.Errorf("OutsPitched: want 720, got %d", r.OutsPitched)
	}
}
