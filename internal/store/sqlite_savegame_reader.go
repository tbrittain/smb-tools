package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"smb-tools/internal/models"
)

// SqliteSaveGameReader implements SaveGameReader against a decompressed SMB
// save game SQLite database. The connection must have been opened read-only.
type SqliteSaveGameReader struct {
	db      *sql.DB
	tmpPath string // temp file to clean up on Close; "" if caller manages it
}

// NewSqliteSaveGameReader wraps an existing read-only *sql.DB.
// Pass tmpPath="" if the caller manages temp file cleanup.
func NewSqliteSaveGameReader(db *sql.DB, tmpPath string) *SqliteSaveGameReader {
	return &SqliteSaveGameReader{db: db, tmpPath: tmpPath}
}

func (r *SqliteSaveGameReader) Close() error {
	err := r.db.Close()
	if r.tmpPath != "" {
		_ = os.Remove(r.tmpPath)
	}
	return err
}

func (r *SqliteSaveGameReader) GetLeagues(ctx context.Context) ([]models.SaveGameLeague, error) {
	// Join pattern mirrors SMB3Explorer's Franchises.sql:
	//   - t_franchise joined via leagueGUID = t_leagues.GUID (not the integer leagueId)
	//   - t_seasons joined via historicalLeagueGUID for the elimination flag
	// FranchiseID non-null → franchise mode
	// FranchiseID null + elimination → elimination mode
	// FranchiseID null + no elimination → season mode
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			COALESCE(hex(l.GUID), '')              AS leagueGUID,
			l.leagueId,
			l.leagueName,
			l.leagueTeamTypeId,
			COALESCE(tt.typeName, ''),
			f.franchiseId,
			MAX(COALESCE(ts.elimination, 0))       AS elimination,
			COUNT(DISTINCT fs.seasonID)            AS numSeasons,
			COALESCE(pt.teamName, '')              AS playerTeamName
		FROM t_leagues l
		LEFT JOIN t_team_types tt ON tt.teamType = l.leagueTeamTypeId
		LEFT JOIN t_franchise f ON f.leagueGUID = l.GUID
		LEFT JOIN t_seasons ts ON ts.historicalLeagueGUID = l.GUID
		LEFT JOIN t_franchise_seasons fs ON fs.franchiseId = f.franchiseId
		LEFT JOIN t_teams pt ON pt.GUID = f.playerTeamGUID
		GROUP BY l.leagueId, l.leagueName, l.leagueTeamTypeId, f.franchiseId
	`)
	if err != nil {
		return nil, fmt.Errorf("querying leagues: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var leagues []models.SaveGameLeague
	for rows.Next() {
		var lg models.SaveGameLeague
		var franchiseID sql.NullInt64
		var elimination int
		if err := rows.Scan(
			&lg.GUID, &lg.ID, &lg.Name, &lg.TeamTypeID, &lg.TypeName,
			&franchiseID, &elimination, &lg.NumSeasons, &lg.PlayerTeamName,
		); err != nil {
			return nil, fmt.Errorf("scanning league: %w", err)
		}
		if franchiseID.Valid {
			id := int(franchiseID.Int64)
			lg.FranchiseID = &id
		}
		lg.Elimination = elimination == 1
		leagues = append(leagues, lg)
	}
	return leagues, rows.Err()
}

func (r *SqliteSaveGameReader) GetCurrentSeason(ctx context.Context, leagueGUID string) (models.SaveGameSeasonInfo, error) {
	var row *sql.Row
	if leagueGUID == "" {
		// SMB3 / single-league: no league filter, just get the most recent season.
		row = r.db.QueryRowContext(ctx, `
			SELECT seasonID, RANK() OVER (ORDER BY seasonID) AS seasonNum
			FROM t_franchise_seasons
			ORDER BY seasonID DESC
			LIMIT 1
		`)
	} else {
		row = r.db.QueryRowContext(ctx, `
			SELECT fs.seasonID, RANK() OVER (ORDER BY fs.seasonID) AS seasonNum
			FROM t_franchise_seasons fs
			JOIN t_franchise f ON f.franchiseId = fs.franchiseId
			WHERE hex(f.leagueGUID) = ?
			ORDER BY fs.seasonID DESC
			LIMIT 1
		`, leagueGUID)
	}
	var info models.SaveGameSeasonInfo
	if err := row.Scan(&info.SeasonID, &info.SeasonNum); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.SaveGameSeasonInfo{}, fmt.Errorf("no seasons found in save game")
		}
		return models.SaveGameSeasonInfo{}, fmt.Errorf("detecting current season: %w", err)
	}
	return info, nil
}

func (r *SqliteSaveGameReader) GetFranchiseSeasons(ctx context.Context, leagueGUID string) ([]models.SaveGameFranchiseSeason, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			fs.seasonID,
			RANK() OVER (PARTITION BY f.franchiseId ORDER BY fs.seasonID) AS seasonNum
		FROM t_franchise_seasons fs
		JOIN t_franchise f ON f.franchiseId = fs.franchiseId
		WHERE hex(f.leagueGUID) = ?
		ORDER BY fs.seasonID
	`, leagueGUID)
	if err != nil {
		return nil, fmt.Errorf("querying franchise seasons: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var seasons []models.SaveGameFranchiseSeason
	for rows.Next() {
		var s models.SaveGameFranchiseSeason
		if err := rows.Scan(&s.SeasonID, &s.SeasonNum); err != nil {
			return nil, fmt.Errorf("scanning franchise season: %w", err)
		}
		seasons = append(seasons, s)
	}
	return seasons, rows.Err()
}

func (r *SqliteSaveGameReader) GetCurrentSeasonPlayers(ctx context.Context, seasonID int) ([]models.SaveGamePlayer, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			hex(bp.GUID)                                  AS playerGUID,
			? AS seasonID,
			COALESCE(sp.firstName, '')                    AS firstName,
			COALESCE(sp.lastName, '')                     AS lastName,
			COALESCE(sp.primaryPosition, '')              AS primaryPos,
			COALESCE(sp.secondaryPosition, '')            AS secondaryPos,
			COALESCE(sp.pitcherRole, '')                  AS pitcherRole,
			COALESCE(st.currentTeamName, '')              AS currentTeam,
			COALESCE(st.mostRecentTeamName, '')           AS previousTeam,
			COALESCE(bp.power, 0),
			COALESCE(bp.contact, 0),
			COALESCE(bp.speed, 0),
			COALESCE(bp.fielding, 0),
			COALESCE(bp.arm, 0),
			COALESCE(bp.velocity, 0),
			COALESCE(bp.junk, 0),
			COALESCE(bp.accuracy, 0),
			COALESCE(bp.age, 0),
			COALESCE(CAST(s.salary * 200 AS INTEGER), 0) AS salary,
			COALESCE(bpt.traits, '[]')                   AS traits
		FROM t_baseball_players bp
		JOIN t_baseball_player_local_ids bpli ON bpli.GUID = bp.GUID
		JOIN t_stats_players sp ON sp.baseballPlayerGUIDIfKnown = bp.GUID
		JOIN t_stats st ON st.aggregatorID = sp.aggregatorID
		JOIN t_season_stats ss ON ss.aggregatorID = st.aggregatorID AND ss.seasonID = ?
		LEFT JOIN t_salary s ON s.baseballPlayerGUID = bp.GUID
		LEFT JOIN t_baseball_player_traits bpt ON bpt.baseballPlayerGUID = bp.GUID
		ORDER BY sp.lastName, sp.firstName
	`, seasonID, seasonID)
	if err != nil {
		return nil, fmt.Errorf("querying season players: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var players []models.SaveGamePlayer
	for rows.Next() {
		var p models.SaveGamePlayer
		var traitsJSON string
		if err := rows.Scan(
			&p.PlayerGUID, &p.SeasonID,
			&p.FirstName, &p.LastName,
			&p.PrimaryPos, &p.SecondaryPos, &p.PitcherRole,
			&p.CurrentTeam, &p.PreviousTeam,
			&p.Power, &p.Contact, &p.Speed, &p.Fielding, &p.Arm,
			&p.Velocity, &p.Junk, &p.Accuracy,
			&p.Age, &p.Salary, &traitsJSON,
		); err != nil {
			return nil, fmt.Errorf("scanning player: %w", err)
		}
		p.Traits = parseTraitJSON(traitsJSON)
		players = append(players, p)
	}
	return players, rows.Err()
}

func (r *SqliteSaveGameReader) GetCurrentSeasonTeams(ctx context.Context, seasonID int) ([]models.SaveGameTeam, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			tli.localID                                          AS teamLocalID,
			hex(t.GUID)                                         AS teamGUID,
			t.teamName,
			? AS seasonID,
			COALESCE(d.divisionName, '')                        AS divisionName,
			COALESCE(c.conferenceName, '')                      AS conferenceName,
			COALESCE(vs.gamesWon, 0),
			COALESCE(vs.gamesLost, 0),
			COALESCE(vs.gamesBack, 0.0),
			COALESCE(CAST(vs.gamesWon AS REAL) /
				NULLIF(vs.gamesWon + vs.gamesLost, 0), 0.0)    AS winPct,
			COALESCE(vs.runsFor - vs.runsAgainst, 0)           AS runDiff,
			COALESCE(vs.runsFor, 0),
			COALESCE(vs.runsAgainst, 0)
		FROM t_team_local_ids tli
		JOIN t_teams t ON t.GUID = tli.GUID
		LEFT JOIN t_division_teams dt ON dt.teamLocalId = tli.localID
		LEFT JOIN t_divisions d ON d.rowid = dt.divisionId
		LEFT JOIN t_conferences c ON c.rowid = d.conferenceId
		LEFT JOIN v_season_standings vs ON vs.teamLocalId = tli.localID
		ORDER BY t.teamName
	`, seasonID)
	if err != nil {
		return nil, fmt.Errorf("querying season teams: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var teams []models.SaveGameTeam
	for rows.Next() {
		var tm models.SaveGameTeam
		if err := rows.Scan(
			&tm.TeamLocalID, &tm.TeamGUID, &tm.TeamName, &tm.SeasonID,
			&tm.DivisionName, &tm.ConferenceName,
			&tm.Wins, &tm.Losses, &tm.GamesBack, &tm.WinPct,
			&tm.RunDifferential, &tm.RunsFor, &tm.RunsAgainst,
		); err != nil {
			return nil, fmt.Errorf("scanning team: %w", err)
		}
		teams = append(teams, tm)
	}
	return teams, rows.Err()
}

func (r *SqliteSaveGameReader) GetSeasonSchedule(ctx context.Context, seasonID int) ([]models.SaveGameGame, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			sg.seasonID,
			sc.gameNumber,
			sc.day,
			sc.homeTeamID,
			hex(ht.GUID)                             AS homeTeamGUID,
			COALESCE(ht.teamName, '')                AS homeTeamName,
			sc.awayTeamID,
			hex(at_.GUID)                            AS awayTeamGUID,
			COALESCE(at_.teamName, '')               AS awayTeamName,
			gr.homeRunsScored,
			gr.awayRunsScored,
			gr.homePitcherLocalID,
			hex(hpbp.GUID),
			COALESCE(hpsp.firstName || ' ' || hpsp.lastName, NULL),
			gr.awayPitcherLocalID,
			hex(apbp.GUID),
			COALESCE(apsp.firstName || ' ' || apsp.lastName, NULL)
		FROM t_season_games sg
		JOIN t_season_schedule sc ON sc.gameNumber = sg.gameNumber
		JOIN t_team_local_ids htli ON htli.localID = sc.homeTeamID
		JOIN t_teams ht ON ht.GUID = htli.GUID
		JOIN t_team_local_ids atli ON atli.localID = sc.awayTeamID
		JOIN t_teams at_ ON at_.GUID = atli.GUID
		LEFT JOIN t_game_results gr ON gr.gameNumber = sc.gameNumber
		LEFT JOIN t_baseball_player_local_ids hpbpli ON hpbpli.localID = gr.homePitcherLocalID
		LEFT JOIN t_baseball_players hpbp ON hpbp.GUID = hpbpli.GUID
		LEFT JOIN t_stats_players hpsp ON hpsp.baseballPlayerGUIDIfKnown = hpbp.GUID
		LEFT JOIN t_baseball_player_local_ids apbpli ON apbpli.localID = gr.awayPitcherLocalID
		LEFT JOIN t_baseball_players apbp ON apbp.GUID = apbpli.GUID
		LEFT JOIN t_stats_players apsp ON apsp.baseballPlayerGUIDIfKnown = apbp.GUID
		WHERE sg.seasonID = ?
		ORDER BY sc.gameNumber
	`, seasonID)
	if err != nil {
		return nil, fmt.Errorf("querying season schedule: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanGames(rows)
}

func (r *SqliteSaveGameReader) GetPlayoffSchedule(ctx context.Context, seasonID int) ([]models.SaveGamePlayoffGame, error) {
	// Get all playoff series for this season, then find the season games that
	// involve those teams. We match by team membership in the playoff series
	// since the save game links t_playoffs to seasons via a GUID chain that
	// varies between SMB3 and SMB4. Matching on team GUID involvement is more
	// robust across both versions.
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			? AS seasonID,
			ps.seriesNumber,
			hex(t1.GUID), COALESCE(t1.teamName, ''), ps.team1Standing,
			hex(t2.GUID), COALESCE(t2.teamName, ''), ps.team2Standing,
			sc.gameNumber,
			hex(ht.GUID), COALESCE(ht.teamName, ''),
			hex(at_.GUID), COALESCE(at_.teamName, ''),
			gr.homeRunsScored, gr.awayRunsScored,
			hex(hpbp.GUID), COALESCE(hpsp.firstName || ' ' || hpsp.lastName, NULL),
			hex(apbp.GUID), COALESCE(apsp.firstName || ' ' || apsp.lastName, NULL)
		FROM t_playoff_series ps
		JOIN t_teams t1 ON t1.GUID = ps.team1GUID
		JOIN t_teams t2 ON t2.GUID = ps.team2GUID
		JOIN t_team_local_ids t1li ON t1li.GUID = ps.team1GUID
		JOIN t_team_local_ids t2li ON t2li.GUID = ps.team2GUID
		JOIN t_season_games sg ON sg.seasonID = ?
		JOIN t_season_schedule sc ON sc.gameNumber = sg.gameNumber
			AND (
				(sc.homeTeamID = t1li.localID AND sc.awayTeamID = t2li.localID) OR
				(sc.homeTeamID = t2li.localID AND sc.awayTeamID = t1li.localID)
			)
		JOIN t_team_local_ids htli ON htli.localID = sc.homeTeamID
		JOIN t_teams ht ON ht.GUID = htli.GUID
		JOIN t_team_local_ids atli ON atli.localID = sc.awayTeamID
		JOIN t_teams at_ ON at_.GUID = atli.GUID
		LEFT JOIN t_game_results gr ON gr.gameNumber = sc.gameNumber
		LEFT JOIN t_baseball_player_local_ids hpbpli ON hpbpli.localID = gr.homePitcherLocalID
		LEFT JOIN t_baseball_players hpbp ON hpbp.GUID = hpbpli.GUID
		LEFT JOIN t_stats_players hpsp ON hpsp.baseballPlayerGUIDIfKnown = hpbp.GUID
		LEFT JOIN t_baseball_player_local_ids apbpli ON apbpli.localID = gr.awayPitcherLocalID
		LEFT JOIN t_baseball_players apbp ON apbp.GUID = apbpli.GUID
		LEFT JOIN t_stats_players apsp ON apsp.baseballPlayerGUIDIfKnown = apbp.GUID
		ORDER BY ps.seriesNumber, sc.gameNumber
	`, seasonID, seasonID)
	if err != nil {
		return nil, fmt.Errorf("querying playoff schedule: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanPlayoffGames(rows)
}

func (r *SqliteSaveGameReader) GetSeasonBattingStats(ctx context.Context, seasonID int) ([]models.SaveGameBattingStat, error) {
	return r.queryBattingStats(ctx,
		`JOIN t_season_stats ss ON ss.aggregatorID = st.aggregatorID AND ss.seasonID = ?`,
		seasonID,
	)
}

func (r *SqliteSaveGameReader) GetCareerBattingStats(ctx context.Context) ([]models.SaveGameBattingStat, error) {
	return r.queryBattingStats(ctx,
		`JOIN t_career_season_stats css ON css.aggregatorID = st.aggregatorID`,
	)
}

func (r *SqliteSaveGameReader) GetSeasonPitchingStats(ctx context.Context, seasonID int) ([]models.SaveGamePitchingStat, error) {
	return r.queryPitchingStats(ctx,
		`JOIN t_season_stats ss ON ss.aggregatorID = st.aggregatorID AND ss.seasonID = ?`,
		seasonID,
	)
}

func (r *SqliteSaveGameReader) GetCareerPitchingStats(ctx context.Context) ([]models.SaveGamePitchingStat, error) {
	return r.queryPitchingStats(ctx,
		`JOIN t_career_season_stats css ON css.aggregatorID = st.aggregatorID`,
	)
}

// ---- private helpers -------------------------------------------------------

// queryBattingStats uses `args ...any` because it passes arguments directly to
// database/sql's QueryContext, which itself accepts `args ...any`. SQL query
// parameters can be int, string, float64, nil, etc. — there is no single
// concrete type for all of them, so the variadic any is appropriate here.
func (r *SqliteSaveGameReader) queryBattingStats(ctx context.Context, joinClause string, args ...any) ([]models.SaveGameBattingStat, error) {
	query := `
		SELECT
			st.aggregatorID,
			COALESCE(hex(sp.baseballPlayerGUIDIfKnown), '') AS playerGUID,
			COALESCE(sp.firstName, ''), COALESCE(sp.lastName, ''),
			COALESCE(st.currentTeamName, ''),
			COALESCE(st.mostRecentTeamName, ''),
			COALESCE(st.secondMostRecentTeamName, ''),
			COALESCE(sp.primaryPosition, ''),
			COALESCE(sp.secondaryPosition, ''),
			COALESCE(sp.pitcherRole, ''),
			COALESCE(sp.age, 0), sp.retirementSeason,
			COALESCE(b.gamesPlayed, 0), COALESCE(b.gamesBatting, 0),
			COALESCE(b.atBats, 0), COALESCE(b.runs, 0), COALESCE(b.hits, 0),
			COALESCE(b.doubles, 0), COALESCE(b.triples, 0),
			COALESCE(b.homeruns, 0), COALESCE(b.rbi, 0),
			COALESCE(b.stolenBases, 0), COALESCE(b.caughtStealing, 0),
			COALESCE(b.baseOnBalls, 0), COALESCE(b.strikeOuts, 0),
			COALESCE(b.hitByPitch, 0), COALESCE(b.sacrificeHits, 0),
			COALESCE(b.sacrificeFlies, 0), COALESCE(b.errors, 0),
			COALESCE(b.passedBalls, 0)
		FROM t_stats st
		JOIN t_stats_players sp ON sp.aggregatorID = st.aggregatorID
		JOIN t_stats_batting b ON b.aggregatorID = st.aggregatorID
		` + joinClause + `
		ORDER BY sp.lastName, sp.firstName`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying batting stats: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanBattingStats(rows)
}

func (r *SqliteSaveGameReader) queryPitchingStats(ctx context.Context, joinClause string, args ...any) ([]models.SaveGamePitchingStat, error) {
	query := `
		SELECT
			st.aggregatorID,
			COALESCE(hex(sp.baseballPlayerGUIDIfKnown), '') AS playerGUID,
			COALESCE(sp.firstName, ''), COALESCE(sp.lastName, ''),
			COALESCE(st.currentTeamName, ''),
			COALESCE(st.mostRecentTeamName, ''),
			COALESCE(st.secondMostRecentTeamName, ''),
			COALESCE(sp.pitcherRole, ''),
			COALESCE(sp.age, 0), sp.retirementSeason,
			COALESCE(p.wins, 0), COALESCE(p.losses, 0),
			COALESCE(p.games, 0), COALESCE(p.gamesStarted, 0),
			COALESCE(p.completeGames, 0), COALESCE(p.totalPitches, 0),
			COALESCE(p.shutouts, 0), COALESCE(p.saves, 0),
			COALESCE(p.outsPitched, 0), COALESCE(p.hits, 0),
			COALESCE(p.earnedRuns, 0), COALESCE(p.homeRuns, 0),
			COALESCE(p.baseOnBalls, 0), COALESCE(p.strikeOuts, 0),
			COALESCE(p.battersHitByPitch, 0), COALESCE(p.battersFaced, 0),
			COALESCE(p.gamesFinished, 0), COALESCE(p.runsAllowed, 0),
			COALESCE(p.wildPitches, 0)
		FROM t_stats st
		JOIN t_stats_players sp ON sp.aggregatorID = st.aggregatorID
		JOIN t_stats_pitching p ON p.aggregatorID = st.aggregatorID
		` + joinClause + `
		ORDER BY sp.lastName, sp.firstName`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying pitching stats: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanPitchingStats(rows)
}

func scanGames(rows *sql.Rows) ([]models.SaveGameGame, error) {
	var games []models.SaveGameGame
	for rows.Next() {
		var g models.SaveGameGame
		if err := rows.Scan(
			&g.SeasonID, &g.GameNumber, &g.Day,
			&g.HomeTeamID, &g.HomeTeamGUID, &g.HomeTeamName,
			&g.AwayTeamID, &g.AwayTeamGUID, &g.AwayTeamName,
			&g.HomeScore, &g.AwayScore,
			&g.HomePitcherID, &g.HomePitcherGUID, &g.HomePitcherName,
			&g.AwayPitcherID, &g.AwayPitcherGUID, &g.AwayPitcherName,
		); err != nil {
			return nil, fmt.Errorf("scanning game: %w", err)
		}
		games = append(games, g)
	}
	return games, rows.Err()
}

func scanPlayoffGames(rows *sql.Rows) ([]models.SaveGamePlayoffGame, error) {
	var games []models.SaveGamePlayoffGame
	for rows.Next() {
		var g models.SaveGamePlayoffGame
		if err := rows.Scan(
			&g.SeasonID, &g.SeriesNum,
			&g.Team1GUID, &g.Team1Name, &g.Team1Seed,
			&g.Team2GUID, &g.Team2Name, &g.Team2Seed,
			&g.GameNumber,
			&g.HomeTeamGUID, &g.HomeTeamName,
			&g.AwayTeamGUID, &g.AwayTeamName,
			&g.HomeScore, &g.AwayScore,
			&g.HomePitcherGUID, &g.HomePitcherName,
			&g.AwayPitcherGUID, &g.AwayPitcherName,
		); err != nil {
			return nil, fmt.Errorf("scanning playoff game: %w", err)
		}
		games = append(games, g)
	}
	return games, rows.Err()
}

func scanBattingStats(rows *sql.Rows) ([]models.SaveGameBattingStat, error) {
	var stats []models.SaveGameBattingStat
	for rows.Next() {
		var s models.SaveGameBattingStat
		if err := rows.Scan(
			&s.AggregatorID, &s.PlayerGUID,
			&s.FirstName, &s.LastName,
			&s.CurrentTeam, &s.PrevTeam, &s.Prev2Team,
			&s.PrimaryPos, &s.SecondaryPos, &s.PitcherRole,
			&s.Age, &s.RetirementSeason,
			&s.GamesPlayed, &s.GamesBatting, &s.AtBats, &s.Runs,
			&s.Hits, &s.Doubles, &s.Triples, &s.HomeRuns, &s.RBI,
			&s.StolenBases, &s.CaughtStealing, &s.Walks, &s.Strikeouts,
			&s.HitByPitch, &s.SacHits, &s.SacFlies, &s.Errors, &s.PassedBalls,
		); err != nil {
			return nil, fmt.Errorf("scanning batting stat: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

func scanPitchingStats(rows *sql.Rows) ([]models.SaveGamePitchingStat, error) {
	var stats []models.SaveGamePitchingStat
	for rows.Next() {
		var s models.SaveGamePitchingStat
		if err := rows.Scan(
			&s.AggregatorID, &s.PlayerGUID,
			&s.FirstName, &s.LastName,
			&s.CurrentTeam, &s.PrevTeam, &s.Prev2Team,
			&s.PitcherRole, &s.Age, &s.RetirementSeason,
			&s.Wins, &s.Losses, &s.Games, &s.GamesStarted,
			&s.CompleteGames, &s.TotalPitches, &s.Shutouts, &s.Saves,
			&s.OutsPitched, &s.HitsAllowed, &s.EarnedRuns, &s.HomeRunsAllowed,
			&s.Walks, &s.Strikeouts, &s.HitBatters, &s.BattersFaced,
			&s.GamesFinished, &s.RunsAllowed, &s.WildPitches,
		); err != nil {
			return nil, fmt.Errorf("scanning pitching stat: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

// parseTraitJSON returns the raw JSON string from t_baseball_player_traits.
// Trait ID → name resolution happens in the service layer.
func parseTraitJSON(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" || raw == "null" {
		return nil
	}
	return []string{raw}
}
