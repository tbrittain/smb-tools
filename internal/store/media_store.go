package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"smb-tools/internal/models"
)

// MediaWithAssocs bundles a media item with its team-season and player associations.
type MediaWithAssocs struct {
	Media       models.Media
	TeamSeasons []MediaTeamSeasonInfo
	Players     []MediaPlayerInfo
}

// MediaTeamSeasonInfo is the denormalized team-season association detail used in DTOs.
type MediaTeamSeasonInfo struct {
	TeamHistoryID int64
	TeamName      string
	SeasonNum     int
}

// MediaPlayerInfo is the denormalized player association detail used in DTOs.
type MediaPlayerInfo struct {
	PlayerID  int64
	FirstName string
	LastName  string
}

// TeamPickerResult is a lightweight team record for the association picker.
type TeamPickerResult struct {
	TeamID   int64
	TeamName string
}

// TeamSeasonPickerResult is a lightweight team-season record for the association picker.
type TeamSeasonPickerResult struct {
	TeamHistoryID int64
	SeasonNum     int
}

// MediaStore handles reads and writes for the media, media_team_seasons, and media_players tables.
type MediaStore struct{}

func NewMediaStore() *MediaStore {
	return &MediaStore{}
}

// InsertMedia inserts a new media record.
func (s *MediaStore) InsertMedia(ctx context.Context, db DBTX, m models.Media) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO media (id, file_path, media_type, name, description, uploaded_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		m.ID,
		m.FilePath,
		string(m.MediaType),
		m.Name,
		m.Description,
		m.UploadedAt.UTC().Format("2006-01-02T15:04:05Z"),
	)
	if err != nil {
		return fmt.Errorf("inserting media: %w", err)
	}
	return nil
}

// InsertTeamSeasonAssoc inserts a media ↔ team-season association.
func (s *MediaStore) InsertTeamSeasonAssoc(ctx context.Context, db DBTX, assoc models.MediaTeamSeason) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO media_team_seasons (id, media_id, team_history_id)
		VALUES (?, ?, ?)
	`, assoc.ID, assoc.MediaID, assoc.TeamHistoryID)
	if err != nil {
		return fmt.Errorf("inserting media team-season association: %w", err)
	}
	return nil
}

// InsertPlayerAssoc inserts a media ↔ player association.
func (s *MediaStore) InsertPlayerAssoc(ctx context.Context, db DBTX, assoc models.MediaPlayer) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO media_players (id, media_id, player_id)
		VALUES (?, ?, ?)
	`, assoc.ID, assoc.MediaID, assoc.PlayerID)
	if err != nil {
		return fmt.Errorf("inserting media player association: %w", err)
	}
	return nil
}

// GetMediaForTeamSeason returns paginated media for a team-season, newest first,
// along with the total count of matching items.
func (s *MediaStore) GetMediaForTeamSeason(
	ctx context.Context, db DBTX,
	teamHistoryID int64, limit, offset int,
) ([]MediaWithAssocs, int, error) {
	var total int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT m.id)
		FROM media m
		JOIN media_team_seasons mts ON mts.media_id = m.id
		WHERE mts.team_history_id = ?
	`, teamHistoryID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting media for team-season %d: %w", teamHistoryID, err)
	}

	rows, err := db.QueryContext(ctx, `
		SELECT m.id, m.file_path, m.media_type, m.name, m.description, m.uploaded_at
		FROM media m
		JOIN media_team_seasons mts ON mts.media_id = m.id
		WHERE mts.team_history_id = ?
		ORDER BY m.uploaded_at DESC
		LIMIT ? OFFSET ?
	`, teamHistoryID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("querying media for team-season %d: %w", teamHistoryID, err)
	}
	defer func() { _ = rows.Close() }()

	items, err := s.scanMediaRows(rows)
	if err != nil {
		return nil, 0, err
	}

	if err := s.loadAssociations(ctx, db, items); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// GetMediaForPlayer returns paginated media for a player, newest first,
// along with the total count of matching items.
func (s *MediaStore) GetMediaForPlayer(
	ctx context.Context, db DBTX,
	playerID int64, limit, offset int,
) ([]MediaWithAssocs, int, error) {
	var total int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT m.id)
		FROM media m
		JOIN media_players mp ON mp.media_id = m.id
		WHERE mp.player_id = ?
	`, playerID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting media for player %d: %w", playerID, err)
	}

	rows, err := db.QueryContext(ctx, `
		SELECT m.id, m.file_path, m.media_type, m.name, m.description, m.uploaded_at
		FROM media m
		JOIN media_players mp ON mp.media_id = m.id
		WHERE mp.player_id = ?
		ORDER BY m.uploaded_at DESC
		LIMIT ? OFFSET ?
	`, playerID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("querying media for player %d: %w", playerID, err)
	}
	defer func() { _ = rows.Close() }()

	items, err := s.scanMediaRows(rows)
	if err != nil {
		return nil, 0, err
	}

	if err := s.loadAssociations(ctx, db, items); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// GetMediaWithAssocs returns the media record with all associations loaded.
