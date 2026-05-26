-- Migration 0005: championship award views and seed records.
--
-- Introduces two SQLite views that encapsulate champion/runner-up detection,
-- replacing the championCTE Go constant that was previously duplicated across
-- multiple queries. Both views inherit the completeness gate: no row is returned
-- if any playoff game for the season lacks a recorded score.

-- season_champions returns one row per season that has a determinable champion.
-- winner_history_id references team_season_history.id.
CREATE VIEW season_champions AS
WITH complete_playoff_seasons AS (
    -- Only seasons where every scheduled playoff game has a score recorded.
    -- Seasons with any NULL score are excluded entirely.
    SELECT season_id
    FROM team_playoff_schedules
    GROUP BY season_id
    HAVING MIN(CASE WHEN home_score IS NOT NULL AND away_score IS NOT NULL
                    THEN 1 ELSE 0 END) = 1
),
final_series AS (
    SELECT season_id, MAX(series_number) AS max_series
    FROM team_playoff_schedules
    WHERE season_id IN (SELECT season_id FROM complete_playoff_seasons)
    GROUP BY season_id
),
series_wins AS (
    SELECT
        g.season_id,
        CASE WHEN g.home_score > g.away_score
             THEN g.home_team_history_id
             ELSE g.away_team_history_id
        END AS winner_history_id,
        COUNT(*) AS wins
    FROM team_playoff_schedules g
    JOIN final_series fs
        ON g.season_id = fs.season_id AND g.series_number = fs.max_series
    GROUP BY g.season_id,
             CASE WHEN g.home_score > g.away_score
                  THEN g.home_team_history_id
                  ELSE g.away_team_history_id END
),
series_totals AS (
    SELECT season_id, SUM(wins) AS total_games
    FROM series_wins
    GROUP BY season_id
),
champion AS (
    SELECT sw.season_id, sw.winner_history_id
    FROM series_wins sw
    JOIN series_totals st ON st.season_id = sw.season_id
    -- Guard: winner must have strictly more than half the games played.
    WHERE sw.wins * 2 > st.total_games
      AND sw.wins = (
          SELECT MAX(wins) FROM series_wins sw2 WHERE sw2.season_id = sw.season_id
      )
)
SELECT winner_history_id, season_id FROM champion;

-- season_conference_champions returns the runner-up for each season: the team
-- that appeared in the final series against the champion but did not win.
-- Inherits the completeness gate via the JOIN to season_champions.
CREATE VIEW season_conference_champions AS
SELECT DISTINCT
    CASE
        WHEN tps.home_team_history_id = sc.winner_history_id
            THEN tps.away_team_history_id
            ELSE tps.home_team_history_id
    END AS runner_up_history_id,
    sc.season_id
FROM team_playoff_schedules tps
JOIN season_champions sc
    ON sc.season_id = tps.season_id
   AND (tps.home_team_history_id = sc.winner_history_id
        OR tps.away_team_history_id = sc.winner_history_id)
WHERE tps.series_number = (
    SELECT MAX(series_number)
    FROM team_playoff_schedules t2
    WHERE t2.season_id = tps.season_id
);

-- Seed the two championship awards.
-- is_user_assignable = 0: computed automatically, never offered in the award
--   delegation UI even though is_playoff_award = 1.
-- League Champion: groupable (omit_from_groupings = 0) — shows "Nx League Champion"
--   in career summary badges.
-- Conference Champion: non-groupable (omit_from_groupings = 1) — appears in
--   per-season award rows only, not in the career summary.
INSERT INTO awards
    (name, original_name, importance, omit_from_groupings,
     is_batting_award, is_pitching_award, is_fielding_award,
     is_playoff_award, is_user_assignable, is_built_in)
VALUES
    ('League Champion',    'League Champion',    1, 0, 0, 0, 0, 1, 0, 1),
    ('Conference Champion','Conference Champion', 1, 1, 0, 0, 0, 1, 0, 1);
