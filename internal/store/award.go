package store

import (
	"context"
	"database/sql"
	"fmt"
	"slices"

	"smb-tools/internal/models"
)

// AwardStore provides read/write operations over the awards tables.
type AwardStore struct {
	db DBTX
}

// NewAwardStore creates an AwardStore backed by the given companion DB.
func NewAwardStore(db DBTX) *AwardStore {
	return &AwardStore{db: db}
}

// ── Award definitions ─────────────────────────────────────────────────────────

// ListAwards returns all award definitions filtered by the playoff flag.
// Pass isPlayoff=true for playoff awards, false for regular-season awards.
func (s *AwardStore) ListAwards(ctx context.Context, isPlayoff bool) ([]models.Award, error) {
	playoff := 0
	if isPlayoff {
		playoff = 1
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, name, original_name, importance, omit_from_groupings,
       is_batting_award, is_pitching_award, is_fielding_award,
       is_playoff_award, is_user_assignable, is_built_in
FROM awards
WHERE is_playoff_award = ?
ORDER BY importance ASC, name ASC
`, playoff)
	if err != nil {
		return nil, fmt.Errorf("listing awards (playoff=%v): %w", isPlayoff, err)
	}
	defer func() { _ = rows.Close() }()
	return scanAwards(rows)
}

// ListAllAwards returns every award definition regardless of playoff flag.
func (s *AwardStore) ListAllAwards(ctx context.Context) ([]models.Award, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, name, original_name, importance, omit_from_groupings,
       is_batting_award, is_pitching_award, is_fielding_award,
       is_playoff_award, is_user_assignable, is_built_in
FROM awards
ORDER BY importance ASC, name ASC
`)
	if err != nil {
		return nil, fmt.Errorf("listing all awards: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanAwards(rows)
}

// CreateCustomAward inserts a user-defined award and returns its new ID.
// Returns an error if the name is already taken.
func (s *AwardStore) CreateCustomAward(ctx context.Context, a models.Award) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
INSERT INTO awards
    (name, original_name, importance, omit_from_groupings,
     is_batting_award, is_pitching_award, is_fielding_award,
     is_playoff_award, is_user_assignable, is_built_in)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
`,
		a.Name, a.OriginalName, a.Importance, boolInt(a.OmitFromGroupings),
		boolInt(a.IsBattingAward), boolInt(a.IsPitchingAward), boolInt(a.IsFieldingAward),
		boolInt(a.IsPlayoffAward), boolInt(a.IsUserAssignable),
	)
	if err != nil {
		return 0, fmt.Errorf("creating custom award %q: %w", a.Name, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting new award ID: %w", err)
	}
	return id, nil
}

// ── Player-season award assignment ────────────────────────────────────────────

// GetSeasonPlayerAwards returns all player-seasons for the given season with
// their currently assigned awards. Uses two queries and merges in Go.
func (s *AwardStore) GetSeasonPlayerAwards(ctx context.Context, seasonID int64) ([]models.PlayerSeasonAwardRow, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT
    ps.id          AS player_season_id,
    p.id           AS player_id,
    p.first_name,
    p.last_name,
    COALESCE(tsh.team_name, '') AS team_name,
    ps.primary_position,
    ps.pitcher_role
FROM player_seasons ps
JOIN players p ON p.id = ps.player_id
LEFT JOIN player_season_teams pst ON pst.player_season_id = ps.id AND pst.sort_order = 0
LEFT JOIN team_season_history tsh ON tsh.id = pst.team_history_id
WHERE ps.season_id = ?
ORDER BY p.last_name ASC, p.first_name ASC
`, seasonID)
	if err != nil {
		return nil, fmt.Errorf("listing player-seasons for season %d: %w", seasonID, err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.PlayerSeasonAwardRow
	psIDToIndex := map[int64]int{}
	for rows.Next() {
		var r models.PlayerSeasonAwardRow
		if err := rows.Scan(
			&r.PlayerSeasonID, &r.PlayerID, &r.FirstName, &r.LastName,
			&r.TeamName, &r.PrimaryPos, &r.PitcherRole,
		); err != nil {
			return nil, fmt.Errorf("scanning player-season row: %w", err)
		}
		psIDToIndex[r.PlayerSeasonID] = len(out)
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return out, nil
	}

	awardRows, err := s.db.QueryContext(ctx, `
SELECT
    psa.player_season_id,
    a.id, a.name, a.original_name, a.importance, a.omit_from_groupings,
    a.is_batting_award, a.is_pitching_award, a.is_fielding_award,
    a.is_playoff_award, a.is_user_assignable, a.is_built_in
FROM player_season_awards psa
JOIN awards a ON a.id = psa.award_id
JOIN player_seasons ps ON ps.id = psa.player_season_id
WHERE ps.season_id = ?
ORDER BY psa.player_season_id, a.importance ASC
`, seasonID)
	if err != nil {
		return nil, fmt.Errorf("loading season awards: %w", err)
	}
	defer func() { _ = awardRows.Close() }()

	for awardRows.Next() {
		var psID int64
		var a models.Award
		if err := awardRows.Scan(
			&psID,
			&a.ID, &a.Name, &a.OriginalName, &a.Importance, &a.OmitFromGroupings,
			&a.IsBattingAward, &a.IsPitchingAward, &a.IsFieldingAward,
			&a.IsPlayoffAward, &a.IsUserAssignable, &a.IsBuiltIn,
		); err != nil {
			return nil, fmt.Errorf("scanning award row: %w", err)
		}
		if idx, ok := psIDToIndex[psID]; ok {
			out[idx].Awards = append(out[idx].Awards, a)
		}
	}
	return out, awardRows.Err()
}

// GetPlayerCareerAwards returns all awards for a player grouped by season_num.
func (s *AwardStore) GetPlayerCareerAwards(ctx context.Context, playerID int64) (map[int][]models.Award, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT
    s.season_num,
    a.id, a.name, a.original_name, a.importance, a.omit_from_groupings,
    a.is_batting_award, a.is_pitching_award, a.is_fielding_award,
    a.is_playoff_award, a.is_user_assignable, a.is_built_in
FROM player_season_awards psa
JOIN player_seasons ps ON ps.id = psa.player_season_id
JOIN seasons s         ON s.id  = ps.season_id
JOIN awards a          ON a.id  = psa.award_id
WHERE ps.player_id = ?
ORDER BY s.season_num ASC, a.importance ASC
`, playerID)
	if err != nil {
		return nil, fmt.Errorf("getting career awards for player %d: %w", playerID, err)
	}
	defer func() { _ = rows.Close() }()

	out := map[int][]models.Award{}
	for rows.Next() {
		var seasonNum int
		var a models.Award
		if err := rows.Scan(
			&seasonNum,
			&a.ID, &a.Name, &a.OriginalName, &a.Importance, &a.OmitFromGroupings,
			&a.IsBattingAward, &a.IsPitchingAward, &a.IsFieldingAward,
			&a.IsPlayoffAward, &a.IsUserAssignable, &a.IsBuiltIn,
		); err != nil {
			return nil, fmt.Errorf("scanning career award row: %w", err)
		}
		out[seasonNum] = append(out[seasonNum], a)
	}
	return out, rows.Err()
}

// SetPlayerSeasonAwards replaces the user-assignable awards for one
// player-season. Auto-computed awards (is_user_assignable=0) are untouched.
func (s *AwardStore) SetPlayerSeasonAwards(ctx context.Context, playerSeasonID int64, awardIDs []int64) error {
	db, ok := s.db.(*sql.DB)
	if !ok {
		return fmt.Errorf("SetPlayerSeasonAwards requires *sql.DB")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `
DELETE FROM player_season_awards
WHERE player_season_id = ?
  AND award_id IN (SELECT id FROM awards WHERE is_user_assignable = 1)
`, playerSeasonID)
	if err != nil {
		return fmt.Errorf("clearing user awards for player-season %d: %w", playerSeasonID, err)
	}

	for _, aid := range awardIDs {
		if _, err = tx.ExecContext(ctx, `
INSERT OR IGNORE INTO player_season_awards (player_season_id, award_id)
VALUES (?, ?)
`, playerSeasonID, aid); err != nil {
			return fmt.Errorf("assigning award %d to player-season %d: %w", aid, playerSeasonID, err)
		}
	}

	return tx.Commit()
}

// ── Auto-computed stat leader awards ─────────────────────────────────────────

// ComputeAndAssignStatLeaderAwards clears all auto-computed awards for the
// given season and recomputes them from regular-season batting and pitching stats.
// Idempotent — safe to call multiple times.
//
// Triple Crown logic: if the BA leader also leads HR and RBI, they receive only
// the Triple Crown award (not the three individual titles). Other leaders in
// each category still receive their individual title. Same rule for ERA/W/K.
func (s *AwardStore) ComputeAndAssignStatLeaderAwards(ctx context.Context, seasonID int64) error {
	db, ok := s.db.(*sql.DB)
	if !ok {
		return fmt.Errorf("ComputeAndAssignStatLeaderAwards requires *sql.DB")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Clear existing auto-computed awards for this season.
	if _, err = tx.ExecContext(ctx, `
DELETE FROM player_season_awards
WHERE award_id IN (SELECT id FROM awards WHERE is_user_assignable = 0)
  AND player_season_id IN (SELECT id FROM player_seasons WHERE season_id = ?)
`, seasonID); err != nil {
		return fmt.Errorf("clearing auto awards for season %d: %w", seasonID, err)
	}

	awardIDs, err := loadAutoAwardIDs(ctx, tx)
	if err != nil {
		return err
	}

	var numGames int
	if err := tx.QueryRowContext(ctx,
		`SELECT num_games FROM seasons WHERE id = ?`, seasonID,
	).Scan(&numGames); err != nil {
		return fmt.Errorf("loading num_games for season %d: %w", seasonID, err)
	}

	// Batting leaders.
	baLeader, err := queryQualifiedLeader(ctx, tx, seasonID, numGames, false)
	if err != nil {
		return err
	}
	hrLeaders, err := queryMaxLeaders(ctx, tx, seasonID, "home_runs", false)
	if err != nil {
		return err
	}
	rbiLeaders, err := queryMaxLeaders(ctx, tx, seasonID, "rbi", false)
	if err != nil {
		return err
	}

	tcBatting := int64(0)
	if baLeader != 0 && slices.Contains(hrLeaders, baLeader) && slices.Contains(rbiLeaders, baLeader) {
		tcBatting = baLeader
	}

	if tcBatting != 0 {
		if err := insertAward(ctx, tx, tcBatting, awardIDs["Triple Crown (Batting)"]); err != nil {
			return err
		}
		for _, id := range hrLeaders {
			if id != tcBatting {
				if err := insertAward(ctx, tx, id, awardIDs["Home Run Title"]); err != nil {
					return err
				}
			}
		}
		for _, id := range rbiLeaders {
			if id != tcBatting {
				if err := insertAward(ctx, tx, id, awardIDs["RBI Title"]); err != nil {
					return err
				}
			}
		}
	} else {
		if baLeader != 0 {
			if err := insertAward(ctx, tx, baLeader, awardIDs["Batting Title"]); err != nil {
				return err
			}
		}
		for _, id := range hrLeaders {
			if err := insertAward(ctx, tx, id, awardIDs["Home Run Title"]); err != nil {
				return err
			}
		}
		for _, id := range rbiLeaders {
			if err := insertAward(ctx, tx, id, awardIDs["RBI Title"]); err != nil {
				return err
			}
		}
	}

	// Pitching leaders.
	eraLeader, err := queryQualifiedLeader(ctx, tx, seasonID, numGames, true)
	if err != nil {
		return err
	}
	wLeaders, err := queryMaxLeaders(ctx, tx, seasonID, "wins", true)
	if err != nil {
		return err
	}
	kLeaders, err := queryMaxLeaders(ctx, tx, seasonID, "strikeouts", true)
	if err != nil {
		return err
	}

	tcPitching := int64(0)
	if eraLeader != 0 && slices.Contains(wLeaders, eraLeader) && slices.Contains(kLeaders, eraLeader) {
		tcPitching = eraLeader
	}

	if tcPitching != 0 {
		if err := insertAward(ctx, tx, tcPitching, awardIDs["Triple Crown (Pitching)"]); err != nil {
			return err
		}
		for _, id := range wLeaders {
			if id != tcPitching {
				if err := insertAward(ctx, tx, id, awardIDs["Wins Title"]); err != nil {
					return err
				}
			}
		}
		for _, id := range kLeaders {
			if id != tcPitching {
				if err := insertAward(ctx, tx, id, awardIDs["Strikeouts Title"]); err != nil {
					return err
				}
			}
		}
	} else {
		if eraLeader != 0 {
			if err := insertAward(ctx, tx, eraLeader, awardIDs["ERA Title"]); err != nil {
				return err
			}
		}
		for _, id := range wLeaders {
			if err := insertAward(ctx, tx, id, awardIDs["Wins Title"]); err != nil {
				return err
			}
		}
		for _, id := range kLeaders {
			if err := insertAward(ctx, tx, id, awardIDs["Strikeouts Title"]); err != nil {
				return err
			}
		}
	}

	if err := s.assignChampionshipAwards(ctx, tx, seasonID, awardIDs); err != nil {
		return err
	}
	return tx.Commit()
}

// AssignChampionshipAwards assigns League Champion and Conference Champion awards
// for the given season within an existing transaction. Safe to call when
// playoffs are incomplete — the season_champions view's completeness gate
// will simply return no rows and no awards will be inserted.
func (s *AwardStore) AssignChampionshipAwards(ctx context.Context, tx *sql.Tx, seasonID int64) error {
	awardIDs, err := loadAutoAwardIDs(ctx, tx)
	if err != nil {
		return fmt.Errorf("loading auto award IDs for championship assignment: %w", err)
	}
	return s.assignChampionshipAwards(ctx, tx, seasonID, awardIDs)
}

func (s *AwardStore) assignChampionshipAwards(ctx context.Context, tx *sql.Tx, seasonID int64, awardIDs map[string]int64) error {
	assign := func(query string, awardID int64) error {
		var histID sql.NullInt64
		if err := tx.QueryRowContext(ctx, query, seasonID).Scan(&histID); err != nil && err != sql.ErrNoRows {
			return err
		}
		if !histID.Valid || histID.Int64 == 0 {
			return nil
		}
		_, err := tx.ExecContext(ctx, `
INSERT OR IGNORE INTO player_season_awards (player_season_id, award_id)
SELECT pst.player_season_id, ?
FROM player_season_teams pst
JOIN player_seasons ps ON ps.id = pst.player_season_id
WHERE pst.team_history_id = ?
  AND ps.season_id         = ?
  AND pst.sort_order       = 0
`, awardID, histID.Int64, seasonID)
		return err
	}
	if id := awardIDs["League Champion"]; id != 0 {
		if err := assign(`SELECT winner_history_id FROM season_champions WHERE season_id = ?`, id); err != nil {
			return fmt.Errorf("assigning League Champion for season %d: %w", seasonID, err)
		}
	}
	if id := awardIDs["Conference Champion"]; id != 0 {
		if err := assign(`SELECT runner_up_history_id FROM season_conference_champions WHERE season_id = ?`, id); err != nil {
			return fmt.Errorf("assigning Conference Champion for season %d: %w", seasonID, err)
		}
	}
	return nil
}

// ── Champion team ─────────────────────────────────────────────────────────────

// GetSeasonChampionTeam returns the team_season_history_id of the playoff
// champion, or nil if the champion cannot yet be determined.
func (s *AwardStore) GetSeasonChampionTeam(ctx context.Context, seasonID int64) (*int64, error) {
	var id sql.NullInt64
	if err := s.db.QueryRowContext(ctx,
		`SELECT winner_history_id FROM season_champions WHERE season_id = ?`,
		seasonID,
	).Scan(&id); err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("getting champion for season %d: %w", seasonID, err)
	}
	if !id.Valid {
		return nil, nil
	}
	v := id.Int64
	return &v, nil
}

// ── Hall of Fame ──────────────────────────────────────────────────────────────

// GetHoFCandidates returns players who appeared in at least one season but are
// absent from the most recently imported season (retired) and not yet inducted.
func (s *AwardStore) GetHoFCandidates(ctx context.Context) ([]models.HoFCandidate, error) {
	return s.queryHoFPlayers(ctx, false)
}

// GetHoFInducted returns all players currently marked as Hall of Famers.
func (s *AwardStore) GetHoFInducted(ctx context.Context) ([]models.HoFCandidate, error) {
	return s.queryHoFPlayers(ctx, true)
}

func (s *AwardStore) queryHoFPlayers(ctx context.Context, inducted bool) ([]models.HoFCandidate, error) {
	var filter string
	if inducted {
		filter = `WHERE p.is_hall_of_famer = 1`
	} else {
		filter = `
WHERE p.is_hall_of_famer = 0
  AND p.id NOT IN (
      SELECT player_id FROM player_seasons
      WHERE season_id = (SELECT id FROM seasons ORDER BY season_num DESC LIMIT 1)
  )`
	}

	rows, err := s.db.QueryContext(ctx, `
SELECT
    p.id,
    p.first_name,
    p.last_name,
    p.is_hall_of_famer,
    MIN(s.season_num)            AS first_season,
    MAX(s.season_num)            AS last_season,
    COUNT(DISTINCT ps.season_id) AS seasons,
    COALESCE(SUM(b.hits),          0) AS hits,
    COALESCE(SUM(b.home_runs),     0) AS home_runs,
    COALESCE(SUM(b.rbi),           0) AS rbi,
    COALESCE(SUM(b.stolen_bases),  0) AS stolen_bases,
    COALESCE(SUM(b.at_bats),       0) AS at_bats,
    COALESCE(SUM(b.walks),         0) AS walks,
    COALESCE(SUM(pi.wins),         0) AS wins,
    COALESCE(SUM(pi.losses),       0) AS losses,
    COALESCE(SUM(pi.outs_pitched), 0) AS outs_pitched,
    COALESCE(SUM(pi.strikeouts),   0) AS strikeouts,
    COALESCE(SUM(pi.earned_runs),  0) AS earned_runs
FROM players p
JOIN player_seasons ps ON ps.player_id = p.id
JOIN seasons s         ON s.id = ps.season_id
LEFT JOIN player_season_batting_stats b
    ON b.player_season_id = ps.id AND b.is_regular_season = 1
LEFT JOIN player_season_pitching_stats pi
    ON pi.player_season_id = ps.id AND pi.is_regular_season = 1
`+filter+`
GROUP BY p.id
ORDER BY hits DESC, home_runs DESC
`)
	if err != nil {
		return nil, fmt.Errorf("querying HoF players (inducted=%v): %w", inducted, err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.HoFCandidate
	for rows.Next() {
		var c models.HoFCandidate
		var hofInt int
		if err := rows.Scan(
			&c.PlayerID, &c.FirstName, &c.LastName, &hofInt,
			&c.FirstSeason, &c.LastSeason, &c.Seasons,
			&c.Hits, &c.HomeRuns, &c.RBI, &c.StolenBases, &c.AtBats, &c.Walks,
			&c.Wins, &c.Losses, &c.OutsPitched, &c.Strikeouts, &c.EarnedRuns,
		); err != nil {
			return nil, fmt.Errorf("scanning HoF candidate: %w", err)
		}
		c.IsHallOfFamer = hofInt == 1
		out = append(out, c)
	}
	return out, rows.Err()
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func scanAwards(rows *sql.Rows) ([]models.Award, error) {
	var out []models.Award
	for rows.Next() {
		var a models.Award
		var omit, bat, pitch, field, playoff, userAssign, builtIn int
		if err := rows.Scan(
			&a.ID, &a.Name, &a.OriginalName, &a.Importance, &omit,
			&bat, &pitch, &field, &playoff, &userAssign, &builtIn,
		); err != nil {
			return nil, fmt.Errorf("scanning award: %w", err)
		}
		a.OmitFromGroupings = omit == 1
		a.IsBattingAward = bat == 1
		a.IsPitchingAward = pitch == 1
		a.IsFieldingAward = field == 1
		a.IsPlayoffAward = playoff == 1
		a.IsUserAssignable = userAssign == 1
		a.IsBuiltIn = builtIn == 1
		out = append(out, a)
	}
	return out, rows.Err()
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// loadAutoAwardIDs returns a name→id map for all auto-computed (non-user-assignable) awards.
func loadAutoAwardIDs(ctx context.Context, tx *sql.Tx) (map[string]int64, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT id, name FROM awards WHERE is_user_assignable = 0`)
	if err != nil {
		return nil, fmt.Errorf("loading auto award IDs: %w", err)
	}
	defer func() { _ = rows.Close() }()
	m := map[string]int64{}
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("scanning auto award: %w", err)
		}
		m[name] = id
	}
	return m, rows.Err()
}

