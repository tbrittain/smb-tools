package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// SaveGameCandidate is a potential SMB save file found on the filesystem.
type SaveGameCandidate struct {
	Path        string
	GameVersion string // "smb3" or "smb4"
}

// DiscoverSaveFiles returns known default save file locations for both SMB3
// and SMB4, filtering to paths that actually exist on disk.
//
// The game stores save files under:
//   - Windows: %LOCALAPPDATA%\Metalhead\Super Mega Baseball {3|4}\{steam_id}\
//   - macOS:   ~/Library/Application Support/Metalhead/Super Mega Baseball {3|4}/{steam_id}/
//   - Linux:   ~/.local/share/Metalhead/Super Mega Baseball {3|4}/{steam_id}/  (Steam via Proton)
func DiscoverSaveFiles() []SaveGameCandidate {
	var candidates []SaveGameCandidate

	roots := saveGameRoots()
	for _, root := range roots {
		found := walkForSaveFiles(root.dir, root.version)
		candidates = append(candidates, found...)
	}
	return candidates
}

type saveRoot struct {
	dir     string
	version string
}

func saveGameRoots() []saveRoot {
	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return nil
		}
		base := filepath.Join(localAppData, "Metalhead")
		return []saveRoot{
			{filepath.Join(base, "Super Mega Baseball 3"), "smb3"},
			{filepath.Join(base, "Super Mega Baseball 4"), "smb4"},
		}
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return nil
		}
		base := filepath.Join(home, "Library", "Application Support", "Metalhead")
		return []saveRoot{
			{filepath.Join(base, "Super Mega Baseball 3"), "smb3"},
			{filepath.Join(base, "Super Mega Baseball 4"), "smb4"},
		}
	default: // Linux (typically via Steam/Proton)
		home, err := os.UserHomeDir()
		if err != nil {
			return nil
		}
		base := filepath.Join(home, ".local", "share", "Metalhead")
		return []saveRoot{
			{filepath.Join(base, "Super Mega Baseball 3"), "smb3"},
			{filepath.Join(base, "Super Mega Baseball 4"), "smb4"},
		}
	}
}

// walkForSaveFiles recurses into the given directory looking for .sav files.
// The game stores saves as: {version_root}/{steam_id}/savedata.sav or {league}.sav
func walkForSaveFiles(root, version string) []SaveGameCandidate {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil
	}

	var found []SaveGameCandidate
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".sav" {
			found = append(found, SaveGameCandidate{Path: path, GameVersion: version})
		}
		return nil
	})
	return found
}

// MasterDBPath returns the path to master.sqlite in the SMB4 installation
// directory at the given install root (e.g. from a Steam library scan).
// master.sqlite contains the league registry (t_league_savedatas).
func MasterDBPath(smb4InstallDir string) string {
	return filepath.Join(smb4InstallDir, "D3D12", "assets", "database", "baseball", "master.sqlite")
}
