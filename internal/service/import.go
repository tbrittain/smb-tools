package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"smb-tools/internal/models"
	"smb-tools/internal/store"
)

// ImportService orchestrates importing a single franchise season from the save
// game into the companion database. It reads via SaveGameReader and writes via
// the store layer, all within a single transaction for atomicity.
//
// Importing the same season multiple times is safe (idempotent): existing
// records are replaced with the latest values from the save game.
type ImportService struct{}

func NewImportService() *ImportService {
	return &ImportService{}
}

// ImportResult summarises the outcome of a season import.
type ImportResult struct {
	SeasonID     int64 // companion DB seasons.id (autoincrement)
	SeasonNum    int   // display season number (save game num + offset)
	Players      int
	Teams        int
	Games        int
	PlayoffGames int
}

// ImportSeason imports all data for the given season from reader into the
// companion database. The entire import runs in a single transaction — either
// all data lands or none does.
//
// saveGameSeasonID is the raw t_seasons.id from the save file (used to query
// the save game reader). leagueGUID identifies which franchise_source produced
// this season and forms part of the companion DB uniqueness key. seasonOffset
// is added to the save game's season number to produce the display season number.
func (svc *ImportService) ImportSeason(
	ctx context.Context,
	companionDB *sql.DB,
	reader store.SaveGameReader,
	saveGameSeasonID int,
	saveGameSeasonNum int,
	leagueGUID string,
	seasonOffset int,
) (ImportResult, error) {
	start := time.Now()
	tx, err := companionDB.BeginTx(ctx, nil)
	if err != nil {
		return ImportResult{}, fmt.Errorf("beginning import transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result, err := svc.importInTx(ctx, tx, reader, saveGameSeasonID, saveGameSeasonNum, leagueGUID, seasonOffset)
	if err != nil {
		return ImportResult{}, err
	}

	if err = tx.Commit(); err != nil {
		return ImportResult{}, fmt.Errorf("committing import transaction: %w", err)
	}
	slog.Info("import: complete",
		"seasonNum", result.SeasonNum,
		"players", result.Players,
		"teams", result.Teams,
		"games", result.Games,
		"playoffGames", result.PlayoffGames,
		"duration", time.Since(start).Round(time.Millisecond),
	)
	return result, nil
}

// importInTx performs the actual import work within the provided transaction.
//
//nolint:gocognit // 14-step sequential pipeline — each step depends on the previous step's output; decomposing would scatter the pipeline state across many call sites without reducing real complexity
func (svc *ImportService) importInTx(
	ctx context.Context,
	tx *sql.Tx,
	reader store.SaveGameReader,
	saveGameSeasonID int,
	saveGameSeasonNum int,
	leagueGUID string,
	seasonOffset int,
) (ImportResult, error) {
	seasonStore := store.NewSeasonStore(tx)
	teamStore := store.NewTeamHistoryStore(tx)
	playerStore := store.NewPlayerSeasonStore(tx)
	scheduleStore := store.NewScheduleStore(tx)

	start := time.Now()
	displaySeasonNum := saveGameSeasonNum + seasonOffset
	result := ImportResult{SeasonNum: displaySeasonNum}
	slog.Info("import: starting", "seasonNum", displaySeasonNum)

	// ── 1. Season record ────────────────────────────────────────────────────
	companionSeasonID, err := seasonStore.Upsert(ctx, store.Season{
		LeagueGUID:       leagueGUID,
		SaveGameSeasonID: saveGameSeasonID,
		SeasonNum:        displaySeasonNum,
	})
	if err != nil {
		slog.Error("import: step 1 failed", "step", "upsert season", "err", err)
		return result, fmt.Errorf("upserting season: %w", err)
	}
	result.SeasonID = companionSeasonID
	slog.Debug("import: step 1 complete", "step", "season record", "seasonID", companionSeasonID)

	// ── 2. Teams ────────────────────────────────────────────────────────────
	teamGUIDToHistoryID, teamNameToHistoryID, teamCount, err := svc.importTeams(ctx, teamStore, reader, saveGameSeasonID, companionSeasonID)
	if err != nil {
		slog.Error("import: step 2 failed", "step", "teams", "err", err)
		return result, err
	}
	result.Teams = teamCount
	slog.Debug("import: step 2 complete", "step", "teams", "teams", teamCount)

	// ── 3. Players ──────────────────────────────────────────────────────────
	playerGUIDToSeasonID, playerIDs, err := svc.importPlayers(ctx, playerStore, reader, saveGameSeasonID, companionSeasonID, teamNameToHistoryID)
	if err != nil {
		slog.Error("import: step 3 failed", "step", "players", "err", err)
		return result, err
	}
	result.Players = len(playerGUIDToSeasonID)
	slog.Debug("import: step 3 complete", "step", "players", "players", result.Players)

	// ── 4. League attribute averages ────────────────────────────────────────
	// Runs after all game stats are written (step 3) so the aggregate is complete.
	if err := ApplyLeagueAvgAttributes(ctx, tx, companionSeasonID); err != nil {
		slog.Error("import: step 4 failed", "step", "league attribute averages", "err", err)
		return result, fmt.Errorf("computing league attribute averages: %w", err)
	}
	slog.Debug("import: step 4 complete", "step", "league attribute averages")

	// ── 5. Player attribute percentiles ─────────────────────────────────────
	// Runs after game stats (step 3) so PERCENT_RANK has complete season data.
	if err := ApplyPlayerAttributePercentiles(ctx, tx, companionSeasonID); err != nil {
		slog.Error("import: step 5 failed", "step", "player attribute percentiles", "err", err)
		return result, fmt.Errorf("computing player attribute percentiles: %w", err)
	}
	slog.Debug("import: step 5 complete", "step", "player attribute percentiles")

	// ── 6–9. Batting and pitching stats (regular season + playoffs) ─────────
	if err := svc.importBattingStats(ctx, playerStore, reader, saveGameSeasonID, true, playerGUIDToSeasonID); err != nil {
		slog.Error("import: step 6 failed", "step", "regular season batting stats", "err", err)
		return result, fmt.Errorf("importing regular season batting stats: %w", err)
	}
	slog.Debug("import: step 6 complete", "step", "regular season batting stats")

	if err := svc.importPitchingStats(ctx, playerStore, reader, saveGameSeasonID, true, playerGUIDToSeasonID); err != nil {
		slog.Error("import: step 7 failed", "step", "regular season pitching stats", "err", err)
		return result, fmt.Errorf("importing regular season pitching stats: %w", err)
	}
	slog.Debug("import: step 7 complete", "step", "regular season pitching stats")

	if err := svc.importBattingStats(ctx, playerStore, reader, saveGameSeasonID, false, playerGUIDToSeasonID); err != nil {
		slog.Error("import: step 8 failed", "step", "playoff batting stats", "err", err)
		return result, fmt.Errorf("importing playoff batting stats: %w", err)
	}
	slog.Debug("import: step 8 complete", "step", "playoff batting stats")

	if err := svc.importPitchingStats(ctx, playerStore, reader, saveGameSeasonID, false, playerGUIDToSeasonID); err != nil {
		slog.Error("import: step 9 failed", "step", "playoff pitching stats", "err", err)
		return result, fmt.Errorf("importing playoff pitching stats: %w", err)
	}
	slog.Debug("import: step 9 complete", "step", "playoff pitching stats")

	// ── 10. Context stats (OPS+, ERA+, FIP, FIP-, smbWAR) ──────────────────
	// Runs after all counting stats are written so league aggregates are complete.
	if err := ApplyContextStats(ctx, tx, companionSeasonID, true); err != nil {
		slog.Error("import: step 10 failed", "step", "context stats (regular season)", "err", err)
		return result, fmt.Errorf("computing context stats (regular season): %w", err)
	}
	if err := ApplyContextStats(ctx, tx, companionSeasonID, false); err != nil {
		slog.Error("import: step 10 failed", "step", "context stats (playoffs)", "err", err)
		return result, fmt.Errorf("computing context stats (playoffs): %w", err)
	}
	slog.Debug("import: step 10 complete", "step", "context stats")

	// ── 11. Career stats ─────────────────────────────────────────────────────
	// Must run after ApplyContextStats so per-season smb_war values are set.
	if err := ApplyCareerStats(ctx, tx, playerIDs); err != nil {
		slog.Error("import: step 11 failed", "step", "career stats", "err", err)
		return result, fmt.Errorf("computing career stats: %w", err)
	}
	slog.Debug("import: step 11 complete", "step", "career stats", "players", len(playerIDs))

	// ── 12. Regular season schedule ──────────────────────────────────────────
	if err := scheduleStore.DeleteSeasonSchedule(ctx, companionSeasonID); err != nil {
		slog.Error("import: step 12 failed", "step", "clear old schedule", "err", err)
		return result, fmt.Errorf("clearing old schedule: %w", err)
	}
	gameCount, err := svc.importRegularSeasonSchedule(ctx, scheduleStore, reader, saveGameSeasonID, companionSeasonID, teamGUIDToHistoryID, playerGUIDToSeasonID)
	if err != nil {
		slog.Error("import: step 12 failed", "step", "regular season schedule", "err", err)
		return result, err
	}
	result.Games = gameCount
	slog.Debug("import: step 12 complete", "step", "regular season schedule", "games", gameCount)

	// Backfill the season's scheduled game count now that it's known. This is required
	// for qualified-player thresholds (PA/IP gating) computed elsewhere from num_games —
	// without it, num_games stays 0 and every player incorrectly qualifies as a leader.
	if _, err := seasonStore.Upsert(ctx, store.Season{
		LeagueGUID:       leagueGUID,
		SaveGameSeasonID: saveGameSeasonID,
		SeasonNum:        displaySeasonNum,
		NumGames:         gameCount,
	}); err != nil {
		slog.Error("import: step 12 failed", "step", "backfill season game count", "err", err)
		return result, fmt.Errorf("backfilling season game count: %w", err)
	}

	// ── 13. Playoff schedule + seeds ─────────────────────────────────────────
	playoffCount, err := svc.importPlayoffScheduleAndSeeds(ctx, scheduleStore, teamStore, reader, saveGameSeasonID, companionSeasonID, teamGUIDToHistoryID, playerGUIDToSeasonID)
	if err != nil {
		slog.Error("import: step 13 failed", "step", "playoff schedule", "err", err)
		return result, err
	}
	result.PlayoffGames = playoffCount
	slog.Debug("import: step 13 complete", "step", "playoff schedule", "games", playoffCount)

	// ── 14. Playoff config ────────────────────────────────────────────────────
	// Prefer the authoritative values from t_playoffs; fall back to inference
	// from the game data we just persisted (covers saves where t_playoffs is
	// missing and the legacy import path).
	cfg, err := reader.GetSeasonPlayoffConfig(ctx, saveGameSeasonID)
	if err != nil {
		slog.Error("import: step 14 failed", "step", "playoff config", "err", err)
		return result, fmt.Errorf("reading playoff config: %w", err)
	}
	if cfg != nil {
		if err := seasonStore.UpdatePlayoffConfig(ctx, companionSeasonID, cfg.Rounds, cfg.SeriesLength); err != nil {
			slog.Error("import: step 14 failed", "step", "persist playoff config", "err", err)
			return result, fmt.Errorf("persisting playoff config: %w", err)
		}
	} else if playoffCount > 0 {
		if err := seasonStore.InferAndSetPlayoffConfig(ctx, companionSeasonID); err != nil {
			slog.Error("import: step 14 failed", "step", "infer playoff config", "err", err)
			return result, fmt.Errorf("inferring playoff config: %w", err)
		}
	}
	slog.Debug("import: step 14 complete", "step", "playoff config", "duration", time.Since(start).Round(time.Millisecond))

	return result, nil
}

// importTeams upserts all teams and season histories for the current season.
// Returns two lookup maps (GUID-keyed for schedule use, name-keyed for player team
// association) along with the team count.
func (svc *ImportService) importTeams(
	ctx context.Context,
	teams *store.TeamHistoryStore,
	reader store.SaveGameReader,
	saveGameSeasonID int,
	companionSeasonID int64,
) (guidMap map[string]int64, nameMap map[string]int64, count int, err error) {
	saveTeams, err := reader.GetCurrentSeasonTeams(ctx, saveGameSeasonID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("reading teams: %w", err)
	}
	guidMap = make(map[string]int64, len(saveTeams))
	nameMap = make(map[string]int64, len(saveTeams))
	for _, t := range saveTeams {
		teamID, err := teams.UpsertTeam(ctx, t.TeamGUID, t.TeamName)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("upserting team %s: %w", t.TeamGUID, err)
		}
		histID, err := teams.UpsertSeasonHistory(ctx, store.TeamSeasonHistory{
			TeamID:         teamID,
			SeasonID:       companionSeasonID,
			TeamName:       t.TeamName,
			DivisionName:   t.DivisionName,
			ConferenceName: t.ConferenceName,
			Budget:         t.Budget,
			Payroll:        t.Payroll,
			Wins:           t.Wins,
			Losses:         t.Losses,
			GamesBack:      t.GamesBack,
			RunsFor:        t.RunsFor,
			RunsAgainst:    t.RunsAgainst,
		})
		if err != nil {
			return nil, nil, 0, fmt.Errorf("upserting team season history for %s: %w", t.TeamGUID, err)
		}
		guidMap[t.TeamGUID] = histID
		nameMap[t.TeamName] = histID
	}
	return guidMap, nameMap, len(saveTeams), nil
}

// importPlayers upserts all players, their season records, game stats, and team
// associations. Returns a GUID→seasonID map and the player ID slice (needed for
// career stat computation).
func (svc *ImportService) importPlayers(
	ctx context.Context,
	players *store.PlayerSeasonStore,
	reader store.SaveGameReader,
	saveGameSeasonID int,
	companionSeasonID int64,
	teamNameToHistoryID map[string]int64,
) (playerGUIDToSeasonID map[string]int64, playerIDs []int64, err error) {
	savePlayers, err := reader.GetCurrentSeasonPlayers(ctx, saveGameSeasonID)
	if err != nil {
		return nil, nil, fmt.Errorf("reading players: %w", err)
	}
	playerGUIDToSeasonID = make(map[string]int64, len(savePlayers))
	playerIDs = make([]int64, 0, len(savePlayers))
	for _, p := range savePlayers {
		playerID, err := players.UpsertPlayer(ctx, store.PlayerIdentity{
			GameGUID:      p.PlayerGUID,
			FirstName:     p.FirstName,
			LastName:      p.LastName,
			BatHand:       p.BatHand,
			ThrowHand:     p.ThrowHand,
			ChemistryType: p.ChemistryType,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("upserting player %s: %w", p.PlayerGUID, err)
		}
		psID, err := players.UpsertSeason(ctx, store.PlayerSeason{
			PlayerID:          playerID,
			SeasonID:          companionSeasonID,
			Age:               p.Age,
			Salary:            p.Salary,
			PrimaryPosition:   p.PrimaryPos,
			SecondaryPosition: p.SecondaryPos,
			PitcherRole:       p.PitcherRole,
			BatHand:           p.BatHand,
			ThrowHand:         p.ThrowHand,
			ChemistryType:     p.ChemistryType,
			TraitsJSON:        joinStrings(p.Traits),
			PitchesJSON:       joinStrings(p.Pitches),
		})
		if err != nil {
			return nil, nil, fmt.Errorf("upserting player season for %s: %w", p.PlayerGUID, err)
		}
		playerGUIDToSeasonID[p.PlayerGUID] = psID
		playerIDs = append(playerIDs, playerID)

		// Populate team associations from all three save-game team name pointers.
		// sort_order: 0=current/final, 1=most recently played prior, 2=two teams ago.
		teamNames := [3]string{p.CurrentTeam, p.PreviousTeam, p.Prev2Team}
		var seasonTeams []store.PlayerSeasonTeam
		seen := map[int64]bool{}
		for order, name := range teamNames {
			if name == "" {
				continue
			}
			hid, ok := teamNameToHistoryID[name]
			if !ok || seen[hid] {
				continue
			}
			seen[hid] = true
			seasonTeams = append(seasonTeams, store.PlayerSeasonTeam{
				PlayerSeasonID: psID,
				TeamHistoryID:  hid,
				SortOrder:      order,
			})
		}
		if err := players.ReplaceSeasonTeams(ctx, psID, seasonTeams); err != nil {
			return nil, nil, fmt.Errorf("upserting season teams for %s: %w", p.PlayerGUID, err)
		}
		if err := players.UpsertGameStats(ctx, store.PlayerSeasonGameStats{
			PlayerSeasonID: psID,
			Power:          p.Power,
			Contact:        p.Contact,
			Speed:          p.Speed,
			Fielding:       p.Fielding,
			Arm:            p.Arm,
			Velocity:       p.Velocity,
			Junk:           p.Junk,
			Accuracy:       p.Accuracy,
		}); err != nil {
			return nil, nil, fmt.Errorf("upserting game stats for %s: %w", p.PlayerGUID, err)
		}
	}
	return playerGUIDToSeasonID, playerIDs, nil
}

// importRegularSeasonSchedule writes regular-season games for the season.
// Returns the number of games written.
func (svc *ImportService) importRegularSeasonSchedule(
	ctx context.Context,
	sched *store.ScheduleStore,
	reader store.SaveGameReader,
	saveGameSeasonID int,
	companionSeasonID int64,
	teamGUIDToHistoryID map[string]int64,
	playerGUIDToSeasonID map[string]int64,
) (int, error) {
	games, err := reader.GetSeasonSchedule(ctx, saveGameSeasonID)
	if err != nil {
		return 0, fmt.Errorf("reading season schedule: %w", err)
	}
	for _, g := range games {
		homeHistID := teamGUIDToHistoryID[g.HomeTeamGUID]
		awayHistID := teamGUIDToHistoryID[g.AwayTeamGUID]
		if homeHistID == 0 || awayHistID == 0 {
			continue
		}
		var homePitcherID, awayPitcherID *int64
		if g.HomePitcherGUID != nil {
			if id, ok := playerGUIDToSeasonID[*g.HomePitcherGUID]; ok {
				homePitcherID = &id
			}
		}
		if g.AwayPitcherGUID != nil {
			if id, ok := playerGUIDToSeasonID[*g.AwayPitcherGUID]; ok {
				awayPitcherID = &id
			}
		}
		if err := sched.UpsertGame(ctx, store.ScheduleGame{
			SeasonID:            companionSeasonID,
			GameNumber:          g.GameNumber,
			Day:                 g.Day,
			HomeTeamHistoryID:   homeHistID,
			AwayTeamHistoryID:   awayHistID,
			HomePitcherSeasonID: homePitcherID,
			AwayPitcherSeasonID: awayPitcherID,
			HomeScore:           g.HomeScore,
			AwayScore:           g.AwayScore,
		}); err != nil {
			return 0, fmt.Errorf("upserting game %d: %w", g.GameNumber, err)
		}
	}
	return len(games), nil
}

// importPlayoffScheduleAndSeeds writes playoff games for the season and records
// playoff seeds. Returns the number of playoff games written.
func (svc *ImportService) importPlayoffScheduleAndSeeds(
	ctx context.Context,
	sched *store.ScheduleStore,
	teams *store.TeamHistoryStore,
	reader store.SaveGameReader,
	saveGameSeasonID int,
	companionSeasonID int64,
	teamGUIDToHistoryID map[string]int64,
	playerGUIDToSeasonID map[string]int64,
) (int, error) {
	games, err := reader.GetPlayoffSchedule(ctx, saveGameSeasonID)
	if err != nil {
		return 0, fmt.Errorf("reading playoff schedule: %w", err)
	}
	for _, g := range games {
		homeHistID := teamGUIDToHistoryID[g.HomeTeamGUID]
		awayHistID := teamGUIDToHistoryID[g.AwayTeamGUID]
		if homeHistID == 0 || awayHistID == 0 {
			continue
		}
		var homePitcherID, awayPitcherID *int64
		if g.HomePitcherGUID != nil {
			if id, ok := playerGUIDToSeasonID[*g.HomePitcherGUID]; ok {
				homePitcherID = &id
			}
		}
		if g.AwayPitcherGUID != nil {
			if id, ok := playerGUIDToSeasonID[*g.AwayPitcherGUID]; ok {
				awayPitcherID = &id
			}
		}
		if err := sched.UpsertPlayoffGame(ctx, store.PlayoffGame{
			SeasonID:            companionSeasonID,
			SeriesNumber:        g.SeriesNum,
			GameNumber:          g.GameNumber,
			HomeTeamHistoryID:   homeHistID,
			AwayTeamHistoryID:   awayHistID,
			HomePitcherSeasonID: homePitcherID,
			AwayPitcherSeasonID: awayPitcherID,
			HomeScore:           g.HomeScore,
			AwayScore:           g.AwayScore,
		}); err != nil {
			return 0, fmt.Errorf("upserting playoff game %d: %w", g.GameNumber, err)
		}
	}
	if len(games) > 0 {
		if err := updatePlayoffSeeds(ctx, teams, games, teamGUIDToHistoryID); err != nil {
			return 0, err
		}
	}
	return len(games), nil
}

// updatePlayoffSeeds records each team's bracket seed from the playoff game data.
// The save game stores seeds 0-indexed (0 = #1 seed); we add 1 for display.
// Multiple games per team carry the same seed; a map deduplicates them.
func updatePlayoffSeeds(ctx context.Context, teams *store.TeamHistoryStore, games []models.SaveGamePlayoffGame, teamGUIDToHistoryID map[string]int64) error {
	seedMap := make(map[int64]int)
	for _, g := range games {
		if histID := teamGUIDToHistoryID[g.Team1GUID]; histID != 0 {
			seedMap[histID] = g.Team1Seed + 1
		}
		if histID := teamGUIDToHistoryID[g.Team2GUID]; histID != 0 {
			seedMap[histID] = g.Team2Seed + 1
		}
	}
	return teams.UpdatePlayoffSeeds(ctx, seedMap)
}

func (svc *ImportService) importBattingStats(
	ctx context.Context,
	players *store.PlayerSeasonStore,
	reader store.SaveGameReader,
	saveGameSeasonID int,
	isRegularSeason bool,
	playerGUIDToSeasonID map[string]int64,
) error {
	var stats []models.SaveGameBattingStat
	var err error
	if isRegularSeason {
		stats, err = reader.GetSeasonBattingStats(ctx, saveGameSeasonID)
	} else {
		stats, err = reader.GetPlayoffBattingStats(ctx, saveGameSeasonID)
	}
	if err != nil {
		return err
	}

	for _, s := range stats {
		psID, ok := playerGUIDToSeasonID[s.PlayerGUID]
		if !ok {
			continue
		}
		tmp := models.CareerBattingStats{
			AtBats: s.AtBats, Hits: s.Hits, Doubles: s.Doubles, Triples: s.Triples,
			HomeRuns: s.HomeRuns, Walks: s.Walks, HitByPitch: s.HitByPitch,
			SacHits: s.SacHits, SacFlies: s.SacFlies, Strikeouts: s.Strikeouts,
		}
		ComputeBattingRates(&tmp)
		if err := players.UpsertBattingStats(ctx, store.PlayerSeasonBattingStats{
			PlayerSeasonID:   psID,
			IsRegularSeason:  isRegularSeason,
			GamesPlayed:      s.GamesPlayed,
			GamesBatting:     s.GamesBatting,
			AtBats:           s.AtBats,
			PlateAppearances: s.AtBats + s.Walks + s.HitByPitch + s.SacHits + s.SacFlies,
			Runs:             s.Runs,
			Hits:             s.Hits,
			Doubles:          s.Doubles,
			Triples:          s.Triples,
			HomeRuns:         s.HomeRuns,
			RBI:              s.RBI,
			StolenBases:      s.StolenBases,
			CaughtStealing:   s.CaughtStealing,
			Walks:            s.Walks,
			Strikeouts:       s.Strikeouts,
			HitByPitch:       s.HitByPitch,
			SacHits:          s.SacHits,
			SacFlies:         s.SacFlies,
			Errors:           s.Errors,
			PassedBalls:      s.PassedBalls,
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
			return fmt.Errorf("batting stats for %s: %w", s.PlayerGUID, err)
		}
	}
	return nil
}

func (svc *ImportService) importPitchingStats(
	ctx context.Context,
	players *store.PlayerSeasonStore,
	reader store.SaveGameReader,
	saveGameSeasonID int,
	isRegularSeason bool,
	playerGUIDToSeasonID map[string]int64,
) error {
	var stats []models.SaveGamePitchingStat
	var err error
	if isRegularSeason {
		stats, err = reader.GetSeasonPitchingStats(ctx, saveGameSeasonID)
	} else {
		stats, err = reader.GetPlayoffPitchingStats(ctx, saveGameSeasonID)
	}
	if err != nil {
		return err
	}

	for _, s := range stats {
		psID, ok := playerGUIDToSeasonID[s.PlayerGUID]
		if !ok {
			continue
		}
		tmp := models.CareerPitchingStats{
			OutsPitched: s.OutsPitched, EarnedRuns: s.EarnedRuns,
			HitsAllowed: s.HitsAllowed, HomeRunsAllowed: s.HomeRunsAllowed,
			Walks: s.Walks, Strikeouts: s.Strikeouts, HitBatters: s.HitBatters,
			BattersFaced: s.BattersFaced, TotalPitches: s.TotalPitches,
			Wins: s.Wins, Losses: s.Losses,
		}
		ComputePitchingRates(&tmp)
		if err := players.UpsertPitchingStats(ctx, store.PlayerSeasonPitchingStats{
			PlayerSeasonID:  psID,
			IsRegularSeason: isRegularSeason,
			Wins:            s.Wins,
			Losses:          s.Losses,
			Games:           s.Games,
			GamesStarted:    s.GamesStarted,
			CompleteGames:   s.CompleteGames,
			Shutouts:        s.Shutouts,
			Saves:           s.Saves,
			OutsPitched:     s.OutsPitched,
			HitsAllowed:     s.HitsAllowed,
			EarnedRuns:      s.EarnedRuns,
			HomeRunsAllowed: s.HomeRunsAllowed,
			Walks:           s.Walks,
			Strikeouts:      s.Strikeouts,
			HitBatters:      s.HitBatters,
			BattersFaced:    s.BattersFaced,
			GamesFinished:   s.GamesFinished,
			RunsAllowed:     s.RunsAllowed,
			WildPitches:     s.WildPitches,
			TotalPitches:    s.TotalPitches,
			ERA:             tmp.ERA,
			WHIP:            tmp.WHIP,
			K9:              tmp.K9,
			BB9:             tmp.BB9,
			H9:              tmp.H9,
			HR9:             tmp.HR9,
			KPerBB:          tmp.KPerBB,
			KPct:            tmp.KPct,
			WinPct:          tmp.WinPct,
			PPerIP:          tmp.PPerIP,
		}); err != nil {
			return fmt.Errorf("pitching stats for %s: %w", s.PlayerGUID, err)
		}
	}
	return nil
}

// ---- helpers ---------------------------------------------------------------

func joinStrings(ss []string) string {
	if len(ss) == 0 {
		return "[]"
	}
	return "[\"" + strings.Join(ss, "\",\"") + "\"]"
}
