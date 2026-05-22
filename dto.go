package main

import (
	"smb-tools/internal/models"
)

// ── Seasons ───────────────────────────────────────────────────────────────────

// SeasonSummaryDTO is the frontend representation of one franchise season.
type SeasonSummaryDTO struct {
	ID                int64  `json:"id"`
	SeasonNum         int    `json:"seasonNum"`
	NumGames          int    `json:"numGames"`
	ImportedAt        string `json:"importedAt"`
	ChampionTeamName  string `json:"championTeamName"`
	ChampionHistoryID *int64 `json:"championHistoryId"`
}

// TeamStandingDTO is one row in a season standings table.
type TeamStandingDTO struct {
	HistoryID      int64   `json:"historyId"`
	TeamID         int64   `json:"teamId"`
	TeamName       string  `json:"teamName"`
	DivisionName   string  `json:"divisionName"`
	ConferenceName string  `json:"conferenceName"`
	Wins           int     `json:"wins"`
	Losses         int     `json:"losses"`
	WinPct         float64 `json:"winPct"`
	GamesBack      float64 `json:"gamesBack"`
	RunsFor        int     `json:"runsFor"`
	RunsAgainst    int     `json:"runsAgainst"`
	RunDiff        int     `json:"runDiff"`
	PlayoffSeed    *int    `json:"playoffSeed"`
}

// StatLeaderDTO is the title leader for one stat category in a season.
type StatLeaderDTO struct {
	PlayerID  int64   `json:"playerId"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	TeamName  string  `json:"teamName"`
	StatValue float64 `json:"statValue"`
}

// StatLeadersDTO groups the six title-leader categories for one season.
type StatLeadersDTO struct {
	SeasonNum  int            `json:"seasonNum"`
	BA         *StatLeaderDTO `json:"ba"`
	HR         *StatLeaderDTO `json:"hr"`
	RBI        *StatLeaderDTO `json:"rbi"`
	ERA        *StatLeaderDTO `json:"era"`
	Wins       *StatLeaderDTO `json:"wins"`
	Strikeouts *StatLeaderDTO `json:"strikeouts"`
}

// CareerLeaderDTO is one entry in an all-time career leaderboard category.
type CareerLeaderDTO struct {
	PlayerID      int64   `json:"playerId"`
	FirstName     string  `json:"firstName"`
	LastName      string  `json:"lastName"`
	StatValue     float64 `json:"statValue"`
	SeasonsPlayed int     `json:"seasonsPlayed"`
}

// CareerLeadersDTO groups the top-5 players for each all-time career category.
type CareerLeadersDTO struct {
	HR         []CareerLeaderDTO `json:"hr"`
	Hits       []CareerLeaderDTO `json:"hits"`
	RBI        []CareerLeaderDTO `json:"rbi"`
	Wins       []CareerLeaderDTO `json:"wins"`
	Strikeouts []CareerLeaderDTO `json:"strikeouts"`
	Saves      []CareerLeaderDTO `json:"saves"`
}

// ── Stats ─────────────────────────────────────────────────────────────────────

// CareerBattingStatsDTO contains counting stats and computed rate fields.
// Rate fields (*float64) are null when their denominator is zero.
type CareerBattingStatsDTO struct {
	GamesPlayed    int      `json:"gamesPlayed"`
	GamesBatting   int      `json:"gamesBatting"`
	AtBats         int      `json:"atBats"`
	Runs           int      `json:"runs"`
	Hits           int      `json:"hits"`
	Doubles        int      `json:"doubles"`
	Triples        int      `json:"triples"`
	HomeRuns       int      `json:"homeRuns"`
	RBI            int      `json:"rbi"`
	StolenBases    int      `json:"stolenBases"`
	CaughtStealing int      `json:"caughtStealing"`
	Walks          int      `json:"walks"`
	Strikeouts     int      `json:"strikeouts"`
	HitByPitch     int      `json:"hitByPitch"`
	SacHits        int      `json:"sacHits"`
	SacFlies       int      `json:"sacFlies"`
	Errors         int      `json:"errors"`
	PassedBalls    int      `json:"passedBalls"`
	BA             *float64 `json:"ba"`
	OBP            *float64 `json:"obp"`
	SLG            *float64 `json:"slg"`
	OPS            *float64 `json:"ops"`
	ISO            *float64 `json:"iso"`
	BABIP          *float64 `json:"babip"`
	KPct           *float64 `json:"kPct"`
	BBPct          *float64 `json:"bbPct"`
	ABPerHR        *float64 `json:"abPerHr"`
}

// CareerPitchingStatsDTO contains counting stats and computed rate fields.
// Rate fields (*float64) are null when their denominator is zero.
type CareerPitchingStatsDTO struct {
	Wins            int      `json:"wins"`
	Losses          int      `json:"losses"`
	Games           int      `json:"games"`
	GamesStarted    int      `json:"gamesStarted"`
	CompleteGames   int      `json:"completeGames"`
	Shutouts        int      `json:"shutouts"`
	Saves           int      `json:"saves"`
	OutsPitched     int      `json:"outsPitched"`
	HitsAllowed     int      `json:"hitsAllowed"`
	EarnedRuns      int      `json:"earnedRuns"`
	HomeRunsAllowed int      `json:"homeRunsAllowed"`
	Walks           int      `json:"walks"`
	Strikeouts      int      `json:"strikeouts"`
	HitBatters      int      `json:"hitBatters"`
	BattersFaced    int      `json:"battersFaced"`
	GamesFinished   int      `json:"gamesFinished"`
	RunsAllowed     int      `json:"runsAllowed"`
	WildPitches     int      `json:"wildPitches"`
	TotalPitches    int      `json:"totalPitches"`
	ERA             *float64 `json:"era"`
	WHIP            *float64 `json:"whip"`
	K9              *float64 `json:"k9"`
	BB9             *float64 `json:"bb9"`
	H9              *float64 `json:"h9"`
	HR9             *float64 `json:"hr9"`
	KPerBB          *float64 `json:"kPerBb"`
	KPct            *float64 `json:"kPct"`
	WinPct          *float64 `json:"winPct"`
	PPerIP          *float64 `json:"pPerIp"`
}

// ── Players ───────────────────────────────────────────────────────────────────

// PlayerSearchResultDTO is a lightweight player record for search results.
type PlayerSearchResultDTO struct {
	PlayerID      int64  `json:"playerId"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	IsHallOfFamer bool   `json:"isHallOfFamer"`
	SeasonsPlayed int    `json:"seasonsPlayed"`
	FirstSeason   int    `json:"firstSeason"`
	LastSeason    int    `json:"lastSeason"`
}

