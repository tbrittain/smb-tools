package service_test

import (
	"context"
	"math"
	"testing"

	"smb-tools/internal/service"
	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

// TestApplyLeagueAvgAttributes_MultiPlayer seeds three active players with
// known attributes and one all-zero player. Verifies stored averages match
// the mean of the active players only (zeros excluded via NULLIF).
func TestApplyLeagueAvgAttributes_MultiPlayer(t *testing.T) {
	ctx := context.Background()
	cdb := testutil.NewTestDB(t)
	seasonStore := store.NewSeasonStore(cdb)
	playerStore := store.NewPlayerSeasonStore(cdb)

	seasonID, err := seasonStore.Upsert(ctx, store.Season{
		LeagueGUID: "avg-test-001", SaveGameSeasonID: 1, SeasonNum: 1,
	})
	if err != nil {
		t.Fatalf("upsert season: %v", err)
	}

	type entry struct {
		guid  string
		power int
		arm   int
	}
	players := []entry{
		{guid: "P1", power: 60, arm: 70},
		{guid: "P2", power: 80, arm: 50},
		{guid: "P3", power: 40, arm: 90},
		// All-zero row — must be excluded from averages.
		{guid: "P0", power: 0, arm: 0},
	}

	for _, p := range players {
		pid, err := playerStore.UpsertPlayer(ctx, store.PlayerIdentity{
			GameGUID: p.guid, FirstName: "A", LastName: "B",
		})
		if err != nil {
			t.Fatalf("upsert player %s: %v", p.guid, err)
		}
		psID, err := playerStore.UpsertSeason(ctx, store.PlayerSeason{
			PlayerID: pid, SeasonID: seasonID,
			BatHand: "R", ThrowHand: "R", ChemistryType: "Competitive",
			TraitsJSON: "[]", PitchesJSON: "[]",
		})
		if err != nil {
			t.Fatalf("upsert player season %s: %v", p.guid, err)
		}
		if err := playerStore.UpsertGameStats(ctx, store.PlayerSeasonGameStats{
			PlayerSeasonID: psID, Power: p.power, Arm: p.arm,
		}); err != nil {
			t.Fatalf("upsert game stats %s: %v", p.guid, err)
		}
	}

	if err := service.ApplyLeagueAvgAttributes(ctx, cdb, seasonID); err != nil {
		t.Fatalf("ApplyLeagueAvgAttributes: %v", err)
	}

	var avgPower, avgArm float64
	if err := cdb.QueryRowContext(ctx,
		`SELECT avg_power, avg_arm FROM season_attribute_averages WHERE season_id = ?`, seasonID,
	).Scan(&avgPower, &avgArm); err != nil {
		t.Fatalf("reading stored averages: %v", err)
	}

	// Expected: AVG(60,80,40)=60.0, AVG(70,50,90)=70.0 — the zero player excluded.
	const epsilon = 0.001
	if math.Abs(avgPower-60.0) > epsilon {
		t.Errorf("avg_power = %.4f, want 60.0", avgPower)
	}
	if math.Abs(avgArm-70.0) > epsilon {
		t.Errorf("avg_arm = %.4f, want 70.0", avgArm)
	}
}

// TestApplyLeagueAvgAttributes_SinglePlayer verifies that when only one player
// has non-zero attributes, the stored average equals that player's own value.
func TestApplyLeagueAvgAttributes_SinglePlayer(t *testing.T) {
	ctx := context.Background()
	cdb := testutil.NewTestDB(t)
	seasonStore := store.NewSeasonStore(cdb)
	playerStore := store.NewPlayerSeasonStore(cdb)

	seasonID, err := seasonStore.Upsert(ctx, store.Season{
		LeagueGUID: "avg-test-002", SaveGameSeasonID: 2, SeasonNum: 2,
	})
	if err != nil {
		t.Fatalf("upsert season: %v", err)
	}

	pid, err := playerStore.UpsertPlayer(ctx, store.PlayerIdentity{
		GameGUID: "SOLO", FirstName: "Solo", LastName: "Player",
	})
	if err != nil {
		t.Fatalf("upsert player: %v", err)
	}
	psID, err := playerStore.UpsertSeason(ctx, store.PlayerSeason{
		PlayerID: pid, SeasonID: seasonID,
		BatHand: "R", ThrowHand: "R", ChemistryType: "Competitive",
		TraitsJSON: "[]", PitchesJSON: "[]",
	})
	if err != nil {
		t.Fatalf("upsert player season: %v", err)
	}
	if err := playerStore.UpsertGameStats(ctx, store.PlayerSeasonGameStats{
		PlayerSeasonID: psID, Power: 75, Contact: 65,
	}); err != nil {
		t.Fatalf("upsert game stats: %v", err)
	}

	if err := service.ApplyLeagueAvgAttributes(ctx, cdb, seasonID); err != nil {
		t.Fatalf("ApplyLeagueAvgAttributes: %v", err)
	}

	var avgPower, avgContact float64
	if err := cdb.QueryRowContext(ctx,
		`SELECT avg_power, avg_contact FROM season_attribute_averages WHERE season_id = ?`, seasonID,
	).Scan(&avgPower, &avgContact); err != nil {
		t.Fatalf("reading stored averages: %v", err)
	}

	const epsilon = 0.001
	if math.Abs(avgPower-75.0) > epsilon {
		t.Errorf("avg_power = %.4f, want 75.0", avgPower)
	}
	if math.Abs(avgContact-65.0) > epsilon {
		t.Errorf("avg_contact = %.4f, want 65.0", avgContact)
	}
}

// TestApplyLeagueAvgAttributes_RoleSplit seeds two batters and one pitcher,
// then verifies that batter-only and pitcher-only averages are computed
// separately and do not bleed into each other.
func TestApplyLeagueAvgAttributes_RoleSplit(t *testing.T) {
	ctx := context.Background()
	cdb := testutil.NewTestDB(t)
	seasonStore := store.NewSeasonStore(cdb)
	playerStore := store.NewPlayerSeasonStore(cdb)

	seasonID, err := seasonStore.Upsert(ctx, store.Season{
		LeagueGUID: "avg-test-004", SaveGameSeasonID: 4, SeasonNum: 4,
	})
	if err != nil {
		t.Fatalf("upsert season: %v", err)
	}

	// Two batters: power 60 and 80 → batter avg power = 70.
	for _, g := range []struct{ guid string; power int }{{guid: "B1", power: 60}, {guid: "B2", power: 80}} {
		pid, err := playerStore.UpsertPlayer(ctx, store.PlayerIdentity{GameGUID: g.guid, FirstName: "B", LastName: "B"})
		if err != nil {
			t.Fatalf("upsert batter %s: %v", g.guid, err)
		}
		psID, err := playerStore.UpsertSeason(ctx, store.PlayerSeason{
			PlayerID: pid, SeasonID: seasonID,
			BatHand: "R", ThrowHand: "R", ChemistryType: "Competitive",
			TraitsJSON: "[]", PitchesJSON: "[]",
		})
		if err != nil {
			t.Fatalf("upsert batter season %s: %v", g.guid, err)
		}
		if err := playerStore.UpsertGameStats(ctx, store.PlayerSeasonGameStats{
			PlayerSeasonID: psID, Power: g.power, Velocity: 0,
		}); err != nil {
			t.Fatalf("upsert batter stats %s: %v", g.guid, err)
		}
	}

	// One pitcher: power 50, velocity 90 → pitcher avg power = 50, pitcher avg velocity = 90.
	pid, err := playerStore.UpsertPlayer(ctx, store.PlayerIdentity{GameGUID: "P1", FirstName: "P", LastName: "P"})
	if err != nil {
		t.Fatalf("upsert pitcher: %v", err)
	}
	psID, err := playerStore.UpsertSeason(ctx, store.PlayerSeason{
		PlayerID: pid, SeasonID: seasonID, PitcherRole: "SP",
		BatHand: "R", ThrowHand: "R", ChemistryType: "Competitive",
		TraitsJSON: "[]", PitchesJSON: "[]",
	})
	if err != nil {
		t.Fatalf("upsert pitcher season: %v", err)
	}
	if err := playerStore.UpsertGameStats(ctx, store.PlayerSeasonGameStats{
		PlayerSeasonID: psID, Power: 50, Velocity: 90,
	}); err != nil {
		t.Fatalf("upsert pitcher stats: %v", err)
	}

	if err := service.ApplyLeagueAvgAttributes(ctx, cdb, seasonID); err != nil {
		t.Fatalf("ApplyLeagueAvgAttributes: %v", err)
	}

	var batterAvgPower, pitcherAvgPower, pitcherAvgVelocity float64
	if err := cdb.QueryRowContext(ctx, `
		SELECT batter_avg_power, pitcher_avg_power, pitcher_avg_velocity
		FROM season_attribute_averages WHERE season_id = ?
	`, seasonID).Scan(&batterAvgPower, &pitcherAvgPower, &pitcherAvgVelocity); err != nil {
		t.Fatalf("reading role averages: %v", err)
	}

	const epsilon = 0.001
	// batter avg power = (60+80)/2 = 70
	if math.Abs(batterAvgPower-70.0) > epsilon {
		t.Errorf("batter_avg_power = %.4f, want 70.0", batterAvgPower)
	}
	// pitcher avg power = 50 (single pitcher)
	if math.Abs(pitcherAvgPower-50.0) > epsilon {
		t.Errorf("pitcher_avg_power = %.4f, want 50.0", pitcherAvgPower)
	}
	// pitcher avg velocity = 90
	if math.Abs(pitcherAvgVelocity-90.0) > epsilon {
		t.Errorf("pitcher_avg_velocity = %.4f, want 90.0", pitcherAvgVelocity)
	}
}

// TestApplyLeagueAvgAttributes_Idempotent verifies calling the function twice
// for the same season does not fail and leaves exactly one row.
func TestApplyLeagueAvgAttributes_Idempotent(t *testing.T) {
	ctx := context.Background()
	cdb := testutil.NewTestDB(t)
	seasonStore := store.NewSeasonStore(cdb)
	playerStore := store.NewPlayerSeasonStore(cdb)

	seasonID, err := seasonStore.Upsert(ctx, store.Season{
		LeagueGUID: "avg-test-003", SaveGameSeasonID: 3, SeasonNum: 3,
	})
	if err != nil {
		t.Fatalf("upsert season: %v", err)
	}

	pid, err := playerStore.UpsertPlayer(ctx, store.PlayerIdentity{
		GameGUID: "IDEM", FirstName: "I", LastName: "D",
	})
	if err != nil {
		t.Fatalf("upsert player: %v", err)
	}
	psID, err := playerStore.UpsertSeason(ctx, store.PlayerSeason{
		PlayerID: pid, SeasonID: seasonID,
		BatHand: "R", ThrowHand: "R", ChemistryType: "Competitive",
		TraitsJSON: "[]", PitchesJSON: "[]",
	})
	if err != nil {
		t.Fatalf("upsert player season: %v", err)
	}
	if err := playerStore.UpsertGameStats(ctx, store.PlayerSeasonGameStats{
		PlayerSeasonID: psID, Power: 80,
	}); err != nil {
		t.Fatalf("upsert game stats: %v", err)
	}

	if err := service.ApplyLeagueAvgAttributes(ctx, cdb, seasonID); err != nil {
		t.Fatalf("first ApplyLeagueAvgAttributes: %v", err)
	}
	if err := service.ApplyLeagueAvgAttributes(ctx, cdb, seasonID); err != nil {
		t.Fatalf("second ApplyLeagueAvgAttributes (idempotent): %v", err)
	}

	var count int
	if err := cdb.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM season_attribute_averages WHERE season_id = ?`, seasonID,
	).Scan(&count); err != nil {
		t.Fatalf("counting rows: %v", err)
	}
	if count != 1 {
		t.Errorf("row count = %d, want 1 (idempotent upsert)", count)
	}
}
