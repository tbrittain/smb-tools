package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

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

// FranchiseDTO is the data transfer object returned to the frontend for
// franchise list and selection operations.
type FranchiseDTO struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	GameVersion      string `json:"gameVersion"`
	HasActiveSource  bool   `json:"hasActiveSource"`
	HasLegacySource  bool   `json:"hasLegacySource"`
	ActiveSourcePath string `json:"activeSourcePath"` // empty when no source configured
	LastSynced       string `json:"lastSynced"`        // ISO-8601 or ""
	LastSeason       int    `json:"lastSeason"`        // 0 if never synced
}

// FranchiseSourceDTO represents one save game source associated with a franchise.
type FranchiseSourceDTO struct {
	ID           int64  `json:"id"`
	SaveFilePath string `json:"saveFilePath"`
	LeagueGUID   string `json:"leagueGUID"`
	SeasonOffset int    `json:"seasonOffset"`
	AddedAt      string `json:"addedAt"` // ISO-8601
	IsLegacy     bool   `json:"isLegacy"`
}

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

// ListFranchises returns all registered franchises enriched with active source info.
func (a *App) ListFranchises() ([]FranchiseDTO, error) {
	if a.franchiseStore == nil {
		return nil, fmt.Errorf("app not initialized")
	}
	franchises, err := a.franchiseStore.List(a.ctx)
	if err != nil {
		return nil, err
	}
	allSources, err := a.franchiseSourceStore.ListAll(a.ctx)
	if err != nil {
		return nil, err
	}
	// Build a map of franchiseID → active source (highest season_offset)
	activeSource := make(map[string]models.FranchiseSource)
	for _, src := range allSources {
		existing, ok := activeSource[src.FranchiseID]
		if !ok || src.SeasonOffset > existing.SeasonOffset {
			activeSource[src.FranchiseID] = src
		}
	}
	dtos := make([]FranchiseDTO, len(franchises))
	for i, f := range franchises {
		dtos[i] = franchiseToDTO(f, activeSource[f.ID])
	}
	return dtos, nil
}

// CreateFranchise creates a new franchise. saveFilePath and leagueGUID may be
// empty if the user wants to configure the save file later via SetInitialSource.
func (a *App) CreateFranchise(name, gameVersion, saveFilePath, leagueGUID string) (FranchiseDTO, error) {
	if a.franchiseService == nil {
		return FranchiseDTO{}, fmt.Errorf("app not initialized")
	}
	v := models.GameVersion(gameVersion)
	f, err := a.franchiseService.CreateFranchise(a.ctx, name, v, saveFilePath, leagueGUID)
	if err != nil {
		return FranchiseDTO{}, err
	}
	src, _ := a.franchiseSourceStore.GetActive(a.ctx, f.ID)
	return franchiseToDTO(f, src), nil
}

// SelectFranchise opens the companion DB for the given franchise and sets it
// as the active franchise. Closes the previously active companion DB if any.
func (a *App) SelectFranchise(id string) (FranchiseDTO, error) {
	if a.franchiseService == nil {
		return FranchiseDTO{}, fmt.Errorf("app not initialized")
	}

	// Close previous companion DB and clear per-franchise stores
	if a.companionDB != nil {
		if err := a.companionDB.Close(); err != nil {
			log.Printf("SelectFranchise: closing previous companion DB: %v", err)
		}
		a.companionDB = nil
		a.activeFranchise = nil
		a.syncService = nil
		a.snapshotStore = nil
		a.seasonStore = nil
		a.seasonQueryStore = nil
		a.playerQueryStore = nil
		a.teamQueryStore = nil
		a.leaderboardQueryStore = nil
		a.awardStore = nil
	}

	companionDB, f, err := a.franchiseService.OpenFranchise(a.ctx, id)
	if err != nil {
		return FranchiseDTO{}, err
	}
	a.companionDB = companionDB
	a.activeFranchise = &f
	a.snapshotStore = store.NewSnapshotStore(companionDB)
	a.seasonStore = store.NewSeasonStore(companionDB)
	snapshotSvc := service.NewSnapshotService(
		a.dirs.SnapshotsDir(id),
		a.snapshotStore,
	)
	a.syncService = service.NewSyncService(snapshotSvc, a.importService)
	a.seasonQueryStore = store.NewSeasonQueryStore(companionDB)
	a.playerQueryStore = store.NewPlayerQueryStore(companionDB)
	a.teamQueryStore = store.NewTeamQueryStore(companionDB)
	a.leaderboardQueryStore = store.NewLeaderboardQueryStore(companionDB)
	a.awardStore = store.NewAwardStore(companionDB)

	src, _ := a.franchiseSourceStore.GetActive(a.ctx, id)
	return franchiseToDTO(f, src), nil
}

