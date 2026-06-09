package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"smb-tools/internal/config"
	"smb-tools/internal/models"
)

// createFile writes a zero-byte file at path, creating parent dirs as needed.
func createFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("fake"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestScanDirShallow_EmptyDir(t *testing.T) {
	root := t.TempDir()
	got := config.ScanDirShallow(root, models.GameVersionSMB4)
	if got == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
	if len(got) != 0 {
		t.Errorf("expected 0 candidates, got %d", len(got))
	}
}

func TestScanDirShallow_NonExistentRoot(t *testing.T) {
	got := config.ScanDirShallow("/does/not/exist", models.GameVersionSMB4)
	if got == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
	if len(got) != 0 {
		t.Errorf("expected 0 candidates for non-existent root, got %d", len(got))
	}
}

func TestScanDirShallow_FindsFilesAtTopLevel(t *testing.T) {
	// Simulates: user points directly at their steam ID dir where saves live flat.
	root := t.TempDir()
	leaguePath := filepath.Join(root, "league-1D454F48-B9BD-42A0-A528-358E46142A64.sav")
	createFile(t, leaguePath)

	got := config.ScanDirShallow(root, models.GameVersionSMB4)
	if len(got) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(got))
	}
	if got[0].Path != leaguePath {
		t.Errorf("expected path %q, got %q", leaguePath, got[0].Path)
	}
	if got[0].GameVersion != models.GameVersionSMB4 {
		t.Errorf("expected version %q, got %q", models.GameVersionSMB4, got[0].GameVersion)
	}
}

func TestScanDirShallow_FindsFilesInImmediateSubdir(t *testing.T) {
	// Simulates: auto-discovery root is "Super Mega Baseball 4/" and saves are
	// one level down in a steam user ID subdirectory — the real SMB4 layout.
	root := t.TempDir()
	steamDir := filepath.Join(root, "76561198034146134")
	leaguePath := filepath.Join(steamDir, "league-1D454F48-B9BD-42A0-A528-358E46142A64.sav")
	createFile(t, leaguePath)

	// Non-matching files in same subdir should be ignored.
	for _, name := range []string{
		"master.sav",
		"mugshots-0E7BD862-11CA-4A56-9483-B73710DB404F.sav",
		"season-CA454E09-CFBA-4F92-8DF0-07D2D5AA09E7.sav",
		"league-1D454F48-B9BD-42A0-A528-358E46142A64.sav.bak",
		"league-1D454F48-B9BD-42A0-A528-358E46142A64.hash",
		"config.json",
	} {
		createFile(t, filepath.Join(steamDir, name))
	}

	got := config.ScanDirShallow(root, models.GameVersionSMB4)
	if len(got) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(got))
	}
	if got[0].Path != leaguePath {
		t.Errorf("expected path %q, got %q", leaguePath, got[0].Path)
	}
}

func TestScanDirShallow_DoesNotRecurseDeeper(t *testing.T) {
	// league-*.sav files more than one level deep must not be returned.
	root := t.TempDir()
	deep := filepath.Join(root, "a", "b", "c")
	createFile(t, filepath.Join(deep, "league-abc.sav"))
	createFile(t, filepath.Join(deep, "league-def.sav"))

	got := config.ScanDirShallow(root, models.GameVersionSMB4)
	if len(got) != 0 {
		t.Errorf("expected 0 candidates (no recursion beyond one level), got %d", len(got))
	}
}

func TestScanDirShallow_IgnoresNonMatchingFiles(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{
		"master.sav",
		"mugshots-abc.sav",
		"season-xyz.sav",
		"league-abc.sav.bak",
		"league-abc.hash",
		"config.json",
		"not-a-league.sav",
	} {
		createFile(t, filepath.Join(root, name))
	}

	got := config.ScanDirShallow(root, models.GameVersionSMB4)
	if len(got) != 0 {
		t.Errorf("expected 0 candidates, got %d: %v", len(got), got)
	}
}

func TestScanDirShallow_MultipleSteamAccounts(t *testing.T) {
	// Multiple steam ID subdirs — all their saves should be found.
	root := t.TempDir()
	paths := []string{
		filepath.Join(root, "76561198034146134", "league-aaa.sav"),
		filepath.Join(root, "76561198034146134", "league-bbb.sav"),
		filepath.Join(root, "76561199012345678", "league-ccc.sav"),
	}
	for _, p := range paths {
		createFile(t, p)
	}

	got := config.ScanDirShallow(root, models.GameVersionSMB4)
	if len(got) != 3 {
		t.Fatalf("expected 3 candidates, got %d", len(got))
	}
}

func TestDiscoverSaveFiles_NeverReturnsNil(t *testing.T) {
	// Regression for the Ubuntu crash: GetSaveFileCandidates was serializing a
	// nil Go slice to JSON null, causing null.filter() to throw in the frontend.
	// DiscoverSaveFiles must always return a non-nil slice even when platform
	// directories are absent.
	got, _ := config.DiscoverSaveFiles()
	if got == nil {
		t.Fatal("DiscoverSaveFiles returned nil slice — Wails serializes this as JSON null")
	}
}
