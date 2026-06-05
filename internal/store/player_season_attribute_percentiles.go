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
// relevant comparison group, making PERCENT_RANK undefined.
//
// Universal stats (power/contact/speed/fielding) carry both a league-wide and a
// role-specific percentile. Role-exclusive stats (arm, velocity, junk, accuracy)
// are always compared within the player's role group.
type PlayerSeasonAttributePercentiles struct {
	PlayerSeasonID int64
	// League-wide (all players in the season)
	PowerPct    *float64
	ContactPct  *float64
	SpeedPct    *float64
	FieldingPct *float64
	// Role-specific (batter vs batters / pitcher vs pitchers)
	ArmPct      *float64 // batter-only stat; pitchers have 0 arm
	VelocityPct *float64 // pitcher-only stat; batters have 0 velocity
	JunkPct     *float64 // pitcher-only stat
	AccuracyPct *float64 // pitcher-only stat
	// Role-specific for universal stats
	PowerPctRole    *float64
	ContactPctRole  *float64
	SpeedPctRole    *float64
	FieldingPctRole *float64
}

// UpsertPlayerSeasonAttributePercentiles inserts or replaces the percentile row
// for a single player-season. Safe to call multiple times (idempotent).
func (s *PlayerSeasonAttributePercentilesStore) UpsertPlayerSeasonAttributePercentiles(ctx context.Context, p PlayerSeasonAttributePercentiles) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO player_season_attribute_percentiles
		    (player_season_id,
		     power_pct, contact_pct, speed_pct, fielding_pct,
		     arm_pct, velocity_pct, junk_pct, accuracy_pct,
		     power_pct_role, contact_pct_role, speed_pct_role, fielding_pct_role)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		p.PlayerSeasonID,
		p.PowerPct, p.ContactPct, p.SpeedPct, p.FieldingPct,
		p.ArmPct, p.VelocityPct, p.JunkPct, p.AccuracyPct,
		p.PowerPctRole, p.ContactPctRole, p.SpeedPctRole, p.FieldingPctRole,
	)
	if err != nil {
		return fmt.Errorf("upserting attribute percentiles for player_season %d: %w", p.PlayerSeasonID, err)
	}
	return nil
}
