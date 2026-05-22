package store_test

import (
	"context"
	"slices"
	"testing"

	"smb-tools/internal/models"
	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

// ── Award seeding ─────────────────────────────────────────────────────────────

func TestAwards_SeedViaMigration(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	all, err := s.ListAllAwards(ctx)
	if err != nil {
		t.Fatalf("ListAllAwards: %v", err)
	}
	if len(all) != 28 {
		t.Errorf("expected 28 built-in awards, got %d", len(all))
	}

	byName := map[string]bool{}
	for _, a := range all {
		byName[a.Name] = true
		if !a.IsBuiltIn {
			t.Errorf("award %q: expected IsBuiltIn=true", a.Name)
		}
	}
	for _, name := range []string{
		"MVP", "Cy Young", "Gold Glove", "Triple Crown (Batting)",
		"Batting Title", "ERA Title", "All-Star", "ROY-5",
	} {
		if !byName[name] {
			t.Errorf("expected built-in award %q to be present", name)
		}
	}
}

// ── Custom awards ─────────────────────────────────────────────────────────────

func TestAwards_CreateCustom(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	id, err := s.CreateCustomAward(ctx, models.Award{
		Name:             "Iron Man",
		OriginalName:     "Iron Man",
		Importance:       3,
		IsBattingAward:   true,
		IsUserAssignable: true,
	})
	if err != nil {
		t.Fatalf("CreateCustomAward: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero ID")
	}

	// Duplicate name must fail.
	_, err = s.CreateCustomAward(ctx, models.Award{Name: "Iron Man", OriginalName: "Iron Man"})
	if err == nil {
		t.Error("expected error for duplicate award name")
	}
}

// ── SetPlayerSeasonAwards ─────────────────────────────────────────────────────

func TestSetPlayerSeasonAwards_Replace(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	team := seedTeam(t, db, "AAA")
	th := seedTeamHistory(t, db, team, season, "Team A", "", "", 20, 20)
	p := seedPlayer(t, db, "P1", "John", "Doe")
	ps := seedPlayerSeason(t, db, p, season, &th)

	all, _ := s.ListAllAwards(ctx)
	var mvpID, cyID int64
	for _, a := range all {
		switch a.Name {
		case "MVP":
			mvpID = a.ID
		case "Cy Young":
			cyID = a.ID
		}
	}

	if err := s.SetPlayerSeasonAwards(ctx, ps, []int64{mvpID}); err != nil {
		t.Fatalf("SetPlayerSeasonAwards MVP: %v", err)
	}
	rows, err := s.GetSeasonPlayerAwards(ctx, season)
	if err != nil {
		t.Fatalf("GetSeasonPlayerAwards: %v", err)
	}
	byPS := awardsByPS(rows)
	if !slices.Contains(byPS[ps], "MVP") {
		t.Errorf("expected MVP for player-season %d, got %v", ps, byPS[ps])
	}

	// Replace with Cy Young.
	if err := s.SetPlayerSeasonAwards(ctx, ps, []int64{cyID}); err != nil {
		t.Fatalf("SetPlayerSeasonAwards CyYoung: %v", err)
	}
	rows, _ = s.GetSeasonPlayerAwards(ctx, season)
	byPS = awardsByPS(rows)
	if slices.Contains(byPS[ps], "MVP") {
		t.Error("MVP should have been removed")
	}
	if !slices.Contains(byPS[ps], "Cy Young") {
		t.Errorf("expected Cy Young for player-season %d, got %v", ps, byPS[ps])
	}
}

func TestSetPlayerSeasonAwards_AutoAwardsUntouched(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	team := seedTeam(t, db, "BBB")
	th := seedTeamHistory(t, db, team, season, "Team B", "", "", 25, 15)

	// BA leader — does NOT lead HR/RBI, so no Triple Crown.
	p := seedPlayer(t, db, "P2", "Jane", "Smith")
	ps := seedPlayerSeason(t, db, p, season, &th)
	seedBatting(t, db, ps, true, 120, 44, 5, 30) // BA=.367, low HR/RBI

	// HR + RBI leader — different player, so Triple Crown won't fire.
	p2 := seedPlayer(t, db, "P3", "Mark", "Power")
	ps2 := seedPlayerSeason(t, db, p2, season, &th)
	seedBatting(t, db, ps2, true, 120, 36, 25, 80)

	// Pitcher so ERA computation has data to work with.
	pp := seedPlayer(t, db, "P2P", "Pete", "Pitcher")
	pps := seedPlayerSeason(t, db, pp, season, &th)
	seedPitching(t, db, pps, true, 15, 5, 120, 10, 120)

	if err := s.ComputeAndAssignStatLeaderAwards(ctx, season); err != nil {
		t.Fatalf("ComputeAndAssignStatLeaderAwards: %v", err)
	}

	all, _ := s.ListAllAwards(ctx)
	var mvpID int64
	for _, a := range all {
		if a.Name == "MVP" {
			mvpID = a.ID
		}
	}

	if err := s.SetPlayerSeasonAwards(ctx, ps, []int64{mvpID}); err != nil {
		t.Fatalf("SetPlayerSeasonAwards: %v", err)
	}

	rows, _ := s.GetSeasonPlayerAwards(ctx, season)
	byPS := awardsByPS(rows)

	if !slices.Contains(byPS[ps], "Batting Title") {
		t.Errorf("expected Batting Title (auto) to survive SetPlayerSeasonAwards, got %v", byPS[ps])
	}
	if !slices.Contains(byPS[ps], "MVP") {
		t.Errorf("expected MVP (user) to be present, got %v", byPS[ps])
	}
	// Auto-computed award on the power hitter should be untouched too.
	if !slices.Contains(byPS[ps2], "Home Run Title") {
		t.Errorf("expected Home Run Title on power hitter, got %v", byPS[ps2])
	}
}

// ── ComputeAndAssignStatLeaderAwards ─────────────────────────────────────────

func TestComputeStatLeaderAwards_BasicLeaders(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	// 40-game season → batting threshold: at_bats >= 120; pitching: outs >= 120.
	season := seedSeason(t, db, 1, 1, 40)
	team := seedTeam(t, db, "CCC")
	th := seedTeamHistory(t, db, team, season, "Team C", "", "", 20, 20)

	// BA leader (.375), does NOT lead HR/RBI (so no Triple Crown).
	p1 := seedPlayer(t, db, "B1", "Al", "Bat")
	ps1 := seedPlayerSeason(t, db, p1, season, &th)
	seedBatting(t, db, ps1, true, 120, 45, 5, 30)

	// HR and RBI leader (lower BA).
	p2 := seedPlayer(t, db, "B2", "Bo", "Power")
	ps2 := seedPlayerSeason(t, db, p2, season, &th)
	seedBatting(t, db, ps2, true, 120, 36, 25, 80)

	// ERA leader: 120 outs (40 IP), 8 ER → ERA 1.80.
	p3 := seedPlayer(t, db, "P1", "Carl", "Arm")
	ps3 := seedPlayerSeason(t, db, p3, season, &th)
	seedPitching(t, db, ps3, true, 18, 5, 120, 8, 150)

	// W and K leader (worse ERA): 120 outs, 20 ER → ERA 4.50.
	p4 := seedPlayer(t, db, "P2", "Dan", "Strike")
	ps4 := seedPlayerSeason(t, db, p4, season, &th)
	seedPitching(t, db, ps4, true, 22, 3, 120, 20, 200)

	if err := s.ComputeAndAssignStatLeaderAwards(ctx, season); err != nil {
		t.Fatalf("ComputeAndAssignStatLeaderAwards: %v", err)
	}

	rows, err := s.GetSeasonPlayerAwards(ctx, season)
	if err != nil {
		t.Fatalf("GetSeasonPlayerAwards: %v", err)
	}

	byPS := awardsByPS(rows)
	assertHasAward(t, byPS, ps1, "Batting Title")
	assertHasAward(t, byPS, ps2, "Home Run Title")
	assertHasAward(t, byPS, ps2, "RBI Title")
	assertHasAward(t, byPS, ps3, "ERA Title")
	assertHasAward(t, byPS, ps4, "Wins Title")
	assertHasAward(t, byPS, ps4, "Strikeouts Title")
}

func TestComputeStatLeaderAwards_TripleCrownBatting(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	team := seedTeam(t, db, "DDD")
	th := seedTeamHistory(t, db, team, season, "Team D", "", "", 30, 10)

	// TC winner: leads BA, HR, RBI.
	p1 := seedPlayer(t, db, "TC1", "Triple", "Crown")
	ps1 := seedPlayerSeason(t, db, p1, season, &th)
	seedBatting(t, db, ps1, true, 120, 48, 30, 100)

	// HR tie partner — should still get Home Run Title even when TC fires.
	p2 := seedPlayer(t, db, "TC2", "Bob", "Runner")
	ps2 := seedPlayerSeason(t, db, p2, season, &th)
	seedBatting(t, db, ps2, true, 120, 36, 30, 70)

	// Third batter: no titles.
	p3 := seedPlayer(t, db, "TC3", "Carl", "Third")
	ps3 := seedPlayerSeason(t, db, p3, season, &th)
	seedBatting(t, db, ps3, true, 120, 30, 5, 50)

	// Pitcher so the pitching branch also executes.
	pp := seedPlayer(t, db, "PP1", "Ed", "Pitcher")
	pps := seedPlayerSeason(t, db, pp, season, &th)
	seedPitching(t, db, pps, true, 15, 5, 120, 10, 120)

	if err := s.ComputeAndAssignStatLeaderAwards(ctx, season); err != nil {
		t.Fatalf("ComputeAndAssignStatLeaderAwards: %v", err)
	}

	byPS := awardsByPS(mustGetSeasonPlayerAwards(t, s, ctx, season))

	// TC winner gets Triple Crown only, not the three individual titles.
	assertHasAward(t, byPS, ps1, "Triple Crown (Batting)")
	assertNoAward(t, byPS, ps1, "Batting Title")
	assertNoAward(t, byPS, ps1, "Home Run Title")
	assertNoAward(t, byPS, ps1, "RBI Title")

	// HR tie partner still gets Home Run Title.
	assertHasAward(t, byPS, ps2, "Home Run Title")
	assertNoAward(t, byPS, ps2, "Triple Crown (Batting)")

	if len(byPS[ps3]) != 0 {
		t.Errorf("third batter expected no awards, got %v", byPS[ps3])
	}
}

func TestComputeStatLeaderAwards_TripleCrownPitching(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	team := seedTeam(t, db, "EEE")
	th := seedTeamHistory(t, db, team, season, "Team E", "", "", 25, 15)

	// TC winner: leads ERA, W, K. 120 outs, 8 ER → ERA 1.80.
	p1 := seedPlayer(t, db, "PTC1", "Tony", "AceTC")
	ps1 := seedPlayerSeason(t, db, p1, season, &th)
	seedPitching(t, db, ps1, true, 22, 3, 120, 8, 220)

	// W tie partner — should still get Wins Title.
	p2 := seedPlayer(t, db, "PTC2", "Sam", "SecondAce")
	ps2 := seedPlayerSeason(t, db, p2, season, &th)
	seedPitching(t, db, ps2, true, 22, 5, 120, 25, 150)

	// Batter so batting branch has data.
	bp := seedPlayer(t, db, "BT1", "Frank", "Hitter")
	bps := seedPlayerSeason(t, db, bp, season, &th)
	seedBatting(t, db, bps, true, 120, 40, 12, 50)

	if err := s.ComputeAndAssignStatLeaderAwards(ctx, season); err != nil {
		t.Fatalf("ComputeAndAssignStatLeaderAwards: %v", err)
	}

	byPS := awardsByPS(mustGetSeasonPlayerAwards(t, s, ctx, season))

	assertHasAward(t, byPS, ps1, "Triple Crown (Pitching)")
	assertNoAward(t, byPS, ps1, "ERA Title")
	assertNoAward(t, byPS, ps1, "Wins Title")
	assertNoAward(t, byPS, ps1, "Strikeouts Title")

	// W tie partner still gets Wins Title.
	assertHasAward(t, byPS, ps2, "Wins Title")
}

func TestComputeStatLeaderAwards_Idempotent(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	team := seedTeam(t, db, "FFF")
	th := seedTeamHistory(t, db, team, season, "Team F", "", "", 20, 20)

	// BA leader — not HR/RBI leader, so no TC.
	p := seedPlayer(t, db, "Q1", "Ida", "Leader")
	ps := seedPlayerSeason(t, db, p, season, &th)
	seedBatting(t, db, ps, true, 120, 42, 5, 30)

	// HR + RBI leader — keeps BA leader from getting TC.
	p2 := seedPlayer(t, db, "Q3", "Jake", "Slugger")
	ps2 := seedPlayerSeason(t, db, p2, season, &th)
	seedBatting(t, db, ps2, true, 120, 30, 25, 80)

	// Pitcher.
	pp := seedPlayer(t, db, "Q2", "Pete", "Pitch")
	pps := seedPlayerSeason(t, db, pp, season, &th)
	seedPitching(t, db, pps, true, 18, 5, 120, 8, 130)

	for i := range 3 {
		if err := s.ComputeAndAssignStatLeaderAwards(ctx, season); err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
	}

	byPS := awardsByPS(mustGetSeasonPlayerAwards(t, s, ctx, season))

	// Exactly 1 Batting Title for the BA leader after 3 idempotent runs.
	count := 0
	for _, name := range byPS[ps] {
		if name == "Batting Title" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 Batting Title after 3 runs, got %d (awards: %v)", count, byPS[ps])
	}

	// Power hitter should have 1 each of HR and RBI titles.
	hrCount := 0
	for _, name := range byPS[ps2] {
		if name == "Home Run Title" {
			hrCount++
		}
	}
	if hrCount != 1 {
		t.Errorf("expected exactly 1 Home Run Title, got %d", hrCount)
	}

	_ = pps // pitcher used in seedPitching
}

// ── Hall of Fame ──────────────────────────────────────────────────────────────

func TestSetHallOfFamer(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	pqs := store.NewPlayerQueryStore(db)
	as := store.NewAwardStore(db)

	season1 := seedSeason(t, db, 1, 1, 40)
	season2 := seedSeason(t, db, 2, 2, 40)
	team := seedTeam(t, db, "GGG")
	th1 := seedTeamHistory(t, db, team, season1, "Team G", "", "", 20, 20)
	th2 := seedTeamHistory(t, db, team, season2, "Team G", "", "", 20, 20)
	p := seedPlayer(t, db, "HOF1", "Harry", "Vetran")
	seedPlayerSeason(t, db, p, season1, &th1)
	seedPlayerSeason(t, db, p, season2, &th2)

	if err := pqs.SetHallOfFamer(ctx, p, true); err != nil {
		t.Fatalf("SetHallOfFamer true: %v", err)
	}

	inducted, err := as.GetHoFInducted(ctx)
	if err != nil {
		t.Fatalf("GetHoFInducted: %v", err)
	}
	if len(inducted) != 1 || inducted[0].PlayerID != p {
		t.Errorf("expected player %d in inducted, got %+v", p, inducted)
	}

	if err := pqs.SetHallOfFamer(ctx, p, false); err != nil {
		t.Fatalf("SetHallOfFamer false: %v", err)
	}
	inducted, _ = as.GetHoFInducted(ctx)
	if len(inducted) != 0 {
		t.Errorf("expected empty inducted after removal, got %v", inducted)
	}
}

func TestGetHoFCandidates(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	as := store.NewAwardStore(db)

	season1 := seedSeason(t, db, 1, 1, 40)
	season2 := seedSeason(t, db, 2, 2, 40)
	team := seedTeam(t, db, "HHH")
	th1 := seedTeamHistory(t, db, team, season1, "Team H", "", "", 25, 15)
	th2 := seedTeamHistory(t, db, team, season2, "Team H", "", "", 25, 15)

	// Retired: only in season 1.
	retired := seedPlayer(t, db, "RET1", "Retired", "Guy")
	rps := seedPlayerSeason(t, db, retired, season1, &th1)
	seedBatting(t, db, rps, true, 100, 30, 5, 20)

	// Active: in both seasons.
	active := seedPlayer(t, db, "ACT1", "Active", "Player")
	seedPlayerSeason(t, db, active, season1, &th1)
	seedPlayerSeason(t, db, active, season2, &th2)

	candidates, err := as.GetHoFCandidates(ctx)
	if err != nil {
		t.Fatalf("GetHoFCandidates: %v", err)
	}

	foundRetired := false
	for _, c := range candidates {
		if c.PlayerID == active {
			t.Error("active player should not be a HoF candidate")
		}
		if c.PlayerID == retired {
			foundRetired = true
		}
	}
	if !foundRetired {
		t.Error("retired player should appear as HoF candidate")
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func awardsByPS(rows []models.PlayerSeasonAwardRow) map[int64][]string {
	m := map[int64][]string{}
	for _, r := range rows {
		for _, aw := range r.Awards {
			m[r.PlayerSeasonID] = append(m[r.PlayerSeasonID], aw.Name)
		}
	}
	return m
}

func mustGetSeasonPlayerAwards(t *testing.T, s *store.AwardStore, ctx context.Context, seasonID int64) []models.PlayerSeasonAwardRow {
	t.Helper()
	rows, err := s.GetSeasonPlayerAwards(ctx, seasonID)
	if err != nil {
		t.Fatalf("GetSeasonPlayerAwards: %v", err)
	}
	return rows
}

func assertHasAward(t *testing.T, byPS map[int64][]string, psID int64, name string) {
	t.Helper()
	if !slices.Contains(byPS[psID], name) {
		t.Errorf("player-season %d: expected award %q, got %v", psID, name, byPS[psID])
	}
}

func assertNoAward(t *testing.T, byPS map[int64][]string, psID int64, name string) {
	t.Helper()
	if slices.Contains(byPS[psID], name) {
		t.Errorf("player-season %d: unexpectedly has award %q", psID, name)
	}
}
