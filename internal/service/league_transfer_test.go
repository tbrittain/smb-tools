package service_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/uuid"

	"smb-tools/internal/config"
	"smb-tools/internal/db"
	"smb-tools/internal/models"
	"smb-tools/internal/service"
	"smb-tools/internal/store"
	"smb-tools/internal/system"
	"smb-tools/internal/testutil"
)

// leagueAFixtureGUID matches the GUID NewTestLeagueSaveDB/WriteCompressedLeagueSaveFixture
// seeds as "League A" (see internal/testutil/leaguesave.go).
var leagueAFixtureGUID = uuid.MustParse("AA000000-0000-0000-0000-000000000000")

type fakeGameRunningChecker struct {
	running bool
	err     error
}

func (f fakeGameRunningChecker) IsGameRunning() (bool, error) { return f.running, f.err }

var _ system.GameRunningChecker = fakeGameRunningChecker{}

// setSMB4RootForLeagueTransfer mirrors internal/config's own test helper
// (not exported across packages) so DiscoverLeagues exercises the real
// config.DiscoverSaveFiles path deterministically on both Windows (dev) and
// Linux (CI).
func setSMB4RootForLeagueTransfer(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	switch runtime.GOOS {
	case "windows":
		t.Setenv("LOCALAPPDATA", tmp)
		return filepath.Join(tmp, "Metalhead", "Super Mega Baseball 4")
	default: // linux and other unix-like — saveGameRoots' default branch reads $HOME, not XDG_DATA_HOME
		t.Setenv("HOME", tmp)
		return filepath.Join(tmp, ".local", "share", "Metalhead", "Super Mega Baseball 4")
	}
}

func newTestLeagueTransferService(t *testing.T, gameRunning bool) (*service.LeagueTransferService, *config.AppDirs) {
	t.Helper()
	dirs := &config.AppDirs{DataDir: filepath.Join(t.TempDir(), "appdata")}
	svc := service.NewLeagueTransferService(dirs, fakeGameRunningChecker{running: gameRunning}, "test-version", uuid.New)
	return svc, dirs
}

// newTestDevLeagueTransferService behaves like newTestLeagueTransferService,
// but reports its version as "dev" and always assigns regeneratedGUID rather
// than a random one — used to deterministically exercise ExportLeague's and
// ExportLeagueWithRename's dev-only GUID regeneration.
func newTestDevLeagueTransferService(t *testing.T, regeneratedGUID uuid.UUID) (*service.LeagueTransferService, *config.AppDirs) {
	t.Helper()
	dirs := &config.AppDirs{DataDir: filepath.Join(t.TempDir(), "appdata")}
	svc := service.NewLeagueTransferService(dirs, fakeGameRunningChecker{running: false}, "dev", func() uuid.UUID { return regeneratedGUID })
	return svc, dirs
}

func TestLeagueTransferService_DiscoverLeagues(t *testing.T) {
	smb4Root := setSMB4RootForLeagueTransfer(t)
	steamDir := filepath.Join(smb4Root, "76561198034146134")
	if err := os.MkdirAll(steamDir, 0o700); err != nil {
		t.Fatalf("creating steam dir: %v", err)
	}
	testutil.WriteCompressedLeagueSaveFixture(t, filepath.Join(steamDir, "league-AA000000-0000-0000-0000-000000000000.sav"))

	svc, _ := newTestLeagueTransferService(t, false)
	overviews, err := svc.DiscoverLeagues(context.Background())
	if err != nil {
		t.Fatalf("DiscoverLeagues: %v", err)
	}
	if len(overviews) != 1 {
		t.Fatalf("expected 1 league, got %d", len(overviews))
	}
	if overviews[0].Name != "League A" {
		t.Errorf("Name = %q, want %q", overviews[0].Name, "League A")
	}
	if len(overviews[0].Conferences) != 2 {
		t.Errorf("expected 2 conferences, got %d", len(overviews[0].Conferences))
	}
	if overviews[0].Mode != models.LeagueModeFranchise {
		t.Errorf("Mode = %q, want %q", overviews[0].Mode, models.LeagueModeFranchise)
	}
}

