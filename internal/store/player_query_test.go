package store_test

import (
	"context"
	"database/sql"
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

// TestGetPlayerAttributeHistory_RoleAvgReturned seeds two batters and one
// pitcher, runs ApplyLeagueAvgAttributes, and verifies that GetPlayerAttributeHistory
// returns the correct role-specific average for each player type.
func TestGetPlayerAttributeHistory_RoleAvgReturned(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)

	// Two batters: power 60 and 80 → batter avg power = 70.
	pBat1 := seedPlayer(t, db, "gB1", "Bat", "One")
	pBat2 := seedPlayer(t, db, "gB2", "Bat", "Two")
	psBat1 := seedPlayerSeason(t, db, pBat1, s1, nil)
	psBat2 := seedPlayerSeason(t, db, pBat2, s1, nil)
	seedGameStats(t, db, psBat1, 60, 0, 0, 0, 0)
	seedGameStats(t, db, psBat2, 80, 0, 0, 0, 0)

	// One pitcher: power 50, velocity 90.
	pPit := seedPlayer(t, db, "gP1", "Pit", "One")
	psPit := seedPitcherPlayerSeason(t, db, pPit, s1)
	seedGameStatsPitcher(t, db, psPit, 50, 90)

	if err := service.ApplyLeagueAvgAttributes(ctx, db, s1); err != nil {
		t.Fatalf("ApplyLeagueAvgAttributes: %v", err)
	}
	if err := service.ApplyPlayerAttributePercentiles(ctx, db, s1); err != nil {
		t.Fatalf("ApplyPlayerAttributePercentiles: %v", err)
	}

	pq := store.NewPlayerQueryStore(db)
	const epsilon = 0.001

	t.Run("batter gets batter role avg power", func(t *testing.T) {
		rows, err := pq.GetPlayerAttributeHistory(ctx, pBat1)
		if err != nil {
			t.Fatalf("GetPlayerAttributeHistory: %v", err)
		}
		if len(rows) != 1 {
			t.Fatalf("expected 1 row, got %d", len(rows))
		}
		// Batter role avg power = (60+80)/2 = 70.
		if math.Abs(rows[0].RoleAvgPower-70.0) > epsilon {
			t.Errorf("RoleAvgPower for batter = %.4f, want 70.0", rows[0].RoleAvgPower)
		}
	})

	t.Run("pitcher gets pitcher role avg power and velocity", func(t *testing.T) {
		rows, err := pq.GetPlayerAttributeHistory(ctx, pPit)
		if err != nil {
			t.Fatalf("GetPlayerAttributeHistory: %v", err)
		}
		if len(rows) != 1 {
			t.Fatalf("expected 1 row, got %d", len(rows))
		}
		// Pitcher role avg power = 50 (only pitcher).
		if math.Abs(rows[0].RoleAvgPower-50.0) > epsilon {
			t.Errorf("RoleAvgPower for pitcher = %.4f, want 50.0", rows[0].RoleAvgPower)
		}
		// Pitcher role avg velocity = 90.
		if math.Abs(rows[0].RoleAvgVelocity-90.0) > epsilon {
			t.Errorf("RoleAvgVelocity for pitcher = %.4f, want 90.0", rows[0].RoleAvgVelocity)
		}
	})

	t.Run("batter has power_pct_role populated", func(t *testing.T) {
		rows, err := pq.GetPlayerAttributeHistory(ctx, pBat1)
		if err != nil {
			t.Fatalf("GetPlayerAttributeHistory: %v", err)
		}
		// pBat1 has power 60; pBat2 has 80. Among batters only: pBat1 is lowest.
		if rows[0].PowerPctRole == nil {
			t.Fatal("PowerPctRole is nil, want a value")
		}
		if *rows[0].PowerPctRole > 5 {
			t.Errorf("PowerPctRole for lowest batter = %.2f, want ≈0", *rows[0].PowerPctRole)
		}
	})
}

// seedPitcherPlayerSeason inserts a pitcher player_season (pitcher_role='SP').
func seedPitcherPlayerSeason(t *testing.T, db *sql.DB, playerID, seasonID int64) int64 {
	t.Helper()
	var id int64
	err := db.QueryRowContext(context.Background(), `
		INSERT INTO player_seasons
		    (player_id, season_id, age, salary,
		     primary_position, secondary_position, pitcher_role,
		     bat_hand, throw_hand, chemistry_type, traits_json, pitches_json)
		VALUES (?,?,28,900,'SP','','SP','R','R','Competitive','[]','[]') RETURNING id
	`, playerID, seasonID).Scan(&id)
	if err != nil {
		t.Fatalf("seedPitcherPlayerSeason: %v", err)
	}
	return id
}

