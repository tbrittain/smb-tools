CREATE TABLE franchises (
    id                   TEXT PRIMARY KEY NOT NULL,
    name                 TEXT NOT NULL,
    game_version         TEXT NOT NULL CHECK (game_version IN ('smb3', 'smb4')),
    save_file_path       TEXT,
    league_guid          TEXT,
    db_path              TEXT NOT NULL,
    created_at           DATETIME NOT NULL DEFAULT (datetime('now')),
    last_synced_at       DATETIME,
    last_synced_season   INTEGER
);
