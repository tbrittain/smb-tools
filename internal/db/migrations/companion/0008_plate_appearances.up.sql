-- Add plate_appearances as a persisted column on player_season_batting_stats.
-- Computed on write (import pipeline and legacy migration) rather than derived on read,
-- so it is available for qualification filtering without inline arithmetic.
ALTER TABLE player_season_batting_stats
    ADD COLUMN plate_appearances INTEGER NOT NULL DEFAULT 0;

-- Backfill existing rows from the components already stored.
UPDATE player_season_batting_stats
SET plate_appearances = at_bats + walks + hit_by_pitch + sac_hits + sac_flies;
