package service

import (
	"context"
	"database/sql"
	"fmt"

	"smb-tools/internal/store"
)

// nullFloat returns a pointer to f.Float64 when f is valid, nil otherwise.
// Used to convert sql.NullFloat64 scan results to *float64 without repeating
// the if-valid pattern per row.
func nullFloat(f sql.NullFloat64) *float64 {
	if !f.Valid {
		return nil
	}
	v := f.Float64
	return &v
}

// ApplyPlayerAttributePercentiles computes per-player attribute percentile
// ranks for all players in the given season and persists them in
// player_season_attribute_percentiles.
//
// Universal stats (power/contact/speed/fielding) produce two percentile columns:
//   - *_pct: league-wide (all players in the season)
//   - *_pct_role: role-specific (batter vs batters, pitcher vs pitchers)
//
// Role-exclusive stats (arm, velocity, junk, accuracy) produce one column each,
// partitioned by role: arm is batter vs batters, pitching stats are pitcher vs pitchers.
//
// Percentile fields are NULL when the comparison group has only one player
// (PERCENT_RANK is undefined for a single-row partition).
//
// Must be called after all player game stats for the season are written
// (same precondition as ApplyLeagueAvgAttributes).
func ApplyPlayerAttributePercentiles(ctx context.Context, db store.DBTX, seasonID int64) error {
	rows, err := db.QueryContext(ctx, `
		SELECT
		    psg.player_season_id,
		    -- league-wide (PARTITION BY season_id only)
		    CASE WHEN COUNT(*) OVER (PARTITION BY ps.season_id) > 1
		         THEN CAST(PERCENT_RANK() OVER (PARTITION BY ps.season_id ORDER BY COALESCE(psg.power,    0)) * 100 AS REAL)
		         ELSE NULL END,
		    CASE WHEN COUNT(*) OVER (PARTITION BY ps.season_id) > 1
		         THEN CAST(PERCENT_RANK() OVER (PARTITION BY ps.season_id ORDER BY COALESCE(psg.contact,  0)) * 100 AS REAL)
		         ELSE NULL END,
		    CASE WHEN COUNT(*) OVER (PARTITION BY ps.season_id) > 1
		         THEN CAST(PERCENT_RANK() OVER (PARTITION BY ps.season_id ORDER BY COALESCE(psg.speed,    0)) * 100 AS REAL)
		         ELSE NULL END,
		    CASE WHEN COUNT(*) OVER (PARTITION BY ps.season_id) > 1
		         THEN CAST(PERCENT_RANK() OVER (PARTITION BY ps.season_id ORDER BY COALESCE(psg.fielding, 0)) * 100 AS REAL)
		         ELSE NULL END,
		    -- role-specific (PARTITION BY season_id, is_pitcher)
		    CASE WHEN COUNT(*) OVER (PARTITION BY ps.season_id, CASE WHEN ps.pitcher_role != '' THEN 1 ELSE 0 END) > 1
		         THEN CAST(PERCENT_RANK() OVER (PARTITION BY ps.season_id, CASE WHEN ps.pitcher_role != '' THEN 1 ELSE 0 END ORDER BY COALESCE(psg.arm,      0)) * 100 AS REAL)
		         ELSE NULL END,
		    CASE WHEN COUNT(*) OVER (PARTITION BY ps.season_id, CASE WHEN ps.pitcher_role != '' THEN 1 ELSE 0 END) > 1
		         THEN CAST(PERCENT_RANK() OVER (PARTITION BY ps.season_id, CASE WHEN ps.pitcher_role != '' THEN 1 ELSE 0 END ORDER BY COALESCE(psg.velocity, 0)) * 100 AS REAL)
		         ELSE NULL END,
		    CASE WHEN COUNT(*) OVER (PARTITION BY ps.season_id, CASE WHEN ps.pitcher_role != '' THEN 1 ELSE 0 END) > 1
		         THEN CAST(PERCENT_RANK() OVER (PARTITION BY ps.season_id, CASE WHEN ps.pitcher_role != '' THEN 1 ELSE 0 END ORDER BY COALESCE(psg.junk,     0)) * 100 AS REAL)
		         ELSE NULL END,
		    CASE WHEN COUNT(*) OVER (PARTITION BY ps.season_id, CASE WHEN ps.pitcher_role != '' THEN 1 ELSE 0 END) > 1
		         THEN CAST(PERCENT_RANK() OVER (PARTITION BY ps.season_id, CASE WHEN ps.pitcher_role != '' THEN 1 ELSE 0 END ORDER BY COALESCE(psg.accuracy, 0)) * 100 AS REAL)
		         ELSE NULL END,
		    -- role-specific for universal stats (new _pct_role columns)
		    CASE WHEN COUNT(*) OVER (PARTITION BY ps.season_id, CASE WHEN ps.pitcher_role != '' THEN 1 ELSE 0 END) > 1
		         THEN CAST(PERCENT_RANK() OVER (PARTITION BY ps.season_id, CASE WHEN ps.pitcher_role != '' THEN 1 ELSE 0 END ORDER BY COALESCE(psg.power,    0)) * 100 AS REAL)
		         ELSE NULL END,
		    CASE WHEN COUNT(*) OVER (PARTITION BY ps.season_id, CASE WHEN ps.pitcher_role != '' THEN 1 ELSE 0 END) > 1
		         THEN CAST(PERCENT_RANK() OVER (PARTITION BY ps.season_id, CASE WHEN ps.pitcher_role != '' THEN 1 ELSE 0 END ORDER BY COALESCE(psg.contact,  0)) * 100 AS REAL)
		         ELSE NULL END,
		    CASE WHEN COUNT(*) OVER (PARTITION BY ps.season_id, CASE WHEN ps.pitcher_role != '' THEN 1 ELSE 0 END) > 1
		         THEN CAST(PERCENT_RANK() OVER (PARTITION BY ps.season_id, CASE WHEN ps.pitcher_role != '' THEN 1 ELSE 0 END ORDER BY COALESCE(psg.speed,    0)) * 100 AS REAL)
		         ELSE NULL END,
		    CASE WHEN COUNT(*) OVER (PARTITION BY ps.season_id, CASE WHEN ps.pitcher_role != '' THEN 1 ELSE 0 END) > 1
		         THEN CAST(PERCENT_RANK() OVER (PARTITION BY ps.season_id, CASE WHEN ps.pitcher_role != '' THEN 1 ELSE 0 END ORDER BY COALESCE(psg.fielding, 0)) * 100 AS REAL)
		         ELSE NULL END
		FROM player_season_game_stats psg
		JOIN player_seasons ps ON ps.id = psg.player_season_id
		WHERE ps.season_id = ?
	`, seasonID)
	if err != nil {
		return fmt.Errorf("ApplyPlayerAttributePercentiles: querying percentiles for season %d: %w", seasonID, err)
	}

	// Collect all rows before closing the cursor. Holding rows open while
	// executing INSERTs would require a second connection — on an in-memory
	// SQLite DB each connection sees its own schema, causing "no such table".
	var percentiles []store.PlayerSeasonAttributePercentiles
	for rows.Next() {
		var p store.PlayerSeasonAttributePercentiles
		var powerPct, contactPct, speedPct, fieldingPct sql.NullFloat64
		var armPct, velocityPct, junkPct, accuracyPct sql.NullFloat64
		var powerPctRole, contactPctRole, speedPctRole, fieldingPctRole sql.NullFloat64
		if err := rows.Scan(
			&p.PlayerSeasonID,
			&powerPct, &contactPct, &speedPct, &fieldingPct,
			&armPct, &velocityPct, &junkPct, &accuracyPct,
			&powerPctRole, &contactPctRole, &speedPctRole, &fieldingPctRole,
		); err != nil {
			_ = rows.Close()
			return fmt.Errorf("ApplyPlayerAttributePercentiles: scanning row: %w", err)
		}
		p.PowerPct        = nullFloat(powerPct)
		p.ContactPct      = nullFloat(contactPct)
		p.SpeedPct        = nullFloat(speedPct)
		p.FieldingPct     = nullFloat(fieldingPct)
		p.ArmPct          = nullFloat(armPct)
		p.VelocityPct     = nullFloat(velocityPct)
		p.JunkPct         = nullFloat(junkPct)
		p.AccuracyPct     = nullFloat(accuracyPct)
		p.PowerPctRole    = nullFloat(powerPctRole)
		p.ContactPctRole  = nullFloat(contactPctRole)
		p.SpeedPctRole    = nullFloat(speedPctRole)
		p.FieldingPctRole = nullFloat(fieldingPctRole)
		percentiles = append(percentiles, p)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("ApplyPlayerAttributePercentiles: closing rows: %w", err)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("ApplyPlayerAttributePercentiles: iterating rows: %w", err)
	}

	s := store.NewPlayerSeasonAttributePercentilesStore(db)
	for _, p := range percentiles {
		if err := s.UpsertPlayerSeasonAttributePercentiles(ctx, p); err != nil {
			return fmt.Errorf("ApplyPlayerAttributePercentiles: %w", err)
		}
	}
	return nil
}
