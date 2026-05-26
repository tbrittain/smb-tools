package models

// GameVersion identifies which Super Mega Baseball title a save file or
// franchise belongs to.
type GameVersion string

const (
	GameVersionSMB3 GameVersion = "smb3"
	GameVersionSMB4 GameVersion = "smb4"
)

// Valid reports whether the GameVersion is one of the known values.
func (v GameVersion) Valid() bool {
	return v == GameVersionSMB3 || v == GameVersionSMB4
}

func (v GameVersion) String() string { return string(v) }

// LeagueMode describes the game mode of a league in an SMB save file.
type LeagueMode string

const (
	LeagueModeFranchise   LeagueMode = "franchise"
	LeagueModeSeason      LeagueMode = "season"
	LeagueModeElimination LeagueMode = "elimination"
	// LeagueModeNone indicates the league exists but no games have been played.
	LeagueModeNone LeagueMode = "none"
)

func (m LeagueMode) String() string { return string(m) }

// CareerStatType discriminates career stat rows in player_career_batting_stats
// and player_career_pitching_stats. TotalCareer combines regular-season and
// playoff counting totals and recomputes all rates from the combined numerators.
type CareerStatType string

const (
	CareerStatTypeRegularSeason CareerStatType = "regular_season"
	CareerStatTypePlayoffs      CareerStatType = "playoffs"
	CareerStatTypeTotalCareer   CareerStatType = "total_career"
)

func (t CareerStatType) String() string { return string(t) }
