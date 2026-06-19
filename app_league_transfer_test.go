package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"smb-tools/internal/config"
	internaldb "smb-tools/internal/db"
	"smb-tools/internal/models"
	"smb-tools/internal/service"
	"smb-tools/internal/store"
	"smb-tools/internal/system"
	"smb-tools/internal/testutil"
)

// newTestAppForLeagueTransfer wires up just enough of App — a registry DB,
// a franchise store, and a LeagueTransferService — to exercise the
// snapshot-export bindings without going through full Wails startup.
func newTestAppForLeagueTransfer(t *testing.T) *App {
	t.Helper()
	tmp := t.TempDir()
	dirs := &config.AppDirs{
		DataDir:       tmp,
		RegistryPath:  filepath.Join(tmp, "registry.db"),
		FranchisesDir: filepath.Join(tmp, "franchises"),
	}

	ctx := context.Background()
	registryDB, err := internaldb.OpenRegistry(ctx, dirs.RegistryPath)
	if err != nil {
		t.Fatalf("opening registry db: %v", err)
	}
	t.Cleanup(func() { _ = registryDB.Close() })

	return &App{
		ctx:                   ctx,
		dirs:                  dirs,
		franchiseStore:        store.NewFranchiseStore(registryDB),
		leagueTransferService: service.NewLeagueTransferService(dirs, system.DefaultGameRunningChecker{}, "test-version", uuid.New),
	}
}

// addTestFranchiseWithSnapshot creates a franchise and a single snapshot for
// it: a raw (uncompressed) single-league SQLite file under its snapshots
// directory, recorded in its companion DB. Returns the franchise ID and the
// recorded snapshot's ID.
func addTestFranchiseWithSnapshot(t *testing.T, a *App, franchiseID, franchiseName string, seasonNum int, compressed bool) (string, int64) {
	t.Helper()
	if err := a.franchiseStore.Create(a.ctx, models.Franchise{
		ID:          franchiseID,
		Name:        franchiseName,
		GameVersion: models.GameVersionSMB4,
	}); err != nil {
		t.Fatalf("creating franchise: %v", err)
	}
	if err := a.dirs.EnsureFranchiseDirs(franchiseID); err != nil {
		t.Fatalf("creating franchise dirs: %v", err)
	}

	companionDB, err := internaldb.OpenCompanion(a.ctx, a.dirs.CompanionDBPath(franchiseID))
	if err != nil {
		t.Fatalf("opening companion db: %v", err)
	}
	defer func() { _ = companionDB.Close() }()

	fileName := store.SnapshotFileName("0001_test.sqlite")
	snapshotPath := filepath.Join(a.dirs.SnapshotsDir(franchiseID), string(fileName))

	// Build a raw, single-league snapshot file by decompressing the
	// standard compressed fixture (it seeds League A + League B; only
	// League A is kept, matching the real single-league-per-file shape).
	compressedPath := filepath.Join(t.TempDir(), "fixture.sav")
	testutil.WriteCompressedLeagueSaveFixture(t, compressedPath)
	rawPath, err := internaldb.DecompressToTempFile(compressedPath)
	if err != nil {
		t.Fatalf("decompressing fixture: %v", err)
	}
	defer func() { _ = os.Remove(rawPath) }()

	rawDB, err := internaldb.OpenForReadWrite(a.ctx, rawPath)
	if err != nil {
		t.Fatalf("opening raw fixture: %v", err)
	}
	leagueAGUID := uuid.MustParse("AA000000-0000-0000-0000-000000000000")
	for _, stmt := range []string{
		`DELETE FROM t_leagues WHERE GUID != ?`,
		`DELETE FROM t_conferences WHERE leagueGUID != ?`,
		`DELETE FROM t_franchise WHERE leagueGUID != ?`,
		`DELETE FROM t_seasons WHERE historicalLeagueGUID != ?`,
		`DELETE FROM t_league_local_ids WHERE GUID != ?`,
	} {
		if _, err := rawDB.Exec(stmt, leagueAGUID[:]); err != nil {
			_ = rawDB.Close()
			t.Fatalf("trimming fixture to a single league: %v", err)
		}
	}
	if err := rawDB.Close(); err != nil {
		t.Fatalf("closing raw fixture: %v", err)
	}

	data, err := os.ReadFile(rawPath)
	if err != nil {
		t.Fatalf("reading raw fixture: %v", err)
	}
	if err := os.WriteFile(snapshotPath, data, 0o600); err != nil {
		t.Fatalf("writing snapshot file: %v", err)
	}

	snapshotID, err := store.NewSnapshotStore(companionDB).Record(a.ctx, store.Snapshot{
		SeasonNum:     seasonNum,
		FileName:      fileName,
		SHA256Hash:    "test-hash",
		FileSizeBytes: int64(len(data)),
		Compressed:    compressed,
	})
	if err != nil {
		t.Fatalf("recording snapshot: %v", err)
	}
	return franchiseID, snapshotID
}

