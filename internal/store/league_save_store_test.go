package store_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"

	"smb-tools/internal/models"
	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

var (
	leagueAGUID = uuid.MustParse("AA000000-0000-0000-0000-000000000000")
	leagueBGUID = uuid.MustParse("BB000000-0000-0000-0000-000000000000")
)

func TestLeagueSaveStore_RewriteLeagueGUID(t *testing.T) {
	db := testutil.NewTestLeagueSaveDB(t)
	s := store.NewLeagueSaveStore(db)
	ctx := context.Background()

	newGUID := uuid.New()
	if err := s.RewriteLeagueGUID(ctx, db, leagueAGUID, newGUID); err != nil {
		t.Fatalf("RewriteLeagueGUID: %v", err)
	}

	// All 6 references for League A must now point at newGUID.
	checks := []struct {
		query string
		want  int
	}{
		{`SELECT COUNT(*) FROM t_leagues WHERE GUID = ?`, 1},
		{`SELECT COUNT(*) FROM t_conferences WHERE leagueGUID = ?`, 2},
		{`SELECT COUNT(*) FROM t_franchise WHERE leagueGUID = ?`, 1},
		{`SELECT COUNT(*) FROM t_seasons WHERE historicalLeagueGUID = ?`, 1},
		{`SELECT COUNT(*) FROM t_league_local_ids WHERE GUID = ?`, 1},
	}
	for _, c := range checks {
		var got int
		if err := db.QueryRowContext(ctx, c.query, newGUID[:]).Scan(&got); err != nil {
			t.Fatalf("query %q: %v", c.query, err)
		}
		if got != c.want {
			t.Errorf("query %q after rewrite: got %d rows referencing newGUID, want %d", c.query, got, c.want)
		}
	}

	// Nothing should reference the old GUID anymore.
	var staleCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_leagues WHERE GUID = ?`, leagueAGUID[:]).Scan(&staleCount); err != nil {
		t.Fatalf("checking for stale old GUID: %v", err)
	}
	if staleCount != 0 {
		t.Errorf("expected 0 rows still referencing the old GUID, got %d", staleCount)
	}

	// League B must be completely untouched.
	var leagueBName string
	if err := db.QueryRowContext(ctx, `SELECT name FROM t_leagues WHERE GUID = ?`, leagueBGUID[:]).Scan(&leagueBName); err != nil {
		t.Fatalf("League B should still exist under its original GUID: %v", err)
	}
	if leagueBName != "League B" {
		t.Errorf("League B name = %q, want %q", leagueBName, "League B")
	}
}

func TestLeagueSaveStore_RewriteLeagueGUID_NoMatchingRows(t *testing.T) {
	db := testutil.NewTestLeagueSaveDB(t)
	s := store.NewLeagueSaveStore(db)
	ctx := context.Background()

	// Rewriting a GUID that doesn't exist in the save should succeed (no
	// rows affected) rather than error — it's a no-op UPDATE, not a lookup.
	unknownGUID := uuid.New()
	newGUID := uuid.New()
	if err := s.RewriteLeagueGUID(ctx, db, unknownGUID, newGUID); err != nil {
		t.Fatalf("RewriteLeagueGUID with no matching rows should not error: %v", err)
	}

	// League A and League B must both be untouched.
	for _, guid := range []uuid.UUID{leagueAGUID, leagueBGUID} {
		var count int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_leagues WHERE GUID = ?`, guid[:]).Scan(&count); err != nil {
			t.Fatalf("checking league %s: %v", guid, err)
		}
		if count != 1 {
			t.Errorf("league %s: got %d rows, want 1 (should be untouched)", guid, count)
		}
	}
}

func TestLeagueSaveStore_ValidateLeagueSaveShape_Valid(t *testing.T) {
	db := testutil.NewTestLeagueSaveDB(t)
	s := store.NewLeagueSaveStore(db)
	ctx := context.Background()

	if err := s.ValidateLeagueSaveShape(ctx); err != nil {
		t.Errorf("ValidateLeagueSaveShape on a real-shaped fixture: %v", err)
	}
}

func TestLeagueSaveStore_ValidateLeagueSaveShape_MissingTable(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec(`CREATE TABLE t_leagues (GUID BLOB PRIMARY KEY)`); err != nil {
		t.Fatalf("creating partial schema: %v", err)
	}

	s := store.NewLeagueSaveStore(db)
	if err := s.ValidateLeagueSaveShape(context.Background()); err == nil {
		t.Error("expected an error for a DB missing most required tables, got nil")
	}
}

func TestLeagueSaveStore_ValidateLeagueSaveShape_PassesWithNoFranchiseRows(t *testing.T) {
	// A stock or season/elimination-mode league has zero rows in t_franchise.
	// Shape validation must check table presence, not row presence.
	db := testutil.NewTestLeagueSaveDB(t)
	if _, err := db.Exec(`DELETE FROM t_franchise`); err != nil {
		t.Fatalf("clearing t_franchise rows: %v", err)
	}

	s := store.NewLeagueSaveStore(db)
	if err := s.ValidateLeagueSaveShape(context.Background()); err != nil {
		t.Errorf("ValidateLeagueSaveShape should pass with an empty (but present) t_franchise table: %v", err)
	}
}

