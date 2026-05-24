package store

import (
	"context"
	"database/sql"
	"fmt"

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

// GetPlayerCareer returns the player's bio and career totals (regular season).
// Rate fields on Batting and Pitching are computed before returning.
// Returns sql.ErrNoRows wrapped in an error if the player does not exist.
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

	// Career batting (SUM over all regular-season player_seasons), with career OPS+
	// computed from career totals vs career-weighted league averages.
	b := &models.CareerBattingStats{}
	err = s.db.QueryRowContext(ctx, `
SELECT
    COALESCE(SUM(bs.games_played),0), COALESCE(SUM(bs.games_batting),0),
    COALESCE(SUM(bs.at_bats),0),      COALESCE(SUM(bs.runs),0),
    COALESCE(SUM(bs.hits),0),         COALESCE(SUM(bs.doubles),0),
    COALESCE(SUM(bs.triples),0),      COALESCE(SUM(bs.home_runs),0),
    COALESCE(SUM(bs.rbi),0),          COALESCE(SUM(bs.stolen_bases),0),
    COALESCE(SUM(bs.caught_stealing),0), COALESCE(SUM(bs.walks),0),
    COALESCE(SUM(bs.strikeouts),0),   COALESCE(SUM(bs.hit_by_pitch),0),
    COALESCE(SUM(bs.sac_hits),0),     COALESCE(SUM(bs.sac_flies),0),
    COALESCE(SUM(bs.errors),0),       COALESCE(SUM(bs.passed_balls),0),
    SUM(bs.smb_war),
    CASE
        WHEN COALESCE(SUM(bs.at_bats), 0) > 0
         AND COALESCE(SUM(bs.at_bats + bs.walks + bs.hit_by_pitch + bs.sac_flies), 0) > 0
         AND COALESCE(SUM(lss.total_at_bats), 0) > 0
         AND COALESCE(SUM(lss.total_at_bats + lss.total_walks + lss.total_hbp + lss.total_sac_flies), 0) > 0
        THEN 100.0 * (
            CAST(SUM(bs.hits + bs.walks + bs.hit_by_pitch) AS REAL)
                / SUM(bs.at_bats + bs.walks + bs.hit_by_pitch + bs.sac_flies)
                / (CAST(SUM(lss.total_hits + lss.total_walks + lss.total_hbp) AS REAL)
                   / SUM(lss.total_at_bats + lss.total_walks + lss.total_hbp + lss.total_sac_flies))
            + CAST(SUM(bs.hits - bs.doubles - bs.triples - bs.home_runs
                        + bs.doubles * 2 + bs.triples * 3 + bs.home_runs * 4) AS REAL)
                / SUM(bs.at_bats)
                / (CAST(SUM(lss.total_hits - lss.total_doubles - lss.total_triples - lss.total_home_runs
                             + lss.total_doubles * 2 + lss.total_triples * 3 + lss.total_home_runs * 4) AS REAL)
                   / SUM(lss.total_at_bats))
            - 1.0
        )
        ELSE NULL
    END AS career_ops_plus
FROM player_season_batting_stats bs
JOIN player_seasons ps ON ps.id = bs.player_season_id
LEFT JOIN league_season_stats lss ON lss.season_id = ps.season_id
    AND lss.is_regular_season = bs.is_regular_season
WHERE ps.player_id = ? AND bs.is_regular_season = 1
`, playerID).Scan(
		&b.GamesPlayed, &b.GamesBatting,
		&b.AtBats, &b.Runs, &b.Hits, &b.Doubles, &b.Triples, &b.HomeRuns,
		&b.RBI, &b.StolenBases, &b.CaughtStealing, &b.Walks,
		&b.Strikeouts, &b.HitByPitch, &b.SacHits, &b.SacFlies,
		&b.Errors, &b.PassedBalls,
		&b.SmbWAR, &b.OPSPlus,
	)
	if err != nil {
		return c, fmt.Errorf("getting career batting for player %d: %w", playerID, err)
	}
	if b.AtBats > 0 || b.GamesPlayed > 0 {
		c.Batting = b
	}

	// Career pitching with career ERA+ from career-weighted league averages.
	p := &models.CareerPitchingStats{}
	err = s.db.QueryRowContext(ctx, `
SELECT
    COALESCE(SUM(pit.wins),0),             COALESCE(SUM(pit.losses),0),
    COALESCE(SUM(pit.games),0),            COALESCE(SUM(pit.games_started),0),
    COALESCE(SUM(pit.complete_games),0),   COALESCE(SUM(pit.shutouts),0),
    COALESCE(SUM(pit.saves),0),            COALESCE(SUM(pit.outs_pitched),0),
    COALESCE(SUM(pit.hits_allowed),0),     COALESCE(SUM(pit.earned_runs),0),
    COALESCE(SUM(pit.home_runs_allowed),0),COALESCE(SUM(pit.walks),0),
    COALESCE(SUM(pit.strikeouts),0),       COALESCE(SUM(pit.hit_batters),0),
    COALESCE(SUM(pit.batters_faced),0),    COALESCE(SUM(pit.games_finished),0),
    COALESCE(SUM(pit.runs_allowed),0),     COALESCE(SUM(pit.wild_pitches),0),
    COALESCE(SUM(pit.total_pitches),0),
    SUM(pit.smb_war),
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
LEFT JOIN league_season_stats lss ON lss.season_id = ps.season_id
    AND lss.is_regular_season = pit.is_regular_season
WHERE ps.player_id = ? AND pit.is_regular_season = 1
`, playerID).Scan(
		&p.Wins, &p.Losses, &p.Games, &p.GamesStarted,
		&p.CompleteGames, &p.Shutouts, &p.Saves, &p.OutsPitched,
		&p.HitsAllowed, &p.EarnedRuns, &p.HomeRunsAllowed, &p.Walks,
		&p.Strikeouts, &p.HitBatters, &p.BattersFaced, &p.GamesFinished,
		&p.RunsAllowed, &p.WildPitches, &p.TotalPitches,
		&p.SmbWAR, &p.ERAPlus,
	)
	if err != nil {
		return c, fmt.Errorf("getting career pitching for player %d: %w", playerID, err)
	}
	if p.OutsPitched > 0 || p.Games > 0 {
		c.Pitching = p
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

// scanSeasonLogRows scans either regular (isRegularSeason=1) or playoff
// (isRegularSeason=0) rows for a player. Returns the rows and a map from
// season_id to slice index (used by the caller to merge playoff stats).
func (s *PlayerQueryStore) scanSeasonLogRows(
	ctx context.Context, playerID int64, isRegularSeason int,
) ([]models.PlayerSeasonLogRow, map[int64]int, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT
    s.id                              AS season_id,
    s.season_num,
    ps.team_history_id,
    tsh.team_id,
    COALESCE(tsh.team_name, '')       AS team_name,
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
    b.ops_plus, b.smb_war,
    -- pitching block (all NULL when no pitching row for this season type)
    pit.outs_pitched,
    pit.wins, pit.losses, pit.games, pit.games_started,
    pit.complete_games, pit.shutouts, pit.saves,
    pit.hits_allowed, pit.earned_runs, pit.home_runs_allowed,
    pit.walks, pit.strikeouts, pit.hit_batters, pit.batters_faced,
    pit.games_finished, pit.runs_allowed, pit.wild_pitches, pit.total_pitches,
    pit.era_plus, pit.fip, pit.fip_minus, pit.smb_war
FROM player_seasons ps
JOIN seasons s ON s.id = ps.season_id
LEFT JOIN team_season_history tsh ON tsh.id = ps.team_history_id
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
		var teamHistID, teamID sql.NullInt64

		var bAtBats sql.NullInt64
		var bGP, bGB, bRuns, bHits, bDB, bTR, bHR, bRBI sql.NullInt64
		var bSB, bCS, bWalks, bK, bHBP, bSH, bSF, bE, bPB sql.NullInt64
		var bOPSPlus, bSmbWAR sql.NullFloat64

		// Pitching sentinel: outs_pitched is NULL when there is no pitching row.
		var pOuts sql.NullInt64
		var pW, pL, pG, pGS, pCG, pSHO, pSV sql.NullInt64
		var pH, pER, pHRA, pWalks, pK, pHBP, pBF, pGF, pRA, pWP, pTP sql.NullInt64
		var pERAPlus, pFIP, pFIPMinus, pSmbWAR sql.NullFloat64

		if err := rows.Scan(
			&row.SeasonID, &row.SeasonNum, &teamHistID, &teamID, &row.TeamName,
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
			&bOPSPlus, &bSmbWAR,
			// pitching block
			&pOuts,
			&pW, &pL, &pG, &pGS, &pCG, &pSHO, &pSV,
			&pH, &pER, &pHRA, &pWalks, &pK, &pHBP, &pBF, &pGF, &pRA, &pWP, &pTP,
			&pERAPlus, &pFIP, &pFIPMinus, &pSmbWAR,
		); err != nil {
			return nil, nil, fmt.Errorf("scanning season log row: %w", err)
		}

		if teamHistID.Valid {
			row.TeamHistoryID = &teamHistID.Int64
			row.TeamID = &teamID.Int64
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
			if bOPSPlus.Valid {
				b.OPSPlus = &bOPSPlus.Float64
			}
			if bSmbWAR.Valid {
				b.SmbWAR = &bSmbWAR.Float64
			}
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
			if pERAPlus.Valid {
				p.ERAPlus = &pERAPlus.Float64
			}
			if pFIP.Valid {
				p.FIP = &pFIP.Float64
			}
			if pFIPMinus.Valid {
				p.FIPMinus = &pFIPMinus.Float64
			}
			if pSmbWAR.Valid {
				p.SmbWAR = &pSmbWAR.Float64
			}
			row.Pitching = p
		}

		index[row.SeasonID] = len(out)
		out = append(out, row)
	}
	return out, index, rows.Err()
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
