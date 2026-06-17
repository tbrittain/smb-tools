package store

import (
	"context"
	"fmt"
)

// ExportPreset is the domain model for a saved export configuration.
type ExportPreset struct {
	ID         string
	Name       string
	DatasetID  string
	ConfigJSON string
	CreatedAt  string
}

// ExportPresetStore manages saved export presets in the companion DB.
type ExportPresetStore struct {
	db DBTX
}

func NewExportPresetStore(db DBTX) *ExportPresetStore {
	return &ExportPresetStore{db: db}
}

// GetExportPresets returns all saved presets ordered by creation time descending.
func (s *ExportPresetStore) GetExportPresets(ctx context.Context) ([]ExportPreset, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, name, dataset_id, config_json, created_at
FROM export_presets
ORDER BY created_at DESC, rowid DESC`)
	if err != nil {
		return nil, fmt.Errorf("GetExportPresets: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []ExportPreset
	for rows.Next() {
		var p ExportPreset
		if err := rows.Scan(&p.ID, &p.Name, &p.DatasetID, &p.ConfigJSON, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("GetExportPresets scan: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetExportPresets rows: %w", err)
	}
	if out == nil {
		out = []ExportPreset{}
	}
	return out, nil
}

// SaveExportPreset inserts a new preset and returns the full inserted row.
func (s *ExportPresetStore) SaveExportPreset(ctx context.Context, name, datasetID, configJSON string) (ExportPreset, error) {
	var p ExportPreset
	err := s.db.QueryRowContext(ctx, `
INSERT INTO export_presets (name, dataset_id, config_json)
VALUES (?, ?, ?)
RETURNING id, name, dataset_id, config_json, created_at`,
		name, datasetID, configJSON,
	).Scan(&p.ID, &p.Name, &p.DatasetID, &p.ConfigJSON, &p.CreatedAt)
	if err != nil {
		return ExportPreset{}, fmt.Errorf("SaveExportPreset: %w", err)
	}
	return p, nil
}

// DeleteExportPreset removes the preset with the given ID. No-op if not found.
func (s *ExportPresetStore) DeleteExportPreset(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM export_presets WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("DeleteExportPreset: %w", err)
	}
	return nil
}
