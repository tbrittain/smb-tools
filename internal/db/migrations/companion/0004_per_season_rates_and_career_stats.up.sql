-- Remove views that computed simple rate stats on read; rates are now stored as columns.
DROP VIEW IF EXISTS v_batting_stats;
DROP VIEW IF EXISTS v_pitching_stats;

-- ── Per-season batting rate columns ──────────────────────────────────────────
ALTER TABLE player_season_batting_stats ADD COLUMN ba       REAL;
ALTER TABLE player_season_batting_stats ADD COLUMN obp      REAL;
ALTER TABLE player_season_batting_stats ADD COLUMN slg      REAL;
ALTER TABLE player_season_batting_stats ADD COLUMN ops      REAL;
ALTER TABLE player_season_batting_stats ADD COLUMN iso      REAL;
ALTER TABLE player_season_batting_stats ADD COLUMN babip    REAL;
ALTER TABLE player_season_batting_stats ADD COLUMN k_pct    REAL;
ALTER TABLE player_season_batting_stats ADD COLUMN bb_pct   REAL;
ALTER TABLE player_season_batting_stats ADD COLUMN ab_per_hr REAL;

-- ── Per-season pitching rate columns ─────────────────────────────────────────
ALTER TABLE player_season_pitching_stats ADD COLUMN era      REAL;
ALTER TABLE player_season_pitching_stats ADD COLUMN whip     REAL;
ALTER TABLE player_season_pitching_stats ADD COLUMN k_per_9  REAL;
ALTER TABLE player_season_pitching_stats ADD COLUMN bb_per_9 REAL;
ALTER TABLE player_season_pitching_stats ADD COLUMN h_per_9  REAL;
ALTER TABLE player_season_pitching_stats ADD COLUMN hr_per_9 REAL;
ALTER TABLE player_season_pitching_stats ADD COLUMN k_per_bb REAL;
ALTER TABLE player_season_pitching_stats ADD COLUMN k_pct    REAL;
ALTER TABLE player_season_pitching_stats ADD COLUMN win_pct  REAL;
ALTER TABLE player_season_pitching_stats ADD COLUMN p_per_ip REAL;

-- ── Career batting stats ──────────────────────────────────────────────────────
-- stat_type discriminates between career regular-season, playoff, and combined totals.
-- 'total_career' aggregates both regular and playoff counting stats and recomputes
-- all rates from the combined numerators/denominators.
CREATE TABLE player_career_batting_stats (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    player_id      INTEGER NOT NULL REFERENCES players(id),
    stat_type      TEXT    NOT NULL
                     CHECK(stat_type IN ('regular_season', 'playoffs', 'total_career')),
    seasons_played INTEGER NOT NULL DEFAULT 0,
    -- counting stats
    games_played   INTEGER NOT NULL DEFAULT 0,
    games_batting  INTEGER NOT NULL DEFAULT 0,
    at_bats        INTEGER NOT NULL DEFAULT 0,
    runs           INTEGER NOT NULL DEFAULT 0,
    hits           INTEGER NOT NULL DEFAULT 0,
    doubles        INTEGER NOT NULL DEFAULT 0,
    triples        INTEGER NOT NULL DEFAULT 0,
    home_runs      INTEGER NOT NULL DEFAULT 0,
    rbi            INTEGER NOT NULL DEFAULT 0,
    stolen_bases   INTEGER NOT NULL DEFAULT 0,
    caught_stealing INTEGER NOT NULL DEFAULT 0,
    walks          INTEGER NOT NULL DEFAULT 0,
    strikeouts     INTEGER NOT NULL DEFAULT 0,
    hit_by_pitch   INTEGER NOT NULL DEFAULT 0,
    sac_hits       INTEGER NOT NULL DEFAULT 0,
    sac_flies      INTEGER NOT NULL DEFAULT 0,
    errors         INTEGER NOT NULL DEFAULT 0,
    passed_balls   INTEGER NOT NULL DEFAULT 0,
    -- simple rate stats (NULL when denominator is zero)
    ba       REAL,
    obp      REAL,
    slg      REAL,
    ops      REAL,
    iso      REAL,
    babip    REAL,
    k_pct    REAL,
    bb_pct   REAL,
    ab_per_hr REAL,
    -- context-dependent stats (NULL when no league data available)
    ops_plus REAL,
    smb_war  REAL,
    UNIQUE(player_id, stat_type)
);

-- ── Career pitching stats ─────────────────────────────────────────────────────
CREATE TABLE player_career_pitching_stats (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    player_id         INTEGER NOT NULL REFERENCES players(id),
    stat_type         TEXT    NOT NULL
                        CHECK(stat_type IN ('regular_season', 'playoffs', 'total_career')),
    seasons_played    INTEGER NOT NULL DEFAULT 0,
    -- counting stats
    wins              INTEGER NOT NULL DEFAULT 0,
    losses            INTEGER NOT NULL DEFAULT 0,
    games             INTEGER NOT NULL DEFAULT 0,
    games_started     INTEGER NOT NULL DEFAULT 0,
    complete_games    INTEGER NOT NULL DEFAULT 0,
    shutouts          INTEGER NOT NULL DEFAULT 0,
    saves             INTEGER NOT NULL DEFAULT 0,
    outs_pitched      INTEGER NOT NULL DEFAULT 0,
    hits_allowed      INTEGER NOT NULL DEFAULT 0,
    earned_runs       INTEGER NOT NULL DEFAULT 0,
    home_runs_allowed INTEGER NOT NULL DEFAULT 0,
    walks             INTEGER NOT NULL DEFAULT 0,
    strikeouts        INTEGER NOT NULL DEFAULT 0,
    hit_batters       INTEGER NOT NULL DEFAULT 0,
    batters_faced     INTEGER NOT NULL DEFAULT 0,
    games_finished    INTEGER NOT NULL DEFAULT 0,
    runs_allowed      INTEGER NOT NULL DEFAULT 0,
    wild_pitches      INTEGER NOT NULL DEFAULT 0,
    total_pitches     INTEGER NOT NULL DEFAULT 0,
    -- simple rate stats (NULL when denominator is zero)
    era      REAL,
    whip     REAL,
    k_per_9  REAL,
    bb_per_9 REAL,
    h_per_9  REAL,
    hr_per_9 REAL,
    k_per_bb REAL,
    k_pct    REAL,
    win_pct  REAL,
    p_per_ip REAL,
    -- context-dependent stats (NULL when no league data available)
    era_plus  REAL,
    fip       REAL,
    fip_minus REAL,
    smb_war   REAL,
    UNIQUE(player_id, stat_type)
);
