package store

import (
	"context"
	"fmt"
)

// Team is the base team record identified by its save game GUID.
type Team struct {
	ID       int64
	GameGUID string // hex-encoded GUID from t_teams
}

// TeamSeasonHistory is the per-season snapshot of a team's performance.
type TeamSeasonHistory struct {
	ID             int64
	TeamID         int64
	SeasonID       int
	TeamName       string
	DivisionName   string
	ConferenceName string
	Budget         int
	Payroll        int
	Wins           int
	Losses         int
	GamesBack      float64
	RunsFor        int
	RunsAgainst    int
	TotalPower     int
	TotalContact   int
	TotalSpeed     int
	TotalFielding  int
	TotalArm       int
	TotalVelocity  int
	TotalJunk      int
	TotalAccuracy  int
	PlayoffSeed    *int
	PlayoffWins    *int
	PlayoffLosses  *int
	PlayoffRunsFor *int
	PlayoffRunsAgainst *int
}

// TeamHistoryStore manages team and team_season_history records.
type TeamHistoryStore struct {
	db DBTX
}

func NewTeamHistoryStore(db DBTX) *TeamHistoryStore {
	return &TeamHistoryStore{db: db}
}

// UpsertTeam inserts or returns the existing team for a given save game GUID.
// Returns the internal team ID.
func (s *TeamHistoryStore) UpsertTeam(ctx context.Context, gameGUID string) (int64, error) {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO teams (game_guid) VALUES (?)`, gameGUID)
	if err != nil {
		return 0, fmt.Errorf("upserting team %s: %w", gameGUID, err)
	}
	var id int64
	if err := s.db.QueryRowContext(ctx,
		`SELECT id FROM teams WHERE game_guid = ?`, gameGUID,
	).Scan(&id); err != nil {
		return 0, fmt.Errorf("getting team id for %s: %w", gameGUID, err)
	}
	return id, nil
}

// UpsertSeasonHistory inserts or replaces a team's season history record.
// Returns the internal history ID.
func (s *TeamHistoryStore) UpsertSeasonHistory(ctx context.Context, h TeamSeasonHistory) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO team_season_history (
			team_id, season_id, team_name, division_name, conference_name,
			budget, payroll,
			wins, losses, games_back, runs_for, runs_against,
			total_power, total_contact, total_speed, total_fielding, total_arm,
			total_velocity, total_junk, total_accuracy,
			playoff_seed, playoff_wins, playoff_losses,
			playoff_runs_for, playoff_runs_against
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(team_id, season_id) DO UPDATE SET
			team_name       = excluded.team_name,
			division_name   = excluded.division_name,
			conference_name = excluded.conference_name,
			budget          = excluded.budget,
			payroll         = excluded.payroll,
			wins            = excluded.wins,
			losses          = excluded.losses,
			games_back      = excluded.games_back,
			runs_for        = excluded.runs_for,
			runs_against    = excluded.runs_against,
			total_power     = excluded.total_power,
			total_contact   = excluded.total_contact,
			total_speed     = excluded.total_speed,
			total_fielding  = excluded.total_fielding,
			total_arm       = excluded.total_arm,
			total_velocity  = excluded.total_velocity,
			total_junk      = excluded.total_junk,
			total_accuracy  = excluded.total_accuracy,
			playoff_seed    = excluded.playoff_seed,
			playoff_wins    = excluded.playoff_wins,
			playoff_losses  = excluded.playoff_losses,
			playoff_runs_for     = excluded.playoff_runs_for,
			playoff_runs_against = excluded.playoff_runs_against
	`,
		h.TeamID, h.SeasonID, h.TeamName, h.DivisionName, h.ConferenceName,
		h.Budget, h.Payroll,
		h.Wins, h.Losses, h.GamesBack, h.RunsFor, h.RunsAgainst,
		h.TotalPower, h.TotalContact, h.TotalSpeed, h.TotalFielding, h.TotalArm,
		h.TotalVelocity, h.TotalJunk, h.TotalAccuracy,
		h.PlayoffSeed, h.PlayoffWins, h.PlayoffLosses,
		h.PlayoffRunsFor, h.PlayoffRunsAgainst,
	)
	if err != nil {
		return 0, fmt.Errorf("upserting team season history (team=%d season=%d): %w", h.TeamID, h.SeasonID, err)
	}
	// ON CONFLICT DO UPDATE doesn't change the rowid — fetch it explicitly.
	var id int64
	if err := s.db.QueryRowContext(ctx,
		`SELECT id FROM team_season_history WHERE team_id = ? AND season_id = ?`,
		h.TeamID, h.SeasonID,
	).Scan(&id); err != nil {
		return 0, fmt.Errorf("getting team history id: %w", err)
	}
	_ = res
	return id, nil
}

// GetHistoryID returns the team_season_history.id for a given team and season.
func (s *TeamHistoryStore) GetHistoryID(ctx context.Context, teamID int64, seasonID int) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx,
		`SELECT id FROM team_season_history WHERE team_id = ? AND season_id = ?`,
		teamID, seasonID,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("getting history id for team=%d season=%d: %w", teamID, seasonID, err)
	}
	return id, nil
}
