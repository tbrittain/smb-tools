package store

import (
	"context"
	"database/sql"
	"fmt"
)

// Fixed, unscaled MLB career postseason qualification minimums (per
// Baseball-Reference): these don't scale with season length or innings per
// game since playoff PA/IP totals are inherently small regardless of how long
// the regular season or its games run.
const (
	BattingPAThresholdPO         = 40 // career playoff PA
	BattingBBHThresholdPO        = 18 // career playoff BB+H
	PitchingOutsThresholdPO      = 90 // career playoff outs pitched (30 IP)
	PitchingDecisionsThresholdPO = 6  // career playoff decisions (W+L)
)

// CareerQualificationThresholds holds the minimums used to gate career
// batting/pitching stat leaders and records. RS thresholds scale Baseball
// Reference's 3000 PA / 1000 IP career minimums off the franchise's first
// season's games-per-season and innings-per-game (relative to a 162-game,
// 9-inning MLB season). PO thresholds are fixed and unscaled.
type CareerQualificationThresholds struct {
	BattingPAThresholdRS         int
	PitchingOutsThresholdRS      int
	BattingPAThresholdPO         int
	BattingBBHThresholdPO        int
	PitchingOutsThresholdPO      int
	PitchingDecisionsThresholdPO int
}

// GetCareerQualificationThresholds computes the career qualification
// thresholds for the franchise. RS thresholds are scaled off the earliest
// season's num_games and innings_per_game. numGames defaults to 162 if no
// seasons exist yet (yielding the unscaled 3000 threshold, though with no
// seasons no players can qualify anyway). innings_per_game can be NULL for
// seasons synced before that column existed — there is no assumed game
// length for those rows, so the innings factor is left unscaled (1.0) until
// the franchise is backfilled via SeasonStore.BackfillInningsPerGame.
func GetCareerQualificationThresholds(ctx context.Context, db DBTX) (CareerQualificationThresholds, error) {
	numGames := 162
	var inningsPerGame sql.NullInt64
	err := db.QueryRowContext(ctx,
		`SELECT num_games, innings_per_game FROM seasons ORDER BY season_num LIMIT 1`,
	).Scan(&numGames, &inningsPerGame)
	if err != nil && err != sql.ErrNoRows {
		return CareerQualificationThresholds{}, fmt.Errorf("GetCareerQualificationThresholds: %w", err)
	}

	inningsFactor := 1.0
	if inningsPerGame.Valid {
		inningsFactor = float64(inningsPerGame.Int64) / 9
	}
	scaledRS := int(3000 * float64(numGames) / 162 * inningsFactor)
	if err == sql.ErrNoRows {
		scaledRS = 0
	}

	return CareerQualificationThresholds{
		BattingPAThresholdRS:         scaledRS,
		PitchingOutsThresholdRS:      scaledRS,
		BattingPAThresholdPO:         BattingPAThresholdPO,
		BattingBBHThresholdPO:        BattingBBHThresholdPO,
		PitchingOutsThresholdPO:      PitchingOutsThresholdPO,
		PitchingDecisionsThresholdPO: PitchingDecisionsThresholdPO,
	}, nil
}
