-- Baseball schema for the companion DB.
--
-- Design principles (see docs/architecture/data-layer.md):
--   • Store raw counting stats only. Simple rate stats (BA, OBP, ERA, etc.)
--     are deterministic functions of stored columns — compute at query time
--     or via generated columns, never as independent writable columns.
--   • No franchise_id anywhere — this is a per-franchise database.
--   • Players and teams are tracked across seasons by their save game GUID.
--     game_guid columns hold the hex-encoded GUID from the SMB save file.

-- ── Seasons ──────────────────────────────────────────────────────────────────

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
