ALTER TABLE franchises ADD COLUMN league_mode TEXT NOT NULL DEFAULT 'franchise'
    CHECK (league_mode IN ('franchise', 'season'));
