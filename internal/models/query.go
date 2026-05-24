package models

import "time"

// ── Seasons ───────────────────────────────────────────────────────────────────

// SeasonSummary is returned by SeasonQueryStore.ListWithChampion.
type SeasonSummary struct {
	ID                int64
	SeasonNum         int
	NumGames          int
	ImportedAt        time.Time
	ChampionTeamName  string // empty when no playoff data exists
	ChampionHistoryID *int64
}

// ── Standings ─────────────────────────────────────────────────────────────────

// TeamStandingRow is one row in a season standings table.
type TeamStandingRow struct {
	HistoryID      int64
	TeamID         int64
	TeamName       string
	DivisionName   string
	ConferenceName string
	Wins           int
	Losses         int
	WinPct         float64
	GamesBack      float64
	RunsFor        int
	RunsAgainst    int
	RunDiff        int
	PlayoffSeed    *int
}

// ── Stat leaders ──────────────────────────────────────────────────────────────

// StatLeader is the leader for one stat category in a season.
type StatLeader struct {
	PlayerID  int64
	FirstName string
	LastName  string
	TeamName  string
	StatValue float64
}

// StatLeaders groups the six title-leader categories for one season.
type StatLeaders struct {
	SeasonNum  int
	BA         *StatLeader
	HR         *StatLeader
	RBI        *StatLeader
	ERA        *StatLeader
	Wins       *StatLeader
	Strikeouts *StatLeader
}

// CareerLeaderRow is one entry in an all-time career leaderboard category.
type CareerLeaderRow struct {
	PlayerID      int64
	FirstName     string
	LastName      string
	StatValue     float64
	SeasonsPlayed int
}

// CareerLeaders groups the top-5 players for each all-time career category.
type CareerLeaders struct {
	HR         []CareerLeaderRow
	Hits       []CareerLeaderRow
	RBI        []CareerLeaderRow
	Wins       []CareerLeaderRow
	Strikeouts []CareerLeaderRow
	Saves      []CareerLeaderRow
}

// ── Player stats ──────────────────────────────────────────────────────────────

// CareerBattingStats holds summed batting counting stats with computed rates.
// Rate fields are nil when the denominator is zero (e.g. zero AB).
type CareerBattingStats struct {
	GamesPlayed    int
	GamesBatting   int
	AtBats         int
	Runs           int
	Hits           int
	Doubles        int
	Triples        int
	HomeRuns       int
	RBI            int
	StolenBases    int
	CaughtStealing int
	Walks          int
	Strikeouts     int
	HitByPitch     int
	SacHits        int
	SacFlies       int
	Errors         int
	PassedBalls    int

	// Computed by service.ComputeBattingRates
	BA      *float64
	OBP     *float64
	SLG     *float64
	OPS     *float64
	ISO     *float64
	BABIP   *float64
	KPct    *float64
	BBPct   *float64
	ABPerHR *float64

	// Context-dependent stats stored at sync time (nil for pre-Phase-8.5 seasons)
	OPSPlus *float64 // 100 × (OBP/lgOBP + SLG/lgSLG − 1); higher = better
	SmbWAR  *float64 // legacy-formula smbWAR: PA-weighted OPS+ + baserunning
}

// CareerPitchingStats holds summed pitching counting stats with computed rates.
// Rate fields are nil when the denominator is zero (e.g. zero outs pitched).
type CareerPitchingStats struct {
	Wins            int
	Losses          int
	Games           int
	GamesStarted    int
	CompleteGames   int
	Shutouts        int
	Saves           int
	OutsPitched     int
	HitsAllowed     int
	EarnedRuns      int
	HomeRunsAllowed int
	Walks           int
	Strikeouts      int
	HitBatters      int
	BattersFaced    int
	GamesFinished   int
	RunsAllowed     int
	WildPitches     int
	TotalPitches    int

	// Computed by service.ComputePitchingRates
	ERA     *float64
	WHIP    *float64
	K9      *float64
	BB9     *float64
	H9      *float64
	HR9     *float64
	KPerBB  *float64
	KPct    *float64
	WinPct  *float64
	PPerIP  *float64

	// Context-dependent stats stored at sync time (nil for pre-Phase-8.5 seasons)
	ERAPlus  *float64 // 100 × lgERA / ERA; higher = better
	FIP      *float64 // Fielding Independent Pitching (league-constant adjusted)
	FIPMinus *float64 // 100 × FIP / lgERA; lower = better
	SmbWAR   *float64 // legacy-formula smbWAR: IP-weighted avg of ERA+ and FIP+
}

// ── Player ────────────────────────────────────────────────────────────────────

// PlayerSearchResult is returned by PlayerQueryStore.SearchPlayers.
type PlayerSearchResult struct {
	PlayerID      int64
	FirstName     string
	LastName      string
	IsHallOfFamer bool
	SeasonsPlayed int
	FirstSeason   int
	LastSeason    int
}

// PlayerCareer is the full career record for one player.
type PlayerCareer struct {
	PlayerID      int64
	FirstName     string
	LastName      string
	IsHallOfFamer bool
	Batting       *CareerBattingStats  // nil if player has no batting rows
	Pitching      *CareerPitchingStats // nil if player has no pitching rows
}