// GetActiveFranchise returns the currently active franchise, or an empty DTO
// if none is selected.
func (a *App) GetActiveFranchise() FranchiseDTO {
	if a.activeFranchise == nil {
		return FranchiseDTO{}
	}
	src, _ := a.franchiseSourceStore.GetActive(a.ctx, a.activeFranchise.ID)
	return franchiseToDTO(*a.activeFranchise, src)
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

// SyncSeasonResult is the DTO returned after a successful sync.
type SyncSeasonResult struct {
	SeasonID     int64  `json:"seasonId"`
	SeasonNum    int    `json:"seasonNum"`
	Players      int    `json:"players"`
	Teams        int    `json:"teams"`
	Games        int    `json:"games"`
	PlayoffGames int    `json:"playoffGames"`
}

// SyncSeason reads the active franchise's current save file source, auto-detects
// the current season, and imports it into the companion database. Safe to call
// multiple times — existing data is replaced with the latest save game state.
func (a *App) SyncSeason() (SyncSeasonResult, error) {
	if a.activeFranchise == nil {
		return SyncSeasonResult{}, fmt.Errorf("no active franchise selected")
	}
	if a.companionDB == nil {
		return SyncSeasonResult{}, fmt.Errorf("companion database not open")
	}

	src, err := a.franchiseSourceStore.GetActive(a.ctx, a.activeFranchise.ID)
	if err != nil {
		return SyncSeasonResult{}, fmt.Errorf("no save file configured for this franchise — add one via franchise settings")
	}

	saveDB, tmpPath, err := internaldb.DecompressAndOpen(a.ctx, src.SaveFilePath)
	if err != nil {
		return SyncSeasonResult{}, fmt.Errorf("opening save game: %w", err)
	}
	defer func() {
		_ = saveDB.Close()
		removePath(tmpPath)
	}()

	reader := store.NewSqliteSaveGameReader(saveDB, "")

	result, err := a.syncService.SyncSeason(a.ctx, a.companionDB, reader, tmpPath, src.LeagueGUID, src.SeasonOffset)
	if err != nil {
		return SyncSeasonResult{}, err
	}
	runtime.LogInfof(a.ctx, "SyncSeason: imported season %d — %d players, %d teams, %d games",
		result.SeasonNum, result.Players, result.Teams, result.Games)

	if err := a.franchiseStore.RecordSync(a.ctx, a.activeFranchise.ID, result.SeasonNum); err != nil {
		log.Printf("SyncSeason: recording sync: %v", err)
	}

	return SyncSeasonResult{
		SeasonID:     result.SeasonID,
		SeasonNum:    result.SeasonNum,
		Players:      result.Players,
		Teams:        result.Teams,
		Games:        result.Games,
		PlayoffGames: result.PlayoffGames,
	}, nil
}

// ListSnapshots returns all save game snapshots for the active franchise,
// ordered by capture time ascending. Each entry includes a FileExists flag
// indicating whether the snapshot file is still present on disk.
func (a *App) ListSnapshots() ([]SnapshotDTO, error) {
	if a.activeFranchise == nil {
		return nil, fmt.Errorf("no active franchise selected")
	}
	if a.snapshotStore == nil {
		return nil, fmt.Errorf("companion database not open")
	}
	snaps, err := a.snapshotStore.List(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing snapshots: %w", err)
	}
	out := make([]SnapshotDTO, len(snaps))
	for i, sn := range snaps {
		absPath := a.dirs.SnapshotsDir(a.activeFranchise.ID) + "/" + string(sn.FileName)
		_, statErr := os.Stat(absPath)
		out[i] = SnapshotDTO{
			ID:            sn.ID,
			SeasonNum:     sn.SeasonNum,
			CapturedAt:    sn.CapturedAt.UTC().Format("2006-01-02T15:04:05Z"),
			FileSizeBytes: sn.FileSizeBytes,
			FileExists:    statErr == nil,
		}
	}
	return out, nil
}

// ReimportSeasonFromSnapshot reimports a specific season using a previously
// captured snapshot file. The season to reimport is identified by seasonNum
// (the companion DB display season number). The snapshot is identified by
// snapshotID. Awards for the season are left untouched.
func (a *App) ReimportSeasonFromSnapshot(snapshotID int64, seasonNum int) (ReimportSeasonResult, error) {
	if a.activeFranchise == nil {
		return ReimportSeasonResult{}, fmt.Errorf("no active franchise selected")
	}
	if a.companionDB == nil {
		return ReimportSeasonResult{}, fmt.Errorf("companion database not open")
	}

	snap, err := a.snapshotStore.GetByID(a.ctx, snapshotID)
	if err != nil {
		return ReimportSeasonResult{}, fmt.Errorf("looking up snapshot: %w", err)
	}

	snapshotPath := a.dirs.SnapshotsDir(a.activeFranchise.ID) + "/" + string(snap.FileName)

	season, err := a.seasonStore.GetBySeasonNum(a.ctx, seasonNum)
	if err != nil {
		return ReimportSeasonResult{}, fmt.Errorf("looking up season %d: %w", seasonNum, err)
	}

	src, err := a.franchiseSourceStore.GetByLeagueGUID(a.ctx, a.activeFranchise.ID, season.LeagueGUID)
	if err != nil {
		return ReimportSeasonResult{}, fmt.Errorf("looking up source for league %q: %w", season.LeagueGUID, err)
	}

	saveGameSeasonNum := seasonNum - src.SeasonOffset

	saveDB, err := internaldb.OpenSnapshot(a.ctx, snapshotPath)
	if err != nil {
		return ReimportSeasonResult{}, fmt.Errorf("opening snapshot: %w", err)
	}
	defer func() { _ = saveDB.Close() }()

	reader := store.NewSqliteSaveGameReader(saveDB, "")

	result, err := a.importService.ImportSeason(
		a.ctx, a.companionDB, reader,
		season.SaveGameSeasonID, saveGameSeasonNum,
		season.LeagueGUID, src.SeasonOffset,
	)
	if err != nil {
		return ReimportSeasonResult{}, fmt.Errorf("reimporting season %d: %w", seasonNum, err)
	}
	runtime.LogInfof(a.ctx, "ReimportSeasonFromSnapshot: reimported season %d from snapshot %d — %d players, %d teams, %d games",
		seasonNum, snapshotID, result.Players, result.Teams, result.Games)

	return ReimportSeasonResult{
		SeasonNum:    result.SeasonNum,
		Players:      result.Players,
		Teams:        result.Teams,
		Games:        result.Games,
		PlayoffGames: result.PlayoffGames,
	}, nil
}

// ---- Save file discovery and probing --------------------------------------

// SaveFileCandidateDTO represents a discovered .sav file with its probed
// league/franchise metadata, giving users enough context to identify which
// save file corresponds to their franchise.
type SaveFileCandidateDTO struct {
	Path           string `json:"path"`
	GameVersion    string `json:"gameVersion"`
	LeagueName     string `json:"leagueName"`
	NumSeasons     int    `json:"numSeasons"`
	// Mode is "franchise", "season", "elimination", or "none".
	// Only franchise mode saves may be associated with a franchise in this app.
	Mode           string `json:"mode"`
	IsFranchise    bool   `json:"isFranchise"` // convenience: Mode == "franchise"
	PlayerTeamName string `json:"playerTeamName"` // team the user controlled; "" if not franchise mode
	LeagueGUID     string `json:"leagueGUID"`
}

// GetSaveFileCandidates scans the default SMB save file locations for .sav
// files and probes each one to return franchise metadata. This gives the user
// enough information (league name, player team, season count) to identify the
// right save file without opening a file browser.
func (a *App) GetSaveFileCandidates() ([]SaveFileCandidateDTO, error) {
	candidates, err := config.DiscoverSaveFiles()
	if err != nil {
		runtime.LogWarningf(a.ctx, "GetSaveFileCandidates: discovery error: %v", err)
		return []SaveFileCandidateDTO{}, nil
	}

	runtime.LogInfof(a.ctx, "GetSaveFileCandidates: found %d league save file(s)", len(candidates))

	var out []SaveFileCandidateDTO
	for _, c := range candidates {
		runtime.LogDebugf(a.ctx, "GetSaveFileCandidates: probing %s", c.Path)
		dto := SaveFileCandidateDTO{
			Path:        c.Path,
			GameVersion: string(c.GameVersion),
		}
		leagues, err := a.probeLeaguesFromPath(c.Path)
		if err != nil {
			runtime.LogWarningf(a.ctx, "GetSaveFileCandidates: probe failed for %s: %v", c.Path, err)
			out = append(out, dto)
			continue
		}
		if len(leagues) > 0 {
			lg := leagues[0]
			dto.LeagueName     = lg.Name
			dto.NumSeasons     = lg.NumSeasons
			dto.Mode           = leagueMode(lg)
			dto.IsFranchise    = lg.Mode == models.LeagueModeFranchise
			dto.PlayerTeamName = lg.PlayerTeamName
			dto.LeagueGUID     = lg.GUID
			runtime.LogInfof(a.ctx, "GetSaveFileCandidates: %s -> mode=%s league=%q team=%q seasons=%d",
				c.Path, dto.Mode, dto.LeagueName, dto.PlayerTeamName, dto.NumSeasons)
		} else {
			runtime.LogWarningf(a.ctx, "GetSaveFileCandidates: %s -> no leagues found in save file", c.Path)
		}
		out = append(out, dto)
	}

	runtime.LogInfof(a.ctx, "GetSaveFileCandidates: returning %d candidate(s)", len(out))
	return out, nil
}

// ProbeFranchiseSaveFile probes the active source's save file for the given
// franchise and returns live metadata. Use this to show "game has N seasons"
// alongside the last-synced state in the franchise list.
func (a *App) ProbeFranchiseSaveFile(franchiseID string) (SaveFileCandidateDTO, error) {
	if a.franchiseSourceStore == nil {
		return SaveFileCandidateDTO{}, fmt.Errorf("app not initialized")
	}
	src, err := a.franchiseSourceStore.GetActive(a.ctx, franchiseID)
	if err != nil {
		return SaveFileCandidateDTO{}, nil // no source configured — not an error
	}
	f, err := a.franchiseStore.GetByID(a.ctx, franchiseID)
	if err != nil {
		return SaveFileCandidateDTO{}, fmt.Errorf("franchise not found: %w", err)
	}
	dto := SaveFileCandidateDTO{
		Path:        src.SaveFilePath,
		GameVersion: string(f.GameVersion),
	}
	leagues, err := a.probeLeaguesFromPath(src.SaveFilePath)
	if err != nil {
		return dto, nil // non-fatal — caller gets basic path info
	}
	for _, lg := range leagues {
		if lg.GUID == src.LeagueGUID || src.LeagueGUID == "" {
			dto.LeagueName     = lg.Name
			dto.NumSeasons     = lg.NumSeasons
			dto.Mode           = leagueMode(lg)
			dto.IsFranchise    = lg.Mode == models.LeagueModeFranchise
			dto.PlayerTeamName = lg.PlayerTeamName
			dto.LeagueGUID     = lg.GUID
			break
		}
	}
	return dto, nil
}

// ListFranchiseSources returns all save game sources for a franchise, ordered
// by season_offset ascending (oldest source first).
func (a *App) ListFranchiseSources(franchiseID string) ([]FranchiseSourceDTO, error) {
	if a.franchiseSourceStore == nil {
		return nil, fmt.Errorf("app not initialized")
	}
	sources, err := a.franchiseSourceStore.ListByFranchise(a.ctx, franchiseID)
	if err != nil {
		return nil, err
	}
	out := make([]FranchiseSourceDTO, len(sources))
	for i, s := range sources {
		out[i] = sourceToDTO(s)
	}
	return out, nil
}

// SetInitialSource adds or replaces the first (season_offset = 0) source for a
// franchise that has no source configured yet. If a source already exists this
// returns an error — use ReplaceActiveFranchiseSource or AddFranchiseSource instead.
func (a *App) SetInitialSource(franchiseID, saveFilePath, leagueGUID string) error {
	if a.franchiseSourceStore == nil {
		return fmt.Errorf("app not initialized")
	}
	_, err := a.franchiseSourceStore.Add(a.ctx, franchiseID, saveFilePath, leagueGUID, 0)
	if err != nil {
		return fmt.Errorf("setting initial source: %w", err)
	}
	return nil
}

// AddFranchiseSource registers a new save game source for a forked league.
// seasonOffset should be the number of seasons already recorded in the franchise
// before this fork (i.e., the current last_synced_season). After this call,
// SyncSeason will read from the new source.
func (a *App) AddFranchiseSource(franchiseID, saveFilePath, leagueGUID string, seasonOffset int) error {
	if a.franchiseSourceStore == nil {
		return fmt.Errorf("app not initialized")
	}
	_, err := a.franchiseSourceStore.Add(a.ctx, franchiseID, saveFilePath, leagueGUID, seasonOffset)
	if err != nil {
		return fmt.Errorf("adding franchise source: %w", err)
	}
	return nil
}

// ReplaceActiveFranchiseSource updates the save file path and league GUID of
// the active (highest season_offset) source in-place. Use this for corrections
// only — e.g., the save file was moved, or the wrong file was linked. This does
// NOT change season_offset or create a new source row.
func (a *App) ReplaceActiveFranchiseSource(franchiseID, saveFilePath, leagueGUID string) error {
	if a.franchiseSourceStore == nil {
		return fmt.Errorf("app not initialized")
	}
	src, err := a.franchiseSourceStore.GetActive(a.ctx, franchiseID)
	if err != nil {
		return fmt.Errorf("no source configured for franchise: %w", err)
	}
	return a.franchiseSourceStore.Replace(a.ctx, src.ID, saveFilePath, leagueGUID)
}

// BrowseSaveFile opens the OS file picker filtered to .sav files and returns
// the selected path. Returns "" if the user cancels.
func (a *App) BrowseSaveFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select SMB Save File",
		Filters: []runtime.FileFilter{
			{DisplayName: "SMB Save Files (*.sav)", Pattern: "*.sav"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("file dialog: %w", err)
	}
	return path, nil
}

// ProbeLeagues opens a save file and returns its league metadata. Use this
// when the user browses to a save file not covered by GetSaveFileCandidates
// (i.e., a file outside the default discovery paths).
func (a *App) ProbeLeagues(saveFilePath string) ([]SaveFileCandidateDTO, error) {
	leagues, err := a.probeLeaguesFromPath(saveFilePath)
	if err != nil {
		return nil, fmt.Errorf("probing save file: %w", err)
	}
	out := make([]SaveFileCandidateDTO, len(leagues))
	for i, lg := range leagues {
		out[i] = SaveFileCandidateDTO{
			Path:           saveFilePath,
			LeagueName:     lg.Name,
			NumSeasons:     lg.NumSeasons,
			Mode:           leagueMode(lg),
			IsFranchise:    lg.Mode == models.LeagueModeFranchise,
			PlayerTeamName: lg.PlayerTeamName,
			LeagueGUID:     lg.GUID,
		}
	}
	return out, nil
}

// probeLeaguesFromPath decompresses a save file and returns its leagues.
// Shared by GetSaveFileCandidates, ProbeFranchiseSaveFile, and ProbeLeagues.
// The decompressed temp file is deleted on return — nothing is persisted.
func (a *App) probeLeaguesFromPath(path string) ([]models.SaveGameLeague, error) {
	saveDB, tmpPath, err := internaldb.DecompressAndOpen(a.ctx, path)
	if err != nil {
		return nil, fmt.Errorf("decompressing %s: %w", path, err)
	}
	defer func() {
		_ = saveDB.Close()
		removePath(tmpPath)
	}()
	reader := store.NewSqliteSaveGameReader(saveDB, "")
	leagues, err := reader.GetLeagues(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("reading leagues from %s: %w", path, err)
	}
	return leagues, nil
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
		a.seasonQueryStore = nil
		a.playerQueryStore = nil
		a.teamQueryStore = nil
		a.leaderboardQueryStore = nil
		a.awardStore = nil
	}
	return a.franchiseService.DeleteFranchise(a.ctx, id)
}

