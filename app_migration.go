package main

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	internaldb "smb-tools/internal/db"
	"smb-tools/internal/models"
	"smb-tools/internal/store"
)

// LegacyFranchiseDTO represents one franchise from a SmbExplorerCompanion database.
type LegacyFranchiseDTO struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	IsSmb3 bool   `json:"isSmb3"`
}

// MigrateLegacyResult is the outcome of a single franchise migration.
type MigrateLegacyResult struct {
	FranchiseID     string `json:"franchiseId"`
	FranchiseName   string `json:"franchiseName"`
	SeasonsMigrated int    `json:"seasonsMigrated"`
	TeamsMigrated   int    `json:"teamsMigrated"`
	PlayersMigrated int    `json:"playersMigrated"`
	AwardsMigrated  int    `json:"awardsMigrated"`
	LogosSkipped    int    `json:"logosSkipped"`
}

// DetectLegacyDB returns the path to SmbExplorerCompanion.db if it exists at
// the default Windows location (%LOCALAPPDATA%\SmbExplorerCompanion\).
// Returns an empty string on non-Windows platforms or when the file is absent.
func (a *App) DetectLegacyDB() string {
	return detectLegacyDBPath()
}

// BrowseLegacyDB opens an OS file picker filtered to *.db files and returns
// the selected path. Returns "" if the user cancels.
func (a *App) BrowseLegacyDB() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select SmbExplorerCompanion Database",
		Filters: []runtime.FileFilter{
			{DisplayName: "SQLite Database (*.db)", Pattern: "*.db"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("file dialog: %w", err)
	}
	return path, nil
}

// ListLegacyFranchises opens the legacy database at dbPath read-only and
// returns all franchises it contains. The caller selects which to migrate.
func (a *App) ListLegacyFranchises(dbPath string) ([]LegacyFranchiseDTO, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("dbPath must not be empty")
	}
	legacyDB, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("opening legacy DB: %w", err)
	}
	defer func() { _ = legacyDB.Close() }()

	reader, err := store.NewLegacyCompanionReader(a.ctx, legacyDB)
	if err != nil {
		return nil, fmt.Errorf("reading legacy DB: %w", err)
	}
	franchises, err := reader.ReadFranchises(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing legacy franchises: %w", err)
	}
	out := make([]LegacyFranchiseDTO, len(franchises))
	for i, f := range franchises {
		out[i] = LegacyFranchiseDTO{ID: f.ID, Name: f.Name, IsSmb3: f.IsSmb3}
	}
	return out, nil
}

// MigrateLegacyFranchise creates a new franchise, migrates all data for
// legacyFranchiseID from the legacy DB at dbPath, and returns a result summary.
//
// gameVersion must be "smb3" or "smb4". newFranchiseName is used as the
// new franchise name (typically pre-filled with the legacy franchise name).
// inningsPerGame must always be supplied by the caller (the legacy schema has
// no source data for it) and must be between store.MinInningsPerGame and
// store.MaxInningsPerGame inclusive — there is no default.
func (a *App) MigrateLegacyFranchise(
	dbPath string,
	legacyFranchiseID int,
	newFranchiseName string,
	gameVersion string,
	inningsPerGame int,
) (MigrateLegacyResult, error) {
	slog.Info("MigrateLegacyFranchise: starting", "legacyID", legacyFranchiseID, "name", newFranchiseName, "version", gameVersion)
	if a.franchiseService == nil || a.legacyMigrationService == nil {
		return MigrateLegacyResult{}, fmt.Errorf("app not initialized")
	}

	version := models.GameVersion(gameVersion)
	if !version.Valid() {
		return MigrateLegacyResult{}, fmt.Errorf("invalid game version %q", gameVersion)
	}

	// Create the new franchise (no live save file source — migration provides data).
	newFranchise, err := a.franchiseService.CreateFranchise(a.ctx, newFranchiseName, version, "", "")
	if err != nil {
		slog.Error("MigrateLegacyFranchise: creating franchise", "err", err)
		return MigrateLegacyResult{}, fmt.Errorf("creating franchise: %w", err)
	}
	slog.Debug("MigrateLegacyFranchise: franchise created", "id", newFranchise.ID)

	// cleanupFranchise deletes the newly created franchise if the migration fails,
	// preventing an orphaned registry entry and empty companion DB. The companion DB
	// must be closed before calling this on Windows (open files cannot be deleted).
	cleanupFranchise := func() {
		if delErr := a.franchiseService.DeleteFranchise(a.ctx, newFranchise.ID); delErr != nil {
			slog.Error("MigrateLegacyFranchise: cleanup after failure", "err", delErr)
		}
	}

	// Open the companion DB.
	companionDB, err := internaldb.OpenCompanion(a.ctx, a.dirs.CompanionDBPath(newFranchise.ID))
	if err != nil {
		cleanupFranchise()
		return MigrateLegacyResult{}, fmt.Errorf("opening companion DB: %w", err)
	}
	defer func() { _ = companionDB.Close() }()

	// Open legacy DB read-only.
	legacyDB, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		_ = companionDB.Close()
		cleanupFranchise()
		return MigrateLegacyResult{}, fmt.Errorf("opening legacy DB: %w", err)
	}
	defer func() { _ = legacyDB.Close() }()

	// Generate a synthetic league GUID for migrated seasons.
	leagueGUID := generateUUID()

	// Register the synthetic source so the franchise has a source entry.
	if _, err := a.franchiseSourceStore.Add(a.ctx, newFranchise.ID,
		legacyMigrationSourcePath, leagueGUID, 0); err != nil {
		_ = companionDB.Close()
		cleanupFranchise()
		return MigrateLegacyResult{}, fmt.Errorf("registering legacy source: %w", err)
	}

	migResult, err := a.legacyMigrationService.Migrate(
		a.ctx, legacyDB, legacyFranchiseID, companionDB, leagueGUID, inningsPerGame,
	)
	if err != nil {
		slog.Error("MigrateLegacyFranchise: migration failed", "legacyID", legacyFranchiseID, "err", err)
		_ = companionDB.Close()
		cleanupFranchise()
		return MigrateLegacyResult{}, fmt.Errorf("migrating franchise: %w", err)
	}
	slog.Info("MigrateLegacyFranchise: complete",
		"franchiseID", newFranchise.ID,
		"seasons", migResult.SeasonsMigrated,
		"players", migResult.PlayersMigrated,
	)

	return MigrateLegacyResult{
		FranchiseID:     newFranchise.ID,
		FranchiseName:   newFranchise.Name,
		SeasonsMigrated: migResult.SeasonsMigrated,
		TeamsMigrated:   migResult.TeamsMigrated,
		PlayersMigrated: migResult.PlayersMigrated,
		AwardsMigrated:  migResult.AwardsMigrated,
		LogosSkipped:    migResult.LogosSkipped,
	}, nil
}
