package store

import (
	"context"
	"fmt"
)

// StatRecordQueryStore runs raw per-player-season counting stat queries used by
// service.StatRecordsService to compute league leaders and all-time records.
type StatRecordQueryStore struct {
	db DBTX
}

func NewStatRecordQueryStore(db DBTX) *StatRecordQueryStore {
	return &StatRecordQueryStore{db: db}
}

// BattingCountRow holds one player-season's batting counting stats.
// PlateAppearances and NumGames are included so lower-is-better stats (e.g. K)
// can be gated to qualified batters (PA ≥ numGames×3.1) before finding the minimum.
type BattingCountRow struct {
	PlayerID         int64
	SeasonNum        int
	GamesPlayed      int
	AtBats           int
	Hits             int
	Doubles          int
	Triples          int
	HomeRuns         int
	RBI              int
	StolenBases      int
	Walks            int
	Strikeouts       int
	PlateAppearances int
	NumGames         int
}

// PitchingCountRow holds one player-season's pitching counting stats.
// NumGames is included so lower-is-better stats (H, ER, BB allowed) can be gated
// to qualified pitchers (outs ≥ numGames×3) before finding the minimum.
type PitchingCountRow struct {
	PlayerID     int64
	SeasonNum    int
	Games        int
	GamesStarted int
	Wins         int
	Losses       int
	Saves        int
	OutsPitched  int
	Strikeouts   int
	Walks        int
	HitsAllowed  int
	EarnedRuns   int
	NumGames     int
}

