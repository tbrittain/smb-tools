-- Context-dependent stat columns and league aggregation table.
--
-- OPS+, ERA+, FIP, FIP-, and smbWAR require league-wide averages that only
-- exist after a full season is imported. They are computed during ImportSeason
-- and stored here so queries can sort and filter without re-computing at read
-- time. They are NOT SQL generated columns because generated columns cannot
-- depend on other rows (league context).
--
-- Columns are nullable: NULL means the season predates this migration and
-- has not been re-synced yet.

ALTER TABLE player_season_batting_stats ADD COLUMN ops_plus REAL;
ALTER TABLE player_season_batting_stats ADD COLUMN smb_war  REAL;

ALTER TABLE player_season_pitching_stats ADD COLUMN era_plus  REAL;
ALTER TABLE player_season_pitching_stats ADD COLUMN fip       REAL;
ALTER TABLE player_season_pitching_stats ADD COLUMN fip_minus REAL;
ALTER TABLE player_season_pitching_stats ADD COLUMN smb_war   REAL;

-- League-level batting and pitching aggregates for each season / game type.
-- Stored so that future re-computation or auditing does not require replaying
-- all player stat rows.
CREATE TABLE league_season_stats (
    id                      INTEGER PRIMARY KEY,
    season_id               INTEGER NOT NULL REFERENCES seasons(id),
    is_regular_season       INTEGER NOT NULL DEFAULT 1,
    -- Raw batting aggregates (SUM over all player_season_batting_stats rows)
    total_plate_appearances INTEGER NOT NULL DEFAULT 0,
    total_at_bats           INTEGER NOT NULL DEFAULT 0,
    total_hits              INTEGER NOT NULL DEFAULT 0,
    total_doubles           INTEGER NOT NULL DEFAULT 0,
    total_triples           INTEGER NOT NULL DEFAULT 0,
    total_home_runs         INTEGER NOT NULL DEFAULT 0,
    total_walks             INTEGER NOT NULL DEFAULT 0,
    total_hbp               INTEGER NOT NULL DEFAULT 0,
    total_sac_flies         INTEGER NOT NULL DEFAULT 0,
    -- Raw pitching aggregates
    total_outs_pitched      INTEGER NOT NULL DEFAULT 0,
    total_earned_runs       INTEGER NOT NULL DEFAULT 0,
    total_hr_allowed        INTEGER NOT NULL DEFAULT 0,
    total_bb_allowed        INTEGER NOT NULL DEFAULT 0,
    total_hbp_allowed       INTEGER NOT NULL DEFAULT 0,
    total_k_pitched         INTEGER NOT NULL DEFAULT 0,
    -- Derived constants computed and stored at import time
    lg_obp       REAL,
    lg_slg       REAL,
    lg_era       REAL,
    fip_constant REAL,
    UNIQUE(season_id, is_regular_season)
);
