package config

import (
	"fmt"
	"os"
	"path/filepath"

	"smb-tools/internal/models"
)

// LeagueTransferDir returns the root directory for League Transfer's own
// data (master.sav backups, export output). This is deliberately not nested
// under FranchisesDir — League Transfer is a top-level feature, independent
// of any franchise (see docs/league-transfer/ux-flow.md).
func (d *AppDirs) LeagueTransferDir() string {
	return filepath.Join(d.DataDir, "league-transfer")
}

// MasterSaveBackupsDir returns the directory where master.sav backups are
// kept, one per import, timestamped (see internal/db.BackupFileTimestamped).
// Backups accumulate as a history rather than being overwritten — the
// MVP keeps every backup ever taken rather than pruning, per the decision in
// docs/league-transfer/implementation-plan.md (Q4).
func (d *AppDirs) MasterSaveBackupsDir() string {
	return filepath.Join(d.LeagueTransferDir(), "backups")
}

// ExportsOutputDir returns the default directory exported league zips are
// written to.
func (d *AppDirs) ExportsOutputDir() string {
	return filepath.Join(d.LeagueTransferDir(), "exports")
}

// EnsureLeagueTransferDirs creates the League Transfer directory structure,
// mirroring EnsureFranchiseDirs' on-demand-creation pattern.
func (d *AppDirs) EnsureLeagueTransferDirs() error {
	for _, dir := range []string{d.MasterSaveBackupsDir(), d.ExportsOutputDir()} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("creating league transfer directory %q: %w", dir, err)
		}
	}
	return nil
}

// DiscoverSteamSaveDirs returns every Steam-profile directory under the SMB4
// save root that contains a master.sav file — i.e., one entry per Steam
// account that has played SMB4 on this machine. This generalizes the
// single-directory assumption in savegame_paths.go (written for franchise
// tracking's common case of one account) without changing that function's
// behavior; League Transfer needs to know about every profile so it can ask
// the user which one to import into when there's more than one (see
// docs/league-transfer/implementation-plan.md, Q2).
//
// League Transfer is SMB4-only (see root CLAUDE.md) — this does not consider
// SMB3 at all, since SMB3 has no master.sav/league-registry concept.
func DiscoverSteamSaveDirs() ([]models.SteamSaveDirCandidate, error) {
	roots, err := saveGameRoots()
	if err != nil {
		return nil, err
	}

	var smb4Root string
	for _, root := range roots {
		if root.version == models.GameVersionSMB4 {
			smb4Root = root.dir
			break
		}
	}
	if smb4Root == "" {
		return nil, fmt.Errorf("no SMB4 save root configured for this platform")
	}

	entries, err := os.ReadDir(smb4Root)
	if err != nil {
		// No SMB4 directory at all yet — not an error, just no candidates.
		return []models.SteamSaveDirCandidate{}, nil
	}

	candidates := []models.SteamSaveDirCandidate{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dirPath := filepath.Join(smb4Root, e.Name())
		masterSavePath := filepath.Join(dirPath, "master.sav")
		if _, err := os.Stat(masterSavePath); err != nil {
			continue
		}
		candidates = append(candidates, models.SteamSaveDirCandidate{
			SteamID:        e.Name(),
			DirPath:        dirPath,
			MasterSavePath: masterSavePath,
		})
	}
	return candidates, nil
}
