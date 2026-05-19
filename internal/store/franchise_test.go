package store_test

import (
	"context"
	"database/sql"
	"testing"

	"smb-tools/internal/models"
	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

func TestFranchiseStore_CreateAndList(t *testing.T) {
	db := testutil.NewTestRegistryDB(t)
	s := store.NewFranchiseStore(db)
	ctx := context.Background()

	f := models.Franchise{
		ID:          "test-id-1",
		Name:        "My Franchise",
		GameVersion: models.GameVersionSMB4,
		DBPath:      "/data/franchises/test-id-1/companion.db",
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

	_ = s.Create(ctx, models.Franchise{ID: "abc", Name: "Test", GameVersion: models.GameVersionSMB3, DBPath: "/x"})

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

func TestFranchiseStore_Rename(t *testing.T) {
	db := testutil.NewTestRegistryDB(t)
	s := store.NewFranchiseStore(db)
	ctx := context.Background()

	_ = s.Create(ctx, models.Franchise{ID: "r1", Name: "Original", GameVersion: models.GameVersionSMB4, DBPath: "/x"})

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

	_ = s.Create(ctx, models.Franchise{ID: "del1", Name: "ToDelete", GameVersion: models.GameVersionSMB4, DBPath: "/x"})
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
