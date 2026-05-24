package store

import (
	"context"
	"database/sql"
	"fmt"
)

// LegacyCompanionReader reads from a SmbExplorerCompanion SQLite database.
// It resolves integer lookup IDs to text domain values at read time so the
// migration service receives clean domain structs, never raw codes.
//
// All SQL uses exact column names from the EF Core migration files
// (20230731165716_Initial.cs and later migrations). Do not rename.
type LegacyCompanionReader struct {
	db              *sql.DB
	positionByID    map[int]string
	batHandByID     map[int]string
	throwHandByID   map[int]string
	pitcherRoleByID map[int]string
	chemistryByID   map[int]string
}

// NewLegacyCompanionReader creates a reader and pre-loads all lookup tables.
func NewLegacyCompanionReader(ctx context.Context, db *sql.DB) (*LegacyCompanionReader, error) {
	r := &LegacyCompanionReader{db: db}
	if err := r.loadLookups(ctx); err != nil {
		return nil, fmt.Errorf("loading legacy lookups: %w", err)
	}
	return r, nil
}

func (r *LegacyCompanionReader) loadLookups(ctx context.Context) error {
	var err error
	if r.positionByID, err = r.loadStringMap(ctx, `SELECT Id, Name FROM Positions`); err != nil {
		return fmt.Errorf("positions: %w", err)
	}
	if r.batHandByID, err = r.loadStringMap(ctx, `SELECT Id, Name FROM BatHandedness`); err != nil {
		return fmt.Errorf("bat handedness: %w", err)
	}
	if r.throwHandByID, err = r.loadStringMap(ctx, `SELECT Id, Name FROM ThrowHandedness`); err != nil {
		return fmt.Errorf("throw handedness: %w", err)
	}
	if r.pitcherRoleByID, err = r.loadStringMap(ctx, `SELECT Id, Name FROM PitcherRoles`); err != nil {
		return fmt.Errorf("pitcher roles: %w", err)
	}
	if r.chemistryByID, err = r.loadStringMap(ctx, `SELECT Id, Name FROM Chemistry`); err != nil {
		return fmt.Errorf("chemistry: %w", err)
	}
	return nil
}

func (r *LegacyCompanionReader) loadStringMap(ctx context.Context, query string) (map[int]string, error) {
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	m := make(map[int]string)
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		m[id] = name
	}
	return m, rows.Err()
}

// --- Domain structs returned by the reader ---

// LegacyFranchise represents one row from the Franchises table.
type LegacyFranchise struct {
	ID     int
	Name   string
	IsSmb3 bool
}

// LegacySeason represents one row from the Seasons table.
type LegacySeason struct {
	ID                    int // Seasons.Id — save game season ID (not auto-generated)
	Number                int // Seasons.Number — display season number
	NumGamesRegularSeason int
}

// LegacyTeam represents a team with all historical GUIDs from TeamGameIdHistory.
type LegacyTeam struct {
	ID       int
	GameGUIDs []string // [0] = primary GUID (lowest Id), [1:] = alt GUIDs
}

// LegacySeasonTeamHistory is one row from SeasonTeamHistory with joins resolved.
type LegacySeasonTeamHistory struct {
	ID             int
	SeasonID       int
	TeamID         int
	TeamName       string
	ConferenceName string
	DivisionName   string
	Budget         int
	Payroll        int
	Wins           int
	Losses         int
	GamesBehind    float64
	RunsScored     int
	RunsAllowed    int
	TotalPower     int
	TotalContact   int
	TotalSpeed     int
	TotalFielding  int
	TotalArm       int
	TotalVelocity  int
	TotalJunk      int
	TotalAccuracy  int
	PlayoffSeed    *int
	PlayoffWins    *int
	PlayoffLosses  *int
	PlayoffRunsScored  *int
	PlayoffRunsAllowed *int
}

// LegacyPlayer represents a player with lookup IDs resolved and all GUIDs loaded.
type LegacyPlayer struct {
	ID              int
	FirstName       string
	LastName        string
	IsHallOfFamer   bool
	BatHand         string
	ThrowHand       string
	PrimaryPosition string
	PitcherRole     string // empty if not a pitcher
	ChemistryType   string // empty if none
	GameGUIDs       []string // [0] = primary GUID, [1:] = alt GUIDs
}

