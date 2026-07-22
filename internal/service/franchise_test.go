package service_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"smb-tools/internal/config"
	"smb-tools/internal/models"
	"smb-tools/internal/service"
	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

func newTestFranchiseService(t *testing.T) (*service.FranchiseService, *store.FranchiseStore) {
	svc, franchiseStore, _ := newTestFranchiseServiceWithSourceStore(t)
	return svc, franchiseStore
}

func newTestFranchiseServiceWithSourceStore(
	t *testing.T,
) (*service.FranchiseService, *store.FranchiseStore, *store.FranchiseSourceStore) {
	t.Helper()
	registryDB := testutil.NewTestRegistryDB(t)
	franchiseStore := store.NewFranchiseStore(registryDB)
	sourceStore := store.NewFranchiseSourceStore(registryDB)

	dirs := &config.AppDirs{
		DataDir:       t.TempDir(),
		FranchisesDir: t.TempDir(),
	}
	dirs.RegistryPath = dirs.DataDir + "/registry.db"

	svc := service.NewFranchiseService(dirs, franchiseStore, sourceStore)
	return svc, franchiseStore, sourceStore
}

func TestFranchiseService_CreateFranchise(t *testing.T) {
	svc, fs := newTestFranchiseService(t)
	ctx := context.Background()

	f, err := svc.CreateFranchise(ctx, "Test League", models.GameVersionSMB4, "", "", models.LeagueModeFranchise)
	if err != nil {
		t.Fatalf("CreateFranchise: %v", err)
	}
	if f.ID == "" {
		t.Error("expected non-empty franchise ID")
	}
	if f.Name != "Test League" {
		t.Errorf("name: got %q, want %q", f.Name, "Test League")
	}
	if f.GameVersion != models.GameVersionSMB4 {
		t.Errorf("game_version: got %q, want %q", f.GameVersion, models.GameVersionSMB4)
	}
	if f.LeagueMode != models.LeagueModeFranchise {
		t.Errorf("league_mode: got %q, want %q", f.LeagueMode, models.LeagueModeFranchise)
	}

	// Verify it's in the registry
	list, _ := fs.List(ctx)
	if len(list) != 1 {
		t.Errorf("expected 1 franchise in registry, got %d", len(list))
	}
}

func TestFranchiseService_CreateFranchise_SeasonMode(t *testing.T) {
	svc, fs := newTestFranchiseService(t)
	ctx := context.Background()

	f, err := svc.CreateFranchise(ctx, "Season League", models.GameVersionSMB4, "", "", models.LeagueModeSeason)
	if err != nil {
		t.Fatalf("CreateFranchise: %v", err)
	}
	if f.LeagueMode != models.LeagueModeSeason {
		t.Errorf("league_mode: got %q, want %q", f.LeagueMode, models.LeagueModeSeason)
	}

	got, err := fs.GetByID(ctx, f.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.LeagueMode != models.LeagueModeSeason {
		t.Errorf("round-tripped league_mode: got %q, want %q", got.LeagueMode, models.LeagueModeSeason)
	}
}

func TestFranchiseService_CreateFranchise_InvalidLeagueMode(t *testing.T) {
	svc, _ := newTestFranchiseService(t)
	_, err := svc.CreateFranchise(context.Background(), "My Franchise", models.GameVersionSMB4, "", "", models.LeagueModeElimination)
	if err == nil {
		t.Error("expected error for unsupported league mode")
	}
}

func TestFranchiseService_CreateFranchise_EmptyName(t *testing.T) {
	svc, _ := newTestFranchiseService(t)
	_, err := svc.CreateFranchise(context.Background(), "", models.GameVersionSMB4, "", "", models.LeagueModeFranchise)
	if err == nil {
		t.Error("expected error for empty franchise name")
	}
}

func TestFranchiseService_CreateFranchise_InvalidVersion(t *testing.T) {
	svc, _ := newTestFranchiseService(t)
	_, err := svc.CreateFranchise(context.Background(), "My Franchise", "smb5", "", "", models.LeagueModeFranchise)
	if err == nil {
		t.Error("expected error for invalid game version")
	}
}

func TestFranchiseService_AddSource_RejectsDuplicates(t *testing.T) {
	svc, _, _ := newTestFranchiseServiceWithSourceStore(t)
	ctx := context.Background()
	dir := t.TempDir()
	originalPath := dir + string(filepath.Separator) + "league-original.sav"
	originalGUID := "b9b0f849-480c-4b01-b9d8-cb632c739a9b"
	f, err := svc.CreateFranchise(
		ctx,
		"Test",
		models.GameVersionSMB4,
		originalPath,
		originalGUID,
		models.LeagueModeFranchise,
	)
	if err != nil {
		t.Fatalf("CreateFranchise: %v", err)
	}

	normalizedDuplicatePath := dir + string(filepath.Separator) + "unused" +
		string(filepath.Separator) + ".." + string(filepath.Separator) + "league-original.sav"
	tests := []struct {
		name       string
		path       string
		leagueGUID string
		wantError  string
	}{
		{
			name:       "exact path and GUID",
			path:       originalPath,
			leagueGUID: originalGUID,
			wantError:  "save file path is already connected",
		},
		{
			name:       "same GUID from copied file",
			path:       dir + string(filepath.Separator) + "league-copy.sav",
			leagueGUID: originalGUID,
			wantError:  "league GUID is already connected",
		},
		{
			name:       "normalized path with different metadata",
			path:       normalizedDuplicatePath,
			leagueGUID: "2fc4219b-9da9-4e62-b6aa-a962f9d677ee",
			wantError:  "save file path is already connected",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.AddSource(ctx, f.ID, tt.path, tt.leagueGUID, 3)
			if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("AddSource error = %v, want error containing %q", err, tt.wantError)
			}
		})
	}
}

