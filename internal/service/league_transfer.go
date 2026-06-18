package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"smb-tools/internal/config"
	internaldb "smb-tools/internal/db"
	"smb-tools/internal/models"
	"smb-tools/internal/store"
	"smb-tools/internal/system"
	internalzip "smb-tools/internal/zip"
)

// LeagueTransferService orchestrates league discovery, export, and import —
// the production implementation of the mechanism validated live in
// docs/league-transfer/validation-results.md. It is independent of
// FranchiseService: League Transfer is a top-level feature, not scoped to
// any franchise (see docs/league-transfer/ux-flow.md).
type LeagueTransferService struct {
	dirs            *config.AppDirs
	gameChecker     system.GameRunningChecker
	smbToolsVersion string
}

func NewLeagueTransferService(dirs *config.AppDirs, gameChecker system.GameRunningChecker, smbToolsVersion string) *LeagueTransferService {
	return &LeagueTransferService{dirs: dirs, gameChecker: gameChecker, smbToolsVersion: smbToolsVersion}
}

// DiscoverLeagues returns the structure of every SMB4 league found on disk,
// regardless of whether it's a smb-tools-tracked franchise. A league save
// that can't be read or doesn't validate is logged and skipped rather than
// failing the whole discovery — one bad file shouldn't hide every other
// league from the list.
func (s *LeagueTransferService) DiscoverLeagues(ctx context.Context) ([]models.LeagueOverview, error) {
	candidates, err := config.DiscoverSaveFiles()
	if err != nil {
		return nil, fmt.Errorf("discovering save files: %w", err)
	}

	overviews := []models.LeagueOverview{}
	for _, c := range candidates {
		if c.GameVersion != models.GameVersionSMB4 {
			continue // League Transfer is SMB4-only (root CLAUDE.md)
		}

		overview, err := s.readLeagueOverview(ctx, c.Path)
		if err != nil {
			slog.Warn("LeagueTransferService.DiscoverLeagues: skipping unreadable league save", "path", c.Path, "err", err)
			continue
		}
		overviews = append(overviews, overview)
	}
	return overviews, nil
}

func (s *LeagueTransferService) readLeagueOverview(ctx context.Context, savPath string) (models.LeagueOverview, error) {
	guid, err := leagueGUIDFromFileName(savPath)
	if err != nil {
		return models.LeagueOverview{}, err
	}

	db, tmpPath, err := internaldb.DecompressAndOpen(ctx, savPath)
	if err != nil {
		return models.LeagueOverview{}, fmt.Errorf("opening league save: %w", err)
	}
	defer func() {
		_ = db.Close()
		_ = os.Remove(tmpPath)
	}()

	overview, err := store.NewLeagueSaveStore(db).GetLeagueOverview(ctx, guid)
	if err != nil {
		return models.LeagueOverview{}, err
	}
	overview.SourcePath = savPath
	return overview, nil
}

// ExportLeague packages the league identified by guid (read from
// sourceSavePath, plus its .sav.bak and optional .hash sidecar in the same
// directory) into a zip under the app's exports directory. Export is
// read-only with respect to the game — see
// docs/league-transfer/legacy-tool-analysis.md, "What the POC Got Right."
func (s *LeagueTransferService) ExportLeague(ctx context.Context, guid uuid.UUID, sourceSavePath string) (outputPath string, err error) {
	fileGUID, err := leagueGUIDFromFileName(sourceSavePath)
	if err != nil {
		return "", err
	}
	if fileGUID != guid {
		return "", fmt.Errorf("league GUID %s does not match the file name at %s", guid, sourceSavePath)
	}

	overview, err := s.readLeagueOverview(ctx, sourceSavePath)
	if err != nil {
		return "", fmt.Errorf("reading league for export: %w", err)
	}

	bakPath := sourceSavePath + ".bak"
	if _, err := os.Stat(bakPath); err != nil {
		return "", fmt.Errorf("missing %s — this league save looks incomplete", bakPath)
	}

	hashPath := strings.TrimSuffix(sourceSavePath, ".sav") + ".hash"
	if _, err := os.Stat(hashPath); err != nil {
		hashPath = "" // optional — not every league has one, see legacy-tool-analysis.md
	}

	outputPath, err = s.packExport(guid, overview.Name, sourceSavePath, bakPath, hashPath)
	if err != nil {
		return "", err
	}

	slog.Info("LeagueTransferService.ExportLeague: exported", "league", overview.Name, "guid", guid, "output", outputPath)
	return outputPath, nil
}

