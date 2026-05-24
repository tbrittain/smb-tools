-- Replace single team_history_id on player_seasons with a normalized junction table.
-- Existing single-team associations are migrated; re-import required to populate
-- multi-team (traded) and FA-fallback data for prior seasons.

CREATE TABLE player_season_teams (
    id               INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    player_season_id INTEGER NOT NULL REFERENCES player_seasons(id) ON DELETE CASCADE,
    team_history_id  INTEGER NOT NULL REFERENCES team_season_history(id),
    sort_order       INTEGER NOT NULL DEFAULT 0, -- 0=current/final, 1=most recent prior, 2=previous
    UNIQUE(player_season_id, team_history_id)
);

-- Migrate any existing single-team associations.
INSERT OR IGNORE INTO player_season_teams (player_season_id, team_history_id, sort_order)
SELECT id, team_history_id, 0
FROM player_seasons
WHERE team_history_id IS NOT NULL;

CREATE INDEX idx_player_season_teams ON player_season_teams(player_season_id);

-- Recreate player_seasons without team_history_id.
-- IDs are preserved so FK references in batting/pitching/game stats remain valid.
CREATE TABLE player_seasons_new (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    player_id           INTEGER NOT NULL REFERENCES players(id),
    season_id           INTEGER NOT NULL REFERENCES seasons(id),
    age                 INTEGER NOT NULL DEFAULT 0,
    salary              INTEGER NOT NULL DEFAULT 0,
    primary_position    TEXT    NOT NULL DEFAULT '',
    secondary_position  TEXT    NOT NULL DEFAULT '',
    pitcher_role        TEXT    NOT NULL DEFAULT '',
    bat_hand            TEXT    NOT NULL DEFAULT '',
    throw_hand          TEXT    NOT NULL DEFAULT '',
    chemistry_type      TEXT    NOT NULL DEFAULT '',
    traits_json         TEXT    NOT NULL DEFAULT '[]',
    pitches_json        TEXT    NOT NULL DEFAULT '[]',
    UNIQUE(player_id, season_id)
);

INSERT INTO player_seasons_new
    SELECT id, player_id, season_id, age, salary,
           primary_position, secondary_position, pitcher_role,
           bat_hand, throw_hand, chemistry_type, traits_json, pitches_json
    FROM player_seasons;

DROP TABLE player_seasons;
ALTER TABLE player_seasons_new RENAME TO player_seasons;

CREATE INDEX idx_player_seasons_season ON player_seasons(season_id);
CREATE INDEX idx_player_seasons_player ON player_seasons(player_id);
