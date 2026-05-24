package store

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"

	"smb-tools/internal/models"
)

// GetSeasonAwardCandidates returns the full candidate set for the award delegation
// page: top overall batters/pitchers, rookies, by team, and by position.
//
// If no user-assignable awards exist for the season yet, award IDs are pre-populated
// with auto-suggestions (All-Star for top 2 per team, Silver Slugger for top 1 per
// position) and AutoSuggested is set to true on the result.
func (s *AwardStore) GetSeasonAwardCandidates(ctx context.Context, seasonID int64) (models.SeasonAwardCandidates, error) {
	var out models.SeasonAwardCandidates
	out.SeasonID = seasonID

	// Season number for display.
	if err := s.db.QueryRowContext(ctx,
		`SELECT season_num FROM seasons WHERE id = ?`, seasonID,
	).Scan(&out.SeasonNum); err != nil {
		return out, fmt.Errorf("loading season num: %w", err)
	}

	// Teams in this season.
	teams, err := s.getSeasonTeams(ctx, seasonID)
	if err != nil {
		return out, err
	}

	// Distinct fielding positions (non-pitcher) active this season.
	positions, err := s.getSeasonPositions(ctx, seasonID)
	if err != nil {
		return out, err
	}

	// ── Top 10 overall ────────────────────────────────────────────────────────
	out.TopBatters, err = s.queryBattingCandidates(ctx, seasonID, 10, 0, "", false)
	if err != nil {
		return out, fmt.Errorf("top batters: %w", err)
	}
	out.TopPitchers, err = s.queryPitchingCandidates(ctx, seasonID, 10, 0, false)
	if err != nil {
		return out, fmt.Errorf("top pitchers: %w", err)
	}

	// ── Top 10 rookies ────────────────────────────────────────────────────────
	out.TopRookieBatters, err = s.queryBattingCandidates(ctx, seasonID, 10, 0, "", true)
	if err != nil {
		return out, fmt.Errorf("rookie batters: %w", err)
	}
	out.TopRookiePitchers, err = s.queryPitchingCandidates(ctx, seasonID, 10, 0, true)
	if err != nil {
		return out, fmt.Errorf("rookie pitchers: %w", err)
	}

	// ── Top 5 per team ────────────────────────────────────────────────────────
	out.ByTeam = make([]models.TeamAwardCandidates, 0, len(teams))
	for _, t := range teams {
		batters, err := s.queryBattingCandidates(ctx, seasonID, 5, t.HistoryID, "", false)
		if err != nil {
			return out, fmt.Errorf("team %d batters: %w", t.HistoryID, err)
		}
		pitchers, err := s.queryPitchingCandidates(ctx, seasonID, 5, t.HistoryID, false)
		if err != nil {
			return out, fmt.Errorf("team %d pitchers: %w", t.HistoryID, err)
		}
		out.ByTeam = append(out.ByTeam, models.TeamAwardCandidates{
			HistoryID: t.HistoryID,
			TeamName:  t.TeamName,
			Batters:   batters,
			Pitchers:  pitchers,
		})
	}

	// ── Top 5 per position ────────────────────────────────────────────────────
	out.ByPosition = make([]models.PositionAwardCandidates, 0, len(positions))
	for _, pos := range positions {
		batters, err := s.queryBattingCandidates(ctx, seasonID, 5, 0, pos, false)
		if err != nil {
			return out, fmt.Errorf("position %s batters: %w", pos, err)
		}
		if len(batters) == 0 {
			continue
		}
		out.ByPosition = append(out.ByPosition, models.PositionAwardCandidates{
			Position: pos,
			Batters:  batters,
		})
	}

	// ── Playoff candidates ────────────────────────────────────────────────────
	out.PlayoffBatters, err = s.queryPlayoffBattingCandidates(ctx, seasonID, false)
	if err != nil {
		return out, fmt.Errorf("playoff batters: %w", err)
	}
	out.PlayoffPitchers, err = s.queryPlayoffPitchingCandidates(ctx, seasonID, false)
	if err != nil {
		return out, fmt.Errorf("playoff pitchers: %w", err)
	}
	out.ChampionBatters, err = s.queryPlayoffBattingCandidates(ctx, seasonID, true)
	if err != nil {
		return out, fmt.Errorf("champion batters: %w", err)
	}
	out.ChampionPitchers, err = s.queryPlayoffPitchingCandidates(ctx, seasonID, true)
	if err != nil {
		return out, fmt.Errorf("champion pitchers: %w", err)
	}

	// ── Load existing user-assignable awards ──────────────────────────────────
	hasUserAwards, err := s.seasonHasUserAwards(ctx, seasonID)
	if err != nil {
		return out, err
	}

	if hasUserAwards {
		// Load current award assignments from the DB.
		awardMap, err := s.loadSeasonUserAwardMap(ctx, seasonID)
		if err != nil {
			return out, err
		}
		populateAllAwardIDs(&out, awardMap)
	} else {
		// Auto-suggest: All-Star for top 2 per team, Silver Slugger for top 1 per position.
		awardMap := map[int64][]int64{}
		if err := s.applyAutoSuggest(ctx, &out, awardMap); err != nil {
			return out, err
		}
		populateAllAwardIDs(&out, awardMap)
		out.AutoSuggested = true
	}

	return out, nil
}

