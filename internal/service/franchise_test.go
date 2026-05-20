package service_test

import (
	"context"
	"testing"

	"smb-tools/internal/config"
	"smb-tools/internal/models"
	"smb-tools/internal/service"
	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

func newTestFranchiseService(t *testing.T) (*service.FranchiseService, *store.FranchiseStore) {
	t.Helper()
	registryDB := testutil.NewTestRegistryDB(t)
	franchiseStore := store.NewFranchiseStore(registryDB)

	// Use t.TempDir() as the data root so filesystem operations are isolated
	dirs := &config.AppDirs{
		DataDir:       t.TempDir(),
		FranchisesDir: t.TempDir(),
	}
	dirs.RegistryPath = dirs.DataDir + "/registry.db"

	svc := service.NewFranchiseService(dirs, franchiseStore)
	return svc, franchiseStore
}

func TestFranchiseService_CreateFranchise(t *testing.T) {
	svc, fs := newTestFranchiseService(t)
	ctx := context.Background()

	f, err := svc.CreateFranchise(ctx, "Test League", models.GameVersionSMB4, "", "")
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

	// Verify it's in the registry
	list, _ := fs.List(ctx)
	if len(list) != 1 {
		t.Errorf("expected 1 franchise in registry, got %d", len(list))
	}
}

func TestFranchiseService_CreateFranchise_EmptyName(t *testing.T) {
	svc, _ := newTestFranchiseService(t)
	_, err := svc.CreateFranchise(context.Background(), "", models.GameVersionSMB4, "", "")
	if err == nil {
		t.Error("expected error for empty franchise name")
	}
}

func TestFranchiseService_CreateFranchise_InvalidVersion(t *testing.T) {
	svc, _ := newTestFranchiseService(t)
	_, err := svc.CreateFranchise(context.Background(), "My Franchise", "smb5", "", "")
	if err == nil {
		t.Error("expected error for invalid game version")
	}
}

func TestFranchiseService_DeleteFranchise(t *testing.T) {
	svc, fs := newTestFranchiseService(t)
	ctx := context.Background()

	f, _ := svc.CreateFranchise(ctx, "To Delete", models.GameVersionSMB4, "", "")
	if err := svc.DeleteFranchise(ctx, f.ID); err != nil {
		t.Fatalf("DeleteFranchise: %v", err)
	}

	list, _ := fs.List(ctx)
	if len(list) != 0 {
		t.Errorf("expected 0 franchises after delete, got %d", len(list))
	}
}
