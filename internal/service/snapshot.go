package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"smb-tools/internal/store"
)

// SnapshotService handles save game snapshot persistence for a single franchise.
// It is responsible for decompressing save files, computing hashes for
// deduplication, writing snapshot files to disk, and recording metadata.
type SnapshotService struct {
	snapshotDir   string // absolute path to franchises/{id}/snapshots/
	snapshotStore *store.SnapshotStore
}

func NewSnapshotService(snapshotDir string, snapshotStore *store.SnapshotStore) *SnapshotService {
	return &SnapshotService{
		snapshotDir:   snapshotDir,
		snapshotStore: snapshotStore,
	}
}

// TakeSnapshot checks whether the decompressed save game bytes differ from the
// most recent snapshot (by SHA-256 hash). If they do, it persists a new
// snapshot file and records metadata. Returns the snapshot ID and whether a
// new snapshot was written (false = identical to the last one).
func (s *SnapshotService) TakeSnapshot(ctx context.Context, decompressedBytes []byte, seasonNum int) (id int64, isNew bool, err error) {
	hash := sha256Sum(decompressedBytes)

	latest, err := s.snapshotStore.LatestHash(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("checking latest snapshot hash: %w", err)
	}
	if latest == hash { //nolint:gocritic // comparing same named types
		return 0, false, nil // identical to last snapshot — skip
	}

	if err := os.MkdirAll(s.snapshotDir, 0o700); err != nil {
		return 0, false, fmt.Errorf("creating snapshots directory: %w", err)
	}

	// Filename: {season_num}_{first 12 chars of hash}.sqlite
	shortHash := string(hash)[:12]
	fileName := store.SnapshotFileName(fmt.Sprintf("%04d_%s.sqlite", seasonNum, shortHash))
	fullPath := filepath.Join(s.snapshotDir, string(fileName))

	if err := os.WriteFile(fullPath, decompressedBytes, 0o600); err != nil {
		return 0, false, fmt.Errorf("writing snapshot file: %w", err)
	}

	snap := store.Snapshot{
		SeasonNum:     seasonNum,
		CapturedAt:    time.Now().UTC(),
		FileName:      fileName,
		SHA256Hash:    hash,
		FileSizeBytes: int64(len(decompressedBytes)),
	}
	snapshotID, err := s.snapshotStore.Record(ctx, snap)
	if err != nil {
		// Best-effort cleanup of the file we just wrote
		_ = os.Remove(fullPath)
		return 0, false, fmt.Errorf("recording snapshot metadata: %w", err)
	}
	return snapshotID, true, nil
}

// TakeSnapshotFromFile reads a decompressed SQLite file from disk and runs
// TakeSnapshot on its contents. Use when the save has already been
// decompressed to a temp file.
func (s *SnapshotService) TakeSnapshotFromFile(ctx context.Context, srcPath string, seasonNum int) (id int64, isNew bool, err error) {
	f, err := os.Open(srcPath)
	if err != nil {
		return 0, false, fmt.Errorf("opening source file: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return 0, false, fmt.Errorf("reading source file: %w", err)
	}
	return s.TakeSnapshot(ctx, data, seasonNum)
}

// sha256Sum returns the SHA256Hex of data.
func sha256Sum(data []byte) store.SHA256Hex {
	h := sha256.Sum256(data)
	return store.SHA256Hex(hex.EncodeToString(h[:]))
}
