package store_test

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"testing"

	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

// ── Validation (errors from buildExportQuery surface through PreviewExportData) ──

func TestPreviewExportData_RejectsUnknownColumn(t *testing.T) {
	s := store.NewExportStore(testutil.NewTestDB(t))
	_, err := s.PreviewExportData(context.Background(), store.ExportOptions{
		DatasetID: "batting_season",
		Columns:   []string{"no_such_column"},
	})
	if err == nil {
		t.Fatal("expected error for unknown column key, got nil")
	}
}

func TestPreviewExportData_RejectsEmptyColumns(t *testing.T) {
	s := store.NewExportStore(testutil.NewTestDB(t))
	_, err := s.PreviewExportData(context.Background(), store.ExportOptions{
		DatasetID: "batting_season",
		Columns:   []string{},
	})
	if err == nil {
		t.Fatal("expected error for empty column list, got nil")
	}
}

func TestPreviewExportData_RejectsUnknownSortField(t *testing.T) {
	s := store.NewExportStore(testutil.NewTestDB(t))
	_, err := s.PreviewExportData(context.Background(), store.ExportOptions{
		DatasetID: "batting_season",
		Columns:   []string{"player_name"},
		SortCol:   "not_a_column",
	})
	if err == nil {
		t.Fatal("expected error for unknown sort column, got nil")
	}
}

func TestPreviewExportData_RejectsUnknownFilterOp(t *testing.T) {
	s := store.NewExportStore(testutil.NewTestDB(t))
	// Column must be valid so the filter-op check is reached (unknown column is silently skipped).
	_, err := s.PreviewExportData(context.Background(), store.ExportOptions{
		DatasetID: "batting_season",
		Columns:   []string{"player_name"},
		Filters:   []store.FilterRow{{Column: "season_num", Op: "BETWEEN", Value: "1"}},
	})
	if err == nil {
		t.Fatal("expected error for unknown filter op, got nil")
	}
}

func TestPreviewExportData_RejectsUnknownDataset(t *testing.T) {
	s := store.NewExportStore(testutil.NewTestDB(t))
	_, err := s.PreviewExportData(context.Background(), store.ExportOptions{
		DatasetID: "no_such_dataset",
		Columns:   []string{"player_name"},
	})
	if err == nil {
		t.Fatal("expected error for unknown dataset, got nil")
	}
}

// ── Dataset smoke tests ───────────────────────────────────────────────────────

func TestPreviewExportData_BattingSeason(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)
	teamID := seedTeam(t, db, "team-bat")
	h1 := seedTeamHistory(t, db, teamID, s1, "Alphas", "East", "NL", 30, 10)
	p := seedPlayer(t, db, "guid-bat", "Alice", "Alpha")
	ps := seedPlayerSeason(t, db, p, s1, &h1)
	seedBatting(t, db, ps, true, 400, 120, 20, 80)

	preview, err := store.NewExportStore(db).PreviewExportData(ctx, store.ExportOptions{
		DatasetID: "batting_season",
		Columns:   []string{"player_name", "season_num", "home_runs"},
	})
	if err != nil {
		t.Fatalf("PreviewExportData batting_season: %v", err)
	}
	if preview.TotalCount != 1 {
		t.Errorf("TotalCount: want 1, got %d", preview.TotalCount)
	}
	if len(preview.Rows) != 1 {
		t.Fatalf("rows: want 1, got %d", len(preview.Rows))
	}
	if hr, _ := preview.Rows[0]["home_runs"].(int64); hr != 20 {
		t.Errorf("home_runs: want 20, got %v", preview.Rows[0]["home_runs"])
	}
}

func TestPreviewExportData_PitchingSeason(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)
	teamID := seedTeam(t, db, "team-pit")
	h1 := seedTeamHistory(t, db, teamID, s1, "Pitchers", "West", "AL", 20, 20)
	p := seedPlayer(t, db, "guid-pit", "Bob", "Beta")
	ps := seedPlayerSeason(t, db, p, s1, &h1)
	seedPitching(t, db, ps, true, 15, 5, 200, 30, 120)

	preview, err := store.NewExportStore(db).PreviewExportData(ctx, store.ExportOptions{
		DatasetID: "pitching_season",
		Columns:   []string{"player_name", "wins", "strikeouts"},
	})
	if err != nil {
		t.Fatalf("PreviewExportData pitching_season: %v", err)
	}
	if preview.TotalCount != 1 {
		t.Errorf("TotalCount: want 1, got %d", preview.TotalCount)
	}
	if w, _ := preview.Rows[0]["wins"].(int64); w != 15 {
		t.Errorf("wins: want 15, got %v", preview.Rows[0]["wins"])
	}
}

