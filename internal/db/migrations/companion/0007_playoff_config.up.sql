-- Migration 0007: persist playoff bracket config and tighten champion detection.
--
-- Adds two NOT NULL columns to seasons (default 0 = "no playoff config yet"):
--   playoff_rounds        – number of single-elimination rounds (e.g. 3 for 8 teams)
--   playoff_series_length – max games per series (e.g. 7 for best-of-7)
--
-- Backfills existing rows from team_playoff_schedules where data is available.
-- For new imports these are read directly from t_playoffs in the save game;
-- for legacy companion imports they are inferred from imported game data.
--
-- The season_champions view is recreated with a stricter gate: playoff_rounds
-- must be > 0 (i.e. a known bracket) AND all 2^rounds - 1 expected series must
-- be present. This prevents mid-playoff imports from prematurely assigning a
-- champion while still working correctly for the legacy import path (which now
-- always has playoff_rounds set via inference).

ALTER TABLE seasons ADD COLUMN playoff_rounds        INTEGER NOT NULL DEFAULT 0;
ALTER TABLE seasons ADD COLUMN playoff_series_length INTEGER NOT NULL DEFAULT 0;

-- ── Backfill playoff_rounds for existing seasons ───────────────────────────────
-- Uses a CASE ladder for bits.Len(total_series) since SQLite has no log().
UPDATE seasons
SET playoff_rounds = (
    SELECT CASE
        WHEN total = 0  THEN 0
        WHEN total <= 1  THEN 1
        WHEN total <= 3  THEN 2
        WHEN total <= 7  THEN 3
        WHEN total <= 15 THEN 4
        WHEN total <= 31 THEN 5
        ELSE 6
    END
    FROM (
        SELECT COUNT(DISTINCT series_number) AS total
        FROM team_playoff_schedules
        WHERE season_id = seasons.id
    )
);

-- ── Backfill playoff_series_length for existing seasons ───────────────────────
-- Derives series length from the max number of wins any team accumulated in a
-- single series across all scored games: seriesLength = 2 * maxWins - 1.
UPDATE seasons
SET playoff_series_length = COALESCE(
    (
        SELECT CASE WHEN max_wins > 0 THEN 2 * max_wins - 1 ELSE 0 END
        FROM (
            SELECT MAX(team_wins) AS max_wins
            FROM (
                SELECT
                    series_number,
                    CASE WHEN home_score > away_score THEN home_team_history_id
                         ELSE away_team_history_id END AS winner_id,
                    COUNT(*) AS team_wins
                FROM team_playoff_schedules
                WHERE season_id = seasons.id
                  AND home_score IS NOT NULL AND away_score IS NOT NULL
                  AND home_score != away_score
                GROUP BY series_number,
                         CASE WHEN home_score > away_score THEN home_team_history_id
                              ELSE away_team_history_id END
            )
        )
    ),
    0
);

-- ── Recreate champion detection views ─────────────────────────────────────────
DROP VIEW season_conference_champions;
DROP VIEW season_champions;

CREATE VIEW season_champions AS
WITH complete_playoff_seasons AS (
    -- Only seasons where playoff_rounds is configured (> 0) and the full
    -- bracket has been played (structural check) with all games scored.
    SELECT tps.season_id
    FROM team_playoff_schedules tps
    JOIN seasons s ON s.id = tps.season_id
    WHERE s.playoff_rounds > 0
    GROUP BY tps.season_id, s.playoff_rounds
    HAVING
        MIN(CASE WHEN tps.home_score IS NOT NULL AND tps.away_score IS NOT NULL
                 THEN 1 ELSE 0 END) = 1
        AND COUNT(DISTINCT tps.series_number) = ((1 << s.playoff_rounds) - 1)
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

-- Recreate season_conference_champions (logic unchanged from migration 0005).
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
