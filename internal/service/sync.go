package service

import (
	"context"
	"database/sql"
	"fmt"

	"smb-tools/internal/store"
)

// SyncService orchestrates a full season sync: snapshot capture followed by
// data import. It is created per-franchise because SnapshotService holds the
// franchise-specific snapshot directory and store.
//
// Snapshot failure is fatal. Every successful sync must have a corresponding
// snapshot on disk. If the snapshot cannot be written the import is blocked
// so the save game state is never silently lost.
type SyncService struct {
	snapshot *SnapshotService
	importer *ImportService
}

// NewSyncService wires a franchise-scoped SnapshotService together with the
// shared ImportService.
func NewSyncService(snapshot *SnapshotService, importer *ImportService) *SyncService {
	return &SyncService{snapshot: snapshot, importer: importer}
}

// SyncResult summarises the outcome of a full season sync.
type SyncResult struct {
	SeasonID      int
	SeasonNum     int
	Players       int
	Teams         int
	Games         int
	PlayoffGames  int
	SnapshotID    int64
	IsNewSnapshot bool
}

// SyncSeason detects the current season from the reader, captures a snapshot
// of the decompressed save game at saveFilePath, then imports the season data
// into the companion DB. The snapshot is taken before the import — if the
// snapshot fails the import does not run and an error is returned.
func (s *SyncService) SyncSeason(
	ctx context.Context,
	companionDB *sql.DB,
	reader store.SaveGameReader,
	saveFilePath string,
	leagueGUID string,
) (SyncResult, error) {
	seasonInfo, err := reader.GetCurrentSeason(ctx, leagueGUID)
	if err != nil {
		return SyncResult{}, fmt.Errorf("detecting current season: %w", err)
	}

	snapshotID, isNew, err := s.snapshot.TakeSnapshotFromFile(ctx, saveFilePath, seasonInfo.SeasonNum)
	if err != nil {
		return SyncResult{}, fmt.Errorf("taking snapshot for season %d: %w", seasonInfo.SeasonNum, err)
	}

	importResult, err := s.importer.ImportSeason(ctx, companionDB, reader, seasonInfo.SeasonID, seasonInfo.SeasonNum)
	if err != nil {
		return SyncResult{}, fmt.Errorf("importing season %d: %w", seasonInfo.SeasonNum, err)
	}

	return SyncResult{
		SeasonID:      importResult.SeasonID,
		SeasonNum:     importResult.SeasonNum,
		Players:       importResult.Players,
		Teams:         importResult.Teams,
		Games:         importResult.Games,
		PlayoffGames:  importResult.PlayoffGames,
		SnapshotID:    snapshotID,
		IsNewSnapshot: isNew,
	}, nil
}
