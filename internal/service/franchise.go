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
// It coordinates between the registry (FranchiseStore + FranchiseSourceStore),
// the per-franchise companion DB, and the filesystem (app directories).
type FranchiseService struct {
	dirs          *config.AppDirs
	franchises    *store.FranchiseStore
	sources       *store.FranchiseSourceStore
}

func NewFranchiseService(
	dirs *config.AppDirs,
	franchises *store.FranchiseStore,
	sources *store.FranchiseSourceStore,
) *FranchiseService {
	return &FranchiseService{dirs: dirs, franchises: franchises, sources: sources}
}

// CreateFranchise registers a new franchise, creates its directory structure,
// and initializes its companion database with migrations. Returns the new
// franchise record.
//
// If saveFilePath and leagueGUID are non-empty, an initial franchise_sources
// row is created with season_offset = 0. Leave them empty when the user has
// not yet configured a save file; they can be added later via AddSource.
func (s *FranchiseService) CreateFranchise(
	ctx context.Context,
	name string,
	version models.GameVersion,
	saveFilePath string,
	leagueGUID string,
) (models.Franchise, error) {
	if name == "" {
		return models.Franchise{}, fmt.Errorf("franchise name must not be empty")
	}
	if !version.Valid() {
		return models.Franchise{}, fmt.Errorf("invalid game version %q", version)
	}

	id := uuid.New().String()
	dbPath := s.dirs.CompanionDBPath(id)

	if err := s.dirs.EnsureFranchiseDirs(id); err != nil {
		return models.Franchise{}, fmt.Errorf("creating franchise directories: %w", err)
	}

	companionDB, err := internaldb.OpenCompanion(ctx, dbPath)
	if err != nil {
		_ = os.RemoveAll(s.dirs.FranchiseDir(id))
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
	if err := s.franchises.Create(ctx, f); err != nil {
		_ = os.RemoveAll(s.dirs.FranchiseDir(id))
		return models.Franchise{}, fmt.Errorf("registering franchise: %w", err)
	}

	if saveFilePath != "" && leagueGUID != "" {
		if _, err := s.sources.Add(ctx, id, saveFilePath, leagueGUID, 0); err != nil {
			// Best-effort cleanup — the franchise row exists but has no source.
			// Return error; caller should handle or surface to the user.
			return models.Franchise{}, fmt.Errorf("adding initial source: %w", err)
		}
	}

	return f, nil
}

// OpenFranchise opens the companion DB for the given franchise ID and returns
// the connection. The caller is responsible for closing it.
func (s *FranchiseService) OpenFranchise(ctx context.Context, id string) (*sql.DB, models.Franchise, error) {
	f, err := s.franchises.GetByID(ctx, id)
	if err != nil {
		return nil, models.Franchise{}, fmt.Errorf("looking up franchise %q: %w", id, err)
	}

	db, err := internaldb.OpenCompanion(ctx, f.DBPath)
	if err != nil {
		return nil, models.Franchise{}, fmt.Errorf("opening companion DB for franchise %q: %w", id, err)
	}
	return db, f, nil
}

// DeleteFranchise removes the franchise and its sources from the registry,
// then deletes the companion DB file and snapshots directory from disk.
// This is irreversible.
func (s *FranchiseService) DeleteFranchise(ctx context.Context, id string) error {
	if err := s.sources.DeleteByFranchise(ctx, id); err != nil {
		return fmt.Errorf("removing franchise sources: %w", err)
	}
	if err := s.franchises.Delete(ctx, id); err != nil {
		return fmt.Errorf("removing franchise from registry: %w", err)
	}
	if err := os.RemoveAll(s.dirs.FranchiseDir(id)); err != nil {
		return fmt.Errorf("removing franchise directory: %w", err)
	}
	return nil
}
