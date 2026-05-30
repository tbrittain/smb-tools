package models

// LeaderboardFilters carries optional filter parameters for all leaderboard queries.
// Zero values mean "no filter applied": empty string = any, 0 = no bound.
type LeaderboardFilters struct {
	IsPlayoffs       bool
	OnlyHallOfFamers bool
	Position         string // primary_position for batting; pitcher_role for pitching
	BatHand          string // "L", "R", "S" — batting only
	ThrowHand        string // "L", "R" — pitching only
	ChemistryType    string
	SeasonStart      int // season_num >= SeasonStart (0 = no lower bound)
	SeasonEnd        int // season_num <= SeasonEnd (0 = no upper bound)
	// Traits filters the season leaderboards to players who have ALL listed traits
	// in their traits_json for that season (AND logic). Max 2 entries; SMB4 only.
	Traits []string
	// Server-side pagination/sort — used by season leader queries only.
	// SortField is the frontend camelCase field name (e.g. "ba", "homeRuns", "smbWar").
	// Empty SortField defaults to smbWAR DESC. Offset/PageSize default to 0/50.
	SortField string
	SortDesc  bool
	Offset    int // 0-based row offset (= DataTable "first")
	PageSize  int // 0 → use default (50)
}

// BattingCareerLeaderRow is one player's aggregated career batting stats for
// the leaderboard. Embeds CareerBattingStats so ComputeBattingRates can be
// called on &rows[i].CareerBattingStats without copying.
type BattingCareerLeaderRow struct {
	PlayerID      int64
	FirstName     string
	LastName      string
	IsHallOfFamer bool
	SeasonsPlayed int
	CareerBattingStats
}

// BattingSeasonLeaderRow is one player-season batting stat row for the leaderboard.
type BattingSeasonLeaderRow struct {
	PlayerID        int64
	FirstName       string
	LastName        string
	IsHallOfFamer   bool
	SeasonNum       int
	TeamName        string
	Age             int
	PrimaryPosition string
	BatHand         string
	ChemistryType   string
	Traits          []string
	CareerBattingStats
}

// PitchingCareerLeaderRow is one player's aggregated career pitching stats for
// the leaderboard. Embeds CareerPitchingStats so ComputePitchingRates can be
// called on &rows[i].CareerPitchingStats without copying.
type PitchingCareerLeaderRow struct {
	PlayerID      int64
	FirstName     string
	LastName      string
	IsHallOfFamer bool
	SeasonsPlayed int
	CareerPitchingStats
}

// PitchingSeasonLeaderRow is one player-season pitching stat row for the leaderboard.
type PitchingSeasonLeaderRow struct {
	PlayerID      int64
	FirstName     string
	LastName      string
	IsHallOfFamer bool
	SeasonNum     int
	TeamName      string
	Age           int
	PitcherRole   string
	ThrowHand     string
	ChemistryType string
	Traits        []string
	CareerPitchingStats
}
