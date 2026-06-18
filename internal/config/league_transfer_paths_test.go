package config_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"smb-tools/internal/config"
)

// setSMB4Root points the platform-specific env var saveGameRoots() reads at
// a temp directory, then returns the resolved "Super Mega Baseball 4" root
// path under it — mirroring the exact branching in
// internal/config/savegame_paths.go so this test is deterministic on both
// Windows (dev) and Linux (CI), per internal/CLAUDE.md's requirement that
// all Go tests pass on Linux.
func setSMB4Root(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()

	switch runtime.GOOS {
	case "windows":
		t.Setenv("LOCALAPPDATA", tmp)
		return filepath.Join(tmp, "Metalhead", "Super Mega Baseball 4")
	default: // linux and other unix-like — saveGameRoots' default branch reads $HOME, not XDG_DATA_HOME
		t.Setenv("HOME", tmp)
		return filepath.Join(tmp, ".local", "share", "Metalhead", "Super Mega Baseball 4")
	}
}

func TestDiscoverSteamSaveDirs_NoSMB4Directory(t *testing.T) {
	setSMB4Root(t)

	got, err := config.DiscoverSteamSaveDirs()
	if err != nil {
		t.Fatalf("DiscoverSteamSaveDirs: %v", err)
	}
	if got == nil {
		t.Fatal("expected a non-nil slice even with no SMB4 directory present")
	}
	if len(got) != 0 {
		t.Errorf("expected 0 candidates, got %d", len(got))
	}
}

func TestDiscoverSteamSaveDirs_OnlyDirsWithMasterSav(t *testing.T) {
	smb4Root := setSMB4Root(t)

	// One profile with master.sav (should be found), one without (should be skipped).
	withMaster := filepath.Join(smb4Root, "76561198034146134")
	withoutMaster := filepath.Join(smb4Root, "76561199099999999")
	createFile(t, filepath.Join(withMaster, "master.sav"))
	createFile(t, filepath.Join(withoutMaster, "league-1D454F48-B9BD-42A0-A528-358E46142A64.sav"))

	got, err := config.DiscoverSteamSaveDirs()
	if err != nil {
		t.Fatalf("DiscoverSteamSaveDirs: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 candidate, got %d: %+v", len(got), got)
	}
	if got[0].SteamID != "76561198034146134" {
		t.Errorf("SteamID = %q, want %q", got[0].SteamID, "76561198034146134")
	}
	if got[0].DirPath != withMaster {
		t.Errorf("DirPath = %q, want %q", got[0].DirPath, withMaster)
	}
	if got[0].MasterSavePath != filepath.Join(withMaster, "master.sav") {
		t.Errorf("MasterSavePath = %q, want %q", got[0].MasterSavePath, filepath.Join(withMaster, "master.sav"))
	}
}

func TestDiscoverSteamSaveDirs_MultipleProfiles(t *testing.T) {
	smb4Root := setSMB4Root(t)

	profiles := []string{"76561198034146134", "76561199012345678"}
	for _, p := range profiles {
		createFile(t, filepath.Join(smb4Root, p, "master.sav"))
	}

	got, err := config.DiscoverSteamSaveDirs()
	if err != nil {
		t.Fatalf("DiscoverSteamSaveDirs: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(got))
	}
}

func TestAppDirs_LeagueTransferPaths(t *testing.T) {
	dirs := &config.AppDirs{DataDir: filepath.Join(t.TempDir(), "data")}

	if got, want := dirs.LeagueTransferDir(), filepath.Join(dirs.DataDir, "league-transfer"); got != want {
		t.Errorf("LeagueTransferDir() = %q, want %q", got, want)
	}
	if got, want := dirs.MasterSaveBackupsDir(), filepath.Join(dirs.LeagueTransferDir(), "backups"); got != want {
		t.Errorf("MasterSaveBackupsDir() = %q, want %q", got, want)
	}
	if got, want := dirs.ExportsOutputDir(), filepath.Join(dirs.LeagueTransferDir(), "exports"); got != want {
		t.Errorf("ExportsOutputDir() = %q, want %q", got, want)
	}

	if err := dirs.EnsureLeagueTransferDirs(); err != nil {
		t.Fatalf("EnsureLeagueTransferDirs: %v", err)
	}
	for _, dir := range []string{dirs.MasterSaveBackupsDir(), dirs.ExportsOutputDir()} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("expected %q to exist after EnsureLeagueTransferDirs: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%q exists but is not a directory", dir)
		}
	}
}
