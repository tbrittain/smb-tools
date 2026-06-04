package store_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"smb-tools/internal/models"
	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

// setupMediaDB seeds a season, team + team_season_history row, and player.
// Returns the test DB, historyID, and playerID.
func setupMediaDB(t *testing.T) (*sql.DB, int64, int64) {
	t.Helper()
	database := testutil.NewTestDB(t)
	ctx := context.Background()

	ss := store.NewSeasonStore(database)
	seasonID := upsertTestSeason(t, ss, "LEAGUE1", 1, 1)

	ts := store.NewTeamHistoryStore(database)
	teamID, err := ts.UpsertTeam(ctx, "GUID1", "Test Team")
	if err != nil {
		t.Fatalf("UpsertTeam: %v", err)
	}
	histID, err := ts.UpsertSeasonHistory(ctx, store.TeamSeasonHistory{
		TeamID: teamID, SeasonID: seasonID, TeamName: "Test Team",
	})
	if err != nil {
		t.Fatalf("UpsertSeasonHistory: %v", err)
	}

	ps := store.NewPlayerSeasonStore(database)
	playerID, err := ps.UpsertPlayer(ctx, store.PlayerIdentity{
		GameGUID: "P1", FirstName: "John", LastName: "Doe",
	})
	if err != nil {
		t.Fatalf("UpsertPlayer: %v", err)
	}

	return database, histID, playerID
}

