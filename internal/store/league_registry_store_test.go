package store_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

func TestLeagueRegistryStore_LeagueExists(t *testing.T) {
	db := testutil.NewTestMasterSaveDB(t)
	s := store.NewLeagueRegistryStore(db)
	ctx := context.Background()

	existing := uuid.MustParse("99F30082-775B-4547-ADD8-8C7D2C94FCE5")
	exists, err := s.LeagueExists(ctx, existing)
	if err != nil {
		t.Fatalf("LeagueExists: %v", err)
	}
	if !exists {
		t.Error("expected the seeded league to exist, got false")
	}

	notRegistered := uuid.New()
	exists, err = s.LeagueExists(ctx, notRegistered)
	if err != nil {
		t.Fatalf("LeagueExists: %v", err)
	}
	if exists {
		t.Error("expected a freshly generated GUID to not exist, got true")
	}
}

func TestLeagueRegistryStore_RegisterLeague(t *testing.T) {
	db := testutil.NewTestMasterSaveDB(t)
	s := store.NewLeagueRegistryStore(db)
	ctx := context.Background()

	newGUID := uuid.New()
	if err := s.RegisterLeague(ctx, newGUID); err != nil {
		t.Fatalf("RegisterLeague: %v", err)
	}

	exists, err := s.LeagueExists(ctx, newGUID)
	if err != nil {
		t.Fatalf("LeagueExists after RegisterLeague: %v", err)
	}
	if !exists {
		t.Fatal("expected newly registered league to exist")
	}

	// Regression test for failure-analysis.md Bug #1: the stored value must
	// be a 16-byte blob, never a 36-character string.
	var (
		guidType string
		guidLen  int
	)
	err = db.QueryRowContext(ctx, `
		SELECT typeof(GUID), length(GUID) FROM t_league_savedatas WHERE GUID = ?
	`, newGUID[:]).Scan(&guidType, &guidLen)
	if err != nil {
		t.Fatalf("querying stored GUID shape: %v", err)
	}
	if guidType != "blob" {
		t.Errorf("stored GUID type = %q, want %q (legacy bug stored it as text)", guidType, "blob")
	}
	if guidLen != 16 {
		t.Errorf("stored GUID length = %d bytes, want 16", guidLen)
	}
}

func TestLeagueRegistryStore_RegisterLeague_DuplicateFails(t *testing.T) {
	db := testutil.NewTestMasterSaveDB(t)
	s := store.NewLeagueRegistryStore(db)
	ctx := context.Background()

	existing := uuid.MustParse("99F30082-775B-4547-ADD8-8C7D2C94FCE5")
	err := s.RegisterLeague(ctx, existing)
	if err == nil {
		t.Fatal("expected an error registering an already-existing GUID (primary key collision), got nil")
	}
}