// ---- Phase 5 query bindings ------------------------------------------------

func (a *App) requireCompanionDB() error {
	if a.companionDB == nil {
		return fmt.Errorf("no active franchise selected")
	}
	return nil
}

// GetSeasonList returns all synced seasons with champion information.
func (a *App) GetSeasonList() ([]SeasonSummaryDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	seasons, err := a.seasonQueryStore.ListWithChampion(a.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]SeasonSummaryDTO, len(seasons))
	for i, s := range seasons {
		out[i] = SeasonSummaryDTO{
			ID:                s.ID,
			SeasonNum:         s.SeasonNum,
			NumGames:          s.NumGames,
			ImportedAt:        s.ImportedAt.Format("2006-01-02T15:04:05Z"),
			ChampionTeamName:  s.ChampionTeamName,
			ChampionHistoryID: s.ChampionHistoryID,
		}
	}
	return out, nil
}

// GetStandings returns all teams' standings for the given season.
func (a *App) GetStandings(seasonID int64) ([]TeamStandingDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	rows, err := a.seasonQueryStore.GetStandings(a.ctx, seasonID)
	if err != nil {
		return nil, err
	}
	out := make([]TeamStandingDTO, len(rows))
	for i, r := range rows {
		out[i] = TeamStandingDTO{
			HistoryID:      r.HistoryID,
			TeamID:         r.TeamID,
			TeamName:       r.TeamName,
			DivisionName:   r.DivisionName,
			ConferenceName: r.ConferenceName,
			Wins:           r.Wins,
			Losses:         r.Losses,
			WinPct:         r.WinPct,
			GamesBack:      r.GamesBack,
			RunsFor:        r.RunsFor,
			RunsAgainst:    r.RunsAgainst,
			RunDiff:        r.RunDiff,
			PlayoffSeed:    r.PlayoffSeed,
		}
	}
	return out, nil
}

