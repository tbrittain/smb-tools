package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	smbExplorerDirName = "SmbExplorerCompanion"
	smbToolsDirName    = "SmbTools"
)

func getAppDataDir(appName string) (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appDataDir := filepath.Join(dir, appName)

	err = os.MkdirAll(appDataDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	return appDataDir, nil
}

func GetSmbToolsDir() (string, error) {
	dir, err := getAppDataDir(smbToolsDirName)
	if err != nil {
		return "", fmt.Errorf("failed to get SMB Tools directory: %w", err)
	}

	return dir, nil
}

// this is Windows-specific, so it will short-circuit on non-Windows systems
func getLocalAppDataDir(appName string) (string, error) {
	// get OS
	if runtime.GOOS == "windows" {
		return "", fmt.Errorf("getLocalAppDataDir is only supported on Windows systems")
	}

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return "", fmt.Errorf("LOCALAPPDATA environment variable is not set")
	}

	// Append the application name to create a unique directory
	appDataDir := filepath.Join(localAppData, appName)

	// not creating it, since it is in a read-only state for this app
	// but short circuit here if it does not already exist
	if _, err := os.Stat(appDataDir); os.IsNotExist(err) {
		return "", fmt.Errorf("application data directory does not exist: %s", appDataDir)
	}

	return appDataDir, nil
}

func GetSmbExplorerCompanionDir() (string, error) {
	dir, err := getLocalAppDataDir(smbExplorerDirName)
	if err != nil {
		return "", fmt.Errorf("failed to get SMB Explorer Companion directory: %w", err)
	}

	return dir, nil
}
