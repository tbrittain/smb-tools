package store_test

import (
	"context"
	"testing"

	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

func seedPlayerSeasonPrereqs(t *testing.T, db interface{ Close() error }) (
	*store.SeasonStore, *store.TeamHistoryStore, *store.PlayerSeasonStore,
) {
	t.Helper()
	// db is already opened — grab it from a separate path
	return nil, nil, nil
}

func TestPlayerSeasonStore_UpsertPlayer_CreatesAndReuses(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := store.NewPlayerSeasonStore(db)
	ctx := context.Background()

	id1, err := s.UpsertPlayer(ctx, store.Player{GameGUID: "AABB", FirstName: "John", LastName: "Doe"})
	if err != nil {
		t.Fatalf("first UpsertPlayer: %v", err)
	}
	// Same GUID, different name (e.g., nickname update) — should reuse ID
	id2, err := s.UpsertPlayer(ctx, store.Player{GameGUID: "AABB", FirstName: "Johnny", LastName: "Doe"})
	if err != nil {
		t.Fatalf("second UpsertPlayer: %v", err)
	}
	if id1 != id2 {
		t.Errorf("expected same player ID on re-upsert, got %d and %d", id1, id2)
	}
}

func TestPlayerSeasonStore_UpsertSeason(t *testing.T) {
	db := testutil.NewTestDB(t)
	ps := store.NewPlayerSeasonStore(db)
	ss := store.NewSeasonStore(db)
	ctx := context.Background()

	_ = ss.Upsert(ctx, store.Season{ID: 100, SeasonNum: 1})
	playerID, _ := ps.UpsertPlayer(ctx, store.Player{GameGUID: "GUID1", FirstName: "A", LastName: "B"})

	seasonID, err := ps.UpsertSeason(ctx, store.PlayerSeason{
		PlayerID: playerID, SeasonID: 100,
		Age: 25, Salary: 500, PrimaryPosition: "CF",
		TraitsJSON: "[]", PitchesJSON: "[]",
	})
	if err != nil {
		t.Fatalf("UpsertSeason: %v", err)
	}
	if seasonID == 0 {
		t.Error("expected non-zero player season ID")
	}

	// Re-upsert with different age — same ID returned
	seasonID2, err := ps.UpsertSeason(ctx, store.PlayerSeason{
		PlayerID: playerID, SeasonID: 100,
		Age: 26, Salary: 550, PrimaryPosition: "CF",
		TraitsJSON: "[]", PitchesJSON: "[]",
	})
	if err != nil {
		t.Fatalf("re-upsert: %v", err)
	}
	if seasonID != seasonID2 {
		t.Errorf("expected same player_season ID on re-upsert (%d), got %d", seasonID, seasonID2)
	}
}

func TestPlayerSeasonStore_UpsertGameStats(t *testing.T) {
	db := testutil.NewTestDB(t)
	ps := store.NewPlayerSeasonStore(db)
	ss := store.NewSeasonStore(db)
	ctx := context.Background()

	_ = ss.Upsert(ctx, store.Season{ID: 100, SeasonNum: 1})
	playerID, _ := ps.UpsertPlayer(ctx, store.Player{GameGUID: "GUID1", FirstName: "A", LastName: "B"})
	psID, _ := ps.UpsertSeason(ctx, store.PlayerSeason{
		PlayerID: playerID, SeasonID: 100, TraitsJSON: "[]", PitchesJSON: "[]",
	})

	if err := ps.UpsertGameStats(ctx, store.PlayerSeasonGameStats{
		PlayerSeasonID: psID, Power: 80, Contact: 75, Speed: 60,
		Fielding: 70, Arm: 65, Velocity: 50, Junk: 50, Accuracy: 50,
	}); err != nil {
		t.Fatalf("UpsertGameStats: %v", err)
	}
	// Re-upsert should not error
	if err := ps.UpsertGameStats(ctx, store.PlayerSeasonGameStats{
		PlayerSeasonID: psID, Power: 82,
	}); err != nil {
		t.Fatalf("re-upsert GameStats: %v", err)
	}
}

func TestPlayerSeasonStore_UpsertBattingStats(t *testing.T) {
	db := testutil.NewTestDB(t)
	ps := store.NewPlayerSeasonStore(db)
	ss := store.NewSeasonStore(db)
	ctx := context.Background()

	_ = ss.Upsert(ctx, store.Season{ID: 100, SeasonNum: 1})
	playerID, _ := ps.UpsertPlayer(ctx, store.Player{GameGUID: "GUID1", FirstName: "A", LastName: "B"})
	psID, _ := ps.UpsertSeason(ctx, store.PlayerSeason{
		PlayerID: playerID, SeasonID: 100, TraitsJSON: "[]", PitchesJSON: "[]",
	})

	bs := store.PlayerSeasonBattingStats{
		PlayerSeasonID: psID, IsRegularSeason: true,
		GamesPlayed: 50, AtBats: 180, Hits: 54, HomeRuns: 12, RBI: 40,
	}
	if err := ps.UpsertBattingStats(ctx, bs); err != nil {
		t.Fatalf("UpsertBattingStats (regular): %v", err)
	}

	// Playoff stats stored separately
	bs.IsRegularSeason = false
	bs.AtBats = 18
	bs.Hits = 6
	if err := ps.UpsertBattingStats(ctx, bs); err != nil {
		t.Fatalf("UpsertBattingStats (playoff): %v", err)
	}
}

func TestPlayerSeasonStore_UpsertPitchingStats(t *testing.T) {
	db := testutil.NewTestDB(t)
	ps := store.NewPlayerSeasonStore(db)
	ss := store.NewSeasonStore(db)
	ctx := context.Background()

	_ = ss.Upsert(ctx, store.Season{ID: 100, SeasonNum: 1})
	playerID, _ := ps.UpsertPlayer(ctx, store.Player{GameGUID: "GUID2", FirstName: "P", LastName: "Pitcher"})
	psID, _ := ps.UpsertSeason(ctx, store.PlayerSeason{
		PlayerID: playerID, SeasonID: 100,
		PrimaryPosition: "P", PitcherRole: "SP",
		TraitsJSON: "[]", PitchesJSON: "[]",
	})

	pitch := store.PlayerSeasonPitchingStats{
		PlayerSeasonID: psID, IsRegularSeason: true,
		Wins: 12, Losses: 8, Games: 25, GamesStarted: 25,
		OutsPitched: 540, HitsAllowed: 140, EarnedRuns: 55, Strikeouts: 180,
	}
	if err := ps.UpsertPitchingStats(ctx, pitch); err != nil {
		t.Fatalf("UpsertPitchingStats: %v", err)
	}
}

func TestPlayerSeasonStore_MultipleSeasonsSamePlayer(t *testing.T) {
	db := testutil.NewTestDB(t)
	ps := store.NewPlayerSeasonStore(db)
	ss := store.NewSeasonStore(db)
	ctx := context.Background()

	_ = ss.Upsert(ctx, store.Season{ID: 1, SeasonNum: 1})
	_ = ss.Upsert(ctx, store.Season{ID: 2, SeasonNum: 2})

	playerID, _ := ps.UpsertPlayer(ctx, store.Player{GameGUID: "PLAYER1", FirstName: "Same", LastName: "Player"})

	id1, err := ps.UpsertSeason(ctx, store.PlayerSeason{
		PlayerID: playerID, SeasonID: 1, Age: 25, TraitsJSON: "[]", PitchesJSON: "[]",
	})
	if err != nil {
		t.Fatalf("season 1 upsert: %v", err)
	}
	id2, err := ps.UpsertSeason(ctx, store.PlayerSeason{
		PlayerID: playerID, SeasonID: 2, Age: 26, TraitsJSON: "[]", PitchesJSON: "[]",
	})
	if err != nil {
		t.Fatalf("season 2 upsert: %v", err)
	}
	if id1 == id2 {
		t.Error("different seasons should produce different player_season IDs")
	}
}