func TestLeagueTransferService_DiscoverLeagues_NoneFound(t *testing.T) {
	setSMB4RootForLeagueTransfer(t)
	svc, _ := newTestLeagueTransferService(t, false)

	overviews, err := svc.DiscoverLeagues(context.Background())
	if err != nil {
		t.Fatalf("DiscoverLeagues: %v", err)
	}
	if overviews == nil {
		t.Fatal("expected a non-nil slice")
	}
	if len(overviews) != 0 {
		t.Errorf("expected 0 leagues, got %d", len(overviews))
	}
}

func setUpSourceLeagueFiles(t *testing.T, dir string, guid uuid.UUID) string {
	t.Helper()
	upper := guid.String()
	savPath := filepath.Join(dir, "league-"+upper+".sav")
	bakPath := savPath + ".bak"

	testutil.WriteCompressedLeagueSaveFixture(t, savPath)
	// Reuse the same compressed bytes for .bak — content doesn't matter for
	// these tests, only that a valid, zlib-compressed sibling file exists.
	data, err := os.ReadFile(savPath)
	if err != nil {
		t.Fatalf("reading generated sav fixture: %v", err)
	}
	if err := os.WriteFile(bakPath, data, 0o600); err != nil {
		t.Fatalf("writing bak fixture: %v", err)
	}
	return savPath
}

func TestLeagueTransferService_ExportLeague(t *testing.T) {
	svc, dirs := newTestLeagueTransferService(t, false)
	sourceDir := t.TempDir()
	savPath := setUpSourceLeagueFiles(t, sourceDir, leagueAFixtureGUID)

	outputPath, err := svc.ExportLeague(context.Background(), leagueAFixtureGUID, savPath)
	if err != nil {
		t.Fatalf("ExportLeague: %v", err)
	}
	if filepath.Dir(outputPath) != dirs.ExportsOutputDir() {
		t.Errorf("export written to %q, want it under %q", outputPath, dirs.ExportsOutputDir())
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Errorf("expected export zip to exist: %v", err)
	}
}

func TestLeagueTransferService_ExportLeagueWithRename(t *testing.T) {
	svc, _ := newTestLeagueTransferService(t, false)
	sourceDir := t.TempDir()
	savPath := setUpSourceLeagueFiles(t, sourceDir, leagueAFixtureGUID)

	beforeBytes, err := os.ReadFile(savPath)
	if err != nil {
		t.Fatalf("reading source .sav before export: %v", err)
	}

	outputPath, err := svc.ExportLeagueWithRename(context.Background(), leagueAFixtureGUID, savPath, "  Renamed League  ")
	if err != nil {
		t.Fatalf("ExportLeagueWithRename: %v", err)
	}

	afterBytes, err := os.ReadFile(savPath)
	if err != nil {
		t.Fatalf("reading source .sav after export: %v", err)
	}
	if string(beforeBytes) != string(afterBytes) {
		t.Error("source .sav was modified — ExportLeagueWithRename must leave it untouched")
	}

	preview, err := svc.PreviewImport(context.Background(), outputPath)
	if err != nil {
		t.Fatalf("PreviewImport on renamed export: %v", err)
	}
	if preview.Overview.Name != "Renamed League" {
		t.Errorf("Overview.Name = %q, want %q (trimmed)", preview.Overview.Name, "Renamed League")
	}
}

func TestLeagueTransferService_ExportLeagueWithRename_EmptyName(t *testing.T) {
	svc, _ := newTestLeagueTransferService(t, false)
	sourceDir := t.TempDir()
	savPath := setUpSourceLeagueFiles(t, sourceDir, leagueAFixtureGUID)

	if _, err := svc.ExportLeagueWithRename(context.Background(), leagueAFixtureGUID, savPath, "   "); err == nil {
		t.Error("expected an error for a blank new name, got nil")
	}
}

