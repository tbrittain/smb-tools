-- Initial companion DB schema.
--
-- Design principles (see docs/architecture/data-layer.md):
--   • Store raw counting stats only. Simple rate stats (BA, OBP, ERA, etc.)
--     are computed on read via views, never as independent writable columns.
--   • No franchise_id anywhere — this is a per-franchise database.
--   • Players and teams are tracked across seasons by their save game GUID.
--     game_guid columns hold the hex-encoded GUID from the SMB save file.
--   • Context-dependent stats (OPS+, ERA+, FIP, wOBA, smbWAR) require
--     league-wide constants and are NOT included here; they are computed
--     during import and stored separately (future phase).

-- ── Snapshots ─────────────────────────────────────────────────────────────────

CREATE TABLE save_game_snapshots (
    id                    INTEGER PRIMARY KEY NOT NULL,
    season_num            INTEGER NOT NULL,
    captured_at           DATETIME NOT NULL DEFAULT (datetime('now')),
    file_name             TEXT NOT NULL,
    sha256_hash           TEXT NOT NULL,
    file_size_bytes       INTEGER NOT NULL,
    compressed            INTEGER NOT NULL DEFAULT 0,
    compressed_size_bytes INTEGER
);

-- ── Seasons ───────────────────────────────────────────────────────────────────

CREATE TABLE seasons (
    id              INTEGER PRIMARY KEY NOT NULL, -- save game seasonID
    season_num      INTEGER NOT NULL,
    num_games       INTEGER NOT NULL DEFAULT 0,   -- regular season length
    imported_at     DATETIME NOT NULL DEFAULT (datetime('now'))
);

-- ── Teams ─────────────────────────────────────────────────────────────────────

CREATE TABLE teams (
    id        INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    game_guid TEXT    NOT NULL UNIQUE -- hex GUID from t_teams
);

CREATE TABLE team_season_history (
    id               INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    team_id          INTEGER NOT NULL REFERENCES teams(id),
    season_id        INTEGER NOT NULL REFERENCES seasons(id),
    team_name        TEXT    NOT NULL,
    division_name    TEXT    NOT NULL DEFAULT '',
    conference_name  TEXT    NOT NULL DEFAULT '',
    -- financials
    budget           INTEGER NOT NULL DEFAULT 0,
    payroll          INTEGER NOT NULL DEFAULT 0,
    -- standings
    wins             INTEGER NOT NULL DEFAULT 0,
    losses           INTEGER NOT NULL DEFAULT 0,
    games_back       REAL    NOT NULL DEFAULT 0.0,
    runs_for         INTEGER NOT NULL DEFAULT 0,
    runs_against     INTEGER NOT NULL DEFAULT 0,
    -- aggregate player attributes (sum of roster)
    total_power      INTEGER NOT NULL DEFAULT 0,
    total_contact    INTEGER NOT NULL DEFAULT 0,
    total_speed      INTEGER NOT NULL DEFAULT 0,
    total_fielding   INTEGER NOT NULL DEFAULT 0,
    total_arm        INTEGER NOT NULL DEFAULT 0,
    total_velocity   INTEGER NOT NULL DEFAULT 0,
    total_junk       INTEGER NOT NULL DEFAULT 0,
    total_accuracy   INTEGER NOT NULL DEFAULT 0,
    -- playoff results (NULL if team did not make playoffs)
    playoff_seed     INTEGER,
    playoff_wins     INTEGER,
    playoff_losses   INTEGER,
    playoff_runs_for INTEGER,
    playoff_runs_against INTEGER,
    UNIQUE(team_id, season_id)
);

-- ── Players ───────────────────────────────────────────────────────────────────

CREATE TABLE players (
    id             INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    game_guid      TEXT    NOT NULL UNIQUE, -- hex GUID from t_baseball_players
    first_name     TEXT    NOT NULL,
    last_name      TEXT    NOT NULL,
    is_hall_of_famer INTEGER NOT NULL DEFAULT 0
);