// GetSeasonStatLeaders returns the six title-leader categories for a season.
func (a *App) GetSeasonStatLeaders(seasonID int64) (StatLeadersDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return StatLeadersDTO{}, err
	}
	l, err := a.seasonQueryStore.GetSeasonStatLeaders(a.ctx, seasonID)
	if err != nil {
		return StatLeadersDTO{}, err
	}
	toDTO := func(sl *models.StatLeader) *StatLeaderDTO {
		if sl == nil {
			return nil
		}
		return &StatLeaderDTO{
			PlayerID:  sl.PlayerID,
			FirstName: sl.FirstName,
			LastName:  sl.LastName,
			TeamName:  sl.TeamName,
			StatValue: sl.StatValue,
		}
	}
	return StatLeadersDTO{
		SeasonNum:  l.SeasonNum,
		BA:         toDTO(l.BA),
		HR:         toDTO(l.HR),
		RBI:        toDTO(l.RBI),
		ERA:        toDTO(l.ERA),
		Wins:       toDTO(l.Wins),
		Strikeouts: toDTO(l.Strikeouts),
	}, nil
}

// GetCareerLeaders returns the top-5 all-time career leaders for each category.
func (a *App) GetCareerLeaders() (CareerLeadersDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return CareerLeadersDTO{}, err
	}
	cl, err := a.seasonQueryStore.GetCareerLeaders(a.ctx)
	if err != nil {
		return CareerLeadersDTO{}, err
	}
	toSlice := func(rows []models.CareerLeaderRow) []CareerLeaderDTO {
		out := make([]CareerLeaderDTO, len(rows))
		for i, r := range rows {
			out[i] = CareerLeaderDTO{
				PlayerID:      r.PlayerID,
				FirstName:     r.FirstName,
				LastName:      r.LastName,
				StatValue:     r.StatValue,
				SeasonsPlayed: r.SeasonsPlayed,
			}
		}
		return out
	}
	return CareerLeadersDTO{
		HR:         toSlice(cl.HR),
		Hits:       toSlice(cl.Hits),
		RBI:        toSlice(cl.RBI),
		Wins:       toSlice(cl.Wins),
		Strikeouts: toSlice(cl.Strikeouts),
		Saves:      toSlice(cl.Saves),
	}, nil
}

