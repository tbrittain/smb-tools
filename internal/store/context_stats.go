package store

import (
	"context"
	"fmt"
)

// ContextStatsStore provides raw DB operations for computing and persisting
// league-level context stats (OPS+, ERA+, FIP, FIP-, smbWAR). All computation
// logic lives in the service layer; this package only handles SQL.
type ContextStatsStore struct {
	db DBTX
}

func NewContextStatsStore(db DBTX) *ContextStatsStore {
	return &ContextStatsStore{db: db}
}

// LeagueBattingTotals holds summed batting counting stats for an entire season.
type LeagueBattingTotals struct {
	AB, Hits, Doubles, Triples, HomeRuns, Walks, HBP, SacHits, SacFlies int64
}

// LeaguePitchingTotals holds summed pitching counting stats for an entire season.
type LeaguePitchingTotals struct {
	OutsPitched, EarnedRuns, HRAllowed, BBAllowed, HBPAllowed, KPitched int64
}

// GetLeagueBattingTotals aggregates all batting counting stats for a season.
func (s *ContextStatsStore) GetLeagueBattingTotals(
	ctx context.Context, seasonID int64, isRegularSeason bool,
) (LeagueBattingTotals, error) {
	isReg := 1
	if !isRegularSeason {
		isReg = 0
	}
	var t LeagueBattingTotals
	err := s.db.QueryRowContext(ctx, `
		SELECT
		    COALESCE(SUM(b.at_bats),        0),
		    COALESCE(SUM(b.hits),            0),
		    COALESCE(SUM(b.doubles),         0),
		    COALESCE(SUM(b.triples),         0),
		    COALESCE(SUM(b.home_runs),       0),
		    COALESCE(SUM(b.walks),           0),
		    COALESCE(SUM(b.hit_by_pitch),    0),
		    COALESCE(SUM(b.sac_hits),        0),
		    COALESCE(SUM(b.sac_flies),       0)
		FROM player_season_batting_stats b
		JOIN player_seasons ps ON ps.id = b.player_season_id
		WHERE ps.season_id = ? AND b.is_regular_season = ?
	`, seasonID, isReg).Scan(
		&t.AB, &t.Hits, &t.Doubles, &t.Triples, &t.HomeRuns,
		&t.Walks, &t.HBP, &t.SacHits, &t.SacFlies,
	)
	if err != nil {
		return LeagueBattingTotals{}, fmt.Errorf("GetLeagueBattingTotals: %w", err)
	}
	return t, nil
}

// GetLeaguePitchingTotals aggregates all pitching counting stats for a season.
func (s *ContextStatsStore) GetLeaguePitchingTotals(
	ctx context.Context, seasonID int64, isRegularSeason bool,
) (LeaguePitchingTotals, error) {
	isReg := 1
	if !isRegularSeason {
		isReg = 0
	}
	var t LeaguePitchingTotals
	err := s.db.QueryRowContext(ctx, `
		SELECT
		    COALESCE(SUM(p.outs_pitched),      0),
		    COALESCE(SUM(p.earned_runs),        0),
		    COALESCE(SUM(p.home_runs_allowed),  0),
		    COALESCE(SUM(p.walks),              0),
		    COALESCE(SUM(p.hit_batters),        0),
		    COALESCE(SUM(p.strikeouts),         0)
		FROM player_season_pitching_stats p
		JOIN player_seasons ps ON ps.id = p.player_season_id
		WHERE ps.season_id = ? AND p.is_regular_season = ?
	`, seasonID, isReg).Scan(
		&t.OutsPitched, &t.EarnedRuns, &t.HRAllowed,
		&t.BBAllowed, &t.HBPAllowed, &t.KPitched,
	)
	if err != nil {
		return LeaguePitchingTotals{}, fmt.Errorf("GetLeaguePitchingTotals: %w", err)
	}
	return t, nil
}

// LeagueSeasonStatsRecord is the row written to league_season_stats.
type LeagueSeasonStatsRecord struct {
	SeasonID         int64
	IsRegularSeason  bool
	BattingTotals    LeagueBattingTotals
	PitchingTotals   LeaguePitchingTotals
	LgOBP            *float64
	LgSLG            *float64
	LgERA            *float64
	FIPConstant      *float64
}