func (s *MediaStore) GetMediaWithAssocs(ctx context.Context, db DBTX, mediaID string) (MediaWithAssocs, error) {
	m, err := s.GetMediaByID(ctx, db, mediaID)
	if err != nil {
		return MediaWithAssocs{}, err
	}
	item := MediaWithAssocs{Media: m}
	tsAssocs, err := s.getTeamSeasonAssocs(ctx, db, mediaID)
	if err != nil {
		return MediaWithAssocs{}, err
	}
	item.TeamSeasons = tsAssocs
	pAssocs, err := s.getPlayerAssocs(ctx, db, mediaID)
	if err != nil {
		return MediaWithAssocs{}, err
	}
	item.Players = pAssocs
	return item, nil
}

// GetMediaByID returns the media record with the given ID.
func (s *MediaStore) GetMediaByID(ctx context.Context, db DBTX, mediaID string) (models.Media, error) {
	var m models.Media
	var uploadedAt string
	err := db.QueryRowContext(ctx, `
		SELECT id, file_path, media_type, name, description, uploaded_at
		FROM media WHERE id = ?
	`, mediaID).Scan(&m.ID, &m.FilePath, (*string)(&m.MediaType), &m.Name, &m.Description, &uploadedAt)
	if err == sql.ErrNoRows {
		return models.Media{}, fmt.Errorf("media %s not found", mediaID)
	}
	if err != nil {
		return models.Media{}, fmt.Errorf("getting media %s: %w", mediaID, err)
	}
	m.UploadedAt, _ = time.Parse("2006-01-02T15:04:05Z", uploadedAt)
	return m, nil
}

// GetAssociationCounts returns the number of team-season and player associations for a media item.
func (s *MediaStore) GetAssociationCounts(ctx context.Context, db DBTX, mediaID string) (teamSeasonCount, playerCount int, err error) {
	err = db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM media_team_seasons WHERE media_id = ?`, mediaID,
	).Scan(&teamSeasonCount)
	if err != nil {
		return 0, 0, fmt.Errorf("counting team-season assocs for media %s: %w", mediaID, err)
	}
	err = db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM media_players WHERE media_id = ?`, mediaID,
	).Scan(&playerCount)
	if err != nil {
		return 0, 0, fmt.Errorf("counting player assocs for media %s: %w", mediaID, err)
	}
	return teamSeasonCount, playerCount, nil
}

// RemoveTeamSeasonAssoc removes a single media ↔ team-season association.
func (s *MediaStore) RemoveTeamSeasonAssoc(ctx context.Context, db DBTX, mediaID string, teamHistoryID int64) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM media_team_seasons WHERE media_id = ? AND team_history_id = ?`,
		mediaID, teamHistoryID,
	)
	if err != nil {
		return fmt.Errorf("removing team-season assoc for media %s: %w", mediaID, err)
	}
	return nil
}

// RemovePlayerAssoc removes a single media ↔ player association.
func (s *MediaStore) RemovePlayerAssoc(ctx context.Context, db DBTX, mediaID string, playerID int64) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM media_players WHERE media_id = ? AND player_id = ?`,
		mediaID, playerID,
	)
	if err != nil {
		return fmt.Errorf("removing player assoc for media %s: %w", mediaID, err)
	}
	return nil
}

