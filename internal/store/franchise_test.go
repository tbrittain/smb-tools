package store_test

import (
	"context"
	"database/sql"
	"testing"

	"smb-tools/internal/models"
	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

func franchiseFixture(id string) models.Franchise {
	return models.Franchise{
		ID:          id,
		Name:        "Fixture Franchise " + id,
		GameVersion: models.GameVersionSMB4,
	}
}

func TestFranchiseStore_CreateAndList(t *testing.T) {
	db := testutil.NewTestRegistryDB(t)
	s := store.NewFranchiseStore(db)
	ctx := context.Background()

	f := models.Franchise{
		ID:          "test-id-1",
		Name:        "My Franchise",
		GameVersion: models.GameVersionSMB4,
	}
	if err := s.Create(ctx, f); err != nil {
		t.Fatalf("Create: %v", err)
	}

	list, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 franchise, got %d", len(list))
	}
	if list[0].Name != f.Name {
		t.Errorf("name: got %q, want %q", list[0].Name, f.Name)
	}
	if list[0].GameVersion != f.GameVersion {
		t.Errorf("game_version: got %q, want %q", list[0].GameVersion, f.GameVersion)
	}
}

func TestFranchiseStore_GetByID(t *testing.T) {
	db := testutil.NewTestRegistryDB(t)
	s := store.NewFranchiseStore(db)
	ctx := context.Background()

	_ = s.Create(ctx, models.Franchise{ID: "abc", Name: "Test", GameVersion: models.GameVersionSMB3})

	got, err := s.GetByID(ctx, "abc")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ID != "abc" {
		t.Errorf("id: got %q, want %q", got.ID, "abc")
	}

	_, err = s.GetByID(ctx, "does-not-exist")
	if err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows for missing ID, got %v", err)
	}
}

func TestFranchiseStore_LeagueMode(t *testing.T) {
	db := testutil.NewTestRegistryDB(t)
	s := store.NewFranchiseStore(db)
	ctx := context.Background()

	if err := s.Create(ctx, models.Franchise{
		ID:          "season-1",
		Name:        "Season Franchise",
		GameVersion: models.GameVersionSMB4,
		LeagueMode:  models.LeagueModeSeason,
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := s.GetByID(ctx, "season-1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.LeagueMode != models.LeagueModeSeason {
		t.Errorf("league_mode: got %q, want %q", got.LeagueMode, models.LeagueModeSeason)
	}

	// Omitting LeagueMode defaults to franchise, mirroring the column default
	// that pre-migration registry rows fall back to.
	if err := s.Create(ctx, models.Franchise{
		ID:          "no-mode",
		Name:        "Unset Mode",
		GameVersion: models.GameVersionSMB4,
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err = s.GetByID(ctx, "no-mode")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.LeagueMode != models.LeagueModeFranchise {
		t.Errorf("league_mode default: got %q, want %q", got.LeagueMode, models.LeagueModeFranchise)
	}

	list, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 franchises, got %d", len(list))
	}
}

func TestFranchiseStore_Rename(t *testing.T) {
	db := testutil.NewTestRegistryDB(t)
	s := store.NewFranchiseStore(db)
	ctx := context.Background()

	_ = s.Create(ctx, models.Franchise{ID: "r1", Name: "Original", GameVersion: models.GameVersionSMB4})

	if err := s.Rename(ctx, "r1", "Renamed"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	f, _ := s.GetByID(ctx, "r1")
	if f.Name != "Renamed" {
		t.Errorf("expected name %q after rename, got %q", "Renamed", f.Name)
	}
}

func TestFranchiseStore_Delete(t *testing.T) {
	db := testutil.NewTestRegistryDB(t)
	s := store.NewFranchiseStore(db)
	ctx := context.Background()

	_ = s.Create(ctx, models.Franchise{ID: "del1", Name: "ToDelete", GameVersion: models.GameVersionSMB4})
	if err := s.Delete(ctx, "del1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := s.GetByID(ctx, "del1")
	if err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows after delete, got %v", err)
	}

	if err := s.Delete(ctx, "del1"); err != sql.ErrNoRows {
		t.Errorf("expected ErrNoRows deleting non-existent, got %v", err)
	}
}