-- One record per player per season. Captures the snapshot at import time.
CREATE TABLE player_seasons (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    player_id           INTEGER NOT NULL REFERENCES players(id),
    season_id           INTEGER NOT NULL REFERENCES seasons(id),
    team_history_id     INTEGER REFERENCES team_season_history(id),
    age                 INTEGER NOT NULL DEFAULT 0,
    salary              INTEGER NOT NULL DEFAULT 0, -- display units (game × 200)
    primary_position    TEXT    NOT NULL DEFAULT '',
    secondary_position  TEXT    NOT NULL DEFAULT '',
    pitcher_role        TEXT    NOT NULL DEFAULT '',
    bat_hand            TEXT    NOT NULL DEFAULT '',
    throw_hand          TEXT    NOT NULL DEFAULT '',
    chemistry_type      TEXT    NOT NULL DEFAULT '',
    -- Raw JSON arrays from the save game. Resolved to names by the UI layer.
    traits_json         TEXT    NOT NULL DEFAULT '[]',
    pitches_json        TEXT    NOT NULL DEFAULT '[]',
    UNIQUE(player_id, season_id)
);

-- Game attributes (Power, Contact, etc.) — one row per player_season.
CREATE TABLE player_season_game_stats (
    id               INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    player_season_id INTEGER NOT NULL UNIQUE REFERENCES player_seasons(id),
    power            INTEGER NOT NULL DEFAULT 0,
    contact          INTEGER NOT NULL DEFAULT 0,
    speed            INTEGER NOT NULL DEFAULT 0,
    fielding         INTEGER NOT NULL DEFAULT 0,
    arm              INTEGER NOT NULL DEFAULT 0,
    velocity         INTEGER NOT NULL DEFAULT 0,
    junk             INTEGER NOT NULL DEFAULT 0,
    accuracy         INTEGER NOT NULL DEFAULT 0
);

-- Raw batting counting stats. is_regular_season=1 for regular season, 0 for playoffs.
CREATE TABLE player_season_batting_stats (
    id               INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    player_season_id INTEGER NOT NULL REFERENCES player_seasons(id),
    is_regular_season INTEGER NOT NULL DEFAULT 1,
    games_played     INTEGER NOT NULL DEFAULT 0,
    games_batting    INTEGER NOT NULL DEFAULT 0,
    at_bats          INTEGER NOT NULL DEFAULT 0,
    runs             INTEGER NOT NULL DEFAULT 0,
    hits             INTEGER NOT NULL DEFAULT 0,
    doubles          INTEGER NOT NULL DEFAULT 0,
    triples          INTEGER NOT NULL DEFAULT 0,
    home_runs        INTEGER NOT NULL DEFAULT 0,
    rbi              INTEGER NOT NULL DEFAULT 0,
    stolen_bases     INTEGER NOT NULL DEFAULT 0,
    caught_stealing  INTEGER NOT NULL DEFAULT 0,
    walks            INTEGER NOT NULL DEFAULT 0,
    strikeouts       INTEGER NOT NULL DEFAULT 0,
    hit_by_pitch     INTEGER NOT NULL DEFAULT 0,
    sac_hits         INTEGER NOT NULL DEFAULT 0,
    sac_flies        INTEGER NOT NULL DEFAULT 0,
    errors           INTEGER NOT NULL DEFAULT 0,
    passed_balls     INTEGER NOT NULL DEFAULT 0,
    UNIQUE(player_season_id, is_regular_season)
);

-- Raw pitching counting stats. outs_pitched / 3 = innings pitched display value.
CREATE TABLE player_season_pitching_stats (
    id               INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    player_season_id INTEGER NOT NULL REFERENCES player_seasons(id),
    is_regular_season INTEGER NOT NULL DEFAULT 1,
    wins             INTEGER NOT NULL DEFAULT 0,
    losses           INTEGER NOT NULL DEFAULT 0,
    games            INTEGER NOT NULL DEFAULT 0,
    games_started    INTEGER NOT NULL DEFAULT 0,
    complete_games   INTEGER NOT NULL DEFAULT 0,
    shutouts         INTEGER NOT NULL DEFAULT 0,
    saves            INTEGER NOT NULL DEFAULT 0,
    outs_pitched     INTEGER NOT NULL DEFAULT 0,
    hits_allowed     INTEGER NOT NULL DEFAULT 0,
    earned_runs      INTEGER NOT NULL DEFAULT 0,
    home_runs_allowed INTEGER NOT NULL DEFAULT 0,
    walks            INTEGER NOT NULL DEFAULT 0,
    strikeouts       INTEGER NOT NULL DEFAULT 0,
    hit_batters      INTEGER NOT NULL DEFAULT 0,
    batters_faced    INTEGER NOT NULL DEFAULT 0,
    games_finished   INTEGER NOT NULL DEFAULT 0,
    runs_allowed     INTEGER NOT NULL DEFAULT 0,
    wild_pitches     INTEGER NOT NULL DEFAULT 0,
    total_pitches    INTEGER NOT NULL DEFAULT 0,
    UNIQUE(player_season_id, is_regular_season)
);

