package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

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
	sourceMu      sync.Mutex
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
//
// leagueMode is immutable once set — a franchise created from a season-mode
// save is always a season-mode franchise. Pass models.LeagueModeFranchise
// when the mode is not yet known (e.g. no save file configured yet).
func (s *FranchiseService) CreateFranchise(
	ctx context.Context,
	name string,
	version models.GameVersion,
	saveFilePath string,
	leagueGUID string,
	leagueMode models.LeagueMode,
) (models.Franchise, error) {
	slog.Info("FranchiseService.CreateFranchise", "name", name, "version", version, "leagueMode", leagueMode)
	if name == "" {
		return models.Franchise{}, fmt.Errorf("franchise name must not be empty")
	}
	if !version.Valid() {
		return models.Franchise{}, fmt.Errorf("invalid game version %q", version)
	}
	if leagueMode != models.LeagueModeFranchise && leagueMode != models.LeagueModeSeason {
		return models.Franchise{}, fmt.Errorf("unsupported league mode %q", leagueMode)
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
		LeagueMode:  leagueMode,
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

	slog.Info("FranchiseService.CreateFranchise: created", "id", id)
	return f, nil
}

// AddSource registers a new save game source after confirming it is not
// already connected to the franchise by normalized path or league GUID.
func (s *FranchiseService) AddSource(
	ctx context.Context,
	franchiseID string,
	saveFilePath string,
	leagueGUID string,
	seasonOffset int,
) error {
	s.sourceMu.Lock()
	defer s.sourceMu.Unlock()

	existingSources, err := s.sources.ListByFranchise(ctx, franchiseID)
	if err != nil {
		return fmt.Errorf("checking existing franchise sources: %w", err)
	}

	normalizedPath, err := filepath.Abs(filepath.Clean(saveFilePath))
	if err != nil {
		return fmt.Errorf("normalizing save file path %q: %w", saveFilePath, err)
	}
	for _, source := range existingSources {
		existingPath, err := filepath.Abs(filepath.Clean(source.SaveFilePath))
		if err != nil {
			return fmt.Errorf("normalizing existing save file path %q: %w", source.SaveFilePath, err)
		}
		pathsMatch := normalizedPath == existingPath
		if runtime.GOOS == "windows" {
			pathsMatch = strings.EqualFold(normalizedPath, existingPath)
		}
		if pathsMatch {
			return fmt.Errorf("save file path is already connected to this franchise")
		}
		if strings.EqualFold(strings.TrimSpace(leagueGUID), strings.TrimSpace(source.LeagueGUID)) {
			return fmt.Errorf("league GUID is already connected to this franchise")
		}
	}

	if _, err := s.sources.Add(ctx, franchiseID, saveFilePath, leagueGUID, seasonOffset); err != nil {
		return fmt.Errorf("storing franchise source: %w", err)
	}
	return nil
}

// OpenFranchise opens the companion DB for the given franchise ID and returns
// the connection. The caller is responsible for closing it.
func (s *FranchiseService) OpenFranchise(ctx context.Context, id string) (*sql.DB, models.Franchise, error) {
	f, err := s.franchises.GetByID(ctx, id)
	if err != nil {
		return nil, models.Franchise{}, fmt.Errorf("looking up franchise %q: %w", id, err)
	}

	db, err := internaldb.OpenCompanion(ctx, s.dirs.CompanionDBPath(f.ID))
	if err != nil {
		return nil, models.Franchise{}, fmt.Errorf("opening companion DB for franchise %q: %w", id, err)
	}
	return db, f, nil
}

// DeleteFranchise removes the franchise and its sources from the registry,
// then deletes the companion DB file and snapshots directory from disk.
// This is irreversible.
func (s *FranchiseService) DeleteFranchise(ctx context.Context, id string) error {
	slog.Info("FranchiseService.DeleteFranchise", "id", id)
	if err := s.sources.DeleteByFranchise(ctx, id); err != nil {
		return fmt.Errorf("removing franchise sources: %w", err)
	}
	if err := s.franchises.Delete(ctx, id); err != nil {
		return fmt.Errorf("removing franchise from registry: %w", err)
	}
	if err := os.RemoveAll(s.dirs.FranchiseDir(id)); err != nil {
		return fmt.Errorf("removing franchise directory: %w", err)
	}
	slog.Info("FranchiseService.DeleteFranchise: deleted", "id", id)
	return nil
}
