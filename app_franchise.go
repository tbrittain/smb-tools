package main

import (
	"fmt"
	"log/slog"
	"os"

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
	// LeagueMode is "franchise" or "season". Immutable once the franchise is created.
	LeagueMode       string `json:"leagueMode"`
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

// SyncSeasonResult is the DTO returned after a successful sync.
type SyncSeasonResult struct {
	SeasonID     int64  `json:"seasonId"`
	SeasonNum    int    `json:"seasonNum"`
	Players      int    `json:"players"`
	Teams        int    `json:"teams"`
	Games        int    `json:"games"`
	PlayoffGames int    `json:"playoffGames"`
}

// ListFranchises returns all registered franchises enriched with active source info.
func (a *App) ListFranchises() ([]FranchiseDTO, error) {
	if a.franchiseStore == nil {
		return nil, fmt.Errorf("app not initialized")
	}
	franchises, err := a.franchiseStore.List(a.ctx)
	if err != nil {
		slog.Error("ListFranchises: listing franchises", "err", err)
		return nil, err
	}
	allSources, err := a.franchiseSourceStore.ListAll(a.ctx)
	if err != nil {
		slog.Error("ListFranchises: listing sources", "err", err)
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
	slog.Info("CreateFranchise", "name", name, "gameVersion", gameVersion)
	if a.franchiseService == nil {
		return FranchiseDTO{}, fmt.Errorf("app not initialized")
	}
	v := models.GameVersion(gameVersion)
	leagueMode := models.LeagueModeFranchise
	if saveFilePath != "" && leagueGUID != "" {
		leagues, err := a.probeLeaguesFromPath(saveFilePath)
		if err != nil {
			return FranchiseDTO{}, fmt.Errorf("probing save file: %w", err)
		}
		found := false
		for _, lg := range leagues {
			if lg.GUID == leagueGUID {
				leagueMode = lg.Mode
				found = true
				break
			}
		}
		if !found {
			return FranchiseDTO{}, fmt.Errorf("league %q not found in save file", leagueGUID)
		}
	}
	f, err := a.franchiseService.CreateFranchise(a.ctx, name, v, saveFilePath, leagueGUID, leagueMode)
	if err != nil {
		slog.Error("CreateFranchise: failed", "err", err)
		return FranchiseDTO{}, err
	}
	slog.Info("CreateFranchise: created", "id", f.ID)
	src, _ := a.franchiseSourceStore.GetActive(a.ctx, f.ID)
	return franchiseToDTO(f, src), nil
}

// SelectFranchise opens the companion DB for the given franchise and sets it
// as the active franchise. Closes the previously active companion DB if any.
func (a *App) SelectFranchise(id string) (FranchiseDTO, error) {
	if a.franchiseService == nil {
		return FranchiseDTO{}, fmt.Errorf("app not initialized")
	}

	slog.Info("SelectFranchise", "id", id)
	// Close previous companion DB and clear per-franchise stores
	if a.companionDB != nil {
		if err := a.companionDB.Close(); err != nil {
			slog.Error("SelectFranchise: closing previous companion DB", "err", err)
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
		a.statRecordsService = nil
		a.exportStore = nil
		a.exportPresetStore = nil
	}

	companionDB, f, err := a.franchiseService.OpenFranchise(a.ctx, id)
	if err != nil {
		slog.Error("SelectFranchise: opening franchise", "id", id, "err", err)
		return FranchiseDTO{}, err
	}
	slog.Info("SelectFranchise: opened", "id", id, "name", f.Name)
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
	a.statRecordsService = service.NewStatRecordsService(store.NewStatRecordQueryStore(companionDB))
	a.exportStore = store.NewExportStore(companionDB)
	a.exportPresetStore = store.NewExportPresetStore(companionDB)

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
	slog.Info("RenameFranchise", "id", id, "newName", newName)
	if a.franchiseStore == nil {
		return fmt.Errorf("app not initialized")
	}
	if err := a.franchiseStore.Rename(a.ctx, id, newName); err != nil {
		slog.Error("RenameFranchise: failed", "id", id, "err", err)
		return err
	}
	// Update in-memory active franchise if it's the one being renamed
	if a.activeFranchise != nil && a.activeFranchise.ID == id {
		a.activeFranchise.Name = newName
	}
	return nil
}

// SyncSeason reads the active franchise's current save file source, auto-detects
// the current season, and imports it into the companion database. Safe to call
// multiple times — existing data is replaced with the latest save game state.
func (a *App) SyncSeason() (SyncSeasonResult, error) {
	slog.Info("SyncSeason: starting")
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
		slog.Error("SyncSeason: failed", "err", err)
		return SyncSeasonResult{}, err
	}
	slog.Info("SyncSeason: imported",
		"seasonNum", result.SeasonNum,
		"players", result.Players,
		"teams", result.Teams,
		"games", result.Games,
		"playoffGames", result.PlayoffGames,
	)

	if err := a.franchiseStore.RecordSync(a.ctx, a.activeFranchise.ID, result.SeasonNum); err != nil {
		slog.Error("SyncSeason: recording sync", "err", err)
	}
	a.statRecordsService.Invalidate()

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

	slog.Info("ReimportSeasonFromSnapshot", "seasonNum", seasonNum, "snapshotID", snapshotID)
	result, err := a.importService.ImportSeason(
		a.ctx, a.companionDB, reader,
		season.SaveGameSeasonID, saveGameSeasonNum,
		season.LeagueGUID, src.SeasonOffset,
	)
	if err != nil {
		slog.Error("ReimportSeasonFromSnapshot: failed", "seasonNum", seasonNum, "err", err)
		return ReimportSeasonResult{}, fmt.Errorf("reimporting season %d: %w", seasonNum, err)
	}
	slog.Info("ReimportSeasonFromSnapshot: complete",
		"seasonNum", seasonNum,
		"snapshotID", snapshotID,
		"players", result.Players,
		"teams", result.Teams,
		"games", result.Games,
	)
	a.statRecordsService.Invalidate()

	return ReimportSeasonResult{
		SeasonNum:    result.SeasonNum,
		Players:      result.Players,
		Teams:        result.Teams,
		Games:        result.Games,
		PlayoffGames: result.PlayoffGames,
	}, nil
}

// GetSaveFileCandidates scans the default SMB save file locations for .sav
// files and probes each one to return franchise metadata. This gives the user
// enough information (league name, player team, season count) to identify the
// right save file without opening a file browser.
func (a *App) GetSaveFileCandidates() ([]SaveFileCandidateDTO, error) {
	candidates, err := config.DiscoverSaveFiles()
	if err != nil {
		slog.Warn("GetSaveFileCandidates: discovery error", "err", err)
		return []SaveFileCandidateDTO{}, nil
	}

	slog.Info("GetSaveFileCandidates: found save files", "count", len(candidates))

	out := []SaveFileCandidateDTO{}
	for _, c := range candidates {
		slog.Debug("GetSaveFileCandidates: probing", "path", c.Path)
		dto := SaveFileCandidateDTO{
			Path:        c.Path,
			GameVersion: string(c.GameVersion),
		}
		leagues, err := a.probeLeaguesFromPath(c.Path)
		if err != nil {
			slog.Warn("GetSaveFileCandidates: probe failed", "path", c.Path, "err", err)
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
			slog.Info("GetSaveFileCandidates: probed",
				"path", c.Path,
				"mode", dto.Mode,
				"league", dto.LeagueName,
				"team", dto.PlayerTeamName,
				"seasons", dto.NumSeasons,
			)
		} else {
			slog.Warn("GetSaveFileCandidates: no leagues found", "path", c.Path)
		}
		out = append(out, dto)
	}

	slog.Info("GetSaveFileCandidates: returning candidates", "count", len(out))
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

// BrowseSaveDirectory opens the OS directory picker and scans the chosen
// directory (and its immediate subdirectories) for franchise save files,
// probing each for league metadata. This mirrors what GetSaveFileCandidates
// does for the default platform paths, giving the user the same scan-and-identify
// experience for a custom location. Returns an empty slice if the user cancels.
func (a *App) BrowseSaveDirectory() ([]SaveFileCandidateDTO, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Folder Containing SMB4 Save Files",
	})
	if err != nil {
		return nil, fmt.Errorf("directory dialog: %w", err)
	}
	if dir == "" {
		return []SaveFileCandidateDTO{}, nil
	}
	candidates := config.ScanDirShallow(dir, models.GameVersionSMB4)
	out := []SaveFileCandidateDTO{}
	for _, c := range candidates {
		dto := SaveFileCandidateDTO{
			Path:        c.Path,
			GameVersion: string(c.GameVersion),
		}
		leagues, err := a.probeLeaguesFromPath(c.Path)
		if err != nil {
			slog.Warn("BrowseSaveDirectory: probe failed", "path", c.Path, "err", err)
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
		}
		out = append(out, dto)
	}
	return out, nil
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
	slog.Info("DeleteFranchise", "id", id)
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
	if err := a.franchiseService.DeleteFranchise(a.ctx, id); err != nil {
		slog.Error("DeleteFranchise: failed", "id", id, "err", err)
		return err
	}
	slog.Info("DeleteFranchise: deleted", "id", id)
	return nil
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
		LeagueMode:  f.LeagueMode.String(),
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
