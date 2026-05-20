package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"smb-tools/internal/models"
)

// SaveGameCandidate is a potential SMB save file found on the filesystem.
type SaveGameCandidate struct {
	Path        string
	GameVersion models.GameVersion
}

// DiscoverSaveFiles returns default save file locations for SMB3 and SMB4 on
// the current platform, filtered to paths that actually exist on disk.
//
// Platform support:
//   - Windows: %LOCALAPPDATA%\Metalhead\Super Mega Baseball {3|4}\{steam_id}\
//   - Linux:   ~/.local/share/Metalhead\Super Mega Baseball {3|4}\{steam_id}\
//     (Steam via Proton — the most common non-Windows path)
//
// macOS is intentionally omitted: there is no confirmed evidence that SMB3 or
// SMB4 ship with native macOS support. If that changes, add a "darwin" case.
func DiscoverSaveFiles() ([]SaveGameCandidate, error) {
	roots, err := saveGameRoots()
	if err != nil {
		return nil, err
	}

	var candidates []SaveGameCandidate
	for _, root := range roots {
		found := WalkForSaveFiles(root.dir, root.version)
		candidates = append(candidates, found...)
	}
	return candidates, nil
}

type saveRoot struct {
	dir     string
	version models.GameVersion
}

func saveGameRoots() ([]saveRoot, error) {
	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return nil, fmt.Errorf("LOCALAPPDATA environment variable not set")
		}
		base := filepath.Join(localAppData, "Metalhead")
		return []saveRoot{
			{filepath.Join(base, "Super Mega Baseball 3"), models.GameVersionSMB3},
			{filepath.Join(base, "Super Mega Baseball 4"), models.GameVersionSMB4},
		}, nil

	case "linux":
		// Linux support is via Steam/Proton. Save paths mirror the Windows layout
		// under the Proton prefix, typically at ~/.local/share/Metalhead.
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolving home directory: %w", err)
		}
		base := filepath.Join(home, ".local", "share", "Metalhead")
		return []saveRoot{
			{filepath.Join(base, "Super Mega Baseball 3"), models.GameVersionSMB3},
			{filepath.Join(base, "Super Mega Baseball 4"), models.GameVersionSMB4},
		}, nil

	default:
		return nil, fmt.Errorf("save file auto-discovery is not supported on %s", runtime.GOOS)
	}
}

// WalkForSaveFiles recurses into root looking for .sav files.
// Exported so it can be called with a custom root in tests and in the UI
// when the user manually points to a save directory.
func WalkForSaveFiles(root string, version models.GameVersion) []SaveGameCandidate {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil
	}

	var found []SaveGameCandidate
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		// Only league save files are valid SMB franchise saves.
		// The naming convention (confirmed from SMB3Explorer) is league-{GUID}.sav.
		// master.sav, mugshots-*.sav, season-*.sav, and *.sav.bak are auxiliary
		// data files that cannot be associated with an app franchise.
		base := filepath.Base(path)
		if filepath.Ext(base) == ".sav" && strings.HasPrefix(base, "league-") {
			found = append(found, SaveGameCandidate{Path: path, GameVersion: version})
		}
		return nil
	})
	return found
}
