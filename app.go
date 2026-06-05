package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"smb-tools/internal/config"
	internaldb "smb-tools/internal/db"
	"smb-tools/internal/models"
	"smb-tools/internal/service"
	"smb-tools/internal/store"
)

// legacyMigrationSourcePath is the placeholder path stored in franchise_source
// for franchises imported via legacy migration. It is not a real file path and
// must never be opened as one.
const legacyMigrationSourcePath = "(legacy migration)"

const docsURL = "https://tbrittain.github.io/smb-tools/"

// App is the Wails application struct. It is intentionally thin: it wires
// dependencies at startup and exposes bindings to the frontend. All business
// logic lives in internal/service and internal/store.
type App struct {
	ctx                 context.Context
	version             string
	dirs                *config.AppDirs
	registryDB          *sql.DB
	companionDB         *sql.DB // active franchise companion DB; nil if none selected
	activeFranchise     *models.Franchise
	franchiseStore      *store.FranchiseStore
	franchiseSourceStore *store.FranchiseSourceStore
	franchiseService        *service.FranchiseService
	importService           *service.ImportService
	syncService             *service.SyncService
	legacyMigrationService  *service.LegacyMigrationService
	// Per-franchise stores — initialised when a franchise is selected,
	// cleared when it is deselected or switched.
	snapshotStore         *store.SnapshotStore
	seasonStore           *store.SeasonStore
	seasonQueryStore      *store.SeasonQueryStore
	playerQueryStore      *store.PlayerQueryStore
	teamQueryStore        *store.TeamQueryStore
	leaderboardQueryStore *store.LeaderboardQueryStore
	awardStore            *store.AwardStore
	statRecordsService    *service.StatRecordsService
	logoStore             *store.LogoStore
	logoService           *service.LogoService
	mediaStore            *store.MediaStore
	mediaService          *service.MediaService
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

	registryDB, err := internaldb.OpenRegistry(ctx, dirs.RegistryPath)
	if err != nil {
		log.Printf("startup: opening registry DB: %v", err)
		return
	}
	a.registryDB = registryDB
	a.franchiseStore = store.NewFranchiseStore(registryDB)
	a.franchiseSourceStore = store.NewFranchiseSourceStore(registryDB)
	a.franchiseService = service.NewFranchiseService(dirs, a.franchiseStore, a.franchiseSourceStore)
	a.importService = service.NewImportService()
	a.legacyMigrationService = service.NewLegacyMigrationService()
	a.logoStore = store.NewLogoStore()
	a.logoService = service.NewLogoService(a.logoStore, dirs)
	a.mediaStore = store.NewMediaStore()
	a.mediaService = service.NewMediaService(a.mediaStore, dirs)

	a.setupMenu(ctx)
}

func (a *App) setupMenu(ctx context.Context) {
	appMenu := menu.NewMenu()
	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("Open App Data Directory", nil, func(_ *menu.CallbackData) {
		if a.dirs == nil {
			return
		}
		if err := openDirectory(a.dirs.DataDir); err != nil {
			log.Printf("open app data dir: %v", err)
		}
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("View Documentation", nil, func(_ *menu.CallbackData) {
		runtime.BrowserOpenURL(ctx, docsURL)
	})
	runtime.MenuSetApplicationMenu(ctx, appMenu)
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

// OpenAppDataDir opens the smb-tools app data directory in the OS file manager.
func (a *App) OpenAppDataDir() error {
	if a.dirs == nil {
		return fmt.Errorf("app data directory not initialized")
	}
	return openDirectory(a.dirs.DataDir)
}

// requireCompanionDB returns an error if no franchise companion DB is open.
func (a *App) requireCompanionDB() error {
	if a.companionDB == nil {
		return fmt.Errorf("no active franchise selected")
	}
	return nil
}
