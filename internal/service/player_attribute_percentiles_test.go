package service_test

import (
	"context"
	"database/sql"
	"testing"

	"smb-tools/internal/service"
	"smb-tools/internal/testutil"
)

func TestApplyPlayerAttributePercentiles_MultiPlayer(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seasonID := seedPctSeason(t, db, 1)
	pA := seedPctPlayer(t, db, "gA")
	pB := seedPctPlayer(t, db, "gB")
	pC := seedPctPlayer(t, db, "gC")
	psA := seedPctPlayerSeason(t, db, pA, seasonID)
	psB := seedPctPlayerSeason(t, db, pB, seasonID)
	psC := seedPctPlayerSeason(t, db, pC, seasonID)

	// Power: A=40 (lowest), B=60 (mid), C=80 (highest).
	seedPctGameStats(t, db, psA, 40)
	seedPctGameStats(t, db, psB, 60)
	seedPctGameStats(t, db, psC, 80)

	if err := service.ApplyPlayerAttributePercentiles(ctx, db, seasonID); err != nil {
		t.Fatalf("ApplyPlayerAttributePercentiles: %v", err)
	}

	rowCount := countPctRows(t, db)
	if rowCount != 3 {
		t.Fatalf("expected 3 rows in player_season_attribute_percentiles, got %d", rowCount)
	}

	// Lowest power → PERCENT_RANK = 0 (0%).
	pctA := fetchPowerPct(t, db, psA)
	if pctA == nil {
		t.Fatal("power_pct for lowest player is nil, want 0")
	}
	if *pctA > 5 {
		t.Errorf("power_pct for lowest player = %.2f, want ≈0", *pctA)
	}

	// Highest power → PERCENT_RANK = 1 (100%).
	pctC := fetchPowerPct(t, db, psC)
	if pctC == nil {
		t.Fatal("power_pct for highest player is nil, want ~100")
	}
	if *pctC < 95 {
		t.Errorf("power_pct for highest player = %.2f, want ≈100", *pctC)
	}

	// Middle power → ≈50th percentile.
	pctB := fetchPowerPct(t, db, psB)
	if pctB == nil {
		t.Fatal("power_pct for middle player is nil, want ~50")
	}
	if *pctB < 40 || *pctB > 60 {
		t.Errorf("power_pct for middle player = %.2f, want [40, 60]", *pctB)
	}
}

func TestApplyPlayerAttributePercentiles_SinglePlayer(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seasonID := seedPctSeason(t, db, 1)
	pid := seedPctPlayer(t, db, "gSolo")
	psID := seedPctPlayerSeason(t, db, pid, seasonID)
	seedPctGameStats(t, db, psID, 75)

	if err := service.ApplyPlayerAttributePercentiles(ctx, db, seasonID); err != nil {
		t.Fatalf("ApplyPlayerAttributePercentiles: %v", err)
	}

	pct := fetchPowerPct(t, db, psID)
	if pct != nil {
		t.Errorf("power_pct = %v, want nil for single-player season", *pct)
	}
}

func TestApplyPlayerAttributePercentiles_Idempotent(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	seasonID := seedPctSeason(t, db, 1)
	pid := seedPctPlayer(t, db, "gIdem")
	psID := seedPctPlayerSeason(t, db, pid, seasonID)
	seedPctGameStats(t, db, psID, 60)

	for i := range 2 {
		if err := service.ApplyPlayerAttributePercentiles(ctx, db, seasonID); err != nil {
			t.Fatalf("call %d: %v", i+1, err)
		}
	}
	if n := countPctRows(t, db); n != 1 {
		t.Errorf("expected 1 row after 2 calls, got %d", n)
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func seedPctSeason(t *testing.T, db *sql.DB, num int) int64 {
	t.Helper()
	var id int64
	err := db.QueryRowContext(context.Background(), `
		INSERT INTO seasons (league_guid, save_game_season_id, season_num, num_games)
		VALUES ('TESTLEAGUE', ?, ?, 100) RETURNING id
	`, num, num).Scan(&id)
	if err != nil {
		t.Fatalf("seedPctSeason: %v", err)
	}
	return id
}

func seedPctPlayer(t *testing.T, db *sql.DB, guid string) int64 {
	t.Helper()
	var id int64
	err := db.QueryRowContext(context.Background(), `
		INSERT INTO players (game_guid, first_name, last_name) VALUES (?, 'F', 'L') RETURNING id
	`, guid).Scan(&id)
	if err != nil {
		t.Fatalf("seedPctPlayer: %v", err)
	}
	return id
}

func seedPctPlayerSeason(t *testing.T, db *sql.DB, playerID, seasonID int64) int64 {
	t.Helper()
	var id int64
	err := db.QueryRowContext(context.Background(), `
		INSERT INTO player_seasons
		    (player_id, season_id, age, salary,
		     primary_position, secondary_position, pitcher_role,
		     bat_hand, throw_hand, chemistry_type, traits_json, pitches_json)
		VALUES (?,?,25,1000,'CF','','','R','R','','[]','[]') RETURNING id
	`, playerID, seasonID).Scan(&id)
	if err != nil {
		t.Fatalf("seedPctPlayerSeason: %v", err)
	}
	return id
}

func seedPctGameStats(t *testing.T, db *sql.DB, psID int64, power int) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO player_season_game_stats
		    (player_season_id, power, contact, speed, fielding, arm, velocity, junk, accuracy)
		VALUES (?, ?, 50, 50, 50, 50, 0, 0, 0)
	`, psID, power)
	if err != nil {
		t.Fatalf("seedPctGameStats: %v", err)
	}
}

func countPctRows(t *testing.T, db *sql.DB) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM player_season_attribute_percentiles`).Scan(&n); err != nil {
		t.Fatalf("countPctRows: %v", err)
	}
	return n
}

func fetchPowerPct(t *testing.T, db *sql.DB, psID int64) *float64 {
	t.Helper()
	var v sql.NullFloat64
	if err := db.QueryRowContext(context.Background(),
		`SELECT power_pct FROM player_season_attribute_percentiles WHERE player_season_id = ?`, psID,
	).Scan(&v); err != nil {
		t.Fatalf("fetchPowerPct: %v", err)
	}
	if !v.Valid {
		return nil
	}
	return &v.Float64
}