func TestLeagueTransferService_ExportLeague_DevModeRegeneratesGUID(t *testing.T) {
	regeneratedGUID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	svc, _ := newTestDevLeagueTransferService(t, regeneratedGUID)
	sourceDir := t.TempDir()
	savPath := setUpSourceLeagueFiles(t, sourceDir, leagueAFixtureGUID)

	beforeBytes, err := os.ReadFile(savPath)
	if err != nil {
		t.Fatalf("reading source .sav before export: %v", err)
	}

	outputPath, err := svc.ExportLeague(context.Background(), leagueAFixtureGUID, savPath)
	if err != nil {
		t.Fatalf("ExportLeague: %v", err)
	}

	afterBytes, err := os.ReadFile(savPath)
	if err != nil {
		t.Fatalf("reading source .sav after export: %v", err)
	}
	if string(beforeBytes) != string(afterBytes) {
		t.Error("source .sav was modified — dev-mode GUID regeneration must leave it untouched")
	}

	preview, err := svc.PreviewImport(context.Background(), outputPath)
	if err != nil {
		t.Fatalf("PreviewImport on dev export: %v", err)
	}
	if preview.Overview.GUID != regeneratedGUID {
		t.Errorf("Overview.GUID = %s, want regenerated GUID %s", preview.Overview.GUID, regeneratedGUID)
	}
	if preview.Overview.Name != "League A" {
		t.Errorf("Overview.Name = %q, want %q", preview.Overview.Name, "League A")
	}
}

func TestLeagueTransferService_ExportLeagueWithRename_DevModeRegeneratesGUID(t *testing.T) {
	regeneratedGUID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	svc, _ := newTestDevLeagueTransferService(t, regeneratedGUID)
	sourceDir := t.TempDir()
	savPath := setUpSourceLeagueFiles(t, sourceDir, leagueAFixtureGUID)

	outputPath, err := svc.ExportLeagueWithRename(context.Background(), leagueAFixtureGUID, savPath, "Renamed Dev League")
	if err != nil {
		t.Fatalf("ExportLeagueWithRename: %v", err)
	}

	preview, err := svc.PreviewImport(context.Background(), outputPath)
	if err != nil {
		t.Fatalf("PreviewImport on dev export: %v", err)
	}
	if preview.Overview.GUID != regeneratedGUID {
		t.Errorf("Overview.GUID = %s, want regenerated GUID %s", preview.Overview.GUID, regeneratedGUID)
	}
	if preview.Overview.Name != "Renamed Dev League" {
		t.Errorf("Overview.Name = %q, want %q", preview.Overview.Name, "Renamed Dev League")
	}
}

func TestLeagueTransferService_ExportLeague_GUIDMismatch(t *testing.T) {
	svc, _ := newTestLeagueTransferService(t, false)
	sourceDir := t.TempDir()
	savPath := setUpSourceLeagueFiles(t, sourceDir, leagueAFixtureGUID)

	wrongGUID := uuid.New()
	if _, err := svc.ExportLeague(context.Background(), wrongGUID, savPath); err == nil {
		t.Error("expected an error when guid doesn't match the file name, got nil")
	}
}

func exportSampleLeague(t *testing.T, svc *service.LeagueTransferService) string {
	t.Helper()
	sourceDir := t.TempDir()
	savPath := setUpSourceLeagueFiles(t, sourceDir, leagueAFixtureGUID)
	zipPath, err := svc.ExportLeague(context.Background(), leagueAFixtureGUID, savPath)
	if err != nil {
		t.Fatalf("ExportLeague (test setup): %v", err)
	}
	return zipPath
}

func TestLeagueTransferService_PreviewImport(t *testing.T) {
	svc, _ := newTestLeagueTransferService(t, false)
	zipPath := exportSampleLeague(t, svc)

	preview, err := svc.PreviewImport(context.Background(), zipPath)
	if err != nil {
		t.Fatalf("PreviewImport: %v", err)
	}
	if preview.Overview.Name != "League A" {
		t.Errorf("Overview.Name = %q, want %q", preview.Overview.Name, "League A")
	}
	if preview.ExportedAt == "" {
		t.Error("expected a non-empty ExportedAt")
	}
}

