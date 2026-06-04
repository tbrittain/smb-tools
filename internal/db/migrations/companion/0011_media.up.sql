CREATE TABLE media (
    id          TEXT     PRIMARY KEY NOT NULL,
    file_path   TEXT     NOT NULL,
    media_type  TEXT     NOT NULL,
    name        TEXT     NOT NULL,
    description TEXT,
    uploaded_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE media_team_seasons (
    id              TEXT    PRIMARY KEY NOT NULL,
    media_id        TEXT    NOT NULL REFERENCES media(id),
    team_history_id INTEGER NOT NULL REFERENCES team_season_history(id),
    UNIQUE(media_id, team_history_id)
);

CREATE INDEX idx_media_ts_team_history ON media_team_seasons(team_history_id);
CREATE INDEX idx_media_ts_media_id     ON media_team_seasons(media_id);

CREATE TABLE media_players (
    id        TEXT    PRIMARY KEY NOT NULL,
    media_id  TEXT    NOT NULL REFERENCES media(id),
    player_id INTEGER NOT NULL REFERENCES players(id),
    UNIQUE(media_id, player_id)
);

CREATE INDEX idx_media_players_player ON media_players(player_id);
CREATE INDEX idx_media_players_media  ON media_players(media_id);
