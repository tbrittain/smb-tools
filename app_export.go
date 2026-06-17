package main

import (
	"encoding/base64"
	"fmt"
	"log/slog"

	"smb-tools/internal/store"
)

// GetTeamsForExport returns all teams for use in the export page team filter.
func (a *App) GetTeamsForExport() ([]TeamPickerResultDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	teams, err := a.teamQueryStore.ListAllTeams(a.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]TeamPickerResultDTO, len(teams))
	for i, t := range teams {
		out[i] = TeamPickerResultDTO{
			TeamID:         t.TeamID,
			TeamName:       t.TeamName,
			ConferenceName: t.ConferenceName,
			DivisionName:   t.DivisionName,
		}
	}
	return out, nil
}

// PreviewExportData returns up to 500 rows for the given export configuration
// plus the total matching row count so the frontend can show a truncation notice.
func (a *App) PreviewExportData(opts ExportOptionsDTO) (ExportPreviewDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return ExportPreviewDTO{}, err
	}
	filters := make([]store.FilterRow, len(opts.Filters))
	for i, f := range opts.Filters {
		filters[i] = store.FilterRow{Column: f.Column, Op: f.Op, Value: f.Value, Value2: f.Value2}
	}
	preview, err := a.exportStore.PreviewExportData(a.ctx, store.ExportOptions{
		DatasetID:      opts.DatasetID,
		Columns:        opts.Columns,
		Filters:        filters,
		SortCol:        opts.SortCol,
		SortDir:        opts.SortDir,
		CareerStatType: opts.CareerStatType,
	})
	if err != nil {
		slog.Error("PreviewExportData", "dataset", opts.DatasetID, "err", err)
		return ExportPreviewDTO{}, fmt.Errorf("preview export: %w", err)
	}
	rows := preview.Rows
	if rows == nil {
		rows = []map[string]any{}
	}
	return ExportPreviewDTO{Rows: rows, TotalCount: preview.TotalCount}, nil
}

// ExportToCSV runs the full export query and returns the CSV as a base64-encoded
// string. The frontend decodes it and triggers a file download.
func (a *App) ExportToCSV(opts ExportOptionsDTO) (string, error) {
	if err := a.requireCompanionDB(); err != nil {
		return "", err
	}
	filters := make([]store.FilterRow, len(opts.Filters))
	for i, f := range opts.Filters {
		filters[i] = store.FilterRow{Column: f.Column, Op: f.Op, Value: f.Value, Value2: f.Value2}
	}
	data, err := a.exportStore.ExportToCSV(a.ctx, store.ExportOptions{
		DatasetID:      opts.DatasetID,
		Columns:        opts.Columns,
		Filters:        filters,
		SortCol:        opts.SortCol,
		SortDir:        opts.SortDir,
		CareerStatType: opts.CareerStatType,
	})
	if err != nil {
		slog.Error("ExportToCSV", "dataset", opts.DatasetID, "err", err)
		return "", fmt.Errorf("export CSV: %w", err)
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// GetExportPresets returns all saved export presets for the active franchise.
func (a *App) GetExportPresets() ([]ExportPresetDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	presets, err := a.exportPresetStore.GetExportPresets(a.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ExportPresetDTO, len(presets))
	for i, p := range presets {
		out[i] = exportPresetToDTO(p)
	}
	return out, nil
}

// SaveExportPreset saves a named export configuration for the active franchise.
func (a *App) SaveExportPreset(name, datasetID, configJSON string) (ExportPresetDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return ExportPresetDTO{}, err
	}
	p, err := a.exportPresetStore.SaveExportPreset(a.ctx, name, datasetID, configJSON)
	if err != nil {
		return ExportPresetDTO{}, err
	}
	return exportPresetToDTO(p), nil
}

// DeleteExportPreset removes a saved export preset by ID. No-op if not found.
func (a *App) DeleteExportPreset(id string) error {
	if err := a.requireCompanionDB(); err != nil {
		return err
	}
	return a.exportPresetStore.DeleteExportPreset(a.ctx, id)
}

func exportPresetToDTO(p store.ExportPreset) ExportPresetDTO {
	return ExportPresetDTO{
		ID:         p.ID,
		Name:       p.Name,
		DatasetID:  p.DatasetID,
		ConfigJSON: p.ConfigJSON,
		CreatedAt:  p.CreatedAt,
	}
}
