CREATE TABLE franchises (
    id                   TEXT PRIMARY KEY NOT NULL,
    name                 TEXT NOT NULL,
    game_version         TEXT NOT NULL CHECK (game_version IN ('smb3', 'smb4')),
    db_path              TEXT NOT NULL,
    created_at           DATETIME NOT NULL DEFAULT (datetime('now')),
    last_synced_at       DATETIME,
    last_synced_season   INTEGER
);

-- Each row represents one save game file + leagueGUID pair associated with a
-- franchise. A franchise starts with one source (season_offset = 0). When the
-- user exports the franchise to a new league in SMB4, a second source is added
-- with season_offset = last synced season number. SyncSeason always reads from
-- the source with the highest season_offset.
CREATE TABLE franchise_sources (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    franchise_id   TEXT    NOT NULL REFERENCES franchises(id),
    save_file_path TEXT    NOT NULL,
    league_guid    TEXT    NOT NULL,
    season_offset  INTEGER NOT NULL DEFAULT 0,
    added_at       DATETIME NOT NULL DEFAULT (datetime('now'))
);
