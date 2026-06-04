package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"smb-tools/internal/config"
	"smb-tools/internal/service"
	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

// newTestMediaService creates a MediaService backed by an in-memory companion DB
// and a temp directory on disk for file I/O.
func newTestMediaService(t *testing.T) (*service.MediaService, *config.AppDirs, string) {
	t.Helper()
	tmpDir := t.TempDir()

	dirs := &config.AppDirs{
		DataDir:       tmpDir,
		FranchisesDir: filepath.Join(tmpDir, "franchises"),
	}
	if err := os.MkdirAll(dirs.FranchisesDir, 0o700); err != nil {
		t.Fatalf("creating franchises dir: %v", err)
	}

	svc := service.NewMediaService(store.NewMediaStore(), dirs)
	return svc, dirs, tmpDir
}

// writeTestFile writes a file with the given name and dummy content to a temp dir.
func writeTestFile(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("dummy content"), 0o600); err != nil {
		t.Fatalf("writing test file %s: %v", name, err)
	}
	return path
}

// ── Extension validation ──────────────────────────────────────────────────────

func TestMediaService_UploadAndAssociate_InvalidExtensionRejected(t *testing.T) {
	svc, _, tmpDir := newTestMediaService(t)
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	srcPath := writeTestFile(t, tmpDir, "video.mkv")

	_, err := svc.UploadAndAssociate(ctx, db, "franchise-1", "Test", "", srcPath, nil, nil)
	if err == nil {
		t.Fatal("expected error for unsupported extension .mkv, got nil")
	}
}

func TestMediaService_UploadAndAssociate_SupportedExtensionsAccepted(t *testing.T) {
	svc, dirs, tmpDir := newTestMediaService(t)
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	for _, ext := range []string{"png", "jpg", "jpeg", "mp4", "mov", "avi"} {
		t.Run(ext, func(t *testing.T) {
			srcPath := writeTestFile(t, tmpDir, "file."+ext)
			m, err := svc.UploadAndAssociate(ctx, db, "franchise-ext", "Test "+ext, "", srcPath, nil, nil)
			if err != nil {
				t.Fatalf("UploadAndAssociate(.%s): %v", ext, err)
			}
			// Verify file was actually copied to MediaDir.
			mediaDir := dirs.MediaDir("franchise-ext")
			dstPath := filepath.Join(mediaDir, filepath.Base(filepath.FromSlash(m.FilePath[len("assets/media/"):])))
			if _, err := os.Stat(dstPath); os.IsNotExist(err) {
				t.Errorf("expected file to exist at %s", dstPath)
			}
		})
	}
}

// ── File copy and DB insert ───────────────────────────────────────────────────

func TestMediaService_UploadAndAssociate_CopiesFileToMediaDir(t *testing.T) {
	svc, dirs, tmpDir := newTestMediaService(t)
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	srcPath := writeTestFile(t, tmpDir, "shot.png")
	const franchiseID = "franchise-copy"

	m, err := svc.UploadAndAssociate(ctx, db, franchiseID, "My Screenshot", "Nice play", srcPath, nil, nil)
	if err != nil {
		t.Fatalf("UploadAndAssociate: %v", err)
	}
	if m.Name != "My Screenshot" {
		t.Errorf("name: want 'My Screenshot', got %q", m.Name)
	}
	if m.Description != "Nice play" {
		t.Errorf("description: want 'Nice play', got %q", m.Description)
	}

	// File must exist in the media directory.
	mediaDir := dirs.MediaDir(franchiseID)
	entries, err := os.ReadDir(mediaDir)
	if err != nil {
		t.Fatalf("reading media dir: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 file in media dir, got %d", len(entries))
	}
}

// ── Orphan cleanup on last association removal ────────────────────────────────