// PlayerSeasonLogRow is one row in a player's season-by-season breakdown.
type PlayerSeasonLogRow struct {
	SeasonNum         int
	SeasonID          int64
	TeamHistoryID     *int64 // nil when player is a free agent
	TeamID            *int64 // nil when player is a free agent
	TeamName          string
	Age               int
	Salary            int
	PrimaryPosition   string
	SecondaryPosition string
	PitcherRole       string
	BatHand           string
	ThrowHand         string
	ChemistryType     string
	TraitsJSON        string
	PitchesJSON       string

	// Attributes (0 when not recorded)
	Power    int
	Contact  int
	Speed    int
	Fielding int
	Arm      int
	Velocity int
	Junk     int
	Accuracy int

	// Regular season stats (nil when no rows exist for this player-season)
	Batting  *CareerBattingStats
	Pitching *CareerPitchingStats

	// Playoff stats
	PlayoffBatting  *CareerBattingStats
	PlayoffPitching *CareerPitchingStats
}

// ── Team ──────────────────────────────────────────────────────────────────────

// TeamSearchResult is returned by TeamQueryStore.SearchTeams.
type TeamSearchResult struct {
	TeamID      int64
	TeamName    string
	Seasons     int
	FirstSeason int
	LastSeason  int
}

// TeamSeasonSummary is one season entry in a team's history.
type TeamSeasonSummary struct {
	HistoryID      int64
	SeasonID       int64
	SeasonNum      int
	TeamName       string
	DivisionName   string
	ConferenceName string
	Wins           int
	Losses         int
	WinPct         float64
	GamesBack      float64
	RunsFor        int
	RunsAgainst    int
	Budget         int
	Payroll        int
	PlayoffSeed    *int
	PlayoffWins    *int
	PlayoffLosses  *int
	PlayoffRunsFor    *int
	PlayoffRunsAgainst *int
	TotalPower    int
	TotalContact  int
	TotalSpeed    int
	TotalFielding int
	TotalArm      int
	TotalVelocity int
	TotalJunk     int
	TotalAccuracy int
	IsChampion    bool
}

// TeamHistory is the full all-time record for one team.
type TeamHistory struct {
	TeamID   int64
	GameGUID string
	Seasons  []TeamSeasonSummary
}

// TeamSeasonListRow is one row in the historical teams list page.
type TeamSeasonListRow struct {
	SeasonNum      int
	HistoryID      int64
	TeamID         int64
	TeamName       string
	ConferenceName string
	DivisionName   string
	Wins           int
	Losses         int
	WinPct         float64
	RunsFor        int
	RunsAgainst    int
	PlayoffSeed    *int
	PlayoffWins    *int
	PlayoffLosses  *int
	IsChampion     bool
}

// HistoricalTeamRow is one row in the aggregated historical teams page.
// Each row represents one team's totals across a user-selected season range.
type HistoricalTeamRow struct {
	TeamID              int64
	TeamName            string
	NumSeasons          int
	FirstSeason         int
	LastSeason          int
	Wins                int
	Losses              int
	PlayoffWins         int
	PlayoffLosses       int
	PlayoffAppearances  int
	DivisionTitles      int
	ConferenceTitles    int
	Championships       int
	ChampionshipDrought int
	RunsFor             int
	RunsAgainst         int
	TotalAB             int
	TotalHits           int
	TotalHR             int
	NumPlayers          int
	NumHoF              int
	TotalEarnedRuns     int
	TotalOutsPitched    int

	// Computed in Go after scanning
	WinPct       float64
	GamesOver500 int
	BA           *float64
	ERA          *float64
}

// ── Roster & schedule ─────────────────────────────────────────────────────────

// RosterPlayer is one player row in a team season roster.
type RosterPlayer struct {
	PlayerID          int64
	FirstName         string
	LastName          string
	IsHallOfFamer     bool
	Age               int
	Salary            int
	PrimaryPosition   string
	SecondaryPosition string
	PitcherRole       string
	BatHand           string
	ThrowHand         string
	ChemistryType     string
	TraitsJSON        string
	PitchesJSON       string

	Power    int
	Contact  int
	Speed    int
	Fielding int
	Arm      int
	Velocity int
	Junk     int
	Accuracy int

	Batting  *CareerBattingStats
	Pitching *CareerPitchingStats
}

// ScheduleGameRow is one game in a team's regular season schedule.
type ScheduleGameRow struct {
	GameNumber        int
	Day               int
	HomeTeamHistoryID int64
	HomeTeamName      string
	AwayTeamHistoryID int64
	AwayTeamName      string
	HomeScore         *int
	AwayScore         *int
	HomePitcherName   string
	AwayPitcherName   string
}

// PlayoffGameRow is one game in a team's playoff schedule.
type PlayoffGameRow struct {
	SeriesNumber      int
	GameNumber        int
	HomeTeamHistoryID int64
	HomeTeamName      string
	AwayTeamHistoryID int64
	AwayTeamName      string
	HomeScore         *int
	AwayScore         *int
	HomePitcherName   string
	AwayPitcherName   string
}

// TeamSeasonDetail bundles roster, schedule, and playoff games for one team season.
type TeamSeasonDetail struct {
	Team     TeamSeasonSummary
	Roster   []RosterPlayer
	Schedule []ScheduleGameRow
	Playoffs []PlayoffGameRow
}