// SubmitMultiplePlayerAwards replaces user-assignable awards for a set of
// player-seasons in a single transaction. Awards not present in entries are
// cleared; auto-computed awards are always preserved.
func (s *AwardStore) SubmitMultiplePlayerAwards(ctx context.Context, entries []models.PlayerAwardEntry) error {
	db, ok := s.db.(*sql.DB)
	if !ok {
		return fmt.Errorf("SubmitMultiplePlayerAwards requires *sql.DB")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, e := range entries {
		if _, err = tx.ExecContext(ctx, `
DELETE FROM player_season_awards
WHERE player_season_id = ?
  AND award_id IN (SELECT id FROM awards WHERE is_user_assignable = 1)
`, e.PlayerSeasonID); err != nil {
			return fmt.Errorf("clearing awards for player-season %d: %w", e.PlayerSeasonID, err)
		}
		for _, aid := range e.AwardIDs {
			if _, err = tx.ExecContext(ctx, `
INSERT OR IGNORE INTO player_season_awards (player_season_id, award_id)
VALUES (?, ?)
`, e.PlayerSeasonID, aid); err != nil {
				return fmt.Errorf("inserting award %d for player-season %d: %w", aid, e.PlayerSeasonID, err)
			}
		}
	}

	return tx.Commit()
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// getSeasonTeams returns all teams (with history IDs) for the given season.
func (s *AwardStore) getSeasonTeams(ctx context.Context, seasonID int64) ([]models.TeamSeasonRef, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT tsh.id, tsh.team_name
FROM team_season_history tsh
WHERE tsh.season_id = ?
ORDER BY tsh.team_name ASC
`, seasonID)
	if err != nil {
		return nil, fmt.Errorf("loading season teams: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []models.TeamSeasonRef
	for rows.Next() {
		var t models.TeamSeasonRef
		if err := rows.Scan(&t.HistoryID, &t.TeamName); err != nil {
			return nil, fmt.Errorf("scanning team: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// getSeasonPositions returns distinct primary fielding positions (non-pitcher)
// that appear in the season, ordered by a conventional position sequence.
func (s *AwardStore) getSeasonPositions(ctx context.Context, seasonID int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT DISTINCT ps.primary_position
FROM player_seasons ps
JOIN player_season_batting_stats b ON b.player_season_id = ps.id
WHERE ps.season_id        = ?
  AND b.is_regular_season  = 1
  AND b.at_bats            > 0
  AND ps.primary_position NOT IN ('P', '')
ORDER BY ps.primary_position ASC
`, seasonID)
	if err != nil {
		return nil, fmt.Errorf("loading season positions: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var positions []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		positions = append(positions, p)
	}
	return positions, rows.Err()
}

// queryBattingCandidates fetches batting candidates with optional filters.
// teamHistoryID=0 means all teams; position="" means all positions.
func (s *AwardStore) queryBattingCandidates(
	ctx context.Context,
	seasonID int64,
	limit int,
	teamHistoryID int64,
	position string,
	rookiesOnly bool,
) ([]models.BattingCandidate, error) {
	conds := []string{
		"ps.season_id = ?",
		"b.is_regular_season = 1",
		"b.at_bats > 0",
	}
	args := []any{seasonID}

	if teamHistoryID != 0 {
		conds = append(conds, "tsh.id = ?")
		args = append(args, teamHistoryID)
	}
	if position != "" {
		conds = append(conds, "ps.primary_position = ?")
		args = append(args, position)
	}
	if rookiesOnly {
		conds = append(conds, `ps.player_id NOT IN (
			SELECT ps2.player_id FROM player_seasons ps2
			JOIN seasons s2 ON s2.id = ps2.season_id
			WHERE s2.season_num < (SELECT season_num FROM seasons WHERE id = ?)
		)`)
		args = append(args, seasonID)
	}

	q := `
SELECT
    ps.id,
    p.id,
    p.first_name,
    p.last_name,
    COALESCE(tsh.team_name, '') AS team_name,
    ps.primary_position,
    ps.pitcher_role,
    COALESCE(b.at_bats,      0),
    COALESCE(b.hits,         0),
    COALESCE(b.home_runs,    0),
    COALESCE(b.rbi,          0),
    COALESCE(b.walks,        0),
    COALESCE(b.runs,         0),
    COALESCE(b.stolen_bases, 0),
    COALESCE(b.strikeouts,   0),
    COALESCE(b.doubles,      0),
    COALESCE(b.triples,      0),
    COALESCE(b.ba,  0.0),
    COALESCE(b.obp, 0.0),
    COALESCE(b.slg, 0.0),
    COALESCE(b.ops, 0.0),
    b.ops_plus,
    b.smb_war
FROM player_seasons ps
JOIN players p ON p.id = ps.player_id
LEFT JOIN player_season_teams pst ON pst.player_season_id = ps.id AND pst.sort_order = 0
LEFT JOIN team_season_history tsh ON tsh.id = pst.team_history_id
JOIN v_batting_stats b ON b.player_season_id = ps.id
WHERE ` + strings.Join(conds, " AND ") + `
ORDER BY COALESCE(b.smb_war, -9999.0) DESC
LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("querying batting candidates: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.BattingCandidate
	for rows.Next() {
		var c models.BattingCandidate
		if err := rows.Scan(
			&c.PlayerSeasonID, &c.PlayerID, &c.FirstName, &c.LastName,
			&c.TeamName, &c.PrimaryPos, &c.PitcherRole,
			&c.AtBats, &c.Hits, &c.HomeRuns, &c.RBI, &c.Walks, &c.Runs,
			&c.StolenBases, &c.Strikeouts, &c.Doubles, &c.Triples,
			&c.BA, &c.OBP, &c.SLG, &c.OPS,
			&c.OPSPlus, &c.SmbWAR,
		); err != nil {
			return nil, fmt.Errorf("scanning batting candidate: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// queryPitchingCandidates fetches pitching candidates with optional filters.
// teamHistoryID=0 means all teams.
func (s *AwardStore) queryPitchingCandidates(
	ctx context.Context,
	seasonID int64,
	limit int,
	teamHistoryID int64,
	rookiesOnly bool,
) ([]models.PitchingCandidate, error) {
	conds := []string{
		"ps.season_id = ?",
		"pit.is_regular_season = 1",
		"pit.outs_pitched > 0",
	}
	args := []any{seasonID}

	if teamHistoryID != 0 {
		conds = append(conds, "tsh.id = ?")
		args = append(args, teamHistoryID)
	}
	if rookiesOnly {
		conds = append(conds, `ps.player_id NOT IN (
			SELECT ps2.player_id FROM player_seasons ps2
			JOIN seasons s2 ON s2.id = ps2.season_id
			WHERE s2.season_num < (SELECT season_num FROM seasons WHERE id = ?)
		)`)
		args = append(args, seasonID)
	}

	q := `
SELECT
    ps.id,
    p.id,
    p.first_name,
    p.last_name,
    COALESCE(tsh.team_name, '') AS team_name,
    ps.primary_position,
    ps.pitcher_role,
    COALESCE(pit.wins,             0),
    COALESCE(pit.losses,           0),
    COALESCE(pit.saves,            0),
    COALESCE(pit.outs_pitched,     0),
    COALESCE(pit.hits_allowed,     0),
    COALESCE(pit.earned_runs,      0),
    COALESCE(pit.walks,            0),
    COALESCE(pit.strikeouts,       0),
    COALESCE(pit.home_runs_allowed,0),
    COALESCE(pit.complete_games,   0),
    COALESCE(pit.shutouts,         0),
    COALESCE(pit.era,    0.0),
    COALESCE(pit.whip,   0.0),
    COALESCE(pit.k_per_9,  0.0),
    COALESCE(pit.bb_per_9, 0.0),
    COALESCE(pit.h_per_9,  0.0),
    COALESCE(pit.hr_per_9, 0.0),
    COALESCE(pit.k_per_bb, 0.0),
    pit.era_plus,
    pit.fip_minus,
    pit.smb_war
FROM player_seasons ps
JOIN players p ON p.id = ps.player_id
LEFT JOIN player_season_teams pst ON pst.player_season_id = ps.id AND pst.sort_order = 0
LEFT JOIN team_season_history tsh ON tsh.id = pst.team_history_id
JOIN v_pitching_stats pit ON pit.player_season_id = ps.id
WHERE ` + strings.Join(conds, " AND ") + `
ORDER BY COALESCE(pit.smb_war, -9999.0) DESC
LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("querying pitching candidates: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.PitchingCandidate
	for rows.Next() {
		var c models.PitchingCandidate
		if err := rows.Scan(
			&c.PlayerSeasonID, &c.PlayerID, &c.FirstName, &c.LastName,
			&c.TeamName, &c.PrimaryPos, &c.PitcherRole,
			&c.Wins, &c.Losses, &c.Saves, &c.OutsPitched,
			&c.HitsAllowed, &c.EarnedRuns, &c.Walks, &c.Strikeouts,
			&c.HomeRunsAllowed, &c.CompleteGames, &c.Shutouts,
			&c.ERA, &c.WHIP, &c.K9, &c.BB9, &c.H9, &c.HR9, &c.KPerBB,
			&c.ERAPlus, &c.FIPMinus, &c.SmbWAR,
		); err != nil {
			return nil, fmt.Errorf("scanning pitching candidate: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// seasonHasUserAwards returns true if any user-assignable awards have already
// been assigned for this season.
func (s *AwardStore) seasonHasUserAwards(ctx context.Context, seasonID int64) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM player_season_awards psa
JOIN awards a ON a.id = psa.award_id
JOIN player_seasons ps ON ps.id = psa.player_season_id
WHERE ps.season_id       = ?
  AND a.is_user_assignable = 1
`, seasonID).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("checking user awards: %w", err)
	}
	return n > 0, nil
}

// loadSeasonUserAwardMap returns a map of playerSeasonID → []awardID for all
// user-assignable awards in the season.
func (s *AwardStore) loadSeasonUserAwardMap(ctx context.Context, seasonID int64) (map[int64][]int64, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT psa.player_season_id, psa.award_id
FROM player_season_awards psa
JOIN awards a ON a.id = psa.award_id
JOIN player_seasons ps ON ps.id = psa.player_season_id
WHERE ps.season_id       = ?
  AND a.is_user_assignable = 1
`, seasonID)
	if err != nil {
		return nil, fmt.Errorf("loading user award map: %w", err)
	}
	defer func() { _ = rows.Close() }()
	m := map[int64][]int64{}
	for rows.Next() {
		var psID, aID int64
		if err := rows.Scan(&psID, &aID); err != nil {
			return nil, err
		}
		m[psID] = append(m[psID], aID)
	}
	return m, rows.Err()
}

// applyAutoSuggest pre-populates awardMap with:
//   - All-Star for top 2 batters and top 2 pitchers per team
//   - Silver Slugger for top 1 batter per position
func (s *AwardStore) applyAutoSuggest(ctx context.Context, candidates *models.SeasonAwardCandidates, awardMap map[int64][]int64) error {
	// Fetch award IDs we need.
	var allStarID, silverSluggerID int64
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name FROM awards WHERE name IN ('All-Star', 'Silver Slugger')`)
	if err != nil {
		return fmt.Errorf("loading auto-suggest award IDs: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		switch name {
		case "All-Star":
			allStarID = id
		case "Silver Slugger":
			silverSluggerID = id
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Top 2 batters per team → All-Star.
	for _, team := range candidates.ByTeam {
		for i, b := range team.Batters {
			if i >= 2 {
				break
			}
			addToAwardMap(awardMap, b.PlayerSeasonID, allStarID)
		}
		for i, p := range team.Pitchers {
			if i >= 2 {
				break
			}
			addToAwardMap(awardMap, p.PlayerSeasonID, allStarID)
		}
	}

	// Top 1 batter per position → Silver Slugger.
	for _, pos := range candidates.ByPosition {
		if len(pos.Batters) > 0 {
			addToAwardMap(awardMap, pos.Batters[0].PlayerSeasonID, silverSluggerID)
		}
	}

	return nil
}

func addToAwardMap(m map[int64][]int64, psID, awardID int64) {
	if !slices.Contains(m[psID], awardID) {
		m[psID] = append(m[psID], awardID)
	}
}

// populateAllAwardIDs sets AwardIDs on every candidate from the map.
func populateAllAwardIDs(out *models.SeasonAwardCandidates, m map[int64][]int64) {
	populateBatters(out.TopBatters, m)
	populatePitchers(out.TopPitchers, m)
	populateBatters(out.TopRookieBatters, m)
	populatePitchers(out.TopRookiePitchers, m)
	for i := range out.ByTeam {
		populateBatters(out.ByTeam[i].Batters, m)
		populatePitchers(out.ByTeam[i].Pitchers, m)
	}
	for i := range out.ByPosition {
		populateBatters(out.ByPosition[i].Batters, m)
	}
	populateBatters(out.PlayoffBatters, m)
	populatePitchers(out.PlayoffPitchers, m)
	populateBatters(out.ChampionBatters, m)
	populatePitchers(out.ChampionPitchers, m)
}

func populateBatters(batters []models.BattingCandidate, m map[int64][]int64) {
	for i := range batters {
		if ids, ok := m[batters[i].PlayerSeasonID]; ok {
			batters[i].AwardIDs = ids
		} else {
			batters[i].AwardIDs = []int64{}
		}
	}
}

func populatePitchers(pitchers []models.PitchingCandidate, m map[int64][]int64) {
	for i := range pitchers {
		if ids, ok := m[pitchers[i].PlayerSeasonID]; ok {
			pitchers[i].AwardIDs = ids
		} else {
			pitchers[i].AwardIDs = []int64{}
		}
	}
}

// ── Playoff candidate queries ─────────────────────────────────────────────────

// queryPlayoffBattingCandidates returns top-10 batters who had playoff at-bats,
// sorted by playoff OPS descending. When championsOnly is true the result is
// restricted to the championship-winning team.
func (s *AwardStore) queryPlayoffBattingCandidates(ctx context.Context, seasonID int64, championsOnly bool) ([]models.BattingCandidate, error) {
	champFilter := ""
	if championsOnly {
		champFilter = "  AND tsh.id IN (SELECT winner_history_id FROM champion WHERE season_id = ?)\n"
	}
	q := championCTE + `
SELECT
    ps.id,
    p.id,
    p.first_name,
    p.last_name,
    COALESCE(tsh.team_name, '') AS team_name,
    ps.primary_position,
    ps.pitcher_role,
    COALESCE(b.at_bats,      0),
    COALESCE(b.hits,         0),
    COALESCE(b.home_runs,    0),
    COALESCE(b.rbi,          0),
    COALESCE(b.walks,        0),
    COALESCE(b.runs,         0),
    COALESCE(b.stolen_bases, 0),
    COALESCE(b.strikeouts,   0),
    COALESCE(b.doubles,      0),
    COALESCE(b.triples,      0),
    COALESCE(b.ba,  0.0),
    COALESCE(b.obp, 0.0),
    COALESCE(b.slg, 0.0),
    COALESCE(b.ops, 0.0),
    CASE WHEN tsh.id IN (SELECT winner_history_id FROM champion WHERE season_id = ?) THEN 1 ELSE 0 END AS is_champion_team
FROM player_seasons ps
JOIN players p ON p.id = ps.player_id
LEFT JOIN player_season_teams pst ON pst.player_season_id = ps.id AND pst.sort_order = 0
LEFT JOIN team_season_history tsh ON tsh.id = pst.team_history_id
JOIN v_batting_stats b ON b.player_season_id = ps.id
WHERE ps.season_id       = ?
  AND b.is_regular_season = 0
  AND b.at_bats           > 0
` + champFilter + `ORDER BY COALESCE(b.ops, 0.0) DESC
LIMIT 10
`
	args := []any{seasonID, seasonID} // CTE arg + is_champion_team subquery arg
	if championsOnly {
		args = append(args, seasonID) // extra arg for the champFilter subquery
	}
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("querying playoff batting candidates: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.BattingCandidate
	for rows.Next() {
		var c models.BattingCandidate
		var isChamp int
		if err := rows.Scan(
			&c.PlayerSeasonID, &c.PlayerID, &c.FirstName, &c.LastName,
			&c.TeamName, &c.PrimaryPos, &c.PitcherRole,
			&c.AtBats, &c.Hits, &c.HomeRuns, &c.RBI, &c.Walks, &c.Runs,
			&c.StolenBases, &c.Strikeouts, &c.Doubles, &c.Triples,
			&c.BA, &c.OBP, &c.SLG, &c.OPS,
			&isChamp,
		); err != nil {
			return nil, fmt.Errorf("scanning playoff batting candidate: %w", err)
		}
		c.IsChampionTeam = isChamp == 1
		out = append(out, c)
	}
	return out, rows.Err()
}

// queryPlayoffPitchingCandidates returns top-10 pitchers who recorded playoff outs,
// sorted by ERA ascending. When championsOnly is true the result is restricted to
// the championship-winning team.
func (s *AwardStore) queryPlayoffPitchingCandidates(ctx context.Context, seasonID int64, championsOnly bool) ([]models.PitchingCandidate, error) {
	champFilter := ""
	if championsOnly {
		champFilter = "  AND tsh.id IN (SELECT winner_history_id FROM champion WHERE season_id = ?)\n"
	}
	q := championCTE + `
SELECT
    ps.id,
    p.id,
    p.first_name,
    p.last_name,
    COALESCE(tsh.team_name, '') AS team_name,
    ps.primary_position,
    ps.pitcher_role,
    COALESCE(pit.wins,             0),
    COALESCE(pit.losses,           0),
    COALESCE(pit.saves,            0),
    COALESCE(pit.outs_pitched,     0),
    COALESCE(pit.hits_allowed,     0),
    COALESCE(pit.earned_runs,      0),
    COALESCE(pit.walks,            0),
    COALESCE(pit.strikeouts,       0),
    COALESCE(pit.home_runs_allowed,0),
    COALESCE(pit.complete_games,   0),
    COALESCE(pit.shutouts,         0),
    COALESCE(pit.era,    0.0),
    COALESCE(pit.whip,   0.0),
    COALESCE(pit.k_per_9,  0.0),
    COALESCE(pit.bb_per_9, 0.0),
    COALESCE(pit.h_per_9,  0.0),
    COALESCE(pit.hr_per_9, 0.0),
    COALESCE(pit.k_per_bb, 0.0),
    CASE WHEN tsh.id IN (SELECT winner_history_id FROM champion WHERE season_id = ?) THEN 1 ELSE 0 END AS is_champion_team
FROM player_seasons ps
JOIN players p ON p.id = ps.player_id
LEFT JOIN player_season_teams pst ON pst.player_season_id = ps.id AND pst.sort_order = 0
LEFT JOIN team_season_history tsh ON tsh.id = pst.team_history_id
JOIN v_pitching_stats pit ON pit.player_season_id = ps.id
WHERE ps.season_id        = ?
  AND pit.is_regular_season = 0
  AND pit.outs_pitched      > 0
` + champFilter + `ORDER BY
    CASE WHEN COALESCE(pit.era, 0.0) = 0.0 THEN 1 ELSE 0 END ASC,
    COALESCE(pit.era, 999.0) ASC
LIMIT 10
`
	args := []any{seasonID, seasonID}
	if championsOnly {
		args = append(args, seasonID)
	}
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("querying playoff pitching candidates: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []models.PitchingCandidate
	for rows.Next() {
		var c models.PitchingCandidate
		var isChamp int
		if err := rows.Scan(
			&c.PlayerSeasonID, &c.PlayerID, &c.FirstName, &c.LastName,
			&c.TeamName, &c.PrimaryPos, &c.PitcherRole,
			&c.Wins, &c.Losses, &c.Saves, &c.OutsPitched,
			&c.HitsAllowed, &c.EarnedRuns, &c.Walks, &c.Strikeouts,
			&c.HomeRunsAllowed, &c.CompleteGames, &c.Shutouts,
			&c.ERA, &c.WHIP, &c.K9, &c.BB9, &c.H9, &c.HR9, &c.KPerBB,
			&isChamp,
		); err != nil {
			return nil, fmt.Errorf("scanning playoff pitching candidate: %w", err)
		}
		c.IsChampionTeam = isChamp == 1
		out = append(out, c)
	}
	return out, rows.Err()
}
