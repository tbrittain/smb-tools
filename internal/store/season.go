package store

import (
	"context"
	"fmt"
)

// Season is the companion DB representation of a franchise season.
type Season struct {
	ID       int // save game seasonID
	SeasonNum int
	NumGames int
}

// SeasonStore manages season records in the companion DB.
type SeasonStore struct {
	db DBTX
}

func NewSeasonStore(db DBTX) *SeasonStore {
	return &SeasonStore{db: db}
}

// Upsert inserts or replaces a season record. Safe to call multiple times
// for the same season (idempotent).
func (s *SeasonStore) Upsert(ctx context.Context, season Season) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO seasons (id, season_num, num_games)
		VALUES (?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			season_num = excluded.season_num,
			num_games  = excluded.num_games
	`, season.ID, season.SeasonNum, season.NumGames)
	if err != nil {
		return fmt.Errorf("upserting season %d: %w", season.ID, err)
	}
	return nil
}

// GetByID returns the season with the given ID, or sql.ErrNoRows.
func (s *SeasonStore) GetByID(ctx context.Context, id int) (Season, error) {
	var season Season
	err := s.db.QueryRowContext(ctx,
		`SELECT id, season_num, num_games FROM seasons WHERE id = ?`, id,
	).Scan(&season.ID, &season.SeasonNum, &season.NumGames)
	if err != nil {
		return Season{}, fmt.Errorf("getting season %d: %w", id, err)
	}
	return season, nil
}

// List returns all seasons ordered by season number ascending.
func (s *SeasonStore) List(ctx context.Context) ([]Season, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, season_num, num_games FROM seasons ORDER BY season_num ASC`)
	if err != nil {
		return nil, fmt.Errorf("listing seasons: %w", err)
	}
	defer rows.Close()

	var seasons []Season
	for rows.Next() {
		var season Season
		if err := rows.Scan(&season.ID, &season.SeasonNum, &season.NumGames); err != nil {
			return nil, fmt.Errorf("scanning season: %w", err)
		}
		seasons = append(seasons, season)
	}
	return seasons, rows.Err()
}