func insertTestMediaRow(t *testing.T, s *store.MediaStore, db *sql.DB, id, name string) models.Media {
	t.Helper()
	m := models.Media{
		ID:         id,
		FilePath:   "assets/media/" + id + ".png",
		MediaType:  models.MediaTypeImage,
		Name:       name,
		UploadedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := s.InsertMedia(context.Background(), db, m); err != nil {
		t.Fatalf("InsertMedia(%s): %v", name, err)
	}
	return m
}

// ── GetMediaForTeamSeason ─────────────────────────────────────────────────────

func TestMediaStore_GetMediaForTeamSeason_HappyPath(t *testing.T) {
	database, histID, _ := setupMediaDB(t)
	s := store.NewMediaStore()
	ctx := context.Background()

	m := insertTestMediaRow(t, s, database, "media-1", "Screenshot 1")
	if err := s.InsertTeamSeasonAssoc(ctx, database, models.MediaTeamSeason{
		ID: "assoc-1", MediaID: m.ID, TeamHistoryID: histID,
	}); err != nil {
		t.Fatalf("InsertTeamSeasonAssoc: %v", err)
	}

	items, total, err := s.GetMediaForTeamSeason(ctx, database, histID, 24, 0)
	if err != nil {
		t.Fatalf("GetMediaForTeamSeason: %v", err)
	}
	if total != 1 {
		t.Errorf("total: want 1, got %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("items len: want 1, got %d", len(items))
	}
	if items[0].Media.ID != m.ID {
		t.Errorf("media ID: want %s, got %s", m.ID, items[0].Media.ID)
	}
	if len(items[0].TeamSeasons) != 1 {
		t.Errorf("team-season assocs: want 1, got %d", len(items[0].TeamSeasons))
	}
	if items[0].TeamSeasons[0].TeamName != "Test Team" {
		t.Errorf("team name in assoc: want 'Test Team', got %q", items[0].TeamSeasons[0].TeamName)
	}
}

// ── GetMediaForPlayer ─────────────────────────────────────────────────────────

func TestMediaStore_GetMediaForPlayer_HappyPath(t *testing.T) {
	database, _, playerID := setupMediaDB(t)
	s := store.NewMediaStore()
	ctx := context.Background()

	m := insertTestMediaRow(t, s, database, "player-clip", "Player Clip")
	if err := s.InsertPlayerAssoc(ctx, database, models.MediaPlayer{
		ID: "passoc-1", MediaID: m.ID, PlayerID: playerID,
	}); err != nil {
		t.Fatalf("InsertPlayerAssoc: %v", err)
	}

	items, total, err := s.GetMediaForPlayer(ctx, database, playerID, 24, 0)
	if err != nil {
		t.Fatalf("GetMediaForPlayer: %v", err)
	}
	if total != 1 {
		t.Errorf("total: want 1, got %d", total)
	}
	if len(items[0].Players) != 1 {
		t.Errorf("player assocs: want 1, got %d", len(items[0].Players))
	}
	if items[0].Players[0].LastName != "Doe" {
		t.Errorf("player last name: want 'Doe', got %q", items[0].Players[0].LastName)
	}
}

// ── Pagination ────────────────────────────────────────────────────────────────

func TestMediaStore_GetMediaForTeamSeason_Pagination(t *testing.T) {
	database, histID, _ := setupMediaDB(t)
	s := store.NewMediaStore()
	ctx := context.Background()

	for i := range 5 {
		id := string(rune('a' + i))
		m := insertTestMediaRow(t, s, database, "item-"+id, "Item "+id)
		if err := s.InsertTeamSeasonAssoc(ctx, database, models.MediaTeamSeason{
			ID: "assoc-" + id, MediaID: m.ID, TeamHistoryID: histID,
		}); err != nil {
			t.Fatalf("InsertTeamSeasonAssoc: %v", err)
		}
	}

	items0, total, err := s.GetMediaForTeamSeason(ctx, database, histID, 3, 0)
	if err != nil {
		t.Fatalf("GetMediaForTeamSeason page 0: %v", err)
	}
	if total != 5 {
		t.Errorf("total: want 5, got %d", total)
	}
	if len(items0) != 3 {
		t.Errorf("page 0 len: want 3, got %d", len(items0))
	}

	items1, total1, err := s.GetMediaForTeamSeason(ctx, database, histID, 3, 3)
	if err != nil {
		t.Fatalf("GetMediaForTeamSeason page 1: %v", err)
	}
	if total1 != 5 {
		t.Errorf("page 1 total: want 5, got %d", total1)
	}
	if len(items1) != 2 {
		t.Errorf("page 1 len: want 2, got %d", len(items1))
	}
}

// ── TotalCount reflects all items regardless of page ─────────────────────────

func TestMediaStore_TotalCountIsConsistentAcrossPages(t *testing.T) {
	database, histID, _ := setupMediaDB(t)
	s := store.NewMediaStore()
	ctx := context.Background()

	for i := range 7 {
		id := string(rune('a' + i))
		m := insertTestMediaRow(t, s, database, "cnt-"+id, "Count "+id)
		if err := s.InsertTeamSeasonAssoc(ctx, database, models.MediaTeamSeason{
			ID: "ca-" + id, MediaID: m.ID, TeamHistoryID: histID,
		}); err != nil {
			t.Fatalf("InsertTeamSeasonAssoc: %v", err)
		}
	}

	_, total, err := s.GetMediaForTeamSeason(ctx, database, histID, 100, 0)
	if err != nil {
		t.Fatalf("GetMediaForTeamSeason: %v", err)
	}
	if total != 7 {
		t.Errorf("total: want 7, got %d", total)
	}
}

// ── RemoveTeamSeasonAssoc — partial removal ───────────────────────────────────

func TestMediaStore_RemoveTeamSeasonAssoc_PartialRemoval(t *testing.T) {
	database, histID, playerID := setupMediaDB(t)
	s := store.NewMediaStore()
	ctx := context.Background()

	m := insertTestMediaRow(t, s, database, "shared-1", "Shared")
	if err := s.InsertTeamSeasonAssoc(ctx, database, models.MediaTeamSeason{
		ID: "ts-shared", MediaID: m.ID, TeamHistoryID: histID,
	}); err != nil {
		t.Fatalf("InsertTeamSeasonAssoc: %v", err)
	}
	if err := s.InsertPlayerAssoc(ctx, database, models.MediaPlayer{
		ID: "p-shared", MediaID: m.ID, PlayerID: playerID,
	}); err != nil {
		t.Fatalf("InsertPlayerAssoc: %v", err)
	}

	if err := s.RemoveTeamSeasonAssoc(ctx, database, m.ID, histID); err != nil {
		t.Fatalf("RemoveTeamSeasonAssoc: %v", err)
	}

	tsCount, pCount, err := s.GetAssociationCounts(ctx, database, m.ID)
	if err != nil {
		t.Fatalf("GetAssociationCounts: %v", err)
	}
	if tsCount != 0 {
		t.Errorf("team-season count after partial remove: want 0, got %d", tsCount)
	}
	if pCount != 1 {
		t.Errorf("player count after partial remove: want 1, got %d", pCount)
	}
}

// ── RemoveTeamSeasonAssoc — last association ──────────────────────────────────

func TestMediaStore_RemoveTeamSeasonAssoc_LastAssoc_CountsZero(t *testing.T) {
	database, histID, _ := setupMediaDB(t)
	s := store.NewMediaStore()
	ctx := context.Background()

	m := insertTestMediaRow(t, s, database, "solo-1", "Solo")
	if err := s.InsertTeamSeasonAssoc(ctx, database, models.MediaTeamSeason{
		ID: "ts-solo", MediaID: m.ID, TeamHistoryID: histID,
	}); err != nil {
		t.Fatalf("InsertTeamSeasonAssoc: %v", err)
	}

	if err := s.RemoveTeamSeasonAssoc(ctx, database, m.ID, histID); err != nil {
		t.Fatalf("RemoveTeamSeasonAssoc: %v", err)
	}

	tsCount, pCount, err := s.GetAssociationCounts(ctx, database, m.ID)
	if err != nil {
		t.Fatalf("GetAssociationCounts: %v", err)
	}
	if tsCount+pCount != 0 {
		t.Errorf("expected 0 total assocs, got ts=%d p=%d", tsCount, pCount)
	}
}

// ── DeleteMedia ───────────────────────────────────────────────────────────────

func TestMediaStore_DeleteMedia_RemovesAllRows(t *testing.T) {
	database, histID, playerID := setupMediaDB(t)
	s := store.NewMediaStore()
	ctx := context.Background()

	m := insertTestMediaRow(t, s, database, "del-1", "Delete Me")
	if err := s.InsertTeamSeasonAssoc(ctx, database, models.MediaTeamSeason{
		ID: "d-ts", MediaID: m.ID, TeamHistoryID: histID,
	}); err != nil {
		t.Fatalf("InsertTeamSeasonAssoc: %v", err)
	}
	if err := s.InsertPlayerAssoc(ctx, database, models.MediaPlayer{
		ID: "d-p", MediaID: m.ID, PlayerID: playerID,
	}); err != nil {
		t.Fatalf("InsertPlayerAssoc: %v", err)
	}

	if err := s.DeleteMedia(ctx, database, m.ID); err != nil {
		t.Fatalf("DeleteMedia: %v", err)
	}

	_, err := s.GetMediaByID(ctx, database, m.ID)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected not-found error after delete, got: %v", err)
	}

	tsCount, pCount, err := s.GetAssociationCounts(ctx, database, m.ID)
	if err != nil {
		t.Fatalf("GetAssociationCounts after delete: %v", err)
	}
	if tsCount+pCount != 0 {
		t.Errorf("expected 0 assocs after delete, got ts=%d p=%d", tsCount, pCount)
	}
}

// ── GetAssociationCounts ──────────────────────────────────────────────────────

func TestMediaStore_GetAssociationCounts(t *testing.T) {
	database, histID, playerID := setupMediaDB(t)
	s := store.NewMediaStore()
	ctx := context.Background()

	m := insertTestMediaRow(t, s, database, "counted-1", "Counted")
	if err := s.InsertTeamSeasonAssoc(ctx, database, models.MediaTeamSeason{
		ID: "cnt-ts", MediaID: m.ID, TeamHistoryID: histID,
	}); err != nil {
		t.Fatalf("InsertTeamSeasonAssoc: %v", err)
	}
	if err := s.InsertPlayerAssoc(ctx, database, models.MediaPlayer{
		ID: "cnt-p", MediaID: m.ID, PlayerID: playerID,
	}); err != nil {
		t.Fatalf("InsertPlayerAssoc: %v", err)
	}

	tsCount, pCount, err := s.GetAssociationCounts(ctx, database, m.ID)
	if err != nil {
		t.Fatalf("GetAssociationCounts: %v", err)
	}
	if tsCount != 1 {
		t.Errorf("team-season count: want 1, got %d", tsCount)
	}
	if pCount != 1 {
		t.Errorf("player count: want 1, got %d", pCount)
	}
}

// ── UNIQUE constraint ─────────────────────────────────────────────────────────

func TestMediaStore_UniqueConstraint_TeamSeason(t *testing.T) {
	database, histID, _ := setupMediaDB(t)
	s := store.NewMediaStore()
	ctx := context.Background()

	m := insertTestMediaRow(t, s, database, "uniq-1", "Unique Test")
	if err := s.InsertTeamSeasonAssoc(ctx, database, models.MediaTeamSeason{
		ID: "uniq-a", MediaID: m.ID, TeamHistoryID: histID,
	}); err != nil {
		t.Fatalf("first InsertTeamSeasonAssoc: %v", err)
	}

	// Duplicate (same media_id + team_history_id) must fail.
	err := s.InsertTeamSeasonAssoc(ctx, database, models.MediaTeamSeason{
		ID: "uniq-b", MediaID: m.ID, TeamHistoryID: histID,
	})
	if err == nil {
		t.Error("expected UNIQUE constraint violation, got nil")
	}
}

// ── SearchTeamsForPicker ──────────────────────────────────────────────────────

func TestMediaStore_SearchTeamsForPicker(t *testing.T) {
	database, _, _ := setupMediaDB(t)
	s := store.NewMediaStore()
	ctx := context.Background()

	results, err := s.SearchTeamsForPicker(ctx, database, "Test")
	if err != nil {
		t.Fatalf("SearchTeamsForPicker: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 match for 'Test', got %d", len(results))
	}
	if len(results) > 0 && results[0].TeamName != "Test Team" {
		t.Errorf("team name: want 'Test Team', got %q", results[0].TeamName)
	}

	none, err := s.SearchTeamsForPicker(ctx, database, "ZZZNoMatch")
	if err != nil {
		t.Fatalf("SearchTeamsForPicker (no match): %v", err)
	}
	if len(none) != 0 {
		t.Errorf("expected 0 results for non-matching query, got %d", len(none))
	}
}

// ── GetTeamSeasonsForPicker ───────────────────────────────────────────────────

func TestMediaStore_GetTeamSeasonsForPicker(t *testing.T) {
	database, _, _ := setupMediaDB(t)
	s := store.NewMediaStore()
	ctx := context.Background()

	teams, err := s.SearchTeamsForPicker(ctx, database, "Test")
	if err != nil || len(teams) == 0 {
		t.Fatalf("SearchTeamsForPicker: %v (results=%d)", err, len(teams))
	}

	seasons, err := s.GetTeamSeasonsForPicker(ctx, database, teams[0].TeamID)
	if err != nil {
		t.Fatalf("GetTeamSeasonsForPicker: %v", err)
	}
	if len(seasons) != 1 {
		t.Errorf("expected 1 season, got %d", len(seasons))
	}
	if len(seasons) > 0 && seasons[0].SeasonNum != 1 {
		t.Errorf("season num: want 1, got %d", seasons[0].SeasonNum)
	}
}
