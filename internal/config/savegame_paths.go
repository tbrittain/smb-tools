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

	candidates := []SaveGameCandidate{}
	for _, root := range roots {
		found := ScanDirShallow(root.dir, root.version)
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

// ScanDirShallow finds league-*.sav files in root and its immediate
// subdirectories. It does not recurse further — the SMB4 save layout is
// {game_dir}/{steam_id}/league-*.sav, so at most one subdir level is needed.
// Always returns a non-nil slice.
func ScanDirShallow(root string, version models.GameVersion) []SaveGameCandidate {
	entries, err := os.ReadDir(root)
	if err != nil {
		return []SaveGameCandidate{}
	}
	out := []SaveGameCandidate{}
	for _, e := range entries {
		fullPath := filepath.Join(root, e.Name())
		if !e.IsDir() {
			if isSaveFile(e.Name()) {
				out = append(out, SaveGameCandidate{Path: fullPath, GameVersion: version})
			}
		} else {
			subEntries, err := os.ReadDir(fullPath)
			if err != nil {
				continue
			}
			for _, sub := range subEntries {
				if !sub.IsDir() && isSaveFile(sub.Name()) {
					out = append(out, SaveGameCandidate{
						Path:        filepath.Join(fullPath, sub.Name()),
						GameVersion: version,
					})
				}
			}
		}
	}
	return out
}

// isSaveFile reports whether name is a league save file (league-*.sav).
// master.sav, mugshots-*.sav, season-*.sav, and *.sav.bak are excluded.
func isSaveFile(name string) bool {
	return filepath.Ext(name) == ".sav" && strings.HasPrefix(name, "league-")
}
