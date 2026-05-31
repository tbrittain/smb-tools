package store_test

import (
	"context"
	"database/sql"
	"testing"

	"smb-tools/internal/models"
	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

// seedPlayerSeasonFull inserts a player_seasons row with all metadata specified
// by the caller. Use seedPlayerSeason for cases where defaults (SS/R/R) suffice.
func seedPlayerSeasonFull(
	t *testing.T, db *sql.DB,
	playerID int64, seasonID int, teamHistID *int64,
	primaryPos, pitcherRole, batHand, throwHand, chemistry string,
) int64 {
	t.Helper()
	res, err := db.ExecContext(context.Background(), `
INSERT INTO player_seasons
    (player_id, season_id, age, salary,
     primary_position, secondary_position, pitcher_role,
     bat_hand, throw_hand, chemistry_type, traits_json, pitches_json)
VALUES (?,?,25,1000,?,?,?,?,?,?,'[]','[]')
`, playerID, seasonID, primaryPos, "", pitcherRole, batHand, throwHand, chemistry)
	if err != nil {
		t.Fatalf("seedPlayerSeasonFull: %v", err)
	}
	id, _ := res.LastInsertId()
	if teamHistID != nil {
		_, err = db.ExecContext(context.Background(), `
INSERT OR IGNORE INTO player_season_teams (player_season_id, team_history_id, sort_order)
VALUES (?, ?, 0)
`, id, *teamHistID)
		if err != nil {
			t.Fatalf("seedPlayerSeasonFull team: %v", err)
		}
	}
	return id
}

// ── shared setup helpers for leaderboard tests ────────────────────────────────

func newLB(db *sql.DB) *store.LeaderboardQueryStore {
	return store.NewLeaderboardQueryStore(db)
}

func noFilters() models.LeaderboardFilters { return models.LeaderboardFilters{} }

// ── Batting career leaderboard ────────────────────────────────────────────────

func TestGetBattingCareerLeaders_BasicAggregation(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	seedSeason(t, db, 2, 2, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)
	h2 := seedTeamHistory(t, db, t1, 2, "Team", "E", "AL", 25, 15)

	pA := seedPlayer(t, db, "gA", "Alice", "Alpha")
	pB := seedPlayer(t, db, "gB", "Bob", "Beta")

	// pA: two seasons, 20+10 HR
	psA1 := seedPlayerSeason(t, db, pA, 1, &h1)
	psA2 := seedPlayerSeason(t, db, pA, 2, &h2)
	seedBatting(t, db, psA1, true, 400, 120, 20, 80)
	seedBatting(t, db, psA2, true, 500, 130, 10, 60)

	// pB: one season, 15 HR
	psB1 := seedPlayerSeason(t, db, pB, 1, &h1)
	seedBatting(t, db, psB1, true, 300, 90, 15, 50)

	rows, err := newLB(db).GetBattingCareerLeaders(ctx, noFilters())
	if err != nil {
		t.Fatalf("GetBattingCareerLeaders: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}

	// Results ordered by last_name: Alpha before Beta
	rA := rows[0]
	if rA.PlayerID != pA {
		t.Errorf("first row: want playerID %d, got %d", pA, rA.PlayerID)
	}
	if rA.SeasonsPlayed != 2 {
		t.Errorf("Alice SeasonsPlayed: want 2, got %d", rA.SeasonsPlayed)
	}
	if rA.HomeRuns != 30 {
		t.Errorf("Alice HR: want 30, got %d", rA.HomeRuns)
	}
	if rA.AtBats != 900 {
		t.Errorf("Alice AB: want 900, got %d", rA.AtBats)
	}
	// Rate fields are nil — caller computes them
	if rA.BA != nil {
		t.Error("BA should be nil before ComputeBattingRates")
	}

	rB := rows[1]
	if rB.HomeRuns != 15 {
		t.Errorf("Bob HR: want 15, got %d", rB.HomeRuns)
	}
	if rB.SeasonsPlayed != 1 {
		t.Errorf("Bob SeasonsPlayed: want 1, got %d", rB.SeasonsPlayed)
	}
}

func TestGetBattingCareerLeaders_HoFFilter(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	pHoF := seedPlayer(t, db, "gH", "Hall", "Famer")
	_, _ = db.ExecContext(ctx, `UPDATE players SET is_hall_of_famer = 1 WHERE id = ?`, pHoF)

	pNorm := seedPlayer(t, db, "gN", "Normal", "Player")

	ps1 := seedPlayerSeason(t, db, pHoF, 1, &h1)
	ps2 := seedPlayerSeason(t, db, pNorm, 1, &h1)
	seedBatting(t, db, ps1, true, 400, 120, 20, 80)
	seedBatting(t, db, ps2, true, 300, 90, 10, 50)

	rows, err := newLB(db).GetBattingCareerLeaders(ctx, models.LeaderboardFilters{OnlyHallOfFamers: true})
	if err != nil {
		t.Fatalf("GetBattingCareerLeaders: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 HoF row, got %d", len(rows))
	}
	if rows[0].PlayerID != pHoF {
		t.Errorf("want HoF player, got playerID %d", rows[0].PlayerID)
	}
}

func TestGetBattingCareerLeaders_PositionFilter(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	pSS := seedPlayer(t, db, "gSS", "Short", "Stop")
	p1B := seedPlayer(t, db, "g1B", "First", "Base")

	psSS := seedPlayerSeasonFull(t, db, pSS, 1, &h1, "SS", "", "R", "R", "")
	ps1B := seedPlayerSeasonFull(t, db, p1B, 1, &h1, "1B", "", "L", "R", "")
	seedBatting(t, db, psSS, true, 400, 120, 10, 50)
	seedBatting(t, db, ps1B, true, 350, 100, 20, 80)

	rows, err := newLB(db).GetBattingCareerLeaders(ctx, models.LeaderboardFilters{Position: "SS"})
	if err != nil {
		t.Fatalf("GetBattingCareerLeaders: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 SS row, got %d", len(rows))
	}
	if rows[0].PlayerID != pSS {
		t.Errorf("want SS player, got playerID %d", rows[0].PlayerID)
	}
}

func TestGetBattingCareerLeaders_SeasonRange(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	seedSeason(t, db, 2, 2, 40)
	seedSeason(t, db, 3, 3, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)
	h2 := seedTeamHistory(t, db, t1, 2, "Team", "E", "AL", 20, 20)
	h3 := seedTeamHistory(t, db, t1, 3, "Team", "E", "AL", 20, 20)

	pid := seedPlayer(t, db, "gP", "Range", "Player")
	ps1 := seedPlayerSeason(t, db, pid, 1, &h1)
	ps2 := seedPlayerSeason(t, db, pid, 2, &h2)
	ps3 := seedPlayerSeason(t, db, pid, 3, &h3)
	seedBatting(t, db, ps1, true, 400, 100, 10, 40) // season 1: 10 HR
	seedBatting(t, db, ps2, true, 400, 110, 15, 50) // season 2: 15 HR
	seedBatting(t, db, ps3, true, 400, 120, 20, 60) // season 3: 20 HR

	// Filter to seasons 2–3 only: expect 35 HR
	rows, err := newLB(db).GetBattingCareerLeaders(ctx, models.LeaderboardFilters{SeasonStart: 2, SeasonEnd: 3})
	if err != nil {
		t.Fatalf("GetBattingCareerLeaders: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	if rows[0].HomeRuns != 35 {
		t.Errorf("HR with range filter: want 35, got %d", rows[0].HomeRuns)
	}
	if rows[0].SeasonsPlayed != 2 {
		t.Errorf("SeasonsPlayed with range filter: want 2, got %d", rows[0].SeasonsPlayed)
	}
}

func TestGetBattingCareerLeaders_PlayoffToggle(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)
	pid := seedPlayer(t, db, "gP", "Playoff", "Test")
	ps1 := seedPlayerSeason(t, db, pid, 1, &h1)
	seedBatting(t, db, ps1, true, 400, 100, 10, 40)  // regular season
	seedBatting(t, db, ps1, false, 50, 15, 3, 10)    // playoffs

	// Regular season: 10 HR
	regRows, err := newLB(db).GetBattingCareerLeaders(ctx, noFilters())
	if err != nil {
		t.Fatalf("regular: %v", err)
	}
	if len(regRows) != 1 || regRows[0].HomeRuns != 10 {
		t.Errorf("regular HR: want 10, got %v", regRows)
	}

	// Playoffs: 3 HR
	playoffRows, err := newLB(db).GetBattingCareerLeaders(ctx, models.LeaderboardFilters{GameType: "playoffs"})
	if err != nil {
		t.Fatalf("playoffs: %v", err)
	}
	if len(playoffRows) != 1 || playoffRows[0].HomeRuns != 3 {
		t.Errorf("playoff HR: want 3, got %v", playoffRows)
	}
}

func TestGetBattingCareerLeaders_CombinedGameType(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	seedSeason(t, db, 2, 2, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)
	h2 := seedTeamHistory(t, db, t1, 2, "Team", "E", "AL", 20, 20)

	pid := seedPlayer(t, db, "gP", "Combined", "Test")
	ps1 := seedPlayerSeason(t, db, pid, 1, &h1)
	ps2 := seedPlayerSeason(t, db, pid, 2, &h2)
	// Season 1: 10 regular, 3 playoff HR
	seedBatting(t, db, ps1, true, 400, 100, 10, 40)
	seedBatting(t, db, ps1, false, 50, 15, 3, 10)
	// Season 2: 15 regular HR only
	seedBatting(t, db, ps2, true, 420, 110, 15, 60)

	rows, err := newLB(db).GetBattingCareerLeaders(ctx, models.LeaderboardFilters{GameType: "combined"})
	if err != nil {
		t.Fatalf("combined: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	r := rows[0]
	if r.HomeRuns != 28 {
		t.Errorf("combined HR: want 28 (10+3+15), got %d", r.HomeRuns)
	}
	// SeasonsPlayed counts distinct game seasons, not stat rows
	if r.SeasonsPlayed != 2 {
		t.Errorf("combined SeasonsPlayed: want 2, got %d", r.SeasonsPlayed)
	}
	// OPS+ is NULL for combined (league context ambiguous)
	if r.OPSPlus != nil {
		t.Errorf("combined OPS+: want nil, got %v", r.OPSPlus)
	}
}

func TestGetBattingCareerLeaders_ExcludesZeroAB(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	// Pitcher with 0 AB — should be excluded from batting leaderboard
	pPit := seedPlayer(t, db, "gPit", "Pure", "Pitcher")
	psBat := seedPlayerSeason(t, db, pPit, 1, &h1)
	seedPitching(t, db, psBat, true, 10, 5, 270, 30, 100)
	// No batting stats row inserted → zero AB

	// Batter with actual AB — should appear
	pBat := seedPlayer(t, db, "gBat", "Real", "Batter")
	psBat2 := seedPlayerSeason(t, db, pBat, 1, &h1)
	seedBatting(t, db, psBat2, true, 400, 100, 10, 40)

	rows, err := newLB(db).GetBattingCareerLeaders(ctx, noFilters())
	if err != nil {
		t.Fatalf("GetBattingCareerLeaders: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row (batter only), got %d", len(rows))
	}
	if rows[0].PlayerID != pBat {
		t.Errorf("want batter, got playerID %d", rows[0].PlayerID)
	}
}

// ── Batting season leaderboard ────────────────────────────────────────────────

func TestGetBattingSeasonLeaders_OneRowPerPlayerSeason(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	seedSeason(t, db, 2, 2, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)
	h2 := seedTeamHistory(t, db, t1, 2, "Team", "E", "AL", 25, 15)

	pA := seedPlayer(t, db, "gA", "Alice", "Alpha")
	pB := seedPlayer(t, db, "gB", "Bob", "Beta")

	psA1 := seedPlayerSeason(t, db, pA, 1, &h1)
	psA2 := seedPlayerSeason(t, db, pA, 2, &h2)
	psB1 := seedPlayerSeason(t, db, pB, 1, &h1)
	psB2 := seedPlayerSeason(t, db, pB, 2, &h2)

	seedBatting(t, db, psA1, true, 400, 100, 10, 40)
	seedBatting(t, db, psA2, true, 420, 110, 15, 50)
	seedBatting(t, db, psB1, true, 380, 95, 8, 35)
	seedBatting(t, db, psB2, true, 350, 90, 12, 45)

	rows, _, err := newLB(db).GetBattingSeasonLeaders(ctx, noFilters())
	if err != nil {
		t.Fatalf("GetBattingSeasonLeaders: %v", err)
	}
	if len(rows) != 4 {
		t.Fatalf("want 4 rows (2 players × 2 seasons), got %d", len(rows))
	}
	// Check SeasonNum is populated
	var s1Count, s2Count int
	for _, r := range rows {
		switch r.SeasonNum {
		case 1:
			s1Count++
		case 2:
			s2Count++
		}
	}
	if s1Count != 2 || s2Count != 2 {
		t.Errorf("want 2 rows per season, got s1=%d s2=%d", s1Count, s2Count)
	}
}

func TestGetBattingSeasonLeaders_BatHandFilter(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	pL := seedPlayer(t, db, "gL", "Lefty", "Batter")
	pR := seedPlayer(t, db, "gR", "Righty", "Batter")

	psL := seedPlayerSeasonFull(t, db, pL, 1, &h1, "LF", "", "L", "L", "")
	psR := seedPlayerSeasonFull(t, db, pR, 1, &h1, "RF", "", "R", "R", "")
	seedBatting(t, db, psL, true, 400, 100, 10, 40)
	seedBatting(t, db, psR, true, 380, 95, 8, 35)

	rows, _, err := newLB(db).GetBattingSeasonLeaders(ctx, models.LeaderboardFilters{BatHand: "L"})
	if err != nil {
		t.Fatalf("GetBattingSeasonLeaders: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 lefty row, got %d", len(rows))
	}
	if rows[0].PlayerID != pL {
		t.Errorf("want lefty player, got playerID %d", rows[0].PlayerID)
	}
}

func TestGetBattingSeasonLeaders_ExcludesZeroAB(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	pid := seedPlayer(t, db, "gP", "Pure", "Pitcher")
	ps := seedPlayerSeason(t, db, pid, 1, &h1)
	seedPitching(t, db, ps, true, 12, 8, 360, 40, 150)
	// No batting row — pitcher appears in pitching leaderboard, not batting

	rows, _, err := newLB(db).GetBattingSeasonLeaders(ctx, noFilters())
	if err != nil {
		t.Fatalf("GetBattingSeasonLeaders: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("want 0 rows for pitcher with no AB, got %d", len(rows))
	}
}

// ── Pitching career leaderboard ───────────────────────────────────────────────

func TestGetPitchingCareerLeaders_BasicAggregation(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	seedSeason(t, db, 2, 2, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)
	h2 := seedTeamHistory(t, db, t1, 2, "Team", "E", "AL", 25, 15)

	pA := seedPlayer(t, db, "gA", "Ace", "Arm")
	psA1 := seedPlayerSeason(t, db, pA, 1, &h1)
	psA2 := seedPlayerSeason(t, db, pA, 2, &h2)
	// season 1: 10W, 5L, 270 outs, 30 ER, 100 K
	// season 2: 12W, 6L, 300 outs, 35 ER, 120 K
	seedPitching(t, db, psA1, true, 10, 5, 270, 30, 100)
	seedPitching(t, db, psA2, true, 12, 6, 300, 35, 120)

	rows, err := newLB(db).GetPitchingCareerLeaders(ctx, noFilters())
	if err != nil {
		t.Fatalf("GetPitchingCareerLeaders: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	r := rows[0]
	if r.SeasonsPlayed != 2 {
		t.Errorf("SeasonsPlayed: want 2, got %d", r.SeasonsPlayed)
	}
	if r.Wins != 22 {
		t.Errorf("Wins: want 22, got %d", r.Wins)
	}
	if r.OutsPitched != 570 {
		t.Errorf("OutsPitched: want 570, got %d", r.OutsPitched)
	}
	if r.Strikeouts != 220 {
		t.Errorf("Strikeouts: want 220, got %d", r.Strikeouts)
	}
	if r.ERA != nil {
		t.Error("ERA should be nil before ComputePitchingRates")
	}
}

func TestGetPitchingCareerLeaders_PitcherRoleFilter(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	pSP := seedPlayer(t, db, "gSP", "Start", "Er")
	pRP := seedPlayer(t, db, "gRP", "Relief", "Pitcher")

	psSP := seedPlayerSeasonFull(t, db, pSP, 1, &h1, "P", "SP", "R", "R", "")
	psRP := seedPlayerSeasonFull(t, db, pRP, 1, &h1, "P", "RP", "R", "R", "")
	seedPitching(t, db, psSP, true, 12, 8, 540, 60, 180)
	seedPitching(t, db, psRP, true, 5, 3, 120, 15, 60)

	rows, err := newLB(db).GetPitchingCareerLeaders(ctx, models.LeaderboardFilters{Position: "SP"})
	if err != nil {
		t.Fatalf("GetPitchingCareerLeaders: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 SP row, got %d", len(rows))
	}
	if rows[0].PlayerID != pSP {
		t.Errorf("want SP player, got playerID %d", rows[0].PlayerID)
	}
}

func TestGetPitchingCareerLeaders_ThrowHandFilter(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	pL := seedPlayer(t, db, "gL", "Lefty", "Pitcher")
	pR := seedPlayer(t, db, "gR", "Righty", "Pitcher")

	psL := seedPlayerSeasonFull(t, db, pL, 1, &h1, "P", "SP", "R", "L", "")
	psR := seedPlayerSeasonFull(t, db, pR, 1, &h1, "P", "SP", "R", "R", "")
	seedPitching(t, db, psL, true, 10, 8, 360, 40, 130)
	seedPitching(t, db, psR, true, 12, 6, 400, 45, 150)

	rows, err := newLB(db).GetPitchingCareerLeaders(ctx, models.LeaderboardFilters{ThrowHand: "L"})
	if err != nil {
		t.Fatalf("GetPitchingCareerLeaders: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 lefty pitcher, got %d", len(rows))
	}
	if rows[0].PlayerID != pL {
		t.Errorf("want lefty, got playerID %d", rows[0].PlayerID)
	}
}

func TestGetPitchingCareerLeaders_CombinedGameType(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	seedSeason(t, db, 2, 2, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)
	h2 := seedTeamHistory(t, db, t1, 2, "Team", "E", "AL", 20, 20)

	pid := seedPlayer(t, db, "gP", "Combined", "Pitcher")
	ps1 := seedPlayerSeason(t, db, pid, 1, &h1)
	ps2 := seedPlayerSeason(t, db, pid, 2, &h2)
	// Season 1: 10W regular, 2W playoff
	seedPitching(t, db, ps1, true, 10, 5, 270, 30, 100)
	seedPitching(t, db, ps1, false, 2, 1, 60, 8, 25)
	// Season 2: 12W regular only
	seedPitching(t, db, ps2, true, 12, 6, 300, 35, 120)

	rows, err := newLB(db).GetPitchingCareerLeaders(ctx, models.LeaderboardFilters{GameType: "combined"})
	if err != nil {
		t.Fatalf("combined: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	r := rows[0]
	if r.Wins != 24 {
		t.Errorf("combined Wins: want 24 (10+2+12), got %d", r.Wins)
	}
	if r.SeasonsPlayed != 2 {
		t.Errorf("combined SeasonsPlayed: want 2, got %d", r.SeasonsPlayed)
	}
	// ERA+ is NULL for combined
	if r.ERAPlus != nil {
		t.Errorf("combined ERA+: want nil, got %v", r.ERAPlus)
	}
}

// ── Pitching season leaderboard ───────────────────────────────────────────────

func TestGetPitchingSeasonLeaders_OneRowPerPlayerSeason(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	seedSeason(t, db, 2, 2, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)
	h2 := seedTeamHistory(t, db, t1, 2, "Team", "E", "AL", 25, 15)

	pA := seedPlayer(t, db, "gA", "Ace", "Arm")
	pB := seedPlayer(t, db, "gB", "Buddy", "Bullpen")

	psA1 := seedPlayerSeason(t, db, pA, 1, &h1)
	psA2 := seedPlayerSeason(t, db, pA, 2, &h2)
	psB1 := seedPlayerSeason(t, db, pB, 1, &h1)
	psB2 := seedPlayerSeason(t, db, pB, 2, &h2)

	seedPitching(t, db, psA1, true, 10, 5, 270, 30, 100)
	seedPitching(t, db, psA2, true, 12, 6, 300, 35, 120)
	seedPitching(t, db, psB1, true, 4, 3, 90, 12, 45)
	seedPitching(t, db, psB2, true, 5, 4, 100, 14, 50)

	rows, _, err := newLB(db).GetPitchingSeasonLeaders(ctx, noFilters())
	if err != nil {
		t.Fatalf("GetPitchingSeasonLeaders: %v", err)
	}
	if len(rows) != 4 {
		t.Fatalf("want 4 rows (2 pitchers × 2 seasons), got %d", len(rows))
	}
}

func TestGetPitchingSeasonLeaders_ExcludesZeroOuts(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	pid := seedPlayer(t, db, "gP", "Position", "Player")
	ps := seedPlayerSeason(t, db, pid, 1, &h1)
	seedBatting(t, db, ps, true, 400, 100, 10, 40)
	// No pitching row — position player should not appear in pitching leaderboard

	rows, _, err := newLB(db).GetPitchingSeasonLeaders(ctx, noFilters())
	if err != nil {
		t.Fatalf("GetPitchingSeasonLeaders: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("want 0 rows for position player with no outs, got %d", len(rows))
	}
}

// ── Trait filter tests ────────────────────────────────────────────────────────

// setTraits updates the traits_json column for an existing player_seasons row.
func setTraits(t *testing.T, db *sql.DB, playerSeasonID int64, traits []string) {
	t.Helper()
	b := "[]"
	if len(traits) > 0 {
		parts := make([]byte, 0, 64)
		parts = append(parts, '[')
		for i, tr := range traits {
			if i > 0 {
				parts = append(parts, ',')
			}
			parts = append(parts, '"')
			parts = append(parts, []byte(tr)...)
			parts = append(parts, '"')
		}
		parts = append(parts, ']')
		b = string(parts)
	}
	_, err := db.ExecContext(context.Background(),
		`UPDATE player_seasons SET traits_json = ? WHERE id = ?`, b, playerSeasonID)
	if err != nil {
		t.Fatalf("setTraits: %v", err)
	}
}

func TestGetBattingSeasonLeaders_SingleTraitFilter(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	pClutch := seedPlayer(t, db, "gC", "Clutch", "Player")
	pNone := seedPlayer(t, db, "gN", "No", "Trait")

	psC := seedPlayerSeason(t, db, pClutch, 1, &h1)
	psN := seedPlayerSeason(t, db, pNone, 1, &h1)
	setTraits(t, db, psC, []string{"Clutch"})
	seedBatting(t, db, psC, true, 400, 120, 20, 80)
	seedBatting(t, db, psN, true, 380, 100, 10, 50)

	rows, total, err := newLB(db).GetBattingSeasonLeaders(ctx, models.LeaderboardFilters{Traits: []string{"Clutch"}})
	if err != nil {
		t.Fatalf("GetBattingSeasonLeaders: %v", err)
	}
	if total != 1 {
		t.Fatalf("want total=1, got %d", total)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	if rows[0].PlayerID != pClutch {
		t.Errorf("want Clutch player, got playerID %d", rows[0].PlayerID)
	}
}

func TestGetBattingSeasonLeaders_TwoTraitANDFilter(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	// pBoth has Clutch+Durable; pOne has only Clutch; pNone has no traits
	pBoth := seedPlayer(t, db, "gB", "Both", "Traits")
	pOne := seedPlayer(t, db, "gO", "One", "Trait")
	pNone := seedPlayer(t, db, "gN", "No", "Trait")

	psB := seedPlayerSeason(t, db, pBoth, 1, &h1)
	psO := seedPlayerSeason(t, db, pOne, 1, &h1)
	psN := seedPlayerSeason(t, db, pNone, 1, &h1)
	setTraits(t, db, psB, []string{"Clutch", "Durable"})
	setTraits(t, db, psO, []string{"Clutch"})
	seedBatting(t, db, psB, true, 400, 120, 20, 80)
	seedBatting(t, db, psO, true, 380, 110, 15, 60)
	seedBatting(t, db, psN, true, 350, 90, 5, 30)

	rows, total, err := newLB(db).GetBattingSeasonLeaders(ctx,
		models.LeaderboardFilters{Traits: []string{"Clutch", "Durable"}})
	if err != nil {
		t.Fatalf("GetBattingSeasonLeaders: %v", err)
	}
	if total != 1 {
		t.Fatalf("want total=1 (AND logic), got %d", total)
	}
	if rows[0].PlayerID != pBoth {
		t.Errorf("want pBoth, got playerID %d", rows[0].PlayerID)
	}
}

func TestGetBattingSeasonLeaders_TraitsPopulatedOnRow(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	pid := seedPlayer(t, db, "gP", "Trait", "Player")
	ps := seedPlayerSeason(t, db, pid, 1, &h1)
	setTraits(t, db, ps, []string{"Clutch", "Durable"})
	seedBatting(t, db, ps, true, 400, 120, 20, 80)

	rows, _, err := newLB(db).GetBattingSeasonLeaders(ctx, noFilters())
	if err != nil {
		t.Fatalf("GetBattingSeasonLeaders: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	got := rows[0].Traits
	if len(got) != 2 || got[0] != "Clutch" || got[1] != "Durable" {
		t.Errorf("Traits: want [Clutch Durable], got %v", got)
	}
}

func TestGetPitchingSeasonLeaders_SingleTraitFilter(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	pWH := seedPlayer(t, db, "gW", "Workhorse", "Pitcher")
	pNone := seedPlayer(t, db, "gN", "Plain", "Pitcher")

	psW := seedPlayerSeason(t, db, pWH, 1, &h1)
	psN := seedPlayerSeason(t, db, pNone, 1, &h1)
	setTraits(t, db, psW, []string{"Workhorse"})
	seedPitching(t, db, psW, true, 15, 5, 540, 50, 200)
	seedPitching(t, db, psN, true, 10, 8, 360, 40, 120)

	rows, total, err := newLB(db).GetPitchingSeasonLeaders(ctx,
		models.LeaderboardFilters{Traits: []string{"Workhorse"}})
	if err != nil {
		t.Fatalf("GetPitchingSeasonLeaders: %v", err)
	}
	if total != 1 {
		t.Fatalf("want total=1, got %d", total)
	}
	if rows[0].PlayerID != pWH {
		t.Errorf("want Workhorse pitcher, got playerID %d", rows[0].PlayerID)
	}
}

func TestGetPitchingSeasonLeaders_TraitsPopulatedOnRow(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	pid := seedPlayer(t, db, "gP", "Trait", "Pitcher")
	ps := seedPlayerSeason(t, db, pid, 1, &h1)
	setTraits(t, db, ps, []string{"Workhorse"})
	seedPitching(t, db, ps, true, 12, 6, 450, 45, 150)

	rows, _, err := newLB(db).GetPitchingSeasonLeaders(ctx, noFilters())
	if err != nil {
		t.Fatalf("GetPitchingSeasonLeaders: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	if len(rows[0].Traits) != 1 || rows[0].Traits[0] != "Workhorse" {
		t.Errorf("Traits: want [Workhorse], got %v", rows[0].Traits)
	}
}

// ── QualifiedOnly filter ──────────────────────────────────────────────────────

// TestGetBattingSeasonLeaders_QualifiedOnly verifies that QualifiedOnly=true
// filters to batters with plate_appearances >= num_games * 3.1 and that
// QualifiedOnly=false returns both qualified and unqualified rows.
func TestGetBattingSeasonLeaders_QualifiedOnly(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	// 100-game season: qualification threshold = 100 * 3.1 = 310 PA.
	seedSeason(t, db, 1, 1, 100)
	t1 := seedTeam(t, db, "tgQ")
	h1 := seedTeamHistory(t, db, t1, 1, "Team Q", "E", "AL", 50, 50)

	pQ := seedPlayer(t, db, "gQQ", "Qual", "Batter")    // 350 AB → 350 PA (qualifies)
	pU := seedPlayer(t, db, "gUU", "Unqual", "Batter")  // 200 AB → 200 PA (does not qualify)

	psQ := seedPlayerSeason(t, db, pQ, 1, &h1)
	psU := seedPlayerSeason(t, db, pU, 1, &h1)
	seedBatting(t, db, psQ, true, 350, 105, 20, 80) // .300 BA
	seedBatting(t, db, psU, true, 200, 60, 5, 25)   // .300 BA, below threshold

	lb := newLB(db)

	// QualifiedOnly=true: only the qualified batter appears.
	rows, total, err := lb.GetBattingSeasonLeaders(ctx, models.LeaderboardFilters{QualifiedOnly: true})
	if err != nil {
		t.Fatalf("QualifiedOnly=true: %v", err)
	}
	if total != 1 {
		t.Errorf("QualifiedOnly=true: want total=1, got %d", total)
	}
	if len(rows) != 1 || rows[0].PlayerID != pQ {
		t.Errorf("QualifiedOnly=true: want playerID %d, got %v", pQ, rows)
	}

	// QualifiedOnly=false: both batters appear.
	rows, total, err = lb.GetBattingSeasonLeaders(ctx, models.LeaderboardFilters{QualifiedOnly: false})
	if err != nil {
		t.Fatalf("QualifiedOnly=false: %v", err)
	}
	if total != 2 {
		t.Errorf("QualifiedOnly=false: want total=2, got %d", total)
	}
	_ = rows
}

// TestGetPitchingSeasonLeaders_QualifiedOnly verifies that QualifiedOnly=true
// filters to pitchers with outs_pitched >= num_games * 3 and that
// QualifiedOnly=false returns both qualified and unqualified rows.
func TestGetPitchingSeasonLeaders_QualifiedOnly(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	// 100-game season: qualification threshold = 100 * 3 = 300 outs (100 IP).
	seedSeason(t, db, 1, 1, 100)
	t1 := seedTeam(t, db, "tgR")
	h1 := seedTeamHistory(t, db, t1, 1, "Team R", "W", "NL", 50, 50)

	pQ := seedPlayer(t, db, "gQP", "Qual", "Pitcher")   // 350 outs (qualifies)
	pU := seedPlayer(t, db, "gUP", "Unqual", "Pitcher") // 100 outs (does not qualify)

	psQ := seedPlayerSeasonFull(t, db, pQ, 1, &h1, "P", "SP", "R", "R", "")
	psU := seedPlayerSeasonFull(t, db, pU, 1, &h1, "P", "RP", "R", "R", "")
	seedPitching(t, db, psQ, true, 15, 5, 350, 35, 200) // 2.70 ERA
	seedPitching(t, db, psU, true, 5, 3, 100, 20, 80)   // 5.40 ERA, below threshold

	lb := newLB(db)

	// QualifiedOnly=true: only the qualified pitcher appears.
	rows, total, err := lb.GetPitchingSeasonLeaders(ctx, models.LeaderboardFilters{QualifiedOnly: true})
	if err != nil {
		t.Fatalf("QualifiedOnly=true: %v", err)
	}
	if total != 1 {
		t.Errorf("QualifiedOnly=true: want total=1, got %d", total)
	}
	if len(rows) != 1 || rows[0].PlayerID != pQ {
		t.Errorf("QualifiedOnly=true: want playerID %d, got %v", pQ, rows)
	}

	// QualifiedOnly=false: both pitchers appear.
	rows, total, err = lb.GetPitchingSeasonLeaders(ctx, models.LeaderboardFilters{QualifiedOnly: false})
	if err != nil {
		t.Fatalf("QualifiedOnly=false: %v", err)
	}
	if total != 2 {
		t.Errorf("QualifiedOnly=false: want total=2, got %d", total)
	}
	_ = rows
}