// SearchPlayers returns up to 50 players matching the query string.
func (a *App) SearchPlayers(query string) ([]PlayerSearchResultDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	results, err := a.playerQueryStore.SearchPlayers(a.ctx, query)
	if err != nil {
		return nil, err
	}
	out := make([]PlayerSearchResultDTO, len(results))
	for i, r := range results {
		out[i] = PlayerSearchResultDTO{
			PlayerID:      r.PlayerID,
			FirstName:     r.FirstName,
			LastName:      r.LastName,
			IsHallOfFamer: r.IsHallOfFamer,
			SeasonsPlayed: r.SeasonsPlayed,
			FirstSeason:   r.FirstSeason,
			LastSeason:    r.LastSeason,
		}
	}
	return out, nil
}

// SearchTeams returns up to 50 teams matching the query string.
func (a *App) SearchTeams(query string) ([]TeamSearchResultDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	results, err := a.teamQueryStore.SearchTeams(a.ctx, query)
	if err != nil {
		return nil, err
	}
	out := make([]TeamSearchResultDTO, len(results))
	for i, r := range results {
		out[i] = TeamSearchResultDTO{
			TeamID:      r.TeamID,
			TeamName:    r.TeamName,
			Seasons:     r.Seasons,
			FirstSeason: r.FirstSeason,
			LastSeason:  r.LastSeason,
		}
	}
	return out, nil
}

// GetPlayerCareer returns a player's bio and career regular-season totals.
// Rate stats are read from pre-computed career tables — no on-read computation.
func (a *App) GetPlayerCareer(playerID int64) (PlayerCareerDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return PlayerCareerDTO{}, err
	}
	career, err := a.playerQueryStore.GetPlayerCareer(a.ctx, playerID)
	if err != nil {
		return PlayerCareerDTO{}, err
	}
	return PlayerCareerDTO{
		PlayerID:      career.PlayerID,
		FirstName:     career.FirstName,
		LastName:      career.LastName,
		IsHallOfFamer: career.IsHallOfFamer,
		Batting:       battingToDTO(career.Batting),
		Pitching:      pitchingToDTO(career.Pitching),
	}, nil
}

// GetPlayerSeasonLog returns a player's season-by-season regular and playoff
// stats. Rate stats are read from stored columns — no on-read computation.
func (a *App) GetPlayerSeasonLog(playerID int64) ([]PlayerSeasonLogDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	rows, err := a.playerQueryStore.GetPlayerSeasonLog(a.ctx, playerID)
	if err != nil {
		return nil, err
	}
	out := make([]PlayerSeasonLogDTO, len(rows))
	for i, r := range rows {
		teams := make([]TeamRefDTO, len(r.Teams))
		for j, t := range r.Teams {
			teams[j] = TeamRefDTO{TeamID: t.TeamID, TeamHistoryID: t.TeamHistoryID, TeamName: t.TeamName, SortOrder: t.SortOrder}
		}

		var traits []string
		if r.TraitsJSON != "" {
			_ = json.Unmarshal([]byte(r.TraitsJSON), &traits)
		}
		if traits == nil {
			traits = []string{}
		}

		var pitches []string
		if r.PitchesJSON != "" {
			_ = json.Unmarshal([]byte(r.PitchesJSON), &pitches)
		}
		if pitches == nil {
			pitches = []string{}
		}

		out[i] = PlayerSeasonLogDTO{
			SeasonNum:         r.SeasonNum,
			SeasonID:          r.SeasonID,
			Teams:             teams,
			Age:               r.Age,
			Salary:            r.Salary,
			PrimaryPosition:   r.PrimaryPosition,
			SecondaryPosition: r.SecondaryPosition,
			PitcherRole:       r.PitcherRole,
			BatHand:           r.BatHand,
			ThrowHand:         r.ThrowHand,
			ChemistryType:     r.ChemistryType,
			Traits:            traits,
			Pitches:           pitches,
			Power:             r.Power,
			Contact:           r.Contact,
			Speed:             r.Speed,
			Fielding:          r.Fielding,
			Arm:               r.Arm,
			Velocity:          r.Velocity,
			Junk:              r.Junk,
			Accuracy:          r.Accuracy,
			Batting:           battingToDTO(r.Batting),
			Pitching:          pitchingToDTO(r.Pitching),
			PlayoffBatting:    battingToDTO(r.PlayoffBatting),
			PlayoffPitching:   pitchingToDTO(r.PlayoffPitching),
		}
	}
	return out, nil
}

// GetTeamHistory returns all seasons played by a team with champion flags.
func (a *App) GetTeamHistory(teamID int64) (TeamHistoryDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return TeamHistoryDTO{}, err
	}
	th, err := a.teamQueryStore.GetTeamHistory(a.ctx, teamID)
	if err != nil {
		return TeamHistoryDTO{}, err
	}
	seasons := make([]TeamSeasonSummaryDTO, len(th.Seasons))
	for i, s := range th.Seasons {
		seasons[i] = teamSeasonSummaryToDTO(s)
	}
	return TeamHistoryDTO{
		TeamID:   th.TeamID,
		GameGUID: th.GameGUID,
		Seasons:  seasons,
	}, nil
}

// GetTeamTopPlayers returns the top 25 all-time players for a team, ranked by
// cumulative smbWAR accumulated while with that team.
func (a *App) GetTeamTopPlayers(teamID int64) ([]TeamTopPlayerDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	players, err := a.teamQueryStore.GetTeamTopPlayers(a.ctx, teamID, 25)
	if err != nil {
		return nil, fmt.Errorf("getting top players for team %d: %w", teamID, err)
	}
	out := make([]TeamTopPlayerDTO, len(players))
	for i, p := range players {
		out[i] = TeamTopPlayerDTO{
			PlayerID:       p.PlayerID,
			FirstName:      p.FirstName,
			LastName:       p.LastName,
			IsHallOfFamer:  p.IsHallOfFamer,
			NumSeasons:     p.NumSeasons,
			SeasonNums:     parseSortedSeasonNums(p.SeasonNumsCSV),
			IsPitcher:      p.IsPitcher,
			Position:       p.PrimaryPosition,
			SmbWARWithTeam: p.TotalSmbWAR,
			AvgOpsPlus:     p.AvgOpsPlus,
			AvgEraPlus:     p.AvgEraPlus,
			Awards:         p.Awards,
		}
	}
	return out, nil
}

