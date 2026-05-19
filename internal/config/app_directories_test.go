package config_test

import (
	"strings"
	"testing"

	"smb-tools/internal/config"
)

func TestNewAppDirs_PathsAreNonEmpty(t *testing.T) {
	dirs, err := config.NewAppDirs()
	if err != nil {
		t.Fatalf("NewAppDirs: %v", err)
	}
	if dirs.DataDir == "" {
		t.Error("DataDir should not be empty")
	}
	if dirs.RegistryPath == "" {
		t.Error("RegistryPath should not be empty")
	}
	if dirs.FranchisesDir == "" {
		t.Error("FranchisesDir should not be empty")
	}
}

func TestNewAppDirs_RegistryUnderDataDir(t *testing.T) {
	dirs, err := config.NewAppDirs()
	if err != nil {
		t.Fatalf("NewAppDirs: %v", err)
	}
	if !strings.HasPrefix(dirs.RegistryPath, dirs.DataDir) {
		t.Errorf("RegistryPath %q should be under DataDir %q", dirs.RegistryPath, dirs.DataDir)
	}
	if !strings.HasPrefix(dirs.FranchisesDir, dirs.DataDir) {
		t.Errorf("FranchisesDir %q should be under DataDir %q", dirs.FranchisesDir, dirs.DataDir)
	}
}

func TestAppDirs_FranchisePaths(t *testing.T) {
	dirs, err := config.NewAppDirs()
	if err != nil {
		t.Fatalf("NewAppDirs: %v", err)
	}
	const id = "test-franchise-123"

	companionPath := dirs.CompanionDBPath(id)
	snapshotsDir := dirs.SnapshotsDir(id)

	if !strings.Contains(companionPath, id) {
		t.Errorf("CompanionDBPath %q should contain franchise ID %q", companionPath, id)
	}
	if !strings.Contains(snapshotsDir, id) {
		t.Errorf("SnapshotsDir %q should contain franchise ID %q", snapshotsDir, id)
	}
	if !strings.HasPrefix(companionPath, dirs.FranchisesDir) {
		t.Errorf("CompanionDBPath %q should be under FranchisesDir %q", companionPath, dirs.FranchisesDir)
	}
}