func TestApp_ListSnapshotExportCandidates_AcrossFranchises(t *testing.T) {
	a := newTestAppForLeagueTransfer(t)
	addTestFranchiseWithSnapshot(t, a, "franchise-1", "Franchise One", 3, false)
	addTestFranchiseWithSnapshot(t, a, "franchise-2", "Franchise Two", 7, false)

	candidates, err := a.ListSnapshotExportCandidates()
	if err != nil {
		t.Fatalf("ListSnapshotExportCandidates: %v", err)
	}
	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates across 2 franchises, got %d", len(candidates))
	}

	var sawFranchiseOne, sawFranchiseTwo bool
	for _, c := range candidates {
		switch c.FranchiseID {
		case "franchise-1":
			sawFranchiseOne = true
			if c.FranchiseName != "Franchise One" || c.SeasonNum != 3 {
				t.Errorf("franchise-1 candidate = %+v, want name %q season 3", c, "Franchise One")
			}
		case "franchise-2":
			sawFranchiseTwo = true
			if c.FranchiseName != "Franchise Two" || c.SeasonNum != 7 {
				t.Errorf("franchise-2 candidate = %+v, want name %q season 7", c, "Franchise Two")
			}
		}
	}
	if !sawFranchiseOne || !sawFranchiseTwo {
		t.Errorf("expected candidates from both franchises, got %+v", candidates)
	}
}

func TestApp_ListSnapshotExportCandidates_SkipsCompressedSnapshot(t *testing.T) {
	a := newTestAppForLeagueTransfer(t)
	addTestFranchiseWithSnapshot(t, a, "franchise-1", "Franchise One", 1, true)

	candidates, err := a.ListSnapshotExportCandidates()
	if err != nil {
		t.Fatalf("ListSnapshotExportCandidates: %v", err)
	}
	if len(candidates) != 0 {
		t.Errorf("expected compressed snapshot to be skipped, got %d candidates", len(candidates))
	}
}

func TestApp_ExportSnapshotAsLeague_HappyPath(t *testing.T) {
	a := newTestAppForLeagueTransfer(t)
	franchiseID, snapshotID := addTestFranchiseWithSnapshot(t, a, "franchise-1", "Franchise One", 5, false)

	outputPath, err := a.ExportSnapshotAsLeague(franchiseID, snapshotID, "Exported From Snapshot")
	if err != nil {
		t.Fatalf("ExportSnapshotAsLeague: %v", err)
	}

	preview, err := a.leagueTransferService.PreviewImport(a.ctx, outputPath)
	if err != nil {
		t.Fatalf("PreviewImport on snapshot export: %v", err)
	}
	if preview.Overview.Name != "Exported From Snapshot" {
		t.Errorf("Overview.Name = %q, want %q", preview.Overview.Name, "Exported From Snapshot")
	}
}

func TestApp_ExportSnapshotAsLeague_UnknownFranchise(t *testing.T) {
	a := newTestAppForLeagueTransfer(t)

	if _, err := a.ExportSnapshotAsLeague("does-not-exist", 1, "Name"); err == nil {
		t.Error("expected an error for an unknown franchise ID, got nil")
	}
}

func TestApp_ExportSnapshotAsLeague_UnknownSnapshot(t *testing.T) {
	a := newTestAppForLeagueTransfer(t)
	franchiseID, _ := addTestFranchiseWithSnapshot(t, a, "franchise-1", "Franchise One", 1, false)

	if _, err := a.ExportSnapshotAsLeague(franchiseID, 999, "Name"); err == nil {
		t.Error("expected an error for an unknown snapshot ID, got nil")
	}
}