// GetBattingCountRows returns one row per player-season with batting counting stats.
// isRegularSeason=true for regular season, false for playoffs.
func (s *StatRecordQueryStore) GetBattingCountRows(ctx context.Context, isRegularSeason bool) ([]BattingCountRow, error) {
	isReg := 0
	if isRegularSeason {
		isReg = 1
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT p.id, s.season_num,
       b.games_played, b.at_bats, b.hits, b.doubles, b.triples,
       b.home_runs, b.rbi, b.stolen_bases, b.walks, b.strikeouts,
       b.plate_appearances, s.num_games
FROM player_season_batting_stats b
JOIN player_seasons ps ON ps.id = b.player_season_id
JOIN seasons s         ON s.id  = ps.season_id
JOIN players p         ON p.id  = ps.player_id
WHERE b.is_regular_season = ?`, isReg)
	if err != nil {
		return nil, fmt.Errorf("GetBattingCountRows: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []BattingCountRow
	for rows.Next() {
		var r BattingCountRow
		if err := rows.Scan(
			&r.PlayerID, &r.SeasonNum,
			&r.GamesPlayed, &r.AtBats, &r.Hits, &r.Doubles, &r.Triples,
			&r.HomeRuns, &r.RBI, &r.StolenBases, &r.Walks, &r.Strikeouts,
			&r.PlateAppearances, &r.NumGames,
		); err != nil {
			return nil, fmt.Errorf("GetBattingCountRows scan: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetPitchingCountRows returns one row per player-season with pitching counting stats.
// isRegularSeason=true for regular season, false for playoffs.
func (s *StatRecordQueryStore) GetPitchingCountRows(ctx context.Context, isRegularSeason bool) ([]PitchingCountRow, error) {
	isReg := 0
	if isRegularSeason {
		isReg = 1
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT p.id, s.season_num,
       pit.games, pit.games_started, pit.wins, pit.losses, pit.saves,
       pit.outs_pitched, pit.strikeouts, pit.walks, pit.hits_allowed, pit.earned_runs,
       s.num_games
FROM player_season_pitching_stats pit
JOIN player_seasons ps ON ps.id = pit.player_season_id
JOIN seasons s         ON s.id  = ps.season_id
JOIN players p         ON p.id  = ps.player_id
WHERE pit.is_regular_season = ?`, isReg)
	if err != nil {
		return nil, fmt.Errorf("GetPitchingCountRows: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []PitchingCountRow
	for rows.Next() {
		var r PitchingCountRow
		if err := rows.Scan(
			&r.PlayerID, &r.SeasonNum,
			&r.Games, &r.GamesStarted, &r.Wins, &r.Losses, &r.Saves,
			&r.OutsPitched, &r.Strikeouts, &r.Walks, &r.HitsAllowed, &r.EarnedRuns,
			&r.NumGames,
		); err != nil {
			return nil, fmt.Errorf("GetPitchingCountRows scan: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ── Rate stat rows ────────────────────────────────────────────────────────────

// BattingRateRow holds one player-season's batting rate stats plus qualification fields.
type BattingRateRow struct {
	PlayerID         int64
	SeasonNum        int
	BA               *float64
	OBP              *float64
	SLG              *float64
	OPS              *float64
	ISO              *float64
	BABIP            *float64
	KPct             *float64
	BBPct            *float64
	ABPerHR          *float64
	OPSPlus          *float64
	SmbWAR           *float64
	PlateAppearances int
	NumGames         int
}

// PitchingRateRow holds one player-season's pitching rate stats plus qualification fields.
type PitchingRateRow struct {
	PlayerID    int64
	SeasonNum   int
	ERA         *float64
	WHIP        *float64
	K9          *float64
	BB9         *float64
	H9          *float64
	HR9         *float64
	KPerBB      *float64
	KPct        *float64
	WinPct      *float64
	PPerIP      *float64
	FIP         *float64
	ERAPlus     *float64
	FIPMinus    *float64
	SmbWAR      *float64
	OutsPitched int
	NumGames    int
}

// BattingCareerRateRow holds one player's career batting rate stats plus qualification field.
type BattingCareerRateRow struct {
	PlayerID int64
	BA       *float64
	OBP      *float64
	SLG      *float64
	OPS      *float64
	ISO      *float64
	BABIP    *float64
	KPct     *float64
	BBPct    *float64
	ABPerHR  *float64
	SmbWAR   *float64
	CareerPA int
	Hits     int // career hits; used with Walks for the unscaled PO BB+H OR-qualifier
	Walks    int
}

// PitchingCareerRateRow holds one player's career pitching rate stats plus qualification field.
type PitchingCareerRateRow struct {
	PlayerID    int64
	ERA         *float64
	WHIP        *float64
	K9          *float64
	BB9         *float64
	H9          *float64
	HR9         *float64
	KPerBB      *float64
	KPct        *float64
	WinPct      *float64
	PPerIP      *float64
	FIP         *float64
	SmbWAR      *float64
	OutsPitched int
	Wins        int // used with Losses for the unscaled PO decisions OR-qualifier
	Losses      int
}

// GetBattingRateRows returns one row per player-season with batting rate stats
// and the plate_appearances and num_games needed for qualification checks.
func (s *StatRecordQueryStore) GetBattingRateRows(ctx context.Context, isRegularSeason bool) ([]BattingRateRow, error) {
	isReg := 0
	if isRegularSeason {
		isReg = 1
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT p.id, s.season_num,
       b.ba, b.obp, b.slg, b.ops, b.iso, b.babip, b.k_pct, b.bb_pct, b.ab_per_hr, b.ops_plus, b.smb_war,
       b.plate_appearances, s.num_games
FROM player_season_batting_stats b
JOIN player_seasons ps ON ps.id = b.player_season_id
JOIN seasons s         ON s.id  = ps.season_id
JOIN players p         ON p.id  = ps.player_id
WHERE b.is_regular_season = ?`, isReg)
	if err != nil {
		return nil, fmt.Errorf("GetBattingRateRows: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []BattingRateRow
	for rows.Next() {
		var r BattingRateRow
		if err := rows.Scan(
			&r.PlayerID, &r.SeasonNum,
			&r.BA, &r.OBP, &r.SLG, &r.OPS, &r.ISO, &r.BABIP, &r.KPct, &r.BBPct, &r.ABPerHR, &r.OPSPlus, &r.SmbWAR,
			&r.PlateAppearances, &r.NumGames,
		); err != nil {
			return nil, fmt.Errorf("GetBattingRateRows scan: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetPitchingRateRows returns one row per player-season with pitching rate stats
// and the outs_pitched and num_games needed for qualification checks.
func (s *StatRecordQueryStore) GetPitchingRateRows(ctx context.Context, isRegularSeason bool) ([]PitchingRateRow, error) {
	isReg := 0
	if isRegularSeason {
		isReg = 1
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT p.id, s.season_num,
       pit.era, pit.whip, pit.k_per_9, pit.bb_per_9, pit.h_per_9, pit.hr_per_9,
       pit.k_per_bb, pit.k_pct, pit.win_pct, pit.p_per_ip, pit.fip, pit.era_plus, pit.fip_minus, pit.smb_war,
       pit.outs_pitched, s.num_games
FROM player_season_pitching_stats pit
JOIN player_seasons ps ON ps.id = pit.player_season_id
JOIN seasons s         ON s.id  = ps.season_id
JOIN players p         ON p.id  = ps.player_id
WHERE pit.is_regular_season = ?`, isReg)
	if err != nil {
		return nil, fmt.Errorf("GetPitchingRateRows: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []PitchingRateRow
	for rows.Next() {
		var r PitchingRateRow
		if err := rows.Scan(
			&r.PlayerID, &r.SeasonNum,
			&r.ERA, &r.WHIP, &r.K9, &r.BB9, &r.H9, &r.HR9,
			&r.KPerBB, &r.KPct, &r.WinPct, &r.PPerIP, &r.FIP, &r.ERAPlus, &r.FIPMinus, &r.SmbWAR,
			&r.OutsPitched, &r.NumGames,
		); err != nil {
			return nil, fmt.Errorf("GetPitchingRateRows scan: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetCareerBattingRateRows returns one row per player from player_career_batting_stats
// for the given stat_type ('regular_season' or 'playoffs').
// CareerPA is computed from the stored counting columns.
func (s *StatRecordQueryStore) GetCareerBattingRateRows(ctx context.Context, statType string) ([]BattingCareerRateRow, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT player_id,
       ba, obp, slg, ops, iso, babip, k_pct, bb_pct, ab_per_hr, smb_war,
       (at_bats + walks + hit_by_pitch + sac_hits + sac_flies) AS career_pa,
       hits, walks
FROM player_career_batting_stats
WHERE stat_type = ?`, statType)
	if err != nil {
		return nil, fmt.Errorf("GetCareerBattingRateRows: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []BattingCareerRateRow
	for rows.Next() {
		var r BattingCareerRateRow
		if err := rows.Scan(
			&r.PlayerID,
			&r.BA, &r.OBP, &r.SLG, &r.OPS, &r.ISO, &r.BABIP, &r.KPct, &r.BBPct, &r.ABPerHR, &r.SmbWAR,
			&r.CareerPA, &r.Hits, &r.Walks,
		); err != nil {
			return nil, fmt.Errorf("GetCareerBattingRateRows scan: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetCareerPitchingRateRows returns one row per player from player_career_pitching_stats
// for the given stat_type ('regular_season' or 'playoffs').
func (s *StatRecordQueryStore) GetCareerPitchingRateRows(ctx context.Context, statType string) ([]PitchingCareerRateRow, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT player_id,
       era, whip, k_per_9, bb_per_9, h_per_9, hr_per_9, k_per_bb, k_pct, win_pct, p_per_ip, fip, smb_war,
       outs_pitched, wins, losses
FROM player_career_pitching_stats
WHERE stat_type = ?`, statType)
	if err != nil {
		return nil, fmt.Errorf("GetCareerPitchingRateRows: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []PitchingCareerRateRow
	for rows.Next() {
		var r PitchingCareerRateRow
		if err := rows.Scan(
			&r.PlayerID,
			&r.ERA, &r.WHIP, &r.K9, &r.BB9, &r.H9, &r.HR9, &r.KPerBB, &r.KPct, &r.WinPct, &r.PPerIP, &r.FIP, &r.SmbWAR,
			&r.OutsPitched, &r.Wins, &r.Losses,
		); err != nil {
			return nil, fmt.Errorf("GetCareerPitchingRateRows scan: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetCareerQualificationThresholds returns the career batting/pitching
// qualification thresholds for the franchise. See the package-level
// GetCareerQualificationThresholds for details.
func (s *StatRecordQueryStore) GetCareerQualificationThresholds(ctx context.Context) (CareerQualificationThresholds, error) {
	return GetCareerQualificationThresholds(ctx, s.db)
}