func TestPreviewExportData_Standings(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)
	teamID := seedTeam(t, db, "team-std")
	seedTeamHistory(t, db, teamID, s1, "Stand Team", "East", "NL", 30, 10)

	preview, err := store.NewExportStore(db).PreviewExportData(ctx, store.ExportOptions{
		DatasetID: "standings",
		Columns:   []string{"team_name", "season_num", "wins", "losses"},
	})
	if err != nil {
		t.Fatalf("PreviewExportData standings: %v", err)
	}
	if preview.TotalCount != 1 {
		t.Errorf("TotalCount: want 1, got %d", preview.TotalCount)
	}
	if name, _ := preview.Rows[0]["team_name"].(string); name != "Stand Team" {
		t.Errorf("team_name: want %q, got %v", "Stand Team", preview.Rows[0]["team_name"])
	}
}

func TestPreviewExportData_CareerBatting(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	p := seedPlayer(t, db, "guid-cb", "Carol", "Career")
	_, err := db.ExecContext(ctx, `
INSERT INTO player_career_batting_stats
    (player_id, stat_type, seasons_played, games_played, games_batting, at_bats,
     runs, hits, doubles, triples, home_runs, rbi, stolen_bases, caught_stealing,
     walks, strikeouts, hit_by_pitch, sac_hits, sac_flies, errors, passed_balls)
VALUES (?,?,2,100,100,350,0,100,0,0,30,90,0,0,0,0,0,0,0,0,0)
`, p, "regular_season")
	if err != nil {
		t.Fatalf("seed career batting: %v", err)
	}

	preview, err := store.NewExportStore(db).PreviewExportData(ctx, store.ExportOptions{
		DatasetID:      "career_batting",
		Columns:        []string{"player_name", "home_runs", "seasons_played"},
		CareerStatType: "regular_season",
	})
	if err != nil {
		t.Fatalf("PreviewExportData career_batting: %v", err)
	}
	if preview.TotalCount != 1 {
		t.Errorf("TotalCount: want 1, got %d", preview.TotalCount)
	}
	if hr, _ := preview.Rows[0]["home_runs"].(int64); hr != 30 {
		t.Errorf("home_runs: want 30, got %v", preview.Rows[0]["home_runs"])
	}
}

func TestPreviewExportData_CareerPitching(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	p := seedPlayer(t, db, "guid-cp", "Dave", "Decker")
	_, err := db.ExecContext(ctx, `
INSERT INTO player_career_pitching_stats
    (player_id, stat_type, seasons_played, wins, losses, games, games_started,
     complete_games, shutouts, saves, outs_pitched, hits_allowed, earned_runs,
     home_runs_allowed, walks, strikeouts, hit_batters, batters_faced,
     games_finished, runs_allowed, wild_pitches, total_pitches)
VALUES (?,?,3,50,20,70,68,5,2,0,540,200,80,10,60,300,5,270,2,90,3,1800)
`, p, "regular_season")
	if err != nil {
		t.Fatalf("seed career pitching: %v", err)
	}

	preview, err := store.NewExportStore(db).PreviewExportData(ctx, store.ExportOptions{
		DatasetID:      "career_pitching",
		Columns:        []string{"player_name", "wins", "strikeouts"},
		CareerStatType: "regular_season",
	})
	if err != nil {
		t.Fatalf("PreviewExportData career_pitching: %v", err)
	}
	if preview.TotalCount != 1 {
		t.Errorf("TotalCount: want 1, got %d", preview.TotalCount)
	}
	if w, _ := preview.Rows[0]["wins"].(int64); w != 50 {
		t.Errorf("wins: want 50, got %v", preview.Rows[0]["wins"])
	}
}

// ── Filter tests ──────────────────────────────────────────────────────────────

func TestPreviewExportData_SeasonRangeFilter(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)
	s2 := seedSeason(t, db, 2, 2, 40)
	s3 := seedSeason(t, db, 3, 3, 40)
	teamID := seedTeam(t, db, "team-srf")
	h1 := seedTeamHistory(t, db, teamID, s1, "RF Team", "East", "NL", 20, 20)
	h2 := seedTeamHistory(t, db, teamID, s2, "RF Team", "East", "NL", 20, 20)
	h3 := seedTeamHistory(t, db, teamID, s3, "RF Team", "East", "NL", 20, 20)
	p := seedPlayer(t, db, "guid-srf", "Range", "Filter")
	ps1 := seedPlayerSeason(t, db, p, s1, &h1)
	ps2 := seedPlayerSeason(t, db, p, s2, &h2)
	ps3 := seedPlayerSeason(t, db, p, s3, &h3)
	seedBatting(t, db, ps1, true, 200, 60, 5, 20)
	seedBatting(t, db, ps2, true, 200, 60, 5, 20)
	seedBatting(t, db, ps3, true, 200, 60, 5, 20)

	preview, err := store.NewExportStore(db).PreviewExportData(ctx, store.ExportOptions{
		DatasetID: "batting_season",
		Columns:   []string{"player_name", "season_num"},
		Filters: []store.FilterRow{
			{Column: "season_num", Op: "gte", Value: "2"},
			{Column: "season_num", Op: "lte", Value: "2"},
		},
	})
	if err != nil {
		t.Fatalf("PreviewExportData season filter: %v", err)
	}
	if preview.TotalCount != 1 {
		t.Errorf("TotalCount: want 1 (season 2 only), got %d", preview.TotalCount)
	}
	if num, _ := preview.Rows[0]["season_num"].(int64); num != 2 {
		t.Errorf("season_num: want 2, got %v", preview.Rows[0]["season_num"])
	}
}

