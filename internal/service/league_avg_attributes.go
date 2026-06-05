package service

import (
	"context"
	"fmt"

	"smb-tools/internal/store"
)

// ApplyLeagueAvgAttributes computes the league-wide mean and role-specific
// means (batter-only, pitcher-only) for each player attribute in the given
// season and persists the result in season_attribute_averages.
//
// Must be called after all player game stats for the season are written.
//
// Players with a zero value for an attribute are excluded via NULLIF so that
// inactive or FA-pool players (whose attributes default to 0) do not drag down
// the league mean. The FILTER clause applies the same exclusion per role group.
func ApplyLeagueAvgAttributes(ctx context.Context, db store.DBTX, seasonID int64) error {
	var avg store.SeasonAttributeAverages
	avg.SeasonID = seasonID

	err := db.QueryRowContext(ctx, `
		SELECT
		    -- league-wide (all players, zeros excluded)
		    COALESCE(AVG(NULLIF(psg.power,    0)), 0),
		    COALESCE(AVG(NULLIF(psg.contact,  0)), 0),
		    COALESCE(AVG(NULLIF(psg.speed,    0)), 0),
		    COALESCE(AVG(NULLIF(psg.fielding, 0)), 0),
		    COALESCE(AVG(NULLIF(psg.arm,      0)), 0),
		    COALESCE(AVG(NULLIF(psg.velocity, 0)), 0),
		    COALESCE(AVG(NULLIF(psg.junk,     0)), 0),
		    COALESCE(AVG(NULLIF(psg.accuracy, 0)), 0),
		    -- batter-only (pitcher_role = '')
		    COALESCE(AVG(NULLIF(psg.power,    0)) FILTER (WHERE ps.pitcher_role = ''), 0),
		    COALESCE(AVG(NULLIF(psg.contact,  0)) FILTER (WHERE ps.pitcher_role = ''), 0),
		    COALESCE(AVG(NULLIF(psg.speed,    0)) FILTER (WHERE ps.pitcher_role = ''), 0),
		    COALESCE(AVG(NULLIF(psg.fielding, 0)) FILTER (WHERE ps.pitcher_role = ''), 0),
		    COALESCE(AVG(NULLIF(psg.arm,      0)) FILTER (WHERE ps.pitcher_role = ''), 0),
		    -- pitcher-only (pitcher_role != '')
		    COALESCE(AVG(NULLIF(psg.power,    0)) FILTER (WHERE ps.pitcher_role != ''), 0),
		    COALESCE(AVG(NULLIF(psg.contact,  0)) FILTER (WHERE ps.pitcher_role != ''), 0),
		    COALESCE(AVG(NULLIF(psg.speed,    0)) FILTER (WHERE ps.pitcher_role != ''), 0),
		    COALESCE(AVG(NULLIF(psg.fielding, 0)) FILTER (WHERE ps.pitcher_role != ''), 0),
		    COALESCE(AVG(NULLIF(psg.velocity, 0)) FILTER (WHERE ps.pitcher_role != ''), 0),
		    COALESCE(AVG(NULLIF(psg.junk,     0)) FILTER (WHERE ps.pitcher_role != ''), 0),
		    COALESCE(AVG(NULLIF(psg.accuracy, 0)) FILTER (WHERE ps.pitcher_role != ''), 0)
		FROM player_season_game_stats psg
		JOIN player_seasons ps ON ps.id = psg.player_season_id
		WHERE ps.season_id = ?
	`, seasonID).Scan(
		&avg.AvgPower, &avg.AvgContact, &avg.AvgSpeed, &avg.AvgFielding,
		&avg.AvgArm, &avg.AvgVelocity, &avg.AvgJunk, &avg.AvgAccuracy,
		&avg.BatterAvgPower, &avg.BatterAvgContact, &avg.BatterAvgSpeed, &avg.BatterAvgFielding, &avg.BatterAvgArm,
		&avg.PitcherAvgPower, &avg.PitcherAvgContact, &avg.PitcherAvgSpeed, &avg.PitcherAvgFielding,
		&avg.PitcherAvgVelocity, &avg.PitcherAvgJunk, &avg.PitcherAvgAccuracy,
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
