package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"smb-tools/internal/models"
	"smb-tools/internal/store"
)

// LegacyMigrationService migrates a single franchise from a SmbExplorerCompanion
// database into a new smb-tools companion database.
//
// The source (legacy) DB is read-only. A single transaction on the companion DB
// wraps the entire migration — either all data lands or none does.
type LegacyMigrationService struct{}

// NewLegacyMigrationService creates a LegacyMigrationService.
func NewLegacyMigrationService() *LegacyMigrationService {
	return &LegacyMigrationService{}
}

// MigrationResult summarises the outcome of a franchise migration.
type MigrationResult struct {
	SeasonsMigrated  int
	TeamsMigrated    int
	PlayersMigrated  int
	AwardsMigrated   int
	LogosSkipped     int // non-zero when legacy had logo blobs
}

// legacyFranchiseData holds all data read from the legacy DB before the transaction opens.
type legacyFranchiseData struct {
	seasons             []store.LegacySeason
	teams               []store.LegacyTeam
	seasonTeamHistories []store.LegacySeasonTeamHistory
	players             []store.LegacyPlayer
	playerSeasons       []store.LegacyPlayerSeason
	seasonTeamsByPSID   map[int][]store.LegacyPlayerSeasonTeam
	gameStats           []store.LegacyGameStats
	battingStats        []store.LegacyBattingStat
	pitchingStats       []store.LegacyPitchingStat
	traitsByPSID        map[int][]string
	pitchesByPSID       map[int][]string
	awardAssignments    []store.LegacyAwardAssignment
	seasonSchedules     []store.LegacyScheduleGame
	playoffSchedules    []store.LegacyPlayoffGame
	championships       []store.LegacyChampionship
}

// readLegacyData reads all franchise data from the legacy DB in one sequential pass.
// Reads happen before the companion transaction opens so legacy reads don't block writes.
func readLegacyData(ctx context.Context, reader *store.LegacyCompanionReader, franchiseID int) (legacyFranchiseData, error) {
	var d legacyFranchiseData
	var err error
	if d.seasons, err = reader.ReadSeasons(ctx, franchiseID); err != nil {
		return legacyFranchiseData{}, fmt.Errorf("reading seasons: %w", err)
	}
	if d.teams, err = reader.ReadTeams(ctx, franchiseID); err != nil {
		return legacyFranchiseData{}, fmt.Errorf("reading teams: %w", err)
	}
	if d.seasonTeamHistories, err = reader.ReadSeasonTeamHistory(ctx, franchiseID); err != nil {
		return legacyFranchiseData{}, fmt.Errorf("reading season team history: %w", err)
	}
	if d.players, err = reader.ReadPlayers(ctx, franchiseID); err != nil {
		return legacyFranchiseData{}, fmt.Errorf("reading players: %w", err)
	}
	if d.playerSeasons, err = reader.ReadPlayerSeasons(ctx, franchiseID); err != nil {
		return legacyFranchiseData{}, fmt.Errorf("reading player seasons: %w", err)
	}
	if d.seasonTeamsByPSID, err = reader.ReadPlayerSeasonTeams(ctx, franchiseID); err != nil {
		return legacyFranchiseData{}, fmt.Errorf("reading player season teams: %w", err)
	}
	if d.gameStats, err = reader.ReadGameStats(ctx, franchiseID); err != nil {
		return legacyFranchiseData{}, fmt.Errorf("reading game stats: %w", err)
	}
	if d.battingStats, err = reader.ReadBattingStats(ctx, franchiseID); err != nil {
		return legacyFranchiseData{}, fmt.Errorf("reading batting stats: %w", err)
	}
	if d.pitchingStats, err = reader.ReadPitchingStats(ctx, franchiseID); err != nil {
		return legacyFranchiseData{}, fmt.Errorf("reading pitching stats: %w", err)
	}
	if d.traitsByPSID, err = reader.ReadTraits(ctx, franchiseID); err != nil {
		return legacyFranchiseData{}, fmt.Errorf("reading traits: %w", err)
	}
	if d.pitchesByPSID, err = reader.ReadPitches(ctx, franchiseID); err != nil {
		return legacyFranchiseData{}, fmt.Errorf("reading pitches: %w", err)
	}
	if d.awardAssignments, err = reader.ReadAwardAssignments(ctx, franchiseID); err != nil {
		return legacyFranchiseData{}, fmt.Errorf("reading award assignments: %w", err)
	}
	if d.seasonSchedules, err = reader.ReadSeasonSchedules(ctx, franchiseID); err != nil {
		return legacyFranchiseData{}, fmt.Errorf("reading season schedules: %w", err)
	}
	if d.playoffSchedules, err = reader.ReadPlayoffSchedules(ctx, franchiseID); err != nil {
		return legacyFranchiseData{}, fmt.Errorf("reading playoff schedules: %w", err)
	}
	if d.championships, err = reader.ReadChampionships(ctx, franchiseID); err != nil {
		return legacyFranchiseData{}, fmt.Errorf("reading championships: %w", err)
	}
	return d, nil
}

