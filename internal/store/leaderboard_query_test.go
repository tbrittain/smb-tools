package store_test

import (
	"context"
	"database/sql"
	"fmt"
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

	// pA: two seasons, 20+10 HR; AB=400+500=900, H=120+130=250
	psA1 := seedPlayerSeason(t, db, pA, 1, &h1)
	psA2 := seedPlayerSeason(t, db, pA, 2, &h2)
	seedBatting(t, db, psA1, true, 400, 120, 20, 80)
	seedBatting(t, db, psA2, true, 500, 130, 10, 60)

	// pB: one season, 15 HR
	psB1 := seedPlayerSeason(t, db, pB, 1, &h1)
	seedBatting(t, db, psB1, true, 300, 90, 15, 50)

	rows, total, err := newLB(db).GetBattingCareerLeaders(ctx, noFilters())
	if err != nil {
		t.Fatalf("GetBattingCareerLeaders: %v", err)
	}
	if total != 2 {
		t.Fatalf("want total=2, got %d", total)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}

	// Default sort is smbWAR DESC; both NULL → tiebreak by last_name: Alpha before Beta.
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
	// CTE computes BA inline: 250H / 900AB ≈ 0.2778
	wantBA := float64(250) / float64(900)
	if rA.BA == nil {
		t.Error("BA should be computed by CTE, got nil")
	} else if got := *rA.BA; got != wantBA {
		t.Errorf("Alice BA: want %f, got %f", wantBA, got)
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

	rows, _, err := newLB(db).GetBattingCareerLeaders(ctx, models.LeaderboardFilters{OnlyHallOfFamers: true})
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

	rows, _, err := newLB(db).GetBattingCareerLeaders(ctx, models.LeaderboardFilters{Position: "SS"})
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
	rows, _, err := newLB(db).GetBattingCareerLeaders(ctx, models.LeaderboardFilters{SeasonStart: 2, SeasonEnd: 3})
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
	regRows, _, err := newLB(db).GetBattingCareerLeaders(ctx, noFilters())
	if err != nil {
		t.Fatalf("regular: %v", err)
	}
	if len(regRows) != 1 || regRows[0].HomeRuns != 10 {
		t.Errorf("regular HR: want 10, got %v", regRows)
	}

	// Playoffs: 3 HR
	playoffRows, _, err := newLB(db).GetBattingCareerLeaders(ctx, models.LeaderboardFilters{GameType: "playoffs"})
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

	rows, _, err := newLB(db).GetBattingCareerLeaders(ctx, models.LeaderboardFilters{GameType: "combined"})
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

	rows, _, err := newLB(db).GetBattingCareerLeaders(ctx, noFilters())
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

	rows, total, err := newLB(db).GetPitchingCareerLeaders(ctx, noFilters())
	if err != nil {
		t.Fatalf("GetPitchingCareerLeaders: %v", err)
	}
	if total != 1 {
		t.Fatalf("want total=1, got %d", total)
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
	// CTE computes ERA inline: (30+35)*27/570 = 65*27/570 ≈ 3.079
	wantERA := float64(65) * 27.0 / float64(570)
	if r.ERA == nil {
		t.Error("ERA should be computed by CTE, got nil")
	} else if got := *r.ERA; got != wantERA {
		t.Errorf("ERA: want %f, got %f", wantERA, got)
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

	rows, _, err := newLB(db).GetPitchingCareerLeaders(ctx, models.LeaderboardFilters{Position: "SP"})
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

	rows, _, err := newLB(db).GetPitchingCareerLeaders(ctx, models.LeaderboardFilters{ThrowHand: "L"})
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

	rows, _, err := newLB(db).GetPitchingCareerLeaders(ctx, models.LeaderboardFilters{GameType: "combined"})
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

// seedBattingFull seeds a batting stat row with explicitly separate games_played
// and plate_appearances values. seedBatting (in season_query_test.go) sets both
// equal to ab, which is fine for most tests but incorrect for qualification tests
// where games_played must reflect team games and pa must reflect actual PA.
func seedBattingFull(t *testing.T, db *sql.DB, playerSeasonID int64, isReg bool, gamesPlayed, pa, ab, hits, hr, rbi int) {
	t.Helper()
	isRegInt := 0
	if isReg {
		isRegInt = 1
	}
	_, err := db.ExecContext(context.Background(), `
INSERT INTO player_season_batting_stats
    (player_season_id, is_regular_season, games_played, games_batting,
     at_bats, plate_appearances, runs, hits, doubles, triples, home_runs, rbi,
     stolen_bases, caught_stealing, walks, strikeouts, hit_by_pitch,
     sac_hits, sac_flies, errors, passed_balls)
VALUES (?,?,?,?,?,?,0,?,0,0,?,?,0,0,0,0,0,0,0,0,0)
`, playerSeasonID, isRegInt, gamesPlayed, gamesPlayed, ab, pa, hits, hr, rbi)
	if err != nil {
		t.Fatalf("seedBattingFull: %v", err)
	}
}

// TestGetBattingSeasonLeaders_QualifiedOnly verifies that QualifiedOnly=true
// filters batters using the actual max games_played in the season (not the
// scheduled num_games), and that QualifiedOnly=false returns all rows.
func TestGetBattingSeasonLeaders_QualifiedOnly(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	// 100-game season. Threshold = MAX(games_played) * 3.1.
	// Both players played all 100 games → MAX = 100 → threshold = 310 PA.
	seedSeason(t, db, 1, 1, 100)
	t1 := seedTeam(t, db, "tgQ")
	h1 := seedTeamHistory(t, db, t1, 1, "Team Q", "E", "AL", 50, 50)

	pQ := seedPlayer(t, db, "gQQ", "Qual", "Batter")   // 350 PA, 100 games → qualifies
	pU := seedPlayer(t, db, "gUU", "Unqual", "Batter") // 200 PA, 100 games → does not qualify

	psQ := seedPlayerSeason(t, db, pQ, 1, &h1)
	psU := seedPlayerSeason(t, db, pU, 1, &h1)
	seedBattingFull(t, db, psQ, true, 100, 350, 350, 105, 20, 80)
	seedBattingFull(t, db, psU, true, 100, 200, 200, 60, 5, 25)

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

// TestGetBattingSeasonLeaders_QualifiedOnly_PartialImport verifies that a
// mid-season import (fewer actual games than the scheduled num_games) uses the
// real game count as the qualification denominator.
func TestGetBattingSeasonLeaders_QualifiedOnly_PartialImport(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	// Scheduled for 100 games, but only 60 were played before the import.
	// Threshold = MAX(games_played) * 3.1 = 60 * 3.1 = 186 PA.
	seedSeason(t, db, 1, 1, 100)
	t1 := seedTeam(t, db, "tgP")
	h1 := seedTeamHistory(t, db, t1, 1, "Team P", "E", "AL", 30, 30)

	pQ := seedPlayer(t, db, "gQP2", "Partial", "Qual")    // 200 PA, 60 games → qualifies (≥186)
	pU := seedPlayer(t, db, "gUP2", "Partial", "Unqual")  // 100 PA, 60 games → below threshold
	psQ := seedPlayerSeason(t, db, pQ, 1, &h1)
	psU := seedPlayerSeason(t, db, pU, 1, &h1)
	seedBattingFull(t, db, psQ, true, 60, 200, 200, 60, 5, 30)
	seedBattingFull(t, db, psU, true, 60, 100, 100, 30, 2, 15)

	rows, total, err := newLB(db).GetBattingSeasonLeaders(ctx, models.LeaderboardFilters{QualifiedOnly: true})
	if err != nil {
		t.Fatalf("partial import QualifiedOnly: %v", err)
	}
	if total != 1 {
		t.Errorf("partial import total: want 1, got %d", total)
	}
	if len(rows) != 1 || rows[0].PlayerID != pQ {
		t.Errorf("partial import: want playerID %d, got %v", pQ, rows)
	}
}

// TestGetBattingSeasonLeaders_QualifiedOnly_Playoffs verifies that playoff
// qualification uses the actual playoff games played (not the regular-season
// schedule length), so the threshold is reachable in a short series.
func TestGetBattingSeasonLeaders_QualifiedOnly_Playoffs(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	// 5-game playoff series. Threshold = MAX(games_played) * 3.1 = 5 * 3.1 = 15.5 PA.
	seedSeason(t, db, 1, 1, 100) // regular-season schedule irrelevant for playoff threshold
	t1 := seedTeam(t, db, "tgPO")
	h1 := seedTeamHistory(t, db, t1, 1, "Team PO", "E", "AL", 60, 40)

	pQ := seedPlayer(t, db, "gQPO", "Playoff", "Qual")   // 20 PA over 5 games → qualifies (≥15.5)
	pU := seedPlayer(t, db, "gUPO", "Playoff", "Unqual") // 8 PA → does not qualify
	psQ := seedPlayerSeason(t, db, pQ, 1, &h1)
	psU := seedPlayerSeason(t, db, pU, 1, &h1)
	// is_regular_season=false for playoff stats
	seedBattingFull(t, db, psQ, false, 5, 20, 18, 6, 1, 4)
	seedBattingFull(t, db, psU, false, 5, 8, 7, 2, 0, 1)

	rows, total, err := newLB(db).GetBattingSeasonLeaders(ctx,
		models.LeaderboardFilters{GameType: "playoffs", QualifiedOnly: true})
	if err != nil {
		t.Fatalf("playoff QualifiedOnly: %v", err)
	}
	if total != 1 {
		t.Errorf("playoff total: want 1, got %d", total)
	}
	if len(rows) != 1 || rows[0].PlayerID != pQ {
		t.Errorf("playoff: want playerID %d, got %v", pQ, rows)
	}
}

// TestGetPitchingSeasonLeaders_QualifiedOnly verifies that QualifiedOnly=true
// filters pitchers using the actual max batting games_played in the season as
// the team-games reference, and that QualifiedOnly=false returns all rows.
func TestGetPitchingSeasonLeaders_QualifiedOnly(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	// 100-game season. Threshold = MAX(batting games_played) * 3 outs = 300 outs.
	seedSeason(t, db, 1, 1, 100)
	t1 := seedTeam(t, db, "tgR")
	h1 := seedTeamHistory(t, db, t1, 1, "Team R", "W", "NL", 50, 50)

	// Seed a position player to establish the games-played denominator.
	// The pitching qualification uses MAX(batting games_played) as the game-count
	// reference, since it accurately tracks team games regardless of pitcher usage.
	anchor := seedPlayer(t, db, "gAnc", "Anchor", "Batter")
	psAnchor := seedPlayerSeason(t, db, anchor, 1, &h1)
	seedBattingFull(t, db, psAnchor, true, 100, 350, 350, 100, 10, 50)

	pQ := seedPlayer(t, db, "gQP", "Qual", "Pitcher")   // 350 outs → qualifies (≥300)
	pU := seedPlayer(t, db, "gUP", "Unqual", "Pitcher") // 100 outs → does not qualify

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

// TestGetBattingCareerLeaders_QualifiedOnly verifies that QualifiedOnly=true
// filters career batters whose total plate_appearances fall below the franchise
// threshold (num_games * 3000 / 162 rounded down) and that QualifiedOnly=false
// returns all rows.
func TestGetBattingCareerLeaders_QualifiedOnly(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	// 40-game season. Career PA threshold = int(3000 * 40 / 162) = 740.
	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tgCBQ")
	h1 := seedTeamHistory(t, db, t1, 1, "Team CQ", "E", "AL", 25, 15)

	pQ := seedPlayer(t, db, "gCBQ", "Qual", "Career")   // 800 career PA → qualifies (≥740)
	pU := seedPlayer(t, db, "gCBU", "Unqual", "Career") // 400 career PA → does not qualify

	psQ := seedPlayerSeason(t, db, pQ, 1, &h1)
	psU := seedPlayerSeason(t, db, pU, 1, &h1)
	seedBattingFull(t, db, psQ, true, 40, 800, 750, 200, 15, 80)
	seedBattingFull(t, db, psU, true, 40, 400, 380, 100, 5, 40)

	lb := newLB(db)

	rows, total, err := lb.GetBattingCareerLeaders(ctx, models.LeaderboardFilters{QualifiedOnly: true})
	if err != nil {
		t.Fatalf("QualifiedOnly=true: %v", err)
	}
	if total != 1 {
		t.Errorf("QualifiedOnly=true: want total=1, got %d", total)
	}
	if len(rows) != 1 || rows[0].PlayerID != pQ {
		t.Errorf("QualifiedOnly=true: want playerID %d, got %v", pQ, rows)
	}

	rows, total, err = lb.GetBattingCareerLeaders(ctx, models.LeaderboardFilters{QualifiedOnly: false})
	if err != nil {
		t.Fatalf("QualifiedOnly=false: %v", err)
	}
	if total != 2 {
		t.Errorf("QualifiedOnly=false: want total=2, got %d", total)
	}
	_ = rows
}

// TestGetPitchingCareerLeaders_QualifiedOnly verifies that QualifiedOnly=true
// filters career pitchers whose total outs_pitched fall below the franchise
// threshold (num_games * 3000 / 162 rounded down) and that QualifiedOnly=false
// returns all rows.
func TestGetPitchingCareerLeaders_QualifiedOnly(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	// 40-game season. Career outs threshold = int(3000 * 40 / 162) = 740.
	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tgCPQ")
	h1 := seedTeamHistory(t, db, t1, 1, "Team CPQ", "W", "NL", 22, 18)

	pQ := seedPlayer(t, db, "gCPQ", "Qual", "Pitcher")   // 800 outs → qualifies (≥740)
	pU := seedPlayer(t, db, "gCPU", "Unqual", "Pitcher") // 300 outs → does not qualify

	psQ := seedPlayerSeasonFull(t, db, pQ, 1, &h1, "P", "SP", "R", "R", "")
	psU := seedPlayerSeasonFull(t, db, pU, 1, &h1, "P", "RP", "R", "R", "")
	seedPitching(t, db, psQ, true, 14, 6, 800, 60, 240)
	seedPitching(t, db, psU, true, 4, 3, 300, 40, 80)

	lb := newLB(db)

	rows, total, err := lb.GetPitchingCareerLeaders(ctx, models.LeaderboardFilters{QualifiedOnly: true})
	if err != nil {
		t.Fatalf("QualifiedOnly=true: %v", err)
	}
	if total != 1 {
		t.Errorf("QualifiedOnly=true: want total=1, got %d", total)
	}
	if len(rows) != 1 || rows[0].PlayerID != pQ {
		t.Errorf("QualifiedOnly=true: want playerID %d, got %v", pQ, rows)
	}

	rows, total, err = lb.GetPitchingCareerLeaders(ctx, models.LeaderboardFilters{QualifiedOnly: false})
	if err != nil {
		t.Fatalf("QualifiedOnly=false: %v", err)
	}
	if total != 2 {
		t.Errorf("QualifiedOnly=false: want total=2, got %d", total)
	}
	_ = rows
}

// ── Career pagination and sort tests ─────────────────────────────────────────

func TestGetBattingCareerLeaders_Pagination(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	// Seed 60 players so we need 2 pages at pageSize=50.
	for i := range 60 {
		guid := fmt.Sprintf("g%03d", i)
		first := fmt.Sprintf("P%03d", i)
		pid := seedPlayer(t, db, guid, first, "Batter")
		ps := seedPlayerSeason(t, db, pid, 1, &h1)
		seedBatting(t, db, ps, true, 400, 100+i, i+1, 50)
	}

	lb := newLB(db)

	// Page 1: 50 rows; total = 60.
	page1, total, err := lb.GetBattingCareerLeaders(ctx, models.LeaderboardFilters{PageSize: 50, Offset: 0})
	if err != nil {
		t.Fatalf("page 1: %v", err)
	}
	if total != 60 {
		t.Errorf("total: want 60, got %d", total)
	}
	if len(page1) != 50 {
		t.Errorf("page 1 len: want 50, got %d", len(page1))
	}

	// Page 2: 10 remaining rows.
	page2, _, err := lb.GetBattingCareerLeaders(ctx, models.LeaderboardFilters{PageSize: 50, Offset: 50})
	if err != nil {
		t.Fatalf("page 2: %v", err)
	}
	if len(page2) != 10 {
		t.Errorf("page 2 len: want 10, got %d", len(page2))
	}

	// No player should appear in both pages.
	p1IDs := make(map[int64]bool, len(page1))
	for _, r := range page1 {
		p1IDs[r.PlayerID] = true
	}
	for _, r := range page2 {
		if p1IDs[r.PlayerID] {
			t.Errorf("player %d appears in both pages", r.PlayerID)
		}
	}
}

func TestGetBattingCareerLeaders_SortByHomeRunsDesc(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	pHigh := seedPlayer(t, db, "gH", "High", "Homer")
	pLow := seedPlayer(t, db, "gL", "Low", "Homer")
	psH := seedPlayerSeason(t, db, pHigh, 1, &h1)
	psL := seedPlayerSeason(t, db, pLow, 1, &h1)
	seedBatting(t, db, psH, true, 400, 100, 40, 100) // 40 HR
	seedBatting(t, db, psL, true, 400, 100, 10, 30)  // 10 HR

	rows, _, err := newLB(db).GetBattingCareerLeaders(ctx,
		models.LeaderboardFilters{SortField: "homeRuns", SortDesc: true})
	if err != nil {
		t.Fatalf("sort by homeRuns: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}
	if rows[0].PlayerID != pHigh {
		t.Errorf("want high HR player first, got playerID %d", rows[0].PlayerID)
	}
	if rows[0].HomeRuns != 40 {
		t.Errorf("first row HR: want 40, got %d", rows[0].HomeRuns)
	}
}

func TestGetBattingCareerLeaders_SortByBADesc(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	// pHigh: 120/300 = .400; pLow: 80/400 = .200
	pHigh := seedPlayer(t, db, "gH", "High", "Average")
	pLow := seedPlayer(t, db, "gL", "Low", "Average")
	psH := seedPlayerSeason(t, db, pHigh, 1, &h1)
	psL := seedPlayerSeason(t, db, pLow, 1, &h1)
	seedBatting(t, db, psH, true, 300, 120, 10, 50)
	seedBatting(t, db, psL, true, 400, 80, 5, 20)

	rows, _, err := newLB(db).GetBattingCareerLeaders(ctx,
		models.LeaderboardFilters{SortField: "ba", SortDesc: true})
	if err != nil {
		t.Fatalf("sort by ba: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}
	if rows[0].PlayerID != pHigh {
		t.Errorf("want high BA player first, got playerID %d", rows[0].PlayerID)
	}
}

func TestGetBattingCareerLeaders_FilterPlusPagination(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	// 3 HoF players, 2 non-HoF.
	for i := range 3 {
		pid := seedPlayer(t, db, fmt.Sprintf("gH%d", i), fmt.Sprintf("HoF%d", i), "Player")
		_, _ = db.ExecContext(ctx, `UPDATE players SET is_hall_of_famer = 1 WHERE id = ?`, pid)
		ps := seedPlayerSeason(t, db, pid, 1, &h1)
		seedBatting(t, db, ps, true, 400, 120, 20, 80)
	}
	for i := range 2 {
		pid := seedPlayer(t, db, fmt.Sprintf("gN%d", i), fmt.Sprintf("Norm%d", i), "Player")
		ps := seedPlayerSeason(t, db, pid, 1, &h1)
		seedBatting(t, db, ps, true, 300, 80, 5, 30)
	}

	rows, total, err := newLB(db).GetBattingCareerLeaders(ctx,
		models.LeaderboardFilters{OnlyHallOfFamers: true, PageSize: 50, Offset: 0})
	if err != nil {
		t.Fatalf("HoF filter + page: %v", err)
	}
	if total != 3 {
		t.Errorf("HoF filter total: want 3, got %d", total)
	}
	if len(rows) != 3 {
		t.Errorf("HoF filter rows: want 3, got %d", len(rows))
	}
}

func TestGetBattingCareerLeaders_EmptyResult(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)
	pid := seedPlayer(t, db, "gP", "Some", "Player")
	ps := seedPlayerSeason(t, db, pid, 1, &h1)
	seedBatting(t, db, ps, true, 400, 120, 20, 80)

	// HoF filter on non-HoF franchise → empty
	rows, total, err := newLB(db).GetBattingCareerLeaders(ctx,
		models.LeaderboardFilters{OnlyHallOfFamers: true})
	if err != nil {
		t.Fatalf("empty result: %v", err)
	}
	if total != 0 {
		t.Errorf("empty total: want 0, got %d", total)
	}
	if len(rows) != 0 {
		t.Errorf("empty rows: want 0, got %d", len(rows))
	}
}

func TestGetPitchingCareerLeaders_Pagination(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	for i := range 60 {
		guid := fmt.Sprintf("p%03d", i)
		first := fmt.Sprintf("P%03d", i)
		pid := seedPlayer(t, db, guid, first, "Pitcher")
		ps := seedPlayerSeason(t, db, pid, 1, &h1)
		seedPitching(t, db, ps, true, i+1, 0, 90+i, 10, 50)
	}

	_, total, err := newLB(db).GetPitchingCareerLeaders(ctx, models.LeaderboardFilters{PageSize: 50, Offset: 0})
	if err != nil {
		t.Fatalf("page 1: %v", err)
	}
	if total != 60 {
		t.Errorf("total: want 60, got %d", total)
	}

	page2, _, err := newLB(db).GetPitchingCareerLeaders(ctx, models.LeaderboardFilters{PageSize: 50, Offset: 50})
	if err != nil {
		t.Fatalf("page 2: %v", err)
	}
	if len(page2) != 10 {
		t.Errorf("page 2 len: want 10, got %d", len(page2))
	}
}

func TestGetPitchingCareerLeaders_SortByWinsDesc(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	pHigh := seedPlayer(t, db, "gH", "Ace", "Winner")
	pLow := seedPlayer(t, db, "gL", "Rook", "Pitcher")
	psH := seedPlayerSeason(t, db, pHigh, 1, &h1)
	psL := seedPlayerSeason(t, db, pLow, 1, &h1)
	seedPitching(t, db, psH, true, 20, 5, 270, 30, 100) // 20W
	seedPitching(t, db, psL, true, 5, 10, 180, 40, 80)  // 5W

	rows, _, err := newLB(db).GetPitchingCareerLeaders(ctx,
		models.LeaderboardFilters{SortField: "wins", SortDesc: true})
	if err != nil {
		t.Fatalf("sort by wins: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}
	if rows[0].PlayerID != pHigh {
		t.Errorf("want high W player first, got playerID %d", rows[0].PlayerID)
	}
}

func TestGetPitchingCareerLeaders_SortByERAAsc(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	h1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	// pLow: 10 ER / 270 OP = ERA 1.00; pHigh: 50 ER / 270 OP = ERA 5.00
	pLow := seedPlayer(t, db, "gL", "Good", "ERA")
	pHigh := seedPlayer(t, db, "gH", "Bad", "ERA")
	psL := seedPlayerSeason(t, db, pLow, 1, &h1)
	psH := seedPlayerSeason(t, db, pHigh, 1, &h1)
	seedPitching(t, db, psL, true, 15, 5, 270, 10, 100)
	seedPitching(t, db, psH, true, 8, 12, 270, 50, 80)

	rows, _, err := newLB(db).GetPitchingCareerLeaders(ctx,
		models.LeaderboardFilters{SortField: "era", SortDesc: false})
	if err != nil {
		t.Fatalf("sort by era asc: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}
	if rows[0].PlayerID != pLow {
		t.Errorf("want low ERA player first (asc sort), got playerID %d", rows[0].PlayerID)
	}
}

func TestGetBattingSeasonLeaders_MultiTeamPlayer(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	t2 := seedTeam(t, db, "tg2")
	h1 := seedTeamHistory(t, db, t1, 1, "Eagles", "E", "AL", 20, 20)
	h2 := seedTeamHistory(t, db, t2, 1, "Falcons", "W", "AL", 22, 18)

	// Traded player: started on Falcons (sort_order=1), finished on Eagles (sort_order=0).
	pTraded := seedPlayer(t, db, "gT", "Traded", "Player")
	// seedPlayerSeason inserts sort_order=0 for h1 (Eagles, the final team).
	psTraded := seedPlayerSeason(t, db, pTraded, 1, &h1)
	// Insert the prior team (Falcons) as sort_order=1.
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO player_season_teams (player_season_id, team_history_id, sort_order) VALUES (?, ?, 1)`,
		psTraded, h2)
	if err != nil {
		t.Fatalf("inserting prior team: %v", err)
	}
	seedBatting(t, db, psTraded, true, 400, 120, 20, 80)

	// Single-team player for comparison.
	pSingle := seedPlayer(t, db, "gS", "Single", "Player")
	psSingle := seedPlayerSeason(t, db, pSingle, 1, &h2)
	seedBatting(t, db, psSingle, true, 350, 100, 10, 50)

	rows, _, err := newLB(db).GetBattingSeasonLeaders(ctx, noFilters())
	if err != nil {
		t.Fatalf("GetBattingSeasonLeaders: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}

	// Find the traded player row and verify both teams are present.
	var tradedRow *models.BattingSeasonLeaderRow
	for i := range rows {
		if rows[i].PlayerID == pTraded {
			tradedRow = &rows[i]
		}
	}
	if tradedRow == nil {
		t.Fatal("traded player row not found")
		return
	}
	if len(tradedRow.Teams) != 2 {
		t.Fatalf("traded player: want 2 teams, got %d", len(tradedRow.Teams))
	}
	// sort_order=0 (Eagles) comes first in the ordered result.
	if tradedRow.Teams[0].TeamName != "Eagles" {
		t.Errorf("teams[0]: want Eagles (final team), got %q", tradedRow.Teams[0].TeamName)
	}
	if tradedRow.Teams[1].TeamName != "Falcons" {
		t.Errorf("teams[1]: want Falcons (prior team), got %q", tradedRow.Teams[1].TeamName)
	}

	// Single-team player should have exactly one team.
	var singleRow *models.BattingSeasonLeaderRow
	for i := range rows {
		if rows[i].PlayerID == pSingle {
			singleRow = &rows[i]
		}
	}
	if singleRow == nil {
		t.Fatal("single-team player row not found")
		return
	}
	if len(singleRow.Teams) != 1 {
		t.Fatalf("single-team player: want 1 team, got %d", len(singleRow.Teams))
	}
}

func TestGetPitchingSeasonLeaders_MultiTeamPlayer(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	t2 := seedTeam(t, db, "tg2")
	h1 := seedTeamHistory(t, db, t1, 1, "Eagles", "E", "AL", 20, 20)
	h2 := seedTeamHistory(t, db, t2, 1, "Falcons", "W", "AL", 22, 18)

	// Traded pitcher: started on Falcons (sort_order=1), finished on Eagles (sort_order=0).
	pTraded := seedPlayer(t, db, "gP", "Traded", "Pitcher")
	psTraded := seedPlayerSeason(t, db, pTraded, 1, &h1)
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO player_season_teams (player_season_id, team_history_id, sort_order) VALUES (?, ?, 1)`,
		psTraded, h2)
	if err != nil {
		t.Fatalf("inserting prior team: %v", err)
	}
	seedPitching(t, db, psTraded, true, 12, 6, 180, 40, 150)

	rows, _, err := newLB(db).GetPitchingSeasonLeaders(ctx, noFilters())
	if err != nil {
		t.Fatalf("GetPitchingSeasonLeaders: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	if len(rows[0].Teams) != 2 {
		t.Fatalf("traded pitcher: want 2 teams, got %d", len(rows[0].Teams))
	}
	if rows[0].Teams[0].TeamName != "Eagles" {
		t.Errorf("teams[0]: want Eagles (final team), got %q", rows[0].Teams[0].TeamName)
	}
	if rows[0].Teams[1].TeamName != "Falcons" {
		t.Errorf("teams[1]: want Falcons (prior team), got %q", rows[0].Teams[1].TeamName)
	}
}
