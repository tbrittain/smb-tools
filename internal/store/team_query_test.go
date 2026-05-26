package store_test

import (
	"context"
	"database/sql"
	"testing"

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
	seedPlayoffGame(t, db, 1, 2, 3, rivalHist1, hist1, 5, 0)  // rival wins at home
	seedPlayoffGame(t, db, 1, 2, 4, hist1, rivalHist1, 6, 1)

	// Season 2: rival wins final series
	seedPlayoffGame(t, db, 2, 2, 1, rivalHist2, hist2, 3, 1)
	seedPlayoffGame(t, db, 2, 2, 2, rivalHist2, hist2, 2, 0)
	seedPlayoffGame(t, db, 2, 2, 3, rivalHist2, hist2, 4, 3)

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
	for i := 0; i < 3; i++ {
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
