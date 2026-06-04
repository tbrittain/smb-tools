package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"smb-tools/internal/service"
	"smb-tools/internal/store"
)

// BrowseMediaFile opens a file dialog filtered to supported image and video types
// and returns the selected file path, or "" if cancelled.
func (a *App) BrowseMediaFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select media file",
		Filters: []runtime.FileFilter{
			{DisplayName: "Images & Videos", Pattern: "*.png;*.jpg;*.jpeg;*.mp4;*.mov;*.avi"},
			{DisplayName: "Images", Pattern: "*.png;*.jpg;*.jpeg"},
			{DisplayName: "Videos", Pattern: "*.mp4;*.mov;*.avi"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("opening media file dialog: %w", err)
	}
	return path, nil
}

// UploadMedia copies the selected file to the franchise's media directory and
// records the given associations in a single transaction.
func (a *App) UploadMedia(req UploadMediaRequest) (MediaItemDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return MediaItemDTO{}, err
	}
	m, err := a.mediaService.UploadAndAssociate(
		a.ctx, a.companionDB, a.activeFranchise.ID,
		req.Name, req.Description, req.FilePath,
		req.TeamHistoryIDs, req.PlayerIDs,
	)
	if err != nil {
		return MediaItemDTO{}, err
	}
	item, err := a.mediaStore.GetMediaWithAssocs(a.ctx, a.companionDB, m.ID)
	if err != nil {
		return MediaItemDTO{}, err
	}
	return mediaItemToDTO(item), nil
}

// GetMediaForTeamSeason returns a paginated gallery of media items for a team-season.
func (a *App) GetMediaForTeamSeason(teamHistoryID int64, page, pageSize int) (MediaGalleryPageDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return MediaGalleryPageDTO{}, err
	}
	if pageSize <= 0 {
		pageSize = 24
	}
	offset := page * pageSize
	items, total, err := a.mediaStore.GetMediaForTeamSeason(a.ctx, a.companionDB, teamHistoryID, pageSize, offset)
	if err != nil {
		return MediaGalleryPageDTO{}, err
	}
	return a.toGalleryPageDTO(items, total, page, pageSize), nil
}

// GetMediaForPlayer returns a paginated gallery of media items for a player.
func (a *App) GetMediaForPlayer(playerID int64, page, pageSize int) (MediaGalleryPageDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return MediaGalleryPageDTO{}, err
	}
	if pageSize <= 0 {
		pageSize = 24
	}
	offset := page * pageSize
	items, total, err := a.mediaStore.GetMediaForPlayer(a.ctx, a.companionDB, playerID, pageSize, offset)
	if err != nil {
		return MediaGalleryPageDTO{}, err
	}
	return a.toGalleryPageDTO(items, total, page, pageSize), nil
}

// RemoveMediaAssociation removes one association for a media item. If it was the
// last association, the file and media record are deleted automatically.
// entityType must be "team_season" or "player".
func (a *App) RemoveMediaAssociation(mediaID string, entityType string, entityID int64) error {
	if err := a.requireCompanionDB(); err != nil {
		return err
	}
	return a.mediaService.RemoveAssociation(a.ctx, a.companionDB, a.activeFranchise.ID, mediaID, entityType, entityID)
}

// DeleteMediaEverywhere removes all associations and the file for a media item.
func (a *App) DeleteMediaEverywhere(mediaID string) error {
	if err := a.requireCompanionDB(); err != nil {
		return err
	}
	return a.mediaService.DeleteMediaEverywhere(a.ctx, a.companionDB, a.activeFranchise.ID, mediaID)
}

// SearchTeamsForMediaPicker returns teams whose name matches the query,
// for use in the media association picker.
func (a *App) SearchTeamsForMediaPicker(query string) ([]TeamPickerResultDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	results, err := a.mediaStore.SearchTeamsForPicker(a.ctx, a.companionDB, query)
	if err != nil {
		return nil, err
	}
	out := make([]TeamPickerResultDTO, len(results))
	for i, r := range results {
		out[i] = TeamPickerResultDTO{TeamID: r.TeamID, TeamName: r.TeamName}
	}
	return out, nil
}

// GetTeamSeasonsForMediaPicker returns all team-season (historyId, seasonNum) pairs
// for a team, for use in the second step of the media association picker.
func (a *App) GetTeamSeasonsForMediaPicker(teamID int64) ([]TeamSeasonPickerResultDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	results, err := a.mediaStore.GetTeamSeasonsForPicker(a.ctx, a.companionDB, teamID)
	if err != nil {
		return nil, err
	}
	out := make([]TeamSeasonPickerResultDTO, len(results))
	for i, r := range results {
		out[i] = TeamSeasonPickerResultDTO{TeamHistoryID: r.TeamHistoryID, SeasonNum: r.SeasonNum}
	}
	return out, nil
}

