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

func NewSyncService(snapshot *SnapshotService, importer *ImportService) *SyncService {
	return &SyncService{snapshot: snapshot, importer: importer}
}

// SyncResult summarises the outcome of a full season sync.
type SyncResult struct {
	SeasonID      int64
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
// into the companion DB. leagueGUID and seasonOffset come from the active
// franchise_source row. The snapshot is taken before the import — if the
// snapshot fails the import does not run.
func (s *SyncService) SyncSeason(
	ctx context.Context,
	companionDB *sql.DB,
	reader store.SaveGameReader,
	saveFilePath string,
	leagueGUID string,
	seasonOffset int,
) (SyncResult, error) {
	seasonInfo, err := reader.GetCurrentSeason(ctx, leagueGUID)
	if err != nil {
		return SyncResult{}, fmt.Errorf("detecting current season: %w", err)
	}

	displaySeasonNum := seasonInfo.SeasonNum + seasonOffset

	snapshotID, isNew, err := s.snapshot.TakeSnapshotFromFile(ctx, saveFilePath, displaySeasonNum)
	if err != nil {
		return SyncResult{}, fmt.Errorf("taking snapshot for season %d: %w", displaySeasonNum, err)
	}

	importResult, err := s.importer.ImportSeason(
		ctx, companionDB, reader,
		seasonInfo.SeasonID, seasonInfo.SeasonNum,
		leagueGUID, seasonOffset,
	)
	if err != nil {
		return SyncResult{}, fmt.Errorf("importing season %d: %w", displaySeasonNum, err)
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