// DeleteMedia removes all associations and the media record itself.
func (s *MediaStore) DeleteMedia(ctx context.Context, db DBTX, mediaID string) error {
	if _, err := db.ExecContext(ctx, `DELETE FROM media_team_seasons WHERE media_id = ?`, mediaID); err != nil {
		return fmt.Errorf("deleting team-season assocs for media %s: %w", mediaID, err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM media_players WHERE media_id = ?`, mediaID); err != nil {
		return fmt.Errorf("deleting player assocs for media %s: %w", mediaID, err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM media WHERE id = ?`, mediaID); err != nil {
		return fmt.Errorf("deleting media %s: %w", mediaID, err)
	}
	return nil
}

// SearchTeamsForPicker returns teams whose name matches the query (case-insensitive LIKE), up to 20 results.
func (s *MediaStore) SearchTeamsForPicker(ctx context.Context, db DBTX, query string) ([]TeamPickerResult, error) {
	pattern := "%" + query + "%"
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT t.id, tsh.team_name
		FROM teams t
		JOIN team_season_history tsh ON tsh.team_id = t.id
		WHERE tsh.team_name LIKE ?
		ORDER BY tsh.team_name
		LIMIT 20
	`, pattern)
	if err != nil {
		return nil, fmt.Errorf("searching teams for picker %q: %w", query, err)
	}
	defer func() { _ = rows.Close() }()

	var out []TeamPickerResult
	for rows.Next() {
		var r TeamPickerResult
		if err := rows.Scan(&r.TeamID, &r.TeamName); err != nil {
			return nil, fmt.Errorf("scanning team picker result: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetTeamSeasonsForPicker returns all (team_history_id, season_num) pairs for a team, ordered by season.
func (s *MediaStore) GetTeamSeasonsForPicker(ctx context.Context, db DBTX, teamID int64) ([]TeamSeasonPickerResult, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT tsh.id, s.season_num
		FROM team_season_history tsh
		JOIN seasons s ON s.id = tsh.season_id
		WHERE tsh.team_id = ?
		ORDER BY s.season_num ASC
	`, teamID)
	if err != nil {
		return nil, fmt.Errorf("getting seasons for team %d: %w", teamID, err)
	}
	defer func() { _ = rows.Close() }()

	var out []TeamSeasonPickerResult
	for rows.Next() {
		var r TeamSeasonPickerResult
		if err := rows.Scan(&r.TeamHistoryID, &r.SeasonNum); err != nil {
			return nil, fmt.Errorf("scanning team season picker result: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *MediaStore) scanMediaRows(rows *sql.Rows) ([]MediaWithAssocs, error) {
	var items []MediaWithAssocs
	for rows.Next() {
		var m models.Media
		var uploadedAt string
		if err := rows.Scan(&m.ID, &m.FilePath, (*string)(&m.MediaType), &m.Name, &m.Description, &uploadedAt); err != nil {
			return nil, fmt.Errorf("scanning media row: %w", err)
		}
		m.UploadedAt, _ = time.Parse("2006-01-02T15:04:05Z", uploadedAt)
		items = append(items, MediaWithAssocs{Media: m})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating media rows: %w", err)
	}
	return items, nil
}

func (s *MediaStore) loadAssociations(ctx context.Context, db DBTX, items []MediaWithAssocs) error {
	for i := range items {
		tsAssocs, err := s.getTeamSeasonAssocs(ctx, db, items[i].Media.ID)
		if err != nil {
			return err
		}
		items[i].TeamSeasons = tsAssocs

		pAssocs, err := s.getPlayerAssocs(ctx, db, items[i].Media.ID)
		if err != nil {
			return err
		}
		items[i].Players = pAssocs
	}
	return nil
}

func (s *MediaStore) getTeamSeasonAssocs(ctx context.Context, db DBTX, mediaID string) ([]MediaTeamSeasonInfo, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT mts.team_history_id, tsh.team_name, s.season_num
		FROM media_team_seasons mts
		JOIN team_season_history tsh ON tsh.id = mts.team_history_id
		JOIN seasons s ON s.id = tsh.season_id
		WHERE mts.media_id = ?
		ORDER BY s.season_num ASC
	`, mediaID)
	if err != nil {
		return nil, fmt.Errorf("loading team-season assocs for media %s: %w", mediaID, err)
	}
	defer func() { _ = rows.Close() }()

	var out []MediaTeamSeasonInfo
	for rows.Next() {
		var info MediaTeamSeasonInfo
		if err := rows.Scan(&info.TeamHistoryID, &info.TeamName, &info.SeasonNum); err != nil {
			return nil, fmt.Errorf("scanning team-season assoc: %w", err)
		}
		out = append(out, info)
	}
	return out, rows.Err()
}

func (s *MediaStore) getPlayerAssocs(ctx context.Context, db DBTX, mediaID string) ([]MediaPlayerInfo, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT mp.player_id, p.first_name, p.last_name
		FROM media_players mp
		JOIN players p ON p.id = mp.player_id
		WHERE mp.media_id = ?
		ORDER BY p.last_name, p.first_name
	`, mediaID)
	if err != nil {
		return nil, fmt.Errorf("loading player assocs for media %s: %w", mediaID, err)
	}
	defer func() { _ = rows.Close() }()

	var out []MediaPlayerInfo
	for rows.Next() {
		var info MediaPlayerInfo
		if err := rows.Scan(&info.PlayerID, &info.FirstName, &info.LastName); err != nil {
			return nil, fmt.Errorf("scanning player assoc: %w", err)
		}
		out = append(out, info)
	}
	return out, rows.Err()
}
