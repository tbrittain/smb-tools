package service_test

import (
	"context"
	"testing"

	"smb-tools/internal/service"
	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

func TestTakeSnapshot_NewSnapshot(t *testing.T) {
	db := testutil.NewTestDB(t)
	snapshotDir := t.TempDir()
	svc := service.NewSnapshotService(snapshotDir, store.NewSnapshotStore(db))

	data := []byte("fake sqlite content for season 1")
	id, isNew, err := svc.TakeSnapshot(context.Background(), data, 1)
	if err != nil {
		t.Fatalf("TakeSnapshot: %v", err)
	}
	if !isNew {
		t.Error("expected isNew=true for first snapshot")
	}
	if id == 0 {
		t.Error("expected non-zero snapshot ID")
	}
}

func TestTakeSnapshot_Idempotent(t *testing.T) {
	db := testutil.NewTestDB(t)
	snapshotDir := t.TempDir()
	svc := service.NewSnapshotService(snapshotDir, store.NewSnapshotStore(db))

	data := []byte("same content both times")
	_, _, err := svc.TakeSnapshot(context.Background(), data, 1)
	if err != nil {
		t.Fatalf("first TakeSnapshot: %v", err)
	}

	id2, isNew, err := svc.TakeSnapshot(context.Background(), data, 1)
	if err != nil {
		t.Fatalf("second TakeSnapshot: %v", err)
	}
	if isNew {
		t.Error("expected isNew=false for duplicate content")
	}
	if id2 != 0 {
		t.Errorf("expected id=0 for deduplicated snapshot, got %d", id2)
	}
}

func TestTakeSnapshot_DifferentContent(t *testing.T) {
	db := testutil.NewTestDB(t)
	snapshotDir := t.TempDir()
	svc := service.NewSnapshotService(snapshotDir, store.NewSnapshotStore(db))

	_, _, _ = svc.TakeSnapshot(context.Background(), []byte("season 1 data"), 1)
	id2, isNew, err := svc.TakeSnapshot(context.Background(), []byte("season 2 data"), 2)
	if err != nil {
		t.Fatalf("second TakeSnapshot: %v", err)
	}
	if !isNew {
		t.Error("expected isNew=true for different content")
	}
	if id2 == 0 {
		t.Error("expected non-zero snapshot ID for new content")
	}
}
