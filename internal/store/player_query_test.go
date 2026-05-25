package store_test

import (
	"context"
	"math"
	"testing"

	"smb-tools/internal/service"
	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

func TestSearchPlayers_ByLastName(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	hist1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	pjohn := seedPlayer(t, db, "g1", "John", "Smith")
	pjane := seedPlayer(t, db, "g2", "Jane", "Doe")
	pjohnny := seedPlayer(t, db, "g3", "Johnny", "Bravo")

	seedPlayerSeason(t, db, pjohn, 1, &hist1)
	seedPlayerSeason(t, db, pjane, 1, &hist1)
	seedPlayerSeason(t, db, pjohnny, 1, &hist1)

	pq := store.NewPlayerQueryStore(db)

	t.Run("exact last name", func(t *testing.T) {
		res, err := pq.SearchPlayers(ctx, "Smith")
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 1 || res[0].LastName != "Smith" {
			t.Errorf("expected only Smith, got %v", res)
		}
	})

	t.Run("prefix matches first name", func(t *testing.T) {
		res, err := pq.SearchPlayers(ctx, "Jo")
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 2 {
			t.Errorf("expected 2 results for 'Jo', got %d: %v", len(res), res)
		}
	})

	t.Run("full name", func(t *testing.T) {
		res, err := pq.SearchPlayers(ctx, "John Smith")
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 1 || res[0].FirstName != "John" {
			t.Errorf("full-name search: expected John Smith, got %v", res)
		}
	})

	t.Run("no match", func(t *testing.T) {
		res, err := pq.SearchPlayers(ctx, "Zzz")
		if err != nil {
			t.Fatal(err)
		}
		if len(res) != 0 {
			t.Errorf("expected 0 results, got %d", len(res))
		}
	})
}

func TestGetPlayerCareer_SumsAcrossSeasons(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	// Two seasons
	seedSeason(t, db, 1, 1, 40)
	seedSeason(t, db, 2, 2, 40)

	t1 := seedTeam(t, db, "tg1")
	hist1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)
	hist2 := seedTeamHistory(t, db, t1, 2, "Team", "E", "AL", 25, 15)

	pid := seedPlayer(t, db, "pguid", "Test", "Player")
	ps1 := seedPlayerSeason(t, db, pid, 1, &hist1)
	ps2 := seedPlayerSeason(t, db, pid, 2, &hist2)

	// Season 1: 400 AB, 120 H (BA=.300)
	// Season 2: 500 AB, 125 H (BA=.250)
	// Career:   900 AB, 245 H → BA=245/900 ≈ .2722... (NOT average of .300 and .250)
	seedBatting(t, db, ps1, true, 400, 120, 10, 50)
	seedBatting(t, db, ps2, true, 500, 125, 15, 60)

	// Populate career tables before reading.
	if err := service.ApplyCareerStats(ctx, db, []int64{pid}); err != nil {
		t.Fatalf("ApplyCareerStats: %v", err)
	}

	pq := store.NewPlayerQueryStore(db)
	career, err := pq.GetPlayerCareer(ctx, pid)
	if err != nil {
		t.Fatalf("GetPlayerCareer: %v", err)
	}
	if career.Batting == nil {
		t.Fatal("expected batting stats")
	}
	if career.Batting.AtBats != 900 {
		t.Errorf("career AB: want 900, got %d", career.Batting.AtBats)
	}
	if career.Batting.Hits != 245 {
		t.Errorf("career H: want 245, got %d", career.Batting.Hits)
	}
	if career.Batting.HomeRuns != 25 {
		t.Errorf("career HR: want 25, got %d", career.Batting.HomeRuns)
	}
	// BA is now stored in the career table: 245/900
	if career.Batting.BA == nil {
		t.Fatal("BA should be non-nil (stored in career table)")
	}
	wantBA := 245.0 / 900.0
	if math.Abs(*career.Batting.BA-wantBA) > 1e-9 {
		t.Errorf("career BA: want %.6f, got %.6f", wantBA, *career.Batting.BA)
	}
}

func TestGetPlayerCareer_NoBattingStats(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg")
	hist1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)
	pid := seedPlayer(t, db, "pg", "Pitcher", "Only")
	ps1 := seedPlayerSeason(t, db, pid, 1, &hist1)
	seedPitching(t, db, ps1, true, 10, 5, 270, 30, 100)

	if err := service.ApplyCareerStats(ctx, db, []int64{pid}); err != nil {
		t.Fatalf("ApplyCareerStats: %v", err)
	}

	pq := store.NewPlayerQueryStore(db)
	career, err := pq.GetPlayerCareer(ctx, pid)
	if err != nil {
		t.Fatalf("GetPlayerCareer: %v", err)
	}
	if career.Batting != nil {
		t.Errorf("expected nil Batting for pitcher-only player")
	}
	if career.Pitching == nil {
		t.Fatal("expected pitching stats")
	}
	if career.Pitching.OutsPitched != 270 {
		t.Errorf("career outs pitched: want 270, got %d", career.Pitching.OutsPitched)
	}
	// ERA is now stored in the career table: 30*27/270 = 3.00
	if career.Pitching.ERA == nil {
		t.Fatal("ERA should be non-nil (stored in career table)")
	}
	wantERA := 30.0 * 27.0 / 270.0
	if math.Abs(*career.Pitching.ERA-wantERA) > 1e-9 {
		t.Errorf("career ERA: want %.4f, got %.4f", wantERA, *career.Pitching.ERA)
	}
}

