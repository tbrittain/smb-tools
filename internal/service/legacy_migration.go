package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
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

// Migrate imports all data for legacyFranchiseID from legacyDB into companionDB.
//
// leagueGUID is a caller-generated UUID written to seasons.league_guid and
// registered as a franchise_source entry (save_file_path = "(legacy migration)").
// The entire write is wrapped in a single transaction.
func (svc *LegacyMigrationService) Migrate(
	ctx context.Context,
	legacyDB *sql.DB,
	legacyFranchiseID int,
	companionDB *sql.DB,
	leagueGUID string,
) (MigrationResult, error) {
	reader, err := store.NewLegacyCompanionReader(ctx, legacyDB)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("creating legacy reader: %w", err)
	}

	// Read all data before opening the transaction so legacy reads don't block writes.
	seasons, err := reader.ReadSeasons(ctx, legacyFranchiseID)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("reading seasons: %w", err)
	}
	teams, err := reader.ReadTeams(ctx, legacyFranchiseID)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("reading teams: %w", err)
	}
	seasonTeamHistories, err := reader.ReadSeasonTeamHistory(ctx, legacyFranchiseID)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("reading season team history: %w", err)
	}
	players, err := reader.ReadPlayers(ctx, legacyFranchiseID)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("reading players: %w", err)
	}
	playerSeasons, err := reader.ReadPlayerSeasons(ctx, legacyFranchiseID)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("reading player seasons: %w", err)
	}
	gameStats, err := reader.ReadGameStats(ctx, legacyFranchiseID)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("reading game stats: %w", err)
	}
	battingStats, err := reader.ReadBattingStats(ctx, legacyFranchiseID)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("reading batting stats: %w", err)
	}
	pitchingStats, err := reader.ReadPitchingStats(ctx, legacyFranchiseID)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("reading pitching stats: %w", err)
	}
	traitsByPSID, err := reader.ReadTraits(ctx, legacyFranchiseID)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("reading traits: %w", err)
	}
	pitchesByPSID, err := reader.ReadPitches(ctx, legacyFranchiseID)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("reading pitches: %w", err)
	}
	awardAssignments, err := reader.ReadAwardAssignments(ctx, legacyFranchiseID)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("reading award assignments: %w", err)
	}
	seasonSchedules, err := reader.ReadSeasonSchedules(ctx, legacyFranchiseID)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("reading season schedules: %w", err)
	}
	playoffSchedules, err := reader.ReadPlayoffSchedules(ctx, legacyFranchiseID)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("reading playoff schedules: %w", err)
	}

	// Check for logo data (not migrated; reported in result).
	logosSkipped, err := svc.countLogos(ctx, legacyDB, legacyFranchiseID)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("checking logos: %w", err)
	}

	// Resolve award IDs outside the main transaction. Custom awards are inserted
	// here; built-ins are matched by original_name.
	awardIDByOriginalName, err := svc.resolveAwardIDs(ctx, awardAssignments, companionDB)
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

	result, err := svc.migrateInTx(ctx, tx, leagueGUID,
		seasons, teams, seasonTeamHistories,
		players, playerSeasons, gameStats, battingStats, pitchingStats,
		traitsByPSID, pitchesByPSID, awardAssignments, awardIDByOriginalName,
		seasonSchedules, playoffSchedules,
	)
	if err != nil {
		return MigrationResult{}, err
	}
	result.LogosSkipped = logosSkipped

	if err = tx.Commit(); err != nil {
		return MigrationResult{}, fmt.Errorf("committing migration transaction: %w", err)
	}
	return result, nil
}