// PlayerCareerDTO is the full career record for one player.
type PlayerCareerDTO struct {
	PlayerID      int64                   `json:"playerId"`
	FirstName     string                  `json:"firstName"`
	LastName      string                  `json:"lastName"`
	IsHallOfFamer bool                    `json:"isHallOfFamer"`
	Batting       *CareerBattingStatsDTO  `json:"batting"`
	Pitching      *CareerPitchingStatsDTO `json:"pitching"`
}

// PlayerSeasonLogDTO is one row in a player's season-by-season breakdown.
type PlayerSeasonLogDTO struct {
	SeasonNum         int                     `json:"seasonNum"`
	SeasonID          int64                   `json:"seasonId"`
	TeamName          string                  `json:"teamName"`
	Age               int                     `json:"age"`
	Salary            int                     `json:"salary"`
	PrimaryPosition   string                  `json:"primaryPosition"`
	SecondaryPosition string                  `json:"secondaryPosition"`
	PitcherRole       string                  `json:"pitcherRole"`
	BatHand           string                  `json:"batHand"`
	ThrowHand         string                  `json:"throwHand"`
	ChemistryType     string                  `json:"chemistryType"`
	TraitsJSON        string                  `json:"traitsJson"`
	PitchesJSON       string                  `json:"pitchesJson"`
	Power             int                     `json:"power"`
	Contact           int                     `json:"contact"`
	Speed             int                     `json:"speed"`
	Fielding          int                     `json:"fielding"`
	Arm               int                     `json:"arm"`
	Velocity          int                     `json:"velocity"`
	Junk              int                     `json:"junk"`
	Accuracy          int                     `json:"accuracy"`
	Batting           *CareerBattingStatsDTO  `json:"batting"`
	Pitching          *CareerPitchingStatsDTO `json:"pitching"`
	PlayoffBatting    *CareerBattingStatsDTO  `json:"playoffBatting"`
	PlayoffPitching   *CareerPitchingStatsDTO `json:"playoffPitching"`
}