// GetTeamSeasonDetail returns the roster, schedule, and playoff results for one
// team season. Rate stats are computed on roster players before returning.
// Only teamHistoryID is required — seasonID is derived from the team summary.
func (a *App) GetTeamSeasonDetail(teamHistoryID int64) (TeamSeasonDetailDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return TeamSeasonDetailDTO{}, err
	}

	teamSummary, err := a.teamQueryStore.GetTeamSeasonSummaryByHistoryID(a.ctx, teamHistoryID)
	if err != nil {
		return TeamSeasonDetailDTO{}, fmt.Errorf("team summary: %w", err)
	}
	seasonID := teamSummary.SeasonID

	roster, err := a.teamQueryStore.GetTeamSeasonRoster(a.ctx, teamHistoryID)
	if err != nil {
		return TeamSeasonDetailDTO{}, fmt.Errorf("roster: %w", err)
	}

	schedule, err := a.teamQueryStore.GetTeamSeasonSchedule(a.ctx, teamHistoryID, seasonID)
	if err != nil {
		return TeamSeasonDetailDTO{}, fmt.Errorf("schedule: %w", err)
	}

	playoffs, err := a.teamQueryStore.GetTeamSeasonPlayoffSchedule(a.ctx, teamHistoryID, seasonID)
	if err != nil {
		return TeamSeasonDetailDTO{}, fmt.Errorf("playoff schedule: %w", err)
	}

	seriesLength, err := a.seasonQueryStore.GetPlayoffSeriesLength(a.ctx, seasonID)
	if err != nil {
		return TeamSeasonDetailDTO{}, fmt.Errorf("playoff series length: %w", err)
	}

	rosterDTOs := make([]RosterPlayerDTO, len(roster))
	for i, r := range roster {
		rosterDTOs[i] = rosterPlayerToDTO(r)
	}
	scheduleDTOs := make([]ScheduleGameDTO, len(schedule))
	for i, g := range schedule {
		scheduleDTOs[i] = scheduleGameToDTO(g)
	}
	playoffDTOs := make([]PlayoffGameDTO, len(playoffs))
	for i, g := range playoffs {
		playoffDTOs[i] = playoffGameToDTO(g)
	}

	return TeamSeasonDetailDTO{
		Team:                teamSeasonSummaryToDTO(teamSummary),
		Roster:              rosterDTOs,
		Schedule:            scheduleDTOs,
		Playoffs:            playoffDTOs,
		PlayoffSeriesLength: seriesLength,
	}, nil
}

// GetHistoricalTeams returns one aggregated row per team for the historical
// teams page, covering the inclusive season range [seasonStart, seasonEnd].
// Results are ordered by total wins descending.
func (a *App) GetHistoricalTeams(seasonStart, seasonEnd int) ([]HistoricalTeamDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	rows, err := a.teamQueryStore.GetHistoricalTeams(a.ctx, seasonStart, seasonEnd)
	if err != nil {
		return nil, err
	}
	out := make([]HistoricalTeamDTO, len(rows))
	for i, r := range rows {
		out[i] = HistoricalTeamDTO{
			TeamID:              r.TeamID,
			TeamName:            r.TeamName,
			HistoryID:           r.HistoryID,
			NumSeasons:          r.NumSeasons,
			FirstSeason:         r.FirstSeason,
			LastSeason:          r.LastSeason,
			Wins:                r.Wins,
			Losses:              r.Losses,
			WinPct:              r.WinPct,
			GamesOver500:        r.GamesOver500,
			PlayoffWins:         r.PlayoffWins,
			PlayoffLosses:       r.PlayoffLosses,
			PlayoffAppearances:  r.PlayoffAppearances,
			DivisionTitles:      r.DivisionTitles,
			ConferenceTitles:    r.ConferenceTitles,
			Championships:       r.Championships,
			ChampionshipDrought: r.ChampionshipDrought,
			RunsFor:             r.RunsFor,
			RunsAgainst:         r.RunsAgainst,
			TotalAB:             r.TotalAB,
			TotalHits:           r.TotalHits,
			TotalHR:             r.TotalHR,
			NumPlayers:          r.NumPlayers,
			NumHoF:              r.NumHoF,
			BA:                  r.BA,
			ERA:                 r.ERA,
		}
	}
	return out, nil
}

// ListAllTeamSeasons returns every team-season for the historical teams page.
func (a *App) ListAllTeamSeasons() ([]TeamSeasonListDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	rows, err := a.teamQueryStore.ListAllTeamSeasons(a.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]TeamSeasonListDTO, len(rows))
	for i, r := range rows {
		out[i] = TeamSeasonListDTO{
			SeasonNum:      r.SeasonNum,
			HistoryID:      r.HistoryID,
			TeamID:         r.TeamID,
			TeamName:       r.TeamName,
			ConferenceName: r.ConferenceName,
			DivisionName:   r.DivisionName,
			Wins:           r.Wins,
			Losses:         r.Losses,
			WinPct:         r.WinPct,
			RunsFor:        r.RunsFor,
			RunsAgainst:    r.RunsAgainst,
			PlayoffSeed:    r.PlayoffSeed,
			PlayoffWins:    r.PlayoffWins,
			PlayoffLosses:  r.PlayoffLosses,
			IsChampion:     r.IsChampion,
		}
	}
	return out, nil
}

// ---- Leaderboard query bindings --------------------------------------------

// GetBattingCareerLeaders returns career batting totals for all players matching
// the given filters. Rate stats are computed before returning. The full result
// set is returned; sorting and pagination are handled client-side.
func (a *App) GetBattingCareerLeaders(filters LeaderboardFiltersDTO) ([]BattingLeaderRowDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	rows, err := a.leaderboardQueryStore.GetBattingCareerLeaders(a.ctx, leaderboardFiltersToDomain(filters))
	if err != nil {
		return nil, err
	}
	out := make([]BattingLeaderRowDTO, len(rows))
	for i := range rows {
		service.ComputeBattingRates(&rows[i].CareerBattingStats)
		out[i] = battingCareerLeaderToDTO(rows[i])
	}
	return out, nil
}