// Migrate imports all data for legacyFranchiseID from legacyDB into companionDB.
//
// leagueGUID is a caller-generated UUID written to seasons.league_guid and
// registered as a franchise_source entry (save_file_path = "(legacy migration)").
// The entire write is wrapped in a single transaction.
// inningsPerGame is the user-supplied innings-per-game for all migrated seasons
// (the legacy schema has no source data for this value). 0/unset defaults to 9.
func (svc *LegacyMigrationService) Migrate(
	ctx context.Context,
	legacyDB *sql.DB,
	legacyFranchiseID int,
	companionDB *sql.DB,
	leagueGUID string,
	inningsPerGame int,
) (MigrationResult, error) {
	if inningsPerGame <= 0 {
		inningsPerGame = 9
	}
	slog.Info("legacy migration: starting", "legacyFranchiseID", legacyFranchiseID, "inningsPerGame", inningsPerGame)
	reader, err := store.NewLegacyCompanionReader(ctx, legacyDB)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("creating legacy reader: %w", err)
	}

	data, err := readLegacyData(ctx, reader, legacyFranchiseID)
	if err != nil {
		slog.Error("legacy migration: reading data", "err", err)
		return MigrationResult{}, err
	}
	slog.Debug("legacy migration: data read",
		"seasons", len(data.seasons),
		"players", len(data.players),
	)

	logosSkipped, err := svc.countLogos(ctx, legacyDB, legacyFranchiseID)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("checking logos: %w", err)
	}

	// Resolve award IDs outside the main transaction. Custom awards are inserted
	// here; built-ins are matched by original_name.
	awardIDByOriginalName, err := svc.resolveAwardIDs(ctx, data.awardAssignments, companionDB)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("resolving award IDs: %w", err)
	}

	// ── Main transaction ────────────────────────────────────────────────────────
	tx, err := companionDB.BeginTx(ctx, nil)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("beginning migration transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result, err := svc.migrateInTx(ctx, tx, leagueGUID, inningsPerGame,
		data.seasons, data.teams, data.seasonTeamHistories,
		data.players, data.playerSeasons, data.seasonTeamsByPSID, data.gameStats, data.battingStats, data.pitchingStats,
		data.traitsByPSID, data.pitchesByPSID, data.awardAssignments, awardIDByOriginalName,
		data.seasonSchedules, data.playoffSchedules, data.championships,
	)
	if err != nil {
		return MigrationResult{}, err
	}
	result.LogosSkipped = logosSkipped

	if err = tx.Commit(); err != nil {
		return MigrationResult{}, fmt.Errorf("committing migration transaction: %w", err)
	}
	slog.Info("legacy migration: complete",
		"legacyFranchiseID", legacyFranchiseID,
		"seasons", result.SeasonsMigrated,
		"players", result.PlayersMigrated,
		"awards", result.AwardsMigrated,
	)
	return result, nil
}

func (svc *LegacyMigrationService) migrateInTx(
	ctx context.Context,
	tx *sql.Tx,
	leagueGUID string,
	inningsPerGame int,
	seasons []store.LegacySeason,
	teams []store.LegacyTeam,
	seasonTeamHistories []store.LegacySeasonTeamHistory,
	players []store.LegacyPlayer,
	playerSeasons []store.LegacyPlayerSeason,
	seasonTeamsByPSID map[int][]store.LegacyPlayerSeasonTeam,
	gameStats []store.LegacyGameStats,
	battingStats []store.LegacyBattingStat,
	pitchingStats []store.LegacyPitchingStat,
	traitsByPSID map[int][]string,
	pitchesByPSID map[int][]string,
	awardAssignments []store.LegacyAwardAssignment,
	awardIDByOriginalName map[string]int64,
	seasonSchedules []store.LegacyScheduleGame,
	playoffSchedules []store.LegacyPlayoffGame,
	championships []store.LegacyChampionship,
) (MigrationResult, error) {
	seasonStore := store.NewSeasonStore(tx)
	teamStore := store.NewTeamHistoryStore(tx)
	playerStore := store.NewPlayerSeasonStore(tx)
	scheduleStore := store.NewScheduleStore(tx)
	result := MigrationResult{}

	var err error
	var legacySeasonIDToNew map[int]int64
	var legacyTeamIDToNew map[int]int64
	var legacySTHIDToNew map[int]int64
	var legacyPlayerIDToNew map[int]int64
	var legacyPlayerIDs []int64
	var legacyPSIDToNew map[int]int64

	legacySeasonIDToNew, result.SeasonsMigrated, err = svc.migrateLegacySeasons(ctx, seasonStore, leagueGUID, inningsPerGame, seasons)
	if err != nil {
		return result, err
	}
	legacyTeamIDToNew, result.TeamsMigrated, err = svc.migrateLegacyTeams(ctx, tx, teamStore, teams)
	if err != nil {
		return result, err
	}
	legacySTHIDToNew, err = svc.migrateLegacySeasonTeamHistory(ctx, teamStore, seasonTeamHistories, legacySeasonIDToNew, legacyTeamIDToNew)
	if err != nil {
		return result, err
	}
	legacyPlayerIDToNew, legacyPlayerIDs, result.PlayersMigrated, err = svc.migrateLegacyPlayers(ctx, tx, playerStore, players)
	if err != nil {
		return result, err
	}
	legacyPSIDToNew, err = svc.migrateLegacyPlayerSeasons(ctx, playerStore, playerSeasons, players, traitsByPSID, pitchesByPSID, seasonTeamsByPSID, legacyPlayerIDToNew, legacySeasonIDToNew, legacySTHIDToNew)
	if err != nil {
		return result, err
	}
	if err := svc.migrateLegacyGameStats(ctx, playerStore, gameStats, legacyPSIDToNew); err != nil {
		return result, err
	}
	if err := migrateLegacyLeagueAvgAttributes(ctx, tx, legacySeasonIDToNew); err != nil {
		return result, err
	}
	if err := migrateLegacyPlayerAttributePercentiles(ctx, tx, legacySeasonIDToNew); err != nil {
		return result, err
	}
	if err := svc.migrateLegacyBattingStats(ctx, playerStore, battingStats, legacyPSIDToNew); err != nil {
		return result, err
	}
	if err := svc.migrateLegacyPitchingStats(ctx, playerStore, pitchingStats, legacyPSIDToNew); err != nil {
		return result, err
	}
	result.AwardsMigrated, err = svc.migrateLegacyAwards(ctx, tx, awardAssignments, awardIDByOriginalName, legacyPSIDToNew)
	if err != nil {
		return result, err
	}
	if err := svc.migrateLegacyRegularSchedule(ctx, scheduleStore, seasonSchedules, legacySeasonIDToNew, legacySTHIDToNew, legacyPSIDToNew); err != nil {
		return result, err
	}
	if err := svc.migrateLegacyPlayoffSchedule(ctx, scheduleStore, playoffSchedules, legacySeasonIDToNew, legacySTHIDToNew, legacyPSIDToNew); err != nil {
		return result, err
	}
	if err := migrateLegacyPlayoffConfig(ctx, seasonStore, playoffSchedules, legacySeasonIDToNew); err != nil {
		return result, err
	}
	if err := migrateLegacyChampionshipAwards(ctx, tx, championships, legacySeasonIDToNew, legacySTHIDToNew); err != nil {
		return result, err
	}
	if err := migrateLegacyContextStats(ctx, tx, legacySeasonIDToNew); err != nil {
		return result, err
	}
	if err := ApplyCareerStats(ctx, tx, legacyPlayerIDs); err != nil {
		return result, fmt.Errorf("computing career stats: %w", err)
	}
	return result, nil
}

