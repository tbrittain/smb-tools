package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/google/uuid"

	"smb-tools/internal/config"
	internaldb "smb-tools/internal/db"
	"smb-tools/internal/models"
	"smb-tools/internal/store"
)

// FranchiseService orchestrates franchise creation and lifecycle.
// It coordinates between the registry (FranchiseStore), the per-franchise
// companion DB, and the filesystem (app directories).
type FranchiseService struct {
	dirs           *config.AppDirs
	franchiseStore *store.FranchiseStore
}

func NewFranchiseService(dirs *config.AppDirs, franchiseStore *store.FranchiseStore) *FranchiseService {
	return &FranchiseService{dirs: dirs, franchiseStore: franchiseStore}
}

// CreateFranchise registers a new franchise, creates its directory structure,
// and initializes its companion database with migrations. Returns the new
// franchise record.
func (s *FranchiseService) CreateFranchise(ctx context.Context, name string, version models.GameVersion) (models.Franchise, error) {
	if name == "" {
		return models.Franchise{}, fmt.Errorf("franchise name must not be empty")
	}
	if !version.Valid() {
		return models.Franchise{}, fmt.Errorf("invalid game version %q", version)
	}

	id := uuid.New().String()
	dbPath := s.dirs.CompanionDBPath(id)

	// Create the franchise directory structure before touching the registry
	if err := s.dirs.EnsureFranchiseDirs(id); err != nil {
		return models.Franchise{}, fmt.Errorf("creating franchise directories: %w", err)
	}

	// Initialize the companion DB (runs migrations)
	companionDB, err := internaldb.OpenCompanion(ctx, dbPath)
	if err != nil {
		_ = os.RemoveAll(s.dirs.FranchiseDir(id)) // best-effort cleanup
		return models.Franchise{}, fmt.Errorf("initializing companion DB: %w", err)
	}
	if err := companionDB.Close(); err != nil {
		return models.Franchise{}, fmt.Errorf("closing companion DB after initialization: %w", err)
	}

	f := models.Franchise{
		ID:          id,
		Name:        name,
		GameVersion: version,
		DBPath:      dbPath,
	}
	if err := s.franchiseStore.Create(ctx, f); err != nil {
		_ = os.RemoveAll(s.dirs.FranchiseDir(id))
		return models.Franchise{}, fmt.Errorf("registering franchise: %w", err)
	}
	return f, nil
}

// OpenFranchise opens the companion DB for the given franchise ID and returns
// the connection. The caller is responsible for closing it.
func (s *FranchiseService) OpenFranchise(ctx context.Context, id string) (*sql.DB, models.Franchise, error) {
	f, err := s.franchiseStore.GetByID(ctx, id)
	if err != nil {
		return nil, models.Franchise{}, fmt.Errorf("looking up franchise %q: %w", id, err)
	}

	db, err := internaldb.OpenCompanion(ctx, f.DBPath)
	if err != nil {
		return nil, models.Franchise{}, fmt.Errorf("opening companion DB for franchise %q: %w", id, err)
	}
	return db, f, nil
}

// DeleteFranchise removes the franchise from the registry. It also deletes
// the companion DB file and snapshots directory from disk.
// This is irreversible.
func (s *FranchiseService) DeleteFranchise(ctx context.Context, id string) error {
	if err := s.franchiseStore.Delete(ctx, id); err != nil {
		return fmt.Errorf("removing franchise from registry: %w", err)
	}
	// Best-effort filesystem cleanup — log but do not fail if removal fails
	if err := os.RemoveAll(s.dirs.FranchiseDir(id)); err != nil {
		return fmt.Errorf("removing franchise directory: %w", err)
	}
	return nil
}
