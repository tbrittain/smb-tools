package models

// SaveGameLeague represents a league entry from t_leagues in the SMB save game.
type SaveGameLeague struct {
	ID          int
	Name        string
	TeamTypeID  int
	TypeName    string
	FranchiseID *int
	NumSeasons  int
	Elimination bool
}

// SaveGameFranchiseSeason represents one season in a franchise (t_franchise_seasons).
type SaveGameFranchiseSeason struct {
	SeasonID  int
	SeasonNum int // computed rank within the franchise
	LeagueGUID string
}

// SaveGamePlayer is a player snapshot from the most recent season, combining
// attributes, traits, salary, and position data.
type SaveGamePlayer struct {
	PlayerGUID    string
	SeasonID      int
	SeasonNum     int
	FirstName     string
	LastName      string
	PrimaryPos    string
	SecondaryPos  string
	PitcherRole   string
	CurrentTeam   string
	PreviousTeam  string
	Power         int
	Contact       int
	Speed         int
	Fielding      int
	Arm           int
	Velocity      int
	Junk          int
	Accuracy      int
	Age           int
	Salary        int // display units (game units × 200)
	Traits        []string
	// SMB4-only fields
	ChemistryType string
	ThrowHand     string
	BatHand       string
	Pitches       []string
}

// SaveGameTeam is a team snapshot from the most recent season.
type SaveGameTeam struct {
	TeamLocalID     int
	TeamGUID        string
	TeamName        string
	SeasonID        int
	SeasonNum       int
	DivisionName    string
	ConferenceName  string
	Budget          int
	Payroll         int
	Surplus         int
	SurplusPerGame  float64
	Wins            int
	Losses          int
	GamesBack       float64
	WinPct          float64
	RunDifferential int
	RunsFor         int
	RunsAgainst     int
	TotalPower      int
	TotalContact    int
	TotalSpeed      int
	TotalFielding   int
	TotalArm        int
	TotalVelocity   int
	TotalJunk       int
	TotalAccuracy   int
}

// SaveGameGame is one scheduled game from t_season_schedule + t_game_results.
type SaveGameGame struct {
	SeasonID       int
	SeasonNum      int
	GameNumber     int
	Day            int
	HomeTeamID     int
	HomeTeamGUID   string
	HomeTeamName   string
	AwayTeamID     int
	AwayTeamGUID   string
	AwayTeamName   string
	HomeScore      *int
	AwayScore      *int
	HomePitcherID  *int
	HomePitcherGUID *string
	HomePitcherName *string
	AwayPitcherID  *int
	AwayPitcherGUID *string
	AwayPitcherName *string
}

// SaveGamePlayoffGame is one playoff game with series context.
type SaveGamePlayoffGame struct {
	SeasonID        int
	SeasonNum       int
	SeriesNum       int
	Team1GUID       string
	Team1Name       string
	Team1Seed       int
	Team2GUID       string
	Team2Name       string
	Team2Seed       int
	GameNumber      int
	HomeTeamGUID    string
	HomeTeamName    string
	AwayTeamGUID    string
	AwayTeamName    string
	HomeScore       *int
	AwayScore       *int
	HomePitcherGUID *string
	HomePitcherName *string
	AwayPitcherGUID *string
	AwayPitcherName *string
}

// SaveGameBattingStat is a season batting stat row from t_stats_batting.
type SaveGameBattingStat struct {
	AggregatorID int
	PlayerGUID   string
	FirstName    string
	LastName     string
	CurrentTeam  string
	PrevTeam     string
	Prev2Team    string
	PrimaryPos   string
	SecondaryPos string
	PitcherRole  string
	Age          int
	RetirementSeason *int
	GamesPlayed  int
	GamesBatting int
	AtBats       int
	Runs         int
	Hits         int
	Doubles      int
	Triples      int
	HomeRuns     int
	RBI          int
	StolenBases  int
	CaughtStealing int
	Walks        int
	Strikeouts   int
	HitByPitch   int
	SacHits      int
	SacFlies     int
	Errors       int
	PassedBalls  int
}

// SaveGamePitchingStat is a season pitching stat row from t_stats_pitching.
type SaveGamePitchingStat struct {
	AggregatorID int
	PlayerGUID   string
	FirstName    string
	LastName     string
	CurrentTeam  string
	PrevTeam     string
	Prev2Team    string
	PitcherRole  string
	Age          int
	RetirementSeason *int
	Wins         int
	Losses       int
	Games        int
	GamesStarted int
	CompleteGames int
	TotalPitches int
	Shutouts     int
	Saves        int
	OutsPitched  int // divide by 3 for IP
	HitsAllowed  int
	EarnedRuns   int
	HomeRunsAllowed int
	Walks        int
	Strikeouts   int
	HitBatters   int
	BattersFaced int
	GamesFinished int
	RunsAllowed  int
	WildPitches  int
}
