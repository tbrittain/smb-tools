package store

import (
	"context"
	"fmt"
	"strings"

	"smb-tools/internal/models"
)

// LeaderboardQueryStore provides read-only queries for the four leaderboard types.
type LeaderboardQueryStore struct {
	db DBTX
}

func NewLeaderboardQueryStore(db DBTX) *LeaderboardQueryStore {
	return &LeaderboardQueryStore{db: db}
}

// whereFragment holds the extra WHERE conditions and their positional args for
// a leaderboard query. Args are []any because database/sql's QueryContext
// requires interface{} for all parameter values regardless of concrete type.
type whereFragment struct {
	conds []string
	args  []any
}

// buildLeaderboardConditions returns extra WHERE conditions for the given filters.
// positionCol is the fully-qualified column name for the position filter
// ("ps.primary_position" for batting, "ps.pitcher_role" for pitching).
// seasonAlias is the seasons table alias (typically "s").
// The base is_regular_season condition is NOT included — callers add it first.
func buildLeaderboardConditions(f models.LeaderboardFilters, positionCol, seasonAlias string) whereFragment {
	var w whereFragment
	if f.OnlyHallOfFamers {
		w.conds = append(w.conds, "p.is_hall_of_famer = 1")
	}
	if f.Position != "" {
		w.conds = append(w.conds, positionCol+" = ?")
		w.args = append(w.args, f.Position)
	}
	if f.BatHand != "" {
		w.conds = append(w.conds, "ps.bat_hand = ?")
		w.args = append(w.args, f.BatHand)
	}
	if f.ThrowHand != "" {
		w.conds = append(w.conds, "ps.throw_hand = ?")
		w.args = append(w.args, f.ThrowHand)
	}
	if f.ChemistryType != "" {
		w.conds = append(w.conds, "ps.chemistry_type = ?")
		w.args = append(w.args, f.ChemistryType)
	}
	if f.SeasonStart > 0 {
		w.conds = append(w.conds, seasonAlias+".season_num >= ?")
		w.args = append(w.args, f.SeasonStart)
	}
	if f.SeasonEnd > 0 {
		w.conds = append(w.conds, seasonAlias+".season_num <= ?")
		w.args = append(w.args, f.SeasonEnd)
	}
	return w
}

