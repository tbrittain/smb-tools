package store

import (
	"context"
	"fmt"
)

// SeasonAttributeAveragesStore provides DB operations for the
// season_attribute_averages table. Computation logic lives in the service
// layer; this package only handles SQL.
type SeasonAttributeAveragesStore struct {
	db DBTX
}

func NewSeasonAttributeAveragesStore(db DBTX) *SeasonAttributeAveragesStore {
	return &SeasonAttributeAveragesStore{db: db}
}

// SeasonAttributeAverages holds both the league-wide mean and the role-specific
// means (batter-only, pitcher-only) for each player attribute in one season.
// Zero means no active players with non-zero values (excluded via NULLIF).
type SeasonAttributeAverages struct {
	SeasonID    int64
	// League-wide averages (all players, zeros excluded via NULLIF)
	AvgPower    float64
	AvgContact  float64
	AvgSpeed    float64
	AvgFielding float64
	AvgArm      float64
	AvgVelocity float64
	AvgJunk     float64
	AvgAccuracy float64
	// Batter-only averages (pitcher_role = '')
	BatterAvgPower    float64
	BatterAvgContact  float64
	BatterAvgSpeed    float64
	BatterAvgFielding float64
	BatterAvgArm      float64
	// Pitcher-only averages (pitcher_role != '')
	PitcherAvgPower    float64
	PitcherAvgContact  float64
	PitcherAvgSpeed    float64
	PitcherAvgFielding float64
	PitcherAvgVelocity float64
	PitcherAvgJunk     float64
	PitcherAvgAccuracy float64
}

// UpsertSeasonAttributeAverages inserts or replaces the league-average and
// role-specific attribute rows for the given season. Safe to call multiple
// times for the same season (idempotent re-import).
func (s *SeasonAttributeAveragesStore) UpsertSeasonAttributeAverages(ctx context.Context, avg SeasonAttributeAverages) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO season_attribute_averages
		    (season_id,
		     avg_power, avg_contact, avg_speed, avg_fielding,
		     avg_arm, avg_velocity, avg_junk, avg_accuracy,
		     batter_avg_power, batter_avg_contact, batter_avg_speed, batter_avg_fielding, batter_avg_arm,
		     pitcher_avg_power, pitcher_avg_contact, pitcher_avg_speed, pitcher_avg_fielding,
		     pitcher_avg_velocity, pitcher_avg_junk, pitcher_avg_accuracy)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		avg.SeasonID,
		avg.AvgPower, avg.AvgContact, avg.AvgSpeed, avg.AvgFielding,
		avg.AvgArm, avg.AvgVelocity, avg.AvgJunk, avg.AvgAccuracy,
		avg.BatterAvgPower, avg.BatterAvgContact, avg.BatterAvgSpeed, avg.BatterAvgFielding, avg.BatterAvgArm,
		avg.PitcherAvgPower, avg.PitcherAvgContact, avg.PitcherAvgSpeed, avg.PitcherAvgFielding,
		avg.PitcherAvgVelocity, avg.PitcherAvgJunk, avg.PitcherAvgAccuracy,
	)
	if err != nil {
		return fmt.Errorf("upserting season attribute averages for season %d: %w", avg.SeasonID, err)
	}
	return nil
}
