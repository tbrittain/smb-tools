package service

import (
	"context"
	"fmt"

	"smb-tools/internal/store"
)

// ApplyLeagueAvgAttributes computes the league-wide mean for each player
// attribute in the given season and persists the result in
// season_attribute_averages. It must be called after all player game stats
// for the season are written so the aggregate is complete.
//
// Players with a zero value for an attribute are excluded from that
// attribute's average via NULLIF — this prevents inactive or FA-pool players
// (whose attributes default to 0) from dragging down the league mean.
func ApplyLeagueAvgAttributes(ctx context.Context, db store.DBTX, seasonID int64) error {
	var avg store.SeasonAttributeAverages
	avg.SeasonID = seasonID

	err := db.QueryRowContext(ctx, `
		SELECT
		    COALESCE(AVG(NULLIF(psg.power,    0)), 0),
		    COALESCE(AVG(NULLIF(psg.contact,  0)), 0),
		    COALESCE(AVG(NULLIF(psg.speed,    0)), 0),
		    COALESCE(AVG(NULLIF(psg.fielding, 0)), 0),
		    COALESCE(AVG(NULLIF(psg.arm,      0)), 0),
		    COALESCE(AVG(NULLIF(psg.velocity, 0)), 0),
		    COALESCE(AVG(NULLIF(psg.junk,     0)), 0),
		    COALESCE(AVG(NULLIF(psg.accuracy, 0)), 0)
		FROM player_season_game_stats psg
		JOIN player_seasons ps ON ps.id = psg.player_season_id
		WHERE ps.season_id = ?
	`, seasonID).Scan(
		&avg.AvgPower, &avg.AvgContact, &avg.AvgSpeed, &avg.AvgFielding,
		&avg.AvgArm, &avg.AvgVelocity, &avg.AvgJunk, &avg.AvgAccuracy,
	)
	if err != nil {
		return fmt.Errorf("ApplyLeagueAvgAttributes: querying averages for season %d: %w", seasonID, err)
	}

	s := store.NewSeasonAttributeAveragesStore(db)
	if err := s.UpsertSeasonAttributeAverages(ctx, avg); err != nil {
		return fmt.Errorf("ApplyLeagueAvgAttributes: %w", err)
	}
	return nil
}
