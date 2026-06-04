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
