package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"smb-tools/internal/service"
	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

// leagueGUIDFixture is the hex-encoded GUID of the test franchise league,
// matching the X'EE000000000000000000000000000000' blob in the fixture.
const leagueGUIDFixture = "EE000000000000000000000000000000"

// newTestSyncService builds a SyncService backed by in-memory databases and a
// real temp directory for snapshot files. It returns the service, the companion
// DB (for post-sync assertions), and the snapshot directory path.
func newTestSyncService(t *testing.T) (*service.SyncService, *store.SnapshotStore, string) {
	t.Helper()
	companionDB := testutil.NewTestDB(t)
	snapshotDir := t.TempDir()
	snapshotStore := store.NewSnapshotStore(companionDB)
	snapshotSvc := service.NewSnapshotService(snapshotDir, snapshotStore)
	syncSvc := service.NewSyncService(snapshotSvc, service.NewImportService())
	return syncSvc, snapshotStore, snapshotDir
}

// writeSaveFile writes bytes to a temp file and returns the path. Used to
// simulate a decompressed save game file for TakeSnapshotFromFile.
func writeSaveFile(t *testing.T, content []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "save.sqlite")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("writeSaveFile: %v", err)
	}
	return path
}

// newTestReader returns a SaveGameReader backed by the standard test fixture.
func newTestReader(t *testing.T) store.SaveGameReader {
	t.Helper()
	saveDB := testutil.NewTestSaveGameDB(t)
	return store.NewSqliteSaveGameReader(saveDB, "")
}

// ── Happy path ───────────────────────────────────────────────────────────────

func TestSyncSeason_TakesSnapshotOnFirstSync(t *testing.T) {
	svc, snapshotStore, snapshotDir := newTestSyncService(t)
	companionDB := testutil.NewTestDB(t)
	reader := newTestReader(t)
	saveFilePath := writeSaveFile(t, []byte("save game bytes v1"))

	result, err := svc.SyncSeason(context.Background(), companionDB, reader, saveFilePath, leagueGUIDFixture)
	if err != nil {
		t.Fatalf("SyncSeason: %v", err)
	}

	// Import counts must be populated — proves import ran.
	if result.Players == 0 {
		t.Error("expected Players > 0")
	}
	if result.Teams == 0 {
		t.Error("expected Teams > 0")
	}

	// Snapshot must be flagged as new.
	if !result.IsNewSnapshot {
		t.Error("expected IsNewSnapshot=true on first sync")
	}
	if result.SnapshotID == 0 {
		t.Error("expected non-zero SnapshotID")
	}

	// Snapshot record must exist in the companion DB.
	snaps, err := snapshotStore.List(context.Background())
	if err != nil {
		t.Fatalf("listing snapshots: %v", err)
	}
	if len(snaps) != 1 {
		t.Fatalf("expected 1 snapshot record, got %d", len(snaps))
	}

	// Snapshot file must exist on disk.
	snapPath := filepath.Join(snapshotDir, string(snaps[0].FileName))
	if _, err := os.Stat(snapPath); os.IsNotExist(err) {
		t.Errorf("snapshot file not found on disk: %s", snapPath)
	}
}

// ── Snapshot is taken before import ─────────────────────────────────────────

func TestSyncSeason_SnapshotFailureBlocksImport(t *testing.T) {
	svc, snapshotStore, _ := newTestSyncService(t)
	companionDB := testutil.NewTestDB(t)
	reader := newTestReader(t)

	// Pass a path that does not exist — TakeSnapshotFromFile will fail to open it.
	_, err := svc.SyncSeason(context.Background(), companionDB, reader, "/nonexistent/save.sqlite", leagueGUIDFixture)
	if err == nil {
		t.Fatal("expected error when snapshot cannot be taken, got nil")
	}

	// No snapshot record should exist.
	snaps, listErr := snapshotStore.List(context.Background())
	if listErr != nil {
		t.Fatalf("listing snapshots: %v", listErr)
	}
	if len(snaps) != 0 {
		t.Errorf("expected 0 snapshot records after failure, got %d", len(snaps))
	}

	// No season data should have been imported.
	var seasonCount int
	_ = companionDB.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM seasons`).Scan(&seasonCount)
	if seasonCount != 0 {
		t.Errorf("expected 0 imported seasons after snapshot failure, got %d", seasonCount)
	}
}

// ── Deduplication ────────────────────────────────────────────────────────────

func TestSyncSeason_DeduplicatesIdenticalSave(t *testing.T) {
	svc, snapshotStore, snapshotDir := newTestSyncService(t)
	companionDB := testutil.NewTestDB(t)
	saveFilePath := writeSaveFile(t, []byte("identical save game bytes"))

	ctx := context.Background()

	// First sync.
	r1 := newTestReader(t)
	result1, err := svc.SyncSeason(ctx, companionDB, r1, saveFilePath, leagueGUIDFixture)
	if err != nil {
		t.Fatalf("first SyncSeason: %v", err)
	}
	if !result1.IsNewSnapshot {
		t.Error("first sync: expected IsNewSnapshot=true")
	}

	// Second sync with the same file content.
	r2 := newTestReader(t)
	result2, err := svc.SyncSeason(ctx, companionDB, r2, saveFilePath, leagueGUIDFixture)
	if err != nil {
		t.Fatalf("second SyncSeason: %v", err)
	}
	if result2.IsNewSnapshot {
		t.Error("second sync with identical file: expected IsNewSnapshot=false")
	}

	// Exactly one snapshot record and one file on disk — no duplicate written.
	snaps, err := snapshotStore.List(ctx)
	if err != nil {
		t.Fatalf("listing snapshots: %v", err)
	}
	if len(snaps) != 1 {
		t.Errorf("expected 1 snapshot record after two identical syncs, got %d", len(snaps))
	}
	entries, _ := os.ReadDir(snapshotDir)
	if len(entries) != 1 {
		t.Errorf("expected 1 snapshot file on disk after two identical syncs, got %d", len(entries))
	}
}

func TestSyncSeason_NewSnapshotWhenSaveChanges(t *testing.T) {
	svc, snapshotStore, snapshotDir := newTestSyncService(t)
	companionDB := testutil.NewTestDB(t)
	ctx := context.Background()

	// First sync.
	save1 := writeSaveFile(t, []byte("save game version 1"))
	r1 := newTestReader(t)
	if _, err := svc.SyncSeason(ctx, companionDB, r1, save1, leagueGUIDFixture); err != nil {
		t.Fatalf("first SyncSeason: %v", err)
	}

	// Second sync with different file content (save game changed between syncs).
	save2 := writeSaveFile(t, []byte("save game version 2 — different content"))
	r2 := newTestReader(t)
	result2, err := svc.SyncSeason(ctx, companionDB, r2, save2, leagueGUIDFixture)
	if err != nil {
		t.Fatalf("second SyncSeason: %v", err)
	}
	if !result2.IsNewSnapshot {
		t.Error("second sync with changed file: expected IsNewSnapshot=true")
	}

	// Two snapshot records and two files on disk.
	snaps, err := snapshotStore.List(ctx)
	if err != nil {
		t.Fatalf("listing snapshots: %v", err)
	}
	if len(snaps) != 2 {
		t.Errorf("expected 2 snapshot records after two different syncs, got %d", len(snaps))
	}
	entries, _ := os.ReadDir(snapshotDir)
	if len(entries) != 2 {
		t.Errorf("expected 2 snapshot files on disk after two different syncs, got %d", len(entries))
	}
}