func TestLeagueTransferService_PreviewImport_FlagsAlreadyRegisteredTarget(t *testing.T) {
	svc, _ := newTestLeagueTransferService(t, false)

	smb4Root := setSMB4RootForLeagueTransfer(t)
	steamDir := filepath.Join(smb4Root, "76561198034146134")
	if err := os.MkdirAll(steamDir, 0o700); err != nil {
		t.Fatalf("creating steam dir: %v", err)
	}
	// Register leagueAFixtureGUID in this target's master.sav ahead of time.
	testutil.WriteCompressedMasterSaveFixture(t, filepath.Join(steamDir, "master.sav"), leagueAFixtureGUID)

	zipPath := exportSampleLeague(t, svc)
	preview, err := svc.PreviewImport(context.Background(), zipPath)
	if err != nil {
		t.Fatalf("PreviewImport: %v", err)
	}
	if len(preview.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(preview.Targets))
	}
	if !preview.Targets[0].AlreadyRegistered {
		t.Error("expected AlreadyRegistered = true")
	}
}

func TestLeagueTransferService_ConfirmImport_HappyPath(t *testing.T) {
	svc, dirs := newTestLeagueTransferService(t, false)
	zipPath := exportSampleLeague(t, svc)

	targetDir := t.TempDir()
	masterSavePath := filepath.Join(targetDir, "master.sav")
	testutil.WriteCompressedMasterSaveFixture(t, masterSavePath) // does not include leagueAFixtureGUID

	if err := svc.ConfirmImport(context.Background(), zipPath, targetDir); err != nil {
		t.Fatalf("ConfirmImport: %v", err)
	}

	upper := leagueAFixtureGUID.String()
	for _, suffix := range []string{".sav", ".sav.bak"} {
		p := filepath.Join(targetDir, "league-"+upper+suffix)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected %s to exist: %v", p, err)
		}
	}

	// master.sav must now have the league registered.
	tmpPath, err := db.DecompressToTempFile(masterSavePath)
	if err != nil {
		t.Fatalf("decompressing resulting master.sav: %v", err)
	}
	defer func() { _ = os.Remove(tmpPath) }()
	registryDB, err := db.OpenForReadWrite(context.Background(), tmpPath)
	if err != nil {
		t.Fatalf("opening resulting master.sav: %v", err)
	}
	defer func() { _ = registryDB.Close() }()

	exists, err := store.NewLeagueRegistryStore(registryDB).LeagueExists(context.Background(), leagueAFixtureGUID)
	if err != nil {
		t.Fatalf("LeagueExists: %v", err)
	}
	if !exists {
		t.Error("expected the league to be registered in master.sav after import")
	}

	// A timestamped backup must exist.
	backups, err := os.ReadDir(dirs.MasterSaveBackupsDir())
	if err != nil {
		t.Fatalf("reading backups dir: %v", err)
	}
	if len(backups) != 1 {
		t.Errorf("expected 1 backup file, got %d", len(backups))
	}
}

func TestLeagueTransferService_ConfirmImport_MultipleImportsKeepDistinctBackups(t *testing.T) {
	svc, dirs := newTestLeagueTransferService(t, false)

	targetDir := t.TempDir()
	masterSavePath := filepath.Join(targetDir, "master.sav")
	testutil.WriteCompressedMasterSaveFixture(t, masterSavePath)

	// Import two different leagues into the same target, one after another.
	zipPath1 := exportSampleLeague(t, svc)
	if err := svc.ConfirmImport(context.Background(), zipPath1, targetDir); err != nil {
		t.Fatalf("first ConfirmImport: %v", err)
	}

	// A second, independent export of the same league GUID — imported into
	// a different target directory, so no collision occurs. This test only
	// needs a second successful import to confirm a second, distinct backup
	// gets written to the shared backup history.
	zipPath2 := exportSampleLeague(t, svc)

	targetDir2 := t.TempDir()
	masterSavePath2 := filepath.Join(targetDir2, "master.sav")
	testutil.WriteCompressedMasterSaveFixture(t, masterSavePath2)
	if err := svc.ConfirmImport(context.Background(), zipPath2, targetDir2); err != nil {
		t.Fatalf("second ConfirmImport (different target): %v", err)
	}

	entries, err := os.ReadDir(dirs.MasterSaveBackupsDir())
	if err != nil {
		t.Fatalf("reading backups dir: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 distinct backup files across two imports, got %d", len(entries))
	}
	if entries[0].Name() == entries[1].Name() {
		t.Error("expected two distinct backup file names, got the same name twice")
	}
}