// GetBattingCareerLeaders returns one row per player with career batting totals
// summed across all seasons matching the filters. Rate fields are left nil —
// the caller must call service.ComputeBattingRates on each row.
func (s *LeaderboardQueryStore) GetBattingCareerLeaders(
	ctx context.Context, f models.LeaderboardFilters,
) ([]models.BattingCareerLeaderRow, error) {
	isReg := 1
	if f.IsPlayoffs {
		isReg = 0
	}

	args := []any{isReg}
	w := buildLeaderboardConditions(f, "ps.primary_position", "s")
	args = append(args, w.args...)

	whereExtra := ""
	if len(w.conds) > 0 {
		whereExtra = " AND " + strings.Join(w.conds, " AND ")
	}

	q := `
SELECT
    p.id, p.first_name, p.last_name, p.is_hall_of_famer,
    COUNT(DISTINCT ps.season_id)           AS seasons_played,
    COALESCE(SUM(b.games_played),    0),
    COALESCE(SUM(b.games_batting),   0),
    COALESCE(SUM(b.at_bats),         0),
    COALESCE(SUM(b.runs),            0),
    COALESCE(SUM(b.hits),            0),
    COALESCE(SUM(b.doubles),         0),
    COALESCE(SUM(b.triples),         0),
    COALESCE(SUM(b.home_runs),       0),
    COALESCE(SUM(b.rbi),             0),
    COALESCE(SUM(b.stolen_bases),    0),
    COALESCE(SUM(b.caught_stealing), 0),
    COALESCE(SUM(b.walks),           0),
    COALESCE(SUM(b.strikeouts),      0),
    COALESCE(SUM(b.hit_by_pitch),    0),
    COALESCE(SUM(b.sac_hits),        0),
    COALESCE(SUM(b.sac_flies),       0),
    COALESCE(SUM(b.errors),          0),
    COALESCE(SUM(b.passed_balls),    0),
    SUM(b.smb_war),
    -- Career OPS+: career totals vs career-weighted league averages from league_season_stats.
    -- NULL when no league_season_stats rows exist (seasons not yet re-synced post-8.5).
    CASE
        WHEN COALESCE(SUM(b.at_bats), 0) > 0
         AND COALESCE(SUM(b.at_bats + b.walks + b.hit_by_pitch + b.sac_flies), 0) > 0
         AND COALESCE(SUM(lss.total_at_bats), 0) > 0
         AND COALESCE(SUM(lss.total_at_bats + lss.total_walks + lss.total_hbp + lss.total_sac_flies), 0) > 0
        THEN 100.0 * (
            CAST(SUM(b.hits + b.walks + b.hit_by_pitch) AS REAL)
                / SUM(b.at_bats + b.walks + b.hit_by_pitch + b.sac_flies)
                / (CAST(SUM(lss.total_hits + lss.total_walks + lss.total_hbp) AS REAL)
                   / SUM(lss.total_at_bats + lss.total_walks + lss.total_hbp + lss.total_sac_flies))
            + CAST(SUM(b.hits - b.doubles - b.triples - b.home_runs
                        + b.doubles * 2 + b.triples * 3 + b.home_runs * 4) AS REAL)
                / SUM(b.at_bats)
                / (CAST(SUM(lss.total_hits - lss.total_doubles - lss.total_triples - lss.total_home_runs
                             + lss.total_doubles * 2 + lss.total_triples * 3 + lss.total_home_runs * 4) AS REAL)
                   / SUM(lss.total_at_bats))
            - 1.0
        )
        ELSE NULL
    END AS career_ops_plus
FROM player_season_batting_stats b
JOIN player_seasons ps ON ps.id = b.player_season_id
JOIN seasons s         ON s.id  = ps.season_id
JOIN players p         ON p.id  = ps.player_id
LEFT JOIN league_season_stats lss ON lss.season_id = ps.season_id
    AND lss.is_regular_season = b.is_regular_season
WHERE b.is_regular_season = ?` + whereExtra + `
GROUP BY p.id
HAVING COALESCE(SUM(b.at_bats), 0) > 0
ORDER BY COALESCE(SUM(b.smb_war), -9999.0) DESC, p.last_name, p.first_name`

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("GetBattingCareerLeaders: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.BattingCareerLeaderRow
	for rows.Next() {
		var r models.BattingCareerLeaderRow
		var hof int
		if err := rows.Scan(
			&r.PlayerID, &r.FirstName, &r.LastName, &hof,
			&r.SeasonsPlayed,
			&r.GamesPlayed, &r.GamesBatting,
			&r.AtBats, &r.Runs, &r.Hits, &r.Doubles, &r.Triples, &r.HomeRuns, &r.RBI,
			&r.StolenBases, &r.CaughtStealing, &r.Walks, &r.Strikeouts,
			&r.HitByPitch, &r.SacHits, &r.SacFlies, &r.Errors, &r.PassedBalls,
			&r.SmbWAR, &r.OPSPlus,
		); err != nil {
			return nil, fmt.Errorf("GetBattingCareerLeaders scan: %w", err)
		}
		r.IsHallOfFamer = hof == 1
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetBattingSeasonLeaders returns one row per player-season with batting stats
// for each individual season matching the filters. Rate fields are left nil.
func (s *LeaderboardQueryStore) GetBattingSeasonLeaders(
	ctx context.Context, f models.LeaderboardFilters,
) ([]models.BattingSeasonLeaderRow, error) {
	isReg := 1
	if f.IsPlayoffs {
		isReg = 0
	}

	args := []any{isReg}
	w := buildLeaderboardConditions(f, "ps.primary_position", "s")
	args = append(args, w.args...)

	whereExtra := ""
	if len(w.conds) > 0 {
		whereExtra = " AND " + strings.Join(w.conds, " AND ")
	}

	q := `
SELECT
    p.id, p.first_name, p.last_name, p.is_hall_of_famer,
    s.season_num,
    COALESCE(tsh.team_name, ''),
    ps.age,
    ps.primary_position,
    ps.bat_hand,
    ps.chemistry_type,
    b.games_played, b.games_batting,
    b.at_bats, b.runs, b.hits, b.doubles, b.triples, b.home_runs, b.rbi,
    b.stolen_bases, b.caught_stealing, b.walks, b.strikeouts,
    b.hit_by_pitch, b.sac_hits, b.sac_flies, b.errors, b.passed_balls,
    b.ops_plus, b.smb_war
FROM player_season_batting_stats b
JOIN player_seasons ps ON ps.id = b.player_season_id
JOIN seasons s         ON s.id  = ps.season_id
JOIN players p         ON p.id  = ps.player_id
LEFT JOIN player_season_teams pst ON pst.player_season_id = ps.id AND pst.sort_order = 0
LEFT JOIN team_season_history tsh ON tsh.id = pst.team_history_id
WHERE b.is_regular_season = ?
  AND b.at_bats > 0` + whereExtra + `
ORDER BY COALESCE(b.smb_war, -9999.0) DESC, p.last_name, p.first_name, s.season_num`

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("GetBattingSeasonLeaders: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.BattingSeasonLeaderRow
	for rows.Next() {
		var r models.BattingSeasonLeaderRow
		var hof int
		if err := rows.Scan(
			&r.PlayerID, &r.FirstName, &r.LastName, &hof,
			&r.SeasonNum, &r.TeamName, &r.Age,
			&r.PrimaryPosition, &r.BatHand, &r.ChemistryType,
			&r.GamesPlayed, &r.GamesBatting,
			&r.AtBats, &r.Runs, &r.Hits, &r.Doubles, &r.Triples, &r.HomeRuns, &r.RBI,
			&r.StolenBases, &r.CaughtStealing, &r.Walks, &r.Strikeouts,
			&r.HitByPitch, &r.SacHits, &r.SacFlies, &r.Errors, &r.PassedBalls,
			&r.OPSPlus, &r.SmbWAR,
		); err != nil {
			return nil, fmt.Errorf("GetBattingSeasonLeaders scan: %w", err)
		}
		r.IsHallOfFamer = hof == 1
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetPitchingCareerLeaders returns one row per player with career pitching totals
// summed across all seasons matching the filters. Rate fields are left nil.
func (s *LeaderboardQueryStore) GetPitchingCareerLeaders(
	ctx context.Context, f models.LeaderboardFilters,
) ([]models.PitchingCareerLeaderRow, error) {
	isReg := 1
	if f.IsPlayoffs {
		isReg = 0
	}

	args := []any{isReg}
	w := buildLeaderboardConditions(f, "ps.pitcher_role", "s")
	args = append(args, w.args...)

	whereExtra := ""
	if len(w.conds) > 0 {
		whereExtra = " AND " + strings.Join(w.conds, " AND ")
	}

	q := `
SELECT
    p.id, p.first_name, p.last_name, p.is_hall_of_famer,
    COUNT(DISTINCT ps.season_id)                   AS seasons_played,
    COALESCE(SUM(pit.wins),             0),
    COALESCE(SUM(pit.losses),           0),
    COALESCE(SUM(pit.games),            0),
    COALESCE(SUM(pit.games_started),    0),
    COALESCE(SUM(pit.complete_games),   0),
    COALESCE(SUM(pit.shutouts),         0),
    COALESCE(SUM(pit.saves),            0),
    COALESCE(SUM(pit.outs_pitched),     0),
    COALESCE(SUM(pit.hits_allowed),     0),
    COALESCE(SUM(pit.earned_runs),      0),
    COALESCE(SUM(pit.home_runs_allowed),0),
    COALESCE(SUM(pit.walks),            0),
    COALESCE(SUM(pit.strikeouts),       0),
    COALESCE(SUM(pit.hit_batters),      0),
    COALESCE(SUM(pit.batters_faced),    0),
    COALESCE(SUM(pit.games_finished),   0),
    COALESCE(SUM(pit.runs_allowed),     0),
    COALESCE(SUM(pit.wild_pitches),     0),
    COALESCE(SUM(pit.total_pitches),    0),
    SUM(pit.smb_war),
    -- Career ERA+: career ERA vs career-weighted league ERA from league_season_stats.
    -- NULL when no league_season_stats rows exist (seasons not yet re-synced post-8.5).
    CASE
        WHEN COALESCE(SUM(pit.outs_pitched), 0) > 0
         AND COALESCE(SUM(pit.earned_runs),  0) > 0
         AND COALESCE(SUM(lss.total_outs_pitched), 0) > 0
        THEN CAST(SUM(lss.total_earned_runs) AS REAL) * 27.0 / SUM(lss.total_outs_pitched)
             / (CAST(SUM(pit.earned_runs) AS REAL) * 27.0 / SUM(pit.outs_pitched))
             * 100.0
        ELSE NULL
    END AS career_era_plus
FROM player_season_pitching_stats pit
JOIN player_seasons ps ON ps.id = pit.player_season_id
JOIN seasons s         ON s.id  = ps.season_id
JOIN players p         ON p.id  = ps.player_id
LEFT JOIN league_season_stats lss ON lss.season_id = ps.season_id
    AND lss.is_regular_season = pit.is_regular_season
WHERE pit.is_regular_season = ?` + whereExtra + `
GROUP BY p.id
HAVING COALESCE(SUM(pit.outs_pitched), 0) > 0
ORDER BY COALESCE(SUM(pit.smb_war), -9999.0) DESC, p.last_name, p.first_name`

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("GetPitchingCareerLeaders: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.PitchingCareerLeaderRow
	for rows.Next() {
		var r models.PitchingCareerLeaderRow
		var hof int
		if err := rows.Scan(
			&r.PlayerID, &r.FirstName, &r.LastName, &hof,
			&r.SeasonsPlayed,
			&r.Wins, &r.Losses, &r.Games, &r.GamesStarted,
			&r.CompleteGames, &r.Shutouts, &r.Saves, &r.OutsPitched,
			&r.HitsAllowed, &r.EarnedRuns, &r.HomeRunsAllowed, &r.Walks,
			&r.Strikeouts, &r.HitBatters, &r.BattersFaced, &r.GamesFinished,
			&r.RunsAllowed, &r.WildPitches, &r.TotalPitches,
			&r.SmbWAR, &r.ERAPlus,
		); err != nil {
			return nil, fmt.Errorf("GetPitchingCareerLeaders scan: %w", err)
		}
		r.IsHallOfFamer = hof == 1
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetPitchingSeasonLeaders returns one row per player-season with pitching stats
// for each individual season matching the filters. Rate fields are left nil.
func (s *LeaderboardQueryStore) GetPitchingSeasonLeaders(
	ctx context.Context, f models.LeaderboardFilters,
) ([]models.PitchingSeasonLeaderRow, error) {
	isReg := 1
	if f.IsPlayoffs {
		isReg = 0
	}

	args := []any{isReg}
	w := buildLeaderboardConditions(f, "ps.pitcher_role", "s")
	args = append(args, w.args...)

	whereExtra := ""
	if len(w.conds) > 0 {
		whereExtra = " AND " + strings.Join(w.conds, " AND ")
	}

	q := `
SELECT
    p.id, p.first_name, p.last_name, p.is_hall_of_famer,
    s.season_num,
    COALESCE(tsh.team_name, ''),
    ps.age,
    ps.pitcher_role,
    ps.throw_hand,
    ps.chemistry_type,
    pit.wins, pit.losses, pit.games, pit.games_started,
    pit.complete_games, pit.shutouts, pit.saves, pit.outs_pitched,
    pit.hits_allowed, pit.earned_runs, pit.home_runs_allowed, pit.walks,
    pit.strikeouts, pit.hit_batters, pit.batters_faced, pit.games_finished,
    pit.runs_allowed, pit.wild_pitches, pit.total_pitches,
    pit.era_plus, pit.fip, pit.fip_minus, pit.smb_war
FROM player_season_pitching_stats pit
JOIN player_seasons ps ON ps.id = pit.player_season_id
JOIN seasons s         ON s.id  = ps.season_id
JOIN players p         ON p.id  = ps.player_id
LEFT JOIN player_season_teams pst ON pst.player_season_id = ps.id AND pst.sort_order = 0
LEFT JOIN team_season_history tsh ON tsh.id = pst.team_history_id
WHERE pit.is_regular_season = ?
  AND pit.outs_pitched > 0` + whereExtra + `
ORDER BY COALESCE(pit.smb_war, -9999.0) DESC, p.last_name, p.first_name, s.season_num`

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("GetPitchingSeasonLeaders: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.PitchingSeasonLeaderRow
	for rows.Next() {
		var r models.PitchingSeasonLeaderRow
		var hof int
		if err := rows.Scan(
			&r.PlayerID, &r.FirstName, &r.LastName, &hof,
			&r.SeasonNum, &r.TeamName, &r.Age,
			&r.PitcherRole, &r.ThrowHand, &r.ChemistryType,
			&r.Wins, &r.Losses, &r.Games, &r.GamesStarted,
			&r.CompleteGames, &r.Shutouts, &r.Saves, &r.OutsPitched,
			&r.HitsAllowed, &r.EarnedRuns, &r.HomeRunsAllowed, &r.Walks,
			&r.Strikeouts, &r.HitBatters, &r.BattersFaced, &r.GamesFinished,
			&r.RunsAllowed, &r.WildPitches, &r.TotalPitches,
			&r.ERAPlus, &r.FIP, &r.FIPMinus, &r.SmbWAR,
		); err != nil {
			return nil, fmt.Errorf("GetPitchingSeasonLeaders scan: %w", err)
		}
		r.IsHallOfFamer = hof == 1
		out = append(out, r)
	}
	return out, rows.Err()
}