// ExportLeagueWithRename behaves like ExportLeague, except the exported copy
// has its display name changed to newName first — letting a user
// disambiguate a league before sharing it (e.g. two leagues that otherwise
// share a name). sourceSavePath itself is never touched: the rename happens
// against decompressed temp copies of the .sav and .sav.bak, which are then
// repackaged in place of the originals.
//
// The .hash sidecar (if present) is still carried over unmodified, same as
// ExportLeague — its format and purpose are unverified (see
// legacy-tool-analysis.md), so mutating the save's content without
// regenerating it is a known theoretical risk. In practice, this exact
// rename-then-export path was exercised manually against real save files
// with no load failures or other adverse effects observed.
func (s *LeagueTransferService) ExportLeagueWithRename(ctx context.Context, guid uuid.UUID, sourceSavePath, newName string) (outputPath string, err error) {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return "", fmt.Errorf("new league name must not be empty")
	}

	fileGUID, err := leagueGUIDFromFileName(sourceSavePath)
	if err != nil {
		return "", err
	}
	if fileGUID != guid {
		return "", fmt.Errorf("league GUID %s does not match the file name at %s", guid, sourceSavePath)
	}

	bakPath := sourceSavePath + ".bak"
	if _, err := os.Stat(bakPath); err != nil {
		return "", fmt.Errorf("missing %s — this league save looks incomplete", bakPath)
	}

	hashPath := strings.TrimSuffix(sourceSavePath, ".sav") + ".hash"
	if _, err := os.Stat(hashPath); err != nil {
		hashPath = "" // optional — not every league has one, see legacy-tool-analysis.md
	}

	renamedDir, err := os.MkdirTemp("", "smb-tools-league-renamed-*")
	if err != nil {
		return "", fmt.Errorf("creating temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(renamedDir) }()

	// internalzip.Unpack/Pack key the package contents off each file's
	// basename containing the league GUID — see leagueGUIDFromFileName and
	// the GUID check in internal/zip.Unpack — so the renamed copies must
	// keep the same "league-{GUID}.sav[.bak]" naming convention, not an
	// arbitrary temp file name.
	upper := strings.ToUpper(guid.String())
	renamedSavPath := filepath.Join(renamedDir, "league-"+upper+".sav")
	renamedBakPath := renamedSavPath + ".bak"

	if err := s.renameLeagueToFile(ctx, guid, sourceSavePath, newName, renamedSavPath); err != nil {
		return "", fmt.Errorf("renaming league save: %w", err)
	}
	if err := s.renameLeagueToFile(ctx, guid, bakPath, newName, renamedBakPath); err != nil {
		return "", fmt.Errorf("renaming league backup save: %w", err)
	}

	outputPath, err = s.packExport(guid, newName, renamedSavPath, renamedBakPath, hashPath)
	if err != nil {
		return "", err
	}

	slog.Info("LeagueTransferService.ExportLeagueWithRename: exported", "league", newName, "guid", guid, "output", outputPath)
	return outputPath, nil
}

// renameLeagueToFile decompresses srcPath, applies RenameLeague against it,
// and recompresses the result to outPath (leaving srcPath completely
// untouched).
func (s *LeagueTransferService) renameLeagueToFile(ctx context.Context, guid uuid.UUID, srcPath, newName, outPath string) error {
	tmpPath, err := internaldb.DecompressToTempFile(srcPath)
	if err != nil {
		return fmt.Errorf("decompressing %s: %w", srcPath, err)
	}
	defer func() { _ = os.Remove(tmpPath) }()

	db, err := internaldb.OpenForReadWrite(ctx, tmpPath)
	if err != nil {
		return err
	}
	if err := store.NewLeagueSaveStore(db).RenameLeague(ctx, guid, newName); err != nil {
		_ = db.Close()
		return err
	}
	if err := db.Close(); err != nil {
		return fmt.Errorf("closing renamed save: %w", err)
	}

	if err := internaldb.CompressFileAtomically(tmpPath, outPath); err != nil {
		return fmt.Errorf("recompressing renamed save: %w", err)
	}
	return nil
}

// packExport prepares the exports directory and writes the zip package for
// the given league files, returning the output path.
func (s *LeagueTransferService) packExport(guid uuid.UUID, leagueName, savPath, bakPath, hashPath string) (string, error) {
	if err := s.dirs.EnsureLeagueTransferDirs(); err != nil {
		return "", fmt.Errorf("preparing exports directory: %w", err)
	}

	timestamp := time.Now().UTC().Format("20060102150405")
	zipName := fmt.Sprintf("%s_%s.zip", sanitizeFileName(leagueName), timestamp)
	outputPath := filepath.Join(s.dirs.ExportsOutputDir(), zipName)

	err := internalzip.Pack(outputPath, internalzip.PackInput{
		GUID:            guid,
		LeagueName:      leagueName,
		SavPath:         savPath,
		BakPath:         bakPath,
		HashPath:        hashPath,
		ExportedAt:      time.Now().UTC().Format(time.RFC3339),
		SmbToolsVersion: s.smbToolsVersion,
	})
	if err != nil {
		return "", fmt.Errorf("packaging export: %w", err)
	}
	return outputPath, nil
}

// PreviewImport validates a league import package and reports what it
// contains, without writing anything to disk. Safe to call repeatedly.
func (s *LeagueTransferService) PreviewImport(ctx context.Context, zipPath string) (models.LeagueImportPreview, error) {
	tempDir, err := os.MkdirTemp("", "smb-tools-league-import-*")
	if err != nil {
		return models.LeagueImportPreview{}, fmt.Errorf("creating temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	unpacked, err := internalzip.Unpack(zipPath, tempDir)
	if err != nil {
		return models.LeagueImportPreview{}, fmt.Errorf("validating package: %w", err)
	}

	overview, err := s.validateAndReadOverview(ctx, unpacked.SavPath, unpacked.Manifest.LeagueGUID)
	if err != nil {
		return models.LeagueImportPreview{}, err
	}

	targets, err := s.buildTargetOptions(ctx, unpacked.Manifest.LeagueGUID)
	if err != nil {
		return models.LeagueImportPreview{}, err
	}

	return models.LeagueImportPreview{
		Overview:   overview,
		ExportedAt: unpacked.Manifest.ExportedAt,
		Targets:    targets,
	}, nil
}

func (s *LeagueTransferService) validateAndReadOverview(ctx context.Context, savPath string, guid uuid.UUID) (models.LeagueOverview, error) {
	tmpPath, err := internaldb.DecompressToTempFile(savPath)
	if err != nil {
		return models.LeagueOverview{}, fmt.Errorf("decompressing package's league save: %w", err)
	}
	defer func() { _ = os.Remove(tmpPath) }()

	db, err := internaldb.OpenForReadWrite(ctx, tmpPath)
	if err != nil {
		return models.LeagueOverview{}, err
	}
	defer func() { _ = db.Close() }()

	saveStore := store.NewLeagueSaveStore(db)
	if err := saveStore.ValidateLeagueSaveShape(ctx); err != nil {
		return models.LeagueOverview{}, fmt.Errorf("package failed validation: %w", err)
	}

	overview, err := saveStore.GetLeagueOverview(ctx, guid)
	if err != nil {
		return models.LeagueOverview{}, fmt.Errorf("reading league from package: %w", err)
	}
	return overview, nil
}

func (s *LeagueTransferService) buildTargetOptions(ctx context.Context, guid uuid.UUID) ([]models.ImportTargetOption, error) {
	candidates, err := config.DiscoverSteamSaveDirs()
	if err != nil {
		return nil, fmt.Errorf("discovering Steam save directories: %w", err)
	}

	targets := make([]models.ImportTargetOption, 0, len(candidates))
	for _, c := range candidates {
		exists, err := s.leagueExistsInMasterSave(ctx, c.MasterSavePath, guid)
		if err != nil {
			return nil, fmt.Errorf("checking existing registration in %s: %w", c.MasterSavePath, err)
		}
		targets = append(targets, models.ImportTargetOption{
			SteamSaveDirCandidate: c,
			AlreadyRegistered:     exists,
		})
	}
	return targets, nil
}

func (s *LeagueTransferService) leagueExistsInMasterSave(ctx context.Context, masterSavePath string, guid uuid.UUID) (bool, error) {
	tmpPath, err := internaldb.DecompressToTempFile(masterSavePath)
	if err != nil {
		return false, fmt.Errorf("decompressing master.sav: %w", err)
	}
	defer func() { _ = os.Remove(tmpPath) }()

	db, err := internaldb.OpenForReadWrite(ctx, tmpPath)
	if err != nil {
		return false, err
	}
	defer func() { _ = db.Close() }()

	return store.NewLeagueRegistryStore(db).LeagueExists(ctx, guid)
}

// ConfirmImport performs the actual import: validates the package again
// (defensively — PreviewImport's result is not assumed still valid),
// refuses if SMB4 is running, registers the league in targetDirPath's
// master.sav, and copies the league's save files into place.
//
// Ordering is deliberate: the league files are copied into targetDirPath
// only after the registration has been prepared (in a temp copy of
// master.sav, not the live file) and confirmed collision-free, but the live
// master.sav is only ever touched as the final step, after a timestamped
// backup. A failure at any point before that final swap leaves the live
// master.sav completely untouched — at worst, an unregistered orphan league
// file sits in the save directory, which is harmless clutter, not
// corruption.
func (s *LeagueTransferService) ConfirmImport(ctx context.Context, zipPath, targetDirPath string) error {
	running, err := s.gameChecker.IsGameRunning()
	if err != nil {
		return fmt.Errorf("checking whether SMB4 is running: %w", err)
	}
	if running {
		return fmt.Errorf("SMB4 is currently running — close the game before importing a league")
	}

	tempDir, err := os.MkdirTemp("", "smb-tools-league-import-*")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	unpacked, err := internalzip.Unpack(zipPath, tempDir)
	if err != nil {
		return fmt.Errorf("validating package: %w", err)
	}
	guid := unpacked.Manifest.LeagueGUID

	if _, err := s.validateAndReadOverview(ctx, unpacked.SavPath, guid); err != nil {
		return err
	}

	targetMasterSavePath := filepath.Join(targetDirPath, "master.sav")
	preMutationInfo, err := os.Stat(targetMasterSavePath)
	if err != nil {
		return fmt.Errorf("could not find master.sav in %s: %w", targetDirPath, err)
	}

	registeredMasterTmpPath, err := s.prepareRegisteredMasterSave(ctx, targetMasterSavePath, guid)
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(registeredMasterTmpPath) }()

	if err := s.copyLeagueFilesInto(targetDirPath, guid, unpacked); err != nil {
		return fmt.Errorf("copying league save files into place: %w", err)
	}

	// Nanosecond precision, not just seconds: two imports completing within
	// the same wall-clock second (e.g. back-to-back in a script, or in
	// tests) must not collide and silently overwrite each other's backup —
	// that would violate the "every backup is kept" safety requirement in
	// docs/league-transfer/implementation-plan.md (Q4).
	backupTimestamp := time.Now().UTC().Format("20060102150405.000000000")
	if _, err := internaldb.BackupFileTimestamped(
		targetMasterSavePath, s.dirs.MasterSaveBackupsDir(), "master", backupTimestamp,
	); err != nil {
		return fmt.Errorf("backing up master.sav before import: %w", err)
	}

	// Defense-in-depth: if master.sav changed since we first looked at it
	// (almost certainly the game writing to it), abort rather than swap in
	// a registration based on stale data. This is an optimistic-concurrency
	// check, not a true OS-level lock — see
	// docs/league-transfer/implementation-plan.md's discussion of why the
	// process-running check above is the primary safety mechanism and this
	// is only a cheap secondary one.
	postInfo, err := os.Stat(targetMasterSavePath)
	if err != nil {
		return fmt.Errorf("re-checking master.sav before final write: %w", err)
	}
	if postInfo.ModTime() != preMutationInfo.ModTime() || postInfo.Size() != preMutationInfo.Size() {
		return fmt.Errorf("master.sav changed during import — please try again (this usually means the game was running)")
	}

	if err := internaldb.CompressFileAtomically(registeredMasterTmpPath, targetMasterSavePath); err != nil {
		return fmt.Errorf("writing updated master.sav: %w", err)
	}

	slog.Info("LeagueTransferService.ConfirmImport: imported", "guid", guid, "target", targetDirPath)
	return nil
}

// prepareRegisteredMasterSave decompresses masterSavePath, checks guid isn't
// already registered, registers it, and returns the path to the mutated
// temp file — without touching masterSavePath itself.
func (s *LeagueTransferService) prepareRegisteredMasterSave(ctx context.Context, masterSavePath string, guid uuid.UUID) (string, error) {
	tmpPath, err := internaldb.DecompressToTempFile(masterSavePath)
	if err != nil {
		return "", fmt.Errorf("decompressing master.sav: %w", err)
	}

	db, err := internaldb.OpenForReadWrite(ctx, tmpPath)
	if err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	defer func() { _ = db.Close() }()

	registry := store.NewLeagueRegistryStore(db)
	exists, err := registry.LeagueExists(ctx, guid)
	if err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("checking existing registration: %w", err)
	}
	if exists {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("this league is already registered in %s", masterSavePath)
	}

	if err := registry.RegisterLeague(ctx, guid); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}

	return tmpPath, nil
}