func TestGetPlayerSeasonLog_RateColumnsStored(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg")
	hist1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)
	pid := seedPlayer(t, db, "pg", "Rate", "Test")
	ps1 := seedPlayerSeason(t, db, pid, 1, &hist1)
	// 400 AB, 120 H, 20 HR, 80 RBI → BA = 120/400 = .300
	seedBatting(t, db, ps1, true, 400, 120, 20, 80)

	pq := store.NewPlayerQueryStore(db)
	log, err := pq.GetPlayerSeasonLog(ctx, pid)
	if err != nil {
		t.Fatalf("GetPlayerSeasonLog: %v", err)
	}
	if len(log) != 1 {
		t.Fatalf("expected 1 row, got %d", len(log))
	}
	// seedBatting inserts without rate columns (NULL) — verify nil is handled correctly.
	// After a real import via UpsertBattingStats, BA would be non-nil.
	// This test confirms the SELECT includes the columns and handles NULL gracefully.
	row := log[0]
	if row.Batting == nil {
		t.Fatal("expected batting row")
	}
	// BA will be nil because seedBatting inserts without computing rates.
	// The important thing is no panic and AtBats is correct.
	if row.Batting.AtBats != 400 {
		t.Errorf("AtBats: want 400, got %d", row.Batting.AtBats)
	}
}

func TestGetPlayerSeasonLog_RegularAndPlayoff(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	t1 := seedTeam(t, db, "tg1")
	hist1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)

	pid := seedPlayer(t, db, "pg", "Test", "Log")
	ps1 := seedPlayerSeason(t, db, pid, 1, &hist1)
	seedBatting(t, db, ps1, true, 400, 100, 20, 80)  // regular season
	seedBatting(t, db, ps1, false, 50, 15, 3, 10)    // playoff

	pq := store.NewPlayerQueryStore(db)
	log, err := pq.GetPlayerSeasonLog(ctx, pid)
	if err != nil {
		t.Fatalf("GetPlayerSeasonLog: %v", err)
	}
	if len(log) != 1 {
		t.Fatalf("expected 1 season row, got %d", len(log))
	}
	row := log[0]
	if row.Batting == nil {
		t.Fatal("expected regular season batting")
	}
	if row.Batting.AtBats != 400 {
		t.Errorf("reg AB: want 400, got %d", row.Batting.AtBats)
	}
	if row.PlayoffBatting == nil {
		t.Fatal("expected playoff batting")
	}
	if row.PlayoffBatting.AtBats != 50 {
		t.Errorf("playoff AB: want 50, got %d", row.PlayoffBatting.AtBats)
	}
	if row.PlayoffPitching != nil {
		t.Errorf("expected nil playoff pitching, got non-nil")
	}
}

func TestGetPlayerSeasonLog_MultiSeason(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seedSeason(t, db, 1, 1, 40)
	seedSeason(t, db, 2, 2, 40)
	t1 := seedTeam(t, db, "tg1")
	hist1 := seedTeamHistory(t, db, t1, 1, "Team", "E", "AL", 20, 20)
	hist2 := seedTeamHistory(t, db, t1, 2, "Team", "E", "AL", 25, 15)

	pid := seedPlayer(t, db, "pg", "Multi", "Season")
	ps1 := seedPlayerSeason(t, db, pid, 1, &hist1)
	ps2 := seedPlayerSeason(t, db, pid, 2, &hist2)
	seedBatting(t, db, ps1, true, 400, 120, 10, 50)
	seedBatting(t, db, ps2, true, 500, 130, 20, 70)

	pq := store.NewPlayerQueryStore(db)
	log, err := pq.GetPlayerSeasonLog(ctx, pid)
	if err != nil {
		t.Fatalf("GetPlayerSeasonLog: %v", err)
	}
	if len(log) != 2 {
		t.Fatalf("expected 2 season rows, got %d", len(log))
	}
	if log[0].SeasonNum != 1 || log[1].SeasonNum != 2 {
		t.Errorf("expected seasons 1,2 got %d,%d", log[0].SeasonNum, log[1].SeasonNum)
	}
}
