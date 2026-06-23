-- Migration 0016: persist innings-per-game on seasons.
--
-- SMB4 supports variable-length games (t_seasons.innings in the save game, a
-- value between 1 and 9). Career qualification thresholds need this alongside
-- num_games to correctly scale 3000 PA / 1000 IP style minimums.
--
-- Nullable with no default: SMB4 games are commonly played at fewer than 9
-- innings, so we cannot assume any particular value. Existing seasons (synced
-- before this column existed) and legacy-migrated seasons land here as NULL —
-- the app must always supply a real value going forward, and existing NULL
-- rows are backfilled later via the Setup page once the user supplies the
-- actual game length they used.

ALTER TABLE seasons ADD COLUMN innings_per_game INTEGER;
