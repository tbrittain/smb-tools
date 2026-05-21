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

	id1, err := s.UpsertTeam(ctx, "AABB", "Team Alpha")
	if err != nil {
		t.Fatalf("first UpsertTeam: %v", err)
	}
	id2, err := s.UpsertTeam(ctx, "AABB", "Team Alpha")
	if err != nil {
		t.Fatalf("second UpsertTeam: %v", err)
	}
	if id1 != id2 {
		t.Errorf("expected same ID on re-upsert, got %d and %d", id1, id2)
	}
}

func TestTeamHistoryStore_UpsertTeam_NameMatchFork(t *testing.T) {
	db := testutil.NewTestDB(t)
	ts := store.NewTeamHistoryStore(db)
	ss := store.NewSeasonStore(db)
	ctx := context.Background()

	// Create a team with GUID1 and a season history row (needed for name match)
	sID := upsertTestSeason(t, ss, "LEAGUE1", 1, 1)
	id1, _ := ts.UpsertTeam(ctx, "GUID1", "Home Squad")
	_, _ = ts.UpsertSeasonHistory(ctx, store.TeamSeasonHistory{
		TeamID: id1, SeasonID: sID, TeamName: "Home Squad",
	})

	// Fork: same team name, new GUID — should resolve via name match
	id2, err := ts.UpsertTeam(ctx, "GUID2", "Home Squad")
	if err != nil {
		t.Fatalf("fork UpsertTeam: %v", err)
	}
	if id1 != id2 {
		t.Errorf("expected name-match to return same team ID (%d), got %d", id1, id2)
	}

	// Next call with GUID2 should hit alt_guids (tier 2), not name match
	id3, err := ts.UpsertTeam(ctx, "GUID2", "Home Squad")
	if err != nil {
		t.Fatalf("alt GUID UpsertTeam: %v", err)
	}
	if id1 != id3 {
		t.Errorf("expected alt GUID lookup to return same team ID (%d), got %d", id1, id3)
	}
}

func TestTeamHistoryStore_UpsertSeasonHistory(t *testing.T) {
	db := testutil.NewTestDB(t)
	ts := store.NewTeamHistoryStore(db)
	ss := store.NewSeasonStore(db)
	ctx := context.Background()

	sID := upsertTestSeason(t, ss, "LEAGUE1", 100, 1)
	teamID, _ := ts.UpsertTeam(ctx, "GUID1", "Home Squad")

	h := store.TeamSeasonHistory{
		TeamID: teamID, SeasonID: sID,
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

	s1ID := upsertTestSeason(t, ss, "LEAGUE1", 1, 1)
	s2ID := upsertTestSeason(t, ss, "LEAGUE1", 2, 2)

	t1, _ := ts.UpsertTeam(ctx, "TEAM1", "Alpha")
	t2, _ := ts.UpsertTeam(ctx, "TEAM2", "Beta")

	for _, tc := range []struct {
		teamID   int64
		seasonID int64
	}{
		{t1, s1ID}, {t1, s2ID}, {t2, s1ID}, {t2, s2ID},
	} {
		h := store.TeamSeasonHistory{TeamID: tc.teamID, SeasonID: tc.seasonID, TeamName: "X"}
		if _, err := ts.UpsertSeasonHistory(ctx, h); err != nil {
			t.Errorf("UpsertSeasonHistory(team=%d season=%d): %v", tc.teamID, tc.seasonID, err)
		}
	}
}
