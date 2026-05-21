package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"os"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"smb-tools/internal/config"
	internaldb "smb-tools/internal/db"
	"smb-tools/internal/models"
	"smb-tools/internal/service"
	"smb-tools/internal/store"
)

// FranchiseDTO is the data transfer object returned to the frontend for
// franchise list and selection operations.
type FranchiseDTO struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	GameVersion      string `json:"gameVersion"`
	HasActiveSource  bool   `json:"hasActiveSource"`
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
	franchiseService    *service.FranchiseService
	importService       *service.ImportService
	syncService         *service.SyncService
	// Read-side query stores — initialised when a franchise is selected,
	// cleared when it is deselected or switched.
	seasonQueryStore      *store.SeasonQueryStore
	playerQueryStore      *store.PlayerQueryStore
	teamQueryStore        *store.TeamQueryStore
	leaderboardQueryStore *store.LeaderboardQueryStore
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

	// Close previous companion DB and clear query stores
	if a.companionDB != nil {
		if err := a.companionDB.Close(); err != nil {
			log.Printf("SelectFranchise: closing previous companion DB: %v", err)
		}
		a.companionDB = nil
		a.activeFranchise = nil
		a.syncService = nil
		a.seasonQueryStore = nil
		a.playerQueryStore = nil
		a.teamQueryStore = nil
		a.leaderboardQueryStore = nil
	}

	companionDB, f, err := a.franchiseService.OpenFranchise(a.ctx, id)
	if err != nil {
		return FranchiseDTO{}, err
	}
	a.companionDB = companionDB
	a.activeFranchise = &f
	snapshotSvc := service.NewSnapshotService(
		a.dirs.SnapshotsDir(id),
		store.NewSnapshotStore(companionDB),
	)
	a.syncService = service.NewSyncService(snapshotSvc, a.importService)
	a.seasonQueryStore = store.NewSeasonQueryStore(companionDB)
	a.playerQueryStore = store.NewPlayerQueryStore(companionDB)
	a.teamQueryStore = store.NewTeamQueryStore(companionDB)
	a.leaderboardQueryStore = store.NewLeaderboardQueryStore(companionDB)

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
// Rate stats (BA, OBP, ERA, etc.) are computed before returning.
func (a *App) GetPlayerCareer(playerID int64) (PlayerCareerDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return PlayerCareerDTO{}, err
	}
	career, err := a.playerQueryStore.GetPlayerCareer(a.ctx, playerID)
	if err != nil {
		return PlayerCareerDTO{}, err
	}
	if career.Batting != nil {
		service.ComputeBattingRates(career.Batting)
	}
	if career.Pitching != nil {
		service.ComputePitchingRates(career.Pitching)
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
// stats. Rate stats are computed on each row before returning.
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
		for _, b := range []*models.CareerBattingStats{r.Batting, r.PlayoffBatting} {
			if b != nil {
				service.ComputeBattingRates(b)
			}
		}
		for _, p := range []*models.CareerPitchingStats{r.Pitching, r.PlayoffPitching} {
			if p != nil {
				service.ComputePitchingRates(p)
			}
		}
		out[i] = PlayerSeasonLogDTO{
			SeasonNum:         r.SeasonNum,
			SeasonID:          r.SeasonID,
			TeamName:          r.TeamName,
			Age:               r.Age,
			Salary:            r.Salary,
			PrimaryPosition:   r.PrimaryPosition,
			SecondaryPosition: r.SecondaryPosition,
			PitcherRole:       r.PitcherRole,
			BatHand:           r.BatHand,
			ThrowHand:         r.ThrowHand,
			ChemistryType:     r.ChemistryType,
			TraitsJSON:        r.TraitsJSON,
			PitchesJSON:       r.PitchesJSON,
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
	for i := range roster {
		if roster[i].Batting != nil {
			service.ComputeBattingRates(roster[i].Batting)
		}
		if roster[i].Pitching != nil {
			service.ComputePitchingRates(roster[i].Pitching)
		}
	}

	schedule, err := a.teamQueryStore.GetTeamSeasonSchedule(a.ctx, teamHistoryID, seasonID)
	if err != nil {
		return TeamSeasonDetailDTO{}, fmt.Errorf("schedule: %w", err)
	}

	playoffs, err := a.teamQueryStore.GetTeamSeasonPlayoffSchedule(a.ctx, teamHistoryID, seasonID)
	if err != nil {
		return TeamSeasonDetailDTO{}, fmt.Errorf("playoff schedule: %w", err)
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
		Team:     teamSeasonSummaryToDTO(teamSummary),
		Roster:   rosterDTOs,
		Schedule: scheduleDTOs,
		Playoffs: playoffDTOs,
	}, nil
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

// GetBattingSeasonLeaders returns per-season batting stats for all player-seasons
// matching the given filters. Rate stats are computed before returning.
func (a *App) GetBattingSeasonLeaders(filters LeaderboardFiltersDTO) ([]BattingLeaderRowDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	rows, err := a.leaderboardQueryStore.GetBattingSeasonLeaders(a.ctx, leaderboardFiltersToDomain(filters))
	if err != nil {
		return nil, err
	}
	out := make([]BattingLeaderRowDTO, len(rows))
	for i := range rows {
		service.ComputeBattingRates(&rows[i].CareerBattingStats)
		out[i] = battingSeasonLeaderToDTO(rows[i])
	}
	return out, nil
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

// GetPitchingSeasonLeaders returns per-season pitching stats for all player-seasons
// matching the given filters. Rate stats are computed before returning.
func (a *App) GetPitchingSeasonLeaders(filters LeaderboardFiltersDTO) ([]PitchingLeaderRowDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	rows, err := a.leaderboardQueryStore.GetPitchingSeasonLeaders(a.ctx, leaderboardFiltersToDomain(filters))
	if err != nil {
		return nil, err
	}
	out := make([]PitchingLeaderRowDTO, len(rows))
	for i := range rows {
		service.ComputePitchingRates(&rows[i].CareerPitchingStats)
		out[i] = pitchingSeasonLeaderToDTO(rows[i])
	}
	return out, nil
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
		HasActiveSource:  src.SaveFilePath != "",
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
	}
}
