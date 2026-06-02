package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

// AppDirs holds all resolved application data directory paths.
type AppDirs struct {
	// DataDir is the root application data directory.
	DataDir string
	// RegistryPath is the path to the franchise registry database.
	RegistryPath string
	// FranchisesDir is the directory containing per-franchise subdirectories.
	FranchisesDir string
}

// NewAppDirs resolves platform-specific app data directories and ensures
// the required directories exist on disk.
func NewAppDirs() (*AppDirs, error) {
	dataDir, err := resolveDataDir()
	if err != nil {
		return nil, fmt.Errorf("resolving app data directory: %w", err)
	}
	franchisesDir := filepath.Join(dataDir, "franchises")
	if err := os.MkdirAll(franchisesDir, 0o700); err != nil {
		return nil, fmt.Errorf("creating franchises directory: %w", err)
	}
	return &AppDirs{
		DataDir:       dataDir,
		RegistryPath:  filepath.Join(dataDir, "registry.db"),
		FranchisesDir: franchisesDir,
	}, nil
}

// FranchiseDir returns the path to the given franchise's subdirectory.
func (d *AppDirs) FranchiseDir(id string) string {
	return filepath.Join(d.FranchisesDir, id)
}

// CompanionDBPath returns the path to the given franchise's companion database.
func (d *AppDirs) CompanionDBPath(franchiseID string) string {
	return filepath.Join(d.FranchiseDir(franchiseID), "companion.db")
}

// SnapshotsDir returns the path to the given franchise's snapshots directory.
func (d *AppDirs) SnapshotsDir(franchiseID string) string {
	return filepath.Join(d.FranchiseDir(franchiseID), "snapshots")
}

// AssetsDir returns the path to the given franchise's assets directory.
func (d *AppDirs) AssetsDir(franchiseID string) string {
	return filepath.Join(d.FranchiseDir(franchiseID), "assets")
}

// TeamLogosDir returns the path to the logo storage directory for a specific team.
func (d *AppDirs) TeamLogosDir(franchiseID string, teamID int) string {
	return filepath.Join(d.AssetsDir(franchiseID), "logos", strconv.Itoa(teamID))
}

// EnsureFranchiseDirs creates the per-franchise directory structure.
func (d *AppDirs) EnsureFranchiseDirs(franchiseID string) error {
	for _, dir := range []string{d.SnapshotsDir(franchiseID), d.AssetsDir(franchiseID)} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("creating franchise directories for %q: %w", franchiseID, err)
		}
	}
	return nil
}

func resolveDataDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		return filepath.Join(appdata, "smb-tools"), nil
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving home directory: %w", err)
		}
		return filepath.Join(home, "Library", "Application Support", "smb-tools"), nil
	default: // Linux and other Unix-like systems
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			return filepath.Join(xdg, "smb-tools"), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving home directory: %w", err)
		}
		return filepath.Join(home, ".local", "share", "smb-tools"), nil
	}
}
