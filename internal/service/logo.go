package service

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"smb-tools/internal/config"
	"smb-tools/internal/models"
	"smb-tools/internal/store"
)

// LogoService handles logo file I/O and coordinates with LogoStore.
type LogoService struct {
	logoStore *store.LogoStore
	dirs      *config.AppDirs
}

func NewLogoService(logoStore *store.LogoStore, dirs *config.AppDirs) *LogoService {
	return &LogoService{logoStore: logoStore, dirs: dirs}
}

// UploadAndAssign copies the source image to the franchise's logo directory,
// inserts a logos row, inserts a logo_assignments row, and returns the DTO.
func (s *LogoService) UploadAndAssign(
	ctx context.Context,
	db *sql.DB,
	franchiseID string,
	teamID int,
	sourceFilePath string,
	startSeason, endSeason *int,
) (models.TeamLogo, models.TeamLogoAssignment, error) {
	ext, err := validateImageFile(sourceFilePath)
	if err != nil {
		return models.TeamLogo{}, models.TeamLogoAssignment{}, fmt.Errorf("invalid image: %w", err)
	}

	logoDir := s.dirs.TeamLogosDir(franchiseID, teamID)
	if err := os.MkdirAll(logoDir, 0o700); err != nil {
		return models.TeamLogo{}, models.TeamLogoAssignment{}, fmt.Errorf("creating logo directory: %w", err)
	}

	logoID := uuid.New().String()
	filename := logoID + "." + ext
	dstPath := filepath.Join(logoDir, filename)

	if err := copyFile(sourceFilePath, dstPath); err != nil {
		return models.TeamLogo{}, models.TeamLogoAssignment{}, fmt.Errorf("copying logo file: %w", err)
	}

	relPath := filepath.Join("assets", "logos", strconv.Itoa(teamID), filename)
	// Normalise to forward slashes for cross-platform DB storage.
	relPath = filepath.ToSlash(relPath)

	now := time.Now().UTC()
	logo := models.TeamLogo{
		ID:         logoID,
		TeamID:     teamID,
		FilePath:   relPath,
		UploadedAt: now,
	}
	if err := s.logoStore.InsertLogo(ctx, db, logo); err != nil {
		_ = os.Remove(dstPath)
		return models.TeamLogo{}, models.TeamLogoAssignment{}, fmt.Errorf("recording logo: %w", err)
	}

	assignmentID := uuid.New().String()
	assignment := models.TeamLogoAssignment{
		ID:          assignmentID,
		LogoID:      logoID,
		StartSeason: startSeason,
		EndSeason:   endSeason,
		AssignedAt:  now,
	}
	if err := s.logoStore.InsertLogoAssignment(ctx, db, assignment); err != nil {
		_ = os.Remove(dstPath)
		return models.TeamLogo{}, models.TeamLogoAssignment{}, fmt.Errorf("recording logo assignment: %w", err)
	}

	return logo, assignment, nil
}

// AssignExisting creates a new assignment for an already-uploaded logo.
func (s *LogoService) AssignExisting(
	ctx context.Context,
	db *sql.DB,
	logoID string,
	startSeason, endSeason *int,
) (models.TeamLogoAssignment, error) {
	assignment := models.TeamLogoAssignment{
		ID:          uuid.New().String(),
		LogoID:      logoID,
		StartSeason: startSeason,
		EndSeason:   endSeason,
		AssignedAt:  time.Now().UTC(),
	}
	if err := s.logoStore.InsertLogoAssignment(ctx, db, assignment); err != nil {
		return models.TeamLogoAssignment{}, fmt.Errorf("recording logo assignment: %w", err)
	}
	return assignment, nil
}

// GetLogoURLForSeason returns the virtual asset URL for the logo covering seasonNum,
// or "" when no logo is assigned for that season.
func (s *LogoService) GetLogoURLForSeason(
	ctx context.Context,
	db *sql.DB,
	teamID int,
	seasonNum int,
) (string, error) {
	logo, err := s.logoStore.GetLogoForSeason(ctx, db, teamID, seasonNum)
	if err != nil {
		return "", err
	}
	if logo == nil {
		return "", nil
	}
	return logoVirtualURL(logo), nil
}

// GetTeamLogos returns all logos for a team with their assignment slices and virtual URLs.
func (s *LogoService) GetTeamLogos(
	ctx context.Context,
	db *sql.DB,
	teamID int,
) ([]store.TeamLogoWithAssignments, error) {
	return s.logoStore.GetLogosForTeam(ctx, db, teamID)
}

// DeleteAssignment removes a logo assignment. If it was the last assignment
// referencing the logo, the file on disk is also deleted.
func (s *LogoService) DeleteAssignment(
	ctx context.Context,
	db *sql.DB,
	franchiseID string,
	assignmentID string,
) error {
	logoID, err := s.logoStore.LogoIDForAssignment(ctx, db, assignmentID)
	if err != nil {
		return fmt.Errorf("resolving assignment: %w", err)
	}

	logo, err := s.logoStore.GetLogoByID(ctx, db, logoID)
	if err != nil {
		return fmt.Errorf("fetching logo: %w", err)
	}

	if err := s.logoStore.DeleteAssignment(ctx, db, assignmentID); err != nil {
		return err
	}

	count, err := s.logoStore.GetAssignmentCountForLogo(ctx, db, logoID)
	if err != nil {
		return fmt.Errorf("checking remaining assignments: %w", err)
	}

	if count == 0 {
		fullPath := filepath.Join(s.dirs.FranchiseDir(franchiseID), filepath.FromSlash(logo.FilePath))
		_ = os.Remove(fullPath)
		if err := s.logoStore.DeleteLogoAndAllAssignments(ctx, db, logoID); err != nil {
			return fmt.Errorf("cleaning up logo record: %w", err)
		}
	}

	return nil
}

// logoVirtualURL constructs the AssetsHandler virtual URL for a logo.
// logo.FilePath is like "assets/logos/42/uuid.png"; virtual URL is "/_logos/42/uuid.png".
func logoVirtualURL(logo *models.TeamLogo) string {
	// FilePath: "assets/logos/{teamId}/{filename}"
	// Virtual:  "/_logos/{teamId}/{filename}"
	const prefix = "assets/logos/"
	if len(logo.FilePath) <= len(prefix) {
		return ""
	}
	return "/_logos/" + logo.FilePath[len(prefix):]
}

// validateImageFile reads the first bytes of the file and checks for a valid
// PNG or JPEG magic number. Returns "png" or "jpg" on success.
func validateImageFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening file: %w", err)
	}
	defer func() { _ = f.Close() }()

	buf := make([]byte, 8)
	if _, err := io.ReadFull(f, buf); err != nil {
		return "", fmt.Errorf("reading file header: %w", err)
	}

	// PNG: \x89 P N G \r \n \x1a \n
	if buf[0] == 0x89 && buf[1] == 0x50 && buf[2] == 0x4E && buf[3] == 0x47 &&
		buf[4] == 0x0D && buf[5] == 0x0A && buf[6] == 0x1A && buf[7] == 0x0A {
		return "png", nil
	}
	// JPEG: \xff \xd8 \xff
	if buf[0] == 0xFF && buf[1] == 0xD8 && buf[2] == 0xFF {
		return "jpg", nil
	}

	return "", fmt.Errorf("file is not a valid PNG or JPEG image")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source: %w", err)
	}
	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copying data: %w", err)
	}
	return nil
}