func (svc *LegacyMigrationService) migrateLegacySeasons(
	ctx context.Context,
	seasonStore *store.SeasonStore,
	leagueGUID string,
	inningsPerGame int,
	seasons []store.LegacySeason,
) (legacySeasonIDToNew map[int]int64, count int, err error) {
	legacySeasonIDToNew = make(map[int]int64, len(seasons))
	for _, s := range seasons {
		newID, err := seasonStore.Upsert(ctx, store.Season{
			LeagueGUID:       leagueGUID,
			SaveGameSeasonID: s.ID,
			SeasonNum:        s.Number,
			NumGames:         s.NumGamesRegularSeason,
			InningsPerGame:   inningsPerGame,
		})
		if err != nil {
			return nil, 0, fmt.Errorf("upserting season %d: %w", s.ID, err)
		}
		legacySeasonIDToNew[s.ID] = newID
		count++
	}
	return legacySeasonIDToNew, count, nil
}

func (svc *LegacyMigrationService) migrateLegacyTeams(
	ctx context.Context,
	tx *sql.Tx,
	teamStore *store.TeamHistoryStore,
	teams []store.LegacyTeam,
) (legacyTeamIDToNew map[int]int64, count int, err error) {
	legacyTeamIDToNew = make(map[int]int64, len(teams))
	for _, t := range teams {
		if len(t.GameGUIDs) == 0 {
			continue
		}
		teamID, err := teamStore.UpsertTeam(ctx, t.GameGUIDs[0], "")
		if err != nil {
			return nil, 0, fmt.Errorf("upserting team %d: %w", t.ID, err)
		}
		for _, altGUID := range t.GameGUIDs[1:] {
			if _, execErr := tx.ExecContext(ctx,
				`INSERT OR IGNORE INTO team_alt_guids (team_id, game_guid) VALUES (?, ?)`,
				teamID, altGUID,
			); execErr != nil {
				return nil, 0, fmt.Errorf("inserting alt GUID for team %d: %w", t.ID, execErr)
			}
		}
		legacyTeamIDToNew[t.ID] = teamID
		count++
	}
	return legacyTeamIDToNew, count, nil
}