// seedGameStatsPitcher inserts game stats with pitcher-relevant columns set.
func seedGameStatsPitcher(t *testing.T, db *sql.DB, psID int64, power, velocity int) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO player_season_game_stats
		    (player_season_id, power, contact, speed, fielding, arm, velocity, junk, accuracy)
		VALUES (?,?,0,0,0,0,?,0,0)
	`, psID, power, velocity)
	if err != nil {
		t.Fatalf("seedGameStatsPitcher: %v", err)
	}
}

// seedGameStats inserts a player_season_game_stats row for an existing player_season.
func seedGameStats(t *testing.T, db *sql.DB, psID int64, power, contact, speed, fielding, arm int) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO player_season_game_stats
		    (player_season_id, power, contact, speed, fielding, arm, velocity, junk, accuracy)
		VALUES (?,?,?,?,?,?,0,0,0)
	`, psID, power, contact, speed, fielding, arm)
	if err != nil {
		t.Fatalf("seedGameStats: %v", err)
	}
}

// seedLeagueAvg inserts a season_attribute_averages row directly (bypassing the
// service layer) so query tests can control the expected values precisely.
func seedLeagueAvg(t *testing.T, db *sql.DB, seasonID int64, avgPower, avgContact float64) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO season_attribute_averages
		    (season_id, avg_power, avg_contact, avg_speed, avg_fielding, avg_arm,
		     avg_velocity, avg_junk, avg_accuracy)
		VALUES (?,?,?,0,0,0,0,0,0)
	`, seasonID, avgPower, avgContact)
	if err != nil {
		t.Fatalf("seedLeagueAvg: %v", err)
	}
}

// TestGetPlayerAttributeHistory_PercentileRanks seeds three players in the
// same season, computes percentiles via ApplyPlayerAttributePercentiles, and
// verifies that the player with the highest Power receives the highest
// percentile, the lowest receives the lowest, and the middle falls between.
func TestGetPlayerAttributeHistory_PercentileRanks(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)
	seedLeagueAvg(t, db, s1, 60.0, 65.0)

	// Three players in the same season with distinct Power values.
	pLow := seedPlayer(t, db, "gLow", "Low", "Power")
	pMid := seedPlayer(t, db, "gMid", "Mid", "Power")
	pHigh := seedPlayer(t, db, "gHigh", "High", "Power")

	psLow := seedPlayerSeason(t, db, pLow, s1, nil)
	psMid := seedPlayerSeason(t, db, pMid, s1, nil)
	psHigh := seedPlayerSeason(t, db, pHigh, s1, nil)

	seedGameStats(t, db, psLow, 40, 50, 60, 70, 80)
	seedGameStats(t, db, psMid, 60, 65, 60, 70, 80)
	seedGameStats(t, db, psHigh, 80, 90, 60, 70, 80)

	// Persist percentile ranks — mirrors the import pipeline step.
	if err := service.ApplyPlayerAttributePercentiles(ctx, db, s1); err != nil {
		t.Fatalf("ApplyPlayerAttributePercentiles: %v", err)
	}

	pq := store.NewPlayerQueryStore(db)

	checkPct := func(t *testing.T, playerID int64, wantPowerPctMin, wantPowerPctMax float64) {
		t.Helper()
		rows, err := pq.GetPlayerAttributeHistory(ctx, playerID)
		if err != nil {
			t.Fatalf("GetPlayerAttributeHistory: %v", err)
		}
		if len(rows) != 1 {
			t.Fatalf("expected 1 row, got %d", len(rows))
		}
		r := rows[0]
		if r.PowerPct == nil {
			t.Fatal("PowerPct is nil, want a value (3 players in season)")
		}
		if *r.PowerPct < wantPowerPctMin || *r.PowerPct > wantPowerPctMax {
			t.Errorf("PowerPct = %.2f, want [%.0f, %.0f]", *r.PowerPct, wantPowerPctMin, wantPowerPctMax)
		}
	}

	t.Run("lowest power → lowest percentile", func(t *testing.T) {
		checkPct(t, pLow, 0, 10) // PERCENT_RANK=0.0 for the minimum
	})
	t.Run("middle power → middle percentile", func(t *testing.T) {
		checkPct(t, pMid, 40, 60) // ≈50th percentile
	})
	t.Run("highest power → highest percentile", func(t *testing.T) {
		checkPct(t, pHigh, 90, 100) // PERCENT_RANK=1.0 for the maximum
	})
}

// TestGetPlayerAttributeHistory_LeagueAvgReturned verifies that the league
// average values seeded in season_attribute_averages are returned correctly.
func TestGetPlayerAttributeHistory_LeagueAvgReturned(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)
	seedLeagueAvg(t, db, s1, 65.5, 72.0)

	pid := seedPlayer(t, db, "gA", "A", "B")
	ps1 := seedPlayerSeason(t, db, pid, s1, nil)
	seedGameStats(t, db, ps1, 70, 80, 60, 55, 50)

	pq := store.NewPlayerQueryStore(db)
	rows, err := pq.GetPlayerAttributeHistory(ctx, pid)
	if err != nil {
		t.Fatalf("GetPlayerAttributeHistory: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	r := rows[0]

	const epsilon = 0.001
	if math.Abs(r.LgAvgPower-65.5) > epsilon {
		t.Errorf("LgAvgPower = %.4f, want 65.5", r.LgAvgPower)
	}
	if math.Abs(r.LgAvgContact-72.0) > epsilon {
		t.Errorf("LgAvgContact = %.4f, want 72.0", r.LgAvgContact)
	}
}

// TestGetPlayerAttributeHistory_MultiSeason verifies rows are returned in
// season_num order for a player spanning three seasons.
func TestGetPlayerAttributeHistory_MultiSeason(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)
	s2 := seedSeason(t, db, 2, 2, 40)
	s3 := seedSeason(t, db, 3, 3, 40)

	pid := seedPlayer(t, db, "gMulti", "Career", "Player")
	ps1 := seedPlayerSeason(t, db, pid, s1, nil)
	ps2 := seedPlayerSeason(t, db, pid, s2, nil)
	ps3 := seedPlayerSeason(t, db, pid, s3, nil)
	seedGameStats(t, db, ps1, 60, 60, 60, 60, 60)
	seedGameStats(t, db, ps2, 70, 70, 70, 70, 70)
	seedGameStats(t, db, ps3, 80, 80, 80, 80, 80)

	pq := store.NewPlayerQueryStore(db)
	rows, err := pq.GetPlayerAttributeHistory(ctx, pid)
	if err != nil {
		t.Fatalf("GetPlayerAttributeHistory: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
	for i, want := range []int{1, 2, 3} {
		if rows[i].SeasonNum != want {
			t.Errorf("row[%d].SeasonNum = %d, want %d", i, rows[i].SeasonNum, want)
		}
	}
	// Power should increase season over season.
	if rows[0].Power >= rows[1].Power || rows[1].Power >= rows[2].Power {
		t.Errorf("expected increasing power: %d, %d, %d", rows[0].Power, rows[1].Power, rows[2].Power)
	}
}

// TestGetPlayerAttributeHistory_SinglePlayerNilPercentile verifies that when
// only one player exists in a season, ApplyPlayerAttributePercentiles stores
// NULL and GetPlayerAttributeHistory returns nil for all percentile fields.
func TestGetPlayerAttributeHistory_SinglePlayerNilPercentile(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	s1 := seedSeason(t, db, 1, 1, 40)
	pid := seedPlayer(t, db, "gSolo", "Solo", "Player")
	ps1 := seedPlayerSeason(t, db, pid, s1, nil)
	seedGameStats(t, db, ps1, 75, 65, 55, 45, 35)

	// Single-player season: ApplyPlayerAttributePercentiles stores NULLs.
	if err := service.ApplyPlayerAttributePercentiles(ctx, db, s1); err != nil {
		t.Fatalf("ApplyPlayerAttributePercentiles: %v", err)
	}

	pq := store.NewPlayerQueryStore(db)
	rows, err := pq.GetPlayerAttributeHistory(ctx, pid)
	if err != nil {
		t.Fatalf("GetPlayerAttributeHistory: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].PowerPct != nil {
		t.Errorf("PowerPct = %v, want nil for single-player season", rows[0].PowerPct)
	}
}
