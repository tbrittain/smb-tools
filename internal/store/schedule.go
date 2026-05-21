package store

import (
	"context"
	"fmt"
)

// ScheduleGame is one regular season game.
type ScheduleGame struct {
	SeasonID           int64
	GameNumber         int
	Day                int
	HomeTeamHistoryID  int64
	AwayTeamHistoryID  int64
	HomePitcherSeasonID *int64
	AwayPitcherSeasonID *int64
	HomeScore          *int
	AwayScore          *int
}

// PlayoffGame is one playoff game with series context.
type PlayoffGame struct {
	SeasonID           int64
	SeriesNumber       int
	GameNumber         int
	HomeTeamHistoryID  int64
	AwayTeamHistoryID  int64
	HomePitcherSeasonID *int64
	AwayPitcherSeasonID *int64
	HomeScore          *int
	AwayScore          *int
}

// ScheduleStore manages schedule records in the companion DB.
type ScheduleStore struct {
	db DBTX
}

func NewScheduleStore(db DBTX) *ScheduleStore {
	return &ScheduleStore{db: db}
}

// UpsertGame inserts or replaces a regular season game record.
func (s *ScheduleStore) UpsertGame(ctx context.Context, g ScheduleGame) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO team_season_schedules (
			season_id, game_number, day,
			home_team_history_id, away_team_history_id,
			home_pitcher_season_id, away_pitcher_season_id,
			home_score, away_score
		) VALUES (?,?,?,?,?,?,?,?,?)
		ON CONFLICT(season_id, game_number) DO UPDATE SET
			day                      = excluded.day,
			home_team_history_id     = excluded.home_team_history_id,
			away_team_history_id     = excluded.away_team_history_id,
			home_pitcher_season_id   = excluded.home_pitcher_season_id,
			away_pitcher_season_id   = excluded.away_pitcher_season_id,
			home_score               = excluded.home_score,
			away_score               = excluded.away_score
	`,
		g.SeasonID, g.GameNumber, g.Day,
		g.HomeTeamHistoryID, g.AwayTeamHistoryID,
		g.HomePitcherSeasonID, g.AwayPitcherSeasonID,
		g.HomeScore, g.AwayScore,
	)
	if err != nil {
		return fmt.Errorf("upserting schedule game %d (season %d): %w", g.GameNumber, g.SeasonID, err)
	}
	return nil
}

// UpsertPlayoffGame inserts or replaces a playoff game record.
func (s *ScheduleStore) UpsertPlayoffGame(ctx context.Context, g PlayoffGame) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO team_playoff_schedules (
			season_id, series_number, game_number,
			home_team_history_id, away_team_history_id,
			home_pitcher_season_id, away_pitcher_season_id,
			home_score, away_score
		) VALUES (?,?,?,?,?,?,?,?,?)
		ON CONFLICT DO NOTHING
	`,
		g.SeasonID, g.SeriesNumber, g.GameNumber,
		g.HomeTeamHistoryID, g.AwayTeamHistoryID,
		g.HomePitcherSeasonID, g.AwayPitcherSeasonID,
		g.HomeScore, g.AwayScore,
	)
	if err != nil {
		return fmt.Errorf("upserting playoff game %d series %d (season %d): %w", g.GameNumber, g.SeriesNumber, g.SeasonID, err)
	}
	return nil
}

// DeleteSeasonSchedule removes all schedule records for a season (used during re-import).
func (s *ScheduleStore) DeleteSeasonSchedule(ctx context.Context, seasonID int64) error {
	if _, err := s.db.ExecContext(ctx,
		`DELETE FROM team_season_schedules WHERE season_id = ?`, seasonID); err != nil {
		return fmt.Errorf("deleting season schedules for season %d: %w", seasonID, err)
	}
	if _, err := s.db.ExecContext(ctx,
		`DELETE FROM team_playoff_schedules WHERE season_id = ?`, seasonID); err != nil {
		return fmt.Errorf("deleting playoff schedules for season %d: %w", seasonID, err)
	}
	return nil
}
