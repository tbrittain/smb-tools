package main

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

// ---- League Transfer Wails bindings ----------------------------------------
//
// League Transfer is a top-level mode independent of franchise tracking —
// see docs/league-transfer/ux-flow.md. These bindings delegate entirely to
// LeagueTransferService; see internal/service/league_transfer.go for the
// actual orchestration.

// DiscoverLeagues returns every SMB4 league found on disk, regardless of
// whether it's a tracked franchise.
func (a *App) DiscoverLeagues() ([]LeagueOverviewDTO, error) {
	if a.leagueTransferService == nil {
		return nil, fmt.Errorf("app not initialized")
	}

	overviews, err := a.leagueTransferService.DiscoverLeagues(a.ctx)
	if err != nil {
		slog.Error("DiscoverLeagues: failed", "err", err)
		return nil, err
	}

	out := make([]LeagueOverviewDTO, len(overviews))
	for i, o := range overviews {
		out[i] = leagueOverviewToDTO(o)
	}
	return out, nil
}

// ExportLeague packages the league identified by leagueGUID (read from
// sourceSavePath) into a zip file, returning the output path.
func (a *App) ExportLeague(leagueGUID, sourceSavePath string) (string, error) {
	slog.Info("ExportLeague", "guid", leagueGUID, "path", sourceSavePath)
	if a.leagueTransferService == nil {
		return "", fmt.Errorf("app not initialized")
	}

	guid, err := uuid.Parse(leagueGUID)
	if err != nil {
		return "", fmt.Errorf("invalid league GUID %q: %w", leagueGUID, err)
	}

	outputPath, err := a.leagueTransferService.ExportLeague(a.ctx, guid, sourceSavePath)
	if err != nil {
		slog.Error("ExportLeague: failed", "err", err)
		return "", err
	}
	return outputPath, nil
}

// PreviewLeagueImport validates an import package and reports what it
// contains, without writing anything to disk.
func (a *App) PreviewLeagueImport(zipPath string) (LeagueImportPreviewDTO, error) {
	if a.leagueTransferService == nil {
		return LeagueImportPreviewDTO{}, fmt.Errorf("app not initialized")
	}

	preview, err := a.leagueTransferService.PreviewImport(a.ctx, zipPath)
	if err != nil {
		slog.Error("PreviewLeagueImport: failed", "err", err)
		return LeagueImportPreviewDTO{}, err
	}
	return leagueImportPreviewToDTO(preview), nil
}

// ConfirmLeagueImport performs the actual import of zipPath into
// targetDirPath (one of the directory paths returned by
// PreviewLeagueImport's Targets). Refuses if SMB4 is currently running or if
// the league is already registered in that target.
func (a *App) ConfirmLeagueImport(zipPath, targetDirPath string) error {
	slog.Info("ConfirmLeagueImport", "zip", zipPath, "target", targetDirPath)
	if a.leagueTransferService == nil {
		return fmt.Errorf("app not initialized")
	}

	if err := a.leagueTransferService.ConfirmImport(a.ctx, zipPath, targetDirPath); err != nil {
		slog.Error("ConfirmLeagueImport: failed", "err", err)
		return err
	}
	return nil
}