func (svc *LegacyMigrationService) migrateLegacySeasonTeamHistory(
	ctx context.Context,
	teamStore *store.TeamHistoryStore,
	seasonTeamHistories []store.LegacySeasonTeamHistory,
	legacySeasonIDToNew map[int]int64,
	legacyTeamIDToNew map[int]int64,
) (legacySTHIDToNew map[int]int64, err error) {
	legacySTHIDToNew = make(map[int]int64, len(seasonTeamHistories))
	for _, h := range seasonTeamHistories {
		newSeasonID, ok := legacySeasonIDToNew[h.SeasonID]
		if !ok {
			continue
		}
		newTeamID, ok := legacyTeamIDToNew[h.TeamID]
		if !ok {
			continue
		}
		newHistID, err := teamStore.UpsertSeasonHistory(ctx, store.TeamSeasonHistory{
			TeamID:             newTeamID,
			SeasonID:           newSeasonID,
			TeamName:           h.TeamName,
			ConferenceName:     h.ConferenceName,
			DivisionName:       h.DivisionName,
			Budget:             h.Budget,
			Payroll:            h.Payroll,
			Wins:               h.Wins,
			Losses:             h.Losses,
			GamesBack:          h.GamesBehind,
			RunsFor:            h.RunsScored,
			RunsAgainst:        h.RunsAllowed,
			TotalPower:         h.TotalPower,
			TotalContact:       h.TotalContact,
			TotalSpeed:         h.TotalSpeed,
			TotalFielding:      h.TotalFielding,
			TotalArm:           h.TotalArm,
			TotalVelocity:      h.TotalVelocity,
			TotalJunk:          h.TotalJunk,
			TotalAccuracy:      h.TotalAccuracy,
			PlayoffSeed:        h.PlayoffSeed,
			PlayoffWins:        h.PlayoffWins,
			PlayoffLosses:      h.PlayoffLosses,
			PlayoffRunsFor:     h.PlayoffRunsScored,
			PlayoffRunsAgainst: h.PlayoffRunsAllowed,
		})
		if err != nil {
			return nil, fmt.Errorf("upserting team history (legacy sth %d): %w", h.ID, err)
		}
		legacySTHIDToNew[h.ID] = newHistID
	}
	return legacySTHIDToNew, nil
}

func (svc *LegacyMigrationService) migrateLegacyPlayers(
	ctx context.Context,
	tx *sql.Tx,
	playerStore *store.PlayerSeasonStore,
	players []store.LegacyPlayer,
) (legacyPlayerIDToNew map[int]int64, playerIDs []int64, count int, err error) {
	legacyPlayerIDToNew = make(map[int]int64, len(players))
	playerIDs = make([]int64, 0, len(players))
	for _, p := range players {
		if len(p.GameGUIDs) == 0 {
			continue
		}
		playerID, err := playerStore.UpsertPlayer(ctx, store.PlayerIdentity{
			GameGUID:      p.GameGUIDs[0],
			FirstName:     p.FirstName,
			LastName:      p.LastName,
			BatHand:       p.BatHand,
			ThrowHand:     p.ThrowHand,
			ChemistryType: p.ChemistryType,
		})
		if err != nil {
			return nil, nil, 0, fmt.Errorf("upserting player %d: %w", p.ID, err)
		}
		if p.IsHallOfFamer {
			if _, execErr := tx.ExecContext(ctx,
				`UPDATE players SET is_hall_of_famer = 1 WHERE id = ?`, playerID,
			); execErr != nil {
				return nil, nil, 0, fmt.Errorf("setting HoF for player %d: %w", p.ID, execErr)
			}
		}
		for _, altGUID := range p.GameGUIDs[1:] {
			if _, execErr := tx.ExecContext(ctx,
				`INSERT OR IGNORE INTO player_alt_guids (player_id, game_guid) VALUES (?, ?)`,
				playerID, altGUID,
			); execErr != nil {
				return nil, nil, 0, fmt.Errorf("inserting alt GUID for player %d: %w", p.ID, execErr)
			}
		}
		legacyPlayerIDToNew[p.ID] = playerID
		playerIDs = append(playerIDs, playerID)
		count++
	}
	return legacyPlayerIDToNew, playerIDs, count, nil
}

func (svc *LegacyMigrationService) migrateLegacyPlayerSeasons(
	ctx context.Context,
	playerStore *store.PlayerSeasonStore,
	playerSeasons []store.LegacyPlayerSeason,
	players []store.LegacyPlayer,
	traitsByPSID map[int][]string,
	pitchesByPSID map[int][]string,
	seasonTeamsByPSID map[int][]store.LegacyPlayerSeasonTeam,
	legacyPlayerIDToNew map[int]int64,
	legacySeasonIDToNew map[int]int64,
	legacySTHIDToNew map[int]int64,
) (legacyPSIDToNew map[int]int64, err error) {
	legacyPlayerByID := make(map[int]store.LegacyPlayer, len(players))
	for _, p := range players {
		legacyPlayerByID[p.ID] = p
	}
	legacyPSIDToNew = make(map[int]int64, len(playerSeasons))
	for _, ps := range playerSeasons {
		newPlayerID, ok := legacyPlayerIDToNew[ps.PlayerID]
		if !ok {
			continue
		}
		newSeasonID, ok := legacySeasonIDToNew[ps.SeasonID]
		if !ok {
			continue
		}
		legacyPlayer := legacyPlayerByID[ps.PlayerID]
		newPSID, err := playerStore.UpsertSeason(ctx, store.PlayerSeason{
			PlayerID:          newPlayerID,
			SeasonID:          newSeasonID,
			Age:               ps.Age,
			Salary:            ps.Salary,
			PrimaryPosition:   legacyPlayer.PrimaryPosition,
			SecondaryPosition: ps.SecondaryPosition,
			PitcherRole:       legacyPlayer.PitcherRole,
			BatHand:           legacyPlayer.BatHand,
			ThrowHand:         legacyPlayer.ThrowHand,
			ChemistryType:     legacyPlayer.ChemistryType,
			TraitsJSON:        joinStrings(traitsByPSID[ps.ID]),
			PitchesJSON:       joinStrings(pitchesByPSID[ps.ID]),
		})
		if err != nil {
			return nil, fmt.Errorf("upserting player season (legacy ps %d): %w", ps.ID, err)
		}
		legacyPSIDToNew[ps.ID] = newPSID
		if err := migrateLegacyPSTeams(ctx, playerStore, newPSID, seasonTeamsByPSID[ps.ID], legacySTHIDToNew); err != nil {
			return nil, fmt.Errorf("migrating season teams for legacy ps %d: %w", ps.ID, err)
		}
	}
	return legacyPSIDToNew, nil
}

