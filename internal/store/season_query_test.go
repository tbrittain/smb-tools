package store_test

import (
	"context"
	"database/sql"
	"testing"

	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

// ── shared seed helpers ───────────────────────────────────────────────────────

// seedSeason inserts a season row using save_game_season_id as the game-side key
// and returns the companion DB autoincrement id.
func seedSeason(t *testing.T, db *sql.DB, sgID, num, numGames int) int64 {
	t.Helper()
	res, err := db.ExecContext(context.Background(),
		`INSERT INTO seasons (league_guid, save_game_season_id, season_num, num_games) VALUES ('TESTLEAGUE',?,?,?)`,
		sgID, num, numGames)
	if err != nil {
		t.Fatalf("seedSeason: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func seedTeam(t *testing.T, db *sql.DB, guid string) int64 {
	t.Helper()
	res, err := db.ExecContext(context.Background(),
		`INSERT INTO teams (game_guid) VALUES (?)`, guid)
	if err != nil {
		t.Fatalf("seedTeam: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func seedTeamHistory(t *testing.T, db *sql.DB, teamID int64, seasonID int64, name, div, conf string, wins, losses int) int64 {
	t.Helper()
	res, err := db.ExecContext(context.Background(), `
INSERT INTO team_season_history
    (team_id, season_id, team_name, division_name, conference_name, wins, losses,
     games_back, runs_for, runs_against,
     total_power, total_contact, total_speed, total_fielding, total_arm,
     total_velocity, total_junk, total_accuracy)
VALUES (?,?,?,?,?,?,?,0,0,0,0,0,0,0,0,0,0,0)
`, teamID, seasonID, name, div, conf, wins, losses)
	if err != nil {
		t.Fatalf("seedTeamHistory: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func seedPlayoffGame(t *testing.T, db *sql.DB, seasonID int64, seriesNum, gameNum int, homeHistID, awayHistID int64, homeScore, awayScore int) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
INSERT INTO team_playoff_schedules
    (season_id, series_number, game_number, home_team_history_id, away_team_history_id,
     home_score, away_score)
VALUES (?,?,?,?,?,?,?)
`, seasonID, seriesNum, gameNum, homeHistID, awayHistID, homeScore, awayScore)
	if err != nil {
		t.Fatalf("seedPlayoffGame: %v", err)
	}
}

func seedPlayer(t *testing.T, db *sql.DB, guid, first, last string) int64 {
	t.Helper()
	res, err := db.ExecContext(context.Background(),
		`INSERT INTO players (game_guid, first_name, last_name) VALUES (?,?,?)`,
		guid, first, last)
	if err != nil {
		t.Fatalf("seedPlayer: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func seedPlayerSeason(t *testing.T, db *sql.DB, playerID int64, seasonID int64, teamHistID *int64) int64 {
	t.Helper()
	res, err := db.ExecContext(context.Background(), `
INSERT INTO player_seasons
    (player_id, season_id, age, salary,
     primary_position, secondary_position, pitcher_role,
     bat_hand, throw_hand, chemistry_type, traits_json, pitches_json)
VALUES (?,?,25,1000,'SS','','','R','R','','[]','[]')
`, playerID, seasonID)
	if err != nil {
		t.Fatalf("seedPlayerSeason: %v", err)
	}
	id, _ := res.LastInsertId()
	if teamHistID != nil {
		_, err = db.ExecContext(context.Background(), `
INSERT OR IGNORE INTO player_season_teams (player_season_id, team_history_id, sort_order)
VALUES (?, ?, 0)
`, id, *teamHistID)
		if err != nil {
			t.Fatalf("seedPlayerSeason team: %v", err)
		}
	}
	return id
}

func seedBatting(t *testing.T, db *sql.DB, playerSeasonID int64, isReg bool, ab, hits, hr, rbi int) {
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
VALUES (?,?,?,?,?,0,?,0,0,?,?,0,0,0,0,0,0,0,0,0)
`, playerSeasonID, isRegInt, ab, ab, ab, hits, hr, rbi)
	if err != nil {
		t.Fatalf("seedBatting: %v", err)
	}
}

func seedPitching(t *testing.T, db *sql.DB, playerSeasonID int64, isReg bool, w, l, outs, er, k int) {
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
VALUES (?,?,?,?,?,?,0,0,0,?,0,?,0,0,?,0,?,0,0,0,0)
`, playerSeasonID, isRegInt, w, l, w+l, w+l, outs, er, k, w+l)
	if err != nil {
		t.Fatalf("seedPitching: %v", err)
	}
}

// ── SeasonQueryStore tests ────────────────────────────────────────────────────

func seedPlayoffGameNullScore(t *testing.T, db *sql.DB, seasonID int64, seriesNum, gameNum int, homeHistID, awayHistID int64) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
INSERT INTO team_playoff_schedules
    (season_id, series_number, game_number, home_team_history_id, away_team_history_id,
     home_score, away_score)
VALUES (?,?,?,?,?,NULL,NULL)
`, seasonID, seriesNum, gameNum, homeHistID, awayHistID)
	if err != nil {
		t.Fatalf("seedPlayoffGameNullScore: %v", err)
	}
}

func TestListWithChampion_NoPlayoffs(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)
	teamID := seedTeam(t, db, "team-guid-a")
	seedTeamHistory(t, db, teamID, s1, "Team A", "East", "American", 30, 10)

	store := store.NewSeasonQueryStore(db)
	seasons, err := store.ListWithChampion(ctx)
	if err != nil {
		t.Fatalf("ListWithChampion: %v", err)
	}
	if len(seasons) != 1 {
		t.Fatalf("expected 1 season, got %d", len(seasons))
	}
	if seasons[0].ChampionTeamName != "" {
		t.Errorf("expected empty champion, got %q", seasons[0].ChampionTeamName)
	}
	if seasons[0].ChampionHistoryID != nil {
		t.Errorf("expected nil champion history ID, got %v", *seasons[0].ChampionHistoryID)
	}
}

func TestListWithChampion_WithPlayoffs(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)
	homeTeamID := seedTeam(t, db, "team-home")
	awayTeamID := seedTeam(t, db, "team-away")
	homeHistID := seedTeamHistory(t, db, homeTeamID, s1, "Home Team", "West", "National", 35, 5)
	awayHistID := seedTeamHistory(t, db, awayTeamID, s1, "Away Team", "East", "National", 25, 15)

	// Final series (series 2): home wins 3 games, away wins 1
	seedPlayoffGame(t, db, s1, 2, 1, homeHistID, awayHistID, 5, 2)
	seedPlayoffGame(t, db, s1, 2, 2, homeHistID, awayHistID, 4, 3)
	seedPlayoffGame(t, db, s1, 2, 3, awayHistID, homeHistID, 3, 1)
	seedPlayoffGame(t, db, s1, 2, 4, homeHistID, awayHistID, 2, 1)

	store := store.NewSeasonQueryStore(db)
	seasons, err := store.ListWithChampion(ctx)
	if err != nil {
		t.Fatalf("ListWithChampion: %v", err)
	}
	if len(seasons) != 1 {
		t.Fatalf("expected 1 season, got %d", len(seasons))
	}
	if seasons[0].ChampionTeamName != "Home Team" {
		t.Errorf("expected champion 'Home Team', got %q", seasons[0].ChampionTeamName)
	}
	if seasons[0].ChampionHistoryID == nil {
		t.Fatal("expected non-nil champion history ID")
	}
	if *seasons[0].ChampionHistoryID != homeHistID {
		t.Errorf("champion history ID: want %d, got %d", homeHistID, *seasons[0].ChampionHistoryID)
	}
}

func TestListWithChampion_PartialPlayoffs_NoChampion(t *testing.T) {
	// If any playoff game has a NULL score (mid-playoffs import), no champion
	// should be returned — even if one team leads the final series.
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)
	homeTeamID := seedTeam(t, db, "team-home-partial")
	awayTeamID := seedTeam(t, db, "team-away-partial")
	homeHistID := seedTeamHistory(t, db, homeTeamID, s1, "Home Partial", "W", "NL", 90, 52)
	awayHistID := seedTeamHistory(t, db, awayTeamID, s1, "Away Partial", "E", "NL", 75, 67)

	seedPlayoffGame(t, db, s1, 2, 1, homeHistID, awayHistID, 4, 1)
	seedPlayoffGame(t, db, s1, 2, 2, homeHistID, awayHistID, 3, 2)
	seedPlayoffGame(t, db, s1, 2, 3, awayHistID, homeHistID, 5, 2)
	seedPlayoffGame(t, db, s1, 2, 4, homeHistID, awayHistID, 2, 1)
	seedPlayoffGameNullScore(t, db, s1, 2, 5, awayHistID, homeHistID)

	sq := store.NewSeasonQueryStore(db)
	seasons, err := sq.ListWithChampion(ctx)
	if err != nil {
		t.Fatalf("ListWithChampion: %v", err)
	}
	if len(seasons) != 1 {
		t.Fatalf("expected 1 season, got %d", len(seasons))
	}
	if seasons[0].ChampionTeamName != "" {
		t.Errorf("partial playoffs: expected no champion, got %q", seasons[0].ChampionTeamName)
	}
	if seasons[0].ChampionHistoryID != nil {
		t.Errorf("partial playoffs: expected nil champion ID, got %v", *seasons[0].ChampionHistoryID)
	}
}

func TestListWithChampion_OrderedBySeasonNum(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 10, 3, 40)
	seedSeason(t, db, 5, 1, 40)
	seedSeason(t, db, 7, 2, 40)
	_ = struct{}{} // IDs not needed — only ordering is tested

	sq := store.NewSeasonQueryStore(db)
	seasons, err := sq.ListWithChampion(ctx)
	if err != nil {
		t.Fatalf("ListWithChampion: %v", err)
	}
	if len(seasons) != 3 {
		t.Fatalf("expected 3 seasons, got %d", len(seasons))
	}
	if seasons[0].SeasonNum != 1 || seasons[1].SeasonNum != 2 || seasons[2].SeasonNum != 3 {
		t.Errorf("unexpected order: %v, %v, %v", seasons[0].SeasonNum, seasons[1].SeasonNum, seasons[2].SeasonNum)
	}
}

func TestGetStandings(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "t1")
	t2 := seedTeam(t, db, "t2")
	seedTeamHistory(t, db, t1, s1, "Alpha", "East", "AL", 30, 10)
	seedTeamHistory(t, db, t2, s1, "Beta", "East", "AL", 20, 20)

	sq := store.NewSeasonQueryStore(db)
	standings, err := sq.GetStandings(ctx, s1)
	if err != nil {
		t.Fatalf("GetStandings: %v", err)
	}
	if len(standings) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(standings))
	}
	// Alpha has more wins, should be first
	if standings[0].TeamName != "Alpha" {
		t.Errorf("expected Alpha first, got %q", standings[0].TeamName)
	}
	// WinPct computed correctly
	want := 30.0 / 40.0
	if standings[0].WinPct != want {
		t.Errorf("WinPct: want %.4f, got %.4f", want, standings[0].WinPct)
	}
}