// queryQualifiedLeader returns the single player_season_id of the BA or ERA
// leader who meets the qualification threshold (at_bats >= numGames*3 for
// batting; outs_pitched >= numGames*3 for pitching). Returns 0 if no one qualifies.
func queryQualifiedLeader(ctx context.Context, tx *sql.Tx, seasonID int64, numGames int, isPitching bool) (int64, error) {
	threshold := numGames * 3
	var q string
	if !isPitching {
		q = `
SELECT ps.id
FROM player_seasons ps
JOIN player_season_batting_stats b ON b.player_season_id = ps.id
WHERE ps.season_id      = ?
  AND b.is_regular_season = 1
  AND b.at_bats          >= ?
  AND b.at_bats          >  0
ORDER BY CAST(b.hits AS REAL) / b.at_bats DESC, b.at_bats DESC
LIMIT 1`
	} else {
		q = `
SELECT ps.id
FROM player_seasons ps
JOIN player_season_pitching_stats p ON p.player_season_id = ps.id
WHERE ps.season_id      = ?
  AND p.is_regular_season = 1
  AND p.outs_pitched     >= ?
  AND p.outs_pitched     >  0
ORDER BY CAST(p.earned_runs AS REAL) / p.outs_pitched ASC, p.outs_pitched DESC
LIMIT 1`
	}
	var id sql.NullInt64
	if err := tx.QueryRowContext(ctx, q, seasonID, threshold).Scan(&id); err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("querying qualified leader (pitching=%v): %w", isPitching, err)
	}
	return id.Int64, nil
}

