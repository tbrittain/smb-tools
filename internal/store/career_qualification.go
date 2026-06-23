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
// season's num_games and innings_per_game; both default to 162 and 9
// respectively if no seasons exist yet (yielding the unscaled 3000 threshold,
// though with no seasons no players can qualify anyway).
func GetCareerQualificationThresholds(ctx context.Context, db DBTX) (CareerQualificationThresholds, error) {
	numGames, inningsPerGame := 162, 9
	err := db.QueryRowContext(ctx,
		`SELECT num_games, innings_per_game FROM seasons ORDER BY season_num LIMIT 1`,
	).Scan(&numGames, &inningsPerGame)
	if err != nil && err != sql.ErrNoRows {
		return CareerQualificationThresholds{}, fmt.Errorf("GetCareerQualificationThresholds: %w", err)
	}

	scaledRS := int(3000 * float64(numGames) / 162 * float64(inningsPerGame) / 9)
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