// GetBattingSeasonLeaders returns a paginated page of per-season batting stats
// matching the given filters. Rate stats are read from stored columns.
func (a *App) GetBattingSeasonLeaders(filters LeaderboardFiltersDTO) (BattingLeaderPageDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return BattingLeaderPageDTO{}, err
	}
	rows, total, err := a.leaderboardQueryStore.GetBattingSeasonLeaders(a.ctx, leaderboardFiltersToDomain(filters))
	if err != nil {
		return BattingLeaderPageDTO{}, err
	}
	out := make([]BattingLeaderRowDTO, len(rows))
	for i := range rows {
		out[i] = battingSeasonLeaderToDTO(rows[i])
	}
	return BattingLeaderPageDTO{Rows: out, Total: total}, nil
}

// GetPitchingCareerLeaders returns career pitching totals for all players matching
// the given filters. Rate stats are computed before returning.
func (a *App) GetPitchingCareerLeaders(filters LeaderboardFiltersDTO) ([]PitchingLeaderRowDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	rows, err := a.leaderboardQueryStore.GetPitchingCareerLeaders(a.ctx, leaderboardFiltersToDomain(filters))
	if err != nil {
		return nil, err
	}
	out := make([]PitchingLeaderRowDTO, len(rows))
	for i := range rows {
		service.ComputePitchingRates(&rows[i].CareerPitchingStats)
		out[i] = pitchingCareerLeaderToDTO(rows[i])
	}
	return out, nil
}

// GetPitchingSeasonLeaders returns a paginated page of per-season pitching stats
// matching the given filters. Rate stats are read from stored columns.
func (a *App) GetPitchingSeasonLeaders(filters LeaderboardFiltersDTO) (PitchingLeaderPageDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return PitchingLeaderPageDTO{}, err
	}
	rows, total, err := a.leaderboardQueryStore.GetPitchingSeasonLeaders(a.ctx, leaderboardFiltersToDomain(filters))
	if err != nil {
		return PitchingLeaderPageDTO{}, err
	}
	out := make([]PitchingLeaderRowDTO, len(rows))
	for i := range rows {
		out[i] = pitchingSeasonLeaderToDTO(rows[i])
	}
	return PitchingLeaderPageDTO{Rows: out, Total: total}, nil
}

// ---- helpers ---------------------------------------------------------------

func leagueMode(lg models.SaveGameLeague) string { return lg.Mode.String() }

func removePath(p string) {
	if p != "" {
		_ = os.Remove(p)
	}
}

func franchiseToDTO(f models.Franchise, src models.FranchiseSource) FranchiseDTO {
	dto := FranchiseDTO{
		ID:          f.ID,
		Name:        f.Name,
		GameVersion: f.GameVersion.String(),
		HasActiveSource:  src.SaveFilePath != "" && src.SaveFilePath != legacyMigrationSourcePath,
		HasLegacySource:  src.SaveFilePath == legacyMigrationSourcePath,
		ActiveSourcePath: src.SaveFilePath,
	}
	if f.LastSyncedAt != nil {
		dto.LastSynced = f.LastSyncedAt.Format("2006-01-02T15:04:05Z")
	}
	if f.LastSyncedSeason != nil {
		dto.LastSeason = *f.LastSyncedSeason
	}
	return dto
}

func sourceToDTO(s models.FranchiseSource) FranchiseSourceDTO {
	return FranchiseSourceDTO{
		ID:           s.ID,
		SaveFilePath: s.SaveFilePath,
		LeagueGUID:   s.LeagueGUID,
		SeasonOffset: s.SeasonOffset,
		AddedAt:      s.AddedAt.Format("2006-01-02T15:04:05Z"),
		IsLegacy:     s.SaveFilePath == legacyMigrationSourcePath,
	}
}

// ── Awards bindings ───────────────────────────────────────────────────────────

// ListAwards returns all award definitions filtered by the playoff flag.
func (a *App) ListAwards(isPlayoff bool) ([]AwardDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	awards, err := a.awardStore.ListAwards(a.ctx, isPlayoff)
	if err != nil {
		return nil, err
	}
	out := make([]AwardDTO, len(awards))
	for i, aw := range awards {
		out[i] = awardToDTO(aw)
	}
	return out, nil
}

// ListAllAwards returns all award definitions regardless of playoff flag.
func (a *App) ListAllAwards() ([]AwardDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	awards, err := a.awardStore.ListAllAwards(a.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]AwardDTO, len(awards))
	for i, aw := range awards {
		out[i] = awardToDTO(aw)
	}
	return out, nil
}

// CreateCustomAward creates a user-defined award and returns it with its new ID.
func (a *App) CreateCustomAward(dto AwardDTO) (AwardDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return AwardDTO{}, err
	}
	m := models.Award{
		Name: dto.Name, OriginalName: dto.OriginalName,
		Importance: dto.Importance, OmitFromGroupings: dto.OmitFromGroupings,
		IsBattingAward: dto.IsBattingAward, IsPitchingAward: dto.IsPitchingAward,
		IsFieldingAward: dto.IsFieldingAward, IsPlayoffAward: dto.IsPlayoffAward,
		IsUserAssignable: dto.IsUserAssignable,
	}
	id, err := a.awardStore.CreateCustomAward(a.ctx, m)
	if err != nil {
		return AwardDTO{}, err
	}
	dto.ID = id
	dto.IsBuiltIn = false
	return dto, nil
}

// GetSeasonPlayerAwards returns all player-seasons for a season with their awards.
func (a *App) GetSeasonPlayerAwards(seasonID int64) ([]SeasonPlayerAwardRowDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	rows, err := a.awardStore.GetSeasonPlayerAwards(a.ctx, seasonID)
	if err != nil {
		return nil, err
	}
	out := make([]SeasonPlayerAwardRowDTO, len(rows))
	for i, r := range rows {
		awards := make([]AwardDTO, len(r.Awards))
		for j, aw := range r.Awards {
			awards[j] = awardToDTO(aw)
		}
		out[i] = SeasonPlayerAwardRowDTO{
			PlayerSeasonID: r.PlayerSeasonID,
			PlayerID:       r.PlayerID,
			FirstName:      r.FirstName,
			LastName:       r.LastName,
			TeamName:       r.TeamName,
			PrimaryPos:     r.PrimaryPos,
			PitcherRole:    r.PitcherRole,
			Awards:         awards,
		}
	}
	return out, nil
}

