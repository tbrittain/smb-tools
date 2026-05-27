package store_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

func seedSnapshot(t *testing.T, s *store.SnapshotStore, seasonNum int) int64 {
	t.Helper()
	snap := store.Snapshot{
		SeasonNum:     seasonNum,
		CapturedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		FileName:      store.SnapshotFileName("0001_abc123.sqlite"),
		SHA256Hash:    store.SHA256Hex("aabbccdd"),
		FileSizeBytes: 1024,
	}
	id, err := s.Record(context.Background(), snap)
	if err != nil {
		t.Fatalf("seedSnapshot: %v", err)
	}
	return id
}

func TestSnapshotStore_GetByID(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := store.NewSnapshotStore(db)
	ctx := context.Background()

	id := seedSnapshot(t, s, 3)

	got, err := s.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ID != id {
		t.Errorf("ID: got %d, want %d", got.ID, id)
	}
	if got.SeasonNum != 3 {
		t.Errorf("SeasonNum: got %d, want 3", got.SeasonNum)
	}
	if got.FileSizeBytes != 1024 {
		t.Errorf("FileSizeBytes: got %d, want 1024", got.FileSizeBytes)
	}
}

func TestSnapshotStore_GetByID_NotFound(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := store.NewSnapshotStore(db)

	_, err := s.GetByID(context.Background(), 9999)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}
