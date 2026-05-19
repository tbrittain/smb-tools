package store_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

func TestSeasonStore_UpsertAndGet(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := store.NewSeasonStore(db)
	ctx := context.Background()

	season := store.Season{ID: 100, SeasonNum: 1, NumGames: 50}
	if err := s.Upsert(ctx, season); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, err := s.GetByID(ctx, 100)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.SeasonNum != 1 || got.NumGames != 50 {
		t.Errorf("got %+v, want SeasonNum=1 NumGames=50", got)
	}
}

func TestSeasonStore_Upsert_Idempotent(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := store.NewSeasonStore(db)
	ctx := context.Background()

	_ = s.Upsert(ctx, store.Season{ID: 1, SeasonNum: 1, NumGames: 50})
	// Re-upsert with updated values
	if err := s.Upsert(ctx, store.Season{ID: 1, SeasonNum: 1, NumGames: 60}); err != nil {
		t.Fatalf("second Upsert: %v", err)
	}
	got, _ := s.GetByID(ctx, 1)
	if got.NumGames != 60 {
		t.Errorf("expected NumGames=60 after re-upsert, got %d", got.NumGames)
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

func TestSeasonStore_List(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := store.NewSeasonStore(db)
	ctx := context.Background()

	_ = s.Upsert(ctx, store.Season{ID: 2, SeasonNum: 2, NumGames: 50})
	_ = s.Upsert(ctx, store.Season{ID: 1, SeasonNum: 1, NumGames: 50})

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
