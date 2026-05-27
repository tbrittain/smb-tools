package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"smb-tools/internal/models"
)

// SeasonQueryStore provides read-only queries over season and standings data.
type SeasonQueryStore struct {
	db DBTX
}

func NewSeasonQueryStore(db DBTX) *SeasonQueryStore {
	return &SeasonQueryStore{db: db}
}

// ListWithChampion returns all seasons ordered by season number ascending.
// Each season is enriched with the name of the playoff champion, if determinable.
func (s *SeasonQueryStore) ListWithChampion(ctx context.Context) ([]models.SeasonSummary, error) {
	q := `
SELECT
    s.id,
    s.season_num,
    s.num_games,
    s.imported_at,
    COALESCE(tsh.team_name, '') AS champion_team_name,
    tsh.id                      AS champion_history_id
FROM seasons s
LEFT JOIN season_champions c  ON c.season_id = s.id
LEFT JOIN team_season_history tsh ON tsh.id = c.winner_history_id
ORDER BY s.season_num ASC
`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("listing seasons with champion: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.SeasonSummary
	for rows.Next() {
		var (
			ss              models.SeasonSummary
			importedAt      string
			championHistID  sql.NullInt64
		)
		if err := rows.Scan(
			&ss.ID, &ss.SeasonNum, &ss.NumGames, &importedAt,
			&ss.ChampionTeamName, &championHistID,
		); err != nil {
			return nil, fmt.Errorf("scanning season summary: %w", err)
		}
		ss.ImportedAt, _ = time.Parse("2006-01-02 15:04:05", importedAt)
		if championHistID.Valid {
			ss.ChampionHistoryID = &championHistID.Int64
		}
		out = append(out, ss)
	}
	return out, rows.Err()
}

// GetPlayoffSeriesLength returns the playoff_series_length for the given season,
// or nil when the season has no playoff config (playoff_series_length = 0).
func (s *SeasonQueryStore) GetPlayoffSeriesLength(ctx context.Context, seasonID int64) (*int, error) {
	var sl int64
	if err := s.db.QueryRowContext(ctx,
		`SELECT playoff_series_length FROM seasons WHERE id = ?`, seasonID,
	).Scan(&sl); err != nil {
		return nil, fmt.Errorf("getting playoff series length for season %d: %w", seasonID, err)
	}
	if sl == 0 {
		return nil, nil
	}
	v := int(sl)
	return &v, nil
}

// GetStandings returns all teams' standings for the given season, ordered by
// conference, then division, then wins descending.
func (s *SeasonQueryStore) GetStandings(ctx context.Context, seasonID int64) ([]models.TeamStandingRow, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT
    tsh.id,
    tsh.team_id,
    tsh.team_name,
    tsh.division_name,
    tsh.conference_name,
    tsh.wins,
    tsh.losses,
    tsh.games_back,
    tsh.runs_for,
    tsh.runs_against,
    tsh.playoff_seed
FROM team_season_history tsh
WHERE tsh.season_id = ?
ORDER BY tsh.conference_name, tsh.division_name, tsh.wins DESC, tsh.losses ASC
`, seasonID)
	if err != nil {
		return nil, fmt.Errorf("getting standings for season %d: %w", seasonID, err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.TeamStandingRow
	for rows.Next() {
		var r models.TeamStandingRow
		var playoffSeed sql.NullInt64
		if err := rows.Scan(
			&r.HistoryID, &r.TeamID, &r.TeamName, &r.DivisionName, &r.ConferenceName,
			&r.Wins, &r.Losses, &r.GamesBack,
			&r.RunsFor, &r.RunsAgainst, &playoffSeed,
		); err != nil {
			return nil, fmt.Errorf("scanning standing row: %w", err)
		}
		if r.Wins+r.Losses > 0 {
			r.WinPct = float64(r.Wins) / float64(r.Wins+r.Losses)
		}
		r.RunDiff = r.RunsFor - r.RunsAgainst
		if playoffSeed.Valid {
			v := int(playoffSeed.Int64)
			r.PlayoffSeed = &v
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetSeasonStatLeaders returns the single-season title leaders for BA, HR, RBI,
// ERA, Wins, and Strikeouts for regular-season play.
//
// Qualification thresholds (applied to avoid outliers):
//   - Batting: at_bats >= num_games * 3  (≈ 3 PA per team game)
//   - ERA:     outs_pitched >= num_games * 3  (≈ 1 IP per team game)
func (s *SeasonQueryStore) GetSeasonStatLeaders(ctx context.Context, seasonID int64) (models.StatLeaders, error) {
	var leaders models.StatLeaders

	// Fetch season_num for the response
	var seasonNum int
	if err := s.db.QueryRowContext(ctx,
		`SELECT season_num FROM seasons WHERE id = ?`, seasonID,
	).Scan(&seasonNum); err != nil {
		return leaders, fmt.Errorf("getting season num for %d: %w", seasonID, err)
	}
	leaders.SeasonNum = seasonNum

	type leaderQuery struct {
		dest *(*models.StatLeader)
		sql  string
		args []any
	}

	queries := []leaderQuery{
		{
			dest: &leaders.BA,
			sql: `
SELECT p.id, p.first_name, p.last_name, COALESCE(tsh.team_name,''),
    CAST(b.hits AS REAL) / b.at_bats AS ba
FROM player_season_batting_stats b
JOIN player_seasons ps ON ps.id = b.player_season_id
JOIN seasons s         ON s.id  = ps.season_id
JOIN players p         ON p.id  = ps.player_id
LEFT JOIN player_season_teams pst ON pst.player_season_id = ps.id AND pst.sort_order = 0
LEFT JOIN team_season_history tsh ON tsh.id = pst.team_history_id
WHERE ps.season_id = ? AND b.is_regular_season = 1
  AND b.at_bats > 0 AND b.at_bats >= s.num_games * 3
ORDER BY ba DESC LIMIT 1`,
			args: []any{seasonID},
		},
		{
			dest: &leaders.HR,
			sql: `
SELECT p.id, p.first_name, p.last_name, COALESCE(tsh.team_name,''),
    CAST(b.home_runs AS REAL)
FROM player_season_batting_stats b
JOIN player_seasons ps ON ps.id = b.player_season_id
JOIN players p         ON p.id  = ps.player_id
LEFT JOIN player_season_teams pst ON pst.player_season_id = ps.id AND pst.sort_order = 0
LEFT JOIN team_season_history tsh ON tsh.id = pst.team_history_id
WHERE ps.season_id = ? AND b.is_regular_season = 1
ORDER BY b.home_runs DESC LIMIT 1`,
			args: []any{seasonID},
		},
		{
			dest: &leaders.RBI,
			sql: `
SELECT p.id, p.first_name, p.last_name, COALESCE(tsh.team_name,''),
    CAST(b.rbi AS REAL)
FROM player_season_batting_stats b
JOIN player_seasons ps ON ps.id = b.player_season_id
JOIN players p         ON p.id  = ps.player_id
LEFT JOIN player_season_teams pst ON pst.player_season_id = ps.id AND pst.sort_order = 0
LEFT JOIN team_season_history tsh ON tsh.id = pst.team_history_id
WHERE ps.season_id = ? AND b.is_regular_season = 1
ORDER BY b.rbi DESC LIMIT 1`,
			args: []any{seasonID},
		},
		{
			dest: &leaders.ERA,
			sql: `
SELECT p.id, p.first_name, p.last_name, COALESCE(tsh.team_name,''),
    CAST(pit.earned_runs AS REAL) * 27.0 / pit.outs_pitched AS era
FROM player_season_pitching_stats pit
JOIN player_seasons ps ON ps.id = pit.player_season_id
JOIN seasons s         ON s.id  = ps.season_id
JOIN players p         ON p.id  = ps.player_id
LEFT JOIN player_season_teams pst ON pst.player_season_id = ps.id AND pst.sort_order = 0
LEFT JOIN team_season_history tsh ON tsh.id = pst.team_history_id
WHERE ps.season_id = ? AND pit.is_regular_season = 1
  AND pit.outs_pitched > 0 AND pit.outs_pitched >= s.num_games * 3
ORDER BY era ASC LIMIT 1`,
			args: []any{seasonID},
		},
		{
			dest: &leaders.Wins,
			sql: `
SELECT p.id, p.first_name, p.last_name, COALESCE(tsh.team_name,''),
    CAST(pit.wins AS REAL)
FROM player_season_pitching_stats pit
JOIN player_seasons ps ON ps.id = pit.player_season_id
JOIN players p         ON p.id  = ps.player_id
LEFT JOIN player_season_teams pst ON pst.player_season_id = ps.id AND pst.sort_order = 0
LEFT JOIN team_season_history tsh ON tsh.id = pst.team_history_id
WHERE ps.season_id = ? AND pit.is_regular_season = 1
ORDER BY pit.wins DESC LIMIT 1`,
			args: []any{seasonID},
		},
		{
			dest: &leaders.Strikeouts,
			sql: `
SELECT p.id, p.first_name, p.last_name, COALESCE(tsh.team_name,''),
    CAST(pit.strikeouts AS REAL)
FROM player_season_pitching_stats pit
JOIN player_seasons ps ON ps.id = pit.player_season_id
JOIN players p         ON p.id  = ps.player_id
LEFT JOIN player_season_teams pst ON pst.player_season_id = ps.id AND pst.sort_order = 0
LEFT JOIN team_season_history tsh ON tsh.id = pst.team_history_id
WHERE ps.season_id = ? AND pit.is_regular_season = 1
ORDER BY pit.strikeouts DESC LIMIT 1`,
			args: []any{seasonID},
		},
	}

	for _, lq := range queries {
		var sl models.StatLeader
		var sv sql.NullFloat64
		err := s.db.QueryRowContext(ctx, lq.sql, lq.args...).Scan(
			&sl.PlayerID, &sl.FirstName, &sl.LastName, &sl.TeamName, &sv,
		)
		if err == sql.ErrNoRows {
			continue // no qualifying player
		}
		if err != nil {
			return leaders, fmt.Errorf("scanning stat leader: %w", err)
		}
		if !sv.Valid {
			continue // computed stat is NULL (e.g. division by zero) — no leader
		}
		sl.StatValue = sv.Float64
		*lq.dest = &sl
	}

	return leaders, nil
}

// GetCareerLeaders returns the top-5 all-time career leaders for HR, Hits, RBI,
// Wins, Strikeouts, and Saves (regular season only).
func (s *SeasonQueryStore) GetCareerLeaders(ctx context.Context) (models.CareerLeaders, error) {
	var out models.CareerLeaders

	type careerQuery struct {
		dest *[]models.CareerLeaderRow
		sql  string
	}

	queries := []careerQuery{
		{
			dest: &out.HR,
			sql: `
SELECT p.id, p.first_name, p.last_name,
    CAST(SUM(b.home_runs) AS REAL) AS val,
    COUNT(DISTINCT ps.season_id)   AS seasons
FROM player_season_batting_stats b
JOIN player_seasons ps ON ps.id = b.player_season_id
JOIN players p         ON p.id  = ps.player_id
WHERE b.is_regular_season = 1
GROUP BY p.id ORDER BY val DESC LIMIT 5`,
		},
		{
			dest: &out.Hits,
			sql: `
SELECT p.id, p.first_name, p.last_name,
    CAST(SUM(b.hits) AS REAL) AS val,
    COUNT(DISTINCT ps.season_id) AS seasons
FROM player_season_batting_stats b
JOIN player_seasons ps ON ps.id = b.player_season_id
JOIN players p         ON p.id  = ps.player_id
WHERE b.is_regular_season = 1
GROUP BY p.id ORDER BY val DESC LIMIT 5`,
		},
		{
			dest: &out.RBI,
			sql: `
SELECT p.id, p.first_name, p.last_name,
    CAST(SUM(b.rbi) AS REAL) AS val,
    COUNT(DISTINCT ps.season_id) AS seasons
FROM player_season_batting_stats b
JOIN player_seasons ps ON ps.id = b.player_season_id
JOIN players p         ON p.id  = ps.player_id
WHERE b.is_regular_season = 1
GROUP BY p.id ORDER BY val DESC LIMIT 5`,
		},
		{
			dest: &out.Wins,
			sql: `
SELECT p.id, p.first_name, p.last_name,
    CAST(SUM(pit.wins) AS REAL) AS val,
    COUNT(DISTINCT ps.season_id) AS seasons
FROM player_season_pitching_stats pit
JOIN player_seasons ps ON ps.id = pit.player_season_id
JOIN players p         ON p.id  = ps.player_id
WHERE pit.is_regular_season = 1
GROUP BY p.id ORDER BY val DESC LIMIT 5`,
		},
		{
			dest: &out.Strikeouts,
			sql: `
SELECT p.id, p.first_name, p.last_name,
    CAST(SUM(pit.strikeouts) AS REAL) AS val,
    COUNT(DISTINCT ps.season_id) AS seasons
FROM player_season_pitching_stats pit
JOIN player_seasons ps ON ps.id = pit.player_season_id
JOIN players p         ON p.id  = ps.player_id
WHERE pit.is_regular_season = 1
GROUP BY p.id ORDER BY val DESC LIMIT 5`,
		},
		{
			dest: &out.Saves,
			sql: `
SELECT p.id, p.first_name, p.last_name,
    CAST(SUM(pit.saves) AS REAL) AS val,
    COUNT(DISTINCT ps.season_id) AS seasons
FROM player_season_pitching_stats pit
JOIN player_seasons ps ON ps.id = pit.player_season_id
JOIN players p         ON p.id  = ps.player_id
WHERE pit.is_regular_season = 1
GROUP BY p.id ORDER BY val DESC LIMIT 5`,
		},
	}

	for _, cq := range queries {
		rows, err := s.db.QueryContext(ctx, cq.sql)
		if err != nil {
			return out, fmt.Errorf("career leader query: %w", err)
		}
		for rows.Next() {
			var row models.CareerLeaderRow
			if err := rows.Scan(&row.PlayerID, &row.FirstName, &row.LastName, &row.StatValue, &row.SeasonsPlayed); err != nil {
				_ = rows.Close()
				return out, fmt.Errorf("scanning career leader: %w", err)
			}
			*cq.dest = append(*cq.dest, row)
		}
		if err := rows.Close(); err != nil {
			return out, err
		}
		if err := rows.Err(); err != nil {
			return out, err
		}
	}

	return out, nil
}