func TestLeagueTransferService_ConfirmImport_RejectsCollision(t *testing.T) {
	svc, dirs := newTestLeagueTransferService(t, false)
	zipPath := exportSampleLeague(t, svc)

	targetDir := t.TempDir()
	masterSavePath := filepath.Join(targetDir, "master.sav")
	testutil.WriteCompressedMasterSaveFixture(t, masterSavePath, leagueAFixtureGUID) // already registered

	beforeBytes, err := os.ReadFile(masterSavePath)
	if err != nil {
		t.Fatalf("reading master.sav before attempt: %v", err)
	}

	err = svc.ConfirmImport(context.Background(), zipPath, targetDir)
	if err == nil {
		t.Fatal("expected an error for an already-registered league, got nil")
	}

	afterBytes, err := os.ReadFile(masterSavePath)
	if err != nil {
		t.Fatalf("reading master.sav after attempt: %v", err)
	}
	if string(beforeBytes) != string(afterBytes) {
		t.Error("master.sav was modified despite the collision — it must be left completely untouched")
	}

	upper := leagueAFixtureGUID.String()
	if _, err := os.Stat(filepath.Join(targetDir, "league-"+upper+".sav")); !os.IsNotExist(err) {
		t.Error("expected no league files to have been copied into the target directory")
	}

	assertNoBackupFilesWritten(t, dirs)
}

func TestLeagueTransferService_ConfirmImport_RefusesWhenGameRunning(t *testing.T) {
	dirs := &config.AppDirs{DataDir: filepath.Join(t.TempDir(), "appdata")}
	svc := service.NewLeagueTransferService(dirs, fakeGameRunningChecker{running: true}, "test-version", uuid.New)

	exportSvc := service.NewLeagueTransferService(dirs, fakeGameRunningChecker{running: false}, "test-version", uuid.New)
	zipPath := exportSampleLeague(t, exportSvc)

	targetDir := t.TempDir()
	masterSavePath := filepath.Join(targetDir, "master.sav")
	testutil.WriteCompressedMasterSaveFixture(t, masterSavePath)
	beforeBytes, err := os.ReadFile(masterSavePath)
	if err != nil {
		t.Fatalf("reading master.sav before attempt: %v", err)
	}

	err = svc.ConfirmImport(context.Background(), zipPath, targetDir)
	if err == nil {
		t.Fatal("expected an error when the game is running, got nil")
	}

	afterBytes, err := os.ReadFile(masterSavePath)
	if err != nil {
		t.Fatalf("reading master.sav after attempt: %v", err)
	}
	if string(beforeBytes) != string(afterBytes) {
		t.Error("master.sav was modified despite the game-running check — it must be left completely untouched")
	}

	assertNoBackupFilesWritten(t, dirs)
}

// assertNoBackupFilesWritten checks that no backup file was written under
// MasterSaveBackupsDir. The directory itself may already exist — test setup
// calling ExportLeague creates the whole league-transfer directory pair via
// EnsureLeagueTransferDirs — so this checks for emptiness, not absence.
func assertNoBackupFilesWritten(t *testing.T, dirs *config.AppDirs) {
	t.Helper()
	entries, err := os.ReadDir(dirs.MasterSaveBackupsDir())
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		t.Fatalf("reading backups dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected no backup files, found %d", len(entries))
	}
}
