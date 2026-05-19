package models

import "time"

// Franchise represents an SMB franchise tracked in the registry.
// Registry-level data only — no baseball stats live here.
type Franchise struct {
	ID                string
	Name              string
	GameVersion       GameVersion
	SaveFilePath      string
	LeagueGUID        string
	DBPath            string
	CreatedAt         time.Time
	LastSyncedAt      *time.Time
	LastSyncedSeason  *int
}
