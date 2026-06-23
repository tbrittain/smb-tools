package store

import (
	"context"
	"database/sql"
	"encoding/json"
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

// gameTypeClause translates the GameType filter into an is_regular_season value and a
// combined flag. "playoffs" maps to is_regular_season=0, "combined" suppresses the
// is_regular_season WHERE condition entirely, and anything else (including "") defaults
// to regular season (is_regular_season=1).
func gameTypeClause(gameType string) (isRegular int, isCombined bool) {
	switch gameType {
	case "playoffs":
		return 0, false
	case "combined":
		return 0, true
	default:
		return 1, false
	}
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

// GetBattingCareerLeaders returns a paginated page of career batting totals
// aggregated across all seasons matching the filters. Rate stats are computed
// inline via a CTE. Returns the page rows and total matching player count.
func (s *LeaderboardQueryStore) GetBattingCareerLeaders(
	ctx context.Context, f models.LeaderboardFilters,
) ([]models.BattingCareerLeaderRow, int, error) {
	isRegArg, isCombined := gameTypeClause(f.GameType)

	var args []any
	var conds []string

	if !isCombined {
		conds = append(conds, "b.is_regular_season = ?")
		args = append(args, isRegArg)
	}

	w := buildLeaderboardConditions(f, "ps.primary_position", "s")
	args = append(args, w.args...)
	conds = append(conds, w.conds...)

	whereClause := ""
	if len(conds) > 0 {
		whereClause = "WHERE " + strings.Join(conds, " AND ")
	}

	// Career OPS+ requires per-season league context; meaningless when combining regular
	// season and playoff rows from different environments. Skip the lss JOIN entirely in
	// combined mode — keeping it without the is_regular_season match condition would
	// cross-join each stat row against two lss rows (regular + playoff), doubling all SUMs.
	var lssJoin, opsPlusExpr string
	if isCombined {
		opsPlusExpr = "NULL"
	} else {
		lssJoin = `
LEFT JOIN league_season_stats lss ON lss.season_id = ps.season_id
    AND lss.is_regular_season = b.is_regular_season`
		opsPlusExpr = `CASE
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
    END`
	}

	// The CTE aggregates counting stats and computes rate fields inline so that
	// the outer SELECT can ORDER BY any column (including rates) without a subquery.
	// Career qualification: regular season (and "combined") use the RS-scaled PA
	// threshold; playoffs use the fixed, unscaled MLB postseason minimums (PA OR
	// BB+H), since a career's worth of playoff PA is tiny compared to a regular
	// season scaled threshold. See export_store.go's career_batting case for the
	// matching fix.
	qualHaving := ""
	if f.QualifiedOnly {
		thresholds, err := GetCareerQualificationThresholds(ctx, s.db)
		if err != nil {
			return nil, 0, err
		}
		if f.GameType == "playoffs" {
			qualHaving = `
  AND (CAST(SUM(b.plate_appearances) AS REAL) >= ?
       OR CAST(SUM(b.hits + b.walks) AS REAL) >= ?)`
			args = append(args, thresholds.BattingPAThresholdPO, thresholds.BattingBBHThresholdPO)
		} else {
			qualHaving = `
  AND CAST(SUM(b.plate_appearances) AS REAL) >= ?`
			args = append(args, thresholds.BattingPAThresholdRS)
		}
	}
	cte := fmt.Sprintf(`
WITH career AS (
    SELECT
        p.id            AS player_id,
        p.first_name,
        p.last_name,
        p.is_hall_of_famer,
        COUNT(DISTINCT ps.season_id)           AS seasons_played,
        COALESCE(SUM(b.games_played),    0)    AS games_played,
        COALESCE(SUM(b.games_batting),   0)    AS games_batting,
        COALESCE(SUM(b.at_bats),         0)    AS at_bats,
        COALESCE(SUM(b.runs),            0)    AS runs,
        COALESCE(SUM(b.hits),            0)    AS hits,
        COALESCE(SUM(b.doubles),         0)    AS doubles,
        COALESCE(SUM(b.triples),         0)    AS triples,
        COALESCE(SUM(b.home_runs),       0)    AS home_runs,
        COALESCE(SUM(b.rbi),             0)    AS rbi,
        COALESCE(SUM(b.stolen_bases),    0)    AS stolen_bases,
        COALESCE(SUM(b.caught_stealing), 0)    AS caught_stealing,
        COALESCE(SUM(b.walks),           0)    AS walks,
        COALESCE(SUM(b.strikeouts),      0)    AS strikeouts,
        COALESCE(SUM(b.hit_by_pitch),    0)    AS hit_by_pitch,
        COALESCE(SUM(b.sac_hits),        0)    AS sac_hits,
        COALESCE(SUM(b.sac_flies),       0)    AS sac_flies,
        COALESCE(SUM(b.errors),          0)    AS errors,
        COALESCE(SUM(b.passed_balls),    0)    AS passed_balls,
        SUM(b.smb_war)                         AS smb_war,
        %s                                     AS ops_plus,
        CAST(SUM(b.hits) AS REAL)
            / NULLIF(SUM(b.at_bats), 0)        AS ba,
        CAST(SUM(b.hits + b.walks + b.hit_by_pitch) AS REAL)
            / NULLIF(SUM(b.at_bats + b.walks + b.hit_by_pitch + b.sac_flies), 0) AS obp,
        CAST(SUM(b.hits - b.doubles - b.triples - b.home_runs
                  + b.doubles * 2 + b.triples * 3 + b.home_runs * 4) AS REAL)
            / NULLIF(SUM(b.at_bats), 0)        AS slg,
        CAST(SUM(b.hits + b.walks + b.hit_by_pitch) AS REAL)
            / NULLIF(SUM(b.at_bats + b.walks + b.hit_by_pitch + b.sac_flies), 0)
          + CAST(SUM(b.hits - b.doubles - b.triples - b.home_runs
                      + b.doubles * 2 + b.triples * 3 + b.home_runs * 4) AS REAL)
            / NULLIF(SUM(b.at_bats), 0)        AS ops,
        CAST(SUM(b.hits - b.doubles - b.triples - b.home_runs
                  + b.doubles * 2 + b.triples * 3 + b.home_runs * 4) AS REAL)
            / NULLIF(SUM(b.at_bats), 0)
          - CAST(SUM(b.hits) AS REAL)
            / NULLIF(SUM(b.at_bats), 0)        AS iso,
        CAST(SUM(b.hits - b.home_runs) AS REAL)
            / NULLIF(SUM(b.at_bats - b.strikeouts - b.home_runs + b.sac_flies), 0) AS babip,
        CAST(SUM(b.strikeouts) AS REAL)
            / NULLIF(SUM(b.at_bats + b.walks + b.hit_by_pitch + b.sac_hits + b.sac_flies), 0) AS k_pct,
        CAST(SUM(b.walks) AS REAL)
            / NULLIF(SUM(b.at_bats + b.walks + b.hit_by_pitch + b.sac_hits + b.sac_flies), 0) AS bb_pct,
        CAST(SUM(b.at_bats) AS REAL)
            / NULLIF(SUM(b.home_runs), 0)      AS ab_per_hr
    FROM player_season_batting_stats b
    JOIN player_seasons ps ON ps.id = b.player_season_id
    JOIN seasons s         ON s.id  = ps.season_id
    JOIN players p         ON p.id  = ps.player_id%s
    %s
    GROUP BY p.id
    HAVING COALESCE(SUM(b.at_bats), 0) > 0%s
)`, opsPlusExpr, lssJoin, whereClause, qualHaving)

	var total int
	if err := s.db.QueryRowContext(ctx, cte+"\nSELECT COUNT(*) FROM career", args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("GetBattingCareerLeaders count: %w", err)
	}

	pageSize := f.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	orderBy := buildOrderBy(f.SortField, f.SortDesc, battingCareerSortCols, battingCareerNullable,
		"COALESCE(smb_war, -9999.0) DESC")

	dataQuery := cte + fmt.Sprintf(`
SELECT
    player_id, first_name, last_name, is_hall_of_famer, seasons_played,
    games_played, games_batting, at_bats, runs, hits, doubles, triples,
    home_runs, rbi, stolen_bases, caught_stealing, walks, strikeouts,
    hit_by_pitch, sac_hits, sac_flies, errors, passed_balls,
    smb_war, ops_plus, ba, obp, slg, ops, iso, babip, k_pct, bb_pct, ab_per_hr
FROM career
ORDER BY %s, last_name, first_name
LIMIT ? OFFSET ?`, orderBy)

	dataArgs := append(args, pageSize, offset)
	rows, err := s.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("GetBattingCareerLeaders: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.BattingCareerLeaderRow
	for rows.Next() {
		var r models.BattingCareerLeaderRow
		var hof int
		var bSmbWAR, bOPSPlus sql.NullFloat64
		var bBA, bOBP, bSLG, bOPS, bISO, bBABIP, bKPct, bBBPct, bABPerHR sql.NullFloat64
		if err := rows.Scan(
			&r.PlayerID, &r.FirstName, &r.LastName, &hof,
			&r.SeasonsPlayed,
			&r.GamesPlayed, &r.GamesBatting,
			&r.AtBats, &r.Runs, &r.Hits, &r.Doubles, &r.Triples, &r.HomeRuns, &r.RBI,
			&r.StolenBases, &r.CaughtStealing, &r.Walks, &r.Strikeouts,
			&r.HitByPitch, &r.SacHits, &r.SacFlies, &r.Errors, &r.PassedBalls,
			&bSmbWAR, &bOPSPlus,
			&bBA, &bOBP, &bSLG, &bOPS, &bISO, &bBABIP, &bKPct, &bBBPct, &bABPerHR,
		); err != nil {
			return nil, 0, fmt.Errorf("GetBattingCareerLeaders scan: %w", err)
		}
		r.IsHallOfFamer = hof == 1
		r.SmbWAR  = nullFloat(bSmbWAR)
		r.OPSPlus = nullFloat(bOPSPlus)
		r.BA      = nullFloat(bBA)
		r.OBP     = nullFloat(bOBP)
		r.SLG     = nullFloat(bSLG)
		r.OPS     = nullFloat(bOPS)
		r.ISO     = nullFloat(bISO)
		r.BABIP   = nullFloat(bBABIP)
		r.KPct    = nullFloat(bKPct)
		r.BBPct   = nullFloat(bBBPct)
		r.ABPerHR = nullFloat(bABPerHR)
		out = append(out, r)
	}
	return out, total, rows.Err()
}

// battingSeasonSortCols maps frontend camelCase field names to safe SQL expressions
// for ORDER BY injection in GetBattingSeasonLeaders.
var battingSeasonSortCols = map[string]string{
	"lastName": "p.last_name", "seasonNum": "s.season_num",
	"teamName": "COALESCE(tsh.team_name, '')", "age": "COALESCE(ps.age, 0)",
	"primaryPosition": "ps.primary_position", "batHand": "ps.bat_hand",
	"gamesPlayed": "b.games_played", "gameBatting": "b.games_batting",
	"atBats": "b.at_bats", "runs": "b.runs", "hits": "b.hits",
	"doubles": "b.doubles", "triples": "b.triples", "homeRuns": "b.home_runs",
	"rbi": "b.rbi", "stolenBases": "b.stolen_bases", "caughtStealing": "b.caught_stealing",
	"walks": "b.walks", "strikeouts": "b.strikeouts",
	"ba": "b.ba", "obp": "b.obp", "slg": "b.slg", "ops": "b.ops",
	"iso": "b.iso", "babip": "b.babip", "kPct": "b.k_pct", "bbPct": "b.bb_pct",
	"abPerHr": "b.ab_per_hr", "opsPlus": "b.ops_plus", "smbWar": "b.smb_war",
}

// battingSeasonNullable is the subset of battingSeasonSortCols whose values can be NULL.
var battingSeasonNullable = map[string]bool{
	"ba": true, "obp": true, "slg": true, "ops": true, "iso": true,
	"babip": true, "kPct": true, "bbPct": true, "abPerHr": true,
	"opsPlus": true, "smbWar": true,
}

// pitchingSeasonSortCols maps frontend camelCase field names to SQL expressions
// for ORDER BY injection in GetPitchingSeasonLeaders.
var pitchingSeasonSortCols = map[string]string{
	"lastName": "p.last_name", "seasonNum": "s.season_num",
	"teamName": "COALESCE(tsh.team_name, '')", "age": "COALESCE(ps.age, 0)",
	"pitcherRole": "ps.pitcher_role", "throwHand": "ps.throw_hand",
	"games": "pit.games", "gamesStarted": "pit.games_started",
	"wins": "pit.wins", "losses": "pit.losses", "saves": "pit.saves",
	"outsPitched": "pit.outs_pitched", "hitsAllowed": "pit.hits_allowed",
	"earnedRuns": "pit.earned_runs", "walks": "pit.walks", "strikeouts": "pit.strikeouts",
	"era": "pit.era", "whip": "pit.whip", "k9": "pit.k_per_9", "bb9": "pit.bb_per_9",
	"kPerBb": "pit.k_per_bb", "kPct": "pit.k_pct", "winPct": "pit.win_pct",
	"eraPlus": "pit.era_plus", "fip": "pit.fip", "fipMinus": "pit.fip_minus",
	"smbWar": "pit.smb_war",
}

// pitchingSeasonNullable is the subset of pitchingSeasonSortCols whose values can be NULL.
var pitchingSeasonNullable = map[string]bool{
	"era": true, "whip": true, "k9": true, "bb9": true, "kPerBb": true,
	"kPct": true, "winPct": true, "eraPlus": true, "fip": true,
	"fipMinus": true, "smbWar": true,
}

// battingCareerSortCols maps frontend camelCase field names to CTE alias names
// for ORDER BY injection in GetBattingCareerLeaders.
var battingCareerSortCols = map[string]string{
	"lastName": "last_name", "seasonsPlayed": "seasons_played",
	"gamesPlayed": "games_played", "atBats": "at_bats", "runs": "runs",
	"hits": "hits", "doubles": "doubles", "triples": "triples",
	"homeRuns": "home_runs", "rbi": "rbi",
	"stolenBases": "stolen_bases", "caughtStealing": "caught_stealing",
	"walks": "walks", "strikeouts": "strikeouts",
	"ba": "ba", "obp": "obp", "slg": "slg", "ops": "ops", "iso": "iso",
	"babip": "babip", "kPct": "k_pct", "bbPct": "bb_pct", "abPerHr": "ab_per_hr",
	"opsPlus": "ops_plus", "smbWar": "smb_war",
}

// battingCareerNullable is the subset of battingCareerSortCols whose values can be NULL.
var battingCareerNullable = map[string]bool{
	"ba": true, "obp": true, "slg": true, "ops": true, "iso": true,
	"babip": true, "kPct": true, "bbPct": true, "abPerHr": true,
	"opsPlus": true, "smbWar": true,
}

// pitchingCareerSortCols maps frontend camelCase field names to CTE alias names
// for ORDER BY injection in GetPitchingCareerLeaders.
var pitchingCareerSortCols = map[string]string{
	"lastName": "last_name", "seasonsPlayed": "seasons_played",
	"wins": "wins", "losses": "losses", "games": "games",
	"gamesStarted": "games_started", "completeGames": "complete_games",
	"shutouts": "shutouts", "saves": "saves", "outsPitched": "outs_pitched",
	"hitsAllowed": "hits_allowed", "earnedRuns": "earned_runs",
	"homeRunsAllowed": "home_runs_allowed", "walks": "walks",
	"strikeouts": "strikeouts", "hitBatters": "hit_batters",
	"battersFaced": "batters_faced", "gamesFinished": "games_finished",
	"era": "era", "whip": "whip", "k9": "k9", "bb9": "bb9",
	"kPerBb": "k_per_bb", "kPct": "k_pct", "winPct": "win_pct",
	"eraPlus": "era_plus", "smbWar": "smb_war",
}

// pitchingCareerNullable is the subset of pitchingCareerSortCols whose values can be NULL.
var pitchingCareerNullable = map[string]bool{
	"era": true, "whip": true, "k9": true, "bb9": true,
	"kPerBb": true, "kPct": true, "winPct": true, "eraPlus": true, "smbWar": true,
}

// buildOrderBy returns a safe ORDER BY expression for the given sort field.
// Nullable columns use COALESCE to push NULLs to the bottom regardless of direction.
func buildOrderBy(fieldKey string, desc bool, colMap map[string]string, nullMap map[string]bool, fallback string) string {
	col, ok := colMap[fieldKey]
	if !ok {
		return fallback
	}
	dir := "ASC"
	if desc {
		dir = "DESC"
	}
	if nullMap[fieldKey] {
		sentinel := "9999.0"
		if desc {
			sentinel = "-9999.0"
		}
		return fmt.Sprintf("COALESCE(%s, %s) %s", col, sentinel, dir)
	}
	return col + " " + dir
}

// GetBattingSeasonLeaders returns one row per player-season with batting stats
// for each individual season matching the filters. Rate fields are read from
// stored columns — no on-read computation required. Returns the page rows and
// total matching row count for server-side pagination.
func (s *LeaderboardQueryStore) GetBattingSeasonLeaders(
	ctx context.Context, f models.LeaderboardFilters,
) ([]models.BattingSeasonLeaderRow, int, error) {
	isRegArg, isCombined := gameTypeClause(f.GameType)

	var filterArgs []any
	var whereParts []string

	if !isCombined {
		whereParts = append(whereParts, "b.is_regular_season = ?")
		filterArgs = append(filterArgs, isRegArg)
	}
	whereParts = append(whereParts, "b.at_bats > 0")

	w := buildLeaderboardConditions(f, "ps.primary_position", "s")
	filterArgs = append(filterArgs, w.args...)
	whereParts = append(whereParts, w.conds...)

	for _, trait := range f.Traits {
		whereParts = append(whereParts, "EXISTS (SELECT 1 FROM json_each(ps.traits_json) WHERE value = ?)")
		filterArgs = append(filterArgs, trait)
	}
	// baseJoins is extended before the WHERE clause so that the qual_gp subquery
	// (when QualifiedOnly is true) can be a plain JOIN rather than a correlated
	// subquery — the correlated form re-ran MAX(games_played) once per row.
	baseJoins := `
FROM player_season_batting_stats b
JOIN player_seasons ps ON ps.id = b.player_season_id
JOIN seasons s         ON s.id  = ps.season_id
JOIN players p         ON p.id  = ps.player_id
LEFT JOIN player_season_teams pst ON pst.player_season_id = ps.id AND pst.sort_order = 0
LEFT JOIN team_season_history tsh ON tsh.id = pst.team_history_id`

	if f.QualifiedOnly {
		// Pre-aggregate max games_played per (season, game-type) once.
		// Joining instead of a correlated subquery keeps this O(n) not O(n²).
		// See the note in docs/store for the future schedule-based improvement.
		baseJoins += `
JOIN (
    SELECT ps2.season_id, b2.is_regular_season, COALESCE(MAX(b2.games_played), 1) AS max_gp
    FROM player_season_batting_stats b2
    JOIN player_seasons ps2 ON ps2.id = b2.player_season_id
    GROUP BY ps2.season_id, b2.is_regular_season
) qual_gp ON qual_gp.season_id = ps.season_id AND qual_gp.is_regular_season = b.is_regular_season`
		whereParts = append(whereParts, "CAST(b.plate_appearances AS REAL) >= qual_gp.max_gp * 3.1")
	}

	joins := baseJoins + "\nWHERE " + strings.Join(whereParts, "\n  AND ")

	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*)`+joins, filterArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("GetBattingSeasonLeaders count: %w", err)
	}

	pageSize := f.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	orderBy := buildOrderBy(f.SortField, f.SortDesc, battingSeasonSortCols, battingSeasonNullable,
		"COALESCE(b.smb_war, -9999.0) DESC")

	q := `SELECT
    p.id, p.first_name, p.last_name, p.is_hall_of_famer,
    s.season_num,
    ps.id,
    ps.age,
    ps.primary_position,
    ps.bat_hand,
    ps.chemistry_type,
    ps.traits_json,
    b.games_played, b.games_batting,
    b.at_bats, b.runs, b.hits, b.doubles, b.triples, b.home_runs, b.rbi,
    b.stolen_bases, b.caught_stealing, b.walks, b.strikeouts,
    b.hit_by_pitch, b.sac_hits, b.sac_flies, b.errors, b.passed_balls,
    b.ba, b.obp, b.slg, b.ops, b.iso, b.babip, b.k_pct, b.bb_pct, b.ab_per_hr,
    b.ops_plus, b.smb_war` + joins + `
ORDER BY ` + orderBy + `, p.last_name, p.first_name, s.season_num
LIMIT ? OFFSET ?`

	dataArgs := append(filterArgs, pageSize, offset)
	rows, err := s.db.QueryContext(ctx, q, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("GetBattingSeasonLeaders: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.BattingSeasonLeaderRow
	for rows.Next() {
		var r models.BattingSeasonLeaderRow
		var hof int
		var traitsJSON string
		var bBA, bOBP, bSLG, bOPS, bISO, bBABIP, bKPct, bBBPct, bABPerHR sql.NullFloat64
		var bOPSPlus, bSmbWAR sql.NullFloat64
		if err := rows.Scan(
			&r.PlayerID, &r.FirstName, &r.LastName, &hof,
			&r.SeasonNum, &r.PlayerSeasonID, &r.Age,
			&r.PrimaryPosition, &r.BatHand, &r.ChemistryType,
			&traitsJSON,
			&r.GamesPlayed, &r.GamesBatting,
			&r.AtBats, &r.Runs, &r.Hits, &r.Doubles, &r.Triples, &r.HomeRuns, &r.RBI,
			&r.StolenBases, &r.CaughtStealing, &r.Walks, &r.Strikeouts,
			&r.HitByPitch, &r.SacHits, &r.SacFlies, &r.Errors, &r.PassedBalls,
			&bBA, &bOBP, &bSLG, &bOPS, &bISO, &bBABIP, &bKPct, &bBBPct, &bABPerHR,
			&bOPSPlus, &bSmbWAR,
		); err != nil {
			return nil, 0, fmt.Errorf("GetBattingSeasonLeaders scan: %w", err)
		}
		r.IsHallOfFamer = hof == 1
		if traitsJSON != "" {
			_ = json.Unmarshal([]byte(traitsJSON), &r.Traits)
		}
		if r.Traits == nil {
			r.Traits = []string{}
		}
		r.BA      = nullFloat(bBA)
		r.OBP     = nullFloat(bOBP)
		r.SLG     = nullFloat(bSLG)
		r.OPS     = nullFloat(bOPS)
		r.ISO     = nullFloat(bISO)
		r.BABIP   = nullFloat(bBABIP)
		r.KPct    = nullFloat(bKPct)
		r.BBPct   = nullFloat(bBBPct)
		r.ABPerHR = nullFloat(bABPerHR)
		r.OPSPlus = nullFloat(bOPSPlus)
		r.SmbWAR  = nullFloat(bSmbWAR)
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if err := s.enrichBattingSeasonWithTeams(ctx, out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// GetPitchingCareerLeaders returns a paginated page of career pitching totals
// aggregated across all seasons matching the filters. Rate stats are computed
// inline via a CTE. Returns the page rows and total matching player count.
func (s *LeaderboardQueryStore) GetPitchingCareerLeaders(
	ctx context.Context, f models.LeaderboardFilters,
) ([]models.PitchingCareerLeaderRow, int, error) {
	isRegArg, isCombined := gameTypeClause(f.GameType)

	var args []any
	var conds []string

	if !isCombined {
		conds = append(conds, "pit.is_regular_season = ?")
		args = append(args, isRegArg)
	}

	w := buildLeaderboardConditions(f, "ps.pitcher_role", "s")
	args = append(args, w.args...)
	conds = append(conds, w.conds...)

	whereClause := ""
	if len(conds) > 0 {
		whereClause = "WHERE " + strings.Join(conds, " AND ")
	}

	// Career ERA+ requires per-season league context; meaningless when combining regular
	// season and playoff rows from different environments. Skip the lss JOIN entirely in
	// combined mode — keeping it without the is_regular_season match condition would
	// cross-join each stat row against two lss rows (regular + playoff), doubling all SUMs.
	var lssJoin, eraPlusExpr string
	if isCombined {
		eraPlusExpr = "NULL"
	} else {
		lssJoin = `
LEFT JOIN league_season_stats lss ON lss.season_id = ps.season_id
    AND lss.is_regular_season = pit.is_regular_season`
		eraPlusExpr = `CASE
        WHEN COALESCE(SUM(pit.outs_pitched), 0) > 0
         AND COALESCE(SUM(pit.earned_runs),  0) > 0
         AND COALESCE(SUM(lss.total_outs_pitched), 0) > 0
        THEN CAST(SUM(lss.total_earned_runs) AS REAL) * 27.0 / SUM(lss.total_outs_pitched)
             / (CAST(SUM(pit.earned_runs) AS REAL) * 27.0 / SUM(pit.outs_pitched))
             * 100.0
        ELSE NULL
    END`
	}

	// Career qualification: see the matching comment in GetBattingCareerLeaders above.
	qualHaving := ""
	if f.QualifiedOnly {
		thresholds, err := GetCareerQualificationThresholds(ctx, s.db)
		if err != nil {
			return nil, 0, err
		}
		if f.GameType == "playoffs" {
			qualHaving = `
  AND (SUM(pit.outs_pitched) >= ?
       OR SUM(pit.wins + pit.losses) >= ?)`
			args = append(args, thresholds.PitchingOutsThresholdPO, thresholds.PitchingDecisionsThresholdPO)
		} else {
			qualHaving = `
  AND SUM(pit.outs_pitched) >= ?`
			args = append(args, thresholds.PitchingOutsThresholdRS)
		}
	}

	cte := fmt.Sprintf(`
WITH career AS (
    SELECT
        p.id            AS player_id,
        p.first_name,
        p.last_name,
        p.is_hall_of_famer,
        COUNT(DISTINCT ps.season_id)                AS seasons_played,
        COALESCE(SUM(pit.wins),             0)      AS wins,
        COALESCE(SUM(pit.losses),           0)      AS losses,
        COALESCE(SUM(pit.games),            0)      AS games,
        COALESCE(SUM(pit.games_started),    0)      AS games_started,
        COALESCE(SUM(pit.complete_games),   0)      AS complete_games,
        COALESCE(SUM(pit.shutouts),         0)      AS shutouts,
        COALESCE(SUM(pit.saves),            0)      AS saves,
        COALESCE(SUM(pit.outs_pitched),     0)      AS outs_pitched,
        COALESCE(SUM(pit.hits_allowed),     0)      AS hits_allowed,
        COALESCE(SUM(pit.earned_runs),      0)      AS earned_runs,
        COALESCE(SUM(pit.home_runs_allowed),0)      AS home_runs_allowed,
        COALESCE(SUM(pit.walks),            0)      AS walks,
        COALESCE(SUM(pit.strikeouts),       0)      AS strikeouts,
        COALESCE(SUM(pit.hit_batters),      0)      AS hit_batters,
        COALESCE(SUM(pit.batters_faced),    0)      AS batters_faced,
        COALESCE(SUM(pit.games_finished),   0)      AS games_finished,
        COALESCE(SUM(pit.runs_allowed),     0)      AS runs_allowed,
        COALESCE(SUM(pit.wild_pitches),     0)      AS wild_pitches,
        COALESCE(SUM(pit.total_pitches),    0)      AS total_pitches,
        SUM(pit.smb_war)                            AS smb_war,
        %s                                          AS era_plus,
        CAST(SUM(pit.earned_runs) AS REAL) * 27.0
            / NULLIF(SUM(pit.outs_pitched), 0)      AS era,
        CAST(SUM(pit.walks + pit.hits_allowed) AS REAL) * 3.0
            / NULLIF(SUM(pit.outs_pitched), 0)      AS whip,
        CAST(SUM(pit.strikeouts) AS REAL) * 27.0
            / NULLIF(SUM(pit.outs_pitched), 0)      AS k9,
        CAST(SUM(pit.walks) AS REAL) * 27.0
            / NULLIF(SUM(pit.outs_pitched), 0)      AS bb9,
        CAST(SUM(pit.strikeouts) AS REAL)
            / NULLIF(SUM(pit.walks), 0)             AS k_per_bb,
        CAST(SUM(pit.strikeouts) AS REAL)
            / NULLIF(SUM(pit.batters_faced), 0)     AS k_pct,
        CAST(SUM(pit.wins) AS REAL)
            / NULLIF(SUM(pit.wins + pit.losses), 0) AS win_pct
    FROM player_season_pitching_stats pit
    JOIN player_seasons ps ON ps.id = pit.player_season_id
    JOIN seasons s         ON s.id  = ps.season_id
    JOIN players p         ON p.id  = ps.player_id%s
    %s
    GROUP BY p.id
    HAVING COALESCE(SUM(pit.outs_pitched), 0) > 0%s
)`, eraPlusExpr, lssJoin, whereClause, qualHaving)

	var total int
	if err := s.db.QueryRowContext(ctx, cte+"\nSELECT COUNT(*) FROM career", args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("GetPitchingCareerLeaders count: %w", err)
	}

	pageSize := f.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	orderBy := buildOrderBy(f.SortField, f.SortDesc, pitchingCareerSortCols, pitchingCareerNullable,
		"COALESCE(smb_war, -9999.0) DESC")

	dataQuery := cte + fmt.Sprintf(`
SELECT
    player_id, first_name, last_name, is_hall_of_famer, seasons_played,
    wins, losses, games, games_started, complete_games, shutouts, saves,
    outs_pitched, hits_allowed, earned_runs, home_runs_allowed,
    walks, strikeouts, hit_batters, batters_faced, games_finished,
    runs_allowed, wild_pitches, total_pitches,
    smb_war, era_plus, era, whip, k9, bb9, k_per_bb, k_pct, win_pct
FROM career
ORDER BY %s, last_name, first_name
LIMIT ? OFFSET ?`, orderBy)

	dataArgs := append(args, pageSize, offset)
	rows, err := s.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("GetPitchingCareerLeaders: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.PitchingCareerLeaderRow
	for rows.Next() {
		var r models.PitchingCareerLeaderRow
		var hof int
		var bSmbWAR, bERAPlus sql.NullFloat64
		var bERA, bWHIP, bK9, bBB9, bKPerBB, bKPct, bWinPct sql.NullFloat64
		if err := rows.Scan(
			&r.PlayerID, &r.FirstName, &r.LastName, &hof,
			&r.SeasonsPlayed,
			&r.Wins, &r.Losses, &r.Games, &r.GamesStarted,
			&r.CompleteGames, &r.Shutouts, &r.Saves, &r.OutsPitched,
			&r.HitsAllowed, &r.EarnedRuns, &r.HomeRunsAllowed, &r.Walks,
			&r.Strikeouts, &r.HitBatters, &r.BattersFaced, &r.GamesFinished,
			&r.RunsAllowed, &r.WildPitches, &r.TotalPitches,
			&bSmbWAR, &bERAPlus,
			&bERA, &bWHIP, &bK9, &bBB9, &bKPerBB, &bKPct, &bWinPct,
		); err != nil {
			return nil, 0, fmt.Errorf("GetPitchingCareerLeaders scan: %w", err)
		}
		r.IsHallOfFamer = hof == 1
		r.SmbWAR  = nullFloat(bSmbWAR)
		r.ERAPlus = nullFloat(bERAPlus)
		r.ERA     = nullFloat(bERA)
		r.WHIP    = nullFloat(bWHIP)
		r.K9      = nullFloat(bK9)
		r.BB9     = nullFloat(bBB9)
		r.KPerBB  = nullFloat(bKPerBB)
		r.KPct    = nullFloat(bKPct)
		r.WinPct  = nullFloat(bWinPct)
		out = append(out, r)
	}
	return out, total, rows.Err()
}

// GetPitchingSeasonLeaders returns one row per player-season with pitching stats
// for each individual season matching the filters. Rate fields are read from
// stored columns — no on-read computation required. Returns the page rows and
// total matching row count for server-side pagination.
func (s *LeaderboardQueryStore) GetPitchingSeasonLeaders(
	ctx context.Context, f models.LeaderboardFilters,
) ([]models.PitchingSeasonLeaderRow, int, error) {
	isRegArg, isCombined := gameTypeClause(f.GameType)

	var filterArgs []any
	var whereParts []string

	if !isCombined {
		whereParts = append(whereParts, "pit.is_regular_season = ?")
		filterArgs = append(filterArgs, isRegArg)
	}
	whereParts = append(whereParts, "pit.outs_pitched > 0")

	w := buildLeaderboardConditions(f, "ps.pitcher_role", "s")
	filterArgs = append(filterArgs, w.args...)
	whereParts = append(whereParts, w.conds...)

	for _, trait := range f.Traits {
		whereParts = append(whereParts, "EXISTS (SELECT 1 FROM json_each(ps.traits_json) WHERE value = ?)")
		filterArgs = append(filterArgs, trait)
	}
	baseJoins := `
FROM player_season_pitching_stats pit
JOIN player_seasons ps ON ps.id = pit.player_season_id
JOIN seasons s         ON s.id  = ps.season_id
JOIN players p         ON p.id  = ps.player_id
LEFT JOIN player_season_teams pst ON pst.player_season_id = ps.id AND pst.sort_order = 0
LEFT JOIN team_season_history tsh ON tsh.id = pst.team_history_id`

	if f.QualifiedOnly {
		// Pre-aggregate max batting games_played per (season, game-type) once.
		// Batting games_played reflects team games better than pitcher appearances.
		// Joining avoids the O(n²) correlated subquery.
		baseJoins += `
JOIN (
    SELECT ps2.season_id, b2.is_regular_season, COALESCE(MAX(b2.games_played), 1) AS max_gp
    FROM player_season_batting_stats b2
    JOIN player_seasons ps2 ON ps2.id = b2.player_season_id
    GROUP BY ps2.season_id, b2.is_regular_season
) qual_gp ON qual_gp.season_id = ps.season_id AND qual_gp.is_regular_season = pit.is_regular_season`
		whereParts = append(whereParts, "pit.outs_pitched >= qual_gp.max_gp * 3")
	}

	joins := baseJoins + "\nWHERE " + strings.Join(whereParts, "\n  AND ")

	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*)`+joins, filterArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("GetPitchingSeasonLeaders count: %w", err)
	}

	pageSize := f.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	orderBy := buildOrderBy(f.SortField, f.SortDesc, pitchingSeasonSortCols, pitchingSeasonNullable,
		"COALESCE(pit.smb_war, -9999.0) DESC")

	q := `SELECT
    p.id, p.first_name, p.last_name, p.is_hall_of_famer,
    s.season_num,
    ps.id,
    ps.age,
    ps.pitcher_role,
    ps.throw_hand,
    ps.chemistry_type,
    ps.traits_json,
    pit.wins, pit.losses, pit.games, pit.games_started,
    pit.complete_games, pit.shutouts, pit.saves, pit.outs_pitched,
    pit.hits_allowed, pit.earned_runs, pit.home_runs_allowed, pit.walks,
    pit.strikeouts, pit.hit_batters, pit.batters_faced, pit.games_finished,
    pit.runs_allowed, pit.wild_pitches, pit.total_pitches,
    pit.era, pit.whip, pit.k_per_9, pit.bb_per_9, pit.h_per_9, pit.hr_per_9,
    pit.k_per_bb, pit.k_pct, pit.win_pct, pit.p_per_ip,
    pit.era_plus, pit.fip, pit.fip_minus, pit.smb_war` + joins + `
ORDER BY ` + orderBy + `, p.last_name, p.first_name, s.season_num
LIMIT ? OFFSET ?`

	dataArgs := append(filterArgs, pageSize, offset)
	rows, err := s.db.QueryContext(ctx, q, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("GetPitchingSeasonLeaders: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.PitchingSeasonLeaderRow
	for rows.Next() {
		var r models.PitchingSeasonLeaderRow
		var hof int
		var traitsJSON string
		var pERA, pWHIP, pK9, pBB9, pH9, pHR9, pKPerBB, pKPct, pWinPct, pPPerIP sql.NullFloat64
		var pERAPlus, pFIP, pFIPMinus, pSmbWAR sql.NullFloat64
		if err := rows.Scan(
			&r.PlayerID, &r.FirstName, &r.LastName, &hof,
			&r.SeasonNum, &r.PlayerSeasonID, &r.Age,
			&r.PitcherRole, &r.ThrowHand, &r.ChemistryType,
			&traitsJSON,
			&r.Wins, &r.Losses, &r.Games, &r.GamesStarted,
			&r.CompleteGames, &r.Shutouts, &r.Saves, &r.OutsPitched,
			&r.HitsAllowed, &r.EarnedRuns, &r.HomeRunsAllowed, &r.Walks,
			&r.Strikeouts, &r.HitBatters, &r.BattersFaced, &r.GamesFinished,
			&r.RunsAllowed, &r.WildPitches, &r.TotalPitches,
			&pERA, &pWHIP, &pK9, &pBB9, &pH9, &pHR9, &pKPerBB, &pKPct, &pWinPct, &pPPerIP,
			&pERAPlus, &pFIP, &pFIPMinus, &pSmbWAR,
		); err != nil {
			return nil, 0, fmt.Errorf("GetPitchingSeasonLeaders scan: %w", err)
		}
		r.IsHallOfFamer = hof == 1
		if traitsJSON != "" {
			_ = json.Unmarshal([]byte(traitsJSON), &r.Traits)
		}
		if r.Traits == nil {
			r.Traits = []string{}
		}
		r.ERA     = nullFloat(pERA)
		r.WHIP    = nullFloat(pWHIP)
		r.K9      = nullFloat(pK9)
		r.BB9     = nullFloat(pBB9)
		r.H9      = nullFloat(pH9)
		r.HR9     = nullFloat(pHR9)
		r.KPerBB  = nullFloat(pKPerBB)
		r.KPct    = nullFloat(pKPct)
		r.WinPct  = nullFloat(pWinPct)
		r.PPerIP  = nullFloat(pPPerIP)
		r.ERAPlus  = nullFloat(pERAPlus)
		r.FIP      = nullFloat(pFIP)
		r.FIPMinus = nullFloat(pFIPMinus)
		r.SmbWAR   = nullFloat(pSmbWAR)
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if err := s.enrichPitchingSeasonWithTeams(ctx, out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// nullFloat converts a sql.NullFloat64 to a *float64 pointer, returning nil when invalid.
func nullFloat(n sql.NullFloat64) *float64 {
	if n.Valid {
		return &n.Float64
	}
	return nil
}

// enrichBattingSeasonWithTeams loads team associations for the given rows and sets each
// row's Teams field. Rows for player-seasons with no team entry receive an empty slice.
func (s *LeaderboardQueryStore) enrichBattingSeasonWithTeams(ctx context.Context, out []models.BattingSeasonLeaderRow) error {
	if len(out) == 0 {
		return nil
	}
	psIDs := make([]int64, len(out))
	for i, r := range out {
		psIDs[i] = r.PlayerSeasonID
	}
	teamsMap, err := s.loadSeasonTeams(ctx, psIDs)
	if err != nil {
		return err
	}
	for i := range out {
		if t, ok := teamsMap[out[i].PlayerSeasonID]; ok {
			out[i].Teams = t
		} else {
			out[i].Teams = []models.PlayerTeamRef{}
		}
	}
	return nil
}

// enrichPitchingSeasonWithTeams is the pitching equivalent of enrichBattingSeasonWithTeams.
func (s *LeaderboardQueryStore) enrichPitchingSeasonWithTeams(ctx context.Context, out []models.PitchingSeasonLeaderRow) error {
	if len(out) == 0 {
		return nil
	}
	psIDs := make([]int64, len(out))
	for i, r := range out {
		psIDs[i] = r.PlayerSeasonID
	}
	teamsMap, err := s.loadSeasonTeams(ctx, psIDs)
	if err != nil {
		return err
	}
	for i := range out {
		if t, ok := teamsMap[out[i].PlayerSeasonID]; ok {
			out[i].Teams = t
		} else {
			out[i].Teams = []models.PlayerTeamRef{}
		}
	}
	return nil
}

// loadSeasonTeams fetches all team associations for the given player_season IDs,
// returning a map from player_season_id to ordered team slice.
func (s *LeaderboardQueryStore) loadSeasonTeams(ctx context.Context, psIDs []int64) (map[int64][]models.PlayerTeamRef, error) {
	placeholders := strings.Repeat("?,", len(psIDs))
	placeholders = placeholders[:len(placeholders)-1]
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
