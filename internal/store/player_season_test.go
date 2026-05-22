package store_test

import (
	"context"
	"testing"

	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

// testSeason is a convenience helper for the new Season schema.
func upsertTestSeason(t *testing.T, ss *store.SeasonStore, leagueGUID string, sgSeasonID, seasonNum int) int64 {
	t.Helper()
	id, err := ss.Upsert(context.Background(), store.Season{
		LeagueGUID:       leagueGUID,
		SaveGameSeasonID: sgSeasonID,
		SeasonNum:        seasonNum,
	})
	if err != nil {
		t.Fatalf("upsert season: %v", err)
	}
	return id
}

func TestPlayerSeasonStore_UpsertPlayer_CreatesAndReuses(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := store.NewPlayerSeasonStore(db)
	ctx := context.Background()

	id1, err := s.UpsertPlayer(ctx, store.PlayerIdentity{GameGUID: "AABB", FirstName: "John", LastName: "Doe"})
	if err != nil {
		t.Fatalf("first UpsertPlayer: %v", err)
	}
	// Same GUID, different name (e.g., nickname update) — should reuse ID
	id2, err := s.UpsertPlayer(ctx, store.PlayerIdentity{GameGUID: "AABB", FirstName: "Johnny", LastName: "Doe"})
	if err != nil {
		t.Fatalf("second UpsertPlayer: %v", err)
	}
	if id1 != id2 {
		t.Errorf("expected same player ID on re-upsert, got %d and %d", id1, id2)
	}
}

func TestPlayerSeasonStore_UpsertPlayer_FuzzyMatch(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := store.NewPlayerSeasonStore(db)
	ss := store.NewSeasonStore(db)
	ctx := context.Background()

	seasonID := upsertTestSeason(t, ss, "LEAGUE1", 1, 1)

	// Create a player with GUID1 and a season row (needed for fuzzy match)
	id1, _ := s.UpsertPlayer(ctx, store.PlayerIdentity{
		GameGUID: "GUID1", FirstName: "Mike", LastName: "Jones",
		BatHand: "R", ThrowHand: "R", ChemistryType: "Competitive",
	})
	_, _ = s.UpsertSeason(ctx, store.PlayerSeason{
		PlayerID: id1, SeasonID: seasonID,
		BatHand: "R", ThrowHand: "R", ChemistryType: "Competitive",
		TraitsJSON: "[]", PitchesJSON: "[]",
	})

	// Same player arrives with a new GUID (franchise fork) — should resolve via fuzzy match
	id2, err := s.UpsertPlayer(ctx, store.PlayerIdentity{
		GameGUID: "GUID2", FirstName: "Mike", LastName: "Jones",
		BatHand: "R", ThrowHand: "R", ChemistryType: "Competitive",
	})
	if err != nil {
		t.Fatalf("fuzzy UpsertPlayer: %v", err)
	}
	if id1 != id2 {
		t.Errorf("expected fuzzy match to return same player ID (%d), got %d", id1, id2)
	}

	// Third import with GUID2 should now hit tier-2 (alt_guids), no fuzzy needed
	id3, err := s.UpsertPlayer(ctx, store.PlayerIdentity{GameGUID: "GUID2", FirstName: "Mike", LastName: "Jones"})
	if err != nil {
		t.Fatalf("alt GUID UpsertPlayer: %v", err)
	}
	if id1 != id3 {
		t.Errorf("expected alt GUID lookup to return same player ID (%d), got %d", id1, id3)
	}
}

func TestPlayerSeasonStore_UpsertSeason(t *testing.T) {
	db := testutil.NewTestDB(t)
	ps := store.NewPlayerSeasonStore(db)
	ss := store.NewSeasonStore(db)
	ctx := context.Background()

	seasonID := upsertTestSeason(t, ss, "LEAGUE1", 100, 1)
	playerID, _ := ps.UpsertPlayer(ctx, store.PlayerIdentity{GameGUID: "GUID1", FirstName: "A", LastName: "B"})

	psID, err := ps.UpsertSeason(ctx, store.PlayerSeason{
		PlayerID: playerID, SeasonID: seasonID,
		Age: 25, Salary: 500, PrimaryPosition: "CF",
		TraitsJSON: "[]", PitchesJSON: "[]",
	})
	if err != nil {
		t.Fatalf("UpsertSeason: %v", err)
	}
	if psID == 0 {
		t.Error("expected non-zero player season ID")
	}

	// Re-upsert with different age — same ID returned
	psID2, err := ps.UpsertSeason(ctx, store.PlayerSeason{
		PlayerID: playerID, SeasonID: seasonID,
		Age: 26, Salary: 550, PrimaryPosition: "CF",
		TraitsJSON: "[]", PitchesJSON: "[]",
	})
	if err != nil {
		t.Fatalf("re-upsert: %v", err)
	}
	if psID != psID2 {
		t.Errorf("expected same player_season ID on re-upsert (%d), got %d", psID, psID2)
	}
}

func TestPlayerSeasonStore_UpsertGameStats(t *testing.T) {
	db := testutil.NewTestDB(t)
	ps := store.NewPlayerSeasonStore(db)
	ss := store.NewSeasonStore(db)
	ctx := context.Background()

	seasonID := upsertTestSeason(t, ss, "LEAGUE1", 100, 1)
	playerID, _ := ps.UpsertPlayer(ctx, store.PlayerIdentity{GameGUID: "GUID1", FirstName: "A", LastName: "B"})
	psID, _ := ps.UpsertSeason(ctx, store.PlayerSeason{
		PlayerID: playerID, SeasonID: seasonID, TraitsJSON: "[]", PitchesJSON: "[]",
	})

	if err := ps.UpsertGameStats(ctx, store.PlayerSeasonGameStats{
		PlayerSeasonID: psID, Power: 80, Contact: 75, Speed: 60,
		Fielding: 70, Arm: 65, Velocity: 50, Junk: 50, Accuracy: 50,
	}); err != nil {
		t.Fatalf("UpsertGameStats: %v", err)
	}
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

	seasonID := upsertTestSeason(t, ss, "LEAGUE1", 100, 1)
	playerID, _ := ps.UpsertPlayer(ctx, store.PlayerIdentity{GameGUID: "GUID1", FirstName: "A", LastName: "B"})
	psID, _ := ps.UpsertSeason(ctx, store.PlayerSeason{
		PlayerID: playerID, SeasonID: seasonID, TraitsJSON: "[]", PitchesJSON: "[]",
	})

	bs := store.PlayerSeasonBattingStats{
		PlayerSeasonID: psID, IsRegularSeason: true,
		GamesPlayed: 50, AtBats: 180, Hits: 54, HomeRuns: 12, RBI: 40,
	}
	if err := ps.UpsertBattingStats(ctx, bs); err != nil {
		t.Fatalf("UpsertBattingStats (regular): %v", err)
	}

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

	seasonID := upsertTestSeason(t, ss, "LEAGUE1", 100, 1)
	playerID, _ := ps.UpsertPlayer(ctx, store.PlayerIdentity{GameGUID: "GUID2", FirstName: "P", LastName: "Pitcher"})
	psID, _ := ps.UpsertSeason(ctx, store.PlayerSeason{
		PlayerID: playerID, SeasonID: seasonID,
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

	s1ID := upsertTestSeason(t, ss, "LEAGUE1", 1, 1)
	s2ID := upsertTestSeason(t, ss, "LEAGUE1", 2, 2)

	playerID, _ := ps.UpsertPlayer(ctx, store.PlayerIdentity{GameGUID: "PLAYER1", FirstName: "Same", LastName: "Player"})

	id1, err := ps.UpsertSeason(ctx, store.PlayerSeason{
		PlayerID: playerID, SeasonID: s1ID, Age: 25, TraitsJSON: "[]", PitchesJSON: "[]",
	})
	if err != nil {
		t.Fatalf("season 1 upsert: %v", err)
	}
	id2, err := ps.UpsertSeason(ctx, store.PlayerSeason{
		PlayerID: playerID, SeasonID: s2ID, Age: 26, TraitsJSON: "[]", PitchesJSON: "[]",
	})
	if err != nil {
		t.Fatalf("season 2 upsert: %v", err)
	}
	if id1 == id2 {
		t.Error("different seasons should produce different player_season IDs")
	}
}