func TestGetSeasonStatLeaders_BALeader(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	// 40-game season → BA threshold is 40*3 = 120 AB
	s1 := seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "t1")
	hist1 := seedTeamHistory(t, db, t1, s1, "Team A", "E", "AL", 20, 20)

	pHigh := seedPlayer(t, db, "phigh", "High", "BA")
	pLow := seedPlayer(t, db, "plow", "Low", "BA")
	pDQ := seedPlayer(t, db, "pdq", "DQ", "Short") // disqualified — too few AB

	psHigh := seedPlayerSeason(t, db, pHigh, s1, &hist1)
	psLow := seedPlayerSeason(t, db, pLow, s1, &hist1)
	psDQ := seedPlayerSeason(t, db, pDQ, s1, &hist1)

	// High: 150 AB, 50 H → BA .333
	seedBatting(t, db, psHigh, true, 150, 50, 0, 0)
	// Low: 150 AB, 40 H → BA .267
	seedBatting(t, db, psLow, true, 150, 40, 0, 0)
	// DQ: 10 AB, 9 H → BA .900 but below threshold
	seedBatting(t, db, psDQ, true, 10, 9, 0, 0)

	sq := store.NewSeasonQueryStore(db)
	leaders, err := sq.GetSeasonStatLeaders(ctx, s1)
	if err != nil {
		t.Fatalf("GetSeasonStatLeaders: %v", err)
	}
	if leaders.BA == nil {
		t.Fatal("BA leader is nil")
	}
	if leaders.BA.LastName != "BA" {
		t.Errorf("BA leader: want 'BA', got %q", leaders.BA.LastName)
	}
}