// LegacyPlayerSeason represents one row from PlayerSeasons with secondary position resolved.
type LegacyPlayerSeason struct {
	ID                int
	PlayerID          int
	SeasonID          int
	Age               int
	Salary            int
	SecondaryPosition string // empty if none
}

// LegacyPlayerSeasonTeam represents one row from PlayerTeamHistory (all Orders).
// Order 1 = current/final team, 2 = most recently played prior, 3 = two teams ago.
// SortOrder is Order-1 to match the companion DB convention (0-indexed).
type LegacyPlayerSeasonTeam struct {
	PlayerSeasonID int
	TeamHistID     int
	SortOrder      int // 0=current, 1=prior, 2=two teams ago
}

// LegacyGameStats represents one row from PlayerSeasonGameStats with nulls coalesced.
type LegacyGameStats struct {
	PlayerSeasonID int
	Power          int
	Contact        int
	Speed          int
	Fielding       int
	Arm            int // NULL → 0
	Velocity       int // NULL → 0
	Junk           int // NULL → 0
	Accuracy       int // NULL → 0
}

// LegacyBattingStat represents one row from PlayerSeasonBattingStats (counting stats only).
type LegacyBattingStat struct {
	PlayerSeasonID  int
	IsRegularSeason bool
	GamesPlayed     int
	GamesBatting    int
	AtBats          int
	Runs            int
	Hits            int
	Doubles         int
	Triples         int
	HomeRuns        int
	RunsBattedIn    int
	StolenBases     int
	CaughtStealing  int
	Walks           int
	Strikeouts      int
	HitByPitch      int
	SacrificeHits   int
	SacrificeFlies  int
	Errors          int
	PassedBalls     int
}

// LegacyPitchingStat represents one row from PlayerSeasonPitchingStats (counting stats only).
type LegacyPitchingStat struct {
	PlayerSeasonID  int
	IsRegularSeason bool
	Wins            int
	Losses          int
	GamesPlayed     int // legacy column name; maps to new schema's "games"
	GamesStarted    int
	CompleteGames   int
	Shutouts        int
	Saves           int
	InningsPitched  *float64 // NULL if never pitched
	HitsAllowed     int      // legacy column "Hits"
	EarnedRuns      int
	HomeRunsAllowed int // legacy column "HomeRuns"
	Walks           int
	Strikeouts      int
	HitBatters      int // legacy column "HitByPitch"
	BattersFaced    int
	GamesFinished   int
	RunsAllowed     int
	WildPitches     int
	TotalPitches    int
}

// LegacyAwardAssignment represents one award given to a player-season.
type LegacyAwardAssignment struct {
	LegacyPlayerSeasonID int
	OriginalName         string
	AwardName            string // PlayerAwards.Name (used for custom awards)
	IsBuiltIn            bool
	Importance           int
	OmitFromGroupings    bool
	IsBattingAward       bool
	IsPitchingAward      bool
	IsFieldingAward      bool
	IsPlayoffAward       bool
	IsUserAssignable     bool
}

// LegacyScheduleGame represents one row from TeamSeasonSchedules.
type LegacyScheduleGame struct {
	LegacySeasonID  int
	HomeTeamHistID  int
	AwayTeamHistID  int
	HomePitcherPSID *int
	AwayPitcherPSID *int
	Day             int
	GlobalGameNum   int
	HomeScore       *int
	AwayScore       *int
}

// LegacyPlayoffGame represents one row from TeamPlayoffSchedules.
type LegacyPlayoffGame struct {
	LegacySeasonID  int
	HomeTeamHistID  int
	AwayTeamHistID  int
	HomePitcherPSID *int
	AwayPitcherPSID *int
	SeriesNumber    int
	GlobalGameNum   int
	HomeScore       *int
	AwayScore       *int
}

// --- Read methods ---

