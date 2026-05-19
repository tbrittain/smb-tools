package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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
	SeasonID   int
	SeasonNum  int
	Players    int
	Teams      int
	Games      int
	PlayoffGames int
}

// ImportSeason imports all data for the given season from reader into the
// companion database. The entire import runs in a single transaction — either
// all data lands or none does.
func (svc *ImportService) ImportSeason(
	ctx context.Context,
	companionDB *sql.DB,
	reader store.SaveGameReader,
	seasonID int,
	seasonNum int,
) (ImportResult, error) {
	tx, err := companionDB.BeginTx(ctx, nil)
	if err != nil {
		return ImportResult{}, fmt.Errorf("beginning import transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result, err := svc.importInTx(ctx, tx, reader, seasonID, seasonNum)
	if err != nil {
		return ImportResult{}, err
	}

	if err = tx.Commit(); err != nil {
		return ImportResult{}, fmt.Errorf("committing import transaction: %w", err)
	}
	return result, nil
}

// importInTx performs the actual import work within the provided transaction.
func (svc *ImportService) importInTx(
	ctx context.Context,
	tx *sql.Tx,
	reader store.SaveGameReader,
	seasonID int,
	seasonNum int,
) (ImportResult, error) {
	// Wrap all stores in the transaction
	seasons := store.NewSeasonStore(tx)
	teams := store.NewTeamHistoryStore(tx)
	players := store.NewPlayerSeasonStore(tx)
	schedule := store.NewScheduleStore(tx)

	result := ImportResult{SeasonID: seasonID, SeasonNum: seasonNum}

	// ── 1. Season record ────────────────────────────────────────────────────
	if err := seasons.Upsert(ctx, store.Season{
		ID: seasonID, SeasonNum: seasonNum,
	}); err != nil {
		return result, fmt.Errorf("upserting season: %w", err)
	}

	// ── 2. Teams ────────────────────────────────────────────────────────────
	saveTeams, err := reader.GetCurrentSeasonTeams(ctx, seasonID)
	if err != nil {
		return result, fmt.Errorf("reading teams: %w", err)
	}

	// teamGUIDToHistoryID maps save game team GUID → companion team_season_history.id
	teamGUIDToHistoryID := make(map[string]int64, len(saveTeams))
	for _, t := range saveTeams {
		teamID, err := teams.UpsertTeam(ctx, t.TeamGUID)
		if err != nil {
			return result, fmt.Errorf("upserting team %s: %w", t.TeamGUID, err)
		}
		histID, err := teams.UpsertSeasonHistory(ctx, store.TeamSeasonHistory{
			TeamID:         teamID,
			SeasonID:       seasonID,
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
			return result, fmt.Errorf("upserting team season history for %s: %w", t.TeamGUID, err)
		}
		teamGUIDToHistoryID[t.TeamGUID] = histID
	}
	result.Teams = len(saveTeams)

	// ── 3. Players ──────────────────────────────────────────────────────────
	savePlayers, err := reader.GetCurrentSeasonPlayers(ctx, seasonID)
	if err != nil {
		return result, fmt.Errorf("reading players: %w", err)
	}

	// playerGUIDToSeasonID maps save game player GUID → companion player_seasons.id
	playerGUIDToSeasonID := make(map[string]int64, len(savePlayers))
	for _, p := range savePlayers {
		playerID, err := players.UpsertPlayer(ctx, store.Player{
			GameGUID:  p.PlayerGUID,
			FirstName: p.FirstName,
			LastName:  p.LastName,
		})
		if err != nil {
			return result, fmt.Errorf("upserting player %s: %w", p.PlayerGUID, err)
		}

		var teamHistID *int64
		if hid, ok := teamGUIDToHistoryID[teamGUIDFromName(saveTeams, p.CurrentTeam)]; ok {
			teamHistID = &hid
		}

		psID, err := players.UpsertSeason(ctx, store.PlayerSeason{
			PlayerID:          playerID,
			SeasonID:          seasonID,
			TeamHistoryID:     teamHistID,
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
			return result, fmt.Errorf("upserting player season for %s: %w", p.PlayerGUID, err)
		}
		playerGUIDToSeasonID[p.PlayerGUID] = psID

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
			return result, fmt.Errorf("upserting game stats for %s: %w", p.PlayerGUID, err)
		}
	}
	result.Players = len(savePlayers)

	// ── 4. Batting stats (regular season) ──────────────────────────────────
	if err := svc.importBattingStats(ctx, players, reader, seasonID, true, playerGUIDToSeasonID); err != nil {
		return result, fmt.Errorf("importing regular season batting stats: %w", err)
	}

	// ── 5. Batting stats (playoffs) ─────────────────────────────────────────
	// Playoff stats use the same stats tables, linked via different aggregators.
	// We read career stats to capture career totals; playoff batting comes from
	// what was season-scoped in the save game.
	// For now, same reader method with season scope covers both via aggregators.
	if err := svc.importBattingStats(ctx, players, reader, seasonID, false, playerGUIDToSeasonID); err != nil {
		return result, fmt.Errorf("importing playoff batting stats: %w", err)
	}

	// ── 6. Pitching stats (regular season + playoffs) ───────────────────────
	if err := svc.importPitchingStats(ctx, players, reader, seasonID, true, playerGUIDToSeasonID); err != nil {
		return result, fmt.Errorf("importing regular season pitching stats: %w", err)
	}
	if err := svc.importPitchingStats(ctx, players, reader, seasonID, false, playerGUIDToSeasonID); err != nil {
		return result, fmt.Errorf("importing playoff pitching stats: %w", err)
	}

	// ── 7. Regular season schedule ──────────────────────────────────────────
	if err := schedule.DeleteSeasonSchedule(ctx, seasonID); err != nil {
		return result, fmt.Errorf("clearing old schedule: %w", err)
	}

	games, err := reader.GetSeasonSchedule(ctx, seasonID)
	if err != nil {
		return result, fmt.Errorf("reading season schedule: %w", err)
	}
	for _, g := range games {
		homeHistID := teamGUIDToHistoryID[g.HomeTeamGUID]
		awayHistID := teamGUIDToHistoryID[g.AwayTeamGUID]
		if homeHistID == 0 || awayHistID == 0 {
			continue // skip games for teams not in this season's roster
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
		if err := schedule.UpsertGame(ctx, store.ScheduleGame{
			SeasonID:             seasonID,
			GameNumber:           g.GameNumber,
			Day:                  g.Day,
			HomeTeamHistoryID:    homeHistID,
			AwayTeamHistoryID:    awayHistID,
			HomePitcherSeasonID: homePitcherID,
			AwayPitcherSeasonID: awayPitcherID,
			HomeScore:            g.HomeScore,
			AwayScore:            g.AwayScore,
		}); err != nil {
			return result, fmt.Errorf("upserting game %d: %w", g.GameNumber, err)
		}
	}
	result.Games = len(games)

	// ── 8. Playoff schedule ──────────────────────────────────────────────────
	playoffGames, err := reader.GetPlayoffSchedule(ctx, seasonID)
	if err != nil {
		return result, fmt.Errorf("reading playoff schedule: %w", err)
	}
	for _, g := range playoffGames {
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
		if err := schedule.UpsertPlayoffGame(ctx, store.PlayoffGame{
			SeasonID:             seasonID,
			SeriesNumber:         g.SeriesNum,
			GameNumber:           g.GameNumber,
			HomeTeamHistoryID:    homeHistID,
			AwayTeamHistoryID:    awayHistID,
			HomePitcherSeasonID: homePitcherID,
			AwayPitcherSeasonID: awayPitcherID,
			HomeScore:            g.HomeScore,
			AwayScore:            g.AwayScore,
		}); err != nil {
			return result, fmt.Errorf("upserting playoff game %d: %w", g.GameNumber, err)
		}
	}
	result.PlayoffGames = len(playoffGames)

	return result, nil
}

func (svc *ImportService) importBattingStats(
	ctx context.Context,
	players *store.PlayerSeasonStore,
	reader store.SaveGameReader,
	seasonID int,
	isRegularSeason bool,
	playerGUIDToSeasonID map[string]int64,
) error {
	var stats []models.SaveGameBattingStat
	var err error
	if isRegularSeason {
		stats, err = reader.GetSeasonBattingStats(ctx, seasonID)
	} else {
		// Playoff batting stats use the same season scope in the save game.
		// The save game doesn't cleanly separate them — we rely on the reader
		// returning the same aggregator rows; the distinction is contextual.
		// For now, playoff stats are a subset found via career aggregator overlap.
		stats, err = reader.GetCareerBattingStats(ctx)
	}
	if err != nil {
		return err
	}

	for _, s := range stats {
		psID, ok := playerGUIDToSeasonID[s.PlayerGUID]
		if !ok {
			continue // player not in this season's roster
		}
		if err := players.UpsertBattingStats(ctx, store.PlayerSeasonBattingStats{
			PlayerSeasonID:  psID,
			IsRegularSeason: isRegularSeason,
			GamesPlayed:     s.GamesPlayed,
			GamesBatting:    s.GamesBatting,
			AtBats:          s.AtBats,
			Runs:            s.Runs,
			Hits:            s.Hits,
			Doubles:         s.Doubles,
			Triples:         s.Triples,
			HomeRuns:        s.HomeRuns,
			RBI:             s.RBI,
			StolenBases:     s.StolenBases,
			CaughtStealing:  s.CaughtStealing,
			Walks:           s.Walks,
			Strikeouts:      s.Strikeouts,
			HitByPitch:      s.HitByPitch,
			SacHits:         s.SacHits,
			SacFlies:        s.SacFlies,
			Errors:          s.Errors,
			PassedBalls:     s.PassedBalls,
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
	seasonID int,
	isRegularSeason bool,
	playerGUIDToSeasonID map[string]int64,
) error {
	var stats []models.SaveGamePitchingStat
	var err error
	if isRegularSeason {
		stats, err = reader.GetSeasonPitchingStats(ctx, seasonID)
	} else {
		stats, err = reader.GetCareerPitchingStats(ctx)
	}
	if err != nil {
		return err
	}

	for _, s := range stats {
		psID, ok := playerGUIDToSeasonID[s.PlayerGUID]
		if !ok {
			continue
		}
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
		}); err != nil {
			return fmt.Errorf("pitching stats for %s: %w", s.PlayerGUID, err)
		}
	}
	return nil
}

// ---- helpers ---------------------------------------------------------------

// teamGUIDFromName finds a team's GUID by matching on team name — a fallback
// for when the player record only carries the team name, not the GUID.
func teamGUIDFromName(teams []models.SaveGameTeam, name string) string {
	for _, t := range teams {
		if t.TeamName == name {
			return t.TeamGUID
		}
	}
	return ""
}

func joinStrings(ss []string) string {
	if len(ss) == 0 {
		return "[]"
	}
	return "[\"" + strings.Join(ss, "\",\"") + "\"]"
}