// migrateLegacyPSTeams writes team associations for one player season.
func migrateLegacyPSTeams(ctx context.Context, playerStore *store.PlayerSeasonStore, newPSID int64, legacyTeams []store.LegacyPlayerSeasonTeam, legacySTHIDToNew map[int]int64) error {
	if len(legacyTeams) == 0 {
		return nil
	}
	var seasonTeams []store.PlayerSeasonTeam
	for _, lt := range legacyTeams {
		if newSTHID, ok := legacySTHIDToNew[lt.TeamHistID]; ok {
			seasonTeams = append(seasonTeams, store.PlayerSeasonTeam{
				PlayerSeasonID: newPSID,
				TeamHistoryID:  newSTHID,
				SortOrder:      lt.SortOrder,
			})
		}
	}
	if len(seasonTeams) == 0 {
		return nil
	}
	return playerStore.ReplaceSeasonTeams(ctx, newPSID, seasonTeams)
}

func (svc *LegacyMigrationService) migrateLegacyGameStats(
	ctx context.Context,
	playerStore *store.PlayerSeasonStore,
	gameStats []store.LegacyGameStats,
	legacyPSIDToNew map[int]int64,
) error {
	for _, gs := range gameStats {
		newPSID, ok := legacyPSIDToNew[gs.PlayerSeasonID]
		if !ok {
			continue
		}
		if err := playerStore.UpsertGameStats(ctx, store.PlayerSeasonGameStats{
			PlayerSeasonID: newPSID,
			Power:          gs.Power,
			Contact:        gs.Contact,
			Speed:          gs.Speed,
			Fielding:       gs.Fielding,
			Arm:            gs.Arm,
			Velocity:       gs.Velocity,
			Junk:           gs.Junk,
			Accuracy:       gs.Accuracy,
		}); err != nil {
			return fmt.Errorf("upserting game stats (legacy ps %d): %w", gs.PlayerSeasonID, err)
		}
	}
	return nil
}

func (svc *LegacyMigrationService) migrateLegacyBattingStats(
	ctx context.Context,
	playerStore *store.PlayerSeasonStore,
	battingStats []store.LegacyBattingStat,
	legacyPSIDToNew map[int]int64,
) error {
	for _, bs := range battingStats {
		newPSID, ok := legacyPSIDToNew[bs.PlayerSeasonID]
		if !ok {
			continue
		}
		tmp := models.CareerBattingStats{
			AtBats: bs.AtBats, Hits: bs.Hits, Doubles: bs.Doubles, Triples: bs.Triples,
			HomeRuns: bs.HomeRuns, Walks: bs.Walks, HitByPitch: bs.HitByPitch,
			SacHits: bs.SacrificeHits, SacFlies: bs.SacrificeFlies, Strikeouts: bs.Strikeouts,
		}
		ComputeBattingRates(&tmp)
		if err := playerStore.UpsertBattingStats(ctx, store.PlayerSeasonBattingStats{
			PlayerSeasonID:   newPSID,
			IsRegularSeason:  bs.IsRegularSeason,
			GamesPlayed:      bs.GamesPlayed,
			GamesBatting:     bs.GamesBatting,
			AtBats:           bs.AtBats,
			PlateAppearances: bs.PlateAppearances,
			Runs:             bs.Runs,
			Hits:             bs.Hits,
			Doubles:          bs.Doubles,
			Triples:          bs.Triples,
			HomeRuns:         bs.HomeRuns,
			RBI:              bs.RunsBattedIn,
			StolenBases:      bs.StolenBases,
			CaughtStealing:   bs.CaughtStealing,
			Walks:            bs.Walks,
			Strikeouts:       bs.Strikeouts,
			HitByPitch:       bs.HitByPitch,
			SacHits:          bs.SacrificeHits,
			SacFlies:         bs.SacrificeFlies,
			Errors:           bs.Errors,
			PassedBalls:      bs.PassedBalls,
			BA:               tmp.BA,
			OBP:              tmp.OBP,
			SLG:              tmp.SLG,
			OPS:              tmp.OPS,
			ISO:              tmp.ISO,
			BABIP:            tmp.BABIP,
			KPct:             tmp.KPct,
			BBPct:            tmp.BBPct,
			ABPerHR:          tmp.ABPerHR,
		}); err != nil {
			return fmt.Errorf("upserting batting stats (legacy ps %d, reg=%v): %w",
				bs.PlayerSeasonID, bs.IsRegularSeason, err)
		}
	}
	return nil
}

