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
