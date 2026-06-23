package store

import (
	"context"
	"fmt"
	"math/bits"
)

// Season is the companion DB representation of a franchise season.
type Season struct {
	ID               int64  // autoincrement companion DB PK
	LeagueGUID       string // leagueGUID from the franchise_source that produced this season
	SaveGameSeasonID int    // raw t_seasons.id from the save game
	SeasonNum        int    // display number: save game season num + source season_offset
	NumGames         int
	InningsPerGame   int // t_seasons.innings from the save game; defaults to 9 for legacy-migrated seasons
}

// SeasonStore manages season records in the companion DB.
type SeasonStore struct {
	db DBTX
}

func NewSeasonStore(db DBTX) *SeasonStore {
	return &SeasonStore{db: db}
}

// Upsert inserts or replaces a season record, keyed on season_num.
// Conflicting on season_num rather than (league_guid, save_game_season_id)
// allows a live sync to supersede a legacy-imported row for the same franchise
// season: the seasons.id is preserved so all child FK references remain valid.
func (s *SeasonStore) Upsert(ctx context.Context, season Season) (int64, error) {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO seasons (league_guid, save_game_season_id, season_num, num_games, innings_per_game)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(season_num) DO UPDATE SET
			league_guid         = excluded.league_guid,
			save_game_season_id = excluded.save_game_season_id,
			num_games           = excluded.num_games,
			innings_per_game    = excluded.innings_per_game
	`, season.LeagueGUID, season.SaveGameSeasonID, season.SeasonNum, season.NumGames, season.InningsPerGame)
	if err != nil {
		return 0, fmt.Errorf("upserting season (league=%s sgID=%d): %w",
			season.LeagueGUID, season.SaveGameSeasonID, err)
	}
	var id int64
	if err := s.db.QueryRowContext(ctx,
		`SELECT id FROM seasons WHERE season_num = ?`,
		season.SeasonNum,
	).Scan(&id); err != nil {
		return 0, fmt.Errorf("fetching season id after upsert: %w", err)
	}
	return id, nil
}

// UpdatePlayoffConfig persists the playoff bracket configuration (rounds and
// series length) for the given season. Both values come from t_playoffs in the
// save game and remain NULL for seasons imported via the legacy path until
// InferAndSetPlayoffConfig backfills them.
func (s *SeasonStore) UpdatePlayoffConfig(ctx context.Context, seasonID int64, rounds, seriesLength int) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE seasons SET playoff_rounds = ?, playoff_series_length = ? WHERE id = ?`,
		rounds, seriesLength, seasonID,
	)
	if err != nil {
		return fmt.Errorf("updating playoff config for season %d: %w", seasonID, err)
	}
	return nil
}

// InferAndSetPlayoffConfig derives playoff_rounds and playoff_series_length from
// the playoff games already stored for the season and persists the inferred values.
// Used for the legacy import path where t_playoffs is unavailable.
//
// Inference rules:
//   - playoff_rounds  = bits.Len(total distinct series count)
//   - playoff_series_length = 2*maxWins - 1, where maxWins is the highest win
//     count recorded by any single team in any single series (scored games only).
//
// No-ops silently if there are no playoff games or no scored games for the season.
func (s *SeasonStore) InferAndSetPlayoffConfig(ctx context.Context, seasonID int64) error {
	var totalSeries int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT series_number) FROM team_playoff_schedules WHERE season_id = ?`,
		seasonID,
	).Scan(&totalSeries); err != nil || totalSeries == 0 {
		return err
	}

	playoffRounds := bits.Len(uint(totalSeries))

	var maxWins int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(team_wins), 0) FROM (
			SELECT
				CASE WHEN home_score > away_score THEN home_team_history_id
				     ELSE away_team_history_id END AS winner_id,
				COUNT(*) AS team_wins
			FROM team_playoff_schedules
			WHERE season_id = ?
			  AND home_score IS NOT NULL AND away_score IS NOT NULL
			  AND home_score != away_score
			GROUP BY series_number,
			         CASE WHEN home_score > away_score THEN home_team_history_id
			              ELSE away_team_history_id END
		)
	`, seasonID).Scan(&maxWins); err != nil {
		return fmt.Errorf("inferring series length for season %d: %w", seasonID, err)
	}
	if maxWins == 0 {
		return nil
	}

	seriesLength := 2*maxWins - 1
	return s.UpdatePlayoffConfig(ctx, seasonID, playoffRounds, seriesLength)
}

// GetBySeasonNum returns the season with the given display season number.
// Returns sql.ErrNoRows if no season with that number exists.
func (s *SeasonStore) GetBySeasonNum(ctx context.Context, seasonNum int) (Season, error) {
	var season Season
	err := s.db.QueryRowContext(ctx,
		`SELECT id, league_guid, save_game_season_id, season_num, num_games, innings_per_game
		 FROM seasons WHERE season_num = ?`, seasonNum,
	).Scan(&season.ID, &season.LeagueGUID, &season.SaveGameSeasonID,
		&season.SeasonNum, &season.NumGames, &season.InningsPerGame)
	if err != nil {
		return Season{}, fmt.Errorf("getting season by num %d: %w", seasonNum, err)
	}
	return season, nil
}

// GetByID returns the season with the given companion DB id.
func (s *SeasonStore) GetByID(ctx context.Context, id int64) (Season, error) {
	var season Season
	err := s.db.QueryRowContext(ctx,
		`SELECT id, league_guid, save_game_season_id, season_num, num_games, innings_per_game
		 FROM seasons WHERE id = ?`, id,
	).Scan(&season.ID, &season.LeagueGUID, &season.SaveGameSeasonID,
		&season.SeasonNum, &season.NumGames, &season.InningsPerGame)
	if err != nil {
		return Season{}, fmt.Errorf("getting season %d: %w", id, err)
	}
	return season, nil
}

// List returns all seasons ordered by season number ascending.
func (s *SeasonStore) List(ctx context.Context) ([]Season, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, league_guid, save_game_season_id, season_num, num_games, innings_per_game
		 FROM seasons ORDER BY season_num ASC`)
	if err != nil {
		return nil, fmt.Errorf("listing seasons: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var seasons []Season
	for rows.Next() {
		var season Season
		if err := rows.Scan(&season.ID, &season.LeagueGUID, &season.SaveGameSeasonID,
			&season.SeasonNum, &season.NumGames, &season.InningsPerGame); err != nil {
			return nil, fmt.Errorf("scanning season: %w", err)
		}
		seasons = append(seasons, season)
	}
	return seasons, rows.Err()
}
