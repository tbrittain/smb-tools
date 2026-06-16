package store_test

import (
	"context"
	"database/sql"
	"testing"

	"smb-tools/internal/models"
	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

func TestSearchTeams(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	t2 := seedTeam(t, db, "tg2")
	seedTeamHistory(t, db, t1, 1, "Red Sox", "E", "AL", 90, 72)
	seedTeamHistory(t, db, t2, 1, "Red Wings", "E", "NL", 80, 82)

	tq := store.NewTeamQueryStore(db)

	t.Run("prefix match", func(t *testing.T) {
		res, err := tq.SearchTeams(ctx, "Red")
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 2 {
			t.Errorf("expected 2 teams for 'Red', got %d", len(res))
		}
	})

	t.Run("exact name", func(t *testing.T) {
		res, err := tq.SearchTeams(ctx, "Sox")
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 1 || res[0].TeamName != "Red Sox" {
			t.Errorf("expected Red Sox, got %v", res)
		}
	})

	t.Run("no match", func(t *testing.T) {
		res, err := tq.SearchTeams(ctx, "Yankees")
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 0 {
			t.Errorf("expected 0, got %d", len(res))
		}
	})
}

func TestGetTeamHistory_WithChampionFlag(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	seedSeason(t, db, 2, 2, 40)
	teamID := seedTeam(t, db, "tgchamp")
	rivalID := seedTeam(t, db, "tgrival")
	hist1 := seedTeamHistory(t, db, teamID, 1, "Champs", "W", "NL", 95, 47)
	hist2 := seedTeamHistory(t, db, teamID, 2, "Champs", "W", "NL", 88, 54)
	rivalHist1 := seedTeamHistory(t, db, rivalID, 1, "Rival", "E", "NL", 80, 62)
	rivalHist2 := seedTeamHistory(t, db, rivalID, 2, "Rival", "E", "NL", 70, 72)

	// Season 1: team wins final series 3-1
	seedPlayoffGame(t, db, 1, 2, 1, hist1, rivalHist1, 4, 1)
	seedPlayoffGame(t, db, 1, 2, 2, hist1, rivalHist1, 3, 2)
	seedPlayoffGame(t, db, 1, 2, 3, rivalHist1, hist1, 5, 0) // rival wins at home
	seedPlayoffGame(t, db, 1, 2, 4, hist1, rivalHist1, 6, 1)
	setPlayoffConfig(t, db, 1, 1, 5)

	// Season 2: rival wins final series
	seedPlayoffGame(t, db, 2, 2, 1, rivalHist2, hist2, 3, 1)
	seedPlayoffGame(t, db, 2, 2, 2, rivalHist2, hist2, 2, 0)
	seedPlayoffGame(t, db, 2, 2, 3, rivalHist2, hist2, 4, 3)
	setPlayoffConfig(t, db, 2, 1, 5)

	tq := store.NewTeamQueryStore(db)
	th, err := tq.GetTeamHistory(ctx, teamID)
	if err != nil {
		t.Fatalf("GetTeamHistory: %v", err)
	}
	if len(th.Seasons) != 2 {
		t.Fatalf("expected 2 seasons, got %d", len(th.Seasons))
	}

	// Season 1: our team is champion (won 3 games, rival won 1)
	if !th.Seasons[0].IsChampion {
		t.Errorf("season 1: expected IsChampion=true")
	}
	// Season 2: rival is champion (won 3 games in final series)
	if th.Seasons[1].IsChampion {
		t.Errorf("season 2: expected IsChampion=false")
	}
}

func TestGetTeamSeasonRoster(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	t2 := seedTeam(t, db, "tg2")
	hist1 := seedTeamHistory(t, db, t1, 1, "Team One", "E", "AL", 20, 20)
	hist2 := seedTeamHistory(t, db, t2, 1, "Team Two", "W", "AL", 30, 10)

	// Three players on team 1, one on team 2
	for i := range 3 {
		guid := string(rune('a' + i))
		pid := seedPlayer(t, db, "g"+guid, "Player", guid)
		psid := seedPlayerSeason(t, db, pid, 1, &hist1)
		seedBatting(t, db, psid, true, 400, 100, 10, 50)
	}
	otherpid := seedPlayer(t, db, "gother", "Other", "Player")
	seedPlayerSeason(t, db, otherpid, 1, &hist2)

	tq := store.NewTeamQueryStore(db)
	roster, err := tq.GetTeamSeasonRoster(ctx, hist1)
	if err != nil {
		t.Fatalf("GetTeamSeasonRoster: %v", err)
	}
	if len(roster) != 3 {
		t.Errorf("expected 3 players on team 1, got %d", len(roster))
	}
	for _, r := range roster {
		if r.Batting == nil {
			t.Errorf("player %s: expected batting stats", r.LastName)
		}
	}
}

func TestGetTeamSeasonSchedule_FiltersByTeam(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	t2 := seedTeam(t, db, "tg2")
	t3 := seedTeam(t, db, "tg3")
	hist1 := seedTeamHistory(t, db, t1, 1, "Team One", "E", "AL", 20, 20)
	hist2 := seedTeamHistory(t, db, t2, 1, "Team Two", "W", "AL", 30, 10)
	hist3 := seedTeamHistory(t, db, t3, 1, "Team Three", "W", "NL", 25, 15)

	// 3 games: Team1 vs Team2 (home), Team1 vs Team3 (away), Team2 vs Team3
	insertScheduleGame(t, db, 1, 1, hist1, hist2, 5, 3)
	insertScheduleGame(t, db, 1, 2, hist3, hist1, 2, 4) // Team1 is away
	insertScheduleGame(t, db, 1, 3, hist2, hist3, 1, 0) // Team1 not involved

	tq := store.NewTeamQueryStore(db)
	games, err := tq.GetTeamSeasonSchedule(ctx, hist1, 1)
	if err != nil {
		t.Fatalf("GetTeamSeasonSchedule: %v", err)
	}
	if len(games) != 2 {
		t.Errorf("expected 2 games for Team One, got %d", len(games))
	}
}

func TestGetTeamSeasonSchedule_ScoresAndOrder(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tgA")
	t2 := seedTeam(t, db, "tgB")
	hist1 := seedTeamHistory(t, db, t1, 1, "Alpha", "E", "AL", 10, 10)
	hist2 := seedTeamHistory(t, db, t2, 1, "Beta", "E", "AL", 10, 10)

	// Game 1: hist1 home, wins 5-3. Game 2: hist1 away at hist2, wins 2-1 as visitor.
	// Game 3: unplayed (NULL scores).
	insertScheduleGame(t, db, 1, 1, hist1, hist2, 5, 3)
	insertScheduleGame(t, db, 1, 2, hist2, hist1, 1, 2) // hist1 is away, wins
	insertScheduleGameNullScore(t, db, 1, 3, hist1, hist2)

	tq := store.NewTeamQueryStore(db)
	games, err := tq.GetTeamSeasonSchedule(ctx, hist1, 1)
	if err != nil {
		t.Fatalf("GetTeamSeasonSchedule: %v", err)
	}
	if len(games) != 3 {
		t.Fatalf("expected 3 games, got %d", len(games))
	}

	// Ordered by game_number; TeamGameNum is 1-based sequential.
	g1, g2, g3 := games[0], games[1], games[2]
	if g1.TeamGameNum != 1 || g2.TeamGameNum != 2 || g3.TeamGameNum != 3 {
		t.Errorf("unexpected TeamGameNum sequence: %d %d %d", g1.TeamGameNum, g2.TeamGameNum, g3.TeamGameNum)
	}

	// Game 1: hist1 is home.
	if g1.HomeTeamHistoryID != hist1 {
		t.Errorf("game 1 home team: want %d, got %d", hist1, g1.HomeTeamHistoryID)
	}
	if g1.HomeScore == nil || *g1.HomeScore != 5 {
		t.Errorf("game 1 home score: want 5, got %v", g1.HomeScore)
	}
	if g1.AwayScore == nil || *g1.AwayScore != 3 {
		t.Errorf("game 1 away score: want 3, got %v", g1.AwayScore)
	}

	// Game 2: hist1 is away, wins 2-1.
	if g2.AwayTeamHistoryID != hist1 {
		t.Errorf("game 2 away team: want %d, got %d", hist1, g2.AwayTeamHistoryID)
	}
	if g2.AwayScore == nil || *g2.AwayScore != 2 {
		t.Errorf("game 2 away (hist1) score: want 2, got %v", g2.AwayScore)
	}

	// Game 3: scores are nil (unplayed).
	if g3.HomeScore != nil || g3.AwayScore != nil {
		t.Errorf("game 3 should have nil scores, got home=%v away=%v", g3.HomeScore, g3.AwayScore)
	}
}

func insertScheduleGameNullScore(t *testing.T, db *sql.DB, seasonID, gameNum int, homeHistID, awayHistID int64) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
INSERT INTO team_season_schedules
    (season_id, game_number, day, home_team_history_id, away_team_history_id,
     home_score, away_score)
VALUES (?,?,1,?,?,NULL,NULL)
`, seasonID, gameNum, homeHistID, awayHistID)
	if err != nil {
		t.Fatalf("insertScheduleGameNullScore: %v", err)
	}
}

func TestListAllTeamSeasons(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	seedSeason(t, db, 2, 2, 40)
	t1 := seedTeam(t, db, "tg1")
	t2 := seedTeam(t, db, "tg2")
	seedTeamHistory(t, db, t1, 1, "Alpha", "E", "AL", 80, 62)
	seedTeamHistory(t, db, t1, 2, "Alpha", "E", "AL", 90, 52)
	seedTeamHistory(t, db, t2, 1, "Beta", "W", "NL", 70, 72)
	seedTeamHistory(t, db, t2, 2, "Beta", "W", "NL", 60, 82)

	tq := store.NewTeamQueryStore(db)
	rows, err := tq.ListAllTeamSeasons(ctx)
	if err != nil {
		t.Fatalf("ListAllTeamSeasons: %v", err)
	}
	if len(rows) != 4 {
		t.Errorf("expected 4 rows, got %d", len(rows))
	}
	// Ordered by season_num DESC: season 2 rows first
	if rows[0].SeasonNum != 2 {
		t.Errorf("expected first row season=2, got %d", rows[0].SeasonNum)
	}
}

func TestGetTeamSeasonPlayoffSchedule_RoundNumbers(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	t2 := seedTeam(t, db, "tg2")
	t3 := seedTeam(t, db, "tg3")
	t4 := seedTeam(t, db, "tg4")
	hist1 := seedTeamHistory(t, db, t1, 1, "Team One", "E", "AL", 90, 52)
	hist2 := seedTeamHistory(t, db, t2, 1, "Team Two", "W", "AL", 80, 62)
	hist3 := seedTeamHistory(t, db, t3, 1, "Team Three", "E", "NL", 70, 72)
	hist4 := seedTeamHistory(t, db, t4, 1, "Team Four", "W", "NL", 60, 82)

	// 4-round bracket using opaque series numbers (as the real save game produces).
	// Series numbers 6, 11, 14, 15 are the 4 distinct series numbers in the season,
	// mapping to round 1 (Round of 16 equivalent in a 4-round bracket has 8 series,
	// but here we use a simplified 4-series bracket for testing: rounds 1-4).
	// Team1 only plays in series 6 (rank 1 of 4) and loses — should see "Round of 8"
	// (fromTop=2 for rank=1,totalSeries=4 is actually fromTop=bits.Len(4)-1=2 → "Round of 8").
	// Actually with 4 total series it's: rank1→"Round of 8", rank2→"Round of 8",
	// rank3→"Conference Championship", rank4→"League Championship".
	// Use 15 total series (standard 16-team bracket) for a realistic test.
	// Simplified: seed only the series our teams actually play in; the round_info
	// is derived from ALL series numbers in the season.
	// Add all 15 series of a 16-team bracket, teams only play in their path.
	// For simplicity, only seed the games the two test teams are in.
	// Series 1-8 = round of 16, 9-12 = round of 8, 13-14 = conf champ, 15 = final.
	// To keep the fixture simple, seed one game per series for the path through.
	// Also seed dummy series entries for the other first-round matchups (no scores needed
	// — the season_champion view won't break because scores are nullable).
	// Seed all 15 series so the bracket is complete.
	for s := 1; s <= 8; s++ {
		if s != 6 { // series 6 is team1 vs team2 — seeded below
			seedPlayoffGame(t, db, 1, s, s*10, hist3, hist4, 3, 1)
		}
	}
	for s := 9; s <= 12; s++ {
		if s != 11 { // series 11 is team2's quarterfinal
			seedPlayoffGame(t, db, 1, s, s*10, hist3, hist4, 2, 1)
		}
	}
	for s := 13; s <= 14; s++ {
		if s != 14 { // series 14 is team2's semifinal
			seedPlayoffGame(t, db, 1, s, s*10, hist3, hist4, 4, 2)
		}
	}
	// Series 6: team1 vs team2, team2 wins
	seedPlayoffGame(t, db, 1, 6, 61, hist2, hist1, 5, 3)
	seedPlayoffGame(t, db, 1, 6, 62, hist1, hist2, 2, 4)
	// Series 11: team2 advances
	seedPlayoffGame(t, db, 1, 11, 111, hist2, hist3, 3, 2)
	// Series 14: team2 advances
	seedPlayoffGame(t, db, 1, 14, 141, hist2, hist3, 4, 1)
	// Series 15: championship
	seedPlayoffGame(t, db, 1, 15, 151, hist2, hist4, 6, 5)

	tq := store.NewTeamQueryStore(db)

	t.Run("eliminated team sees round of 16 label despite only playing in series 6", func(t *testing.T) {
		games, err := tq.GetTeamSeasonPlayoffSchedule(ctx, hist1, 1)
		if err != nil {
			t.Fatalf("GetTeamSeasonPlayoffSchedule: %v", err)
		}
		if len(games) != 2 {
			t.Fatalf("expected 2 games for team1, got %d", len(games))
		}
		for _, g := range games {
			if g.RoundNumber != 1 {
				t.Errorf("expected RoundNumber=1, got %d", g.RoundNumber)
			}
			if g.RoundLabel != "Round of 16" {
				t.Errorf("expected RoundLabel='Round of 16', got %q", g.RoundLabel)
			}
		}
	})

	t.Run("finalist team sees correct labels through all rounds", func(t *testing.T) {
		games, err := tq.GetTeamSeasonPlayoffSchedule(ctx, hist2, 1)
		if err != nil {
			t.Fatalf("GetTeamSeasonPlayoffSchedule: %v", err)
		}
		if len(games) != 5 {
			t.Fatalf("expected 5 games for team2, got %d", len(games))
		}
		wantLabel := []string{"Round of 16", "Round of 16", "Round of 8", "Conference Championship", "League Championship"}
		wantRound := []int{1, 1, 2, 3, 4}
		for i, g := range games {
			if g.RoundNumber != wantRound[i] {
				t.Errorf("game[%d]: expected RoundNumber=%d, got %d", i, wantRound[i], g.RoundNumber)
			}
			if g.RoundLabel != wantLabel[i] {
				t.Errorf("game[%d]: expected RoundLabel=%q, got %q", i, wantLabel[i], g.RoundLabel)
			}
		}
	})
}

func TestPlayoffRoundInfo(t *testing.T) {
	cases := []struct {
		totalSeries int
		rank        int
		wantNum     int
		wantLabel   string
	}{
		// 1-round bracket (1 series = championship only)
		{1, 1, 1, "League Championship"},
		// 2-round bracket (3 series: 2 semis + 1 final)
		{3, 1, 1, "Conference Championship"},
		{3, 2, 1, "Conference Championship"},
		{3, 3, 2, "League Championship"},
		// 3-round bracket (7 series: 4 + 2 + 1)
		{7, 1, 1, "Round of 8"},
		{7, 4, 1, "Round of 8"},
		{7, 5, 2, "Conference Championship"},
		{7, 6, 2, "Conference Championship"},
		{7, 7, 3, "League Championship"},
		// 4-round bracket (15 series: 8 + 4 + 2 + 1)
		{15, 1, 1, "Round of 16"},
		{15, 8, 1, "Round of 16"},
		{15, 9, 2, "Round of 8"},
		{15, 12, 2, "Round of 8"},
		{15, 13, 3, "Conference Championship"},
		{15, 14, 3, "Conference Championship"},
		{15, 15, 4, "League Championship"},
	}
	for _, tc := range cases {
		n, l := store.PlayoffRoundInfo(tc.rank, tc.totalSeries)
		if n != tc.wantNum || l != tc.wantLabel {
			t.Errorf("PlayoffRoundInfo(rank=%d, total=%d): got (%d,%q), want (%d,%q)",
				tc.rank, tc.totalSeries, n, l, tc.wantNum, tc.wantLabel)
		}
	}
}

// TestGetTeamSeasonPlayoffSchedule_RoundLabels_MidPlayoff verifies that when
// playoff_rounds is set on the season, the round labels are derived from the
// full expected bracket size — not just the series already in the DB. Without
// this fix, a mid-playoff import (4 of 7 series present) would call the final
// round "League Championship" instead of "Round of 8".
func TestGetTeamSeasonPlayoffSchedule_RoundLabels_MidPlayoff(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)
	setPlayoffConfig(t, db, s1, 3, 7) // 3 rounds → 7 total series expected
	t1 := seedTeam(t, db, "mid-t1")
	t2 := seedTeam(t, db, "mid-t2")
	hist1 := seedTeamHistory(t, db, t1, s1, "Team One", "E", "AL", 90, 52)
	hist2 := seedTeamHistory(t, db, t2, s1, "Team Two", "W", "AL", 80, 62)

	// Only 4 of the 7 series are present (mid-playoff import).
	// Team 1 plays in series 4 (the highest-numbered present series).
	// Without the fix, series 4 of 4 would look like "League Championship".
	// With the fix, totalSeries=7, so series 4 is still in "Round of 8" territory.
	for s := 1; s <= 3; s++ {
		seedPlayoffGame(t, db, s1, s, 1, hist2, hist1, 4, 2)
	}
	seedPlayoffGame(t, db, s1, 4, 1, hist1, hist2, 3, 1)
	seedPlayoffGame(t, db, s1, 4, 2, hist2, hist1, 5, 4)

	tq := store.NewTeamQueryStore(db)
	games, err := tq.GetTeamSeasonPlayoffSchedule(ctx, hist1, s1)
	if err != nil {
		t.Fatalf("GetTeamSeasonPlayoffSchedule: %v", err)
	}
	if len(games) == 0 {
		t.Fatal("expected games for team1, got none")
	}
	for _, g := range games {
		if g.RoundLabel != "Round of 8" {
			t.Errorf("expected 'Round of 8' (mid-playoff, rounds=3, totalSeries=7), got %q", g.RoundLabel)
		}
	}
}

func TestGetTeamTopPlayers(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)
	s2 := seedSeason(t, db, 2, 2, 40)
	t1 := seedTeam(t, db, "tp-t1")
	t2 := seedTeam(t, db, "tp-t2")
	h1s1 := seedTeamHistory(t, db, t1, s1, "Team One", "E", "AL", 90, 52)
	h1s2 := seedTeamHistory(t, db, t1, s2, "Team One", "E", "AL", 85, 57)
	h2s1 := seedTeamHistory(t, db, t2, s1, "Team Two", "W", "AL", 80, 62)

	// Player A: 2 seasons on team 1, total smbWAR = 7.5
	pA := seedPlayer(t, db, "tp-pa", "Aaron", "Alpha")
	psA1 := seedPlayerSeason(t, db, pA, s1, &h1s1)
	psA2 := seedPlayerSeason(t, db, pA, s2, &h1s2)
	seedBatting(t, db, psA1, true, 500, 150, 20, 80)
	seedBatting(t, db, psA2, true, 480, 140, 18, 75)
	setTopPlayerBattingWAR(t, db, psA1, 125.0, 4.0)
	setTopPlayerBattingWAR(t, db, psA2, 118.0, 3.5)

	// Player B: 1 season on team 1, smbWAR = 1.5
	pB := seedPlayer(t, db, "tp-pb", "Brad", "Beta")
	psB1 := seedPlayerSeason(t, db, pB, s1, &h1s1)
	seedBatting(t, db, psB1, true, 400, 110, 10, 55)
	setTopPlayerBattingWAR(t, db, psB1, 105.0, 1.5)

	// Player C: only on team 2 — must not appear in team 1 results
	pC := seedPlayer(t, db, "tp-pc", "Carl", "Charlie")
	psC1 := seedPlayerSeason(t, db, pC, s1, &h2s1)
	seedBatting(t, db, psC1, true, 500, 160, 30, 100)
	setTopPlayerBattingWAR(t, db, psC1, 145.0, 6.0)

	tq := store.NewTeamQueryStore(db)

	t.Run("excludes players from other teams", func(t *testing.T) {
		results, err := tq.GetTeamTopPlayers(ctx, t1, 25)
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range results {
			if r.PlayerID == pC {
				t.Errorf("player C (team 2 only) should not appear in team 1 results")
			}
		}
		if len(results) != 2 {
			t.Errorf("expected 2 players for team 1, got %d", len(results))
		}
	})

	t.Run("sorted by total smbWAR descending", func(t *testing.T) {
		results, err := tq.GetTeamTopPlayers(ctx, t1, 25)
		if err != nil {
			t.Fatal(err)
		}
		if results[0].PlayerID != pA {
			t.Errorf("expected player A first (highest smbWAR 7.5), got playerID=%d", results[0].PlayerID)
		}
		const want = 7.5
		if abs64(results[0].TotalSmbWAR-want) > 0.001 {
			t.Errorf("expected player A total smbWAR=%.1f, got %.4f", want, results[0].TotalSmbWAR)
		}
	})

	t.Run("num_seasons counts correctly for multi-season player", func(t *testing.T) {
		results, err := tq.GetTeamTopPlayers(ctx, t1, 25)
		if err != nil {
			t.Fatal(err)
		}
		var got int
		for _, r := range results {
			if r.PlayerID == pA {
				got = r.NumSeasons
			}
		}
		if got != 2 {
			t.Errorf("expected player A NumSeasons=2, got %d", got)
		}
	})

	t.Run("capped at limit", func(t *testing.T) {
		results, err := tq.GetTeamTopPlayers(ctx, t1, 1)
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 1 {
			t.Errorf("expected 1 row with limit=1, got %d", len(results))
		}
		if results[0].PlayerID != pA {
			t.Errorf("expected player A as top result, got playerID=%d", results[0].PlayerID)
		}
	})

	t.Run("player with null smbWAR gets 0.0 and sorts to bottom", func(t *testing.T) {
		// Player D: on team 1 season 2 but no context stats set (smb_war remains NULL)
		pD := seedPlayer(t, db, "tp-pd", "Dave", "Delta")
		psD2 := seedPlayerSeason(t, db, pD, s2, &h1s2)
		seedBatting(t, db, psD2, true, 300, 80, 5, 30)
		// No setTopPlayerBattingWAR call — smb_war stays NULL

		results, err := tq.GetTeamTopPlayers(ctx, t1, 25)
		if err != nil {
			t.Fatal(err)
		}
		last := results[len(results)-1]
		if last.PlayerID != pD {
			t.Errorf("expected player D (null smbWAR→0) to sort last, got playerID=%d", last.PlayerID)
		}
		if last.TotalSmbWAR != 0.0 {
			t.Errorf("expected TotalSmbWAR=0.0 for null smb_war, got %f", last.TotalSmbWAR)
		}
	})

	t.Run("hof flag reflected correctly", func(t *testing.T) {
		_, err := db.ExecContext(ctx, `UPDATE players SET is_hall_of_famer=1 WHERE id=?`, pA)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_, _ = db.ExecContext(ctx, `UPDATE players SET is_hall_of_famer=0 WHERE id=?`, pA)
		})
		results, err := tq.GetTeamTopPlayers(ctx, t1, 25)
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range results {
			switch r.PlayerID {
			case pA:
				if !r.IsHallOfFamer {
					t.Errorf("player A: expected IsHallOfFamer=true")
				}
			case pB:
				if r.IsHallOfFamer {
					t.Errorf("player B: expected IsHallOfFamer=false")
				}
			}
		}
	})

	t.Run("awards scoped to seasons with target team", func(t *testing.T) {
		s3 := seedSeason(t, db, 3, 3, 40)
		h2s3 := seedTeamHistory(t, db, t2, s3, "Team Two", "W", "AL", 75, 67)
		psA3 := seedPlayerSeason(t, db, pA, s3, &h2s3)
		seedBatting(t, db, psA3, true, 490, 145, 22, 82)
		// Award earned with team 1 (s1)
		linkTopPlayerAward(t, db, psA1, "All-Star")
		// Award earned with team 2 (s3) — must not appear in team 1 results
		linkTopPlayerAward(t, db, psA3, "MVP")

		results, err := tq.GetTeamTopPlayers(ctx, t1, 25)
		if err != nil {
			t.Fatal(err)
		}
		var playerA *models.TeamTopPlayer
		for i := range results {
			if results[i].PlayerID == pA {
				playerA = &results[i]
			}
		}
		if playerA == nil {
			t.Fatal("player A not found in results")
			return
		}
		for _, aw := range playerA.Awards {
			if aw == "MVP" {
				t.Errorf("award 'MVP' earned on team 2 should not appear in team 1 results")
			}
		}
		found := false
		for _, aw := range playerA.Awards {
			if aw == "All-Star" {
				found = true
			}
		}
		if !found {
			t.Errorf("award 'All-Star' earned on team 1 should appear in results; got %v", playerA.Awards)
		}
	})

	t.Run("traded player appears for both teams", func(t *testing.T) {
		// Player E has a player_season linked to both teams in s1 (traded mid-season).
		pE := seedPlayer(t, db, "tp-pe", "Eli", "Echo")
		// seedPlayerSeason inserts with sort_order=0 → team 1 is final team
		psE1 := seedPlayerSeason(t, db, pE, s1, &h1s1)
		// Add second association: team 2 (the team player came from, sort_order=1)
		addTopPlayerTeam(t, db, psE1, h2s1, 1)
		seedBatting(t, db, psE1, true, 450, 130, 15, 70)
		setTopPlayerBattingWAR(t, db, psE1, 110.0, 2.0)

		team1Results, err := tq.GetTeamTopPlayers(ctx, t1, 25)
		if err != nil {
			t.Fatal(err)
		}
		team2Results, err := tq.GetTeamTopPlayers(ctx, t2, 25)
		if err != nil {
			t.Fatal(err)
		}

		foundInT1, foundInT2 := false, false
		for _, r := range team1Results {
			if r.PlayerID == pE {
				foundInT1 = true
			}
		}
		for _, r := range team2Results {
			if r.PlayerID == pE {
				foundInT2 = true
			}
		}
		if !foundInT1 {
			t.Errorf("traded player E should appear in team 1 results (final team)")
		}
		if !foundInT2 {
			t.Errorf("traded player E should appear in team 2 results (traded from)")
		}
	})
}

// ── Helpers for TestGetTeamTopPlayers ─────────────────────────────────────────

func setTopPlayerBattingWAR(t *testing.T, db *sql.DB, playerSeasonID int64, opsPlus, smbWAR float64) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		`UPDATE player_season_batting_stats SET ops_plus=?, smb_war=? WHERE player_season_id=? AND is_regular_season=1`,
		opsPlus, smbWAR, playerSeasonID)
	if err != nil {
		t.Fatalf("setTopPlayerBattingWAR: %v", err)
	}
}

func setTopPlayerPitchingWAR(t *testing.T, db *sql.DB, playerSeasonID int64, eraPlus, smbWAR float64) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		`UPDATE player_season_pitching_stats SET era_plus=?, smb_war=? WHERE player_season_id=? AND is_regular_season=1`,
		eraPlus, smbWAR, playerSeasonID)
	if err != nil {
		t.Fatalf("setTopPlayerPitchingWAR: %v", err)
	}
}

func linkTopPlayerAward(t *testing.T, db *sql.DB, playerSeasonID int64, awardName string) {
	t.Helper()
	var awardID int64
	err := db.QueryRowContext(context.Background(),
		`SELECT id FROM awards WHERE original_name=? LIMIT 1`, awardName).Scan(&awardID)
	if err != nil {
		t.Fatalf("linkTopPlayerAward: looking up %q: %v", awardName, err)
	}
	_, err = db.ExecContext(context.Background(),
		`INSERT OR IGNORE INTO player_season_awards (player_season_id, award_id) VALUES (?,?)`,
		playerSeasonID, awardID)
	if err != nil {
		t.Fatalf("linkTopPlayerAward: inserting %q: %v", awardName, err)
	}
}

func addTopPlayerTeam(t *testing.T, db *sql.DB, playerSeasonID, teamHistID int64, sortOrder int) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		`INSERT OR IGNORE INTO player_season_teams (player_season_id, team_history_id, sort_order) VALUES (?,?,?)`,
		playerSeasonID, teamHistID, sortOrder)
	if err != nil {
		t.Fatalf("addTopPlayerTeam: %v", err)
	}
}

func abs64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func insertScheduleGame(t *testing.T, db *sql.DB, seasonID, gameNum int, homeHistID, awayHistID int64, homeScore, awayScore int) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
INSERT INTO team_season_schedules
    (season_id, game_number, day, home_team_history_id, away_team_history_id,
     home_score, away_score)
VALUES (?,?,1,?,?,?,?)
`, seasonID, gameNum, homeHistID, awayHistID, homeScore, awayScore)
	if err != nil {
		t.Fatalf("insertScheduleGame: %v", err)
	}
}
