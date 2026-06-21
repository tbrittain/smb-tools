package main

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	internaldb "smb-tools/internal/db"
	"smb-tools/internal/store"
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

// ExportLeagueWithRename packages the league identified by leagueGUID (read
// from sourceSavePath) into a zip file, with its display name changed to
// newName first, returning the output path.
func (a *App) ExportLeagueWithRename(leagueGUID, sourceSavePath, newName string) (string, error) {
	slog.Info("ExportLeagueWithRename", "guid", leagueGUID, "path", sourceSavePath, "newName", newName)
	if a.leagueTransferService == nil {
		return "", fmt.Errorf("app not initialized")
	}

	guid, err := uuid.Parse(leagueGUID)
	if err != nil {
		return "", fmt.Errorf("invalid league GUID %q: %w", leagueGUID, err)
	}

	outputPath, err := a.leagueTransferService.ExportLeagueWithRename(a.ctx, guid, sourceSavePath, newName)
	if err != nil {
		slog.Error("ExportLeagueWithRename: failed", "err", err)
		return "", err
	}
	return outputPath, nil
}

// OpenLeagueExportDir opens the folder containing exportedFilePath in the OS
// file manager, so the user can immediately locate the file they just
// exported.
func (a *App) OpenLeagueExportDir(exportedFilePath string) error {
	return openDirectory(filepath.Dir(exportedFilePath))
}

// BrowseLeagueImportZip opens the OS file picker filtered to .zip files and
// returns the selected path. Returns "" if the user cancels.
func (a *App) BrowseLeagueImportZip() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select League Export File",
		Filters: []runtime.FileFilter{
			{DisplayName: "League Export Files (*.zip)", Pattern: "*.zip"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("file dialog: %w", err)
	}
	return path, nil
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

// ListSnapshotExportCandidates returns every franchise snapshot across all
// registered franchises, for the League Transfer "From Snapshot" export
// flow. A franchise whose companion database can't be opened is logged and
// skipped, not fatal to the whole list, mirroring DiscoverLeagues' tolerance
// of unreadable saves. Snapshots that have been compressed in place
// (Snapshot.Compressed) are skipped too: nothing in the codebase produces a
// compressed snapshot today (MarkCompressed has no callers), so there is no
// decompression step to run before export — silently exporting one would
// risk packaging garbage.
func (a *App) ListSnapshotExportCandidates() ([]SnapshotExportCandidateDTO, error) {
	if a.franchiseStore == nil || a.dirs == nil {
		return nil, fmt.Errorf("app not initialized")
	}

	franchises, err := a.franchiseStore.List(a.ctx)
	if err != nil {
		slog.Error("ListSnapshotExportCandidates: listing franchises", "err", err)
		return nil, err
	}

	out := []SnapshotExportCandidateDTO{}
	for _, f := range franchises {
		db, err := internaldb.OpenCompanion(a.ctx, a.dirs.CompanionDBPath(f.ID))
		if err != nil {
			slog.Warn("ListSnapshotExportCandidates: skipping franchise with unreadable companion DB", "franchiseID", f.ID, "err", err)
			continue
		}

		snaps, err := store.NewSnapshotStore(db).List(a.ctx)
		closeErr := db.Close()
		if err != nil {
			slog.Warn("ListSnapshotExportCandidates: skipping franchise with unreadable snapshots", "franchiseID", f.ID, "err", err)
			continue
		}
		if closeErr != nil {
			slog.Warn("ListSnapshotExportCandidates: closing companion DB", "franchiseID", f.ID, "err", closeErr)
		}

		for _, sn := range snaps {
			if sn.Compressed {
				slog.Warn("ListSnapshotExportCandidates: skipping compressed snapshot (unsupported)", "franchiseID", f.ID, "snapshotID", sn.ID)
				continue
			}
			out = append(out, SnapshotExportCandidateDTO{
				FranchiseID:   f.ID,
				FranchiseName: f.Name,
				SnapshotID:    sn.ID,
				SeasonNum:     sn.SeasonNum,
				CapturedAt:    sn.CapturedAt.UTC().Format("2006-01-02T15:04:05Z"),
				FileSizeBytes: sn.FileSizeBytes,
			})
		}
	}
	return out, nil
}

// ExportSnapshotAsLeague packages the snapshot identified by franchiseID and
// snapshotID into a league export zip, with its display name set to
// newName. newName is mandatory — see LeagueTransferService.ExportSnapshot.
func (a *App) ExportSnapshotAsLeague(franchiseID string, snapshotID int64, newName string) (string, error) {
	slog.Info("ExportSnapshotAsLeague", "franchiseID", franchiseID, "snapshotID", snapshotID, "newName", newName)
	if a.franchiseStore == nil || a.dirs == nil || a.leagueTransferService == nil {
		return "", fmt.Errorf("app not initialized")
	}

	if _, err := a.franchiseStore.GetByID(a.ctx, franchiseID); err != nil {
		return "", fmt.Errorf("looking up franchise %q: %w", franchiseID, err)
	}

	db, err := internaldb.OpenCompanion(a.ctx, a.dirs.CompanionDBPath(franchiseID))
	if err != nil {
		return "", fmt.Errorf("opening companion database for franchise %q: %w", franchiseID, err)
	}
	defer func() { _ = db.Close() }()

	snap, err := store.NewSnapshotStore(db).GetByID(a.ctx, snapshotID)
	if err != nil {
		return "", fmt.Errorf("looking up snapshot %d: %w", snapshotID, err)
	}
	if snap.Compressed {
		return "", fmt.Errorf("snapshot %d is compressed — exporting compressed snapshots is not supported", snapshotID)
	}

	snapshotPath := filepath.Join(a.dirs.SnapshotsDir(franchiseID), string(snap.FileName))

	outputPath, err := a.leagueTransferService.ExportSnapshot(a.ctx, snapshotPath, newName)
	if err != nil {
		slog.Error("ExportSnapshotAsLeague: failed", "err", err)
		return "", err
	}
	return outputPath, nil
}

// BrowseLeagueExportDirectory opens a directory picker and scans the chosen
// folder (and its immediate subdirectories) for SMB4 league saves, the
// League Transfer export-side equivalent of franchise creation's
// BrowseSaveDirectory. Returns an empty slice if the user cancels.
func (a *App) BrowseLeagueExportDirectory() ([]LeagueOverviewDTO, error) {
	if a.leagueTransferService == nil {
		return nil, fmt.Errorf("app not initialized")
	}

	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Folder Containing League Save Files",
	})
	if err != nil {
		return nil, fmt.Errorf("directory dialog: %w", err)
	}
	if dir == "" {
		return []LeagueOverviewDTO{}, nil
	}

	overviews := a.leagueTransferService.DiscoverLeaguesInDir(a.ctx, dir)
	out := make([]LeagueOverviewDTO, len(overviews))
	for i, o := range overviews {
		out[i] = leagueOverviewToDTO(o)
	}
	return out, nil
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
