package store_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

const testLG = "TESTLEAGUE000000000000000000000000"

func TestSeasonStore_UpsertAndGet(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := store.NewSeasonStore(db)
	ctx := context.Background()

	id, err := s.Upsert(ctx, store.Season{LeagueGUID: testLG, SaveGameSeasonID: 100, SeasonNum: 1, NumGames: 50})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero companion DB ID")
	}

	got, err := s.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.SeasonNum != 1 || got.NumGames != 50 {
		t.Errorf("got %+v, want SeasonNum=1 NumGames=50", got)
	}
	if got.LeagueGUID != testLG {
		t.Errorf("league_guid: got %q, want %q", got.LeagueGUID, testLG)
	}
	if got.SaveGameSeasonID != 100 {
		t.Errorf("save_game_season_id: got %d, want 100", got.SaveGameSeasonID)
	}
}

func TestSeasonStore_Upsert_Idempotent(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := store.NewSeasonStore(db)
	ctx := context.Background()

	id1, _ := s.Upsert(ctx, store.Season{LeagueGUID: testLG, SaveGameSeasonID: 1, SeasonNum: 1, NumGames: 50})
	// Re-upsert same (league_guid, save_game_season_id) with updated values
	id2, err := s.Upsert(ctx, store.Season{LeagueGUID: testLG, SaveGameSeasonID: 1, SeasonNum: 1, NumGames: 60})
	if err != nil {
		t.Fatalf("second Upsert: %v", err)
	}
	if id1 != id2 {
		t.Errorf("expected same companion ID on re-upsert, got %d then %d", id1, id2)
	}
	got, _ := s.GetByID(ctx, id2)
	if got.NumGames != 60 {
		t.Errorf("expected NumGames=60 after re-upsert, got %d", got.NumGames)
	}
}

func TestSeasonStore_ForkNoCollision(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := store.NewSeasonStore(db)
	ctx := context.Background()

	// Two different leagues both have a season with save_game_season_id=1
	// — they must not collide.
	id1, err := s.Upsert(ctx, store.Season{LeagueGUID: "LEAGUE-A", SaveGameSeasonID: 1, SeasonNum: 1})
	if err != nil {
		t.Fatalf("Upsert LEAGUE-A: %v", err)
	}
	id2, err := s.Upsert(ctx, store.Season{LeagueGUID: "LEAGUE-B", SaveGameSeasonID: 1, SeasonNum: 16})
	if err != nil {
		t.Fatalf("Upsert LEAGUE-B: %v", err)
	}
	if id1 == id2 {
		t.Error("expected different companion IDs for same save_game_season_id from different leagues")
	}
}

func TestSeasonStore_GetByID_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := store.NewSeasonStore(db)
	_, err := s.GetByID(context.Background(), 999)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected ErrNoRows, got %v", err)
	}
}

func TestSeasonStore_GetBySeasonNum(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := store.NewSeasonStore(db)
	ctx := context.Background()

	_, _ = s.Upsert(ctx, store.Season{LeagueGUID: testLG, SaveGameSeasonID: 42, SeasonNum: 5, NumGames: 82})

	got, err := s.GetBySeasonNum(ctx, 5)
	if err != nil {
		t.Fatalf("GetBySeasonNum: %v", err)
	}
	if got.SeasonNum != 5 {
		t.Errorf("SeasonNum: got %d, want 5", got.SeasonNum)
	}
	if got.SaveGameSeasonID != 42 {
		t.Errorf("SaveGameSeasonID: got %d, want 42", got.SaveGameSeasonID)
	}
	if got.LeagueGUID != testLG {
		t.Errorf("LeagueGUID: got %q, want %q", got.LeagueGUID, testLG)
	}
}

func TestSeasonStore_GetBySeasonNum_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := store.NewSeasonStore(db)

	_, err := s.GetBySeasonNum(context.Background(), 99)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected ErrNoRows, got %v", err)
	}
}

func TestSeasonStore_List(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := store.NewSeasonStore(db)
	ctx := context.Background()

	_, _ = s.Upsert(ctx, store.Season{LeagueGUID: testLG, SaveGameSeasonID: 2, SeasonNum: 2, NumGames: 50})
	_, _ = s.Upsert(ctx, store.Season{LeagueGUID: testLG, SaveGameSeasonID: 1, SeasonNum: 1, NumGames: 50})

	seasons, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(seasons) != 2 {
		t.Fatalf("expected 2 seasons, got %d", len(seasons))
	}
	if seasons[0].SeasonNum != 1 {
		t.Errorf("expected seasons ordered by season_num, got %d first", seasons[0].SeasonNum)
	}
}