func TestMediaService_RemoveAssociation_LastAssoc_DeletesFile(t *testing.T) {
	svc, dirs, tmpDir := newTestMediaService(t)
	database := testutil.NewTestDB(t)
	ctx := context.Background()

	srcPath := writeTestFile(t, tmpDir, "clip.mp4")
	const franchiseID = "franchise-orphan"

	// Upload with no associations (zero team-season IDs, zero player IDs).
	// We'll manually insert a team-season assoc row using the store directly.
	ms := store.NewMediaStore()
	ss := store.NewSeasonStore(database)
	seasonID, err := ss.Upsert(ctx, store.Season{LeagueGUID: "L1", SaveGameSeasonID: 1, SeasonNum: 1})
	if err != nil {
		t.Fatalf("Upsert season: %v", err)
	}
	ts := store.NewTeamHistoryStore(database)
	teamID, _ := ts.UpsertTeam(ctx, "GUID-O", "Orphan Team")
	histID, err := ts.UpsertSeasonHistory(ctx, store.TeamSeasonHistory{
		TeamID: teamID, SeasonID: seasonID, TeamName: "Orphan Team",
	})
	if err != nil {
		t.Fatalf("UpsertSeasonHistory: %v", err)
	}

	m, err := svc.UploadAndAssociate(ctx, database, franchiseID, "Clip", "", srcPath, []int64{histID}, nil)
	if err != nil {
		t.Fatalf("UploadAndAssociate: %v", err)
	}

	// Verify the file exists before removal.
	mediaDir := dirs.MediaDir(franchiseID)
	entries, _ := os.ReadDir(mediaDir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 file before removal, got %d", len(entries))
	}

	// Remove the only association — should trigger orphan cleanup.
	if err := svc.RemoveAssociation(ctx, database, franchiseID, m.ID, "team_season", histID); err != nil {
		t.Fatalf("RemoveAssociation: %v", err)
	}

	// File must be gone.
	entries2, _ := os.ReadDir(mediaDir)
	if len(entries2) != 0 {
		t.Errorf("expected file deleted after orphan cleanup, but %d file(s) remain", len(entries2))
	}

	// DB record must be gone.
	_, err = ms.GetMediaByID(ctx, database, m.ID)
	if err == nil {
		t.Error("expected media record deleted, but GetMediaByID returned no error")
	}
}

// ── Partial removal — file persists ──────────────────────────────────────────

func TestMediaService_RemoveAssociation_NotLastAssoc_FileRemains(t *testing.T) {
	svc, dirs, tmpDir := newTestMediaService(t)
	database := testutil.NewTestDB(t)
	ctx := context.Background()

	ss := store.NewSeasonStore(database)
	seasonID, _ := ss.Upsert(ctx, store.Season{LeagueGUID: "L1", SaveGameSeasonID: 1, SeasonNum: 1})
	ts := store.NewTeamHistoryStore(database)
	teamID, _ := ts.UpsertTeam(ctx, "GUID-R", "Remain Team")
	histID, _ := ts.UpsertSeasonHistory(ctx, store.TeamSeasonHistory{
		TeamID: teamID, SeasonID: seasonID, TeamName: "Remain Team",
	})
	ps := store.NewPlayerSeasonStore(database)
	playerID, _ := ps.UpsertPlayer(ctx, store.PlayerIdentity{
		GameGUID: "P-R", FirstName: "Jane", LastName: "Smith",
	})

	srcPath := writeTestFile(t, tmpDir, "shot2.png")
	const franchiseID = "franchise-remain"

	m, err := svc.UploadAndAssociate(ctx, database, franchiseID, "Shot", "", srcPath,
		[]int64{histID}, []int64{playerID},
	)
	if err != nil {
		t.Fatalf("UploadAndAssociate: %v", err)
	}

	// Remove the team-season association; player association remains.
	if err := svc.RemoveAssociation(ctx, database, franchiseID, m.ID, "team_season", histID); err != nil {
		t.Fatalf("RemoveAssociation: %v", err)
	}

	// File must still exist.
	mediaDir := dirs.MediaDir(franchiseID)
	entries, _ := os.ReadDir(mediaDir)
	if len(entries) != 1 {
		t.Errorf("expected file to remain after partial removal, got %d files", len(entries))
	}
}

// ── DeleteMediaEverywhere ─────────────────────────────────────────────────────

func TestMediaService_DeleteMediaEverywhere_DeletesFileAndRows(t *testing.T) {
	svc, dirs, tmpDir := newTestMediaService(t)
	database := testutil.NewTestDB(t)
	ctx := context.Background()
	ms := store.NewMediaStore()

	srcPath := writeTestFile(t, tmpDir, "gone.png")
	const franchiseID = "franchise-delete"

	m, err := svc.UploadAndAssociate(ctx, database, franchiseID, "Gone", "", srcPath, nil, nil)
	if err != nil {
		t.Fatalf("UploadAndAssociate: %v", err)
	}

	if err := svc.DeleteMediaEverywhere(ctx, database, franchiseID, m.ID); err != nil {
		t.Fatalf("DeleteMediaEverywhere: %v", err)
	}

	mediaDir := dirs.MediaDir(franchiseID)
	entries, _ := os.ReadDir(mediaDir)
	if len(entries) != 0 {
		t.Errorf("expected 0 files after DeleteMediaEverywhere, got %d", len(entries))
	}

	_, err = ms.GetMediaByID(ctx, database, m.ID)
	if err == nil {
		t.Error("expected media record deleted, but GetMediaByID returned no error")
	}
}