// ── Teams ─────────────────────────────────────────────────────────────────────

// TeamSearchResultDTO is a lightweight team record for search results.
type TeamSearchResultDTO struct {
	TeamID      int64  `json:"teamId"`
	TeamName    string `json:"teamName"`
	Seasons     int    `json:"seasons"`
	FirstSeason int    `json:"firstSeason"`
	LastSeason  int    `json:"lastSeason"`
}

// TeamSeasonSummaryDTO is one season entry in a team's history.
type TeamSeasonSummaryDTO struct {
	HistoryID          int64   `json:"historyId"`
	SeasonID           int64   `json:"seasonId"`
	SeasonNum          int     `json:"seasonNum"`
	TeamName           string  `json:"teamName"`
	DivisionName       string  `json:"divisionName"`
	ConferenceName     string  `json:"conferenceName"`
	Wins               int     `json:"wins"`
	Losses             int     `json:"losses"`
	WinPct             float64 `json:"winPct"`
	GamesBack          float64 `json:"gamesBack"`
	RunsFor            int     `json:"runsFor"`
	RunsAgainst        int     `json:"runsAgainst"`
	Budget             int     `json:"budget"`
	Payroll            int     `json:"payroll"`
	PlayoffSeed        *int    `json:"playoffSeed"`
	PlayoffWins        *int    `json:"playoffWins"`
	PlayoffLosses      *int    `json:"playoffLosses"`
	PlayoffRunsFor     *int    `json:"playoffRunsFor"`
	PlayoffRunsAgainst *int    `json:"playoffRunsAgainst"`
	TotalPower         int     `json:"totalPower"`
	TotalContact       int     `json:"totalContact"`
	TotalSpeed         int     `json:"totalSpeed"`
	TotalFielding      int     `json:"totalFielding"`
	TotalArm           int     `json:"totalArm"`
	TotalVelocity      int     `json:"totalVelocity"`
	TotalJunk          int     `json:"totalJunk"`
	TotalAccuracy      int     `json:"totalAccuracy"`
	IsChampion         bool    `json:"isChampion"`
}

// TeamHistoryDTO is the full all-time record for one team.
type TeamHistoryDTO struct {
	TeamID   int64                  `json:"teamId"`
	GameGUID string                 `json:"gameGuid"`
	Seasons  []TeamSeasonSummaryDTO `json:"seasons"`
}

// TeamSeasonListDTO is one row in the historical teams list page.
type TeamSeasonListDTO struct {
	SeasonNum      int     `json:"seasonNum"`
	HistoryID      int64   `json:"historyId"`
	TeamID         int64   `json:"teamId"`
	TeamName       string  `json:"teamName"`
	ConferenceName string  `json:"conferenceName"`
	DivisionName   string  `json:"divisionName"`
	Wins           int     `json:"wins"`
	Losses         int     `json:"losses"`
	WinPct         float64 `json:"winPct"`
	RunsFor        int     `json:"runsFor"`
	RunsAgainst    int     `json:"runsAgainst"`
	PlayoffSeed    *int    `json:"playoffSeed"`
	PlayoffWins    *int    `json:"playoffWins"`
	PlayoffLosses  *int    `json:"playoffLosses"`
	IsChampion     bool    `json:"isChampion"`
}

// RosterPlayerDTO is one player row in a team season roster.
type RosterPlayerDTO struct {
	PlayerID          int64                   `json:"playerId"`
	FirstName         string                  `json:"firstName"`
	LastName          string                  `json:"lastName"`
	IsHallOfFamer     bool                    `json:"isHallOfFamer"`
	Age               int                     `json:"age"`
	Salary            int                     `json:"salary"`
	PrimaryPosition   string                  `json:"primaryPosition"`
	SecondaryPosition string                  `json:"secondaryPosition"`
	PitcherRole       string                  `json:"pitcherRole"`
	BatHand           string                  `json:"batHand"`
	ThrowHand         string                  `json:"throwHand"`
	ChemistryType     string                  `json:"chemistryType"`
	TraitsJSON        string                  `json:"traitsJson"`
	PitchesJSON       string                  `json:"pitchesJson"`
	Power             int                     `json:"power"`
	Contact           int                     `json:"contact"`
	Speed             int                     `json:"speed"`
	Fielding          int                     `json:"fielding"`
	Arm               int                     `json:"arm"`
	Velocity          int                     `json:"velocity"`
	Junk              int                     `json:"junk"`
	Accuracy          int                     `json:"accuracy"`
	Batting           *CareerBattingStatsDTO  `json:"batting"`
	Pitching          *CareerPitchingStatsDTO `json:"pitching"`
}