func (s *LeagueTransferService) copyLeagueFilesInto(targetDirPath string, guid uuid.UUID, unpacked internalzip.UnpackResult) error {
	upper := strings.ToUpper(guid.String())
	destSav := filepath.Join(targetDirPath, "league-"+upper+".sav")
	destBak := destSav + ".bak"

	if err := copyFile(unpacked.SavPath, destSav); err != nil {
		return err
	}
	if err := copyFile(unpacked.BakPath, destBak); err != nil {
		return err
	}
	if unpacked.HashPath != "" {
		destHash := strings.TrimSuffix(destSav, ".sav") + ".hash"
		if err := copyFile(unpacked.HashPath, destHash); err != nil {
			return err
		}
	}
	return nil
}

// leagueGUIDFromFileName parses the GUID out of a "league-{GUID}.sav"
// (or "...sav.bak", "...hash") path — the naming convention confirmed
// against real save files in docs/league-transfer/legacy-tool-analysis.md.
func leagueGUIDFromFileName(path string) (uuid.UUID, error) {
	name := filepath.Base(path)
	name = strings.TrimSuffix(name, ".bak")
	name = strings.TrimSuffix(name, ".sav")
	name = strings.TrimSuffix(name, ".hash")
	name = strings.TrimPrefix(name, "league-")

	guid, err := uuid.Parse(name)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("could not parse a league GUID from file name %q: %w", filepath.Base(path), err)
	}
	return guid, nil
}

// sanitizeFileName mirrors the legacy tool's filename sanitization
// (legacy-tool-analysis.md) — keep alphanumerics, dots, underscores, and
// hyphens; replace everything else with an underscore.
func sanitizeFileName(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	sanitized := strings.Trim(b.String(), "_")
	if sanitized == "" {
		return "league"
	}
	return sanitized
}
