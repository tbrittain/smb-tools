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
	// Column names match the real SMB save game schema (confirmed from SMB3Explorer):
	//   t_leagues: GUID (blob PK), name, allowedTeamType
	//   t_franchise: GUID (blob PK), leagueGUID (blob FK), playerTeamGUID (blob FK)
	//   t_seasons: id (int PK), historicalLeagueGUID (blob FK), elimination
	// Mode logic mirrors SMB3Explorer's LeagueModeExtensions.Parse:
	//   franchise GUID present → franchise
	//   no franchise + elimination flag → elimination
	//   no franchise + seasons played → season
	//   else → none
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			COALESCE(hex(l.GUID), '')              AS leagueGUID,
			l.name,
			l.allowedTeamType,
			COALESCE(tt.typeName, ''),
			f.GUID                                 AS franchiseGUID,
			MAX(COALESCE(ts.elimination, 0))       AS elimination,
			COUNT(DISTINCT ts.id)                  AS numSeasons,
			COALESCE(pt.teamName, '')              AS playerTeamName
		FROM t_leagues l
		LEFT JOIN t_team_types tt ON tt.teamType = l.allowedTeamType
		LEFT JOIN t_franchise f ON f.leagueGUID = l.GUID
		LEFT JOIN t_seasons ts ON ts.historicalLeagueGUID = l.GUID
		LEFT JOIN t_teams pt ON pt.GUID = f.playerTeamGUID
		GROUP BY hex(l.GUID), l.name, l.allowedTeamType, hex(f.GUID)
	`)
	if err != nil {
		return nil, fmt.Errorf("querying leagues: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var leagues []models.SaveGameLeague
	for rows.Next() {
		var lg models.SaveGameLeague
		var franchiseGUID sql.NullString
		var elimination int
		if err := rows.Scan(
			&lg.GUID, &lg.Name, &lg.AllowedTeamType, &lg.TypeName,
			&franchiseGUID, &elimination, &lg.NumSeasons, &lg.PlayerTeamName,
		); err != nil {
			return nil, fmt.Errorf("scanning league: %w", err)
		}
		switch {
		case franchiseGUID.Valid && franchiseGUID.String != "":
			lg.Mode = models.LeagueModeFranchise
		case elimination == 1:
			lg.Mode = models.LeagueModeElimination
		case lg.NumSeasons > 0:
			lg.Mode = models.LeagueModeSeason
		default:
			lg.Mode = models.LeagueModeNone
		}
		leagues = append(leagues, lg)
	}
	return leagues, rows.Err()
}

func (r *SqliteSaveGameReader) GetCurrentSeason(ctx context.Context, leagueGUID string) (models.SaveGameSeasonInfo, error) {
	// Use t_seasons.id as the season key — it's the integer PK referenced by
	// t_season_stats.seasonID throughout the save game. RANK() over id gives the
	// human-facing season number. Matches SMB3Explorer's FranchiseSeasons.sql.
	var row *sql.Row
	if leagueGUID == "" {
		// SMB3 / single-league: no league filter needed.
		row = r.db.QueryRowContext(ctx, `
			SELECT ts.id, RANK() OVER (ORDER BY ts.id) AS seasonNum
			FROM t_seasons ts
			ORDER BY ts.id DESC
			LIMIT 1
		`)
	} else {
		row = r.db.QueryRowContext(ctx, `
			SELECT ts.id, RANK() OVER (ORDER BY ts.id) AS seasonNum
			FROM t_seasons ts
			JOIN t_leagues tl ON ts.historicalLeagueGUID = tl.GUID
			WHERE hex(tl.GUID) = ?
			ORDER BY ts.id DESC
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
	// Mirrors SMB3Explorer's FranchiseSeasons.sql: join t_seasons → t_leagues by GUID.
	// t_seasons.id is the integer PK used throughout the save game as the season key.
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			ts.id,
			RANK() OVER (ORDER BY ts.id) AS seasonNum
		FROM t_seasons ts
		JOIN t_leagues tl ON ts.historicalLeagueGUID = tl.GUID
		WHERE hex(tl.GUID) = ?
		ORDER BY ts.id
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
			COALESCE(vbpi.firstName,       sp.firstName, '')  AS firstName,
			COALESCE(vbpi.lastName,        sp.lastName,  '')  AS lastName,
			COALESCE(vbpi.primaryPosition, sp.primaryPos,'')  AS primaryPos,
			COALESCE(sp.secondaryPos, '')                     AS secondaryPos,
			COALESCE(vbpi.pitcherRole,     sp.pitcherRole,'') AS pitcherRole,
			COALESCE(ct.teamName, '')                         AS currentTeam,
			COALESCE(mrt.teamName, '')                        AS previousTeam,
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
			COALESCE(
				(SELECT json_group_array(json_object('traitId', t.trait, 'subtypeId', t.subType))
				 FROM t_baseball_player_traits t
				 WHERE t.baseballPlayerLocalID = bpli.localID),
				'[]'
			) AS traits
		FROM t_baseball_players bp
		JOIN t_baseball_player_local_ids bpli ON bpli.GUID = bp.GUID
		JOIN t_stats_players sp ON sp.baseballPlayerLocalID = bpli.localID
		JOIN t_stats st ON st.statsPlayerID = sp.statsPlayerID
		JOIN t_season_stats ss ON ss.aggregatorID = st.aggregatorID AND ss.seasonID = ?
		LEFT JOIN t_team_local_ids ctli  ON ctli.localID  = st.currentTeamLocalID
		LEFT JOIN t_teams ct             ON ct.GUID        = ctli.GUID
		LEFT JOIN t_team_local_ids mrtli ON mrtli.localID = st.mostRecentlyPlayedTeamLocalID
		LEFT JOIN t_teams mrt            ON mrt.GUID       = mrtli.GUID
		LEFT JOIN v_baseball_player_info vbpi ON vbpi.baseballPlayerGUID = bpli.GUID
		LEFT JOIN t_salary s ON s.baseballPlayerGUID = bp.GUID
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
			COALESCE(d.name, '')                                AS divisionName,
			COALESCE(c.name, '')                                AS conferenceName,
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
		LEFT JOIN t_division_teams dt ON dt.teamGUID = t.GUID
		LEFT JOIN t_divisions d ON d.GUID = dt.divisionGUID
		LEFT JOIN t_conferences c ON c.GUID = d.conferenceGUID
		LEFT JOIN v_season_standings vs ON vs.teamGUID = t.GUID
			AND vs.seasonGUID = (SELECT ts.GUID FROM t_seasons ts WHERE ts.id = ?)
		ORDER BY t.teamName
	`, seasonID, seasonID)
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
	// In the real game t_season_schedule has no gameNumber or day columns — those
	// are computed via RANK/ROW_NUMBER in SMB3Explorer. We derive game number from
	// t_game_results.ID order and set day to 0 (not critical for the import).
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			sg.seasonID,
			RANK() OVER (ORDER BY gr.ID)             AS gameNumber,
			0                                        AS day,
			gr.homeTeamLocalID                       AS homeTeamID,
			hex(htli.GUID)                           AS homeTeamGUID,
			COALESCE(ht.teamName, '')                AS homeTeamName,
			gr.awayTeamLocalID                       AS awayTeamID,
			hex(atli.GUID)                           AS awayTeamGUID,
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
		JOIN t_game_results gr ON gr.ID = sg.gameID
		JOIN t_team_local_ids htli ON htli.localID = gr.homeTeamLocalID
		JOIN t_teams ht ON ht.GUID = htli.GUID
		JOIN t_team_local_ids atli ON atli.localID = gr.awayTeamLocalID
		JOIN t_teams at_ ON at_.GUID = atli.GUID
		LEFT JOIN t_baseball_player_local_ids hpbpli ON hpbpli.localID = gr.homePitcherLocalID
		LEFT JOIN t_baseball_players hpbp ON hpbp.GUID = hpbpli.GUID
		LEFT JOIN t_stats_players hpsp ON hpsp.baseballPlayerLocalID = hpbpli.localID
		LEFT JOIN t_baseball_player_local_ids apbpli ON apbpli.localID = gr.awayPitcherLocalID
		LEFT JOIN t_baseball_players apbp ON apbp.GUID = apbpli.GUID
		LEFT JOIN t_stats_players apsp ON apsp.baseballPlayerLocalID = apbpli.localID
		WHERE sg.seasonID = ?
		ORDER BY gr.ID
	`, seasonID)
	if err != nil {
		return nil, fmt.Errorf("querying season schedule: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanGames(rows)
}

func (r *SqliteSaveGameReader) GetPlayoffSchedule(ctx context.Context, seasonID int) ([]models.SaveGamePlayoffGame, error) {
	// Playoff games are linked via t_playoffs.seasonGUID → t_seasons.GUID,
	// then t_playoff_games.gameID → t_game_results.ID (not via t_season_games).
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			ts.id                                                    AS seasonID,
			tps.seriesNumber,
			hex(tps.team1GUID), COALESCE(t1.teamName, ''), tps.team1Standing,
			hex(tps.team2GUID), COALESCE(t2.teamName, ''), tps.team2Standing,
			RANK() OVER (PARTITION BY tps.seriesNumber ORDER BY gr.ID) AS gameNumber,
			hex(htli.GUID), COALESCE(ht.teamName, ''),
			hex(atli.GUID), COALESCE(at_.teamName, ''),
			gr.homeRunsScored, gr.awayRunsScored,
			hex(hpbp.GUID), COALESCE(hpsp.firstName || ' ' || hpsp.lastName, NULL),
			hex(apbp.GUID), COALESCE(apsp.firstName || ' ' || apsp.lastName, NULL)
		FROM t_playoffs tp
		JOIN t_seasons ts ON ts.GUID = tp.seasonGUID
		JOIN t_playoff_series tps ON tps.playoffGUID = tp.GUID
		JOIN t_playoff_games tpg ON tpg.playoffGUID = tp.GUID
			AND tpg.seriesNumber = tps.seriesNumber
		JOIN t_game_results gr ON gr.ID = tpg.gameID
		JOIN t_teams t1 ON t1.GUID = tps.team1GUID
		JOIN t_teams t2 ON t2.GUID = tps.team2GUID
		JOIN t_team_local_ids htli ON htli.localID = gr.homeTeamLocalID
		JOIN t_teams ht ON ht.GUID = htli.GUID
		JOIN t_team_local_ids atli ON atli.localID = gr.awayTeamLocalID
		JOIN t_teams at_ ON at_.GUID = atli.GUID
		LEFT JOIN t_baseball_player_local_ids hpbpli ON hpbpli.localID = gr.homePitcherLocalID
		LEFT JOIN t_baseball_players hpbp ON hpbp.GUID = hpbpli.GUID
		LEFT JOIN t_stats_players hpsp ON hpsp.baseballPlayerLocalID = hpbpli.localID
		LEFT JOIN t_baseball_player_local_ids apbpli ON apbpli.localID = gr.awayPitcherLocalID
		LEFT JOIN t_baseball_players apbp ON apbp.GUID = apbpli.GUID
		LEFT JOIN t_stats_players apsp ON apsp.baseballPlayerLocalID = apbpli.localID
		WHERE ts.id = ?
		ORDER BY tps.seriesNumber, gr.ID
	`, seasonID)
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
			COALESCE(hex(bpli.GUID), '') AS playerGUID,
			COALESCE(vbpi.firstName, sp.firstName, ''), COALESCE(vbpi.lastName, sp.lastName, ''),
			COALESCE(ct.teamName, ''),
			COALESCE(mrt.teamName, ''),
			COALESCE(pmrt.teamName, ''),
			COALESCE(sp.primaryPos, ''),
			COALESCE(sp.secondaryPos, ''),
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
		JOIN t_stats_players sp ON sp.statsPlayerID = st.statsPlayerID
		JOIN t_stats_batting b ON b.aggregatorID = st.aggregatorID
		LEFT JOIN t_baseball_player_local_ids bpli  ON bpli.localID  = sp.baseballPlayerLocalID
		LEFT JOIN v_baseball_player_info vbpi        ON vbpi.baseballPlayerGUID = bpli.GUID
		LEFT JOIN t_team_local_ids ctli             ON ctli.localID  = st.currentTeamLocalID
		LEFT JOIN t_teams ct                        ON ct.GUID        = ctli.GUID
		LEFT JOIN t_team_local_ids mrtli            ON mrtli.localID = st.mostRecentlyPlayedTeamLocalID
		LEFT JOIN t_teams mrt                       ON mrt.GUID       = mrtli.GUID
		LEFT JOIN t_team_local_ids pmrtli           ON pmrtli.localID = st.previousRecentlyPlayedTeamLocalID
		LEFT JOIN t_teams pmrt                      ON pmrt.GUID      = pmrtli.GUID
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
			COALESCE(hex(bpli.GUID), '') AS playerGUID,
			COALESCE(vbpi.firstName, sp.firstName, ''), COALESCE(vbpi.lastName, sp.lastName, ''),
			COALESCE(ct.teamName, ''),
			COALESCE(mrt.teamName, ''),
			COALESCE(pmrt.teamName, ''),
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
		JOIN t_stats_players sp ON sp.statsPlayerID = st.statsPlayerID
		JOIN t_stats_pitching p ON p.aggregatorID = st.aggregatorID
		LEFT JOIN t_baseball_player_local_ids bpli  ON bpli.localID  = sp.baseballPlayerLocalID
		LEFT JOIN v_baseball_player_info vbpi        ON vbpi.baseballPlayerGUID = bpli.GUID
		LEFT JOIN t_team_local_ids ctli             ON ctli.localID  = st.currentTeamLocalID
		LEFT JOIN t_teams ct                        ON ct.GUID        = ctli.GUID
		LEFT JOIN t_team_local_ids mrtli            ON mrtli.localID = st.mostRecentlyPlayedTeamLocalID
		LEFT JOIN t_teams mrt                       ON mrt.GUID       = mrtli.GUID
		LEFT JOIN t_team_local_ids pmrtli           ON pmrtli.localID = st.previousRecentlyPlayedTeamLocalID
		LEFT JOIN t_teams pmrt                      ON pmrt.GUID      = pmrtli.GUID
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