// ScheduleGameDTO is one game in a team's regular season schedule.
type ScheduleGameDTO struct {
	GameNumber        int    `json:"gameNumber"`
	Day               int    `json:"day"`
	HomeTeamHistoryID int64  `json:"homeTeamHistoryId"`
	HomeTeamName      string `json:"homeTeamName"`
	AwayTeamHistoryID int64  `json:"awayTeamHistoryId"`
	AwayTeamName      string `json:"awayTeamName"`
	HomeScore         *int   `json:"homeScore"`
	AwayScore         *int   `json:"awayScore"`
	HomePitcherName   string `json:"homePitcherName"`
	AwayPitcherName   string `json:"awayPitcherName"`
}

// PlayoffGameDTO is one game in a team's playoff schedule.
type PlayoffGameDTO struct {
	SeriesNumber      int    `json:"seriesNumber"`
	GameNumber        int    `json:"gameNumber"`
	HomeTeamHistoryID int64  `json:"homeTeamHistoryId"`
	HomeTeamName      string `json:"homeTeamName"`
	AwayTeamHistoryID int64  `json:"awayTeamHistoryId"`
	AwayTeamName      string `json:"awayTeamName"`
	HomeScore         *int   `json:"homeScore"`
	AwayScore         *int   `json:"awayScore"`
	HomePitcherName   string `json:"homePitcherName"`
	AwayPitcherName   string `json:"awayPitcherName"`
}

// TeamSeasonDetailDTO bundles everything needed for the team season detail page.
type TeamSeasonDetailDTO struct {
	Team     TeamSeasonSummaryDTO `json:"team"`
	Roster   []RosterPlayerDTO    `json:"roster"`
	Schedule []ScheduleGameDTO    `json:"schedule"`
	Playoffs []PlayoffGameDTO     `json:"playoffs"`
}

// ── Mapping helpers ───────────────────────────────────────────────────────────

func battingToDTO(b *models.CareerBattingStats) *CareerBattingStatsDTO {
	if b == nil {
		return nil
	}
	return &CareerBattingStatsDTO{
		GamesPlayed:    b.GamesPlayed,
		GamesBatting:   b.GamesBatting,
		AtBats:         b.AtBats,
		Runs:           b.Runs,
		Hits:           b.Hits,
		Doubles:        b.Doubles,
		Triples:        b.Triples,
		HomeRuns:       b.HomeRuns,
		RBI:            b.RBI,
		StolenBases:    b.StolenBases,
		CaughtStealing: b.CaughtStealing,
		Walks:          b.Walks,
		Strikeouts:     b.Strikeouts,
		HitByPitch:     b.HitByPitch,
		SacHits:        b.SacHits,
		SacFlies:       b.SacFlies,
		Errors:         b.Errors,
		PassedBalls:    b.PassedBalls,
		BA:             b.BA,
		OBP:            b.OBP,
		SLG:            b.SLG,
		OPS:            b.OPS,
		ISO:            b.ISO,
		BABIP:          b.BABIP,
		KPct:           b.KPct,
		BBPct:          b.BBPct,
		ABPerHR:        b.ABPerHR,
	}
}

func pitchingToDTO(p *models.CareerPitchingStats) *CareerPitchingStatsDTO {
	if p == nil {
		return nil
	}
	return &CareerPitchingStatsDTO{
		Wins:            p.Wins,
		Losses:          p.Losses,
		Games:           p.Games,
		GamesStarted:    p.GamesStarted,
		CompleteGames:   p.CompleteGames,
		Shutouts:        p.Shutouts,
		Saves:           p.Saves,
		OutsPitched:     p.OutsPitched,
		HitsAllowed:     p.HitsAllowed,
		EarnedRuns:      p.EarnedRuns,
		HomeRunsAllowed: p.HomeRunsAllowed,
		Walks:           p.Walks,
		Strikeouts:      p.Strikeouts,
		HitBatters:      p.HitBatters,
		BattersFaced:    p.BattersFaced,
		GamesFinished:   p.GamesFinished,
		RunsAllowed:     p.RunsAllowed,
		WildPitches:     p.WildPitches,
		TotalPitches:    p.TotalPitches,
		ERA:             p.ERA,
		WHIP:            p.WHIP,
		K9:              p.K9,
		BB9:             p.BB9,
		H9:              p.H9,
		HR9:             p.HR9,
		KPerBB:          p.KPerBB,
		KPct:            p.KPct,
		WinPct:          p.WinPct,
		PPerIP:          p.PPerIP,
	}
}

