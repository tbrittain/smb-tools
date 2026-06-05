-- Per-player, per-season attribute percentile ranks. One row per player_season.
-- Computed eagerly during season import and legacy migration so that
-- GetPlayerAttributeHistory can retrieve them via a plain JOIN instead of
-- running PERCENT_RANK() window functions at query time.
--
-- Percentile fields are NULL when the player is the only one in a season
-- (PERCENT_RANK is meaningless for a single-row partition).

CREATE TABLE player_season_attribute_percentiles (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    player_season_id INTEGER NOT NULL UNIQUE REFERENCES player_seasons(id),
    power_pct        REAL,
    contact_pct      REAL,
    speed_pct        REAL,
    fielding_pct     REAL,
    arm_pct          REAL,
    velocity_pct     REAL,
    junk_pct         REAL,
    accuracy_pct     REAL
);

-- Backfill percentile ranks for all player seasons that existed before this
-- migration. Uses the same CASE WHEN COUNT(*) OVER > 1 guard as the service
-- layer so single-player seasons produce NULL rather than a misleading 0.
INSERT INTO player_season_attribute_percentiles
    (player_season_id, power_pct, contact_pct, speed_pct, fielding_pct,
     arm_pct, velocity_pct, junk_pct, accuracy_pct)
SELECT
    psg.player_season_id,
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
    CASE WHEN COUNT(*) OVER (PARTITION BY ps.season_id) > 1
         THEN CAST(PERCENT_RANK() OVER (PARTITION BY ps.season_id ORDER BY COALESCE(psg.arm,      0)) * 100 AS REAL)
         ELSE NULL END,
    CASE WHEN COUNT(*) OVER (PARTITION BY ps.season_id) > 1
         THEN CAST(PERCENT_RANK() OVER (PARTITION BY ps.season_id ORDER BY COALESCE(psg.velocity, 0)) * 100 AS REAL)
         ELSE NULL END,
    CASE WHEN COUNT(*) OVER (PARTITION BY ps.season_id) > 1
         THEN CAST(PERCENT_RANK() OVER (PARTITION BY ps.season_id ORDER BY COALESCE(psg.junk,     0)) * 100 AS REAL)
         ELSE NULL END,
    CASE WHEN COUNT(*) OVER (PARTITION BY ps.season_id) > 1
         THEN CAST(PERCENT_RANK() OVER (PARTITION BY ps.season_id ORDER BY COALESCE(psg.accuracy, 0)) * 100 AS REAL)
         ELSE NULL END
FROM player_season_game_stats psg
JOIN player_seasons ps ON ps.id = psg.player_season_id;