func TestPreviewExportData_TeamFilter(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)
	teamA := seedTeam(t, db, "team-tf-a")
	teamB := seedTeam(t, db, "team-tf-b")
	hA := seedTeamHistory(t, db, teamA, s1, "Alpha Squad", "East", "NL", 25, 15)
	hB := seedTeamHistory(t, db, teamB, s1, "Beta Boys", "West", "NL", 15, 25)
	pA := seedPlayer(t, db, "guid-tfa", "Team", "Alpha")
	pB := seedPlayer(t, db, "guid-tfb", "Team", "Beta")
	psA := seedPlayerSeason(t, db, pA, s1, &hA)
	psB := seedPlayerSeason(t, db, pB, s1, &hB)
	seedBatting(t, db, psA, true, 200, 60, 5, 20)
	seedBatting(t, db, psB, true, 200, 60, 3, 15)

	preview, err := store.NewExportStore(db).PreviewExportData(ctx, store.ExportOptions{
		DatasetID: "batting_season",
		Columns:   []string{"player_name", "team_name"},
		Filters:   []store.FilterRow{{Column: "team_name", Op: "eq", Value: "Alpha Squad"}},
	})
	if err != nil {
		t.Fatalf("PreviewExportData team filter: %v", err)
	}
	if preview.TotalCount != 1 {
		t.Errorf("TotalCount: want 1 (Alpha Squad only), got %d", preview.TotalCount)
	}
	if name, _ := preview.Rows[0]["team_name"].(string); name != "Alpha Squad" {
		t.Errorf("team_name: want %q, got %v", "Alpha Squad", preview.Rows[0]["team_name"])
	}
}

func TestPreviewExportData_CareerStatTypeFilter(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	p := seedPlayer(t, db, "guid-stf", "Stat", "Type")
	statHR := map[string]int{"regular_season": 30, "playoffs": 5, "total_career": 35}
	for _, st := range []string{"regular_season", "playoffs", "total_career"} {
		_, err := db.ExecContext(ctx, `
INSERT INTO player_career_batting_stats
    (player_id, stat_type, seasons_played, games_played, games_batting, at_bats,
     runs, hits, doubles, triples, home_runs, rbi, stolen_bases, caught_stealing,
     walks, strikeouts, hit_by_pitch, sac_hits, sac_flies, errors, passed_balls)
VALUES (?,?,2,100,100,350,0,100,0,0,?,90,0,0,0,0,0,0,0,0,0)
`, p, st, statHR[st])
		if err != nil {
			t.Fatalf("seed career batting %s: %v", st, err)
		}
	}

	preview, err := store.NewExportStore(db).PreviewExportData(ctx, store.ExportOptions{
		DatasetID:      "career_batting",
		Columns:        []string{"player_name", "home_runs"},
		CareerStatType: "playoffs",
	})
	if err != nil {
		t.Fatalf("PreviewExportData career stat type: %v", err)
	}
	if preview.TotalCount != 1 {
		t.Fatalf("TotalCount: want 1 (playoffs only), got %d", preview.TotalCount)
	}
	if hr, _ := preview.Rows[0]["home_runs"].(int64); hr != 5 {
		t.Errorf("home_runs (playoffs): want 5, got %v", preview.Rows[0]["home_runs"])
	}
}