func teamSeasonSummaryToDTO(ts models.TeamSeasonSummary) TeamSeasonSummaryDTO {
	return TeamSeasonSummaryDTO{
		HistoryID:          ts.HistoryID,
		SeasonID:           ts.SeasonID,
		SeasonNum:          ts.SeasonNum,
		TeamName:           ts.TeamName,
		DivisionName:       ts.DivisionName,
		ConferenceName:     ts.ConferenceName,
		Wins:               ts.Wins,
		Losses:             ts.Losses,
		WinPct:             ts.WinPct,
		GamesBack:          ts.GamesBack,
		RunsFor:            ts.RunsFor,
		RunsAgainst:        ts.RunsAgainst,
		Budget:             ts.Budget,
		Payroll:            ts.Payroll,
		PlayoffSeed:        ts.PlayoffSeed,
		PlayoffWins:        ts.PlayoffWins,
		PlayoffLosses:      ts.PlayoffLosses,
		PlayoffRunsFor:     ts.PlayoffRunsFor,
		PlayoffRunsAgainst: ts.PlayoffRunsAgainst,
		TotalPower:         ts.TotalPower,
		TotalContact:       ts.TotalContact,
		TotalSpeed:         ts.TotalSpeed,
		TotalFielding:      ts.TotalFielding,
		TotalArm:           ts.TotalArm,
		TotalVelocity:      ts.TotalVelocity,
		TotalJunk:          ts.TotalJunk,
		TotalAccuracy:      ts.TotalAccuracy,
		IsChampion:         ts.IsChampion,
	}
}

func rosterPlayerToDTO(r models.RosterPlayer) RosterPlayerDTO {
	return RosterPlayerDTO{
		PlayerID:          r.PlayerID,
		FirstName:         r.FirstName,
		LastName:          r.LastName,
		IsHallOfFamer:     r.IsHallOfFamer,
		Age:               r.Age,
		Salary:            r.Salary,
		PrimaryPosition:   r.PrimaryPosition,
		SecondaryPosition: r.SecondaryPosition,
		PitcherRole:       r.PitcherRole,
		BatHand:           r.BatHand,
		ThrowHand:         r.ThrowHand,
		ChemistryType:     r.ChemistryType,
		TraitsJSON:        r.TraitsJSON,
		PitchesJSON:       r.PitchesJSON,
		Power:             r.Power,
		Contact:           r.Contact,
		Speed:             r.Speed,
		Fielding:          r.Fielding,
		Arm:               r.Arm,
		Velocity:          r.Velocity,
		Junk:              r.Junk,
		Accuracy:          r.Accuracy,
		Batting:           battingToDTO(r.Batting),
		Pitching:          pitchingToDTO(r.Pitching),
	}
}

func scheduleGameToDTO(g models.ScheduleGameRow) ScheduleGameDTO {
	return ScheduleGameDTO{
		GameNumber:        g.GameNumber,
		Day:               g.Day,
		HomeTeamHistoryID: g.HomeTeamHistoryID,
		HomeTeamName:      g.HomeTeamName,
		AwayTeamHistoryID: g.AwayTeamHistoryID,
		AwayTeamName:      g.AwayTeamName,
		HomeScore:         g.HomeScore,
		AwayScore:         g.AwayScore,
		HomePitcherName:   g.HomePitcherName,
		AwayPitcherName:   g.AwayPitcherName,
	}
}

func playoffGameToDTO(g models.PlayoffGameRow) PlayoffGameDTO {
	return PlayoffGameDTO{
		SeriesNumber:      g.SeriesNumber,
		GameNumber:        g.GameNumber,
		HomeTeamHistoryID: g.HomeTeamHistoryID,
		HomeTeamName:      g.HomeTeamName,
		AwayTeamHistoryID: g.AwayTeamHistoryID,
		AwayTeamName:      g.AwayTeamName,
		HomeScore:         g.HomeScore,
		AwayScore:         g.AwayScore,
		HomePitcherName:   g.HomePitcherName,
		AwayPitcherName:   g.AwayPitcherName,
	}
}

