-- Rate stat views for the companion DB.
--
-- Simple rate stats (BA, OBP, ERA, WHIP, etc.) are deterministic functions of
-- stored counting columns. Rather than storing them independently (where they
-- can diverge from inputs on re-import) or relying on ALTER TABLE ADD COLUMN
-- VIRTUAL (not portable across SQLite versions), these views compute rates on
-- read. Queries needing rate stats JOIN through v_batting_stats or
-- v_pitching_stats instead of the raw tables.
--
-- Context-dependent stats (OPS+, ERA+, FIP, wOBA, smbWAR) require league-wide
-- constants available only at import time and are NOT included here.

-- ── Batting ───────────────────────────────────────────────────────────────────

CREATE VIEW v_batting_stats AS
SELECT
    b.*,

    -- Counting derivations
    (b.at_bats + b.walks + b.hit_by_pitch + b.sac_hits + b.sac_flies)
        AS plate_appearances,

    (b.hits - b.doubles - b.triples - b.home_runs
     + (b.doubles   * 2)
     + (b.triples   * 3)
     + (b.home_runs * 4))
        AS total_bases,

    -- Rate stats (NULL when denominator is zero)
    CASE WHEN b.at_bats > 0
         THEN CAST(b.hits AS REAL) / b.at_bats
         ELSE NULL END AS ba,

    CASE WHEN (b.at_bats + b.walks + b.hit_by_pitch + b.sac_flies) > 0
         THEN CAST(b.hits + b.walks + b.hit_by_pitch AS REAL)
              / (b.at_bats + b.walks + b.hit_by_pitch + b.sac_flies)
         ELSE NULL END AS obp,

    CASE WHEN b.at_bats > 0
         THEN CAST(
                  b.hits - b.doubles - b.triples - b.home_runs
                  + (b.doubles   * 2)
                  + (b.triples   * 3)
                  + (b.home_runs * 4)
              AS REAL) / b.at_bats
         ELSE NULL END AS slg,

    -- OPS: OBP + SLG (both sub-expressions inlined to avoid CTEs in a view)
    CASE WHEN b.at_bats > 0
              AND (b.at_bats + b.walks + b.hit_by_pitch + b.sac_flies) > 0
         THEN (
             CAST(b.hits + b.walks + b.hit_by_pitch AS REAL)
             / (b.at_bats + b.walks + b.hit_by_pitch + b.sac_flies)
             + CAST(
                   b.hits - b.doubles - b.triples - b.home_runs
                   + (b.doubles   * 2)
                   + (b.triples   * 3)
                   + (b.home_runs * 4)
               AS REAL) / b.at_bats
         )
         ELSE NULL END AS ops,

    -- ISO = SLG − BA
    CASE WHEN b.at_bats > 0
         THEN CAST(
                  b.hits - b.doubles - b.triples - b.home_runs
                  + (b.doubles   * 2)
                  + (b.triples   * 3)
                  + (b.home_runs * 4)
              AS REAL) / b.at_bats
              - CAST(b.hits AS REAL) / b.at_bats
         ELSE NULL END AS iso,

    -- BABIP = (H − HR) / (AB − K − HR + SF)
    CASE WHEN (b.at_bats - b.strikeouts - b.home_runs + b.sac_flies) > 0
         THEN CAST(b.hits - b.home_runs AS REAL)
              / (b.at_bats - b.strikeouts - b.home_runs + b.sac_flies)
         ELSE NULL END AS babip,

    CASE WHEN (b.at_bats + b.walks + b.hit_by_pitch + b.sac_hits + b.sac_flies) > 0
         THEN CAST(b.strikeouts AS REAL)
              / (b.at_bats + b.walks + b.hit_by_pitch + b.sac_hits + b.sac_flies)
         ELSE NULL END AS k_pct,

    CASE WHEN (b.at_bats + b.walks + b.hit_by_pitch + b.sac_hits + b.sac_flies) > 0
         THEN CAST(b.walks AS REAL)
              / (b.at_bats + b.walks + b.hit_by_pitch + b.sac_hits + b.sac_flies)
         ELSE NULL END AS bb_pct,

    CASE WHEN b.home_runs > 0
         THEN CAST(b.at_bats AS REAL) / b.home_runs
         ELSE NULL END AS ab_per_hr

FROM player_season_batting_stats b;

-- ── Pitching ──────────────────────────────────────────────────────────────────

CREATE VIEW v_pitching_stats AS
SELECT
    p.*,

    -- IP display: e.g. 97 outs → "32.1", 99 outs → "33.0", 100 outs → "33.1"
    -- The fractional digit is remainder outs (0, 1, or 2), not tenths of an inning.
    CAST(p.outs_pitched / 3 AS TEXT) || '.' || CAST(p.outs_pitched % 3 AS TEXT)
        AS ip_display,

    -- Rate stats (NULL when no IP recorded)
    CASE WHEN p.outs_pitched > 0
         THEN CAST(p.earned_runs AS REAL) * 27.0 / p.outs_pitched
         ELSE NULL END AS era,

    CASE WHEN p.outs_pitched > 0
         THEN CAST(p.walks + p.hits_allowed AS REAL) * 3.0 / p.outs_pitched
         ELSE NULL END AS whip,

    CASE WHEN p.outs_pitched > 0
         THEN CAST(p.strikeouts AS REAL) * 27.0 / p.outs_pitched
         ELSE NULL END AS k_per_9,

    CASE WHEN p.outs_pitched > 0
         THEN CAST(p.walks AS REAL) * 27.0 / p.outs_pitched
         ELSE NULL END AS bb_per_9,

    CASE WHEN p.outs_pitched > 0
         THEN CAST(p.hits_allowed AS REAL) * 27.0 / p.outs_pitched
         ELSE NULL END AS h_per_9,

    CASE WHEN p.outs_pitched > 0
         THEN CAST(p.home_runs_allowed AS REAL) * 27.0 / p.outs_pitched
         ELSE NULL END AS hr_per_9,

    CASE WHEN p.walks > 0
         THEN CAST(p.strikeouts AS REAL) / p.walks
         ELSE NULL END AS k_per_bb,

    CASE WHEN p.batters_faced > 0
         THEN CAST(p.strikeouts AS REAL) / p.batters_faced
         ELSE NULL END AS k_pct,

    CASE WHEN (p.wins + p.losses) > 0
         THEN CAST(p.wins AS REAL) / (p.wins + p.losses)
         ELSE NULL END AS win_pct,

    CASE WHEN p.outs_pitched > 0
         THEN CAST(p.total_pitches AS REAL) * 3.0 / p.outs_pitched
         ELSE NULL END AS p_per_ip

FROM player_season_pitching_stats p;
