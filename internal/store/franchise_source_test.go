package store_test

import (
	"context"
	"database/sql"
	"testing"

	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

func TestFranchiseSourceStore_AddAndGetActive(t *testing.T) {
	db := testutil.NewTestRegistryDB(t)
	fs := store.NewFranchiseStore(db)
	ss := store.NewFranchiseSourceStore(db)
	ctx := context.Background()

	_ = fs.Create(ctx, franchiseFixture("fid1"))

	src, err := ss.Add(ctx, "fid1", "/path/to/save.sav", "LEAGUE-GUID-1", 0)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if src.ID == 0 {
		t.Error("expected non-zero source ID")
	}

	active, err := ss.GetActive(ctx, "fid1")
	if err != nil {
		t.Fatalf("GetActive: %v", err)
	}
	if active.LeagueGUID != "LEAGUE-GUID-1" {
		t.Errorf("LeagueGUID: got %q, want %q", active.LeagueGUID, "LEAGUE-GUID-1")
	}
	if active.SeasonOffset != 0 {
		t.Errorf("SeasonOffset: got %d, want 0", active.SeasonOffset)
	}
}

func TestFranchiseSourceStore_GetActiveReturnsHighestOffset(t *testing.T) {
	db := testutil.NewTestRegistryDB(t)
	fs := store.NewFranchiseStore(db)
	ss := store.NewFranchiseSourceStore(db)
	ctx := context.Background()

	_ = fs.Create(ctx, franchiseFixture("fid2"))
	_, _ = ss.Add(ctx, "fid2", "/save1.sav", "LEAGUE-1", 0)
	_, _ = ss.Add(ctx, "fid2", "/save2.sav", "LEAGUE-2", 15)

	active, err := ss.GetActive(ctx, "fid2")
	if err != nil {
		t.Fatalf("GetActive: %v", err)
	}
	if active.LeagueGUID != "LEAGUE-2" {
		t.Errorf("expected active to be LEAGUE-2 (offset=15), got %q", active.LeagueGUID)
	}
	if active.SeasonOffset != 15 {
		t.Errorf("SeasonOffset: got %d, want 15", active.SeasonOffset)
	}
}

func TestFranchiseSourceStore_GetActiveNoSources(t *testing.T) {
	db := testutil.NewTestRegistryDB(t)
	fs := store.NewFranchiseStore(db)
	ss := store.NewFranchiseSourceStore(db)
	ctx := context.Background()

	_ = fs.Create(ctx, franchiseFixture("fid3"))

	_, err := ss.GetActive(ctx, "fid3")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows for franchise with no sources, got %v", err)
	}
}

func TestFranchiseSourceStore_Replace(t *testing.T) {
	db := testutil.NewTestRegistryDB(t)
	fs := store.NewFranchiseStore(db)
	ss := store.NewFranchiseSourceStore(db)
	ctx := context.Background()

	_ = fs.Create(ctx, franchiseFixture("fid4"))
	src, _ := ss.Add(ctx, "fid4", "/old.sav", "OLD-GUID", 0)

	if err := ss.Replace(ctx, src.ID, "/new.sav", "NEW-GUID"); err != nil {
		t.Fatalf("Replace: %v", err)
	}

	active, _ := ss.GetActive(ctx, "fid4")
	if active.SaveFilePath != "/new.sav" {
		t.Errorf("SaveFilePath: got %q, want /new.sav", active.SaveFilePath)
	}
	if active.LeagueGUID != "NEW-GUID" {
		t.Errorf("LeagueGUID: got %q, want NEW-GUID", active.LeagueGUID)
	}
	if active.SeasonOffset != 0 {
		t.Errorf("SeasonOffset should be unchanged: got %d", active.SeasonOffset)
	}
}

func TestFranchiseSourceStore_ListByFranchise(t *testing.T) {
	db := testutil.NewTestRegistryDB(t)
	fs := store.NewFranchiseStore(db)
	ss := store.NewFranchiseSourceStore(db)
	ctx := context.Background()

	_ = fs.Create(ctx, franchiseFixture("fid5"))
	_, _ = ss.Add(ctx, "fid5", "/s1.sav", "L1", 0)
	_, _ = ss.Add(ctx, "fid5", "/s2.sav", "L2", 10)
	_, _ = ss.Add(ctx, "fid5", "/s3.sav", "L3", 20)

	sources, err := ss.ListByFranchise(ctx, "fid5")
	if err != nil {
		t.Fatalf("ListByFranchise: %v", err)
	}
	if len(sources) != 3 {
		t.Fatalf("expected 3 sources, got %d", len(sources))
	}
	if sources[0].SeasonOffset != 0 || sources[1].SeasonOffset != 10 || sources[2].SeasonOffset != 20 {
		t.Errorf("sources not in offset-ascending order: offsets = %d, %d, %d",
			sources[0].SeasonOffset, sources[1].SeasonOffset, sources[2].SeasonOffset)
	}
}

func TestFranchiseSourceStore_DeleteByFranchise(t *testing.T) {
	db := testutil.NewTestRegistryDB(t)
	fs := store.NewFranchiseStore(db)
	ss := store.NewFranchiseSourceStore(db)
	ctx := context.Background()

	_ = fs.Create(ctx, franchiseFixture("fid6"))
	_, _ = ss.Add(ctx, "fid6", "/s.sav", "L1", 0)

	if err := ss.DeleteByFranchise(ctx, "fid6"); err != nil {
		t.Fatalf("DeleteByFranchise: %v", err)
	}

	sources, _ := ss.ListByFranchise(ctx, "fid6")
	if len(sources) != 0 {
		t.Errorf("expected 0 sources after delete, got %d", len(sources))
	}
}
