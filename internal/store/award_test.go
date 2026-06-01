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
	if len(all) != 30 {
		t.Errorf("expected 30 built-in awards, got %d", len(all))
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
		"League Champion", "Conference Champion",
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
	seedBatting(t, db, ps, true, 125, 44, 5, 30) // BA=.352, low HR/RBI

	// HR + RBI leader — different player, so Triple Crown won't fire.
	p2 := seedPlayer(t, db, "P3", "Mark", "Power")
	ps2 := seedPlayerSeason(t, db, p2, season, &th)
	seedBatting(t, db, ps2, true, 125, 36, 25, 80)

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

	// 40-game season → batting threshold: plate_appearances >= 124 (40*3.1); pitching: outs >= 120.
	season := seedSeason(t, db, 1, 1, 40)
	team := seedTeam(t, db, "CCC")
	th := seedTeamHistory(t, db, team, season, "Team C", "", "", 20, 20)

	// BA leader (.360), does NOT lead HR/RBI (so no Triple Crown).
	p1 := seedPlayer(t, db, "B1", "Al", "Bat")
	ps1 := seedPlayerSeason(t, db, p1, season, &th)
	seedBatting(t, db, ps1, true, 125, 45, 5, 30)

	// HR and RBI leader (lower BA).
	p2 := seedPlayer(t, db, "B2", "Bo", "Power")
	ps2 := seedPlayerSeason(t, db, p2, season, &th)
	seedBatting(t, db, ps2, true, 125, 36, 25, 80)

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
	seedBatting(t, db, ps1, true, 125, 48, 30, 100)

	// HR tie partner — should still get Home Run Title even when TC fires.
	p2 := seedPlayer(t, db, "TC2", "Bob", "Runner")
	ps2 := seedPlayerSeason(t, db, p2, season, &th)
	seedBatting(t, db, ps2, true, 125, 36, 30, 70)

	// Third batter: no titles.
	p3 := seedPlayer(t, db, "TC3", "Carl", "Third")
	ps3 := seedPlayerSeason(t, db, p3, season, &th)
	seedBatting(t, db, ps3, true, 125, 30, 5, 50)

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
	seedBatting(t, db, ps, true, 125, 42, 5, 30)

	// HR + RBI leader — keeps BA leader from getting TC.
	p2 := seedPlayer(t, db, "Q3", "Jake", "Slugger")
	ps2 := seedPlayerSeason(t, db, p2, season, &th)
	seedBatting(t, db, ps2, true, 125, 30, 25, 80)

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

// ── Championship awards ───────────────────────────────────────────────────────

func TestChampionshipAwards_AssignedToCorrectTeams(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	champTeam := seedTeam(t, db, "CHMP")
	champTH := seedTeamHistory(t, db, champTeam, season, "Champions", "", "", 30, 10)
	runnerTeam := seedTeam(t, db, "RUNP")
	runnerTH := seedTeamHistory(t, db, runnerTeam, season, "Runners-Up", "", "", 28, 12)
	otherTeam := seedTeam(t, db, "OTHR")
	otherTH := seedTeamHistory(t, db, otherTeam, season, "Others", "", "", 20, 20)

	// Final series (series 1): champion wins 3-1.
	seedPlayoffGame(t, db, season, 1, 1, champTH, runnerTH, 5, 2)
	seedPlayoffGame(t, db, season, 1, 2, runnerTH, champTH, 2, 4)
	seedPlayoffGame(t, db, season, 1, 3, champTH, runnerTH, 3, 1)
	seedPlayoffGame(t, db, season, 1, 4, runnerTH, champTH, 3, 2)
	setPlayoffConfig(t, db, season, 1, 5)

	p1 := seedPlayer(t, db, "CH1", "Alice", "Champion")
	ps1 := seedPlayerSeason(t, db, p1, season, &champTH)
	p2 := seedPlayer(t, db, "CH2", "Bob", "Champ")
	ps2 := seedPlayerSeason(t, db, p2, season, &champTH)
	p3 := seedPlayer(t, db, "RU1", "Carol", "Runner")
	ps3 := seedPlayerSeason(t, db, p3, season, &runnerTH)
	p4 := seedPlayer(t, db, "OT1", "Dave", "Other")
	ps4 := seedPlayerSeason(t, db, p4, season, &otherTH)

	if err := s.ComputeAndAssignStatLeaderAwards(ctx, season); err != nil {
		t.Fatalf("ComputeAndAssignStatLeaderAwards: %v", err)
	}

	byPS := awardsByPS(mustGetSeasonPlayerAwards(t, s, ctx, season))

	assertHasAward(t, byPS, ps1, "League Champion")
	assertHasAward(t, byPS, ps2, "League Champion")
	assertHasAward(t, byPS, ps3, "Conference Champion")
	assertNoAward(t, byPS, ps3, "League Champion")
	assertNoAward(t, byPS, ps4, "League Champion")
	assertNoAward(t, byPS, ps4, "Conference Champion")
}

func TestChampionshipAwards_Idempotent(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	champTeam := seedTeam(t, db, "IDM1")
	champTH := seedTeamHistory(t, db, champTeam, season, "Champs", "", "", 30, 10)
	otherTeam := seedTeam(t, db, "IDM2")
	otherTH := seedTeamHistory(t, db, otherTeam, season, "Others", "", "", 20, 20)
	seedPlayoffGame(t, db, season, 1, 1, champTH, otherTH, 5, 1)
	seedPlayoffGame(t, db, season, 1, 2, champTH, otherTH, 4, 2)
	seedPlayoffGame(t, db, season, 1, 3, champTH, otherTH, 6, 0)
	setPlayoffConfig(t, db, season, 1, 5)

	p := seedPlayer(t, db, "IDM", "Ida", "Leader")
	ps := seedPlayerSeason(t, db, p, season, &champTH)

	for i := range 3 {
		if err := s.ComputeAndAssignStatLeaderAwards(ctx, season); err != nil {
			t.Fatalf("run %d: ComputeAndAssignStatLeaderAwards: %v", i, err)
		}
	}

	byPS := awardsByPS(mustGetSeasonPlayerAwards(t, s, ctx, season))

	count := 0
	for _, name := range byPS[ps] {
		if name == "League Champion" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 League Champion after 3 runs, got %d (awards: %v)", count, byPS[ps])
	}
}

func TestChampionshipAwards_IncompletePlayoffs_NoAwards(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	champTeam := seedTeam(t, db, "INC1")
	champTH := seedTeamHistory(t, db, champTeam, season, "Pending", "", "", 30, 10)
	otherTeam := seedTeam(t, db, "INC2")
	otherTH := seedTeamHistory(t, db, otherTeam, season, "Other", "", "", 20, 20)

	// Two scored games, one unscored — completeness gate blocks champion detection.
	seedPlayoffGame(t, db, season, 1, 1, champTH, otherTH, 5, 1)
	seedPlayoffGame(t, db, season, 1, 2, champTH, otherTH, 4, 2)
	seedPlayoffGameNullScore(t, db, season, 1, 3, champTH, otherTH)

	p := seedPlayer(t, db, "INC", "Incomplete", "Season")
	ps := seedPlayerSeason(t, db, p, season, &champTH)

	if err := s.ComputeAndAssignStatLeaderAwards(ctx, season); err != nil {
		t.Fatalf("ComputeAndAssignStatLeaderAwards: %v", err)
	}

	byPS := awardsByPS(mustGetSeasonPlayerAwards(t, s, ctx, season))
	assertNoAward(t, byPS, ps, "League Champion")
	assertNoAward(t, byPS, ps, "Conference Champion")
}

func TestChampionshipAwards_CurrentTeamDeterminesEligibility(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	champTeam := seedTeam(t, db, "CUR1")
	champTH := seedTeamHistory(t, db, champTeam, season, "Champs", "", "", 30, 10)
	otherTeam := seedTeam(t, db, "CUR2")
	otherTH := seedTeamHistory(t, db, otherTeam, season, "Others", "", "", 20, 20)
	seedPlayoffGame(t, db, season, 1, 1, champTH, otherTH, 5, 1)
	seedPlayoffGame(t, db, season, 1, 2, champTH, otherTH, 4, 2)
	seedPlayoffGame(t, db, season, 1, 3, champTH, otherTH, 6, 0)
	setPlayoffConfig(t, db, season, 1, 5)

	// Player A: current team (sort_order=0) IS the champion.
	pA := seedPlayer(t, db, "CRA", "Traded", "To")
	psA := seedPlayerSeason(t, db, pA, season, &champTH)

	// Player B: current team (sort_order=0) is NOT the champion; has a historical
	// record (sort_order=1) on the champion team — should NOT get the award.
	pB := seedPlayer(t, db, "CRB", "Traded", "Away")
	psB := seedPlayerSeason(t, db, pB, season, &otherTH)
	_, err := db.ExecContext(context.Background(), `
INSERT OR IGNORE INTO player_season_teams (player_season_id, team_history_id, sort_order)
VALUES (?, ?, 1)
`, psB, champTH)
	if err != nil {
		t.Fatalf("inserting historical team for player B: %v", err)
	}

	if err := s.ComputeAndAssignStatLeaderAwards(ctx, season); err != nil {
		t.Fatalf("ComputeAndAssignStatLeaderAwards: %v", err)
	}

	byPS := awardsByPS(mustGetSeasonPlayerAwards(t, s, ctx, season))
	assertHasAward(t, byPS, psA, "League Champion")
	assertNoAward(t, byPS, psB, "League Champion")
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

	page, err := as.GetHoFInducted(ctx, 1, 25, 100)
	if err != nil {
		t.Fatalf("GetHoFInducted: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].PlayerID != p {
		t.Errorf("expected player %d in inducted, got %+v", p, page.Items)
	}

	if err := pqs.SetHallOfFamer(ctx, p, false); err != nil {
		t.Fatalf("SetHallOfFamer false: %v", err)
	}
	page, _ = as.GetHoFInducted(ctx, 1, 25, 100)
	if len(page.Items) != 0 {
		t.Errorf("expected empty inducted after removal, got %v", page.Items)
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

	candidatesPage, err := as.GetHoFCandidates(ctx, 1, 25, 100)
	if err != nil {
		t.Fatalf("GetHoFCandidates: %v", err)
	}

	foundRetired := false
	for _, c := range candidatesPage.Items {
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

// ── GetSeasonAwardCandidates — smbWAR regression ─────────────────────────────

func TestGetSeasonAwardCandidates_SmbWARNonNilWhenPresent(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	team := seedTeam(t, db, "SWAR1")
	th := seedTeamHistory(t, db, team, season, "WAR Team", "", "", 20, 20)
	p := seedPlayer(t, db, "SW1", "War", "Batter")
	ps := seedPlayerSeason(t, db, p, season, &th)
	seedBatting(t, db, ps, true, 125, 45, 15, 60)
	_, err := db.ExecContext(ctx,
		`UPDATE player_season_batting_stats SET smb_war = 3.7 WHERE player_season_id = ?`, ps)
	if err != nil {
		t.Fatalf("seeding smb_war: %v", err)
	}

	candidates, err := s.GetSeasonAwardCandidates(ctx, season)
	if err != nil {
		t.Fatalf("GetSeasonAwardCandidates: %v", err)
	}
	if len(candidates.TopBatters) == 0 {
		t.Fatal("expected at least one top batter")
	}
	if candidates.TopBatters[0].SmbWAR == nil {
		t.Error("expected non-nil SmbWAR on BattingCandidate when DB row has value")
	} else if *candidates.TopBatters[0].SmbWAR != 3.7 {
		t.Errorf("expected SmbWAR 3.7, got %v", *candidates.TopBatters[0].SmbWAR)
	}
}

func TestGetSeasonAwardCandidates_SmbWARNilWhenNull(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	team := seedTeam(t, db, "SWAR2")
	th := seedTeamHistory(t, db, team, season, "NoWAR Team", "", "", 20, 20)
	p := seedPlayer(t, db, "SW2", "NoWar", "Batter")
	ps := seedPlayerSeason(t, db, p, season, &th)
	seedBatting(t, db, ps, true, 125, 40, 10, 50)
	// smb_war left NULL (default from migration)

	candidates, err := s.GetSeasonAwardCandidates(ctx, season)
	if err != nil {
		t.Fatalf("GetSeasonAwardCandidates: %v", err)
	}
	if len(candidates.TopBatters) == 0 {
		t.Fatal("expected at least one top batter")
	}
	if candidates.TopBatters[0].SmbWAR != nil {
		t.Errorf("expected nil SmbWAR when DB has NULL, got %v", *candidates.TopBatters[0].SmbWAR)
	}
}

// ── GetSeasonAwardSummary ─────────────────────────────────────────────────────

func TestGetSeasonAwardSummary_EmptyWhenNoAwards(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)

	summary, err := s.GetSeasonAwardSummary(ctx, season)
	if err != nil {
		t.Fatalf("GetSeasonAwardSummary: %v", err)
	}
	if summary.Groups == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(summary.Groups) != 0 {
		t.Errorf("expected 0 groups, got %d", len(summary.Groups))
	}
}

func TestGetSeasonAwardSummary_GroupsOrderedByImportance(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	team := seedTeam(t, db, "ORD1")
	th := seedTeamHistory(t, db, team, season, "Order Team", "", "", 20, 20)

	p1 := seedPlayer(t, db, "OR1", "First", "Player")
	ps1 := seedPlayerSeason(t, db, p1, season, &th)
	p2 := seedPlayer(t, db, "OR2", "Second", "Player")
	ps2 := seedPlayerSeason(t, db, p2, season, &th)

	all, _ := s.ListAllAwards(ctx)
	var mvpID, silverSluggerID int64
	for _, a := range all {
		switch a.Name {
		case "MVP":
			mvpID = a.ID
		case "Silver Slugger":
			silverSluggerID = a.ID
		}
	}
	if mvpID == 0 || silverSluggerID == 0 {
		t.Fatal("required awards not found in seed data")
	}

	if err := s.SetPlayerSeasonAwards(ctx, ps1, []int64{mvpID}); err != nil {
		t.Fatalf("set MVP: %v", err)
	}
	if err := s.SetPlayerSeasonAwards(ctx, ps2, []int64{silverSluggerID}); err != nil {
		t.Fatalf("set Silver Slugger: %v", err)
	}

	summary, err := s.GetSeasonAwardSummary(ctx, season)
	if err != nil {
		t.Fatalf("GetSeasonAwardSummary: %v", err)
	}
	if len(summary.Groups) < 2 {
		t.Fatalf("expected at least 2 groups, got %d", len(summary.Groups))
	}
	// MVP importance < Silver Slugger importance → MVP first.
	if summary.Groups[0].Award.Importance > summary.Groups[1].Award.Importance {
		t.Errorf("groups not ordered by importance: first=%d second=%d",
			summary.Groups[0].Award.Importance, summary.Groups[1].Award.Importance)
	}
}

func TestGetSeasonAwardSummary_MultipleWinnersGroupedTogether(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	team := seedTeam(t, db, "GRP1")
	th := seedTeamHistory(t, db, team, season, "Group Team", "", "", 20, 20)

	p1 := seedPlayer(t, db, "GR1", "Alice", "Star")
	ps1 := seedPlayerSeason(t, db, p1, season, &th)
	p2 := seedPlayer(t, db, "GR2", "Bob", "Star")
	ps2 := seedPlayerSeason(t, db, p2, season, &th)

	all, _ := s.ListAllAwards(ctx)
	var allStarID int64
	for _, a := range all {
		if a.Name == "All-Star" {
			allStarID = a.ID
			break
		}
	}
	if allStarID == 0 {
		t.Fatal("All-Star award not found in seed data")
	}

	if err := s.SetPlayerSeasonAwards(ctx, ps1, []int64{allStarID}); err != nil {
		t.Fatalf("set All-Star ps1: %v", err)
	}
	if err := s.SetPlayerSeasonAwards(ctx, ps2, []int64{allStarID}); err != nil {
		t.Fatalf("set All-Star ps2: %v", err)
	}

	summary, err := s.GetSeasonAwardSummary(ctx, season)
	if err != nil {
		t.Fatalf("GetSeasonAwardSummary: %v", err)
	}
	if len(summary.Groups) != 1 {
		t.Fatalf("expected 1 group for All-Star, got %d", len(summary.Groups))
	}
	if summary.Groups[0].Award.Name != "All-Star" {
		t.Errorf("expected All-Star group, got %q", summary.Groups[0].Award.Name)
	}
	if len(summary.Groups[0].Winners) != 2 {
		t.Errorf("expected 2 winners in All-Star group, got %d", len(summary.Groups[0].Winners))
	}
}

func TestGetSeasonAwardSummary_ChampionshipAwardsExcluded(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	champTeam := seedTeam(t, db, "EXC1")
	champTH := seedTeamHistory(t, db, champTeam, season, "Champs", "", "", 30, 10)
	otherTeam := seedTeam(t, db, "EXC2")
	otherTH := seedTeamHistory(t, db, otherTeam, season, "Others", "", "", 20, 20)
	seedPlayoffGame(t, db, season, 1, 1, champTH, otherTH, 5, 1)
	seedPlayoffGame(t, db, season, 1, 2, champTH, otherTH, 4, 2)
	seedPlayoffGame(t, db, season, 1, 3, champTH, otherTH, 6, 0)
	setPlayoffConfig(t, db, season, 1, 5)

	p := seedPlayer(t, db, "EX1", "Champ", "Player")
	ps := seedPlayerSeason(t, db, p, season, &champTH)
	seedBatting(t, db, ps, true, 125, 40, 10, 50)

	if err := s.ComputeAndAssignStatLeaderAwards(ctx, season); err != nil {
		t.Fatalf("ComputeAndAssignStatLeaderAwards: %v", err)
	}

	// Confirm League Champion was assigned via the standard path.
	awards := awardsByPS(mustGetSeasonPlayerAwards(t, s, ctx, season))
	assertHasAward(t, awards, ps, "League Champion")

	// GetSeasonAwardSummary must NOT include League Champion.
	summary, err := s.GetSeasonAwardSummary(ctx, season)
	if err != nil {
		t.Fatalf("GetSeasonAwardSummary: %v", err)
	}
	for _, g := range summary.Groups {
		if g.Award.Name == "League Champion" || g.Award.Name == "Conference Champion" {
			t.Errorf("championship award %q must not appear in award summary", g.Award.Name)
		}
	}
}

func TestGetSeasonAwardSummary_BattingStatsOnWinner(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	team := seedTeam(t, db, "BST1")
	th := seedTeamHistory(t, db, team, season, "Bat Team", "", "", 20, 20)

	p := seedPlayer(t, db, "BS1", "Slugger", "Stats")
	ps := seedPlayerSeason(t, db, p, season, &th)
	seedBatting(t, db, ps, true, 125, 45, 20, 80)
	// Seed ba and smb_war as stored rate stats.
	_, err := db.ExecContext(ctx,
		`UPDATE player_season_batting_stats SET ba = 0.360, smb_war = 4.2 WHERE player_season_id = ?`, ps)
	if err != nil {
		t.Fatalf("seeding rates: %v", err)
	}

	all, _ := s.ListAllAwards(ctx)
	var silverSluggerID int64
	for _, a := range all {
		if a.Name == "Silver Slugger" {
			silverSluggerID = a.ID
			break
		}
	}
	if err := s.SetPlayerSeasonAwards(ctx, ps, []int64{silverSluggerID}); err != nil {
		t.Fatalf("set award: %v", err)
	}

	summary, err := s.GetSeasonAwardSummary(ctx, season)
	if err != nil {
		t.Fatalf("GetSeasonAwardSummary: %v", err)
	}
	if len(summary.Groups) == 0 {
		t.Fatal("expected at least one group")
	}
	w := summary.Groups[0].Winners[0]
	if w.HR != 20 {
		t.Errorf("expected HR=20, got %d", w.HR)
	}
	if w.RBI != 80 {
		t.Errorf("expected RBI=80, got %d", w.RBI)
	}
	if w.BA != 0.360 {
		t.Errorf("expected BA=0.360, got %f", w.BA)
	}
	if w.SmbWAR == nil {
		t.Error("expected non-nil SmbWAR")
	} else if *w.SmbWAR != 4.2 {
		t.Errorf("expected SmbWAR=4.2, got %f", *w.SmbWAR)
	}
}

func TestGetSeasonAwardSummary_PitchingStatsOnWinner(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	team := seedTeam(t, db, "PST1")
	th := seedTeamHistory(t, db, team, season, "Pitch Team", "", "", 20, 20)

	p := seedPlayer(t, db, "PS1", "Ace", "Pitcher")
	ps := seedPlayerSeason(t, db, p, season, &th)
	seedPitching(t, db, ps, true, 18, 5, 120, 10, 160)
	_, err := db.ExecContext(ctx,
		`UPDATE player_season_pitching_stats SET era = 2.25, smb_war = 5.1 WHERE player_season_id = ?`, ps)
	if err != nil {
		t.Fatalf("seeding rates: %v", err)
	}

	all, _ := s.ListAllAwards(ctx)
	var cyID int64
	for _, a := range all {
		if a.Name == "Cy Young" {
			cyID = a.ID
			break
		}
	}
	if err := s.SetPlayerSeasonAwards(ctx, ps, []int64{cyID}); err != nil {
		t.Fatalf("set award: %v", err)
	}

	summary, err := s.GetSeasonAwardSummary(ctx, season)
	if err != nil {
		t.Fatalf("GetSeasonAwardSummary: %v", err)
	}
	if len(summary.Groups) == 0 {
		t.Fatal("expected at least one group")
	}
	w := summary.Groups[0].Winners[0]
	if w.Wins != 18 {
		t.Errorf("expected Wins=18, got %d", w.Wins)
	}
	if w.Strikeouts != 160 {
		t.Errorf("expected Strikeouts=160, got %d", w.Strikeouts)
	}
	if w.ERA != 2.25 {
		t.Errorf("expected ERA=2.25, got %f", w.ERA)
	}
	if w.SmbWAR == nil {
		t.Error("expected non-nil SmbWAR on pitching winner")
	} else if *w.SmbWAR != 5.1 {
		t.Errorf("expected SmbWAR=5.1, got %f", *w.SmbWAR)
	}
}

func TestGetSeasonAwardSummary_RunnerUpGroupedUnderParent(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	team := seedTeam(t, db, "RUB1")
	th := seedTeamHistory(t, db, team, season, "RunUp Team", "", "", 20, 20)

	winner := seedPlayer(t, db, "RU1", "Win", "Player")
	psWin := seedPlayerSeason(t, db, winner, season, &th)
	seedBatting(t, db, psWin, true, 125, 45, 20, 80)

	ruPlayer := seedPlayer(t, db, "RU2", "Run", "Up")
	psRU := seedPlayerSeason(t, db, ruPlayer, season, &th)
	seedBatting(t, db, psRU, true, 125, 38, 15, 60)

	all, _ := s.ListAllAwards(ctx)
	var mvpID, mvp2ID int64
	for _, a := range all {
		switch a.Name {
		case "MVP":
			mvpID = a.ID
		case "MVP-2":
			mvp2ID = a.ID
		}
	}
	if mvpID == 0 || mvp2ID == 0 {
		t.Fatal("required awards not found in seed data")
	}

	if err := s.SetPlayerSeasonAwards(ctx, psWin, []int64{mvpID}); err != nil {
		t.Fatalf("set MVP: %v", err)
	}
	if err := s.SetPlayerSeasonAwards(ctx, psRU, []int64{mvp2ID}); err != nil {
		t.Fatalf("set MVP-2: %v", err)
	}

	summary, err := s.GetSeasonAwardSummary(ctx, season)
	if err != nil {
		t.Fatalf("GetSeasonAwardSummary: %v", err)
	}
	// MVP and MVP-2 must collapse into one group, not two.
	if len(summary.Groups) != 1 {
		t.Fatalf("expected 1 group (MVP-2 grouped under MVP), got %d", len(summary.Groups))
	}
	g := summary.Groups[0]
	if g.Award.Name != "MVP" {
		t.Errorf("expected group award name MVP, got %q", g.Award.Name)
	}
	if len(g.Winners) != 1 {
		t.Fatalf("expected 1 winner, got %d", len(g.Winners))
	}
	if g.Winners[0].PlayerID != winner {
		t.Errorf("expected winner player %d, got %d", winner, g.Winners[0].PlayerID)
	}
	if len(g.RunnerUps) != 1 {
		t.Fatalf("expected 1 runner-up, got %d", len(g.RunnerUps))
	}
	if g.RunnerUps[0].PlayerID != ruPlayer {
		t.Errorf("expected runner-up player %d, got %d", ruPlayer, g.RunnerUps[0].PlayerID)
	}
}

func TestGetSeasonAwardSummary_MultipleRunnerUpsInRankOrder(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	team := seedTeam(t, db, "RUM1")
	th := seedTeamHistory(t, db, team, season, "Multi RU Team", "", "", 20, 20)

	winner := seedPlayer(t, db, "MRU1", "Win", "Player")
	psWin := seedPlayerSeason(t, db, winner, season, &th)
	seedBatting(t, db, psWin, true, 125, 45, 20, 80)

	ru1 := seedPlayer(t, db, "MRU2", "Second", "Place")
	psRU1 := seedPlayerSeason(t, db, ru1, season, &th)
	seedBatting(t, db, psRU1, true, 125, 38, 15, 60)

	ru2 := seedPlayer(t, db, "MRU3", "Third", "Place")
	psRU2 := seedPlayerSeason(t, db, ru2, season, &th)
	seedBatting(t, db, psRU2, true, 125, 30, 10, 50)

	all, _ := s.ListAllAwards(ctx)
	var mvpID, mvp2ID, mvp3ID int64
	for _, a := range all {
		switch a.Name {
		case "MVP":
			mvpID = a.ID
		case "MVP-2":
			mvp2ID = a.ID
		case "MVP-3":
			mvp3ID = a.ID
		}
	}
	if mvpID == 0 || mvp2ID == 0 || mvp3ID == 0 {
		t.Fatal("required awards not found in seed data")
	}

	if err := s.SetPlayerSeasonAwards(ctx, psWin, []int64{mvpID}); err != nil {
		t.Fatalf("set MVP: %v", err)
	}
	if err := s.SetPlayerSeasonAwards(ctx, psRU1, []int64{mvp2ID}); err != nil {
		t.Fatalf("set MVP-2: %v", err)
	}
	if err := s.SetPlayerSeasonAwards(ctx, psRU2, []int64{mvp3ID}); err != nil {
		t.Fatalf("set MVP-3: %v", err)
	}

	summary, err := s.GetSeasonAwardSummary(ctx, season)
	if err != nil {
		t.Fatalf("GetSeasonAwardSummary: %v", err)
	}
	if len(summary.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(summary.Groups))
	}
	g := summary.Groups[0]
	if len(g.RunnerUps) != 2 {
		t.Fatalf("expected 2 runner-ups, got %d", len(g.RunnerUps))
	}
	// RunnerUps must be in rank order: MVP-2 (rank 1) before MVP-3 (rank 2).
	if g.RunnerUps[0].PlayerID != ru1 {
		t.Errorf("expected RunnerUps[0] player %d (rank-1), got %d", ru1, g.RunnerUps[0].PlayerID)
	}
	if g.RunnerUps[1].PlayerID != ru2 {
		t.Errorf("expected RunnerUps[1] player %d (rank-2), got %d", ru2, g.RunnerUps[1].PlayerID)
	}
}

func TestGetSeasonAwardSummary_NoRunnerUpsWhenNoneAssigned(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	team := seedTeam(t, db, "NRU1")
	th := seedTeamHistory(t, db, team, season, "NoRU Team", "", "", 20, 20)

	p := seedPlayer(t, db, "NRU", "Solo", "Winner")
	ps := seedPlayerSeason(t, db, p, season, &th)
	seedBatting(t, db, ps, true, 125, 40, 15, 60)

	all, _ := s.ListAllAwards(ctx)
	var mvpID int64
	for _, a := range all {
		if a.Name == "MVP" {
			mvpID = a.ID
			break
		}
	}
	if err := s.SetPlayerSeasonAwards(ctx, ps, []int64{mvpID}); err != nil {
		t.Fatalf("set MVP: %v", err)
	}

	summary, err := s.GetSeasonAwardSummary(ctx, season)
	if err != nil {
		t.Fatalf("GetSeasonAwardSummary: %v", err)
	}
	if len(summary.Groups) == 0 {
		t.Fatal("expected at least one group")
	}
	if len(summary.Groups[0].RunnerUps) != 0 {
		t.Errorf("expected empty RunnerUps when no runner-up award assigned, got %d", len(summary.Groups[0].RunnerUps))
	}
}

func TestGetSeasonAwardSummary_MultiWinnerAwardHasNoRunnerUps(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	team := seedTeam(t, db, "MWR1")
	th := seedTeamHistory(t, db, team, season, "Multi Team", "", "", 20, 20)

	p1 := seedPlayer(t, db, "MW1", "Multi", "One")
	ps1 := seedPlayerSeason(t, db, p1, season, &th)
	p2 := seedPlayer(t, db, "MW2", "Multi", "Two")
	ps2 := seedPlayerSeason(t, db, p2, season, &th)

	all, _ := s.ListAllAwards(ctx)
	var allStarID int64
	for _, a := range all {
		if a.Name == "All-Star" {
			allStarID = a.ID
			break
		}
	}
	if err := s.SetPlayerSeasonAwards(ctx, ps1, []int64{allStarID}); err != nil {
		t.Fatalf("set All-Star ps1: %v", err)
	}
	if err := s.SetPlayerSeasonAwards(ctx, ps2, []int64{allStarID}); err != nil {
		t.Fatalf("set All-Star ps2: %v", err)
	}

	summary, err := s.GetSeasonAwardSummary(ctx, season)
	if err != nil {
		t.Fatalf("GetSeasonAwardSummary: %v", err)
	}
	if len(summary.Groups) == 0 {
		t.Fatal("expected at least one group")
	}
	if len(summary.Groups[0].RunnerUps) != 0 {
		t.Errorf("expected empty RunnerUps for All-Star (no child awards), got %d", len(summary.Groups[0].RunnerUps))
	}
}

func TestGetSeasonAwardSummary_NilSmbWARWhenNull(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewAwardStore(db)

	season := seedSeason(t, db, 1, 1, 40)
	team := seedTeam(t, db, "NWR1")
	th := seedTeamHistory(t, db, team, season, "NullWAR Team", "", "", 20, 20)

	p := seedPlayer(t, db, "NW1", "Null", "War")
	ps := seedPlayerSeason(t, db, p, season, &th)
	seedBatting(t, db, ps, true, 125, 40, 10, 50)
	// smb_war deliberately left NULL

	all, _ := s.ListAllAwards(ctx)
	var mvpID int64
	for _, a := range all {
		if a.Name == "MVP" {
			mvpID = a.ID
			break
		}
	}
	if err := s.SetPlayerSeasonAwards(ctx, ps, []int64{mvpID}); err != nil {
		t.Fatalf("set award: %v", err)
	}

	summary, err := s.GetSeasonAwardSummary(ctx, season)
	if err != nil {
		t.Fatalf("GetSeasonAwardSummary: %v", err)
	}
	if len(summary.Groups) == 0 {
		t.Fatal("expected at least one group")
	}
	if summary.Groups[0].Winners[0].SmbWAR != nil {
		t.Errorf("expected nil SmbWAR when DB has NULL, got %v", *summary.Groups[0].Winners[0].SmbWAR)
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
