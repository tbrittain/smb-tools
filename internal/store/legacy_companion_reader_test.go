package store_test

import (
	"context"
	"testing"

	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

// TestReadPlayerSeasonTeams_FASlotSkipped verifies that a PlayerTeamHistory row
// with NULL SeasonTeamHistoryId (free agent — no team to point at) is silently
// skipped, while valid team associations for earlier Orders are still returned.
//
// In the legacy schema, Order=1 represents the player's current/final state.
// When a player ends the season as a free agent, the companion app writes
// Order=1 with SeasonTeamHistoryId=NULL. Any team(s) they actually played for
// appear as Order=2 (and Order=3 if traded more than once).
func TestReadPlayerSeasonTeams_FASlotSkipped(t *testing.T) {
	db := testutil.NewLegacyCompanionDB(t)
	ctx := context.Background()

	// Insert a new player season (Id=100) for Alex Power (PlayerId=1) in Season 10.
	// The legacy schema has no UNIQUE(PlayerId, SeasonId) constraint, so this is valid.
	// PlayerTeamHistory:
	//   Order=1 → NULL  (ended season as free agent)
	//   Order=2 → Id=1  (Alpha Squad — the team he actually played for)
	_, err := db.ExecContext(ctx, `
		INSERT INTO PlayerSeasons (Id, PlayerId, SeasonId, Age, Salary)
		VALUES (100, 1, 10, 28, 0);
		INSERT INTO PlayerTeamHistory (PlayerSeasonId, SeasonTeamHistoryId, "Order")
		VALUES (100, NULL, 1),
		       (100, 1,    2);
	`)
	if err != nil {
		t.Fatalf("seeding FA scenario: %v", err)
	}

	reader, err := store.NewLegacyCompanionReader(ctx, db)
	if err != nil {
		t.Fatalf("NewLegacyCompanionReader: %v", err)
	}

	teamsByPS, err := reader.ReadPlayerSeasonTeams(ctx, 1)
	if err != nil {
		t.Fatalf("ReadPlayerSeasonTeams: %v", err)
	}

	teams, ok := teamsByPS[100]
	if !ok {
		t.Fatal("expected an entry for PlayerSeasonId=100; got none")
	}
	if len(teams) != 1 {
		t.Fatalf("expected 1 team for ps 100 (FA NULL slot should be skipped), got %d", len(teams))
	}
	if teams[0].TeamHistID != 1 {
		t.Errorf("expected TeamHistID=1 (Alpha Squad), got %d", teams[0].TeamHistID)
	}
	// Legacy Order=2 maps to SortOrder=1 (0-indexed).
	if teams[0].SortOrder != 1 {
		t.Errorf("expected SortOrder=1 (Order-1), got %d", teams[0].SortOrder)
	}
}

// TestReadPlayerSeasonTeams_MultiTeam verifies that a player traded between two
// teams during a season has both associations returned with correct sort orders.
func TestReadPlayerSeasonTeams_MultiTeam(t *testing.T) {
	db := testutil.NewLegacyCompanionDB(t)
	ctx := context.Background()

	// Insert a new player season (Id=101) for Alex Power (PlayerId=1) in Season 11.
	// PlayerTeamHistory:
	//   Order=1 → Id=3  (Alpha Squad S11 — current team at end of season)
	//   Order=2 → Id=4  (Beta Ballers S11 — team before the trade)
	_, err := db.ExecContext(ctx, `
		INSERT INTO PlayerSeasons (Id, PlayerId, SeasonId, Age, Salary)
		VALUES (101, 1, 11, 29, 0);
		INSERT INTO PlayerTeamHistory (PlayerSeasonId, SeasonTeamHistoryId, "Order")
		VALUES (101, 3, 1),
		       (101, 4, 2);
	`)
	if err != nil {
		t.Fatalf("seeding multi-team scenario: %v", err)
	}

	reader, err := store.NewLegacyCompanionReader(ctx, db)
	if err != nil {
		t.Fatalf("NewLegacyCompanionReader: %v", err)
	}

	teamsByPS, err := reader.ReadPlayerSeasonTeams(ctx, 1)
	if err != nil {
		t.Fatalf("ReadPlayerSeasonTeams: %v", err)
	}

	teams, ok := teamsByPS[101]
	if !ok {
		t.Fatal("expected an entry for PlayerSeasonId=101; got none")
	}
	if len(teams) != 2 {
		t.Fatalf("expected 2 teams for ps 101, got %d", len(teams))
	}
	// Results are ordered by Order ASC, so index 0 = Order=1 = SortOrder=0.
	if teams[0].TeamHistID != 3 || teams[0].SortOrder != 0 {
		t.Errorf("teams[0]: want {TeamHistID:3 SortOrder:0}, got {TeamHistID:%d SortOrder:%d}",
			teams[0].TeamHistID, teams[0].SortOrder)
	}
	if teams[1].TeamHistID != 4 || teams[1].SortOrder != 1 {
		t.Errorf("teams[1]: want {TeamHistID:4 SortOrder:1}, got {TeamHistID:%d SortOrder:%d}",
			teams[1].TeamHistID, teams[1].SortOrder)
	}
}
