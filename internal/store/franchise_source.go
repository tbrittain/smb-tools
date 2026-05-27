package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"smb-tools/internal/models"
)

// FranchiseSourceStore handles reads and writes for franchise_sources in the
// registry DB.
type FranchiseSourceStore struct {
	db DBTX
}

func NewFranchiseSourceStore(db DBTX) *FranchiseSourceStore {
	return &FranchiseSourceStore{db: db}
}

// Add inserts a new franchise source row and returns it with the generated ID.
func (s *FranchiseSourceStore) Add(
	ctx context.Context,
	franchiseID, saveFilePath, leagueGUID string,
	seasonOffset int,
) (models.FranchiseSource, error) {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO franchise_sources (franchise_id, save_file_path, league_guid, season_offset)
		VALUES (?, ?, ?, ?)
	`, franchiseID, saveFilePath, leagueGUID, seasonOffset)
	if err != nil {
		return models.FranchiseSource{}, fmt.Errorf("adding source for franchise %q: %w", franchiseID, err)
	}
	id, _ := res.LastInsertId()
	return models.FranchiseSource{
		ID:           id,
		FranchiseID:  franchiseID,
		SaveFilePath: saveFilePath,
		LeagueGUID:   leagueGUID,
		SeasonOffset: seasonOffset,
		AddedAt:      time.Now().UTC(),
	}, nil
}

// ListByFranchise returns all sources for a franchise ordered by season_offset
// ascending (oldest/lowest-offset source first).
func (s *FranchiseSourceStore) ListByFranchise(ctx context.Context, franchiseID string) ([]models.FranchiseSource, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, franchise_id, save_file_path, league_guid, season_offset, added_at
		FROM franchise_sources
		WHERE franchise_id = ?
		ORDER BY season_offset ASC, added_at ASC
	`, franchiseID)
	if err != nil {
		return nil, fmt.Errorf("listing sources for franchise %q: %w", franchiseID, err)
	}
	defer func() { _ = rows.Close() }()
	return scanSources(rows)
}

// GetActive returns the active source for a franchise — the one with the
// highest season_offset (ties broken by added_at DESC). Returns sql.ErrNoRows
// if the franchise has no sources configured.
func (s *FranchiseSourceStore) GetActive(ctx context.Context, franchiseID string) (models.FranchiseSource, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, franchise_id, save_file_path, league_guid, season_offset, added_at
		FROM franchise_sources
		WHERE franchise_id = ?
		ORDER BY season_offset DESC, added_at DESC
		LIMIT 1
	`, franchiseID)
	if err != nil {
		return models.FranchiseSource{}, fmt.Errorf("getting active source for franchise %q: %w", franchiseID, err)
	}
	defer func() { _ = rows.Close() }()

	sources, err := scanSources(rows)
	if err != nil {
		return models.FranchiseSource{}, err
	}
	if len(sources) == 0 {
		return models.FranchiseSource{}, sql.ErrNoRows
	}
	return sources[0], nil
}

// ListAll returns all sources for all franchises, ordered by franchise_id then
// season_offset. Used to batch-load sources when building franchise list DTOs.
func (s *FranchiseSourceStore) ListAll(ctx context.Context) ([]models.FranchiseSource, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, franchise_id, save_file_path, league_guid, season_offset, added_at
		FROM franchise_sources
		ORDER BY franchise_id, season_offset ASC, added_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("listing all sources: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanSources(rows)
}

// GetByLeagueGUID returns the source for a franchise that matches the given
// leagueGUID. Returns sql.ErrNoRows if no such source exists.
func (s *FranchiseSourceStore) GetByLeagueGUID(ctx context.Context, franchiseID, leagueGUID string) (models.FranchiseSource, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, franchise_id, save_file_path, league_guid, season_offset, added_at
		FROM franchise_sources
		WHERE franchise_id = ? AND league_guid = ?
		LIMIT 1
	`, franchiseID, leagueGUID)
	if err != nil {
		return models.FranchiseSource{}, fmt.Errorf("getting source by league GUID for franchise %q: %w", franchiseID, err)
	}
	defer func() { _ = rows.Close() }()

	sources, err := scanSources(rows)
	if err != nil {
		return models.FranchiseSource{}, err
	}
	if len(sources) == 0 {
		return models.FranchiseSource{}, sql.ErrNoRows
	}
	return sources[0], nil
}

// Replace updates the save_file_path and league_guid of an existing source
// in-place. Used for path corrections (e.g. save file moved) — does not change
// season_offset or create a new source row.
func (s *FranchiseSourceStore) Replace(ctx context.Context, sourceID int64, saveFilePath, leagueGUID string) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE franchise_sources
		SET save_file_path = ?, league_guid = ?
		WHERE id = ?
	`, saveFilePath, leagueGUID, sourceID)
	if err != nil {
		return fmt.Errorf("replacing source %d: %w", sourceID, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteByFranchise removes all sources for a franchise. Called when the
// franchise itself is deleted.
func (s *FranchiseSourceStore) DeleteByFranchise(ctx context.Context, franchiseID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM franchise_sources WHERE franchise_id = ?`, franchiseID)
	if err != nil {
		return fmt.Errorf("deleting sources for franchise %q: %w", franchiseID, err)
	}
	return nil
}

// ---- helpers ---------------------------------------------------------------

func scanSources(rows *sql.Rows) ([]models.FranchiseSource, error) {
	var out []models.FranchiseSource
	for rows.Next() {
		var src models.FranchiseSource
		var addedAt string
		if err := rows.Scan(
			&src.ID, &src.FranchiseID, &src.SaveFilePath,
			&src.LeagueGUID, &src.SeasonOffset, &addedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning franchise source: %w", err)
		}
		t, _ := time.Parse("2006-01-02T15:04:05Z", addedAt)
		src.AddedAt = t
		out = append(out, src)
	}
	return out, rows.Err()
}
