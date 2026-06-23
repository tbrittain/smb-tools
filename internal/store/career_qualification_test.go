package store_test

import (
	"context"
	"testing"

	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

func TestGetCareerQualificationThresholds_ScalesByGamesAndInnings(t *testing.T) {
	tests := []struct {
		name           string
		numGames       int
		inningsPerGame int
		wantRS         int
	}{
		{"162 games / 9 innings (unscaled MLB baseline)", 162, 9, 3000},
		{"40 games / 9 innings", 40, 9, 740},
		{"81 games / 9 innings (half season)", 81, 9, 1500},
		{"162 games / 6 innings (shorter games)", 162, 6, 2000},
		{"162 games / 12 innings (longer games)", 162, 12, 4000},
		{"40 games / 6 innings (both scaled down)", 40, 6, 493},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := testutil.NewTestDB(t)
			ctx := context.Background()
			if _, err := db.ExecContext(ctx,
				`INSERT INTO seasons (league_guid, save_game_season_id, season_num, num_games, innings_per_game)
				 VALUES ('TESTLEAGUE', 1, 1, ?, ?)`,
				tc.numGames, tc.inningsPerGame,
			); err != nil {
				t.Fatalf("seeding season: %v", err)
			}

			got, err := store.GetCareerQualificationThresholds(ctx, db)
			if err != nil {
				t.Fatalf("GetCareerQualificationThresholds: %v", err)
			}
			if got.BattingPAThresholdRS != tc.wantRS {
				t.Errorf("BattingPAThresholdRS = %d, want %d", got.BattingPAThresholdRS, tc.wantRS)
			}
			if got.PitchingOutsThresholdRS != tc.wantRS {
				t.Errorf("PitchingOutsThresholdRS = %d, want %d", got.PitchingOutsThresholdRS, tc.wantRS)
			}
		})
	}
}

// TestGetCareerQualificationThresholds_NullInningsLeavesUnscaled verifies that
// a season predating the innings_per_game column (NULL, not 9) doesn't get an
// assumed game length — the RS threshold scales by games only until the
// franchise is backfilled via SeasonStore.BackfillInningsPerGame.
func TestGetCareerQualificationThresholds_NullInningsLeavesUnscaled(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO seasons (league_guid, save_game_season_id, season_num, num_games)
		 VALUES ('TESTLEAGUE', 1, 1, 40)`,
	); err != nil {
		t.Fatalf("seeding season: %v", err)
	}

	got, err := store.GetCareerQualificationThresholds(ctx, db)
	if err != nil {
		t.Fatalf("GetCareerQualificationThresholds: %v", err)
	}
	// Games-only scaling: 3000 * 40/162 = 740 (innings factor 1.0, not assumed-9).
	if got.BattingPAThresholdRS != 740 {
		t.Errorf("BattingPAThresholdRS = %d, want 740 (games-only scaling when innings unknown)", got.BattingPAThresholdRS)
	}
}

func TestGetCareerQualificationThresholds_UsesFirstSeasonOnly(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO seasons (league_guid, save_game_season_id, season_num, num_games, innings_per_game)
		 VALUES ('TESTLEAGUE', 1, 1, 40, 7)`,
	); err != nil {
		t.Fatalf("seeding season 1: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO seasons (league_guid, save_game_season_id, season_num, num_games, innings_per_game)
		 VALUES ('TESTLEAGUE', 2, 2, 162, 9)`,
	); err != nil {
		t.Fatalf("seeding season 2: %v", err)
	}

	got, err := store.GetCareerQualificationThresholds(ctx, db)
	if err != nil {
		t.Fatalf("GetCareerQualificationThresholds: %v", err)
	}
	// Should scale off season 1 (40 games, 7 innings) only, ignoring season 2.
	numGames, innings := 40, 7
	wantF := 3000 * float64(numGames) / 162 * float64(innings) / 9
	want := int(wantF)
	if got.BattingPAThresholdRS != want {
		t.Errorf("BattingPAThresholdRS = %d, want %d (first season only)", got.BattingPAThresholdRS, want)
	}
}

func TestGetCareerQualificationThresholds_NoSeasons(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	got, err := store.GetCareerQualificationThresholds(ctx, db)
	if err != nil {
		t.Fatalf("GetCareerQualificationThresholds: %v", err)
	}
	if got.BattingPAThresholdRS != 0 {
		t.Errorf("BattingPAThresholdRS = %d, want 0 when no seasons exist", got.BattingPAThresholdRS)
	}
	if got.PitchingOutsThresholdRS != 0 {
		t.Errorf("PitchingOutsThresholdRS = %d, want 0 when no seasons exist", got.PitchingOutsThresholdRS)
	}
}

func TestGetCareerQualificationThresholds_POConstantsAreFixed(t *testing.T) {
	db := testutil.NewTestDB(t)
	ctx := context.Background()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO seasons (league_guid, save_game_season_id, season_num, num_games, innings_per_game)
		 VALUES ('TESTLEAGUE', 1, 1, 40, 6)`,
	); err != nil {
		t.Fatalf("seeding season: %v", err)
	}

	got, err := store.GetCareerQualificationThresholds(ctx, db)
	if err != nil {
		t.Fatalf("GetCareerQualificationThresholds: %v", err)
	}
	if got.BattingPAThresholdPO != store.BattingPAThresholdPO {
		t.Errorf("BattingPAThresholdPO = %d, want fixed %d regardless of season length/innings", got.BattingPAThresholdPO, store.BattingPAThresholdPO)
	}
	if got.BattingBBHThresholdPO != store.BattingBBHThresholdPO {
		t.Errorf("BattingBBHThresholdPO = %d, want fixed %d", got.BattingBBHThresholdPO, store.BattingBBHThresholdPO)
	}
	if got.PitchingOutsThresholdPO != store.PitchingOutsThresholdPO {
		t.Errorf("PitchingOutsThresholdPO = %d, want fixed %d", got.PitchingOutsThresholdPO, store.PitchingOutsThresholdPO)
	}
	if got.PitchingDecisionsThresholdPO != store.PitchingDecisionsThresholdPO {
		t.Errorf("PitchingDecisionsThresholdPO = %d, want fixed %d", got.PitchingDecisionsThresholdPO, store.PitchingDecisionsThresholdPO)
	}
}