func TestGetSeasonStatLeaders_NoData(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)

	sq := store.NewSeasonQueryStore(db)
	leaders, err := sq.GetSeasonStatLeaders(ctx, s1)
	if err != nil {
		t.Fatalf("GetSeasonStatLeaders on empty season: %v", err)
	}
	if leaders.BA != nil || leaders.HR != nil || leaders.ERA != nil {
		t.Errorf("expected all nil leaders for empty season")
	}
}

func TestGetCareerLeaders(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "t1")
	hist1 := seedTeamHistory(t, db, t1, s1, "Team", "E", "AL", 20, 20)

	for i, hr := range []int{40, 30, 20} {
		guid := string(rune('a' + i))
		pid := seedPlayer(t, db, "guid-"+guid, "Player", guid)
		psid := seedPlayerSeason(t, db, pid, s1, &hist1)
		seedBatting(t, db, psid, true, 500, 150, hr, 100)
	}

	sq := store.NewSeasonQueryStore(db)
	leaders, err := sq.GetCareerLeaders(ctx)
	if err != nil {
		t.Fatalf("GetCareerLeaders: %v", err)
	}
	if len(leaders.HR) == 0 {
		t.Fatal("HR leaders empty")
	}
	if leaders.HR[0].StatValue != 40 {
		t.Errorf("top HR: want 40, got %.0f", leaders.HR[0].StatValue)
	}
}
