package store

import (
	"context"
	"fmt"
)

// PlayerSeasonAttributePercentilesStore provides DB operations for the
// player_season_attribute_percentiles table.
type PlayerSeasonAttributePercentilesStore struct {
	db DBTX
}

func NewPlayerSeasonAttributePercentilesStore(db DBTX) *PlayerSeasonAttributePercentilesStore {
	return &PlayerSeasonAttributePercentilesStore{db: db}
}

// PlayerSeasonAttributePercentiles holds the percentile rank for each attribute
// for a single player-season. Nil means the player was the only one in the
// season, making PERCENT_RANK undefined.
type PlayerSeasonAttributePercentiles struct {
	PlayerSeasonID int64
	PowerPct       *float64
	ContactPct     *float64
	SpeedPct       *float64
	FieldingPct    *float64
	ArmPct         *float64
	VelocityPct    *float64
	JunkPct        *float64
	AccuracyPct    *float64
}

// UpsertPlayerSeasonAttributePercentiles inserts or replaces the percentile row
// for a single player-season. Safe to call multiple times (idempotent).
func (s *PlayerSeasonAttributePercentilesStore) UpsertPlayerSeasonAttributePercentiles(ctx context.Context, p PlayerSeasonAttributePercentiles) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO player_season_attribute_percentiles
		    (player_season_id, power_pct, contact_pct, speed_pct, fielding_pct,
		     arm_pct, velocity_pct, junk_pct, accuracy_pct)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		p.PlayerSeasonID,
		p.PowerPct, p.ContactPct, p.SpeedPct, p.FieldingPct,
		p.ArmPct, p.VelocityPct, p.JunkPct, p.AccuracyPct,
	)
	if err != nil {
		return fmt.Errorf("upserting attribute percentiles for player_season %d: %w", p.PlayerSeasonID, err)
	}
	return nil
}