// ── Leaderboards ──────────────────────────────────────────────────────────────

// LeaderboardFiltersDTO carries filter parameters from the frontend.
// Zero values (empty string, false, 0) mean "no filter applied".
type LeaderboardFiltersDTO struct {
	IsPlayoffs       bool   `json:"isPlayoffs"`
	OnlyHallOfFamers bool   `json:"onlyHallOfFamers"`
	Position         string `json:"position"`
	BatHand          string `json:"batHand"`
	ThrowHand        string `json:"throwHand"`
	ChemistryType    string `json:"chemistryType"`
	SeasonStart      int    `json:"seasonStart"`
	SeasonEnd        int    `json:"seasonEnd"`
}

// BattingLeaderRowDTO is one row in a batting leaderboard (career or season).
// The DTO is flat so PrimeVue DataTable sort-field can reference top-level keys.
// Career rows have SeasonsPlayed > 0; season rows have SeasonNum > 0.
type BattingLeaderRowDTO struct {
	PlayerID        int64  `json:"playerId"`
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	IsHallOfFamer   bool   `json:"isHallOfFamer"`
	SeasonsPlayed   int    `json:"seasonsPlayed"`
	SeasonNum       int    `json:"seasonNum"`
	TeamName        string `json:"teamName"`
	Age             int    `json:"age"`
	PrimaryPosition string `json:"primaryPosition"`
	BatHand         string `json:"batHand"`
	ChemistryType   string `json:"chemistryType"`
	// Counting stats
	GamesPlayed    int `json:"gamesPlayed"`
	GamesBatting   int `json:"gamesBatting"`
	AtBats         int `json:"atBats"`
	Runs           int `json:"runs"`
	Hits           int `json:"hits"`
	Doubles        int `json:"doubles"`
	Triples        int `json:"triples"`
	HomeRuns       int `json:"homeRuns"`
	RBI            int `json:"rbi"`
	StolenBases    int `json:"stolenBases"`
	CaughtStealing int `json:"caughtStealing"`
	Walks          int `json:"walks"`
	Strikeouts     int `json:"strikeouts"`
	HitByPitch     int `json:"hitByPitch"`
	SacHits        int `json:"sacHits"`
	SacFlies       int `json:"sacFlies"`
	Errors         int `json:"errors"`
	PassedBalls    int `json:"passedBalls"`
	// Computed rate fields (nil when denominator is zero)
	BA      *float64 `json:"ba"`
	OBP     *float64 `json:"obp"`
	SLG     *float64 `json:"slg"`
	OPS     *float64 `json:"ops"`
	ISO     *float64 `json:"iso"`
	BABIP   *float64 `json:"babip"`
	KPct    *float64 `json:"kPct"`
	BBPct   *float64 `json:"bbPct"`
	ABPerHR *float64 `json:"abPerHr"`
}

