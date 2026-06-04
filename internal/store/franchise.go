package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"smb-tools/internal/models"
)

// FranchiseStore handles reads and writes for the franchise registry.
// It operates against registry.db, not a per-franchise companion DB.
type FranchiseStore struct {
	db DBTX
}

func NewFranchiseStore(db DBTX) *FranchiseStore {
	return &FranchiseStore{db: db}
}

// Create inserts a new franchise record. id must be a unique string (use UUID).
func (s *FranchiseStore) Create(ctx context.Context, f models.Franchise) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO franchises (id, name, game_version)
		VALUES (?, ?, ?)
	`, f.ID, f.Name, f.GameVersion)
	if err != nil {
		return fmt.Errorf("creating franchise %q: %w", f.Name, err)
	}
	return nil
}

// List returns all franchises ordered by creation time ascending.
func (s *FranchiseStore) List(ctx context.Context) ([]models.Franchise, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, game_version, created_at,
		       last_synced_at, last_synced_season
		FROM franchises
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("listing franchises: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanFranchises(rows)
}

// GetByID returns the franchise with the given ID, or sql.ErrNoRows if not found.
func (s *FranchiseStore) GetByID(ctx context.Context, id string) (models.Franchise, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, game_version, created_at,
		       last_synced_at, last_synced_season
		FROM franchises
		WHERE id = ?
	`, id)
	if err != nil {
		return models.Franchise{}, fmt.Errorf("getting franchise %q: %w", id, err)
	}
	defer func() { _ = rows.Close() }()

	fs, err := scanFranchises(rows)
	if err != nil {
		return models.Franchise{}, err
	}
	if len(fs) == 0 {
		return models.Franchise{}, sql.ErrNoRows
	}
	return fs[0], nil
}

// Rename updates the display name of a franchise.
func (s *FranchiseStore) Rename(ctx context.Context, id, newName string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE franchises SET name = ? WHERE id = ?`, newName, id)
	if err != nil {
		return fmt.Errorf("renaming franchise %q: %w", id, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// RecordSync updates the last_synced_at timestamp and last_synced_season.
func (s *FranchiseStore) RecordSync(ctx context.Context, id string, seasonNum int) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE franchises
		SET last_synced_at = datetime('now'), last_synced_season = ?
		WHERE id = ?
	`, seasonNum, id)
	if err != nil {
		return fmt.Errorf("recording sync for franchise %q: %w", id, err)
	}
	return nil
}

// Delete removes a franchise from the registry. It does NOT delete the
// companion DB file or snapshots on disk — that is the caller's responsibility.
func (s *FranchiseStore) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM franchises WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting franchise %q: %w", id, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ---- helpers ---------------------------------------------------------------

func scanFranchises(rows *sql.Rows) ([]models.Franchise, error) {
	var fs []models.Franchise
	for rows.Next() {
		var f models.Franchise
		var createdAt string
		var lastSyncedAt sql.NullString
		var lastSyncedSeason sql.NullInt64
		var gameVersion string
		if err := rows.Scan(
			&f.ID, &f.Name, &gameVersion,
			&createdAt, &lastSyncedAt, &lastSyncedSeason,
		); err != nil {
			return nil, fmt.Errorf("scanning franchise: %w", err)
		}
		f.GameVersion = models.GameVersion(gameVersion)
		t, _ := time.Parse("2006-01-02T15:04:05Z", createdAt)
		f.CreatedAt = t
		if lastSyncedAt.Valid {
			st, _ := time.Parse("2006-01-02T15:04:05Z", lastSyncedAt.String)
			f.LastSyncedAt = &st
		}
		if lastSyncedSeason.Valid {
			n := int(lastSyncedSeason.Int64)
			f.LastSyncedSeason = &n
		}
		fs = append(fs, f)
	}
	return fs, rows.Err()
}