// UpsertLeagueSeasonStats persists the league aggregate row.
func (s *ContextStatsStore) UpsertLeagueSeasonStats(ctx context.Context, rec LeagueSeasonStatsRecord) error {
	isReg := 1
	if !rec.IsRegularSeason {
		isReg = 0
	}
	bt := rec.BattingTotals
	pt := rec.PitchingTotals
	lgPA := bt.AB + bt.Walks + bt.HBP + bt.SacHits + bt.SacFlies

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO league_season_stats (
		    season_id, is_regular_season,
		    total_plate_appearances, total_at_bats, total_hits,
		    total_doubles, total_triples, total_home_runs,
		    total_walks, total_hbp, total_sac_flies,
		    total_outs_pitched, total_earned_runs,
		    total_hr_allowed, total_bb_allowed, total_hbp_allowed, total_k_pitched,
		    lg_obp, lg_slg, lg_era, fip_constant
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(season_id, is_regular_season) DO UPDATE SET
		    total_plate_appearances = excluded.total_plate_appearances,
		    total_at_bats           = excluded.total_at_bats,
		    total_hits              = excluded.total_hits,
		    total_doubles           = excluded.total_doubles,
		    total_triples           = excluded.total_triples,
		    total_home_runs         = excluded.total_home_runs,
		    total_walks             = excluded.total_walks,
		    total_hbp               = excluded.total_hbp,
		    total_sac_flies         = excluded.total_sac_flies,
		    total_outs_pitched      = excluded.total_outs_pitched,
		    total_earned_runs       = excluded.total_earned_runs,
		    total_hr_allowed        = excluded.total_hr_allowed,
		    total_bb_allowed        = excluded.total_bb_allowed,
		    total_hbp_allowed       = excluded.total_hbp_allowed,
		    total_k_pitched         = excluded.total_k_pitched,
		    lg_obp                  = excluded.lg_obp,
		    lg_slg                  = excluded.lg_slg,
		    lg_era                  = excluded.lg_era,
		    fip_constant            = excluded.fip_constant
	`,
		rec.SeasonID, isReg,
		lgPA, bt.AB, bt.Hits, bt.Doubles, bt.Triples, bt.HomeRuns,
		bt.Walks, bt.HBP, bt.SacFlies,
		pt.OutsPitched, pt.EarnedRuns,
		pt.HRAllowed, pt.BBAllowed, pt.HBPAllowed, pt.KPitched,
		rec.LgOBP, rec.LgSLG, rec.LgERA, rec.FIPConstant,
	)
	if err != nil {
		return fmt.Errorf("UpsertLeagueSeasonStats: %w", err)
	}
	return nil
}

// BattingContextRow holds the minimal stats needed to compute OPS+ and smbWAR.
type BattingContextRow struct {
	ID             int64
	AtBats         int
	Hits           int
	Doubles        int
	Triples        int
	HomeRuns       int
	Walks          int
	HitByPitch     int
	SacHits        int
	SacFlies       int
	StolenBases    int
	CaughtStealing int
}

// GetBattingRowsForContext fetches all batting stat rows for a season that need
// context-stat computation.
func (s *ContextStatsStore) GetBattingRowsForContext(
	ctx context.Context, seasonID int64, isRegularSeason bool,
) ([]BattingContextRow, error) {
	isReg := 1
	if !isRegularSeason {
		isReg = 0
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT
		    b.id, b.at_bats, b.hits, b.doubles, b.triples, b.home_runs,
		    b.walks, b.hit_by_pitch, b.sac_hits, b.sac_flies,
		    b.stolen_bases, b.caught_stealing
		FROM player_season_batting_stats b
		JOIN player_seasons ps ON ps.id = b.player_season_id
		WHERE ps.season_id = ? AND b.is_regular_season = ?
	`, seasonID, isReg)
	if err != nil {
		return nil, fmt.Errorf("GetBattingRowsForContext: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []BattingContextRow
	for rows.Next() {
		var r BattingContextRow
		if err := rows.Scan(
			&r.ID, &r.AtBats, &r.Hits, &r.Doubles, &r.Triples, &r.HomeRuns,
			&r.Walks, &r.HitByPitch, &r.SacHits, &r.SacFlies,
			&r.StolenBases, &r.CaughtStealing,
		); err != nil {
			return nil, fmt.Errorf("GetBattingRowsForContext scan: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// UpdateBattingContextStats sets ops_plus and smb_war on one batting stat row.
func (s *ContextStatsStore) UpdateBattingContextStats(
	ctx context.Context, id int64, opsPlus, smbWAR *float64,
) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE player_season_batting_stats SET ops_plus = ?, smb_war = ? WHERE id = ?`,
		opsPlus, smbWAR, id,
	)
	if err != nil {
		return fmt.Errorf("UpdateBattingContextStats(%d): %w", id, err)
	}
	return nil
}

// PitchingContextRow holds the minimal stats needed to compute ERA+, FIP, FIP-, and smbWAR.
type PitchingContextRow struct {
	ID              int64
	OutsPitched     int
	EarnedRuns      int
	HomeRunsAllowed int
	Walks           int
	HitBatters      int
	Strikeouts      int
}

// GetPitchingRowsForContext fetches all pitching stat rows for a season that need
// context-stat computation.
func (s *ContextStatsStore) GetPitchingRowsForContext(
	ctx context.Context, seasonID int64, isRegularSeason bool,
) ([]PitchingContextRow, error) {
	isReg := 1
	if !isRegularSeason {
		isReg = 0
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT
		    p.id, p.outs_pitched, p.earned_runs, p.home_runs_allowed,
		    p.walks, p.hit_batters, p.strikeouts
		FROM player_season_pitching_stats p
		JOIN player_seasons ps ON ps.id = p.player_season_id
		WHERE ps.season_id = ? AND p.is_regular_season = ?
	`, seasonID, isReg)
	if err != nil {
		return nil, fmt.Errorf("GetPitchingRowsForContext: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []PitchingContextRow
	for rows.Next() {
		var r PitchingContextRow
		if err := rows.Scan(
			&r.ID, &r.OutsPitched, &r.EarnedRuns, &r.HomeRunsAllowed,
			&r.Walks, &r.HitBatters, &r.Strikeouts,
		); err != nil {
			return nil, fmt.Errorf("GetPitchingRowsForContext scan: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// UpdatePitchingContextStats sets era_plus, fip, fip_minus, and smb_war on one pitching stat row.
func (s *ContextStatsStore) UpdatePitchingContextStats(
	ctx context.Context, id int64, eraPlus, fip, fipMinus, smbWAR *float64,
) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE player_season_pitching_stats SET era_plus = ?, fip = ?, fip_minus = ?, smb_war = ? WHERE id = ?`,
		eraPlus, fip, fipMinus, smbWAR, id,
	)
	if err != nil {
		return fmt.Errorf("UpdatePitchingContextStats(%d): %w", id, err)
	}
	return nil
}
