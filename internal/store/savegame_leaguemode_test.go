package store_test

import (
	"context"
	"testing"

	"smb-tools/internal/models"
	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

const (
	franchiseLeagueGUIDHex = "EE000000000000000000000000000000"
	seasonLeagueGUIDHex    = "FE000000000000000000000000000000"
)

func TestSqliteSaveGameReader_GetLeagues_FranchiseMode(t *testing.T) {
	db := testutil.NewTestSaveGameDB(t)
	r := store.NewSqliteSaveGameReader(db, "")

	leagues, err := r.GetLeagues(context.Background())
	if err != nil {
		t.Fatalf("GetLeagues: %v", err)
	}
	if len(leagues) != 1 {
		t.Fatalf("expected 1 league, got %d", len(leagues))
	}
	if leagues[0].Mode != models.LeagueModeFranchise {
		t.Errorf("mode: got %q, want %q", leagues[0].Mode, models.LeagueModeFranchise)
	}
	if leagues[0].NumSeasons != 2 {
		t.Errorf("numSeasons: got %d, want 2", leagues[0].NumSeasons)
	}
}

// TestSqliteSaveGameReader_GetLeagues_SeasonMode proves mode detection works
// for a league with no t_franchise row and elimination=0 — the real save-game
// signature of Season Mode. This is the core regression test for the season
// mode feature: GetLeagues must resolve LeagueModeSeason without any franchise
// row present.
func TestSqliteSaveGameReader_GetLeagues_SeasonMode(t *testing.T) {
	db := testutil.NewTestSaveGameDB_SeasonMode(t)
	r := store.NewSqliteSaveGameReader(db, "")

	leagues, err := r.GetLeagues(context.Background())
	if err != nil {
		t.Fatalf("GetLeagues: %v", err)
	}
	if len(leagues) != 1 {
		t.Fatalf("expected 1 league, got %d", len(leagues))
	}
	if leagues[0].Mode != models.LeagueModeSeason {
		t.Errorf("mode: got %q, want %q", leagues[0].Mode, models.LeagueModeSeason)
	}
	if leagues[0].NumSeasons != 2 {
		t.Errorf("numSeasons: got %d, want 2", leagues[0].NumSeasons)
	}
	if leagues[0].PlayerTeamName != "" {
		t.Errorf("playerTeamName: expected empty for season mode (no franchise-controlled team), got %q", leagues[0].PlayerTeamName)
	}
}

// TestSqliteSaveGameReader_GetCurrentSeason_SeasonMode proves season
// progression doesn't depend on t_franchise — GetCurrentSeason must resolve
// the latest season for a season-mode league exactly like franchise mode.
func TestSqliteSaveGameReader_GetCurrentSeason_SeasonMode(t *testing.T) {
	db := testutil.NewTestSaveGameDB_SeasonMode(t)
	r := store.NewSqliteSaveGameReader(db, "")

	info, err := r.GetCurrentSeason(context.Background(), seasonLeagueGUIDHex)
	if err != nil {
		t.Fatalf("GetCurrentSeason: %v", err)
	}
	if info.SeasonID != 201 {
		t.Errorf("seasonID: got %d, want 201", info.SeasonID)
	}
	if info.SeasonNum != 2 {
		t.Errorf("seasonNum: got %d, want 2", info.SeasonNum)
	}
}

// TestSqliteSaveGameReader_GetFranchiseSeasons_SeasonMode proves multi-season
// enumeration works for a season-mode league with no t_franchise row.
func TestSqliteSaveGameReader_GetFranchiseSeasons_SeasonMode(t *testing.T) {
	db := testutil.NewTestSaveGameDB_SeasonMode(t)
	r := store.NewSqliteSaveGameReader(db, "")

	seasons, err := r.GetFranchiseSeasons(context.Background(), seasonLeagueGUIDHex)
	if err != nil {
		t.Fatalf("GetFranchiseSeasons: %v", err)
	}
	if len(seasons) != 2 {
		t.Fatalf("expected 2 seasons, got %d", len(seasons))
	}
	if seasons[0].SeasonID != 200 || seasons[0].SeasonNum != 1 {
		t.Errorf("season[0]: got id=%d num=%d, want id=200 num=1", seasons[0].SeasonID, seasons[0].SeasonNum)
	}
	if seasons[1].SeasonID != 201 || seasons[1].SeasonNum != 2 {
		t.Errorf("season[1]: got id=%d num=%d, want id=201 num=2", seasons[1].SeasonID, seasons[1].SeasonNum)
	}
}
