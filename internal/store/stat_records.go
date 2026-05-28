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
type BattingCountRow struct {
	PlayerID    int64
	SeasonNum   int
	GamesPlayed int
	AtBats      int
	Hits        int
	Doubles     int
	Triples     int
	HomeRuns    int
	RBI         int
	StolenBases int
	Walks       int
	Strikeouts  int
}

// PitchingCountRow holds one player-season's pitching counting stats.
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
       b.home_runs, b.rbi, b.stolen_bases, b.walks, b.strikeouts
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
       pit.outs_pitched, pit.strikeouts, pit.walks, pit.hits_allowed, pit.earned_runs
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
		); err != nil {
			return nil, fmt.Errorf("GetPitchingCountRows scan: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