// ReadFranchises returns all franchises in the legacy database.
func (r *LegacyCompanionReader) ReadFranchises(ctx context.Context) ([]LegacyFranchise, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT Id, Name, IsSmb3 FROM Franchises ORDER BY Id ASC`)
	if err != nil {
		return nil, fmt.Errorf("querying legacy franchises: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []LegacyFranchise
	for rows.Next() {
		var f LegacyFranchise
		var isSmb3 int
		if err := rows.Scan(&f.ID, &f.Name, &isSmb3); err != nil {
			return nil, fmt.Errorf("scanning legacy franchise: %w", err)
		}
		f.IsSmb3 = isSmb3 != 0
		out = append(out, f)
	}
	return out, rows.Err()
}

// ReadSeasons returns all seasons for the given franchise, ordered by Number ASC.
func (r *LegacyCompanionReader) ReadSeasons(ctx context.Context, franchiseID int) ([]LegacySeason, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT Id, Number, NumGamesRegularSeason
		FROM Seasons
		WHERE FranchiseId = ?
		ORDER BY Number ASC
	`, franchiseID)
	if err != nil {
		return nil, fmt.Errorf("querying legacy seasons: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []LegacySeason
	for rows.Next() {
		var s LegacySeason
		if err := rows.Scan(&s.ID, &s.Number, &s.NumGamesRegularSeason); err != nil {
			return nil, fmt.Errorf("scanning legacy season: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// ReadTeams returns all teams for the franchise. Each team includes all historical
// GUIDs from TeamGameIdHistory, ordered by Id ASC (oldest first = primary GUID).
func (r *LegacyCompanionReader) ReadTeams(ctx context.Context, franchiseID int) ([]LegacyTeam, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.Id, tgih.GameId
		FROM Teams t
		JOIN TeamGameIdHistory tgih ON tgih.TeamId = t.Id
		WHERE t.FranchiseId = ?
		ORDER BY t.Id ASC, tgih.Id ASC
	`, franchiseID)
	if err != nil {
		return nil, fmt.Errorf("querying legacy teams: %w", err)
	}
	defer func() { _ = rows.Close() }()

	byID := make(map[int]*LegacyTeam)
	var order []int
	for rows.Next() {
		var teamID int
		var gameID string
		if err := rows.Scan(&teamID, &gameID); err != nil {
			return nil, fmt.Errorf("scanning legacy team: %w", err)
		}
		if _, exists := byID[teamID]; !exists {
			byID[teamID] = &LegacyTeam{ID: teamID}
			order = append(order, teamID)
		}
		byID[teamID].GameGUIDs = append(byID[teamID].GameGUIDs, gameID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]LegacyTeam, 0, len(order))
	for _, id := range order {
		out = append(out, *byID[id])
	}
	return out, nil
}

// ReadSeasonTeamHistory returns all SeasonTeamHistory rows for the franchise,
// with conference, division, and team name resolved.
func (r *LegacyCompanionReader) ReadSeasonTeamHistory(ctx context.Context, franchiseID int) ([]LegacySeasonTeamHistory, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT sth.Id, sth.SeasonId, sth.TeamId,
		       tnh.Name,
		       c.Name, d.Name,
		       sth.Budget, sth.Payroll,
		       sth.Wins, sth.Losses, sth.GamesBehind,
		       sth.RunsScored, sth.RunsAllowed,
		       sth.TotalPower, sth.TotalContact, sth.TotalSpeed, sth.TotalFielding,
		       sth.TotalArm, sth.TotalVelocity, sth.TotalJunk, sth.TotalAccuracy,
		       sth.PlayoffSeed, sth.PlayoffWins, sth.PlayoffLosses,
		       sth.PlayoffRunsScored, sth.PlayoffRunsAllowed
		FROM SeasonTeamHistory sth
		JOIN Seasons s ON s.Id = sth.SeasonId
		JOIN TeamNameHistory tnh ON tnh.Id = sth.TeamNameHistoryId
		JOIN Divisions d ON d.Id = sth.DivisionId
		JOIN Conferences c ON c.Id = d.ConferenceId
		WHERE s.FranchiseId = ?
		ORDER BY s.Number ASC, sth.Id ASC
	`, franchiseID)
	if err != nil {
		return nil, fmt.Errorf("querying legacy season team history: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []LegacySeasonTeamHistory
	for rows.Next() {
		var h LegacySeasonTeamHistory
		if err := rows.Scan(
			&h.ID, &h.SeasonID, &h.TeamID,
			&h.TeamName,
			&h.ConferenceName, &h.DivisionName,
			&h.Budget, &h.Payroll,
			&h.Wins, &h.Losses, &h.GamesBehind,
			&h.RunsScored, &h.RunsAllowed,
			&h.TotalPower, &h.TotalContact, &h.TotalSpeed, &h.TotalFielding,
			&h.TotalArm, &h.TotalVelocity, &h.TotalJunk, &h.TotalAccuracy,
			&h.PlayoffSeed, &h.PlayoffWins, &h.PlayoffLosses,
			&h.PlayoffRunsScored, &h.PlayoffRunsAllowed,
		); err != nil {
			return nil, fmt.Errorf("scanning legacy season team history: %w", err)
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// ReadPlayers returns all players for the franchise with lookup IDs resolved and GUIDs loaded.
func (r *LegacyCompanionReader) ReadPlayers(ctx context.Context, franchiseID int) ([]LegacyPlayer, error) {
	// Step 1: player identity rows (one row per player)
	pRows, err := r.db.QueryContext(ctx, `
		SELECT Id, FirstName, LastName, IsHallOfFamer,
		       BatHandednessId, ThrowHandednessId, PrimaryPositionId,
		       PitcherRoleId, ChemistryId
		FROM Players
		WHERE FranchiseId = ?
		ORDER BY Id ASC
	`, franchiseID)
	if err != nil {
		return nil, fmt.Errorf("querying legacy players: %w", err)
	}
	defer func() { _ = pRows.Close() }()

	byID := make(map[int]*LegacyPlayer)
	var order []int
	for pRows.Next() {
		var p LegacyPlayer
		var isHoF, batHandID, throwHandID, primaryPosID int
		var pitcherRoleID, chemistryID sql.NullInt64
		if err := pRows.Scan(
			&p.ID, &p.FirstName, &p.LastName, &isHoF,
			&batHandID, &throwHandID, &primaryPosID,
			&pitcherRoleID, &chemistryID,
		); err != nil {
			return nil, fmt.Errorf("scanning legacy player: %w", err)
		}
		p.IsHallOfFamer = isHoF != 0
		p.BatHand = r.batHandByID[batHandID]
		p.ThrowHand = r.throwHandByID[throwHandID]
		p.PrimaryPosition = r.positionByID[primaryPosID]
		if pitcherRoleID.Valid {
			p.PitcherRole = r.pitcherRoleByID[int(pitcherRoleID.Int64)]
		}
		if chemistryID.Valid {
			p.ChemistryType = r.chemistryByID[int(chemistryID.Int64)]
		}
		byID[p.ID] = &p
		order = append(order, p.ID)
	}
	if err := pRows.Err(); err != nil {
		return nil, err
	}

	// Step 2: load GUIDs (one row per GUID, multiple per player on fork)
	gRows, err := r.db.QueryContext(ctx, `
		SELECT pgih.PlayerId, pgih.GameId
		FROM PlayerGameIdHistory pgih
		JOIN Players p ON p.Id = pgih.PlayerId
		WHERE p.FranchiseId = ?
		ORDER BY pgih.PlayerId ASC, pgih.Id ASC
	`, franchiseID)
	if err != nil {
		return nil, fmt.Errorf("querying legacy player GUIDs: %w", err)
	}
	defer func() { _ = gRows.Close() }()
	for gRows.Next() {
		var playerID int
		var gameID string
		if err := gRows.Scan(&playerID, &gameID); err != nil {
			return nil, fmt.Errorf("scanning player GUID: %w", err)
		}
		if p, ok := byID[playerID]; ok {
			p.GameGUIDs = append(p.GameGUIDs, gameID)
		}
	}
	if err := gRows.Err(); err != nil {
		return nil, err
	}

	out := make([]LegacyPlayer, 0, len(order))
	for _, id := range order {
		out = append(out, *byID[id])
	}
	return out, nil
}

// ReadPlayerSeasons returns all PlayerSeason rows for the franchise with secondary
// position resolved. Team associations are returned separately by ReadPlayerSeasonTeams.
func (r *LegacyCompanionReader) ReadPlayerSeasons(ctx context.Context, franchiseID int) ([]LegacyPlayerSeason, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ps.Id, ps.PlayerId, ps.SeasonId, ps.Age, ps.Salary,
		       ps.SecondaryPositionId
		FROM PlayerSeasons ps
		JOIN Players p ON p.Id = ps.PlayerId
		WHERE p.FranchiseId = ?
		ORDER BY ps.SeasonId ASC, ps.PlayerId ASC
	`, franchiseID)
	if err != nil {
		return nil, fmt.Errorf("querying legacy player seasons: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []LegacyPlayerSeason
	for rows.Next() {
		var ps LegacyPlayerSeason
		var secPosID sql.NullInt64
		if err := rows.Scan(
			&ps.ID, &ps.PlayerID, &ps.SeasonID, &ps.Age, &ps.Salary,
			&secPosID,
		); err != nil {
			return nil, fmt.Errorf("scanning legacy player season: %w", err)
		}
		if secPosID.Valid {
			ps.SecondaryPosition = r.positionByID[int(secPosID.Int64)]
		}
		out = append(out, ps)
	}
	return out, rows.Err()
}

// ReadPlayerSeasonTeams returns all PlayerTeamHistory rows for the franchise, keyed by
// legacy player season ID. All Order values (1, 2, 3) are included; SortOrder = Order-1.
func (r *LegacyCompanionReader) ReadPlayerSeasonTeams(ctx context.Context, franchiseID int) (map[int][]LegacyPlayerSeasonTeam, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT pth.PlayerSeasonId, pth.SeasonTeamHistoryId, pth."Order"
		FROM PlayerTeamHistory pth
		JOIN PlayerSeasons ps ON ps.Id = pth.PlayerSeasonId
		JOIN Players p ON p.Id = ps.PlayerId
		WHERE p.FranchiseId = ?
		ORDER BY pth.PlayerSeasonId ASC, pth."Order" ASC
	`, franchiseID)
	if err != nil {
		return nil, fmt.Errorf("querying legacy player season teams: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := make(map[int][]LegacyPlayerSeasonTeam)
	for rows.Next() {
		var psID int
		var teamHistID sql.NullInt64
		var order int
		if err := rows.Scan(&psID, &teamHistID, &order); err != nil {
			return nil, fmt.Errorf("scanning legacy player season team: %w", err)
		}
		if !teamHistID.Valid {
			continue
		}
		t := LegacyPlayerSeasonTeam{
			PlayerSeasonID: psID,
			TeamHistID:     int(teamHistID.Int64),
			SortOrder:      order - 1,
		}
		out[psID] = append(out[psID], t)
	}
	return out, rows.Err()
}

// ReadGameStats returns all PlayerSeasonGameStats rows for the franchise.
// Nullable Arm/Velocity/Junk/Accuracy are coalesced to 0.
func (r *LegacyCompanionReader) ReadGameStats(ctx context.Context, franchiseID int) ([]LegacyGameStats, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT psg.PlayerSeasonId,
		       psg.Power, psg.Contact, psg.Speed, psg.Fielding,
		       COALESCE(psg.Arm, 0),
		       COALESCE(psg.Velocity, 0),
		       COALESCE(psg.Junk, 0),
		       COALESCE(psg.Accuracy, 0)
		FROM PlayerSeasonGameStats psg
		JOIN PlayerSeasons ps ON ps.Id = psg.PlayerSeasonId
		JOIN Players p ON p.Id = ps.PlayerId
		WHERE p.FranchiseId = ?
	`, franchiseID)
	if err != nil {
		return nil, fmt.Errorf("querying legacy game stats: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []LegacyGameStats
	for rows.Next() {
		var gs LegacyGameStats
		if err := rows.Scan(
			&gs.PlayerSeasonID,
			&gs.Power, &gs.Contact, &gs.Speed, &gs.Fielding,
			&gs.Arm, &gs.Velocity, &gs.Junk, &gs.Accuracy,
		); err != nil {
			return nil, fmt.Errorf("scanning legacy game stats: %w", err)
		}
		out = append(out, gs)
	}
	return out, rows.Err()
}

// ReadBattingStats returns all PlayerSeasonBattingStats (counting columns only) for the franchise.
func (r *LegacyCompanionReader) ReadBattingStats(ctx context.Context, franchiseID int) ([]LegacyBattingStat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT psb.PlayerSeasonId, psb.IsRegularSeason,
		       psb.GamesPlayed, psb.GamesBatting, psb.AtBats, psb.Runs, psb.Hits,
		       psb.Doubles, psb.Triples, psb.HomeRuns, psb.RunsBattedIn,
		       psb.StolenBases, psb.CaughtStealing, psb.Walks, psb.Strikeouts,
		       psb.HitByPitch, psb.SacrificeHits, psb.SacrificeFlies,
		       psb.Errors, psb.PassedBalls
		FROM PlayerSeasonBattingStats psb
		JOIN PlayerSeasons ps ON ps.Id = psb.PlayerSeasonId
		JOIN Players p ON p.Id = ps.PlayerId
		WHERE p.FranchiseId = ?
	`, franchiseID)
	if err != nil {
		return nil, fmt.Errorf("querying legacy batting stats: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []LegacyBattingStat
	for rows.Next() {
		var bs LegacyBattingStat
		var isReg int
		if err := rows.Scan(
			&bs.PlayerSeasonID, &isReg,
			&bs.GamesPlayed, &bs.GamesBatting, &bs.AtBats, &bs.Runs, &bs.Hits,
			&bs.Doubles, &bs.Triples, &bs.HomeRuns, &bs.RunsBattedIn,
			&bs.StolenBases, &bs.CaughtStealing, &bs.Walks, &bs.Strikeouts,
			&bs.HitByPitch, &bs.SacrificeHits, &bs.SacrificeFlies,
			&bs.Errors, &bs.PassedBalls,
		); err != nil {
			return nil, fmt.Errorf("scanning legacy batting stat: %w", err)
		}
		bs.IsRegularSeason = isReg != 0
		out = append(out, bs)
	}
	return out, rows.Err()
}

// ReadPitchingStats returns all PlayerSeasonPitchingStats (counting columns only) for the franchise.
func (r *LegacyCompanionReader) ReadPitchingStats(ctx context.Context, franchiseID int) ([]LegacyPitchingStat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT psp.PlayerSeasonId, psp.IsRegularSeason,
		       psp.Wins, psp.Losses, psp.GamesPlayed, psp.GamesStarted,
		       psp.CompleteGames, psp.Shutouts, psp.Saves, psp.InningsPitched,
		       psp.Hits, psp.EarnedRuns, psp.HomeRuns, psp.Walks, psp.Strikeouts,
		       psp.HitByPitch, psp.BattersFaced, psp.GamesFinished,
		       psp.RunsAllowed, psp.WildPitches, psp.TotalPitches
		FROM PlayerSeasonPitchingStats psp
		JOIN PlayerSeasons ps ON ps.Id = psp.PlayerSeasonId
		JOIN Players p ON p.Id = ps.PlayerId
		WHERE p.FranchiseId = ?
	`, franchiseID)
	if err != nil {
		return nil, fmt.Errorf("querying legacy pitching stats: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []LegacyPitchingStat
	for rows.Next() {
		var ps LegacyPitchingStat
		var isReg int
		if err := rows.Scan(
			&ps.PlayerSeasonID, &isReg,
			&ps.Wins, &ps.Losses, &ps.GamesPlayed, &ps.GamesStarted,
			&ps.CompleteGames, &ps.Shutouts, &ps.Saves, &ps.InningsPitched,
			&ps.HitsAllowed, &ps.EarnedRuns, &ps.HomeRunsAllowed, &ps.Walks, &ps.Strikeouts,
			&ps.HitBatters, &ps.BattersFaced, &ps.GamesFinished,
			&ps.RunsAllowed, &ps.WildPitches, &ps.TotalPitches,
		); err != nil {
			return nil, fmt.Errorf("scanning legacy pitching stat: %w", err)
		}
		ps.IsRegularSeason = isReg != 0
		out = append(out, ps)
	}
	return out, rows.Err()
}

// ReadTraits returns a map of legacy PlayerSeason ID → slice of trait names.
func (r *LegacyCompanionReader) ReadTraits(ctx context.Context, franchiseID int) (map[int][]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT pst.PlayerSeasonsId, tr.Name
		FROM PlayerSeasonTrait pst
		JOIN Traits tr ON tr.Id = pst.TraitsId
		JOIN PlayerSeasons ps ON ps.Id = pst.PlayerSeasonsId
		JOIN Players p ON p.Id = ps.PlayerId
		WHERE p.FranchiseId = ?
		ORDER BY pst.PlayerSeasonsId ASC
	`, franchiseID)
	if err != nil {
		return nil, fmt.Errorf("querying legacy traits: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := make(map[int][]string)
	for rows.Next() {
		var psID int
		var name string
		if err := rows.Scan(&psID, &name); err != nil {
			return nil, fmt.Errorf("scanning legacy trait: %w", err)
		}
		out[psID] = append(out[psID], name)
	}
	return out, rows.Err()
}

// ReadPitches returns a map of legacy PlayerSeason ID → slice of pitch type names.
func (r *LegacyCompanionReader) ReadPitches(ctx context.Context, franchiseID int) (map[int][]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ptps.PlayerSeasonsId, pt.Name
		FROM PitchTypePlayerSeason ptps
		JOIN PitchTypes pt ON pt.Id = ptps.PitchTypesId
		JOIN PlayerSeasons ps ON ps.Id = ptps.PlayerSeasonsId
		JOIN Players p ON p.Id = ps.PlayerId
		WHERE p.FranchiseId = ?
		ORDER BY ptps.PlayerSeasonsId ASC
	`, franchiseID)
	if err != nil {
		return nil, fmt.Errorf("querying legacy pitches: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := make(map[int][]string)
	for rows.Next() {
		var psID int
		var name string
		if err := rows.Scan(&psID, &name); err != nil {
			return nil, fmt.Errorf("scanning legacy pitch type: %w", err)
		}
		out[psID] = append(out[psID], name)
	}
	return out, rows.Err()
}

// ReadAwardAssignments returns all award assignments for player-seasons in the franchise.
func (r *LegacyCompanionReader) ReadAwardAssignments(ctx context.Context, franchiseID int) ([]LegacyAwardAssignment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT paps.PlayerSeasonsId,
		       pa.OriginalName, pa.Name, pa.IsBuiltIn, pa.Importance,
		       pa.OmitFromGroupings,
		       pa.IsBattingAward, pa.IsPitchingAward, pa.IsFieldingAward,
		       pa.IsPlayoffAward, pa.IsUserAssignable
		FROM PlayerAwardPlayerSeason paps
		JOIN PlayerAwards pa ON pa.Id = paps.AwardsId
		JOIN PlayerSeasons ps ON ps.Id = paps.PlayerSeasonsId
		JOIN Players p ON p.Id = ps.PlayerId
		WHERE p.FranchiseId = ?
	`, franchiseID)
	if err != nil {
		return nil, fmt.Errorf("querying legacy award assignments: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []LegacyAwardAssignment
	for rows.Next() {
		var a LegacyAwardAssignment
		var isBuiltIn, omitFromGroupings, isBatting, isPitching, isFielding, isPlayoff, isUserAssignable int
		if err := rows.Scan(
			&a.LegacyPlayerSeasonID,
			&a.OriginalName, &a.AwardName, &isBuiltIn, &a.Importance,
			&omitFromGroupings,
			&isBatting, &isPitching, &isFielding, &isPlayoff, &isUserAssignable,
		); err != nil {
			return nil, fmt.Errorf("scanning legacy award assignment: %w", err)
		}
		a.IsBuiltIn = isBuiltIn != 0
		a.OmitFromGroupings = omitFromGroupings != 0
		a.IsBattingAward = isBatting != 0
		a.IsPitchingAward = isPitching != 0
		a.IsFieldingAward = isFielding != 0
		a.IsPlayoffAward = isPlayoff != 0
		a.IsUserAssignable = isUserAssignable != 0
		out = append(out, a)
	}
	return out, rows.Err()
}

// ReadSeasonSchedules returns all regular season schedule rows for the franchise.
func (r *LegacyCompanionReader) ReadSeasonSchedules(ctx context.Context, franchiseID int) ([]LegacyScheduleGame, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT sth.SeasonId,
		       tss.HomeTeamHistoryId, tss.AwayTeamHistoryId,
		       tss.HomePitcherSeasonId, tss.AwayPitcherSeasonId,
		       tss.Day, tss.GlobalGameNumber,
		       tss.HomeScore, tss.AwayScore
		FROM TeamSeasonSchedules tss
		JOIN SeasonTeamHistory sth ON sth.Id = tss.HomeTeamHistoryId
		JOIN Seasons s ON s.Id = sth.SeasonId
		WHERE s.FranchiseId = ?
		ORDER BY sth.SeasonId ASC, tss.GlobalGameNumber ASC
	`, franchiseID)
	if err != nil {
		return nil, fmt.Errorf("querying legacy season schedules: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []LegacyScheduleGame
	for rows.Next() {
		var g LegacyScheduleGame
		var homePitcherID, awayPitcherID sql.NullInt64
		if err := rows.Scan(
			&g.LegacySeasonID,
			&g.HomeTeamHistID, &g.AwayTeamHistID,
			&homePitcherID, &awayPitcherID,
			&g.Day, &g.GlobalGameNum,
			&g.HomeScore, &g.AwayScore,
		); err != nil {
			return nil, fmt.Errorf("scanning legacy schedule game: %w", err)
		}
		if homePitcherID.Valid {
			v := int(homePitcherID.Int64)
			g.HomePitcherPSID = &v
		}
		if awayPitcherID.Valid {
			v := int(awayPitcherID.Int64)
			g.AwayPitcherPSID = &v
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

// ReadPlayoffSchedules returns all playoff schedule rows for the franchise.
func (r *LegacyCompanionReader) ReadPlayoffSchedules(ctx context.Context, franchiseID int) ([]LegacyPlayoffGame, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT sth.SeasonId,
		       tps.HomeTeamHistoryId, tps.AwayTeamHistoryId,
		       tps.HomePitcherSeasonId, tps.AwayPitcherSeasonId,
		       tps.SeriesNumber, tps.GlobalGameNumber,
		       tps.HomeScore, tps.AwayScore
		FROM TeamPlayoffSchedules tps
		JOIN SeasonTeamHistory sth ON sth.Id = tps.HomeTeamHistoryId
		JOIN Seasons s ON s.Id = sth.SeasonId
		WHERE s.FranchiseId = ?
		ORDER BY sth.SeasonId ASC, tps.SeriesNumber ASC, tps.GlobalGameNumber ASC
	`, franchiseID)
	if err != nil {
		return nil, fmt.Errorf("querying legacy playoff schedules: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []LegacyPlayoffGame
	for rows.Next() {
		var g LegacyPlayoffGame
		var homePitcherID, awayPitcherID sql.NullInt64
		if err := rows.Scan(
			&g.LegacySeasonID,
			&g.HomeTeamHistID, &g.AwayTeamHistID,
			&homePitcherID, &awayPitcherID,
			&g.SeriesNumber, &g.GlobalGameNum,
			&g.HomeScore, &g.AwayScore,
		); err != nil {
			return nil, fmt.Errorf("scanning legacy playoff game: %w", err)
		}
		if homePitcherID.Valid {
			v := int(homePitcherID.Int64)
			g.HomePitcherPSID = &v
		}
		if awayPitcherID.Valid {
			v := int(awayPitcherID.Int64)
			g.AwayPitcherPSID = &v
		}
		out = append(out, g)
	}
	return out, rows.Err()
}
