package main

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"smb-tools/internal/models"
)

// BrowseLogoFile opens a file picker filtered to PNG and JPEG images.
// Returns the selected file path, or "" if cancelled.
func (a *App) BrowseLogoFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Logo Image",
		Filters: []runtime.FileFilter{
			{DisplayName: "Image Files (*.png;*.jpg;*.jpeg)", Pattern: "*.png;*.jpg;*.jpeg"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("opening file dialog: %w", err)
	}
	return path, nil
}

// UploadAndAssignTeamLogo copies the source image to the franchise's logo directory,
// records it in the DB, and assigns it to the given season range.
func (a *App) UploadAndAssignTeamLogo(
	teamID int,
	sourceFilePath string,
	startSeason, endSeason *int,
) (TeamLogoDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return TeamLogoDTO{}, err
	}
	logo, assignment, err := a.logoService.UploadAndAssign(
		a.ctx, a.companionDB, a.activeFranchise.ID, teamID, sourceFilePath, startSeason, endSeason,
	)
	if err != nil {
		return TeamLogoDTO{}, err
	}
	return TeamLogoDTO{
		ID:         logo.ID,
		TeamID:     logo.TeamID,
		LogoURL:    "/_logos/" + strconv.Itoa(logo.TeamID) + "/" + filepath.Base(filepath.FromSlash(logo.FilePath)),
		UploadedAt: logo.UploadedAt.Format("2006-01-02T15:04:05Z"),
		Assignments: []TeamLogoAssignmentDTO{
			logoAssignmentToDTO(assignment),
		},
	}, nil
}

// AssignExistingTeamLogo creates a new season-range assignment for an already-uploaded logo.
func (a *App) AssignExistingTeamLogo(
	logoID string,
	startSeason, endSeason *int,
) (TeamLogoAssignmentDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return TeamLogoAssignmentDTO{}, err
	}
	assignment, err := a.logoService.AssignExisting(a.ctx, a.companionDB, logoID, startSeason, endSeason)
	if err != nil {
		return TeamLogoAssignmentDTO{}, err
	}
	return logoAssignmentToDTO(assignment), nil
}

// GetTeamLogos returns all uploaded logos for the given team, each with their
// assignment slices and virtual asset URLs.
func (a *App) GetTeamLogos(teamID int) ([]TeamLogoDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	all, err := a.logoService.GetTeamLogos(a.ctx, a.companionDB, teamID)
	if err != nil {
		return nil, err
	}
	out := make([]TeamLogoDTO, len(all))
	for i, lwa := range all {
		assignments := make([]TeamLogoAssignmentDTO, len(lwa.Assignments))
		for j, a := range lwa.Assignments {
			assignments[j] = logoAssignmentToDTO(a)
		}
		out[i] = TeamLogoDTO{
			ID:          lwa.Logo.ID,
			TeamID:      lwa.Logo.TeamID,
			LogoURL:     "/_logos/" + strconv.Itoa(lwa.Logo.TeamID) + "/" + filepath.Base(filepath.FromSlash(lwa.Logo.FilePath)),
			UploadedAt:  lwa.Logo.UploadedAt.Format("2006-01-02T15:04:05Z"),
			Assignments: assignments,
		}
	}
	return out, nil
}

// GetLogoURLForSeason returns the virtual asset URL for the logo covering the given
// season, or "" when no logo is assigned for that season.
func (a *App) GetLogoURLForSeason(teamID int, seasonNum int) (string, error) {
	if err := a.requireCompanionDB(); err != nil {
		return "", err
	}
	return a.logoService.GetLogoURLForSeason(a.ctx, a.companionDB, teamID, seasonNum)
}

// DeleteTeamLogoAssignment removes a logo assignment. If it was the last assignment
// for that logo, the file on disk is also deleted.
func (a *App) DeleteTeamLogoAssignment(assignmentID string) error {
	if err := a.requireCompanionDB(); err != nil {
		return err
	}
	return a.logoService.DeleteAssignment(a.ctx, a.companionDB, a.activeFranchise.ID, assignmentID)
}

func logoAssignmentToDTO(a models.TeamLogoAssignment) TeamLogoAssignmentDTO {
	return TeamLogoAssignmentDTO{
		ID:          a.ID,
		LogoID:      a.LogoID,
		StartSeason: a.StartSeason,
		EndSeason:   a.EndSeason,
		AssignedAt:  a.AssignedAt.Format("2006-01-02T15:04:05Z"),
	}
}
