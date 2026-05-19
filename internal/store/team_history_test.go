package store_test

import (
	"context"
	"testing"

	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

func TestTeamHistoryStore_UpsertTeam_CreatesAndReuses(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := store.NewTeamHistoryStore(db)
	ctx := context.Background()

	id1, err := s.UpsertTeam(ctx, "AABB")
	if err != nil {
		t.Fatalf("first UpsertTeam: %v", err)
	}
	id2, err := s.UpsertTeam(ctx, "AABB")
	if err != nil {
		t.Fatalf("second UpsertTeam: %v", err)
	}
	if id1 != id2 {
		t.Errorf("expected same ID on re-upsert, got %d and %d", id1, id2)
	}
}

func TestTeamHistoryStore_UpsertSeasonHistory(t *testing.T) {
	db := testutil.NewTestDB(t)
	ts := store.NewTeamHistoryStore(db)
	ss := store.NewSeasonStore(db)
	ctx := context.Background()

	_ = ss.Upsert(ctx, store.Season{ID: 100, SeasonNum: 1})
	teamID, _ := ts.UpsertTeam(ctx, "GUID1")

	h := store.TeamSeasonHistory{
		TeamID: teamID, SeasonID: 100,
		TeamName: "Home Squad", DivisionName: "East",
		Wins: 30, Losses: 20, RunsFor: 200, RunsAgainst: 170,
	}
	histID, err := ts.UpsertSeasonHistory(ctx, h)
	if err != nil {
		t.Fatalf("UpsertSeasonHistory: %v", err)
	}
	if histID == 0 {
		t.Error("expected non-zero history ID")
	}

	// Re-upsert with different wins — should update
	h.Wins = 35
	newHistID, err := ts.UpsertSeasonHistory(ctx, h)
	if err != nil {
		t.Fatalf("re-upsert: %v", err)
	}
	if newHistID != histID {
		t.Errorf("expected same history ID on re-upsert (%d), got %d", histID, newHistID)
	}
}

func TestTeamHistoryStore_TwoTeamsTwoSeasons(t *testing.T) {
	db := testutil.NewTestDB(t)
	ts := store.NewTeamHistoryStore(db)
	ss := store.NewSeasonStore(db)
	ctx := context.Background()

	_ = ss.Upsert(ctx, store.Season{ID: 1, SeasonNum: 1})
	_ = ss.Upsert(ctx, store.Season{ID: 2, SeasonNum: 2})

	t1, _ := ts.UpsertTeam(ctx, "TEAM1")
	t2, _ := ts.UpsertTeam(ctx, "TEAM2")

	for _, tc := range []struct{ teamID int64; seasonID int }{
		{t1, 1}, {t1, 2}, {t2, 1}, {t2, 2},
	} {
		h := store.TeamSeasonHistory{TeamID: tc.teamID, SeasonID: tc.seasonID, TeamName: "X"}
		if _, err := ts.UpsertSeasonHistory(ctx, h); err != nil {
			t.Errorf("UpsertSeasonHistory(team=%d season=%d): %v", tc.teamID, tc.seasonID, err)
		}
	}
}