func TestFranchiseService_AddSource_AcceptsDistinctSourceAndOffset(t *testing.T) {
	svc, _, sourceStore := newTestFranchiseServiceWithSourceStore(t)
	ctx := context.Background()
	dir := t.TempDir()
	f, err := svc.CreateFranchise(
		ctx,
		"Test",
		models.GameVersionSMB4,
		dir+string(filepath.Separator)+"league-original.sav",
		"b9b0f849-480c-4b01-b9d8-cb632c739a9b",
		models.LeagueModeFranchise,
	)
	if err != nil {
		t.Fatalf("CreateFranchise: %v", err)
	}

	const seasonOffset = 3
	if err := svc.AddSource(
		ctx,
		f.ID,
		dir+string(filepath.Separator)+"league-fork.sav",
		"2fc4219b-9da9-4e62-b6aa-a962f9d677ee",
		seasonOffset,
	); err != nil {
		t.Fatalf("AddSource: %v", err)
	}

	sources, err := sourceStore.ListByFranchise(ctx, f.ID)
	if err != nil {
		t.Fatalf("ListByFranchise: %v", err)
	}
	if len(sources) != 2 {
		t.Fatalf("source count = %d, want 2", len(sources))
	}
	if sources[1].SeasonOffset != seasonOffset {
		t.Errorf("season offset = %d, want %d", sources[1].SeasonOffset, seasonOffset)
	}
}

func TestFranchiseService_OpenFranchise(t *testing.T) {
	svc, _ := newTestFranchiseService(t)
	ctx := context.Background()

	f, err := svc.CreateFranchise(ctx, "Test", models.GameVersionSMB4, "", "", models.LeagueModeFranchise)
	if err != nil {
		t.Fatalf("CreateFranchise: %v", err)
	}

	db, got, err := svc.OpenFranchise(ctx, f.ID)
	if err != nil {
		t.Fatalf("OpenFranchise: %v", err)
	}
	defer func() { _ = db.Close() }()

	if got.ID != f.ID {
		t.Errorf("ID: got %q, want %q", got.ID, f.ID)
	}
	if err := db.PingContext(ctx); err != nil {
		t.Errorf("DB not usable after open: %v", err)
	}
}

func TestFranchiseService_DeleteFranchise(t *testing.T) {
	svc, fs := newTestFranchiseService(t)
	ctx := context.Background()

	f, _ := svc.CreateFranchise(ctx, "To Delete", models.GameVersionSMB4, "", "", models.LeagueModeFranchise)
	if err := svc.DeleteFranchise(ctx, f.ID); err != nil {
		t.Fatalf("DeleteFranchise: %v", err)
	}

	list, _ := fs.List(ctx)
	if len(list) != 0 {
		t.Errorf("expected 0 franchises after delete, got %d", len(list))
	}
}
