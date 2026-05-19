package store

import (
	"context"
	"database/sql"
	"fmt"

	"smb-tools/internal/models"
)

// TeamQueryStore provides read-only queries over team, roster, and schedule data.
type TeamQueryStore struct {
	db DBTX
}

func NewTeamQueryStore(db DBTX) *TeamQueryStore {
	return &TeamQueryStore{db: db}
}

// SearchTeams returns up to 50 teams whose name (in any season) matches the
// query string (case-insensitive LIKE). Results show each team once, using the
// most recent name it had.
func (s *TeamQueryStore) SearchTeams(ctx context.Context, query string) ([]models.TeamSearchResult, error) {
	pattern := "%" + query + "%"
	rows, err := s.db.QueryContext(ctx, `
SELECT
    t.id AS team_id,
    (
        SELECT tsh2.team_name
        FROM team_season_history tsh2
        JOIN seasons s2 ON s2.id = tsh2.season_id
        WHERE tsh2.team_id = t.id
        ORDER BY s2.season_num DESC
        LIMIT 1
    ) AS current_name,
    COUNT(DISTINCT tsh.season_id) AS seasons,
    MIN(s.season_num)             AS first_season,
    MAX(s.season_num)             AS last_season
FROM teams t
JOIN team_season_history tsh ON tsh.team_id = t.id
JOIN seasons s               ON s.id = tsh.season_id
WHERE tsh.team_name LIKE ?
GROUP BY t.id
ORDER BY current_name
LIMIT 50
`, pattern)
	if err != nil {
		return nil, fmt.Errorf("searching teams %q: %w", query, err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.TeamSearchResult
	for rows.Next() {
		var r models.TeamSearchResult
		if err := rows.Scan(&r.TeamID, &r.TeamName, &r.Seasons, &r.FirstSeason, &r.LastSeason); err != nil {
			return nil, fmt.Errorf("scanning team search result: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListAllTeamSeasons returns every team-season row ordered by season descending,
// then team name ascending. Each row includes a champion flag.
func (s *TeamQueryStore) ListAllTeamSeasons(ctx context.Context) ([]models.TeamSeasonListRow, error) {
	q := championCTE + `
SELECT
    s.season_num,
    tsh.id          AS history_id,
    tsh.team_id,
    tsh.team_name,
    tsh.conference_name,
    tsh.division_name,
    tsh.wins,
    tsh.losses,
    tsh.runs_for,
    tsh.runs_against,
    tsh.playoff_seed,
    tsh.playoff_wins,
    tsh.playoff_losses,
    CASE WHEN c.winner_history_id = tsh.id THEN 1 ELSE 0 END AS is_champion
FROM team_season_history tsh
JOIN seasons s ON s.id = tsh.season_id
LEFT JOIN champion c ON c.season_id = tsh.season_id
ORDER BY s.season_num DESC, tsh.team_name ASC
`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("listing all team seasons: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.TeamSeasonListRow
	for rows.Next() {
		var r models.TeamSeasonListRow
		var seed, pw, pl sql.NullInt64
		var isChampion int
		if err := rows.Scan(
			&r.SeasonNum, &r.HistoryID, &r.TeamID, &r.TeamName,
			&r.ConferenceName, &r.DivisionName,
			&r.Wins, &r.Losses, &r.RunsFor, &r.RunsAgainst,
			&seed, &pw, &pl, &isChampion,
		); err != nil {
			return nil, fmt.Errorf("scanning team season list row: %w", err)
		}
		if r.Wins+r.Losses > 0 {
			r.WinPct = float64(r.Wins) / float64(r.Wins+r.Losses)
		}
		if seed.Valid {
			v := int(seed.Int64)
			r.PlayoffSeed = &v
		}
		if pw.Valid {
			v := int(pw.Int64)
			r.PlayoffWins = &v
		}
		if pl.Valid {
			v := int(pl.Int64)
			r.PlayoffLosses = &v
		}
		r.IsChampion = isChampion == 1
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetTeamHistory returns all seasons played by the given team, enriched with
// the champion flag, ordered by season number ascending.
// Returns sql.ErrNoRows wrapped in an error if the team does not exist.
func (s *TeamQueryStore) GetTeamHistory(ctx context.Context, teamID int64) (models.TeamHistory, error) {
	var th models.TeamHistory
	if err := s.db.QueryRowContext(ctx,
		`SELECT id, game_guid FROM teams WHERE id = ?`, teamID,
	).Scan(&th.TeamID, &th.GameGUID); err != nil {
		return th, fmt.Errorf("getting team %d: %w", teamID, err)
	}

	q := championCTE + `
SELECT
    tsh.id,
    s.season_num,
    tsh.team_name,
    tsh.division_name,
    tsh.conference_name,
    tsh.wins,
    tsh.losses,
    tsh.games_back,
    tsh.runs_for,
    tsh.runs_against,
    tsh.budget,
    tsh.payroll,
    tsh.playoff_seed,
    tsh.playoff_wins,
    tsh.playoff_losses,
    tsh.playoff_runs_for,
    tsh.playoff_runs_against,
    tsh.total_power,    tsh.total_contact,  tsh.total_speed,
    tsh.total_fielding, tsh.total_arm,
    tsh.total_velocity, tsh.total_junk,     tsh.total_accuracy,
    CASE WHEN c.winner_history_id = tsh.id THEN 1 ELSE 0 END AS is_champion
FROM team_season_history tsh
JOIN seasons s ON s.id = tsh.season_id
LEFT JOIN champion c ON c.season_id = tsh.season_id
WHERE tsh.team_id = ?
ORDER BY s.season_num ASC
`
	rows, err := s.db.QueryContext(ctx, q, teamID)
	if err != nil {
		return th, fmt.Errorf("getting team history for %d: %w", teamID, err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var ts models.TeamSeasonSummary
		var seed, pw, pl, prf, pra sql.NullInt64
		var isChampion int
		if err := rows.Scan(
			&ts.HistoryID, &ts.SeasonNum, &ts.TeamName,
			&ts.DivisionName, &ts.ConferenceName,
			&ts.Wins, &ts.Losses, &ts.GamesBack, &ts.RunsFor, &ts.RunsAgainst,
			&ts.Budget, &ts.Payroll,
			&seed, &pw, &pl, &prf, &pra,
			&ts.TotalPower, &ts.TotalContact, &ts.TotalSpeed,
			&ts.TotalFielding, &ts.TotalArm,
			&ts.TotalVelocity, &ts.TotalJunk, &ts.TotalAccuracy,
			&isChampion,
		); err != nil {
			return th, fmt.Errorf("scanning team history row: %w", err)
		}
		if ts.Wins+ts.Losses > 0 {
			ts.WinPct = float64(ts.Wins) / float64(ts.Wins+ts.Losses)
		}
		nullIntToPtr := func(n sql.NullInt64) *int {
			if !n.Valid {
				return nil
			}
			v := int(n.Int64)
			return &v
		}
		ts.PlayoffSeed = nullIntToPtr(seed)
		ts.PlayoffWins = nullIntToPtr(pw)
		ts.PlayoffLosses = nullIntToPtr(pl)
		ts.PlayoffRunsFor = nullIntToPtr(prf)
		ts.PlayoffRunsAgainst = nullIntToPtr(pra)
		ts.IsChampion = isChampion == 1
		th.Seasons = append(th.Seasons, ts)
	}
	return th, rows.Err()
}

// GetTeamSeasonRoster returns all players on a team in a given season (by
// team_history_id), with regular season batting and pitching stats.
func (s *TeamQueryStore) GetTeamSeasonRoster(ctx context.Context, teamHistoryID int64) ([]models.RosterPlayer, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT
    p.id, p.first_name, p.last_name, p.is_hall_of_famer,
    ps.age, ps.salary,
    ps.primary_position, ps.secondary_position, ps.pitcher_role,
    ps.bat_hand, ps.throw_hand, ps.chemistry_type,
    ps.traits_json, ps.pitches_json,
    COALESCE(gs.power,0),    COALESCE(gs.contact,0),
    COALESCE(gs.speed,0),    COALESCE(gs.fielding,0), COALESCE(gs.arm,0),
    COALESCE(gs.velocity,0), COALESCE(gs.junk,0),     COALESCE(gs.accuracy,0),
    -- batting sentinel first, then the rest
    b.at_bats,
    b.games_played, b.games_batting, b.runs, b.hits,
    b.doubles, b.triples, b.home_runs, b.rbi,
    b.stolen_bases, b.caught_stealing, b.walks, b.strikeouts,
    b.hit_by_pitch, b.sac_hits, b.sac_flies, b.errors, b.passed_balls,
    -- pitching sentinel first, then the rest
    pit.outs_pitched,
    pit.wins, pit.losses, pit.games, pit.games_started,
    pit.complete_games, pit.shutouts, pit.saves,
    pit.hits_allowed, pit.earned_runs, pit.home_runs_allowed,
    pit.walks, pit.strikeouts, pit.hit_batters, pit.batters_faced,
    pit.games_finished, pit.runs_allowed, pit.wild_pitches, pit.total_pitches
FROM player_seasons ps
JOIN players p ON p.id = ps.player_id
LEFT JOIN player_season_game_stats gs ON gs.player_season_id = ps.id
LEFT JOIN player_season_batting_stats b
    ON b.player_season_id = ps.id AND b.is_regular_season = 1
LEFT JOIN player_season_pitching_stats pit
    ON pit.player_season_id = ps.id AND pit.is_regular_season = 1
WHERE ps.team_history_id = ?
ORDER BY ps.primary_position, p.last_name
`, teamHistoryID)
	if err != nil {
		return nil, fmt.Errorf("getting roster for team history %d: %w", teamHistoryID, err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.RosterPlayer
	for rows.Next() {
		var r models.RosterPlayer
		var hof int

		var bAtBats sql.NullInt64
		var bGP, bGB, bRuns, bHits, bDB, bTR, bHR, bRBI sql.NullInt64
		var bSB, bCS, bWalks, bK, bHBP, bSH, bSF, bE, bPB sql.NullInt64

		var pOuts sql.NullInt64
		var pW, pL, pG, pGS, pCG, pSHO, pSV sql.NullInt64
		var pH, pER, pHRA, pWalks, pK, pHBP, pBF, pGF, pRA, pWP, pTP sql.NullInt64

		if err := rows.Scan(
			&r.PlayerID, &r.FirstName, &r.LastName, &hof,
			&r.Age, &r.Salary,
			&r.PrimaryPosition, &r.SecondaryPosition, &r.PitcherRole,
			&r.BatHand, &r.ThrowHand, &r.ChemistryType,
			&r.TraitsJSON, &r.PitchesJSON,
			&r.Power, &r.Contact, &r.Speed, &r.Fielding, &r.Arm,
			&r.Velocity, &r.Junk, &r.Accuracy,
			&bAtBats,
			&bGP, &bGB, &bRuns, &bHits, &bDB, &bTR, &bHR, &bRBI,
			&bSB, &bCS, &bWalks, &bK, &bHBP, &bSH, &bSF, &bE, &bPB,
			&pOuts,
			&pW, &pL, &pG, &pGS, &pCG, &pSHO, &pSV,
			&pH, &pER, &pHRA, &pWalks, &pK, &pHBP, &pBF, &pGF, &pRA, &pWP, &pTP,
		); err != nil {
			return nil, fmt.Errorf("scanning roster player: %w", err)
		}
		r.IsHallOfFamer = hof == 1

		if bAtBats.Valid {
			r.Batting = &models.CareerBattingStats{
				AtBats: int(bAtBats.Int64), GamesPlayed: int(bGP.Int64), GamesBatting: int(bGB.Int64),
				Runs: int(bRuns.Int64), Hits: int(bHits.Int64),
				Doubles: int(bDB.Int64), Triples: int(bTR.Int64), HomeRuns: int(bHR.Int64),
				RBI: int(bRBI.Int64), StolenBases: int(bSB.Int64), CaughtStealing: int(bCS.Int64),
				Walks: int(bWalks.Int64), Strikeouts: int(bK.Int64), HitByPitch: int(bHBP.Int64),
				SacHits: int(bSH.Int64), SacFlies: int(bSF.Int64),
				Errors: int(bE.Int64), PassedBalls: int(bPB.Int64),
			}
		}
		if pOuts.Valid {
			r.Pitching = &models.CareerPitchingStats{
				OutsPitched: int(pOuts.Int64),
				Wins: int(pW.Int64), Losses: int(pL.Int64), Games: int(pG.Int64),
				GamesStarted: int(pGS.Int64), CompleteGames: int(pCG.Int64),
				Shutouts: int(pSHO.Int64), Saves: int(pSV.Int64),
				HitsAllowed: int(pH.Int64), EarnedRuns: int(pER.Int64),
				HomeRunsAllowed: int(pHRA.Int64), Walks: int(pWalks.Int64),
				Strikeouts: int(pK.Int64), HitBatters: int(pHBP.Int64),
				BattersFaced: int(pBF.Int64), GamesFinished: int(pGF.Int64),
				RunsAllowed: int(pRA.Int64), WildPitches: int(pWP.Int64),
				TotalPitches: int(pTP.Int64),
			}
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetTeamSeasonSchedule returns all regular season games (home and away) for a
// team in a given season, ordered by game number.
func (s *TeamQueryStore) GetTeamSeasonSchedule(ctx context.Context, teamHistoryID int64, seasonID int) ([]models.ScheduleGameRow, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT
    g.game_number,
    g.day,
    g.home_team_history_id,
    home.team_name AS home_team_name,
    g.away_team_history_id,
    away.team_name AS away_team_name,
    g.home_score,
    g.away_score,
    COALESCE(hp.first_name || ' ' || hp.last_name, '') AS home_pitcher,
    COALESCE(ap.first_name || ' ' || ap.last_name, '') AS away_pitcher
FROM team_season_schedules g
JOIN team_season_history home ON home.id = g.home_team_history_id
JOIN team_season_history away ON away.id = g.away_team_history_id
LEFT JOIN player_seasons hps ON hps.id = g.home_pitcher_season_id
LEFT JOIN players hp          ON hp.id  = hps.player_id
LEFT JOIN player_seasons aps ON aps.id = g.away_pitcher_season_id
LEFT JOIN players ap          ON ap.id  = aps.player_id
WHERE g.season_id = ?
  AND (g.home_team_history_id = ? OR g.away_team_history_id = ?)
ORDER BY g.game_number ASC
`, seasonID, teamHistoryID, teamHistoryID)
	if err != nil {
		return nil, fmt.Errorf("getting schedule for team history %d season %d: %w", teamHistoryID, seasonID, err)
	}
	defer func() { _ = rows.Close() }()

	return scanScheduleRows(rows)
}

// GetTeamSeasonPlayoffSchedule returns all playoff games for a team in a given
// season, ordered by series number then game number.
func (s *TeamQueryStore) GetTeamSeasonPlayoffSchedule(ctx context.Context, teamHistoryID int64, seasonID int) ([]models.PlayoffGameRow, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT
    g.series_number,
    g.game_number,
    g.home_team_history_id,
    home.team_name AS home_team_name,
    g.away_team_history_id,
    away.team_name AS away_team_name,
    g.home_score,
    g.away_score,
    COALESCE(hp.first_name || ' ' || hp.last_name, '') AS home_pitcher,
    COALESCE(ap.first_name || ' ' || ap.last_name, '') AS away_pitcher
FROM team_playoff_schedules g
JOIN team_season_history home ON home.id = g.home_team_history_id
JOIN team_season_history away ON away.id = g.away_team_history_id
LEFT JOIN player_seasons hps ON hps.id = g.home_pitcher_season_id
LEFT JOIN players hp          ON hp.id  = hps.player_id
LEFT JOIN player_seasons aps ON aps.id = g.away_pitcher_season_id
LEFT JOIN players ap          ON ap.id  = aps.player_id
WHERE g.season_id = ?
  AND (g.home_team_history_id = ? OR g.away_team_history_id = ?)
ORDER BY g.series_number ASC, g.game_number ASC
`, seasonID, teamHistoryID, teamHistoryID)
	if err != nil {
		return nil, fmt.Errorf("getting playoff schedule for team history %d season %d: %w", teamHistoryID, seasonID, err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.PlayoffGameRow
	for rows.Next() {
		var r models.PlayoffGameRow
		var homeScore, awayScore sql.NullInt64
		if err := rows.Scan(
			&r.SeriesNumber, &r.GameNumber,
			&r.HomeTeamHistoryID, &r.HomeTeamName,
			&r.AwayTeamHistoryID, &r.AwayTeamName,
			&homeScore, &awayScore,
			&r.HomePitcherName, &r.AwayPitcherName,
		); err != nil {
			return nil, fmt.Errorf("scanning playoff game row: %w", err)
		}
		if homeScore.Valid {
			v := int(homeScore.Int64)
			r.HomeScore = &v
		}
		if awayScore.Valid {
			v := int(awayScore.Int64)
			r.AwayScore = &v
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func scanScheduleRows(rows *sql.Rows) ([]models.ScheduleGameRow, error) {
	var out []models.ScheduleGameRow
	for rows.Next() {
		var r models.ScheduleGameRow
		var homeScore, awayScore sql.NullInt64
		if err := rows.Scan(
			&r.GameNumber, &r.Day,
			&r.HomeTeamHistoryID, &r.HomeTeamName,
			&r.AwayTeamHistoryID, &r.AwayTeamName,
			&homeScore, &awayScore,
			&r.HomePitcherName, &r.AwayPitcherName,
		); err != nil {
			return nil, fmt.Errorf("scanning schedule game row: %w", err)
		}
		if homeScore.Valid {
			v := int(homeScore.Int64)
			r.HomeScore = &v
		}
		if awayScore.Valid {
			v := int(awayScore.Int64)
			r.AwayScore = &v
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
