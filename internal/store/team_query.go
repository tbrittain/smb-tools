package store

import (
	"context"
	"database/sql"
	"fmt"
	"math/bits"

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

// ListAllTeams returns all teams with their most recent name, ordered by name.
// Used to populate the team filter picker on the Export page.
func (s *TeamQueryStore) ListAllTeams(ctx context.Context) ([]models.TeamSearchResult, error) {
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
GROUP BY t.id
ORDER BY current_name`)
	if err != nil {
		return nil, fmt.Errorf("ListAllTeams: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.TeamSearchResult
	for rows.Next() {
		var r models.TeamSearchResult
		if err := rows.Scan(&r.TeamID, &r.TeamName, &r.Seasons, &r.FirstSeason, &r.LastSeason); err != nil {
			return nil, fmt.Errorf("ListAllTeams scan: %w", err)
		}
		out = append(out, r)
	}
	if out == nil {
		out = []models.TeamSearchResult{}
	}
	return out, rows.Err()
}

// ListAllTeamSeasons returns every team-season row ordered by season descending,
// then team name ascending. Each row includes a champion flag.
func (s *TeamQueryStore) ListAllTeamSeasons(ctx context.Context) ([]models.TeamSeasonListRow, error) {
	q := `
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
LEFT JOIN season_champions c ON c.season_id = tsh.season_id
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

// GetHistoricalTeams returns one aggregated row per team covering the given
// inclusive season range. Stats are summed across all seasons in the range;
// rate stats (BA, ERA) and GamesOver500 are computed in Go after scanning.
func (s *TeamQueryStore) GetHistoricalTeams(ctx context.Context, seasonStart, seasonEnd int) ([]models.HistoricalTeamRow, error) {
	q := `
WITH
conf_champs AS (
    SELECT DISTINCT tsh.team_id, tsh.season_id
    FROM (
        SELECT season_id, MAX(series_number) AS max_series
        FROM team_playoff_schedules
        GROUP BY season_id
    ) fs
    JOIN team_playoff_schedules tps ON tps.season_id = fs.season_id AND tps.series_number = fs.max_series
    JOIN team_season_history tsh ON tsh.id = tps.home_team_history_id OR tsh.id = tps.away_team_history_id
    JOIN seasons s ON s.id = tsh.season_id
    WHERE s.season_num BETWEEN ? AND ?
),
batting_agg AS (
    SELECT
        tsh.team_id,
        SUM(COALESCE(b.at_bats, 0))    AS total_ab,
        SUM(COALESCE(b.hits, 0))        AS total_hits,
        SUM(COALESCE(b.home_runs, 0))   AS total_hr,
        COUNT(DISTINCT ps.player_id)    AS num_players,
        COUNT(DISTINCT CASE WHEN p.is_hall_of_famer = 1 THEN ps.player_id END) AS num_hof
    FROM team_season_history tsh
    JOIN seasons s ON s.id = tsh.season_id
    JOIN player_season_teams pst ON pst.team_history_id = tsh.id AND pst.sort_order = 0
    JOIN player_seasons ps ON ps.id = pst.player_season_id
    JOIN players p ON p.id = ps.player_id
    LEFT JOIN player_season_batting_stats b ON b.player_season_id = ps.id AND b.is_regular_season = 1
    WHERE s.season_num BETWEEN ? AND ?
    GROUP BY tsh.team_id
),
pitching_agg AS (
    SELECT
        tsh.team_id,
        SUM(COALESCE(pit.earned_runs, 0))  AS total_er,
        SUM(COALESCE(pit.outs_pitched, 0)) AS total_outs
    FROM team_season_history tsh
    JOIN seasons s ON s.id = tsh.season_id
    JOIN player_season_teams pst ON pst.team_history_id = tsh.id AND pst.sort_order = 0
    JOIN player_seasons ps ON ps.id = pst.player_season_id
    LEFT JOIN player_season_pitching_stats pit ON pit.player_season_id = ps.id AND pit.is_regular_season = 1
    WHERE s.season_num BETWEEN ? AND ?
    GROUP BY tsh.team_id
)
SELECT
    t.id AS team_id,
    (
        SELECT tsh2.team_name
        FROM team_season_history tsh2
        JOIN seasons s2 ON s2.id = tsh2.season_id
        WHERE tsh2.team_id = t.id AND s2.season_num BETWEEN ? AND ?
        ORDER BY s2.season_num DESC
        LIMIT 1
    ) AS current_name,
    COUNT(DISTINCT tsh.id)                                           AS num_seasons,
    MIN(s.season_num)                                                AS first_season,
    MAX(s.season_num)                                                AS last_season,
    SUM(tsh.wins)                                                    AS wins,
    SUM(tsh.losses)                                                  AS losses,
    SUM(COALESCE(tsh.playoff_wins, 0))                               AS playoff_wins,
    SUM(COALESCE(tsh.playoff_losses, 0))                             AS playoff_losses,
    COUNT(CASE WHEN tsh.playoff_seed IS NOT NULL THEN 1 END)         AS playoff_appearances,
    COUNT(CASE WHEN tsh.games_back = 0 THEN 1 END)                   AS division_titles,
    COUNT(DISTINCT CASE WHEN cc.season_id IS NOT NULL
                        THEN tsh.season_id END)                      AS conference_titles,
    COUNT(CASE WHEN c.winner_history_id = tsh.id THEN 1 END)         AS championships,
    SUM(tsh.runs_for + COALESCE(tsh.playoff_runs_for, 0))            AS runs_for,
    SUM(tsh.runs_against + COALESCE(tsh.playoff_runs_against, 0))    AS runs_against,
    COALESCE(ba.total_ab, 0)                                         AS total_ab,
    COALESCE(ba.total_hits, 0)                                       AS total_hits,
    COALESCE(ba.total_hr, 0)                                         AS total_hr,
    COALESCE(ba.num_players, 0)                                      AS num_players,
    COALESCE(ba.num_hof, 0)                                          AS num_hof,
    COALESCE(pit.total_er, 0)                                        AS total_er,
    COALESCE(pit.total_outs, 0)                                      AS total_outs,
    -- Championship drought: seasons since last title (or since the franchise began)
    (
        SELECT COALESCE(MAX(s2.season_num), 0)
        FROM seasons s2
        WHERE s2.season_num <= ?
    ) - COALESCE(
        (
            SELECT MAX(s3.season_num)
            FROM team_season_history tsh3
            JOIN seasons s3 ON s3.id = tsh3.season_id
            WHERE tsh3.team_id = t.id
              AND s3.season_num <= ?
              AND tsh3.id IN (SELECT winner_history_id FROM season_champions)
        ), 0
    )                                                                AS championship_drought,
    -- Only populated when the query covers exactly one season
    CASE WHEN COUNT(DISTINCT tsh.id) = 1 THEN MIN(tsh.id) ELSE NULL END AS history_id
FROM teams t
JOIN team_season_history tsh ON tsh.team_id = t.id
JOIN seasons s ON s.id = tsh.season_id
LEFT JOIN season_champions c ON c.season_id = tsh.season_id
LEFT JOIN conf_champs cc ON cc.team_id = tsh.team_id AND cc.season_id = tsh.season_id
LEFT JOIN batting_agg ba ON ba.team_id = t.id
LEFT JOIN pitching_agg pit ON pit.team_id = t.id
WHERE s.season_num BETWEEN ? AND ?
GROUP BY t.id
ORDER BY SUM(tsh.wins) DESC
`
	rows, err := s.db.QueryContext(ctx, q,
		seasonStart, seasonEnd, // conf_champs BETWEEN
		seasonStart, seasonEnd, // batting_agg BETWEEN
		seasonStart, seasonEnd, // pitching_agg BETWEEN
		seasonStart, seasonEnd, // current_name subquery BETWEEN
		seasonEnd,              // drought: max season_num in range
		seasonEnd,              // drought: last champ season for this team
		seasonStart, seasonEnd, // main WHERE BETWEEN
	)
	if err != nil {
		return nil, fmt.Errorf("getting historical teams: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.HistoricalTeamRow
	for rows.Next() {
		var r models.HistoricalTeamRow
		if err := rows.Scan(
			&r.TeamID, &r.TeamName,
			&r.NumSeasons, &r.FirstSeason, &r.LastSeason,
			&r.Wins, &r.Losses,
			&r.PlayoffWins, &r.PlayoffLosses,
			&r.PlayoffAppearances,
			&r.DivisionTitles, &r.ConferenceTitles, &r.Championships,
			&r.RunsFor, &r.RunsAgainst,
			&r.TotalAB, &r.TotalHits, &r.TotalHR,
			&r.NumPlayers, &r.NumHoF,
			&r.TotalEarnedRuns, &r.TotalOutsPitched,
			&r.ChampionshipDrought,
			&r.HistoryID,
		); err != nil {
			return nil, fmt.Errorf("scanning historical team row: %w", err)
		}

		total := r.Wins + r.Losses
		if total > 0 {
			r.WinPct = float64(r.Wins) / float64(total)
		}
		r.GamesOver500 = r.Wins - r.Losses

		if r.TotalAB > 0 {
			ba := float64(r.TotalHits) / float64(r.TotalAB)
			r.BA = &ba
		}
		if r.TotalOutsPitched > 0 {
			era := float64(r.TotalEarnedRuns) / float64(r.TotalOutsPitched) * 27.0
			r.ERA = &era
		}

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

	q := `
SELECT
    tsh.id,
    s.id  AS season_id,
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
LEFT JOIN season_champions c ON c.season_id = tsh.season_id
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
			&ts.HistoryID, &ts.SeasonID, &ts.SeasonNum, &ts.TeamName,
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
    pst.sort_order,
    -- batting sentinel first, then the rest
    b.at_bats,
    b.games_played, b.games_batting, b.runs, b.hits,
    b.doubles, b.triples, b.home_runs, b.rbi,
    b.stolen_bases, b.caught_stealing, b.walks, b.strikeouts,
    b.hit_by_pitch, b.sac_hits, b.sac_flies, b.errors, b.passed_balls,
    b.ba, b.obp, b.slg, b.ops, b.iso, b.babip, b.k_pct, b.bb_pct, b.ab_per_hr,
    b.ops_plus, b.smb_war,
    -- pitching sentinel first, then the rest
    pit.outs_pitched,
    pit.wins, pit.losses, pit.games, pit.games_started,
    pit.complete_games, pit.shutouts, pit.saves,
    pit.hits_allowed, pit.earned_runs, pit.home_runs_allowed,
    pit.walks, pit.strikeouts, pit.hit_batters, pit.batters_faced,
    pit.games_finished, pit.runs_allowed, pit.wild_pitches, pit.total_pitches,
    pit.era, pit.whip, pit.k_per_9, pit.bb_per_9, pit.h_per_9, pit.hr_per_9,
    pit.k_per_bb, pit.k_pct, pit.win_pct, pit.p_per_ip,
    pit.era_plus, pit.fip, pit.fip_minus, pit.smb_war
FROM player_seasons ps
JOIN players p ON p.id = ps.player_id
JOIN player_season_teams pst ON pst.player_season_id = ps.id AND pst.team_history_id = ?
LEFT JOIN player_season_game_stats gs ON gs.player_season_id = ps.id
LEFT JOIN player_season_batting_stats b
    ON b.player_season_id = ps.id AND b.is_regular_season = 1
LEFT JOIN player_season_pitching_stats pit
    ON pit.player_season_id = ps.id AND pit.is_regular_season = 1
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
		var bBA, bOBP, bSLG, bOPS, bISO, bBABIP, bKPct, bBBPct, bABPerHR sql.NullFloat64
		var bOPSPlus, bSmbWAR sql.NullFloat64

		var pOuts sql.NullInt64
		var pW, pL, pG, pGS, pCG, pSHO, pSV sql.NullInt64
		var pH, pER, pHRA, pWalks, pK, pHBP, pBF, pGF, pRA, pWP, pTP sql.NullInt64
		var pERA, pWHIP, pK9, pBB9, pH9, pHR9, pKPerBB, pPKPct, pWinPct, pPPerIP sql.NullFloat64
		var pERAPlus, pFIP, pFIPMinus, pSmbWAR sql.NullFloat64

		if err := rows.Scan(
			&r.PlayerID, &r.FirstName, &r.LastName, &hof,
			&r.Age, &r.Salary,
			&r.PrimaryPosition, &r.SecondaryPosition, &r.PitcherRole,
			&r.BatHand, &r.ThrowHand, &r.ChemistryType,
			&r.TraitsJSON, &r.PitchesJSON,
			&r.Power, &r.Contact, &r.Speed, &r.Fielding, &r.Arm,
			&r.Velocity, &r.Junk, &r.Accuracy,
			&r.SortOrder,
			&bAtBats,
			&bGP, &bGB, &bRuns, &bHits, &bDB, &bTR, &bHR, &bRBI,
			&bSB, &bCS, &bWalks, &bK, &bHBP, &bSH, &bSF, &bE, &bPB,
			&bBA, &bOBP, &bSLG, &bOPS, &bISO, &bBABIP, &bKPct, &bBBPct, &bABPerHR,
			&bOPSPlus, &bSmbWAR,
			&pOuts,
			&pW, &pL, &pG, &pGS, &pCG, &pSHO, &pSV,
			&pH, &pER, &pHRA, &pWalks, &pK, &pHBP, &pBF, &pGF, &pRA, &pWP, &pTP,
			&pERA, &pWHIP, &pK9, &pBB9, &pH9, &pHR9, &pKPerBB, &pPKPct, &pWinPct, &pPPerIP,
			&pERAPlus, &pFIP, &pFIPMinus, &pSmbWAR,
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
				BA:      nullFloat64Ptr(bBA),
				OBP:     nullFloat64Ptr(bOBP),
				SLG:     nullFloat64Ptr(bSLG),
				OPS:     nullFloat64Ptr(bOPS),
				ISO:     nullFloat64Ptr(bISO),
				BABIP:   nullFloat64Ptr(bBABIP),
				KPct:    nullFloat64Ptr(bKPct),
				BBPct:   nullFloat64Ptr(bBBPct),
				ABPerHR: nullFloat64Ptr(bABPerHR),
				OPSPlus: nullFloat64Ptr(bOPSPlus),
				SmbWAR:  nullFloat64Ptr(bSmbWAR),
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
				ERA:      nullFloat64Ptr(pERA),
				WHIP:     nullFloat64Ptr(pWHIP),
				K9:       nullFloat64Ptr(pK9),
				BB9:      nullFloat64Ptr(pBB9),
				H9:       nullFloat64Ptr(pH9),
				HR9:      nullFloat64Ptr(pHR9),
				KPerBB:   nullFloat64Ptr(pKPerBB),
				KPct:     nullFloat64Ptr(pPKPct),
				WinPct:   nullFloat64Ptr(pWinPct),
				PPerIP:   nullFloat64Ptr(pPPerIP),
				ERAPlus:  nullFloat64Ptr(pERAPlus),
				FIP:      nullFloat64Ptr(pFIP),
				FIPMinus: nullFloat64Ptr(pFIPMinus),
				SmbWAR:   nullFloat64Ptr(pSmbWAR),
			}
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetTeamSeasonSummaryByHistoryID returns the TeamSeasonSummary for a specific
// team_season_history row, including the champion flag.
// Returns sql.ErrNoRows wrapped in an error if the history ID does not exist.
func (s *TeamQueryStore) GetTeamSeasonSummaryByHistoryID(ctx context.Context, historyID int64) (models.TeamSeasonSummary, error) {
	q := `
SELECT
    tsh.id,
    s.id  AS season_id,
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
LEFT JOIN season_champions c ON c.season_id = tsh.season_id
WHERE tsh.id = ?
`
	var ts models.TeamSeasonSummary
	var seed, pw, pl, prf, pra sql.NullInt64
	var isChampion int
	err := s.db.QueryRowContext(ctx, q, historyID).Scan(
		&ts.HistoryID, &ts.SeasonID, &ts.SeasonNum, &ts.TeamName,
		&ts.DivisionName, &ts.ConferenceName,
		&ts.Wins, &ts.Losses, &ts.GamesBack, &ts.RunsFor, &ts.RunsAgainst,
		&ts.Budget, &ts.Payroll,
		&seed, &pw, &pl, &prf, &pra,
		&ts.TotalPower, &ts.TotalContact, &ts.TotalSpeed,
		&ts.TotalFielding, &ts.TotalArm,
		&ts.TotalVelocity, &ts.TotalJunk, &ts.TotalAccuracy,
		&isChampion,
	)
	if err != nil {
		return ts, fmt.Errorf("getting team season summary for history %d: %w", historyID, err)
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
	return ts, nil
}

// GetTeamSeasonSchedule returns all regular season games (home and away) for a
// team in a given season, ordered by game number. TeamGameNum is the 1-based
// sequential index of each game within this team's schedule.
func (s *TeamQueryStore) GetTeamSeasonSchedule(ctx context.Context, teamHistoryID int64, seasonID int64) ([]models.ScheduleGameRow, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT
    ROW_NUMBER() OVER (ORDER BY g.game_number) AS team_game_num,
    g.game_number,
    g.day,
    g.home_team_history_id,
    home.team_name AS home_team_name,
    home.team_id   AS home_team_id,
    g.away_team_history_id,
    away.team_name AS away_team_name,
    away.team_id   AS away_team_id,
    g.home_score,
    g.away_score,
    COALESCE(hp.first_name || ' ' || hp.last_name, '') AS home_pitcher,
    COALESCE(ap.first_name || ' ' || ap.last_name, '') AS away_pitcher,
    hp.id AS home_pitcher_player_id,
    ap.id AS away_pitcher_player_id
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
// season. RoundNumber and RoundLabel are computed from the full set of playoff
// series in the season so that a team eliminated in round 1 of a 4-round bracket
// still sees RoundNumber=1 / RoundLabel="Round of 16".
//nolint:gocognit // two correlated queries (series numbers → round mapping, then game rows) with nullable field scan; splitting queries would lose the single round-mapping pass
func (s *TeamQueryStore) GetTeamSeasonPlayoffSchedule(ctx context.Context, teamHistoryID int64, seasonID int64) ([]models.PlayoffGameRow, error) {
	// Step 1: fetch all distinct series numbers for the season so we can map each
	// to its correct bracket round regardless of which team we're viewing.
	seriesRows, err := s.db.QueryContext(ctx,
		`SELECT DISTINCT series_number FROM team_playoff_schedules WHERE season_id = ? ORDER BY series_number ASC`,
		seasonID)
	if err != nil {
		return nil, fmt.Errorf("fetching playoff series numbers for season %d: %w", seasonID, err)
	}
	defer func() { _ = seriesRows.Close() }()

	var seriesNumbers []int
	for seriesRows.Next() {
		var n int
		if err := seriesRows.Scan(&n); err != nil {
			return nil, fmt.Errorf("scanning series number: %w", err)
		}
		seriesNumbers = append(seriesNumbers, n)
	}
	if err := seriesRows.Err(); err != nil {
		return nil, fmt.Errorf("iterating series numbers: %w", err)
	}
	_ = seriesRows.Close()

	// If playoff_rounds is configured for this season (> 0), use it to derive
	// the true bracket size so round labels are correct even mid-playoff.
	var dbRounds int64
	_ = s.db.QueryRowContext(ctx,
		`SELECT playoff_rounds FROM seasons WHERE id = ?`, seasonID,
	).Scan(&dbRounds)

	totalSeries := len(seriesNumbers)
	if dbRounds > 0 {
		totalSeries = (1 << int(dbRounds)) - 1
	}
	// Map opaque series_number → (roundNumber, roundLabel).
	type roundInfo struct {
		number int
		label  string
	}
	seriesRoundInfo := make(map[int]roundInfo, totalSeries)
	for rank, sn := range seriesNumbers {
		n, l := PlayoffRoundInfo(rank+1, totalSeries)
		seriesRoundInfo[sn] = roundInfo{number: n, label: l}
	}

	// Step 2: fetch this team's playoff games.
	rows, err := s.db.QueryContext(ctx, `
SELECT
    g.series_number,
    g.game_number,
    g.home_team_history_id,
    home.team_name AS home_team_name,
    home.team_id   AS home_team_id,
    g.away_team_history_id,
    away.team_name AS away_team_name,
    away.team_id   AS away_team_id,
    g.home_score,
    g.away_score,
    COALESCE(hp.first_name || ' ' || hp.last_name, '') AS home_pitcher,
    COALESCE(ap.first_name || ' ' || ap.last_name, '') AS away_pitcher,
    hp.id AS home_pitcher_player_id,
    ap.id AS away_pitcher_player_id
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
		var seriesNumber int
		var homeScore, awayScore sql.NullInt64
		var homePitcherID, awayPitcherID sql.NullInt64
		if err := rows.Scan(
			&seriesNumber, &r.GameNumber,
			&r.HomeTeamHistoryID, &r.HomeTeamName, &r.HomeTeamID,
			&r.AwayTeamHistoryID, &r.AwayTeamName, &r.AwayTeamID,
			&homeScore, &awayScore,
			&r.HomePitcherName, &r.AwayPitcherName,
			&homePitcherID, &awayPitcherID,
		); err != nil {
			return nil, fmt.Errorf("scanning playoff game row: %w", err)
		}
		if ri, ok := seriesRoundInfo[seriesNumber]; ok {
			r.RoundNumber = ri.number
			r.RoundLabel = ri.label
		}
		if homeScore.Valid {
			v := int(homeScore.Int64)
			r.HomeScore = &v
		}
		if awayScore.Valid {
			v := int(awayScore.Int64)
			r.AwayScore = &v
		}
		if homePitcherID.Valid {
			v := homePitcherID.Int64
			r.HomePitcherPlayerID = &v
		}
		if awayPitcherID.Valid {
			v := awayPitcherID.Int64
			r.AwayPitcherPlayerID = &v
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// PlayoffRoundInfo maps a series' rank (1-based, sorted ascending by series_number)
// within a bracket of totalSeries series to its round number and display label.
//
// The bracket structure is assumed to be single-elimination with power-of-2 team
// counts (4, 8, 16, ...), so totalSeries = 2^R - 1 for R rounds.
//
// Algorithm: bits.Len gives floor(log2(n))+1. The distance from the
// championship ("fromTop") = bits.Len(totalSeries+1-rank) - 1, and
// roundNumber = totalRounds - fromTop.
func PlayoffRoundInfo(rank, totalSeries int) (roundNumber int, roundLabel string) {
	if totalSeries <= 0 {
		return 1, "League Championship"
	}
	totalRounds := bits.Len(uint(totalSeries))
	fromTop := bits.Len(uint(totalSeries+1-rank)) - 1
	roundNumber = totalRounds - fromTop
	switch fromTop {
	case 0:
		roundLabel = "League Championship"
	case 1:
		roundLabel = "Conference Championship"
	default:
		roundLabel = fmt.Sprintf("Round of %d", 1<<(fromTop+1))
	}
	return
}

func nullFloat64Ptr(n sql.NullFloat64) *float64 {
	if !n.Valid {
		return nil
	}
	v := n.Float64
	return &v
}

func scanScheduleRows(rows *sql.Rows) ([]models.ScheduleGameRow, error) {
	var out []models.ScheduleGameRow
	for rows.Next() {
		var r models.ScheduleGameRow
		var homeScore, awayScore sql.NullInt64
		var homePitcherID, awayPitcherID sql.NullInt64
		if err := rows.Scan(
			&r.TeamGameNum, &r.GameNumber, &r.Day,
			&r.HomeTeamHistoryID, &r.HomeTeamName, &r.HomeTeamID,
			&r.AwayTeamHistoryID, &r.AwayTeamName, &r.AwayTeamID,
			&homeScore, &awayScore,
			&r.HomePitcherName, &r.AwayPitcherName,
			&homePitcherID, &awayPitcherID,
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
		if homePitcherID.Valid {
			v := homePitcherID.Int64
			r.HomePitcherPlayerID = &v
		}
		if awayPitcherID.Valid {
			v := awayPitcherID.Int64
			r.AwayPitcherPlayerID = &v
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetTeamTopPlayers returns the top `limit` players by cumulative smbWAR
// accumulated while playing for the given team, across all seasons. A player
// who was traded mid-season contributes fully to both the team they left and
// the team they joined — granular split stats are not stored.
func (s *TeamQueryStore) GetTeamTopPlayers(ctx context.Context, teamID int64, limit int) ([]models.TeamTopPlayer, error) {
	rows, err := s.db.QueryContext(ctx, `
WITH player_team_seasons AS (
    SELECT DISTINCT
        ps.player_id,
        ps.id          AS player_season_id,
        ps.season_id,
        ps.pitcher_role
    FROM player_seasons ps
    JOIN player_season_teams pst ON pst.player_season_id = ps.id
    JOIN team_season_history  tsh ON tsh.id = pst.team_history_id
    WHERE tsh.team_id = ?
)
SELECT
    p.id,
    p.first_name,
    p.last_name,
    p.is_hall_of_famer,
    COUNT(DISTINCT pts.season_id)                                       AS num_seasons,
    GROUP_CONCAT(s.season_num)                                          AS season_nums_csv,
    MAX(CASE WHEN pts.pitcher_role != '' THEN 1 ELSE 0 END)            AS is_pitcher,
    (
        SELECT ps2.primary_position
        FROM player_team_seasons pts2
        JOIN player_seasons ps2 ON ps2.id = pts2.player_season_id
        JOIN seasons        s2  ON s2.id  = pts2.season_id
        WHERE pts2.player_id = p.id
        ORDER BY s2.season_num DESC
        LIMIT 1
    )                                                                   AS primary_position,
    COALESCE(SUM(b.smb_war),   0) +
    COALESCE(SUM(pit.smb_war), 0)                                       AS total_smb_war,
    AVG(CASE WHEN b.ops_plus   IS NOT NULL THEN b.ops_plus   END)       AS avg_ops_plus,
    AVG(CASE WHEN pit.era_plus IS NOT NULL THEN pit.era_plus END)       AS avg_era_plus
FROM player_team_seasons pts
JOIN players p ON p.id = pts.player_id
JOIN seasons  s ON s.id = pts.season_id
LEFT JOIN player_season_batting_stats  b
    ON b.player_season_id   = pts.player_season_id AND b.is_regular_season = 1
LEFT JOIN player_season_pitching_stats pit
    ON pit.player_season_id = pts.player_season_id AND pit.is_regular_season = 1
GROUP BY p.id, p.first_name, p.last_name, p.is_hall_of_famer
ORDER BY total_smb_war DESC
LIMIT ?
`, teamID, limit)
	if err != nil {
		return nil, fmt.Errorf("getting top players for team %d: %w", teamID, err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.TeamTopPlayer
	for rows.Next() {
		var r models.TeamTopPlayer
		var hof, isPitcher int
		var seasonNumsCSV sql.NullString
		var avgOpsPlus, avgEraPlus sql.NullFloat64
		if err := rows.Scan(
			&r.PlayerID, &r.FirstName, &r.LastName, &hof,
			&r.NumSeasons, &seasonNumsCSV, &isPitcher, &r.PrimaryPosition,
			&r.TotalSmbWAR, &avgOpsPlus, &avgEraPlus,
		); err != nil {
			return nil, fmt.Errorf("scanning team top player: %w", err)
		}
		r.IsHallOfFamer = hof == 1
		r.IsPitcher = isPitcher == 1
		r.SeasonNumsCSV = seasonNumsCSV.String
		if avgOpsPlus.Valid {
			v := avgOpsPlus.Float64
			r.AvgOpsPlus = &v
		}
		if avgEraPlus.Valid {
			v := avgEraPlus.Float64
			r.AvgEraPlus = &v
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating team top players: %w", err)
	}

	if len(out) == 0 {
		return out, nil
	}

	// Second query: awards scoped to each player's seasons with this team.
	awardRows, err := s.db.QueryContext(ctx, `
SELECT
    ps.player_id,
    a.original_name
FROM player_season_awards psa
JOIN player_seasons        ps  ON ps.id  = psa.player_season_id
JOIN awards                a   ON a.id   = psa.award_id
JOIN player_season_teams   pst ON pst.player_season_id = psa.player_season_id
JOIN team_season_history   tsh ON tsh.id = pst.team_history_id
WHERE tsh.team_id = ?
ORDER BY ps.player_id ASC, a.importance ASC, a.name ASC
`, teamID)
	if err != nil {
		return nil, fmt.Errorf("getting awards for team %d top players: %w", teamID, err)
	}
	defer func() { _ = awardRows.Close() }()

	awardsByPlayer := make(map[int64][]string)
	for awardRows.Next() {
		var playerID int64
		var originalName string
		if err := awardRows.Scan(&playerID, &originalName); err != nil {
			return nil, fmt.Errorf("scanning team top player award: %w", err)
		}
		awardsByPlayer[playerID] = append(awardsByPlayer[playerID], originalName)
	}
	if err := awardRows.Err(); err != nil {
		return nil, fmt.Errorf("iterating team top player awards: %w", err)
	}

	for i := range out {
		if awards, ok := awardsByPlayer[out[i].PlayerID]; ok {
			out[i].Awards = awards
		}
	}
	return out, nil
}
