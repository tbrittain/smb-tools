package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"smb-tools/internal/models"
)

// PlayerQueryStore provides read-only queries over player and career stat data.
type PlayerQueryStore struct {
	db DBTX
}

func NewPlayerQueryStore(db DBTX) *PlayerQueryStore {
	return &PlayerQueryStore{db: db}
}

// SearchPlayers returns up to 50 players whose first name, last name, or full
// name matches the query string (case-insensitive LIKE).
func (s *PlayerQueryStore) SearchPlayers(ctx context.Context, query string) ([]models.PlayerSearchResult, error) {
	pattern := "%" + query + "%"
	rows, err := s.db.QueryContext(ctx, `
SELECT
    p.id,
    p.first_name,
    p.last_name,
    p.is_hall_of_famer,
    COUNT(DISTINCT ps.season_id)  AS seasons_played,
    MIN(s.season_num)             AS first_season,
    MAX(s.season_num)             AS last_season
FROM players p
JOIN player_seasons ps ON ps.player_id = p.id
JOIN seasons s         ON s.id = ps.season_id
WHERE p.first_name LIKE ?
   OR p.last_name  LIKE ?
   OR (p.first_name || ' ' || p.last_name) LIKE ?
GROUP BY p.id
ORDER BY p.last_name, p.first_name
LIMIT 50
`, pattern, pattern, pattern)
	if err != nil {
		return nil, fmt.Errorf("searching players %q: %w", query, err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.PlayerSearchResult
	for rows.Next() {
		var r models.PlayerSearchResult
		var hof int
		if err := rows.Scan(
			&r.PlayerID, &r.FirstName, &r.LastName, &hof,
			&r.SeasonsPlayed, &r.FirstSeason, &r.LastSeason,
		); err != nil {
			return nil, fmt.Errorf("scanning player search result: %w", err)
		}
		r.IsHallOfFamer = hof == 1
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetPlayerCareer returns the player's bio and combined career totals (regular season + playoffs).
// Reads from pre-computed career tables (stat_type='total_career') — no on-read rate computation.
// Returns sql.ErrNoRows wrapped in an error if the player does not exist.
//nolint:gocognit // two independent queries (batting + pitching career tables) each with ~12 nullable field assignments; the dual-query structure is a schema constraint, not reducible without splitting queries
func (s *PlayerQueryStore) GetPlayerCareer(ctx context.Context, playerID int64) (models.PlayerCareer, error) {
	var c models.PlayerCareer
	var hof int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, first_name, last_name, is_hall_of_famer FROM players WHERE id = ?`,
		playerID,
	).Scan(&c.PlayerID, &c.FirstName, &c.LastName, &hof)
	if err != nil {
		return c, fmt.Errorf("getting player %d: %w", playerID, err)
	}
	c.IsHallOfFamer = hof == 1

	b := &models.CareerBattingStats{}
	var bBA, bOBP, bSLG, bOPS, bISO, bBABIP, bKPct, bBBPct, bABPerHR sql.NullFloat64
	var bOPSPlus, bSmbWAR sql.NullFloat64
	err = s.db.QueryRowContext(ctx, `
SELECT
    games_played, games_batting, at_bats, runs, hits,
    doubles, triples, home_runs, rbi, stolen_bases, caught_stealing,
    walks, strikeouts, hit_by_pitch, sac_hits, sac_flies, errors, passed_balls,
    ba, obp, slg, ops, iso, babip, k_pct, bb_pct, ab_per_hr,
    ops_plus, smb_war
FROM player_career_batting_stats
WHERE player_id = ? AND stat_type = 'total_career'
`, playerID).Scan(
		&b.GamesPlayed, &b.GamesBatting, &b.AtBats, &b.Runs, &b.Hits,
		&b.Doubles, &b.Triples, &b.HomeRuns, &b.RBI, &b.StolenBases, &b.CaughtStealing,
		&b.Walks, &b.Strikeouts, &b.HitByPitch, &b.SacHits, &b.SacFlies, &b.Errors, &b.PassedBalls,
		&bBA, &bOBP, &bSLG, &bOPS, &bISO, &bBABIP, &bKPct, &bBBPct, &bABPerHR,
		&bOPSPlus, &bSmbWAR,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return c, fmt.Errorf("getting career batting for player %d: %w", playerID, err)
	}
	if err == nil {
		if bBA.Valid      { b.BA      = &bBA.Float64 }
		if bOBP.Valid     { b.OBP     = &bOBP.Float64 }
		if bSLG.Valid     { b.SLG     = &bSLG.Float64 }
		if bOPS.Valid     { b.OPS     = &bOPS.Float64 }
		if bISO.Valid     { b.ISO     = &bISO.Float64 }
		if bBABIP.Valid   { b.BABIP   = &bBABIP.Float64 }
		if bKPct.Valid    { b.KPct    = &bKPct.Float64 }
		if bBBPct.Valid   { b.BBPct   = &bBBPct.Float64 }
		if bABPerHR.Valid { b.ABPerHR = &bABPerHR.Float64 }
		if bOPSPlus.Valid { b.OPSPlus = &bOPSPlus.Float64 }
		if bSmbWAR.Valid  { b.SmbWAR  = &bSmbWAR.Float64 }
		if b.AtBats > 0 || b.GamesPlayed > 0 {
			c.Batting = b
		}
	}

	p := &models.CareerPitchingStats{}
	var pERA, pWHIP, pK9, pBB9, pH9, pHR9, pKPerBB, pKPct, pWinPct, pPPerIP sql.NullFloat64
	var pERAPlus, pFIP, pFIPMinus, pSmbWAR sql.NullFloat64
	err = s.db.QueryRowContext(ctx, `
SELECT
    wins, losses, games, games_started, complete_games, shutouts, saves,
    outs_pitched, hits_allowed, earned_runs, home_runs_allowed, walks, strikeouts,
    hit_batters, batters_faced, games_finished, runs_allowed, wild_pitches, total_pitches,
    era, whip, k_per_9, bb_per_9, h_per_9, hr_per_9, k_per_bb, k_pct, win_pct, p_per_ip,
    era_plus, fip, fip_minus, smb_war
FROM player_career_pitching_stats
WHERE player_id = ? AND stat_type = 'total_career'
`, playerID).Scan(
		&p.Wins, &p.Losses, &p.Games, &p.GamesStarted, &p.CompleteGames, &p.Shutouts, &p.Saves,
		&p.OutsPitched, &p.HitsAllowed, &p.EarnedRuns, &p.HomeRunsAllowed, &p.Walks, &p.Strikeouts,
		&p.HitBatters, &p.BattersFaced, &p.GamesFinished, &p.RunsAllowed, &p.WildPitches, &p.TotalPitches,
		&pERA, &pWHIP, &pK9, &pBB9, &pH9, &pHR9, &pKPerBB, &pKPct, &pWinPct, &pPPerIP,
		&pERAPlus, &pFIP, &pFIPMinus, &pSmbWAR,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return c, fmt.Errorf("getting career pitching for player %d: %w", playerID, err)
	}
	if err == nil {
		if pERA.Valid     { p.ERA     = &pERA.Float64 }
		if pWHIP.Valid    { p.WHIP    = &pWHIP.Float64 }
		if pK9.Valid      { p.K9      = &pK9.Float64 }
		if pBB9.Valid     { p.BB9     = &pBB9.Float64 }
		if pH9.Valid      { p.H9      = &pH9.Float64 }
		if pHR9.Valid     { p.HR9     = &pHR9.Float64 }
		if pKPerBB.Valid  { p.KPerBB  = &pKPerBB.Float64 }
		if pKPct.Valid    { p.KPct    = &pKPct.Float64 }
		if pWinPct.Valid  { p.WinPct  = &pWinPct.Float64 }
		if pPPerIP.Valid  { p.PPerIP  = &pPPerIP.Float64 }
		if pERAPlus.Valid  { p.ERAPlus  = &pERAPlus.Float64 }
		if pFIP.Valid      { p.FIP      = &pFIP.Float64 }
		if pFIPMinus.Valid { p.FIPMinus = &pFIPMinus.Float64 }
		if pSmbWAR.Valid   { p.SmbWAR   = &pSmbWAR.Float64 }
		if p.OutsPitched > 0 || p.Games > 0 {
			c.Pitching = p
		}
	}

	return c, nil
}

// GetPlayerSeasonLog returns the player's season-by-season regular and playoff
// stats, ordered by season number ascending. Rate fields are computed on each row.
func (s *PlayerQueryStore) GetPlayerSeasonLog(ctx context.Context, playerID int64) ([]models.PlayerSeasonLogRow, error) {
	out, seasonIndex, err := s.scanSeasonLogRows(ctx, playerID, 1)
	if err != nil {
		return nil, err
	}

	// Load team associations for each player_season (one query for all seasons).
	if len(out) > 0 {
		psIDs := make([]int64, len(out))
		for i, r := range out {
			psIDs[i] = r.PlayerSeasonID
		}
		teamsMap, err := s.loadSeasonTeams(ctx, psIDs)
		if err != nil {
			return nil, err
		}
		for i := range out {
			if t, ok := teamsMap[out[i].PlayerSeasonID]; ok {
				out[i].Teams = t
			} else {
				out[i].Teams = []models.PlayerTeamRef{}
			}
		}
	}

	// Fetch playoff rows and merge by season_id.
	playoffOut, _, err := s.scanSeasonLogRows(ctx, playerID, 0)
	if err != nil {
		return nil, err
	}
	for _, pr := range playoffOut {
		if idx, ok := seasonIndex[pr.SeasonID]; ok {
			out[idx].PlayoffBatting = pr.Batting
			out[idx].PlayoffPitching = pr.Pitching
		}
	}

	return out, nil
}

// loadSeasonTeams fetches all team associations for the given player_season IDs,
// returning a map from player_season_id to ordered team slice.
func (s *PlayerQueryStore) loadSeasonTeams(ctx context.Context, psIDs []int64) (map[int64][]models.PlayerTeamRef, error) {
	if len(psIDs) == 0 {
		return nil, nil
	}
	placeholders := strings.Repeat("?,", len(psIDs))
	placeholders = placeholders[:len(placeholders)-1] // trim trailing comma
	args := make([]any, len(psIDs))
	for i, id := range psIDs {
		args[i] = id
	}
	//nolint:gosec // placeholder count is controlled internally, not from user input
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT pst.player_season_id, tsh.team_id, tsh.id, tsh.team_name, pst.sort_order
		FROM player_season_teams pst
		JOIN team_season_history tsh ON tsh.id = pst.team_history_id
		WHERE pst.player_season_id IN (%s)
		ORDER BY pst.player_season_id, pst.sort_order
	`, placeholders), args...)
	if err != nil {
		return nil, fmt.Errorf("loading season teams: %w", err)
	}
	defer func() { _ = rows.Close() }()

	result := map[int64][]models.PlayerTeamRef{}
	for rows.Next() {
		var psID int64
		var ref models.PlayerTeamRef
		if err := rows.Scan(&psID, &ref.TeamID, &ref.TeamHistoryID, &ref.TeamName, &ref.SortOrder); err != nil {
			return nil, fmt.Errorf("scanning season team: %w", err)
		}
		result[psID] = append(result[psID], ref)
	}
	return result, rows.Err()
}

// scanSeasonLogRows scans either regular (isRegularSeason=1) or playoff
// (isRegularSeason=0) rows for a player. Returns the rows and a map from
// season_id to slice index (used by the caller to merge playoff stats).
//nolint:gocognit // single scan handles batting × pitching × regular/playoff in one pass; sentinel-based detection of optional stat objects is correct — splitting into multiple queries loses the single-pass guarantee
func (s *PlayerQueryStore) scanSeasonLogRows(
	ctx context.Context, playerID int64, isRegularSeason int,
) ([]models.PlayerSeasonLogRow, map[int64]int, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT
    ps.id                             AS player_season_id,
    s.id                              AS season_id,
    s.season_num,
    ps.age,
    ps.salary,
    ps.primary_position,
    ps.secondary_position,
    ps.pitcher_role,
    ps.bat_hand,
    ps.throw_hand,
    ps.chemistry_type,
    ps.traits_json,
    ps.pitches_json,
    COALESCE(gs.power,0),
    COALESCE(gs.contact,0),
    COALESCE(gs.speed,0),
    COALESCE(gs.fielding,0),
    COALESCE(gs.arm,0),
    COALESCE(gs.velocity,0),
    COALESCE(gs.junk,0),
    COALESCE(gs.accuracy,0),
    -- batting block (all NULL when no batting row for this season type)
    b.at_bats,
    b.games_played, b.games_batting,
    b.runs, b.hits, b.doubles, b.triples, b.home_runs, b.rbi,
    b.stolen_bases, b.caught_stealing, b.walks, b.strikeouts,
    b.hit_by_pitch, b.sac_hits, b.sac_flies, b.errors, b.passed_balls,
    b.ba, b.obp, b.slg, b.ops, b.iso, b.babip, b.k_pct, b.bb_pct, b.ab_per_hr,
    b.ops_plus, b.smb_war,
    -- pitching block (all NULL when no pitching row for this season type)
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
JOIN seasons s ON s.id = ps.season_id
LEFT JOIN player_season_game_stats gs ON gs.player_season_id = ps.id
LEFT JOIN player_season_batting_stats b
    ON b.player_season_id = ps.id AND b.is_regular_season = ?
LEFT JOIN player_season_pitching_stats pit
    ON pit.player_season_id = ps.id AND pit.is_regular_season = ?
WHERE ps.player_id = ?
ORDER BY s.season_num ASC
`, isRegularSeason, isRegularSeason, playerID)
	if err != nil {
		return nil, nil, fmt.Errorf("scanning season log (reg=%d) for player %d: %w", isRegularSeason, playerID, err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.PlayerSeasonLogRow
	index := map[int64]int{}

	for rows.Next() {
		var row models.PlayerSeasonLogRow

		// Batting sentinel: at_bats is NULL when there is no batting row.
		var bAtBats sql.NullInt64
		var bGP, bGB, bRuns, bHits, bDB, bTR, bHR, bRBI sql.NullInt64
		var bSB, bCS, bWalks, bK, bHBP, bSH, bSF, bE, bPB sql.NullInt64
		var bBA, bOBP, bSLG, bOPS, bISO, bBABIP, bKPct, bBBPct, bABPerHR sql.NullFloat64
		var bOPSPlus, bSmbWAR sql.NullFloat64

		// Pitching sentinel: outs_pitched is NULL when there is no pitching row.
		var pOuts sql.NullInt64
		var pW, pL, pG, pGS, pCG, pSHO, pSV sql.NullInt64
		var pH, pER, pHRA, pWalks, pK, pHBP, pBF, pGF, pRA, pWP, pTP sql.NullInt64
		var pERA, pWHIP, pK9, pBB9, pH9, pHR9, pKPerBB, pPKPct, pWinPct, pPPerIP sql.NullFloat64
		var pERAPlus, pFIP, pFIPMinus, pSmbWAR sql.NullFloat64

		if err := rows.Scan(
			&row.PlayerSeasonID, &row.SeasonID, &row.SeasonNum,
			&row.Age, &row.Salary,
			&row.PrimaryPosition, &row.SecondaryPosition, &row.PitcherRole,
			&row.BatHand, &row.ThrowHand, &row.ChemistryType,
			&row.TraitsJSON, &row.PitchesJSON,
			&row.Power, &row.Contact, &row.Speed, &row.Fielding, &row.Arm,
			&row.Velocity, &row.Junk, &row.Accuracy,
			// batting block
			&bAtBats,
			&bGP, &bGB, &bRuns, &bHits, &bDB, &bTR, &bHR, &bRBI,
			&bSB, &bCS, &bWalks, &bK, &bHBP, &bSH, &bSF, &bE, &bPB,
			&bBA, &bOBP, &bSLG, &bOPS, &bISO, &bBABIP, &bKPct, &bBBPct, &bABPerHR,
			&bOPSPlus, &bSmbWAR,
			// pitching block
			&pOuts,
			&pW, &pL, &pG, &pGS, &pCG, &pSHO, &pSV,
			&pH, &pER, &pHRA, &pWalks, &pK, &pHBP, &pBF, &pGF, &pRA, &pWP, &pTP,
			&pERA, &pWHIP, &pK9, &pBB9, &pH9, &pHR9, &pKPerBB, &pPKPct, &pWinPct, &pPPerIP,
			&pERAPlus, &pFIP, &pFIPMinus, &pSmbWAR,
		); err != nil {
			return nil, nil, fmt.Errorf("scanning season log row: %w", err)
		}

		if bAtBats.Valid {
			b := &models.CareerBattingStats{
				AtBats:         int(bAtBats.Int64),
				GamesPlayed:    int(bGP.Int64), GamesBatting: int(bGB.Int64),
				Runs: int(bRuns.Int64), Hits: int(bHits.Int64),
				Doubles: int(bDB.Int64), Triples: int(bTR.Int64), HomeRuns: int(bHR.Int64),
				RBI: int(bRBI.Int64), StolenBases: int(bSB.Int64), CaughtStealing: int(bCS.Int64),
				Walks: int(bWalks.Int64), Strikeouts: int(bK.Int64), HitByPitch: int(bHBP.Int64),
				SacHits: int(bSH.Int64), SacFlies: int(bSF.Int64),
				Errors: int(bE.Int64), PassedBalls: int(bPB.Int64),
			}
			if bBA.Valid      { b.BA      = &bBA.Float64 }
			if bOBP.Valid     { b.OBP     = &bOBP.Float64 }
			if bSLG.Valid     { b.SLG     = &bSLG.Float64 }
			if bOPS.Valid     { b.OPS     = &bOPS.Float64 }
			if bISO.Valid     { b.ISO     = &bISO.Float64 }
			if bBABIP.Valid   { b.BABIP   = &bBABIP.Float64 }
			if bKPct.Valid    { b.KPct    = &bKPct.Float64 }
			if bBBPct.Valid   { b.BBPct   = &bBBPct.Float64 }
			if bABPerHR.Valid { b.ABPerHR = &bABPerHR.Float64 }
			if bOPSPlus.Valid { b.OPSPlus = &bOPSPlus.Float64 }
			if bSmbWAR.Valid  { b.SmbWAR  = &bSmbWAR.Float64 }
			row.Batting = b
		}

		if pOuts.Valid {
			p := &models.CareerPitchingStats{
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
			if pERA.Valid    { p.ERA    = &pERA.Float64 }
			if pWHIP.Valid   { p.WHIP   = &pWHIP.Float64 }
			if pK9.Valid     { p.K9     = &pK9.Float64 }
			if pBB9.Valid    { p.BB9    = &pBB9.Float64 }
			if pH9.Valid     { p.H9     = &pH9.Float64 }
			if pHR9.Valid    { p.HR9    = &pHR9.Float64 }
			if pKPerBB.Valid { p.KPerBB = &pKPerBB.Float64 }
			if pPKPct.Valid  { p.KPct   = &pPKPct.Float64 }
			if pWinPct.Valid { p.WinPct = &pWinPct.Float64 }
			if pPPerIP.Valid { p.PPerIP = &pPPerIP.Float64 }
			if pERAPlus.Valid  { p.ERAPlus  = &pERAPlus.Float64 }
			if pFIP.Valid      { p.FIP      = &pFIP.Float64 }
			if pFIPMinus.Valid { p.FIPMinus = &pFIPMinus.Float64 }
			if pSmbWAR.Valid   { p.SmbWAR   = &pSmbWAR.Float64 }
			row.Pitching = p
		}

		index[row.SeasonID] = len(out)
		out = append(out, row)
	}
	return out, index, rows.Err()
}

// PlayerAttributeHistoryRow holds one season's attribute snapshot for the
// trend chart.
//
// Percentile fields are nil when the player is the only one in the relevant
// comparison group (PERCENT_RANK is undefined for a single-row partition).
//
// Universal stats (power/contact/speed/fielding) carry both a league-wide
// percentile (*Pct) and a role-specific one (*PctRole). Role-exclusive stats
// (arm, velocity, junk, accuracy) have only a role-specific percentile.
//
// RoleAvg* fields carry the average for the player's own role group and are
// used as the reference-line value when comparing vs. own role.
type PlayerAttributeHistoryRow struct {
	SeasonID    int64
	SeasonNum   int
	Power       int
	Contact     int
	Speed       int
	Fielding    int
	Arm         int
	Velocity    int
	Junk        int
	Accuracy    int
	// League-wide percentiles (universal stats only)
	PowerPct    *float64
	ContactPct  *float64
	SpeedPct    *float64
	FieldingPct *float64
	// Role-specific percentiles (all 8 stats)
	ArmPct          *float64
	VelocityPct     *float64
	JunkPct         *float64
	AccuracyPct     *float64
	PowerPctRole    *float64
	ContactPctRole  *float64
	SpeedPctRole    *float64
	FieldingPctRole *float64
	// League-wide averages
	LgAvgPower    float64
	LgAvgContact  float64
	LgAvgSpeed    float64
	LgAvgFielding float64
	LgAvgArm      float64
	LgAvgVelocity float64
	LgAvgJunk     float64
	LgAvgAccuracy float64
	// Role-specific averages (batter or pitcher depending on player's role)
	RoleAvgPower    float64
	RoleAvgContact  float64
	RoleAvgSpeed    float64
	RoleAvgFielding float64
	RoleAvgArm      float64
	RoleAvgVelocity float64
	RoleAvgJunk     float64
	RoleAvgAccuracy float64
}

// GetPlayerAttributeHistory returns one row per season for the given player,
// ordered by season number ascending. Each row includes raw attribute values,
// pre-computed league-wide and role-specific percentile ranks, league-wide
// averages, and role-specific averages.
//
// All percentile and average data is read from eagerly-populated tables so
// this query is a plain indexed JOIN with no window functions at read time.
//nolint:gocognit // 12 nullable float assignments (one per percentile column that is NULL for single-player seasons) count as branches — splitting the scan into separate queries would lose the single-round-trip guarantee
func (s *PlayerQueryStore) GetPlayerAttributeHistory(ctx context.Context, playerID int64) ([]PlayerAttributeHistoryRow, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT
    ps.season_id,
    s.season_num,
    COALESCE(psg.power,    0),
    COALESCE(psg.contact,  0),
    COALESCE(psg.speed,    0),
    COALESCE(psg.fielding, 0),
    COALESCE(psg.arm,      0),
    COALESCE(psg.velocity, 0),
    COALESCE(psg.junk,     0),
    COALESCE(psg.accuracy, 0),
    -- league-wide percentiles (universal stats)
    psap.power_pct, psap.contact_pct, psap.speed_pct, psap.fielding_pct,
    -- role-specific percentiles (role-exclusive stats + role variants of universal)
    psap.arm_pct, psap.velocity_pct, psap.junk_pct, psap.accuracy_pct,
    psap.power_pct_role, psap.contact_pct_role, psap.speed_pct_role, psap.fielding_pct_role,
    -- league-wide averages
    COALESCE(saa.avg_power,    0),
    COALESCE(saa.avg_contact,  0),
    COALESCE(saa.avg_speed,    0),
    COALESCE(saa.avg_fielding, 0),
    COALESCE(saa.avg_arm,      0),
    COALESCE(saa.avg_velocity, 0),
    COALESCE(saa.avg_junk,     0),
    COALESCE(saa.avg_accuracy, 0),
    -- role-specific averages: pick batter or pitcher avg based on this player's role
    COALESCE(CASE WHEN ps.pitcher_role != '' THEN saa.pitcher_avg_power    ELSE saa.batter_avg_power    END, 0),
    COALESCE(CASE WHEN ps.pitcher_role != '' THEN saa.pitcher_avg_contact  ELSE saa.batter_avg_contact  END, 0),
    COALESCE(CASE WHEN ps.pitcher_role != '' THEN saa.pitcher_avg_speed    ELSE saa.batter_avg_speed    END, 0),
    COALESCE(CASE WHEN ps.pitcher_role != '' THEN saa.pitcher_avg_fielding ELSE saa.batter_avg_fielding END, 0),
    COALESCE(saa.batter_avg_arm,      0),
    COALESCE(saa.pitcher_avg_velocity, 0),
    COALESCE(saa.pitcher_avg_junk,     0),
    COALESCE(saa.pitcher_avg_accuracy, 0)
FROM player_season_game_stats psg
JOIN player_seasons ps  ON ps.id  = psg.player_season_id
JOIN seasons s          ON s.id   = ps.season_id
LEFT JOIN player_season_attribute_percentiles psap ON psap.player_season_id = psg.player_season_id
LEFT JOIN season_attribute_averages saa            ON saa.season_id = ps.season_id
WHERE ps.player_id = ?
ORDER BY s.season_num ASC
`, playerID)
	if err != nil {
		return nil, fmt.Errorf("querying attribute history for player %d: %w", playerID, err)
	}
	defer func() { _ = rows.Close() }()

	var out []PlayerAttributeHistoryRow
	for rows.Next() {
		var r PlayerAttributeHistoryRow
		var powerPct, contactPct, speedPct, fieldingPct sql.NullFloat64
		var armPct, velocityPct, junkPct, accuracyPct sql.NullFloat64
		var powerPctRole, contactPctRole, speedPctRole, fieldingPctRole sql.NullFloat64
		if err := rows.Scan(
			&r.SeasonID, &r.SeasonNum,
			&r.Power, &r.Contact, &r.Speed, &r.Fielding, &r.Arm,
			&r.Velocity, &r.Junk, &r.Accuracy,
			&powerPct, &contactPct, &speedPct, &fieldingPct,
			&armPct, &velocityPct, &junkPct, &accuracyPct,
			&powerPctRole, &contactPctRole, &speedPctRole, &fieldingPctRole,
			&r.LgAvgPower, &r.LgAvgContact, &r.LgAvgSpeed, &r.LgAvgFielding, &r.LgAvgArm,
			&r.LgAvgVelocity, &r.LgAvgJunk, &r.LgAvgAccuracy,
			&r.RoleAvgPower, &r.RoleAvgContact, &r.RoleAvgSpeed, &r.RoleAvgFielding, &r.RoleAvgArm,
			&r.RoleAvgVelocity, &r.RoleAvgJunk, &r.RoleAvgAccuracy,
		); err != nil {
			return nil, fmt.Errorf("scanning attribute history row: %w", err)
		}
		if powerPct.Valid        { r.PowerPct        = &powerPct.Float64 }
		if contactPct.Valid      { r.ContactPct      = &contactPct.Float64 }
		if speedPct.Valid        { r.SpeedPct        = &speedPct.Float64 }
		if fieldingPct.Valid     { r.FieldingPct     = &fieldingPct.Float64 }
		if armPct.Valid          { r.ArmPct          = &armPct.Float64 }
		if velocityPct.Valid     { r.VelocityPct     = &velocityPct.Float64 }
		if junkPct.Valid         { r.JunkPct         = &junkPct.Float64 }
		if accuracyPct.Valid     { r.AccuracyPct     = &accuracyPct.Float64 }
		if powerPctRole.Valid    { r.PowerPctRole    = &powerPctRole.Float64 }
		if contactPctRole.Valid  { r.ContactPctRole  = &contactPctRole.Float64 }
		if speedPctRole.Valid    { r.SpeedPctRole    = &speedPctRole.Float64 }
		if fieldingPctRole.Valid { r.FieldingPctRole = &fieldingPctRole.Float64 }
		out = append(out, r)
	}
	return out, rows.Err()
}

// SetHallOfFamer updates the is_hall_of_famer flag for the given player.
func (s *PlayerQueryStore) SetHallOfFamer(ctx context.Context, playerID int64, isHoF bool) error {
	v := 0
	if isHoF {
		v = 1
	}
	_, err := s.db.ExecContext(ctx,
		`UPDATE players SET is_hall_of_famer = ? WHERE id = ?`, v, playerID)
	if err != nil {
		return fmt.Errorf("setting hall of famer for player %d: %w", playerID, err)
	}
	return nil
}
