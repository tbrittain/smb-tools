package store_test

import (
	"context"
	"testing"

	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

func TestSaveAndGetExportPresets(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewExportPresetStore(db)

	saved, err := s.SaveExportPreset(ctx, "My Preset", "batting_season", `{"columns":["player_name"]}`)
	if err != nil {
		t.Fatalf("SaveExportPreset: %v", err)
	}
	if saved.ID == "" {
		t.Error("saved preset has empty ID")
	}
	if saved.Name != "My Preset" {
		t.Errorf("Name: want %q, got %q", "My Preset", saved.Name)
	}
	if saved.DatasetID != "batting_season" {
		t.Errorf("DatasetID: want %q, got %q", "batting_season", saved.DatasetID)
	}
	if saved.ConfigJSON != `{"columns":["player_name"]}` {
		t.Errorf("ConfigJSON: want %q, got %q", `{"columns":["player_name"]}`, saved.ConfigJSON)
	}
	if saved.CreatedAt == "" {
		t.Error("CreatedAt should not be empty")
	}

	presets, err := s.GetExportPresets(ctx)
	if err != nil {
		t.Fatalf("GetExportPresets: %v", err)
	}
	if len(presets) != 1 {
		t.Fatalf("want 1 preset, got %d", len(presets))
	}
	if presets[0].ID != saved.ID {
		t.Errorf("preset ID mismatch: want %q, got %q", saved.ID, presets[0].ID)
	}
}

func TestGetExportPresets_EmptyReturnsSliceNotNil(t *testing.T) {
	db := testutil.NewTestDB(t)
	presets, err := store.NewExportPresetStore(db).GetExportPresets(context.Background())
	if err != nil {
		t.Fatalf("GetExportPresets: %v", err)
	}
	if presets == nil {
		t.Error("expected non-nil empty slice, got nil")
	}
	if len(presets) != 0 {
		t.Errorf("expected 0 presets, got %d", len(presets))
	}
}

func TestDeleteExportPreset(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewExportPresetStore(db)

	p, err := s.SaveExportPreset(ctx, "To Delete", "standings", `{}`)
	if err != nil {
		t.Fatalf("SaveExportPreset: %v", err)
	}

	if err := s.DeleteExportPreset(ctx, p.ID); err != nil {
		t.Fatalf("DeleteExportPreset: %v", err)
	}

	presets, err := s.GetExportPresets(ctx)
	if err != nil {
		t.Fatalf("GetExportPresets after delete: %v", err)
	}
	if len(presets) != 0 {
		t.Errorf("expected 0 presets after delete, got %d", len(presets))
	}
}

func TestDeleteExportPreset_NoopIfNotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	// Delete of a non-existent ID should not error.
	err := store.NewExportPresetStore(db).DeleteExportPreset(context.Background(), "nonexistent-id")
	if err != nil {
		t.Errorf("DeleteExportPreset nonexistent: want nil, got %v", err)
	}
}

func TestGetExportPresets_OrderedByCreatedAtDesc(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	s := store.NewExportPresetStore(db)

	first, err := s.SaveExportPreset(ctx, "First", "batting_season", `{}`)
	if err != nil {
		t.Fatalf("SaveExportPreset first: %v", err)
	}
	second, err := s.SaveExportPreset(ctx, "Second", "pitching_season", `{}`)
	if err != nil {
		t.Fatalf("SaveExportPreset second: %v", err)
	}

	presets, err := s.GetExportPresets(ctx)
	if err != nil {
		t.Fatalf("GetExportPresets: %v", err)
	}
	if len(presets) != 2 {
		t.Fatalf("want 2 presets, got %d", len(presets))
	}
	// Most recent first — second inserted should appear first.
	if presets[0].ID != second.ID {
		t.Errorf("expected second preset first (most recent); got %q", presets[0].Name)
	}
	if presets[1].ID != first.ID {
		t.Errorf("expected first preset second; got %q", presets[1].Name)
	}
}
