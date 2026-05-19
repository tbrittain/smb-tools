package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"smb-tools/internal/config"
	"smb-tools/internal/models"
)

func TestWalkForSaveFiles_FindsSavFiles(t *testing.T) {
	root := t.TempDir()

	// Create a fake Steam ID subdirectory with a .sav file
	steamDir := filepath.Join(root, "12345678")
	if err := os.MkdirAll(steamDir, 0o700); err != nil {
		t.Fatal(err)
	}
	savPath := filepath.Join(steamDir, "savedata.sav")
	if err := os.WriteFile(savPath, []byte("fake"), 0o600); err != nil {
		t.Fatal(err)
	}
	// Also create a non-.sav file that should be ignored
	if err := os.WriteFile(filepath.Join(steamDir, "config.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	candidates := config.WalkForSaveFiles(root, models.GameVersionSMB4)
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(candidates))
	}
	if candidates[0].Path != savPath {
		t.Errorf("expected path %q, got %q", savPath, candidates[0].Path)
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
