package main

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
)

// detectLegacyDBPath returns the default SmbExplorerCompanion DB path on Windows
// if the file exists, otherwise returns "". Always returns "" on non-Windows platforms.
func detectLegacyDBPath() string {
	if runtime.GOOS != "windows" {
		return ""
	}
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return ""
	}
	candidate := filepath.Join(localAppData, "SmbExplorerCompanion", "SmbExplorerCompanion.db")
	if _, err := os.Stat(candidate); err != nil {
		return ""
	}
	return candidate
}

func generateUUID() string {
	return uuid.New().String()
}
