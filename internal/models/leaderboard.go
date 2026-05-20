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
	CareerPitchingStats
}
