-- Migration 0016: persist innings-per-game on seasons.
--
-- SMB4 supports variable-length games (t_seasons.innings in the save game).
-- Career qualification thresholds need this alongside num_games to correctly
-- scale 3000 PA / 1000 IP style minimums. Defaults to 9 (the standard game
-- length) for existing rows and for legacy-migrated seasons, which have no
-- source data for this value.

ALTER TABLE seasons ADD COLUMN innings_per_game INTEGER NOT NULL DEFAULT 9;