func (svc *LegacyMigrationService) migrateLegacyPitchingStats(
	ctx context.Context,
	playerStore *store.PlayerSeasonStore,
	pitchingStats []store.LegacyPitchingStat,
	legacyPSIDToNew map[int]int64,
) error {
	for _, ps := range pitchingStats {
		newPSID, ok := legacyPSIDToNew[ps.PlayerSeasonID]
		if !ok {
			continue
		}
		outs := ipToOuts(ps.InningsPitched)
		ptmp := models.CareerPitchingStats{
			Wins: ps.Wins, Losses: ps.Losses, OutsPitched: int(outs),
			HitsAllowed: ps.HitsAllowed, EarnedRuns: ps.EarnedRuns,
			HomeRunsAllowed: ps.HomeRunsAllowed, Walks: ps.Walks,
			Strikeouts: ps.Strikeouts, HitBatters: ps.HitBatters,
			BattersFaced: ps.BattersFaced, TotalPitches: ps.TotalPitches,
		}
		ComputePitchingRates(&ptmp)
		if err := playerStore.UpsertPitchingStats(ctx, store.PlayerSeasonPitchingStats{
			PlayerSeasonID:  newPSID,
			IsRegularSeason: ps.IsRegularSeason,
			Wins:            ps.Wins,
			Losses:          ps.Losses,
			Games:           ps.GamesPlayed,
			GamesStarted:    ps.GamesStarted,
			CompleteGames:   ps.CompleteGames,
			Shutouts:        ps.Shutouts,
			Saves:           ps.Saves,
			OutsPitched:     int(outs),
			HitsAllowed:     ps.HitsAllowed,
			EarnedRuns:      ps.EarnedRuns,
			HomeRunsAllowed: ps.HomeRunsAllowed,
			Walks:           ps.Walks,
			Strikeouts:      ps.Strikeouts,
			HitBatters:      ps.HitBatters,
			BattersFaced:    ps.BattersFaced,
			GamesFinished:   ps.GamesFinished,
			RunsAllowed:     ps.RunsAllowed,
			WildPitches:     ps.WildPitches,
			TotalPitches:    ps.TotalPitches,
			ERA:             ptmp.ERA,
			WHIP:            ptmp.WHIP,
			K9:              ptmp.K9,
			BB9:             ptmp.BB9,
			H9:              ptmp.H9,
			HR9:             ptmp.HR9,
			KPerBB:          ptmp.KPerBB,
			KPct:            ptmp.KPct,
			WinPct:          ptmp.WinPct,
			PPerIP:          ptmp.PPerIP,
		}); err != nil {
			return fmt.Errorf("upserting pitching stats (legacy ps %d, reg=%v): %w",
				ps.PlayerSeasonID, ps.IsRegularSeason, err)
		}
	}
	return nil
}

