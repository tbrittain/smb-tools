package models

import "time"

// Franchise represents an SMB franchise tracked in the registry.
// Registry-level data only — no baseball stats live here.
type Franchise struct {
	ID               string
	Name             string
	GameVersion      GameVersion
	LeagueMode       LeagueMode
	CreatedAt        time.Time
	LastSyncedAt     *time.Time
	LastSyncedSeason *int
}

// FranchiseSource represents one save game file + leagueGUID pair associated
// with a franchise. A franchise starts with one source (SeasonOffset = 0).
// When the user forks to a new league, a second source is added with
// SeasonOffset equal to the last synced season number.
type FranchiseSource struct {
	ID           int64
	FranchiseID  string
	SaveFilePath string
	LeagueGUID   string
	SeasonOffset int
	AddedAt      time.Time
}
