package store

import (
	"context"
	"fmt"
)

// Season is the companion DB representation of a franchise season.
type Season struct {
	ID                int64  // autoincrement companion DB PK
	LeagueGUID        string // leagueGUID from the franchise_source that produced this season
	SaveGameSeasonID  int    // raw t_seasons.id from the save game
	SeasonNum         int    // display number: save game season num + source season_offset
	NumGames          int
}

// SeasonStore manages season records in the companion DB.
type SeasonStore struct {
	db DBTX
}

func NewSeasonStore(db DBTX) *SeasonStore {
	return &SeasonStore{db: db}
}

// Upsert inserts or replaces a season record. Safe to call multiple times for
// the same (league_guid, save_game_season_id) pair — idempotent.
func (s *SeasonStore) Upsert(ctx context.Context, season Season) (int64, error) {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO seasons (league_guid, save_game_season_id, season_num, num_games)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(league_guid, save_game_season_id) DO UPDATE SET
			season_num = excluded.season_num,
			num_games  = excluded.num_games
	`, season.LeagueGUID, season.SaveGameSeasonID, season.SeasonNum, season.NumGames)
	if err != nil {
		return 0, fmt.Errorf("upserting season (league=%s sgID=%d): %w",
			season.LeagueGUID, season.SaveGameSeasonID, err)
	}
	var id int64
	if err := s.db.QueryRowContext(ctx,
		`SELECT id FROM seasons WHERE league_guid = ? AND save_game_season_id = ?`,
		season.LeagueGUID, season.SaveGameSeasonID,
	).Scan(&id); err != nil {
		return 0, fmt.Errorf("fetching season id after upsert: %w", err)
	}
	return id, nil
}

// GetByID returns the season with the given companion DB id.
func (s *SeasonStore) GetByID(ctx context.Context, id int64) (Season, error) {
	var season Season
	err := s.db.QueryRowContext(ctx,
		`SELECT id, league_guid, save_game_season_id, season_num, num_games
		 FROM seasons WHERE id = ?`, id,
	).Scan(&season.ID, &season.LeagueGUID, &season.SaveGameSeasonID,
		&season.SeasonNum, &season.NumGames)
	if err != nil {
		return Season{}, fmt.Errorf("getting season %d: %w", id, err)
	}
	return season, nil
}

// List returns all seasons ordered by season number ascending.
func (s *SeasonStore) List(ctx context.Context) ([]Season, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, league_guid, save_game_season_id, season_num, num_games
		 FROM seasons ORDER BY season_num ASC`)
	if err != nil {
		return nil, fmt.Errorf("listing seasons: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var seasons []Season
	for rows.Next() {
		var season Season
		if err := rows.Scan(&season.ID, &season.LeagueGUID, &season.SaveGameSeasonID,
			&season.SeasonNum, &season.NumGames); err != nil {
			return nil, fmt.Errorf("scanning season: %w", err)
		}
		seasons = append(seasons, season)
	}
	return seasons, rows.Err()
}