-- ── Schedules ─────────────────────────────────────────────────────────────────

CREATE TABLE team_season_schedules (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    season_id             INTEGER NOT NULL REFERENCES seasons(id),
    game_number           INTEGER NOT NULL,
    day                   INTEGER NOT NULL DEFAULT 0,
    home_team_history_id  INTEGER NOT NULL REFERENCES team_season_history(id),
    away_team_history_id  INTEGER NOT NULL REFERENCES team_season_history(id),
    home_pitcher_season_id INTEGER REFERENCES player_seasons(id),
    away_pitcher_season_id INTEGER REFERENCES player_seasons(id),
    home_score            INTEGER, -- NULL if game not yet played
    away_score            INTEGER,
    UNIQUE(season_id, game_number)
);

CREATE TABLE team_playoff_schedules (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    season_id             INTEGER NOT NULL REFERENCES seasons(id),
    series_number         INTEGER NOT NULL,
    game_number           INTEGER NOT NULL,
    home_team_history_id  INTEGER NOT NULL REFERENCES team_season_history(id),
    away_team_history_id  INTEGER NOT NULL REFERENCES team_season_history(id),
    home_pitcher_season_id INTEGER REFERENCES player_seasons(id),
    away_pitcher_season_id INTEGER REFERENCES player_seasons(id),
    home_score            INTEGER,
    away_score            INTEGER
);

-- ── Indexes ───────────────────────────────────────────────────────────────────

CREATE INDEX idx_player_seasons_season     ON player_seasons(season_id);
CREATE INDEX idx_player_seasons_player     ON player_seasons(player_id);
CREATE INDEX idx_team_history_season       ON team_season_history(season_id);
CREATE INDEX idx_team_history_team         ON team_season_history(team_id);
CREATE INDEX idx_batting_stats_season      ON player_season_batting_stats(player_season_id);
CREATE INDEX idx_pitching_stats_season     ON player_season_pitching_stats(player_season_id);
CREATE INDEX idx_schedules_season          ON team_season_schedules(season_id);
CREATE INDEX idx_playoff_schedules_season  ON team_playoff_schedules(season_id);

-- ── Rate stat views ───────────────────────────────────────────────────────────
--
-- Simple rate stats (BA, OBP, ERA, WHIP, etc.) are computed on read via
-- these views. Queries that need rate stats JOIN through here rather than
-- storing derived values alongside the raw counts they depend on.
-- Context-dependent stats (OPS+, ERA+, FIP, wOBA) are excluded.