// GetPlayerCareerAwards returns awards grouped by season number for a player.
func (a *App) GetPlayerCareerAwards(playerID int64) (map[string][]AwardDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	bySeasonNum, err := a.awardStore.GetPlayerCareerAwards(a.ctx, playerID)
	if err != nil {
		return nil, err
	}
	out := make(map[string][]AwardDTO, len(bySeasonNum))
	for sn, awards := range bySeasonNum {
		key := fmt.Sprintf("%d", sn)
		dtos := make([]AwardDTO, len(awards))
		for i, aw := range awards {
			dtos[i] = awardToDTO(aw)
		}
		out[key] = dtos
	}
	return out, nil
}

// SetPlayerSeasonAwards replaces the user-assignable awards for one player-season.
func (a *App) SetPlayerSeasonAwards(req SetPlayerAwardsRequestDTO) error {
	if err := a.requireCompanionDB(); err != nil {
		return err
	}
	return a.awardStore.SetPlayerSeasonAwards(a.ctx, req.PlayerSeasonID, req.AwardIDs)
}

// ComputeSeasonStatLeaderAwards computes and stores auto-calculated stat title
// awards (BA, HR, RBI, ERA, W, K, Triple Crown) for the given season.
func (a *App) ComputeSeasonStatLeaderAwards(seasonID int64) error {
	if err := a.requireCompanionDB(); err != nil {
		return err
	}
	return a.awardStore.ComputeAndAssignStatLeaderAwards(a.ctx, seasonID)
}

// GetSeasonChampionTeamHistoryID returns the team_season_history_id of the
// playoff champion for the season, or nil if not yet determinable.
func (a *App) GetSeasonChampionTeamHistoryID(seasonID int64) (*int64, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	return a.awardStore.GetSeasonChampionTeam(a.ctx, seasonID)
}

// GetHoFCandidates returns a paginated list of retired players eligible for Hall
// of Fame induction. page is 1-based; pageSize controls rows per page;
// lastSeasons limits results to players whose last season falls within the past
// lastSeasons seasons.
func (a *App) GetHoFCandidates(page, pageSize, lastSeasons int) (HoFPageDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return HoFPageDTO{}, err
	}
	result, err := a.awardStore.GetHoFCandidates(a.ctx, page, pageSize, lastSeasons)
	if err != nil {
		return HoFPageDTO{}, err
	}
	items := make([]HoFCandidateDTO, len(result.Items))
	for i, c := range result.Items {
		items[i] = hofCandidateToDTO(c)
	}
	return HoFPageDTO{Items: items, Total: result.Total}, nil
}

// GetHoFInducted returns a paginated list of current Hall of Fame members,
// filtered and paginated with the same semantics as GetHoFCandidates.
func (a *App) GetHoFInducted(page, pageSize, lastSeasons int) (HoFPageDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return HoFPageDTO{}, err
	}
	result, err := a.awardStore.GetHoFInducted(a.ctx, page, pageSize, lastSeasons)
	if err != nil {
		return HoFPageDTO{}, err
	}
	items := make([]HoFCandidateDTO, len(result.Items))
	for i, c := range result.Items {
		items[i] = hofCandidateToDTO(c)
	}
	return HoFPageDTO{Items: items, Total: result.Total}, nil
}

// GetSeasonAwardCandidates returns all award delegation candidate groups for a season:
// top batters/pitchers overall, rookies, by team, and by position. Award IDs are
// pre-populated from existing assignments, or auto-suggested if none exist yet.
func (a *App) GetSeasonAwardCandidates(seasonID int64) (SeasonAwardCandidatesDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return SeasonAwardCandidatesDTO{}, err
	}
	m, err := a.awardStore.GetSeasonAwardCandidates(a.ctx, seasonID)
	if err != nil {
		return SeasonAwardCandidatesDTO{}, err
	}
	return seasonAwardCandidatesToDTO(m), nil
}

// SubmitSeasonAwards replaces user-assignable awards for all specified player-seasons
// in a single transaction. Players omitted from the request are not modified.
func (a *App) SubmitSeasonAwards(req SubmitSeasonAwardsDTO) error {
	if err := a.requireCompanionDB(); err != nil {
		return err
	}
	entries := make([]models.PlayerAwardEntry, len(req.PlayerAwards))
	for i, e := range req.PlayerAwards {
		entries[i] = models.PlayerAwardEntry{
			PlayerSeasonID: e.PlayerSeasonID,
			AwardIDs:       e.AwardIDs,
		}
	}
	return a.awardStore.SubmitMultiplePlayerAwards(a.ctx, entries)
}

// SetHallOfFamer updates the Hall of Fame status for a player.
func (a *App) SetHallOfFamer(playerID int64, isHoF bool) error {
	if err := a.requireCompanionDB(); err != nil {
		return err
	}
	return a.playerQueryStore.SetHallOfFamer(a.ctx, playerID, isHoF)
}

// ── Legacy migration ──────────────────────────────────────────────────────────

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
func (a *App) MigrateLegacyFranchise(
	dbPath string,
	legacyFranchiseID int,
	newFranchiseName string,
	gameVersion string,
) (MigrateLegacyResult, error) {
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
		return MigrateLegacyResult{}, fmt.Errorf("creating franchise: %w", err)
	}

	// cleanupFranchise deletes the newly created franchise if the migration fails,
	// preventing an orphaned registry entry and empty companion DB. The companion DB
	// must be closed before calling this on Windows (open files cannot be deleted).
	cleanupFranchise := func() {
		if delErr := a.franchiseService.DeleteFranchise(a.ctx, newFranchise.ID); delErr != nil {
			log.Printf("MigrateLegacyFranchise: cleanup after failure: %v", delErr)
		}
	}

	// Open the companion DB.
	companionDB, err := internaldb.OpenCompanion(a.ctx, newFranchise.DBPath)
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
		a.ctx, legacyDB, legacyFranchiseID, companionDB, leagueGUID,
	)
	if err != nil {
		_ = companionDB.Close()
		cleanupFranchise()
		return MigrateLegacyResult{}, fmt.Errorf("migrating franchise: %w", err)
	}

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

// parseSortedSeasonNums converts a comma-separated season number string
// (as returned by GROUP_CONCAT) into a sorted integer slice.
func parseSortedSeasonNums(csv string) []int {
	if csv == "" {
		return []int{}
	}
	parts := strings.Split(csv, ",")
	nums := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err == nil {
			nums = append(nums, n)
		}
	}
	sort.Ints(nums)
	return nums
}
