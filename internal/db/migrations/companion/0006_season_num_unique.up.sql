-- Enforce one season_num per franchise companion DB.
-- This allows a live save game sync to supersede a legacy-imported season
-- for the same franchise season number: the upsert fires on season_num,
-- updates league_guid + save_game_season_id in place, and all child records
-- (player_seasons, schedule rows, etc.) remain valid via the unchanged seasons.id.
CREATE UNIQUE INDEX idx_seasons_season_num ON seasons(season_num);