func TestPreviewExportData_CareerStatTypeDefaultsToRegularSeason(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	p := seedPlayer(t, db, "guid-def", "Default", "Season")
	for st, hr := range map[string]int{"regular_season": 40, "playoffs": 2} {
		_, err := db.ExecContext(ctx, `
INSERT INTO player_career_batting_stats
    (player_id, stat_type, seasons_played, games_played, games_batting, at_bats,
     runs, hits, doubles, triples, home_runs, rbi, stolen_bases, caught_stealing,
     walks, strikeouts, hit_by_pitch, sac_hits, sac_flies, errors, passed_balls)
VALUES (?,?,2,100,100,350,0,100,0,0,?,90,0,0,0,0,0,0,0,0,0)
`, p, st, hr)
		if err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	// Empty CareerStatType should default to regular_season.
	preview, err := store.NewExportStore(db).PreviewExportData(ctx, store.ExportOptions{
		DatasetID:      "career_batting",
		Columns:        []string{"player_name", "home_runs"},
		CareerStatType: "",
	})
	if err != nil {
		t.Fatalf("PreviewExportData default stat type: %v", err)
	}
	if preview.TotalCount != 1 {
		t.Fatalf("TotalCount: want 1, got %d", preview.TotalCount)
	}
	if hr, _ := preview.Rows[0]["home_runs"].(int64); hr != 40 {
		t.Errorf("home_runs (default→regular_season): want 40, got %v", preview.Rows[0]["home_runs"])
	}
}

func TestPreviewExportData_LimitRespected(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	// Career batting has minimal join requirements (player + career row only),
	// making it practical to insert 501 rows for the limit test.
	const total = 501
	for i := range total {
		p := seedPlayer(t, db, fmt.Sprintf("guid-lim-%d", i), fmt.Sprintf("P%d", i), "Last")
		_, err := db.ExecContext(ctx, `
INSERT INTO player_career_batting_stats
    (player_id, stat_type, seasons_played, games_played, games_batting, at_bats,
     runs, hits, doubles, triples, home_runs, rbi, stolen_bases, caught_stealing,
     walks, strikeouts, hit_by_pitch, sac_hits, sac_flies, errors, passed_balls)
VALUES (?,?,1,50,50,200,0,60,0,0,10,40,0,0,0,0,0,0,0,0,0)
`, p, "regular_season")
		if err != nil {
			t.Fatalf("seed player %d: %v", i, err)
		}
	}

	preview, err := store.NewExportStore(db).PreviewExportData(ctx, store.ExportOptions{
		DatasetID:      "career_batting",
		Columns:        []string{"player_name", "home_runs"},
		CareerStatType: "regular_season",
	})
	if err != nil {
		t.Fatalf("PreviewExportData limit: %v", err)
	}
	if preview.TotalCount != total {
		t.Errorf("TotalCount: want %d, got %d", total, preview.TotalCount)
	}
	if len(preview.Rows) > 500 {
		t.Errorf("preview rows: want ≤500, got %d", len(preview.Rows))
	}
	if preview.TotalCount <= len(preview.Rows) {
		t.Errorf("TotalCount (%d) should exceed preview row count (%d)", preview.TotalCount, len(preview.Rows))
	}
}

// ── CSV export tests ──────────────────────────────────────────────────────────

func TestExportToCSV_HeaderRow(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)
	teamID := seedTeam(t, db, "team-csv")
	h1 := seedTeamHistory(t, db, teamID, s1, "CSV Team", "East", "NL", 20, 20)
	p := seedPlayer(t, db, "guid-csv", "CSV", "Hero")
	ps := seedPlayerSeason(t, db, p, s1, &h1)
	seedBatting(t, db, ps, true, 300, 90, 15, 50)

	data, err := store.NewExportStore(db).ExportToCSV(ctx, store.ExportOptions{
		DatasetID: "batting_season",
		Columns:   []string{"player_name", "season_num", "home_runs"},
	})
	if err != nil {
		t.Fatalf("ExportToCSV: %v", err)
	}

	records, err := csv.NewReader(strings.NewReader(string(data))).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	if len(records) < 1 {
		t.Fatal("CSV has no rows")
	}

	// First row must be the column labels (not keys).
	want := []string{"Player", "Season", "HR"}
	if len(records[0]) != len(want) {
		t.Fatalf("header length: want %d cols, got %d", len(want), len(records[0]))
	}
	for i, h := range want {
		if records[0][i] != h {
			t.Errorf("header[%d]: want %q, got %q", i, h, records[0][i])
		}
	}
}

func TestExportToCSV_RowCount(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)
	teamID := seedTeam(t, db, "team-rc")
	h1 := seedTeamHistory(t, db, teamID, s1, "Row Count", "East", "NL", 20, 20)

	const numPlayers = 5
	for i := range numPlayers {
		p := seedPlayer(t, db, fmt.Sprintf("guid-rc-%d", i), fmt.Sprintf("P%d", i), "Last")
		ps := seedPlayerSeason(t, db, p, s1, &h1)
		seedBatting(t, db, ps, true, 200, 60, 5, 20)
	}

	data, err := store.NewExportStore(db).ExportToCSV(ctx, store.ExportOptions{
		DatasetID: "batting_season",
		Columns:   []string{"player_name", "home_runs"},
	})
	if err != nil {
		t.Fatalf("ExportToCSV: %v", err)
	}

	records, err := csv.NewReader(strings.NewReader(string(data))).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	// First row is the header; rest are data.
	if got := len(records) - 1; got != numPlayers {
		t.Errorf("data rows: want %d, got %d", numPlayers, got)
	}
}
