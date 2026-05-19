package store

import "database/sql"

// FranchiseStore handles reads and writes for the franchise registry.
// It operates against registry.db, not a per-franchise companion DB.
type FranchiseStore struct {
	db *sql.DB
}

func NewFranchiseStore(db *sql.DB) *FranchiseStore {
	return &FranchiseStore{db: db}
}

// CRUD methods are added in Phase 3 alongside the franchise management UI.
