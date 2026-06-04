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

// SeasonAttributeAverages holds the league-wide mean for each player attribute
// for a single season. Zero means no active players had a non-zero value for
// that attribute (should not occur in practice for a real season).
type SeasonAttributeAverages struct {
	SeasonID    int64
	AvgPower    float64
	AvgContact  float64
	AvgSpeed    float64
	AvgFielding float64
	AvgArm      float64
	AvgVelocity float64
	AvgJunk     float64
	AvgAccuracy float64
}

// UpsertSeasonAttributeAverages inserts or replaces the league-average
// attribute row for the given season. Safe to call multiple times for the same
// season (idempotent re-import).
func (s *SeasonAttributeAveragesStore) UpsertSeasonAttributeAverages(ctx context.Context, avg SeasonAttributeAverages) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO season_attribute_averages
		    (season_id, avg_power, avg_contact, avg_speed, avg_fielding,
		     avg_arm, avg_velocity, avg_junk, avg_accuracy)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		avg.SeasonID,
		avg.AvgPower, avg.AvgContact, avg.AvgSpeed, avg.AvgFielding,
		avg.AvgArm, avg.AvgVelocity, avg.AvgJunk, avg.AvgAccuracy,
	)
	if err != nil {
		return fmt.Errorf("upserting season attribute averages for season %d: %w", avg.SeasonID, err)
	}
	return nil
}
