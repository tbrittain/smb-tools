package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"smb-tools/internal/config"
	"smb-tools/internal/models"
	"smb-tools/internal/store"
)

// MediaService handles media file I/O and coordinates with MediaStore.
type MediaService struct {
	mediaStore *store.MediaStore
	dirs       *config.AppDirs
}

func NewMediaService(mediaStore *store.MediaStore, dirs *config.AppDirs) *MediaService {
	return &MediaService{mediaStore: mediaStore, dirs: dirs}
}

// UploadAndAssociate copies the source file to the franchise's media directory,
// inserts a media row, and creates the requested associations atomically.
func (s *MediaService) UploadAndAssociate(
	ctx context.Context,
	db *sql.DB,
	franchiseID string,
	name, description, srcPath string,
	teamHistoryIDs []int64,
	playerIDs []int64,
) (models.Media, error) {
	ext, mediaType, err := validateMediaFile(srcPath)
	if err != nil {
		return models.Media{}, fmt.Errorf("invalid media file: %w", err)
	}

	mediaDir := s.dirs.MediaDir(franchiseID)
	if err := os.MkdirAll(mediaDir, 0o700); err != nil {
		return models.Media{}, fmt.Errorf("creating media directory: %w", err)
	}

	mediaID := uuid.New().String()
	filename := mediaID + "." + ext
	dstPath := filepath.Join(mediaDir, filename)

	if err := copyFile(srcPath, dstPath); err != nil {
		return models.Media{}, fmt.Errorf("copying media file: %w", err)
	}

	// Relative path stored in DB, forward-slashes for cross-platform consistency.
	relPath := filepath.ToSlash(filepath.Join("assets", "media", filename))

	now := time.Now().UTC()
	m := models.Media{
		ID:          mediaID,
		FilePath:    relPath,
		MediaType:   mediaType,
		Name:        name,
		Description: description,
		UploadedAt:  now,
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		_ = os.Remove(dstPath)
		return models.Media{}, fmt.Errorf("beginning transaction: %w", err)
	}

	if err := s.mediaStore.InsertMedia(ctx, tx, m); err != nil {
		_ = tx.Rollback()
		_ = os.Remove(dstPath)
		return models.Media{}, fmt.Errorf("recording media: %w", err)
	}

	for _, thID := range teamHistoryIDs {
		assoc := models.MediaTeamSeason{
			ID:            uuid.New().String(),
			MediaID:       mediaID,
			TeamHistoryID: thID,
		}
		if err := s.mediaStore.InsertTeamSeasonAssoc(ctx, tx, assoc); err != nil {
			_ = tx.Rollback()
			_ = os.Remove(dstPath)
			return models.Media{}, fmt.Errorf("recording team-season association: %w", err)
		}
	}

	for _, pID := range playerIDs {
		assoc := models.MediaPlayer{
			ID:       uuid.New().String(),
			MediaID:  mediaID,
			PlayerID: pID,
		}
		if err := s.mediaStore.InsertPlayerAssoc(ctx, tx, assoc); err != nil {
			_ = tx.Rollback()
			_ = os.Remove(dstPath)
			return models.Media{}, fmt.Errorf("recording player association: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		_ = os.Remove(dstPath)
		return models.Media{}, fmt.Errorf("committing media upload: %w", err)
	}

	return m, nil
}

// RemoveAssociation removes one association for the given media item.
// If no associations remain after removal, the file and media record are deleted.
func (s *MediaService) RemoveAssociation(
	ctx context.Context,
	db *sql.DB,
	franchiseID string,
	mediaID string,
	entityType string,
	entityID int64,
) error {
	switch entityType {
	case "team_season":
		if err := s.mediaStore.RemoveTeamSeasonAssoc(ctx, db, mediaID, entityID); err != nil {
			return err
		}
	case "player":
		if err := s.mediaStore.RemovePlayerAssoc(ctx, db, mediaID, entityID); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown entity type %q", entityType)
	}

	return s.cleanupIfOrphaned(ctx, db, franchiseID, mediaID)
}

// DeleteMediaEverywhere removes all associations and the file unconditionally.
func (s *MediaService) DeleteMediaEverywhere(
	ctx context.Context,
	db *sql.DB,
	franchiseID string,
	mediaID string,
) error {
	m, err := s.mediaStore.GetMediaByID(ctx, db, mediaID)
	if err != nil {
		return fmt.Errorf("fetching media for deletion: %w", err)
	}

	if err := s.mediaStore.DeleteMedia(ctx, db, mediaID); err != nil {
		return err
	}

	fullPath := filepath.Join(s.dirs.FranchiseDir(franchiseID), filepath.FromSlash(m.FilePath))
	_ = os.Remove(fullPath)
	return nil
}

// MediaVirtualURL constructs the virtual asset URL for a media item.
// FilePath is "assets/media/{filename}"; virtual URL is "/_media/{filename}".
func MediaVirtualURL(m models.Media) string {
	const prefix = "assets/media/"
	if len(m.FilePath) <= len(prefix) {
		return ""
	}
	return "/_media/" + m.FilePath[len(prefix):]
}

func (s *MediaService) cleanupIfOrphaned(ctx context.Context, db *sql.DB, franchiseID, mediaID string) error {
	tsCount, pCount, err := s.mediaStore.GetAssociationCounts(ctx, db, mediaID)
	if err != nil {
		return fmt.Errorf("checking remaining associations: %w", err)
	}
	if tsCount+pCount > 0 {
		return nil
	}
	return s.DeleteMediaEverywhere(ctx, db, franchiseID, mediaID)
}

// validateMediaFile checks the file extension and returns (ext, MediaType, error).
// Extension-based validation is used for video since magic-byte checks for AVI/MOV/MP4
// require parsing container headers; the simpler extension check is acceptable here.
func validateMediaFile(path string) (string, models.MediaType, error) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	switch ext {
	case "png", "jpg", "jpeg":
		return ext, models.MediaTypeImage, nil
	case "mp4", "mov", "avi":
		return ext, models.MediaTypeVideo, nil
	}
	return "", "", fmt.Errorf("unsupported file type %q: accepted types are png, jpg, jpeg, mp4, mov, avi", ext)
}
