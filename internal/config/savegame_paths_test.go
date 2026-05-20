package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"smb-tools/internal/config"
	"smb-tools/internal/models"
)

func TestWalkForSaveFiles_FindsLeagueSavFiles(t *testing.T) {
	root := t.TempDir()

	// Create a fake Steam ID subdirectory mirroring the real layout
	steamDir := filepath.Join(root, "76561198034146134")
	if err := os.MkdirAll(steamDir, 0o700); err != nil {
		t.Fatal(err)
	}
	leaguePath := filepath.Join(steamDir, "league-1D454F48-B9BD-42A0-A528-358E46142A64.sav")
	if err := os.WriteFile(leaguePath, []byte("fake"), 0o600); err != nil {
		t.Fatal(err)
	}
	// These should all be ignored: wrong prefix or not a real league save
	for _, name := range []string{
		"master.sav",
		"mugshots-0E7BD862-11CA-4A56-9483-B73710DB404F.sav",
		"season-CA454E09-CFBA-4F92-8DF0-07D2D5AA09E7.sav",
		"league-1D454F48-B9BD-42A0-A528-358E46142A64.sav.bak",
		"league-1D454F48-B9BD-42A0-A528-358E46142A64.hash",
		"config.json",
	} {
		if err := os.WriteFile(filepath.Join(steamDir, name), []byte("fake"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	candidates := config.WalkForSaveFiles(root, models.GameVersionSMB4)
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	if candidates[0].Path != leaguePath {
		t.Errorf("expected path %q, got %q", leaguePath, candidates[0].Path)
	}
	if candidates[0].GameVersion != models.GameVersionSMB4 {
		t.Errorf("expected version %q, got %q", models.GameVersionSMB4, candidates[0].GameVersion)
	}
}

func TestWalkForSaveFiles_NonExistentRoot(t *testing.T) {
	candidates := config.WalkForSaveFiles("/does/not/exist", models.GameVersionSMB3)
	if len(candidates) != 0 {
		t.Errorf("expected 0 candidates for non-existent root, got %d", len(candidates))
	}
}

func TestWalkForSaveFiles_EmptyDirectory(t *testing.T) {
	root := t.TempDir()
	candidates := config.WalkForSaveFiles(root, models.GameVersionSMB3)
	if len(candidates) != 0 {
		t.Errorf("expected 0 candidates in empty directory, got %d", len(candidates))
	}
}

func TestWalkForSaveFiles_NestedDirectories(t *testing.T) {
	root := t.TempDir()

	// .sav files can be nested multiple levels deep
	deep := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(deep, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"league-abc.sav", "league-def.sav"} {
		p := filepath.Join(deep, name)
		if err := os.WriteFile(p, []byte("fake"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	candidates := config.WalkForSaveFiles(root, models.GameVersionSMB4)
	if len(candidates) != 2 {
		t.Errorf("expected 2 candidates, got %d", len(candidates))
	}
}