CREATE VIEW v_batting_stats AS
SELECT
    b.*,

    -- Counting derivations
    (b.at_bats + b.walks + b.hit_by_pitch + b.sac_hits + b.sac_flies)
        AS plate_appearances,

    (b.hits - b.doubles - b.triples - b.home_runs
     + (b.doubles   * 2)
     + (b.triples   * 3)
     + (b.home_runs * 4))
        AS total_bases,

    -- Rate stats (NULL when denominator is zero)
    CASE WHEN b.at_bats > 0
         THEN CAST(b.hits AS REAL) / b.at_bats
         ELSE NULL END AS ba,

    CASE WHEN (b.at_bats + b.walks + b.hit_by_pitch + b.sac_flies) > 0
         THEN CAST(b.hits + b.walks + b.hit_by_pitch AS REAL)
              / (b.at_bats + b.walks + b.hit_by_pitch + b.sac_flies)
         ELSE NULL END AS obp,

    CASE WHEN b.at_bats > 0
         THEN CAST(
                  b.hits - b.doubles - b.triples - b.home_runs
                  + (b.doubles   * 2)
                  + (b.triples   * 3)
                  + (b.home_runs * 4)
              AS REAL) / b.at_bats
         ELSE NULL END AS slg,

    -- OPS: OBP + SLG (both sub-expressions inlined to avoid CTEs in a view)
    CASE WHEN b.at_bats > 0
              AND (b.at_bats + b.walks + b.hit_by_pitch + b.sac_flies) > 0
         THEN (
             CAST(b.hits + b.walks + b.hit_by_pitch AS REAL)
             / (b.at_bats + b.walks + b.hit_by_pitch + b.sac_flies)
             + CAST(
                   b.hits - b.doubles - b.triples - b.home_runs
                   + (b.doubles   * 2)
                   + (b.triples   * 3)
                   + (b.home_runs * 4)
               AS REAL) / b.at_bats
         )
         ELSE NULL END AS ops,

    -- ISO = SLG − BA
    CASE WHEN b.at_bats > 0
         THEN CAST(
                  b.hits - b.doubles - b.triples - b.home_runs
                  + (b.doubles   * 2)
                  + (b.triples   * 3)
                  + (b.home_runs * 4)
              AS REAL) / b.at_bats
              - CAST(b.hits AS REAL) / b.at_bats
         ELSE NULL END AS iso,

    -- BABIP = (H − HR) / (AB − K − HR + SF)
    CASE WHEN (b.at_bats - b.strikeouts - b.home_runs + b.sac_flies) > 0
         THEN CAST(b.hits - b.home_runs AS REAL)
              / (b.at_bats - b.strikeouts - b.home_runs + b.sac_flies)
         ELSE NULL END AS babip,

    CASE WHEN (b.at_bats + b.walks + b.hit_by_pitch + b.sac_hits + b.sac_flies) > 0
         THEN CAST(b.strikeouts AS REAL)
              / (b.at_bats + b.walks + b.hit_by_pitch + b.sac_hits + b.sac_flies)
         ELSE NULL END AS k_pct,

    CASE WHEN (b.at_bats + b.walks + b.hit_by_pitch + b.sac_hits + b.sac_flies) > 0
         THEN CAST(b.walks AS REAL)
              / (b.at_bats + b.walks + b.hit_by_pitch + b.sac_hits + b.sac_flies)
         ELSE NULL END AS bb_pct,

    CASE WHEN b.home_runs > 0
         THEN CAST(b.at_bats AS REAL) / b.home_runs
         ELSE NULL END AS ab_per_hr

FROM player_season_batting_stats b;

CREATE VIEW v_pitching_stats AS
SELECT
    p.*,

    -- IP display: e.g. 97 outs → "32.1", 99 outs → "33.0", 100 outs → "33.1"
    -- The fractional digit is remainder outs (0, 1, or 2), not tenths of an inning.
    CAST(p.outs_pitched / 3 AS TEXT) || '.' || CAST(p.outs_pitched % 3 AS TEXT)
        AS ip_display,

    -- Rate stats (NULL when no IP recorded)
    CASE WHEN p.outs_pitched > 0
         THEN CAST(p.earned_runs AS REAL) * 27.0 / p.outs_pitched
         ELSE NULL END AS era,

    CASE WHEN p.outs_pitched > 0
         THEN CAST(p.walks + p.hits_allowed AS REAL) * 3.0 / p.outs_pitched
         ELSE NULL END AS whip,

    CASE WHEN p.outs_pitched > 0
         THEN CAST(p.strikeouts AS REAL) * 27.0 / p.outs_pitched
         ELSE NULL END AS k_per_9,

    CASE WHEN p.outs_pitched > 0
         THEN CAST(p.walks AS REAL) * 27.0 / p.outs_pitched
         ELSE NULL END AS bb_per_9,

    CASE WHEN p.outs_pitched > 0
         THEN CAST(p.hits_allowed AS REAL) * 27.0 / p.outs_pitched
         ELSE NULL END AS h_per_9,

    CASE WHEN p.outs_pitched > 0
         THEN CAST(p.home_runs_allowed AS REAL) * 27.0 / p.outs_pitched
         ELSE NULL END AS hr_per_9,

    CASE WHEN p.walks > 0
         THEN CAST(p.strikeouts AS REAL) / p.walks
         ELSE NULL END AS k_per_bb,

    CASE WHEN p.batters_faced > 0
         THEN CAST(p.strikeouts AS REAL) / p.batters_faced
         ELSE NULL END AS k_pct,

    CASE WHEN (p.wins + p.losses) > 0
         THEN CAST(p.wins AS REAL) / (p.wins + p.losses)
         ELSE NULL END AS win_pct,

    CASE WHEN p.outs_pitched > 0
         THEN CAST(p.total_pitches AS REAL) * 3.0 / p.outs_pitched
         ELSE NULL END AS p_per_ip

FROM player_season_pitching_stats p;