func (svc *LegacyMigrationService) migrateInTx(
	ctx context.Context,
	tx *sql.Tx,
	leagueGUID string,
	seasons []store.LegacySeason,
	teams []store.LegacyTeam,
	seasonTeamHistories []store.LegacySeasonTeamHistory,
	players []store.LegacyPlayer,
	playerSeasons []store.LegacyPlayerSeason,
	gameStats []store.LegacyGameStats,
	battingStats []store.LegacyBattingStat,
	pitchingStats []store.LegacyPitchingStat,
	traitsByPSID map[int][]string,
	pitchesByPSID map[int][]string,
	awardAssignments []store.LegacyAwardAssignment,
	awardIDByOriginalName map[string]int64,
	seasonSchedules []store.LegacyScheduleGame,
	playoffSchedules []store.LegacyPlayoffGame,
) (MigrationResult, error) {
	seasonStore := store.NewSeasonStore(tx)
	teamStore := store.NewTeamHistoryStore(tx)
	playerStore := store.NewPlayerSeasonStore(tx)
	scheduleStore := store.NewScheduleStore(tx)

	result := MigrationResult{}

	// ID maps: legacy ID → new companion DB ID
	// (needed to remap foreign keys in schedule rows)
	legacySeasonIDToNew := make(map[int]int64, len(seasons))
	legacySTHIDToNew := make(map[int]int64, len(seasonTeamHistories))
	legacyPSIDToNew := make(map[int]int64, len(playerSeasons))

	// ── 1. Seasons ──────────────────────────────────────────────────────────────
	for _, s := range seasons {
		newID, err := seasonStore.Upsert(ctx, store.Season{
			LeagueGUID:       leagueGUID,
			SaveGameSeasonID: s.ID,
			SeasonNum:        s.Number,
			NumGames:         s.NumGamesRegularSeason,
		})
		if err != nil {
			return result, fmt.Errorf("upserting season %d: %w", s.ID, err)
		}
		legacySeasonIDToNew[s.ID] = newID
		result.SeasonsMigrated++
	}

	// ── 2. Teams ────────────────────────────────────────────────────────────────
	legacyTeamIDToNew := make(map[int]int64, len(teams))
	for _, t := range teams {
		if len(t.GameGUIDs) == 0 {
			continue
		}
		teamID, err := teamStore.UpsertTeam(ctx, t.GameGUIDs[0], "")
		if err != nil {
			return result, fmt.Errorf("upserting team %d: %w", t.ID, err)
		}
		for _, altGUID := range t.GameGUIDs[1:] {
			if _, execErr := tx.ExecContext(ctx,
				`INSERT OR IGNORE INTO team_alt_guids (team_id, game_guid) VALUES (?, ?)`,
				teamID, altGUID,
			); execErr != nil {
				return result, fmt.Errorf("inserting alt GUID for team %d: %w", t.ID, execErr)
			}
		}
		legacyTeamIDToNew[t.ID] = teamID
		result.TeamsMigrated++
	}

	// ── 3. Season team history ───────────────────────────────────────────────────
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
			TeamID:         newTeamID,
			SeasonID:       newSeasonID,
			TeamName:       h.TeamName,
			ConferenceName: h.ConferenceName,
			DivisionName:   h.DivisionName,
			Budget:         h.Budget,
			Payroll:        h.Payroll,
			Wins:           h.Wins,
			Losses:         h.Losses,
			GamesBack:      h.GamesBehind,
			RunsFor:        h.RunsScored,
			RunsAgainst:    h.RunsAllowed,
			TotalPower:     h.TotalPower,
			TotalContact:   h.TotalContact,
			TotalSpeed:     h.TotalSpeed,
			TotalFielding:  h.TotalFielding,
			TotalArm:       h.TotalArm,
			TotalVelocity:  h.TotalVelocity,
			TotalJunk:      h.TotalJunk,
			TotalAccuracy:  h.TotalAccuracy,
			PlayoffSeed:    h.PlayoffSeed,
			PlayoffWins:    h.PlayoffWins,
			PlayoffLosses:  h.PlayoffLosses,
			PlayoffRunsFor: h.PlayoffRunsScored,
			PlayoffRunsAgainst: h.PlayoffRunsAllowed,
		})
		if err != nil {
			return result, fmt.Errorf("upserting team history (legacy sth %d): %w", h.ID, err)
		}
		legacySTHIDToNew[h.ID] = newHistID
	}

	// ── 4. Players ──────────────────────────────────────────────────────────────
	legacyPlayerIDToNew := make(map[int]int64, len(players))
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
			return result, fmt.Errorf("upserting player %d: %w", p.ID, err)
		}
		if p.IsHallOfFamer {
			if _, execErr := tx.ExecContext(ctx,
				`UPDATE players SET is_hall_of_famer = 1 WHERE id = ?`, playerID,
			); execErr != nil {
				return result, fmt.Errorf("setting HoF for player %d: %w", p.ID, execErr)
			}
		}
		for _, altGUID := range p.GameGUIDs[1:] {
			if _, execErr := tx.ExecContext(ctx,
				`INSERT OR IGNORE INTO player_alt_guids (player_id, game_guid) VALUES (?, ?)`,
				playerID, altGUID,
			); execErr != nil {
				return result, fmt.Errorf("inserting alt GUID for player %d: %w", p.ID, execErr)
			}
		}
		legacyPlayerIDToNew[p.ID] = playerID
		result.PlayersMigrated++
	}

	// Build lookup for player attributes needed when upserting player seasons.
	legacyPlayerByID := make(map[int]store.LegacyPlayer, len(players))
	for _, p := range players {
		legacyPlayerByID[p.ID] = p
	}

	// ── 5. Player seasons ────────────────────────────────────────────────────────
	for _, ps := range playerSeasons {
		newPlayerID, ok := legacyPlayerIDToNew[ps.PlayerID]
		if !ok {
			continue
		}
		newSeasonID, ok := legacySeasonIDToNew[ps.SeasonID]
		if !ok {
			continue
		}
		var teamHistID *int64
		if ps.CurrentTeamHistID != nil {
			if newSTHID, ok := legacySTHIDToNew[*ps.CurrentTeamHistID]; ok {
				teamHistID = &newSTHID
			}
		}
		traits := joinStrings(traitsByPSID[ps.ID])
		pitches := joinStrings(pitchesByPSID[ps.ID])
		legacyPlayer := legacyPlayerByID[ps.PlayerID]
		newPSID, err := playerStore.UpsertSeason(ctx, store.PlayerSeason{
			PlayerID:          newPlayerID,
			SeasonID:          newSeasonID,
			TeamHistoryID:     teamHistID,
			Age:               ps.Age,
			Salary:            ps.Salary,
			PrimaryPosition:   legacyPlayer.PrimaryPosition,
			SecondaryPosition: ps.SecondaryPosition,
			PitcherRole:       legacyPlayer.PitcherRole,
			BatHand:           legacyPlayer.BatHand,
			ThrowHand:         legacyPlayer.ThrowHand,
			ChemistryType:     legacyPlayer.ChemistryType,
			TraitsJSON:        traits,
			PitchesJSON:       pitches,
		})
		if err != nil {
			return result, fmt.Errorf("upserting player season (legacy ps %d): %w", ps.ID, err)
		}
		legacyPSIDToNew[ps.ID] = newPSID
	}

	// ── 6. Game stats ────────────────────────────────────────────────────────────
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
			return result, fmt.Errorf("upserting game stats (legacy ps %d): %w", gs.PlayerSeasonID, err)
		}
	}

	// ── 7. Batting stats ─────────────────────────────────────────────────────────
	for _, bs := range battingStats {
		newPSID, ok := legacyPSIDToNew[bs.PlayerSeasonID]
		if !ok {
			continue
		}
		if err := playerStore.UpsertBattingStats(ctx, store.PlayerSeasonBattingStats{
			PlayerSeasonID:  newPSID,
			IsRegularSeason: bs.IsRegularSeason,
			GamesPlayed:     bs.GamesPlayed,
			GamesBatting:    bs.GamesBatting,
			AtBats:          bs.AtBats,
			Runs:            bs.Runs,
			Hits:            bs.Hits,
			Doubles:         bs.Doubles,
			Triples:         bs.Triples,
			HomeRuns:        bs.HomeRuns,
			RBI:             bs.RunsBattedIn,
			StolenBases:     bs.StolenBases,
			CaughtStealing:  bs.CaughtStealing,
			Walks:           bs.Walks,
			Strikeouts:      bs.Strikeouts,
			HitByPitch:      bs.HitByPitch,
			SacHits:         bs.SacrificeHits,
			SacFlies:        bs.SacrificeFlies,
			Errors:          bs.Errors,
			PassedBalls:     bs.PassedBalls,
		}); err != nil {
			return result, fmt.Errorf("upserting batting stats (legacy ps %d, reg=%v): %w",
				bs.PlayerSeasonID, bs.IsRegularSeason, err)
		}
	}

	// ── 8. Pitching stats ────────────────────────────────────────────────────────
	for _, ps := range pitchingStats {
		newPSID, ok := legacyPSIDToNew[ps.PlayerSeasonID]
		if !ok {
			continue
		}
		outs := ipToOuts(ps.InningsPitched)
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
		}); err != nil {
			return result, fmt.Errorf("upserting pitching stats (legacy ps %d, reg=%v): %w",
				ps.PlayerSeasonID, ps.IsRegularSeason, err)
		}
	}

	// ── 9. Awards ────────────────────────────────────────────────────────────────
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
			return result, fmt.Errorf("inserting award %q for ps %d: %w",
				a.OriginalName, a.LegacyPlayerSeasonID, execErr)
		}
		result.AwardsMigrated++
	}

	// ── 10. Regular season schedule ──────────────────────────────────────────────
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
			return result, fmt.Errorf("upserting schedule game %d (season %d): %w",
				g.GlobalGameNum, g.LegacySeasonID, err)
		}
	}

	// ── 11. Playoff schedule ─────────────────────────────────────────────────────
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
			return result, fmt.Errorf("upserting playoff game %d series %d (season %d): %w",
				g.GlobalGameNum, g.SeriesNumber, g.LegacySeasonID, err)
		}
	}

	// ── 12. Context stats (OPS+, ERA+, FIP, FIP-, smbWAR) ───────────────────────
	// ImportSeason computes these immediately after writing counting stats.
	// The legacy migration must do the same after all batting/pitching rows land.
	for _, newSeasonID := range legacySeasonIDToNew {
		if err := ApplyContextStats(ctx, tx, newSeasonID, true); err != nil {
			return result, fmt.Errorf("computing context stats for season %d (regular): %w", newSeasonID, err)
		}
		if err := ApplyContextStats(ctx, tx, newSeasonID, false); err != nil {
			return result, fmt.Errorf("computing context stats for season %d (playoffs): %w", newSeasonID, err)
		}
	}

	return result, nil
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
