-- Role-split attribute percentiles and averages.
--
-- Universal batting stats (power/contact/speed/fielding) support two comparisons:
--   *_pct (existing)      = vs. entire league (all players in season)
--   *_pct_role (new)      = vs. own role group (batters vs batters, pitchers vs pitchers)
-- Role-exclusive stats are corrected to their natural group:
--   arm_pct               = batter vs batters (pitchers have 0 arm)
--   velocity/junk/accuracy_pct = pitcher vs pitchers (batters have 0 pitching stats)
--
-- season_attribute_averages gains role-specific means so the chart can show a
-- role-appropriate reference line:
--   batter_avg_{power,contact,speed,fielding,arm}
--   pitcher_avg_{power,contact,speed,fielding,velocity,junk,accuracy}

ALTER TABLE player_season_attribute_percentiles ADD COLUMN power_pct_role   REAL;
ALTER TABLE player_season_attribute_percentiles ADD COLUMN contact_pct_role  REAL;
ALTER TABLE player_season_attribute_percentiles ADD COLUMN speed_pct_role    REAL;
ALTER TABLE player_season_attribute_percentiles ADD COLUMN fielding_pct_role REAL;

ALTER TABLE season_attribute_averages ADD COLUMN batter_avg_power    REAL NOT NULL DEFAULT 0;
ALTER TABLE season_attribute_averages ADD COLUMN batter_avg_contact  REAL NOT NULL DEFAULT 0;
ALTER TABLE season_attribute_averages ADD COLUMN batter_avg_speed    REAL NOT NULL DEFAULT 0;
ALTER TABLE season_attribute_averages ADD COLUMN batter_avg_fielding REAL NOT NULL DEFAULT 0;
ALTER TABLE season_attribute_averages ADD COLUMN batter_avg_arm      REAL NOT NULL DEFAULT 0;
ALTER TABLE season_attribute_averages ADD COLUMN pitcher_avg_power    REAL NOT NULL DEFAULT 0;
ALTER TABLE season_attribute_averages ADD COLUMN pitcher_avg_contact  REAL NOT NULL DEFAULT 0;
ALTER TABLE season_attribute_averages ADD COLUMN pitcher_avg_speed    REAL NOT NULL DEFAULT 0;
ALTER TABLE season_attribute_averages ADD COLUMN pitcher_avg_fielding REAL NOT NULL DEFAULT 0;
ALTER TABLE season_attribute_averages ADD COLUMN pitcher_avg_velocity REAL NOT NULL DEFAULT 0;
ALTER TABLE season_attribute_averages ADD COLUMN pitcher_avg_junk     REAL NOT NULL DEFAULT 0;
ALTER TABLE season_attribute_averages ADD COLUMN pitcher_avg_accuracy REAL NOT NULL DEFAULT 0;

-- Backfill percentile rows. Replaces rows written by migration 0013 (which used
-- league-wide partition for all 8 attributes) with correct values:
--   power/contact/speed/fielding keep league-wide AND gain _pct_role variants.
--   arm/velocity/junk/accuracy are re-partitioned by role (correcting 0013).
INSERT OR REPLACE INTO player_season_attribute_percentiles
    (player_season_id,
     power_pct, contact_pct, speed_pct, fielding_pct,
     arm_pct, velocity_pct, junk_pct, accuracy_pct,
     power_pct_role, contact_pct_role, speed_pct_role, fielding_pct_role)
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
    -- role-specific for universal stats (new columns)
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
JOIN player_seasons ps ON ps.id = psg.player_season_id;

