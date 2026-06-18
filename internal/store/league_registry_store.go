package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// LeagueRegistryStore handles reads and writes against master.sav's
// t_league_savedatas table — the game's own registry of leagues it knows
// about for a Steam profile. It operates on an already-decompressed,
// already-opened master.sav SQLite connection (see internal/db.DecompressToTempFile),
// never on the live .sav file directly.
//
// See docs/domain/master-save-schema.md for the verified real schema this
// store is built against, and docs/league-transfer/failure-analysis.md for
// why GUID must be bound as its raw 16-byte form, never a string.
type LeagueRegistryStore struct {
	db DBTX
}

func NewLeagueRegistryStore(db DBTX) *LeagueRegistryStore {
	return &LeagueRegistryStore{db: db}
}

// LeagueExists reports whether guid already has a row in t_league_savedatas.
func (s *LeagueRegistryStore) LeagueExists(ctx context.Context, guid uuid.UUID) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM t_league_savedatas WHERE GUID = ?
	`, guid[:]).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("checking league existence for %s: %w", guid, err)
	}
	return count > 0, nil
}

// RegisterLeague inserts a new row into t_league_savedatas for guid, with
// isMissing = 0. guid is bound as its raw 16-byte slice (guid[:]), never a
// stringified form — this is the exact correction for the bug confirmed in
// docs/league-transfer/failure-analysis.md (Bug #1), where the legacy tool
// bound a 36-character uppercase string into this BLOB-affinity column.
//
// Callers must check LeagueExists first; RegisterLeague does not itself
// guard against the primary key collision (see
// docs/league-transfer/ux-flow.md's "hard stop" decision for import).
func (s *LeagueRegistryStore) RegisterLeague(ctx context.Context, guid uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO t_league_savedatas (GUID, isMissing) VALUES (?, 0)
	`, guid[:])
	if err != nil {
		return fmt.Errorf("registering league %s: %w", guid, err)
	}
	return nil
}