func (svc *LegacyMigrationService) migrateLegacyAwards(
	ctx context.Context,
	tx *sql.Tx,
	awardAssignments []store.LegacyAwardAssignment,
	awardIDByOriginalName map[string]int64,
	legacyPSIDToNew map[int]int64,
) (count int, err error) {
	for _, a := range awardAssignments {
		newPSID, ok := legacyPSIDToNew[a.LegacyPlayerSeasonID]
		if !ok {
			continue
		}
		newAwardID, ok := awardIDByOriginalName[a.OriginalName]
		if !ok {
			continue
		}
		if _, execErr := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO player_season_awards (player_season_id, award_id) VALUES (?, ?)`,
			newPSID, newAwardID,
		); execErr != nil {
			return count, fmt.Errorf("inserting award %q for ps %d: %w",
				a.OriginalName, a.LegacyPlayerSeasonID, execErr)
		}
		count++
	}
	return count, nil
}

func (svc *LegacyMigrationService) migrateLegacyRegularSchedule(
	ctx context.Context,
	scheduleStore *store.ScheduleStore,
	seasonSchedules []store.LegacyScheduleGame,
	legacySeasonIDToNew map[int]int64,
	legacySTHIDToNew map[int]int64,
	legacyPSIDToNew map[int]int64,
) error {
	for _, g := range seasonSchedules {
		newSeasonID, ok := legacySeasonIDToNew[g.LegacySeasonID]
		if !ok {
			continue
		}
		homeHistID, ok := legacySTHIDToNew[g.HomeTeamHistID]
		if !ok {
			continue
		}
		awayHistID, ok := legacySTHIDToNew[g.AwayTeamHistID]
		if !ok {
			continue
		}
		var homePitcherID, awayPitcherID *int64
		if g.HomePitcherPSID != nil {
			if id, ok := legacyPSIDToNew[*g.HomePitcherPSID]; ok {
				homePitcherID = &id
			}
		}
		if g.AwayPitcherPSID != nil {
			if id, ok := legacyPSIDToNew[*g.AwayPitcherPSID]; ok {
				awayPitcherID = &id
			}
		}
		if err := scheduleStore.UpsertGame(ctx, store.ScheduleGame{
			SeasonID:            newSeasonID,
			GameNumber:          g.GlobalGameNum,
			Day:                 g.Day,
			HomeTeamHistoryID:   homeHistID,
			AwayTeamHistoryID:   awayHistID,
			HomePitcherSeasonID: homePitcherID,
			AwayPitcherSeasonID: awayPitcherID,
			HomeScore:           g.HomeScore,
			AwayScore:           g.AwayScore,
		}); err != nil {
			return fmt.Errorf("upserting schedule game %d (season %d): %w",
				g.GlobalGameNum, g.LegacySeasonID, err)
		}
	}
	return nil
}

func (svc *LegacyMigrationService) migrateLegacyPlayoffSchedule(
	ctx context.Context,
	scheduleStore *store.ScheduleStore,
	playoffSchedules []store.LegacyPlayoffGame,
	legacySeasonIDToNew map[int]int64,
	legacySTHIDToNew map[int]int64,
	legacyPSIDToNew map[int]int64,
) error {
	for _, g := range playoffSchedules {
		newSeasonID, ok := legacySeasonIDToNew[g.LegacySeasonID]
		if !ok {
			continue
		}
		homeHistID, ok := legacySTHIDToNew[g.HomeTeamHistID]
		if !ok {
			continue
		}
		awayHistID, ok := legacySTHIDToNew[g.AwayTeamHistID]
		if !ok {
			continue
		}
		var homePitcherID, awayPitcherID *int64
		if g.HomePitcherPSID != nil {
			if id, ok := legacyPSIDToNew[*g.HomePitcherPSID]; ok {
				homePitcherID = &id
			}
		}
		if g.AwayPitcherPSID != nil {
			if id, ok := legacyPSIDToNew[*g.AwayPitcherPSID]; ok {
				awayPitcherID = &id
			}
		}
		if err := scheduleStore.UpsertPlayoffGame(ctx, store.PlayoffGame{
			SeasonID:            newSeasonID,
			SeriesNumber:        g.SeriesNumber,
			GameNumber:          g.GlobalGameNum,
			HomeTeamHistoryID:   homeHistID,
			AwayTeamHistoryID:   awayHistID,
			HomePitcherSeasonID: homePitcherID,
			AwayPitcherSeasonID: awayPitcherID,
			HomeScore:           g.HomeScore,
			AwayScore:           g.AwayScore,
		}); err != nil {
			return fmt.Errorf("upserting playoff game %d series %d (season %d): %w",
				g.GlobalGameNum, g.SeriesNumber, g.LegacySeasonID, err)
		}
	}
	return nil
}

// migrateLegacyPlayoffConfig infers and persists playoff rounds/series-length for each
// season that has playoff game data. t_playoffs is unavailable in the legacy path.
func migrateLegacyPlayoffConfig(
	ctx context.Context,
	seasonStore *store.SeasonStore,
	playoffSchedules []store.LegacyPlayoffGame,
	legacySeasonIDToNew map[int]int64,
) error {
	seen := make(map[int]struct{})
	for _, g := range playoffSchedules {
		seen[g.LegacySeasonID] = struct{}{}
	}
	for legacySeasonID := range seen {
		newSeasonID, ok := legacySeasonIDToNew[legacySeasonID]
		if !ok {
			continue
		}
		if err := seasonStore.InferAndSetPlayoffConfig(ctx, newSeasonID); err != nil {
			return fmt.Errorf("inferring playoff config for legacy season %d: %w", legacySeasonID, err)
		}
	}
	return nil
}

// migrateLegacyChampionshipAwards assigns League Champion / Conference Champion awards.
// The legacy DB stores championship winners in ChampionshipWinners (not as player awards)
// and may contain unplayed game rows with NULL scores that would fail the season_champions
// view's completeness gate — we bypass it and use the IDs directly.
func migrateLegacyChampionshipAwards(
	ctx context.Context,
	tx *sql.Tx,
	championships []store.LegacyChampionship,
	legacySeasonIDToNew map[int]int64,
	legacySTHIDToNew map[int]int64,
) error {
	awardStore := store.NewAwardStore(tx)
	for _, champ := range championships {
		newSeasonID, ok := legacySeasonIDToNew[champ.LegacySeasonID]
		if !ok {
			continue
		}
		newChampHistID, ok := legacySTHIDToNew[champ.ChampionSTHID]
		if !ok {
			continue
		}
		var newRunnerUpHistID int64
		if champ.RunnerUpSTHID != 0 {
			newRunnerUpHistID = legacySTHIDToNew[champ.RunnerUpSTHID]
		}
		if err := awardStore.AssignChampionshipAwardsForTeams(ctx, tx, newSeasonID, newChampHistID, newRunnerUpHistID); err != nil {
			return fmt.Errorf("assigning championship awards for legacy season %d: %w", champ.LegacySeasonID, err)
		}
	}
	return nil
}

// migrateLegacyContextStats computes OPS+, ERA+, FIP, FIP-, and smbWAR for every
// migrated season. Must run after all batting and pitching rows are written.
func migrateLegacyContextStats(ctx context.Context, tx *sql.Tx, legacySeasonIDToNew map[int]int64) error {
	for _, newSeasonID := range legacySeasonIDToNew {
		if err := ApplyContextStats(ctx, tx, newSeasonID, true); err != nil {
			return fmt.Errorf("computing context stats for season %d (regular): %w", newSeasonID, err)
		}
		if err := ApplyContextStats(ctx, tx, newSeasonID, false); err != nil {
			return fmt.Errorf("computing context stats for season %d (playoffs): %w", newSeasonID, err)
		}
	}
	return nil
}

// migrateLegacyLeagueAvgAttributes computes and persists league-average
// attribute values for every migrated season. Must run after all game stats
// rows are written (i.e. after migrateLegacyGameStats).
func migrateLegacyLeagueAvgAttributes(ctx context.Context, tx *sql.Tx, legacySeasonIDToNew map[int]int64) error {
	for _, newSeasonID := range legacySeasonIDToNew {
		if err := ApplyLeagueAvgAttributes(ctx, tx, newSeasonID); err != nil {
			return fmt.Errorf("computing league attribute averages for season %d: %w", newSeasonID, err)
		}
	}
	return nil
}

// migrateLegacyPlayerAttributePercentiles computes and persists per-player
// attribute percentile ranks for every migrated season. Must run after
// migrateLegacyGameStats so the game stats rows are present.
func migrateLegacyPlayerAttributePercentiles(ctx context.Context, tx *sql.Tx, legacySeasonIDToNew map[int]int64) error {
	for _, newSeasonID := range legacySeasonIDToNew {
		if err := ApplyPlayerAttributePercentiles(ctx, tx, newSeasonID); err != nil {
			return fmt.Errorf("computing player attribute percentiles for season %d: %w", newSeasonID, err)
		}
	}
	return nil
}

// resolveAwardIDs builds a map of OriginalName → new awards.id.
// Custom (non-built-in) awards not already in the companion DB are inserted.
func (svc *LegacyMigrationService) resolveAwardIDs(
	ctx context.Context,
	assignments []store.LegacyAwardAssignment,
	companionDB *sql.DB,
) (map[string]int64, error) {
	// Collect unique awards from the assignments.
	type awardMeta struct {
		name              string
		originalName      string
		isBuiltIn         bool
		importance        int
		omitFromGroupings bool
		isBatting         bool
		isPitching        bool
		isFielding        bool
		isPlayoff         bool
		isUserAssignable  bool
	}
	unique := make(map[string]awardMeta)
	for _, a := range assignments {
		if _, exists := unique[a.OriginalName]; !exists {
			unique[a.OriginalName] = awardMeta{
				name:              a.AwardName,
				originalName:      a.OriginalName,
				isBuiltIn:         a.IsBuiltIn,
				importance:        a.Importance,
				omitFromGroupings: a.OmitFromGroupings,
				isBatting:         a.IsBattingAward,
				isPitching:        a.IsPitchingAward,
				isFielding:        a.IsFieldingAward,
				isPlayoff:         a.IsPlayoffAward,
				isUserAssignable:  a.IsUserAssignable,
			}
		}
	}

	result := make(map[string]int64, len(unique))
	for originalName, meta := range unique {
		var id int64
		err := companionDB.QueryRowContext(ctx,
			`SELECT id FROM awards WHERE original_name = ?`, originalName,
		).Scan(&id)
		if err == nil {
			result[originalName] = id
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("looking up award %q: %w", originalName, err)
		}
		// Not found — insert as custom award (is_built_in=0).
		res, insertErr := companionDB.ExecContext(ctx, `
			INSERT OR IGNORE INTO awards
			    (name, original_name, importance, omit_from_groupings,
			     is_batting_award, is_pitching_award, is_fielding_award,
			     is_playoff_award, is_user_assignable, is_built_in)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
		`,
			meta.name, originalName, meta.importance, legacyBoolToInt(meta.omitFromGroupings),
			legacyBoolToInt(meta.isBatting), legacyBoolToInt(meta.isPitching), legacyBoolToInt(meta.isFielding),
			legacyBoolToInt(meta.isPlayoff), legacyBoolToInt(meta.isUserAssignable),
		)
		if insertErr != nil {
			return nil, fmt.Errorf("inserting custom award %q: %w", originalName, insertErr)
		}
		rows, _ := res.RowsAffected()
		if rows > 0 {
			id, _ = res.LastInsertId()
		} else {
			// Name conflict with an existing award — look up by name.
			if lookupErr := companionDB.QueryRowContext(ctx,
				`SELECT id FROM awards WHERE name = ?`, meta.name,
			).Scan(&id); lookupErr != nil {
				return nil, fmt.Errorf("looking up award by name %q after conflict: %w",
					meta.name, lookupErr)
			}
		}
		result[originalName] = id
	}
	return result, nil
}

// countLogos returns the number of TeamLogoHistory rows linked to teams in the franchise.
func (svc *LegacyMigrationService) countLogos(ctx context.Context, legacyDB *sql.DB, franchiseID int) (int, error) {
	var n int
	err := legacyDB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM TeamLogoHistory tlh
		JOIN TeamNameHistory tnh ON tnh.TeamLogoHistoryId = tlh.Id
		JOIN SeasonTeamHistory sth ON sth.TeamNameHistoryId = tnh.Id
		JOIN Seasons s ON s.Id = sth.SeasonId
		WHERE s.FranchiseId = ?
	`, franchiseID).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("counting logos: %w", err)
	}
	return n, nil
}

// ipToOuts converts a legacy InningsPitched value (REAL) to integer outs.
// The decimal digit represents additional outs (0, 1, or 2), not fractions.
// e.g. 6.2 = 6*3 + 2 = 20 outs.  NULL → 0.
func ipToOuts(ip *float64) int64 {
	if ip == nil {
		return 0
	}
	whole := math.Floor(*ip)
	frac := *ip - whole
	return int64(whole)*3 + int64(math.Round(frac*10))
}

func legacyBoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
