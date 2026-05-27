package models

// Award represents one award definition (built-in or user-created).
type Award struct {
	ID                int64
	Name              string
	OriginalName      string
	Importance        int
	OmitFromGroupings bool
	IsBattingAward    bool
	IsPitchingAward   bool
	IsFieldingAward   bool
	IsPlayoffAward    bool
	IsUserAssignable  bool
	IsBuiltIn         bool
}

// PlayerSeasonAwardRow is one player-season entry returned for the awards
// delegation page, pre-loaded with the awards already assigned.
type PlayerSeasonAwardRow struct {
	PlayerSeasonID int64
	PlayerID       int64
	FirstName      string
	LastName       string
	TeamName       string
	PrimaryPos     string
	PitcherRole    string
	Awards         []Award
}

// ── Award delegation candidates ───────────────────────────────────────────────

// BattingCandidate is one player row in an award delegation batting section.
type BattingCandidate struct {
	PlayerSeasonID int64
	PlayerID       int64
	FirstName      string
	LastName       string
	TeamName       string
	PrimaryPos     string
	PitcherRole    string
	// Regular-season batting counting stats
	AtBats      int
	Hits        int
	HomeRuns    int
	RBI         int
	Walks       int
	Runs        int
	StolenBases int
	Strikeouts  int
	Doubles     int
	Triples     int
	// Rate stats (0 when no plate appearances)
	BA  float64
	OBP float64
	SLG float64
	OPS float64
	// Context stats (nil for pre-Phase-8.5 seasons)
	OPSPlus *float64
	SmbWAR  *float64
	// IsChampionTeam is true when this player's team won the championship.
	IsChampionTeam bool
	// Current user-assignable award IDs (pre-populated from DB or auto-suggested)
	AwardIDs []int64
}

// PitchingCandidate is one player row in an award delegation pitching section.
type PitchingCandidate struct {
	PlayerSeasonID  int64
	PlayerID        int64
	FirstName       string
	LastName        string
	TeamName        string
	PrimaryPos      string
	PitcherRole     string
	// Regular-season pitching counting stats
	Wins            int
	Losses          int
	Saves           int
	OutsPitched     int // divide by 3 for IP display
	HitsAllowed     int
	EarnedRuns      int
	Walks           int
	Strikeouts      int
	HomeRunsAllowed int
	CompleteGames   int
	Shutouts        int
	// Rate stats (0 when no outs pitched)
	ERA    float64
	WHIP   float64
	K9     float64
	BB9    float64
	H9     float64
	HR9    float64
	KPerBB float64
	// Context stats (nil for pre-Phase-8.5 seasons)
	ERAPlus  *float64
	FIPMinus *float64
	SmbWAR   *float64
	// IsChampionTeam is true when this player's team won the championship.
	IsChampionTeam bool
	// Current user-assignable award IDs
	AwardIDs []int64
}

// TeamAwardCandidates groups top batters and pitchers for one team in a season.
type TeamAwardCandidates struct {
	HistoryID int64
	TeamName  string
	Batters   []BattingCandidate
	Pitchers  []PitchingCandidate
}

// PositionAwardCandidates groups top batters at one fielding position.
type PositionAwardCandidates struct {
	Position string
	Batters  []BattingCandidate
}

// SeasonAwardCandidates is the full payload returned for the awards delegation page.
type SeasonAwardCandidates struct {
	SeasonID          int64
	SeasonNum         int
	TopBatters        []BattingCandidate
	TopPitchers       []PitchingCandidate
	TopRookieBatters  []BattingCandidate
	TopRookiePitchers []PitchingCandidate
	ByTeam            []TeamAwardCandidates
	ByPosition        []PositionAwardCandidates
	// PlayoffBatters/PlayoffPitchers: top-10 performers across all playoff participants.
	PlayoffBatters  []BattingCandidate
	PlayoffPitchers []PitchingCandidate
	// ChampionBatters/ChampionPitchers: top-10 performers from the championship-winning
	// team only. Populated server-side so the "Champions only" toggle never filters a
	// pre-truncated list.
	ChampionBatters  []BattingCandidate
	ChampionPitchers []PitchingCandidate
	// AutoSuggested is true when award IDs were pre-populated because no
	// user-assignable awards existed for this season yet.
	AutoSuggested bool
}

// TeamSeasonRef identifies a team within a specific season.
type TeamSeasonRef struct {
	HistoryID int64
	TeamName  string
}

// PlayerAwardEntry is one (playerSeasonID, []awardID) pair used in batch submission.
type PlayerAwardEntry struct {
	PlayerSeasonID int64
	AwardIDs       []int64
}

// HoFCandidate carries career-aggregate stats for a player who is eligible
// for Hall of Fame consideration (retired, not active in the latest season).
type HoFCandidate struct {
	PlayerID       int64
	FirstName      string
	LastName       string
	IsHallOfFamer  bool
	FirstSeason    int
	LastSeason     int
	Seasons        int
	Hits           int
	HomeRuns       int
	RBI            int
	StolenBases    int
	AtBats         int
	Walks          int
	Wins           int
	Losses         int
	OutsPitched    int
	Strikeouts     int
	EarnedRuns     int
	SmbWAR         float64
}

// HoFPage is a paginated response for Hall of Fame candidates or inductees.
type HoFPage struct {
	Items []HoFCandidate
	Total int
}