// GetAllMediaForTeam returns all media for all seasons of a team, grouped by season (most recent first).
func (a *App) GetAllMediaForTeam(teamID int64) ([]TeamSeasonMediaGroupDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	rows, err := a.mediaStore.GetAllMediaForTeam(a.ctx, a.companionDB, teamID)
	if err != nil {
		return nil, err
	}

	var groups []TeamSeasonMediaGroupDTO
	groupIdx := map[int64]int{} // teamHistoryID → index in groups slice
	for _, row := range rows {
		idx, ok := groupIdx[row.GroupTeamHistoryID]
		if !ok {
			groups = append(groups, TeamSeasonMediaGroupDTO{
				SeasonNum:     row.GroupSeasonNum,
				TeamHistoryID: row.GroupTeamHistoryID,
				TeamName:      row.GroupTeamName,
			})
			idx = len(groups) - 1
			groupIdx[row.GroupTeamHistoryID] = idx
		}
		groups[idx].Items = append(groups[idx].Items, mediaItemToDTO(row.MediaWithAssocs))
	}
	return groups, nil
}

func (a *App) toGalleryPageDTO(items []store.MediaWithAssocs, total, page, pageSize int) MediaGalleryPageDTO {
	dtos := make([]MediaItemDTO, len(items))
	for i, item := range items {
		dtos[i] = mediaItemToDTO(item)
	}
	return MediaGalleryPageDTO{
		Items:      dtos,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}
}

func mediaItemToDTO(item store.MediaWithAssocs) MediaItemDTO {
	tsAssocs := make([]MediaTeamSeasonAssocDTO, len(item.TeamSeasons))
	for i, ts := range item.TeamSeasons {
		tsAssocs[i] = MediaTeamSeasonAssocDTO{
			TeamHistoryID: ts.TeamHistoryID,
			TeamName:      ts.TeamName,
			SeasonNum:     ts.SeasonNum,
		}
	}
	pAssocs := make([]MediaPlayerAssocDTO, len(item.Players))
	for i, p := range item.Players {
		pAssocs[i] = MediaPlayerAssocDTO{
			PlayerID:  p.PlayerID,
			FirstName: p.FirstName,
			LastName:  p.LastName,
		}
	}
	return MediaItemDTO{
		ID:                    item.Media.ID,
		Name:                  item.Media.Name,
		Description:           item.Media.Description,
		MediaType:             string(item.Media.MediaType),
		URL:                   service.MediaVirtualURL(item.Media),
		UploadedAt:            item.Media.UploadedAt.Format("2006-01-02T15:04:05Z"),
		TotalAssociationCount: len(tsAssocs) + len(pAssocs),
		TeamSeasonAssocs:      tsAssocs,
		PlayerAssocs:          pAssocs,
	}
}

// ── Asset middleware ──────────────────────────────────────────────────────────

// logoAssetMiddleware returns a Wails AssetServer middleware that intercepts
// /_logos/{teamId}/{filename} and /_media/{filename} requests, serving them
// directly from the active franchise's assets directory. Using Middleware (not
// Handler) ensures these requests are handled before the Vite dev-proxy in
// wails dev mode.
func (a *App) logoAssetMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasPrefix(r.URL.Path, "/_logos/"):
				a.serveLogoAsset(w, r)
			case strings.HasPrefix(r.URL.Path, "/_media/"):
				a.serveMediaAsset(w, r)
			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}

func (a *App) serveLogoAsset(w http.ResponseWriter, r *http.Request) {
	if a.dirs == nil || a.activeFranchise == nil {
		http.NotFound(w, r)
		return
	}

	// Strip "/_logos/" prefix and split into exactly [teamId, filename].
	trimmed := strings.TrimPrefix(r.URL.Path, "/_logos/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		http.NotFound(w, r)
		return
	}
	teamIDStr, filename := parts[0], parts[1]

	// Reject any path traversal attempts.
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") ||
		strings.Contains(filename, "..") || strings.Contains(teamIDStr, "..") {
		http.NotFound(w, r)
		return
	}

	franchiseDir := a.dirs.FranchiseDir(a.activeFranchise.ID)
	fullPath := filepath.Join(franchiseDir, "assets", "logos", teamIDStr, filename)

	// Verify the resolved path stays within the franchise directory.
	if !strings.HasPrefix(fullPath, filepath.Clean(franchiseDir)+string(filepath.Separator)) {
		http.NotFound(w, r)
		return
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	default:
		http.NotFound(w, r)
		return
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_, _ = w.Write(data)
}

func (a *App) serveMediaAsset(w http.ResponseWriter, r *http.Request) {
	if a.dirs == nil || a.activeFranchise == nil {
		http.NotFound(w, r)
		return
	}

	filename := strings.TrimPrefix(r.URL.Path, "/_media/")
	if filename == "" {
		http.NotFound(w, r)
		return
	}

	// Reject path traversal.
	if strings.Contains(filename, "/") || strings.Contains(filename, "\\") ||
		strings.Contains(filename, "..") {
		http.NotFound(w, r)
		return
	}

	franchiseDir := a.dirs.FranchiseDir(a.activeFranchise.ID)
	fullPath := filepath.Join(franchiseDir, "assets", "media", filename)

	if !strings.HasPrefix(fullPath, filepath.Clean(franchiseDir)+string(filepath.Separator)) {
		http.NotFound(w, r)
		return
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".mp4":
		w.Header().Set("Content-Type", "video/mp4")
	case ".mov":
		w.Header().Set("Content-Type", "video/quicktime")
	case ".avi":
		w.Header().Set("Content-Type", "video/x-msvideo")
	default:
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, fullPath)
}
