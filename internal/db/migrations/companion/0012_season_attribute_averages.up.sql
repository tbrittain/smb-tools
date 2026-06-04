-- Per-season league-wide attribute averages, computed eagerly during season
-- import and legacy migration. One row per season. Used as a reference overlay
-- on the player attribute trend chart and to avoid re-aggregating at query time.
--
-- avg_* columns store the mean of each attribute across all players in the
-- season who have a non-zero value for that attribute (inactive/FA-pool players
-- with all-zero attributes are excluded via NULLIF in the computation).

CREATE TABLE season_attribute_averages (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    season_id    INTEGER NOT NULL UNIQUE REFERENCES seasons(id),
    avg_power    REAL NOT NULL DEFAULT 0,
    avg_contact  REAL NOT NULL DEFAULT 0,
    avg_speed    REAL NOT NULL DEFAULT 0,
    avg_fielding REAL NOT NULL DEFAULT 0,
    avg_arm      REAL NOT NULL DEFAULT 0,
    avg_velocity REAL NOT NULL DEFAULT 0,
    avg_junk     REAL NOT NULL DEFAULT 0,
    avg_accuracy REAL NOT NULL DEFAULT 0
);

-- Backfill averages for all seasons that existed before this migration. Uses the
-- same NULLIF(x, 0) exclusion as ApplyLeagueAvgAttributes so inactive/FA-pool
-- players with all-zero attributes do not drag down the league mean.
INSERT INTO season_attribute_averages
    (season_id, avg_power, avg_contact, avg_speed, avg_fielding,
     avg_arm, avg_velocity, avg_junk, avg_accuracy)
SELECT
    ps.season_id,
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
GROUP BY ps.season_id;