// queryMaxLeaders returns all player_season_ids tied for the max value of the
// given counting stat column. isPitching=true queries pitching stats.
// Allowed column names: home_runs, rbi (batting); wins, strikeouts (pitching).
func queryMaxLeaders(ctx context.Context, tx *sql.Tx, seasonID int64, col string, isPitching bool) ([]int64, error) {
	table := "player_season_batting_stats"
	if isPitching {
		table = "player_season_pitching_stats"
	}
	// col is a controlled internal value (not user input), so interpolation is safe.
	q := fmt.Sprintf(`
SELECT ps.id
FROM player_seasons ps
JOIN %s stat ON stat.player_season_id = ps.id
WHERE ps.season_id       = ?
  AND stat.is_regular_season = 1
  AND stat.%s = (
      SELECT MAX(stat2.%s)
      FROM player_seasons ps2
      JOIN %s stat2 ON stat2.player_season_id = ps2.id
      WHERE ps2.season_id        = ?
        AND stat2.is_regular_season = 1
  )
`, table, col, col, table)
	rows, err := tx.QueryContext(ctx, q, seasonID, seasonID)
	if err != nil {
		return nil, fmt.Errorf("querying max leaders for %s: %w", col, err)
	}
	defer func() { _ = rows.Close() }()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning leader id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func insertAward(ctx context.Context, tx *sql.Tx, playerSeasonID, awardID int64) error {
	if awardID == 0 {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
INSERT OR IGNORE INTO player_season_awards (player_season_id, award_id)
VALUES (?, ?)
`, playerSeasonID, awardID)
	if err != nil {
		return fmt.Errorf("inserting award %d for player-season %d: %w", awardID, playerSeasonID, err)
	}
	return nil
}