-- Backfill role-specific averages using FILTER aggregates (SQLite 3.30+).
WITH role_avgs AS (
    SELECT
        ps.season_id,
        COALESCE(AVG(NULLIF(psg.power,    0)) FILTER (WHERE ps.pitcher_role = ''), 0) AS batter_avg_power,
        COALESCE(AVG(NULLIF(psg.contact,  0)) FILTER (WHERE ps.pitcher_role = ''), 0) AS batter_avg_contact,
        COALESCE(AVG(NULLIF(psg.speed,    0)) FILTER (WHERE ps.pitcher_role = ''), 0) AS batter_avg_speed,
        COALESCE(AVG(NULLIF(psg.fielding, 0)) FILTER (WHERE ps.pitcher_role = ''), 0) AS batter_avg_fielding,
        COALESCE(AVG(NULLIF(psg.arm,      0)) FILTER (WHERE ps.pitcher_role = ''), 0) AS batter_avg_arm,
        COALESCE(AVG(NULLIF(psg.power,    0)) FILTER (WHERE ps.pitcher_role != ''), 0) AS pitcher_avg_power,
        COALESCE(AVG(NULLIF(psg.contact,  0)) FILTER (WHERE ps.pitcher_role != ''), 0) AS pitcher_avg_contact,
        COALESCE(AVG(NULLIF(psg.speed,    0)) FILTER (WHERE ps.pitcher_role != ''), 0) AS pitcher_avg_speed,
        COALESCE(AVG(NULLIF(psg.fielding, 0)) FILTER (WHERE ps.pitcher_role != ''), 0) AS pitcher_avg_fielding,
        COALESCE(AVG(NULLIF(psg.velocity, 0)) FILTER (WHERE ps.pitcher_role != ''), 0) AS pitcher_avg_velocity,
        COALESCE(AVG(NULLIF(psg.junk,     0)) FILTER (WHERE ps.pitcher_role != ''), 0) AS pitcher_avg_junk,
        COALESCE(AVG(NULLIF(psg.accuracy, 0)) FILTER (WHERE ps.pitcher_role != ''), 0) AS pitcher_avg_accuracy
    FROM player_season_game_stats psg
    JOIN player_seasons ps ON ps.id = psg.player_season_id
    GROUP BY ps.season_id
)
UPDATE season_attribute_averages
SET
    batter_avg_power    = (SELECT batter_avg_power    FROM role_avgs WHERE role_avgs.season_id = season_attribute_averages.season_id),
    batter_avg_contact  = (SELECT batter_avg_contact  FROM role_avgs WHERE role_avgs.season_id = season_attribute_averages.season_id),
    batter_avg_speed    = (SELECT batter_avg_speed    FROM role_avgs WHERE role_avgs.season_id = season_attribute_averages.season_id),
    batter_avg_fielding = (SELECT batter_avg_fielding FROM role_avgs WHERE role_avgs.season_id = season_attribute_averages.season_id),
    batter_avg_arm      = (SELECT batter_avg_arm      FROM role_avgs WHERE role_avgs.season_id = season_attribute_averages.season_id),
    pitcher_avg_power    = (SELECT pitcher_avg_power    FROM role_avgs WHERE role_avgs.season_id = season_attribute_averages.season_id),
    pitcher_avg_contact  = (SELECT pitcher_avg_contact  FROM role_avgs WHERE role_avgs.season_id = season_attribute_averages.season_id),
    pitcher_avg_speed    = (SELECT pitcher_avg_speed    FROM role_avgs WHERE role_avgs.season_id = season_attribute_averages.season_id),
    pitcher_avg_fielding = (SELECT pitcher_avg_fielding FROM role_avgs WHERE role_avgs.season_id = season_attribute_averages.season_id),
    pitcher_avg_velocity = (SELECT pitcher_avg_velocity FROM role_avgs WHERE role_avgs.season_id = season_attribute_averages.season_id),
    pitcher_avg_junk     = (SELECT pitcher_avg_junk     FROM role_avgs WHERE role_avgs.season_id = season_attribute_averages.season_id),
    pitcher_avg_accuracy = (SELECT pitcher_avg_accuracy FROM role_avgs WHERE role_avgs.season_id = season_attribute_averages.season_id)
WHERE EXISTS (SELECT 1 FROM role_avgs WHERE role_avgs.season_id = season_attribute_averages.season_id);