func TestLeagueSaveStore_GetLeagueOverview_WithDivisions(t *testing.T) {
	db := testutil.NewTestLeagueSaveDB(t)
	s := store.NewLeagueSaveStore(db)
	ctx := context.Background()

	overview, err := s.GetLeagueOverview(ctx, leagueAGUID)
	if err != nil {
		t.Fatalf("GetLeagueOverview: %v", err)
	}

	if overview.Name != "League A" {
		t.Errorf("Name = %q, want %q", overview.Name, "League A")
	}
	if len(overview.Conferences) != 2 {
		t.Fatalf("expected 2 conferences, got %d", len(overview.Conferences))
	}

	var foundEast, foundWest bool
	for _, c := range overview.Conferences {
		teamCount := 0
		for _, d := range c.Divisions {
			teamCount += len(d.Teams)
		}
		switch c.Name {
		case "East":
			foundEast = true
			if len(c.Divisions) != 1 {
				t.Errorf("East: expected 1 division, got %d", len(c.Divisions))
			}
			if teamCount != 2 {
				t.Errorf("East: expected 2 teams, got %d", teamCount)
			}
		case "West":
			foundWest = true
			if len(c.Divisions) != 0 {
				t.Errorf("West: expected 0 divisions (divisions are optional), got %d", len(c.Divisions))
			}
		}
	}

	if !foundEast {
		t.Error("expected an 'East' conference")
	}
	if !foundWest {
		t.Error("expected a 'West' conference")
	}
	if overview.Mode != models.LeagueModeFranchise {
		t.Errorf("Mode = %q, want %q (League A has a t_franchise row)", overview.Mode, models.LeagueModeFranchise)
	}
}

func TestLeagueSaveStore_GetLeagueOverview_Mode(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(t *testing.T, db *sql.DB)
		wantMode models.LeagueMode
	}{
		{
			name:     "franchise present",
			mutate:   func(t *testing.T, db *sql.DB) {},
			wantMode: models.LeagueModeFranchise,
		},
		{
			name: "no franchise, elimination season",
			mutate: func(t *testing.T, db *sql.DB) {
				if _, err := db.Exec(`DELETE FROM t_franchise WHERE leagueGUID = ?`, leagueAGUID[:]); err != nil {
					t.Fatalf("clearing franchise: %v", err)
				}
				if _, err := db.Exec(`UPDATE t_seasons SET elimination = 1 WHERE historicalLeagueGUID = ?`, leagueAGUID[:]); err != nil {
					t.Fatalf("marking season as elimination: %v", err)
				}
			},
			wantMode: models.LeagueModeElimination,
		},
		{
			name: "no franchise, regular season",
			mutate: func(t *testing.T, db *sql.DB) {
				if _, err := db.Exec(`DELETE FROM t_franchise WHERE leagueGUID = ?`, leagueAGUID[:]); err != nil {
					t.Fatalf("clearing franchise: %v", err)
				}
			},
			wantMode: models.LeagueModeSeason,
		},
		{
			name: "no franchise, no seasons - empty shell",
			mutate: func(t *testing.T, db *sql.DB) {
				if _, err := db.Exec(`DELETE FROM t_franchise WHERE leagueGUID = ?`, leagueAGUID[:]); err != nil {
					t.Fatalf("clearing franchise: %v", err)
				}
				if _, err := db.Exec(`DELETE FROM t_seasons WHERE historicalLeagueGUID = ?`, leagueAGUID[:]); err != nil {
					t.Fatalf("clearing seasons: %v", err)
				}
			},
			wantMode: models.LeagueModeNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testutil.NewTestLeagueSaveDB(t)
			tt.mutate(t, db)

			s := store.NewLeagueSaveStore(db)
			overview, err := s.GetLeagueOverview(context.Background(), leagueAGUID)
			if err != nil {
				t.Fatalf("GetLeagueOverview: %v", err)
			}
			if overview.Mode != tt.wantMode {
				t.Errorf("Mode = %q, want %q", overview.Mode, tt.wantMode)
			}
		})
	}
}

func TestLeagueSaveStore_RenameLeague(t *testing.T) {
	db := testutil.NewTestLeagueSaveDB(t)
	s := store.NewLeagueSaveStore(db)
	ctx := context.Background()

	if err := s.RenameLeague(ctx, leagueAGUID, "Renamed League A"); err != nil {
		t.Fatalf("RenameLeague: %v", err)
	}

	var name string
	if err := db.QueryRowContext(ctx, `SELECT name FROM t_leagues WHERE GUID = ?`, leagueAGUID[:]).Scan(&name); err != nil {
		t.Fatalf("reading renamed league: %v", err)
	}
	if name != "Renamed League A" {
		t.Errorf("name = %q, want %q", name, "Renamed League A")
	}

	// League B must be completely untouched.
	var leagueBName string
	if err := db.QueryRowContext(ctx, `SELECT name FROM t_leagues WHERE GUID = ?`, leagueBGUID[:]).Scan(&leagueBName); err != nil {
		t.Fatalf("League B should still exist: %v", err)
	}
	if leagueBName != "League B" {
		t.Errorf("League B name = %q, want %q", leagueBName, "League B")
	}
}

func TestLeagueSaveStore_RenameLeague_UnknownGUID(t *testing.T) {
	db := testutil.NewTestLeagueSaveDB(t)
	s := store.NewLeagueSaveStore(db)

	if err := s.RenameLeague(context.Background(), uuid.New(), "New Name"); err == nil {
		t.Error("expected an error for an unknown league GUID, got nil")
	}
}

func TestLeagueSaveStore_GetLeagueOverview_UnknownLeague(t *testing.T) {
	db := testutil.NewTestLeagueSaveDB(t)
	s := store.NewLeagueSaveStore(db)

	_, err := s.GetLeagueOverview(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected an error for an unknown league GUID, got nil")
	}
}
