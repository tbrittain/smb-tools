CREATE TABLE logos (
    id          TEXT     PRIMARY KEY NOT NULL,
    team_id     INTEGER  NOT NULL,
    file_path   TEXT     NOT NULL,
    uploaded_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE logo_assignments (
    id           TEXT     PRIMARY KEY NOT NULL,
    logo_id      TEXT     NOT NULL REFERENCES logos(id),
    start_season INTEGER,
    end_season   INTEGER,
    assigned_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_logo_assignments_logo_id ON logo_assignments(logo_id);
