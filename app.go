package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"smb-tools/internal/config"
	internaldb "smb-tools/internal/db"
	"smb-tools/internal/logger"
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
	logFilePath         string   // path to current session log file; empty if logger failed
	logCleanup          func()   // closes the session log file
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
		fmt.Fprintf(os.Stderr, "startup: resolving app directories: %v\n", err)
		return
	}
	a.dirs = dirs

	cleanup, sessionFile, err := logger.Setup(dirs.LogsDir, a.version == "dev")
	if err != nil {
		fmt.Fprintf(os.Stderr, "startup: initializing logger: %v\n", err)
	} else {
		a.logCleanup = cleanup
		a.logFilePath = sessionFile
	}

	slog.Info("startup", "version", a.version)

	registryDB, err := internaldb.OpenRegistry(ctx, dirs.RegistryPath)
	if err != nil {
		slog.Error("startup: opening registry DB", "err", err)
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

	a.setupMenu(ctx, nil)

	go func() {
		info := a.CheckForUpdate()
		if info.Available {
			a.setupMenu(ctx, &info)
			runtime.EventsEmit(ctx, "updateAvailable", info)
		}
	}()
}

func (a *App) setupMenu(ctx context.Context, update *UpdateInfo) {
	appMenu := menu.NewMenu()

	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("Open App Data Directory", nil, func(_ *menu.CallbackData) {
		if a.dirs == nil {
			return
		}
		if err := openDirectory(a.dirs.DataDir); err != nil {
			slog.Error("open app data dir", "err", err)
		}
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("View Documentation", nil, func(_ *menu.CallbackData) {
		runtime.BrowserOpenURL(ctx, docsURL)
	})
	if update != nil && update.Available {
		fileMenu.AddSeparator()
		url := update.URL
		fileMenu.AddText("Update Available: "+update.Tag, nil, func(_ *menu.CallbackData) {
			runtime.BrowserOpenURL(ctx, url)
		})
	}

	helpMenu := appMenu.AddSubmenu("Help")
	helpMenu.AddText("Report a Bug", nil, func(_ *menu.CallbackData) {
		runtime.EventsEmit(ctx, "openBugReport")
	})

	runtime.MenuSetApplicationMenu(ctx, appMenu)
}

func (a *App) shutdown(_ context.Context) {
	if a.companionDB != nil {
		if err := a.companionDB.Close(); err != nil {
			slog.Error("shutdown: closing companion DB", "err", err)
		}
	}
	if a.registryDB != nil {
		if err := a.registryDB.Close(); err != nil {
			slog.Error("shutdown: closing registry DB", "err", err)
		}
	}
	slog.Info("shutdown")
	if a.logCleanup != nil {
		a.logCleanup()
	}
}

// ---- Wails bindings --------------------------------------------------------

// GetVersion returns the running app version.
func (a *App) GetVersion() string { return a.version }

// CheckForUpdate queries the GitHub releases API and reports whether a newer
// version is available. Returns an empty UpdateInfo when running a dev build
// or when the request fails — callers should treat an empty result as "no update".
func (a *App) CheckForUpdate() UpdateInfo {
	if a.version == "dev" {
		return UpdateInfo{}
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(a.ctx, http.MethodGet,
		"https://api.github.com/repos/tbrittain/smb-tools/releases/latest", nil)
	if err != nil {
		slog.Warn("CheckForUpdate: failed to build request", "err", err)
		return UpdateInfo{}
	}
	req.Header.Set("User-Agent", "smb-tools/"+a.version)

	resp, err := client.Do(req)
	if err != nil {
		slog.Warn("CheckForUpdate: request failed", "err", err)
		return UpdateInfo{}
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Warn("CheckForUpdate: close response body", "err", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("CheckForUpdate: unexpected status", "status", resp.StatusCode)
		return UpdateInfo{}
	}

	var release struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		slog.Warn("CheckForUpdate: failed to decode response", "err", err)
		return UpdateInfo{}
	}

	if release.TagName == "" || release.TagName == a.version {
		return UpdateInfo{}
	}

	return UpdateInfo{
		Available: true,
		Tag:       release.TagName,
		URL:       release.HTMLURL,
	}
}

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
