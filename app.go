package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"smb-tools/internal/config"
	"smb-tools/internal/db"
	"smb-tools/internal/models"
	"smb-tools/internal/service"
	"smb-tools/internal/store"
)

// FranchiseDTO is the data transfer object returned to the frontend for
// franchise list and selection operations.
type FranchiseDTO struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	GameVersion  string `json:"gameVersion"`
	SaveFilePath string `json:"saveFilePath"`
	LastSynced   string `json:"lastSynced"`   // ISO-8601 or ""
	LastSeason   int    `json:"lastSeason"`   // 0 if never synced
}

// App is the Wails application struct. It is intentionally thin: it wires
// dependencies at startup and exposes bindings to the frontend. All business
// logic lives in internal/service and internal/store.
type App struct {
	ctx              context.Context
	version          string
	dirs             *config.AppDirs
	registryDB       *sql.DB
	companionDB      *sql.DB     // active franchise companion DB; nil if none selected
	activeFranchise  *models.Franchise
	franchiseStore   *store.FranchiseStore
	franchiseService *service.FranchiseService
}

func NewApp(version string) *App {
	return &App{version: version}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	dirs, err := config.NewAppDirs()
	if err != nil {
		log.Printf("startup: resolving app directories: %v", err)
		return
	}
	a.dirs = dirs

	registryDB, err := db.OpenRegistry(ctx, dirs.RegistryPath)
	if err != nil {
		log.Printf("startup: opening registry DB: %v", err)
		return
	}
	a.registryDB = registryDB
	a.franchiseStore = store.NewFranchiseStore(registryDB)
	a.franchiseService = service.NewFranchiseService(dirs, a.franchiseStore)
}

func (a *App) shutdown(_ context.Context) {
	if a.companionDB != nil {
		if err := a.companionDB.Close(); err != nil {
			log.Printf("shutdown: closing companion DB: %v", err)
		}
	}
	if a.registryDB != nil {
		if err := a.registryDB.Close(); err != nil {
			log.Printf("shutdown: closing registry DB: %v", err)
		}
	}
}

// ---- Wails bindings --------------------------------------------------------

// GetVersion returns the running app version.
func (a *App) GetVersion() string { return a.version }

// ListFranchises returns all registered franchises.
func (a *App) ListFranchises() ([]FranchiseDTO, error) {
	if a.franchiseStore == nil {
		return nil, fmt.Errorf("app not initialized")
	}
	franchises, err := a.franchiseStore.List(a.ctx)
	if err != nil {
		return nil, err
	}
	dtos := make([]FranchiseDTO, len(franchises))
	for i, f := range franchises {
		dtos[i] = franchiseToDTO(f)
	}
	return dtos, nil
}

// CreateFranchise creates a new franchise with the given name and game version.
func (a *App) CreateFranchise(name string, gameVersion string) (FranchiseDTO, error) {
	if a.franchiseService == nil {
		return FranchiseDTO{}, fmt.Errorf("app not initialized")
	}
	v := models.GameVersion(gameVersion)
	f, err := a.franchiseService.CreateFranchise(a.ctx, name, v)
	if err != nil {
		return FranchiseDTO{}, err
	}
	return franchiseToDTO(f), nil
}

// SelectFranchise opens the companion DB for the given franchise and sets it
// as the active franchise. Closes the previously active companion DB if any.
func (a *App) SelectFranchise(id string) (FranchiseDTO, error) {
	if a.franchiseService == nil {
		return FranchiseDTO{}, fmt.Errorf("app not initialized")
	}

	// Close previous companion DB
	if a.companionDB != nil {
		if err := a.companionDB.Close(); err != nil {
			log.Printf("SelectFranchise: closing previous companion DB: %v", err)
		}
		a.companionDB = nil
		a.activeFranchise = nil
	}

	companionDB, f, err := a.franchiseService.OpenFranchise(a.ctx, id)
	if err != nil {
		return FranchiseDTO{}, err
	}
	a.companionDB = companionDB
	a.activeFranchise = &f
	return franchiseToDTO(f), nil
}

// GetActiveFranchise returns the currently active franchise, or an empty DTO
// if none is selected.
func (a *App) GetActiveFranchise() FranchiseDTO {
	if a.activeFranchise == nil {
		return FranchiseDTO{}
	}
	return franchiseToDTO(*a.activeFranchise)
}

// RenameFranchise updates the display name of a franchise.
func (a *App) RenameFranchise(id, newName string) error {
	if a.franchiseStore == nil {
		return fmt.Errorf("app not initialized")
	}
	if err := a.franchiseStore.Rename(a.ctx, id, newName); err != nil {
		return err
	}
	// Update in-memory active franchise if it's the one being renamed
	if a.activeFranchise != nil && a.activeFranchise.ID == id {
		a.activeFranchise.Name = newName
	}
	return nil
}

// DeleteFranchise removes a franchise and deletes its data directory.
// If the deleted franchise is currently active, it deselects it.
func (a *App) DeleteFranchise(id string) error {
	if a.franchiseService == nil {
		return fmt.Errorf("app not initialized")
	}
	// Deselect if active
	if a.activeFranchise != nil && a.activeFranchise.ID == id {
		if a.companionDB != nil {
			_ = a.companionDB.Close()
			a.companionDB = nil
		}
		a.activeFranchise = nil
	}
	return a.franchiseService.DeleteFranchise(a.ctx, id)
}

// ---- helpers ---------------------------------------------------------------

func franchiseToDTO(f models.Franchise) FranchiseDTO {
	dto := FranchiseDTO{
		ID:          f.ID,
		Name:        f.Name,
		GameVersion: f.GameVersion.String(),
		SaveFilePath: f.SaveFilePath,
	}
	if f.LastSyncedAt != nil {
		dto.LastSynced = f.LastSyncedAt.Format("2006-01-02T15:04:05Z")
	}
	if f.LastSyncedSeason != nil {
		dto.LastSeason = *f.LastSyncedSeason
	}
	return dto
}