// PitchingLeaderRowDTO is one row in a pitching leaderboard (career or season).
// Career rows have SeasonsPlayed > 0; season rows have SeasonNum > 0.
type PitchingLeaderRowDTO struct {
	PlayerID      int64  `json:"playerId"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	IsHallOfFamer bool   `json:"isHallOfFamer"`
	SeasonsPlayed int    `json:"seasonsPlayed"`
	SeasonNum     int    `json:"seasonNum"`
	TeamName      string `json:"teamName"`
	Age           int    `json:"age"`
	PitcherRole   string `json:"pitcherRole"`
	ThrowHand     string `json:"throwHand"`
	ChemistryType string `json:"chemistryType"`
	// Counting stats
	Wins            int `json:"wins"`
	Losses          int `json:"losses"`
	Games           int `json:"games"`
	GamesStarted    int `json:"gamesStarted"`
	CompleteGames   int `json:"completeGames"`
	Shutouts        int `json:"shutouts"`
	Saves           int `json:"saves"`
	OutsPitched     int `json:"outsPitched"`
	HitsAllowed     int `json:"hitsAllowed"`
	EarnedRuns      int `json:"earnedRuns"`
	HomeRunsAllowed int `json:"homeRunsAllowed"`
	Walks           int `json:"walks"`
	Strikeouts      int `json:"strikeouts"`
	HitBatters      int `json:"hitBatters"`
	BattersFaced    int `json:"battersFaced"`
	GamesFinished   int `json:"gamesFinished"`
	RunsAllowed     int `json:"runsAllowed"`
	WildPitches     int `json:"wildPitches"`
	TotalPitches    int `json:"totalPitches"`
	// Computed rate fields (nil when denominator is zero)
	ERA    *float64 `json:"era"`
	WHIP   *float64 `json:"whip"`
	K9     *float64 `json:"k9"`
	BB9    *float64 `json:"bb9"`
	H9     *float64 `json:"h9"`
	HR9    *float64 `json:"hr9"`
	KPerBB *float64 `json:"kPerBb"`
	KPct   *float64 `json:"kPct"`
	WinPct *float64 `json:"winPct"`
	PPerIP *float64 `json:"pPerIp"`
}

// ── Leaderboard mapping helpers ───────────────────────────────────────────────

func leaderboardFiltersToDomain(f LeaderboardFiltersDTO) models.LeaderboardFilters {
	return models.LeaderboardFilters{
		IsPlayoffs:       f.IsPlayoffs,
		OnlyHallOfFamers: f.OnlyHallOfFamers,
		Position:         f.Position,
		BatHand:          f.BatHand,
		ThrowHand:        f.ThrowHand,
		ChemistryType:    f.ChemistryType,
		SeasonStart:      f.SeasonStart,
		SeasonEnd:        f.SeasonEnd,
	}
}

func battingStatsToLeaderDTO(b models.CareerBattingStats) (
	gp, gb, ab, r, h, d, tr, hr, rbi, sb, cs, bb, k, hbp, sh, sf, e, pb int,
	ba, obp, slg, ops, iso, babip, kpct, bbpct, abhr *float64,
) {
	return b.GamesPlayed, b.GamesBatting,
		b.AtBats, b.Runs, b.Hits, b.Doubles, b.Triples, b.HomeRuns, b.RBI,
		b.StolenBases, b.CaughtStealing, b.Walks, b.Strikeouts,
		b.HitByPitch, b.SacHits, b.SacFlies, b.Errors, b.PassedBalls,
		b.BA, b.OBP, b.SLG, b.OPS, b.ISO, b.BABIP, b.KPct, b.BBPct, b.ABPerHR
}

func battingCareerLeaderToDTO(r models.BattingCareerLeaderRow) BattingLeaderRowDTO {
	gp, gb, ab, runs, h, d, tr, hr, rbi, sb, cs, bb, k, hbp, sh, sf, e, pb,
		ba, obp, slg, ops, iso, babip, kpct, bbpct, abhr := battingStatsToLeaderDTO(r.CareerBattingStats)
	return BattingLeaderRowDTO{
		PlayerID: r.PlayerID, FirstName: r.FirstName, LastName: r.LastName,
		IsHallOfFamer: r.IsHallOfFamer, SeasonsPlayed: r.SeasonsPlayed,
		GamesPlayed: gp, GamesBatting: gb,
		AtBats: ab, Runs: runs, Hits: h, Doubles: d, Triples: tr, HomeRuns: hr, RBI: rbi,
		StolenBases: sb, CaughtStealing: cs, Walks: bb, Strikeouts: k,
		HitByPitch: hbp, SacHits: sh, SacFlies: sf, Errors: e, PassedBalls: pb,
		BA: ba, OBP: obp, SLG: slg, OPS: ops, ISO: iso,
		BABIP: babip, KPct: kpct, BBPct: bbpct, ABPerHR: abhr,
	}
}

func battingSeasonLeaderToDTO(r models.BattingSeasonLeaderRow) BattingLeaderRowDTO {
	gp, gb, ab, runs, h, d, tr, hr, rbi, sb, cs, bb, k, hbp, sh, sf, e, pb,
		ba, obp, slg, ops, iso, babip, kpct, bbpct, abhr := battingStatsToLeaderDTO(r.CareerBattingStats)
	return BattingLeaderRowDTO{
		PlayerID: r.PlayerID, FirstName: r.FirstName, LastName: r.LastName,
		IsHallOfFamer: r.IsHallOfFamer,
		SeasonNum: r.SeasonNum, TeamName: r.TeamName, Age: r.Age,
		PrimaryPosition: r.PrimaryPosition, BatHand: r.BatHand, ChemistryType: r.ChemistryType,
		GamesPlayed: gp, GamesBatting: gb,
		AtBats: ab, Runs: runs, Hits: h, Doubles: d, Triples: tr, HomeRuns: hr, RBI: rbi,
		StolenBases: sb, CaughtStealing: cs, Walks: bb, Strikeouts: k,
		HitByPitch: hbp, SacHits: sh, SacFlies: sf, Errors: e, PassedBalls: pb,
		BA: ba, OBP: obp, SLG: slg, OPS: ops, ISO: iso,
		BABIP: babip, KPct: kpct, BBPct: bbpct, ABPerHR: abhr,
	}
}

func pitchingStatsToLeaderDTO(p models.CareerPitchingStats) (
	w, l, g, gs, cg, sho, sv, outs, h, er, hra, bb, k, hb, bf, gf, ra, wp, tp int,
	era, whip, k9, bb9, h9, hr9, kperbb, kpct, winpct, pperip *float64,
) {
	return p.Wins, p.Losses, p.Games, p.GamesStarted,
		p.CompleteGames, p.Shutouts, p.Saves, p.OutsPitched,
		p.HitsAllowed, p.EarnedRuns, p.HomeRunsAllowed, p.Walks,
		p.Strikeouts, p.HitBatters, p.BattersFaced, p.GamesFinished,
		p.RunsAllowed, p.WildPitches, p.TotalPitches,
		p.ERA, p.WHIP, p.K9, p.BB9, p.H9, p.HR9, p.KPerBB, p.KPct, p.WinPct, p.PPerIP
}

func pitchingCareerLeaderToDTO(r models.PitchingCareerLeaderRow) PitchingLeaderRowDTO {
	w, l, g, gs, cg, sho, sv, outs, h, er, hra, bb, k, hb, bf, gf, ra, wp, tp,
		era, whip, k9, bb9, h9, hr9, kperbb, kpct, winpct, pperip := pitchingStatsToLeaderDTO(r.CareerPitchingStats)
	return PitchingLeaderRowDTO{
		PlayerID: r.PlayerID, FirstName: r.FirstName, LastName: r.LastName,
		IsHallOfFamer: r.IsHallOfFamer, SeasonsPlayed: r.SeasonsPlayed,
		Wins: w, Losses: l, Games: g, GamesStarted: gs,
		CompleteGames: cg, Shutouts: sho, Saves: sv, OutsPitched: outs,
		HitsAllowed: h, EarnedRuns: er, HomeRunsAllowed: hra, Walks: bb,
		Strikeouts: k, HitBatters: hb, BattersFaced: bf, GamesFinished: gf,
		RunsAllowed: ra, WildPitches: wp, TotalPitches: tp,
		ERA: era, WHIP: whip, K9: k9, BB9: bb9, H9: h9, HR9: hr9,
		KPerBB: kperbb, KPct: kpct, WinPct: winpct, PPerIP: pperip,
	}
}

func pitchingSeasonLeaderToDTO(r models.PitchingSeasonLeaderRow) PitchingLeaderRowDTO {
	w, l, g, gs, cg, sho, sv, outs, h, er, hra, bb, k, hb, bf, gf, ra, wp, tp,
		era, whip, k9, bb9, h9, hr9, kperbb, kpct, winpct, pperip := pitchingStatsToLeaderDTO(r.CareerPitchingStats)
	return PitchingLeaderRowDTO{
		PlayerID: r.PlayerID, FirstName: r.FirstName, LastName: r.LastName,
		IsHallOfFamer: r.IsHallOfFamer,
		SeasonNum: r.SeasonNum, TeamName: r.TeamName, Age: r.Age,
		PitcherRole: r.PitcherRole, ThrowHand: r.ThrowHand, ChemistryType: r.ChemistryType,
		Wins: w, Losses: l, Games: g, GamesStarted: gs,
		CompleteGames: cg, Shutouts: sho, Saves: sv, OutsPitched: outs,
		HitsAllowed: h, EarnedRuns: er, HomeRunsAllowed: hra, Walks: bb,
		Strikeouts: k, HitBatters: hb, BattersFaced: bf, GamesFinished: gf,
		RunsAllowed: ra, WildPitches: wp, TotalPitches: tp,
		ERA: era, WHIP: whip, K9: k9, BB9: bb9, H9: h9, HR9: hr9,
		KPerBB: kperbb, KPct: kpct, WinPct: winpct, PPerIP: pperip,
	}
}
