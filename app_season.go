package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"

	"smb-tools/internal/models"
)

// HasSeasonsMissingInningsPerGame reports whether the active franchise has any
// season rows that predate the innings_per_game column (synced or
// legacy-migrated before this feature existed). The Setup page uses this to
// show or hide the one-time backfill prompt.
func (a *App) HasSeasonsMissingInningsPerGame() (bool, error) {
	if err := a.requireCompanionDB(); err != nil {
		return false, err
	}
	return a.seasonStore.HasSeasonsMissingInningsPerGame(a.ctx)
}

// BackfillInningsPerGame sets innings_per_game on every season row that
// predates the column, to the actual game length the user supplies. There is
// no derivable value for these rows, so the user must provide it explicitly.
func (a *App) BackfillInningsPerGame(innings int) error {
	if err := a.requireCompanionDB(); err != nil {
		return err
	}
	return a.seasonStore.BackfillInningsPerGame(a.ctx, innings)
}

// GetSeasonList returns all synced seasons with champion information.
func (a *App) GetSeasonList() ([]SeasonSummaryDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	seasons, err := a.seasonQueryStore.ListWithChampion(a.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]SeasonSummaryDTO, len(seasons))
	for i, s := range seasons {
		out[i] = SeasonSummaryDTO{
			ID:                s.ID,
			SeasonNum:         s.SeasonNum,
			NumGames:          s.NumGames,
			ImportedAt:        s.ImportedAt.Format("2006-01-02T15:04:05Z"),
			ChampionTeamName:  s.ChampionTeamName,
			ChampionHistoryID: s.ChampionHistoryID,
		}
	}
	return out, nil
}

// GetStandings returns all teams' standings for the given season.
func (a *App) GetStandings(seasonID int64) ([]TeamStandingDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	rows, err := a.seasonQueryStore.GetStandings(a.ctx, seasonID)
	if err != nil {
		return nil, err
	}
	out := make([]TeamStandingDTO, len(rows))
	for i, r := range rows {
		out[i] = TeamStandingDTO{
			HistoryID:      r.HistoryID,
			TeamID:         r.TeamID,
			TeamName:       r.TeamName,
			DivisionName:   r.DivisionName,
			ConferenceName: r.ConferenceName,
			Wins:           r.Wins,
			Losses:         r.Losses,
			WinPct:         r.WinPct,
			GamesBack:      r.GamesBack,
			RunsFor:        r.RunsFor,
			RunsAgainst:    r.RunsAgainst,
			RunDiff:        r.RunDiff,
			PlayoffSeed:    r.PlayoffSeed,
		}
	}
	return out, nil
}

// GetSeasonStatLeaders returns the six title-leader categories for a season.
func (a *App) GetSeasonStatLeaders(seasonID int64) (StatLeadersDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return StatLeadersDTO{}, err
	}
	l, err := a.seasonQueryStore.GetSeasonStatLeaders(a.ctx, seasonID)
	if err != nil {
		return StatLeadersDTO{}, err
	}
	toDTO := func(sl *models.StatLeader) *StatLeaderDTO {
		if sl == nil {
			return nil
		}
		return &StatLeaderDTO{
			PlayerID:  sl.PlayerID,
			FirstName: sl.FirstName,
			LastName:  sl.LastName,
			TeamName:  sl.TeamName,
			StatValue: sl.StatValue,
		}
	}
	return StatLeadersDTO{
		SeasonNum:  l.SeasonNum,
		BA:         toDTO(l.BA),
		HR:         toDTO(l.HR),
		RBI:        toDTO(l.RBI),
		ERA:        toDTO(l.ERA),
		Wins:       toDTO(l.Wins),
		Strikeouts: toDTO(l.Strikeouts),
	}, nil
}

// GetCareerLeaders returns the top-5 all-time career leaders for each category.
func (a *App) GetCareerLeaders() (CareerLeadersDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return CareerLeadersDTO{}, err
	}
	cl, err := a.seasonQueryStore.GetCareerLeaders(a.ctx)
	if err != nil {
		return CareerLeadersDTO{}, err
	}
	toSlice := func(rows []models.CareerLeaderRow) []CareerLeaderDTO {
		out := make([]CareerLeaderDTO, len(rows))
		for i, r := range rows {
			out[i] = CareerLeaderDTO{
				PlayerID:      r.PlayerID,
				FirstName:     r.FirstName,
				LastName:      r.LastName,
				StatValue:     r.StatValue,
				SeasonsPlayed: r.SeasonsPlayed,
			}
		}
		return out
	}
	return CareerLeadersDTO{
		HR:         toSlice(cl.HR),
		Hits:       toSlice(cl.Hits),
		RBI:        toSlice(cl.RBI),
		Wins:       toSlice(cl.Wins),
		Strikeouts: toSlice(cl.Strikeouts),
		Saves:      toSlice(cl.Saves),
	}, nil
}

// SearchPlayers returns up to 50 players matching the query string.
func (a *App) SearchPlayers(query string) ([]PlayerSearchResultDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	results, err := a.playerQueryStore.SearchPlayers(a.ctx, query)
	if err != nil {
		return nil, err
	}
	out := make([]PlayerSearchResultDTO, len(results))
	for i, r := range results {
		out[i] = PlayerSearchResultDTO{
			PlayerID:      r.PlayerID,
			FirstName:     r.FirstName,
			LastName:      r.LastName,
			IsHallOfFamer: r.IsHallOfFamer,
			SeasonsPlayed: r.SeasonsPlayed,
			FirstSeason:   r.FirstSeason,
			LastSeason:    r.LastSeason,
		}
	}
	return out, nil
}

// SearchTeams returns up to 50 teams matching the query string.
func (a *App) SearchTeams(query string) ([]TeamSearchResultDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	results, err := a.teamQueryStore.SearchTeams(a.ctx, query)
	if err != nil {
		return nil, err
	}
	out := make([]TeamSearchResultDTO, len(results))
	for i, r := range results {
		out[i] = TeamSearchResultDTO{
			TeamID:      r.TeamID,
			TeamName:    r.TeamName,
			Seasons:     r.Seasons,
			FirstSeason: r.FirstSeason,
			LastSeason:  r.LastSeason,
		}
	}
	return out, nil
}

// GetPlayerCareer returns a player's bio and career regular-season totals.
// Rate stats are read from pre-computed career tables — no on-read computation.
func (a *App) GetPlayerCareer(playerID int64) (PlayerCareerDTO, error) {
	slog.Debug("GetPlayerCareer", "playerID", playerID)
	if err := a.requireCompanionDB(); err != nil {
		return PlayerCareerDTO{}, err
	}
	career, err := a.playerQueryStore.GetPlayerCareer(a.ctx, playerID)
	if err != nil {
		slog.Error("GetPlayerCareer", "playerID", playerID, "err", err)
		return PlayerCareerDTO{}, err
	}
	return PlayerCareerDTO{
		PlayerID:      career.PlayerID,
		FirstName:     career.FirstName,
		LastName:      career.LastName,
		IsHallOfFamer: career.IsHallOfFamer,
		Batting:       battingToDTO(career.Batting),
		Pitching:      pitchingToDTO(career.Pitching),
	}, nil
}

// GetPlayerSeasonLog returns a player's season-by-season regular and playoff
// stats. Rate stats are read from stored columns — no on-read computation.
func (a *App) GetPlayerSeasonLog(playerID int64) ([]PlayerSeasonLogDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	rows, err := a.playerQueryStore.GetPlayerSeasonLog(a.ctx, playerID)
	if err != nil {
		return nil, err
	}
	out := make([]PlayerSeasonLogDTO, len(rows))
	for i, r := range rows {
		teams := make([]TeamRefDTO, len(r.Teams))
		for j, t := range r.Teams {
			teams[j] = TeamRefDTO{TeamID: t.TeamID, TeamHistoryID: t.TeamHistoryID, TeamName: t.TeamName, SortOrder: t.SortOrder}
		}

		var traits []string
		if r.TraitsJSON != "" {
			_ = json.Unmarshal([]byte(r.TraitsJSON), &traits)
		}
		if traits == nil {
			traits = []string{}
		}

		var pitches []string
		if r.PitchesJSON != "" {
			_ = json.Unmarshal([]byte(r.PitchesJSON), &pitches)
		}
		if pitches == nil {
			pitches = []string{}
		}

		out[i] = PlayerSeasonLogDTO{
			SeasonNum:         r.SeasonNum,
			SeasonID:          r.SeasonID,
			Teams:             teams,
			Age:               r.Age,
			Salary:            r.Salary,
			PrimaryPosition:   r.PrimaryPosition,
			SecondaryPosition: r.SecondaryPosition,
			PitcherRole:       r.PitcherRole,
			BatHand:           r.BatHand,
			ThrowHand:         r.ThrowHand,
			ChemistryType:     r.ChemistryType,
			Traits:            traits,
			Pitches:           pitches,
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
			PlayoffBatting:    battingToDTO(r.PlayoffBatting),
			PlayoffPitching:   pitchingToDTO(r.PlayoffPitching),
		}
	}
	return out, nil
}

// GetPlayerAttributeHistory returns one entry per season for the given player,
// carrying raw attribute values, league-wide percentile ranks, and the eagerly
// persisted league averages. Results are ordered by season number ascending.
func (a *App) GetPlayerAttributeHistory(playerID int64) ([]PlayerAttributeSeasonDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	rows, err := a.playerQueryStore.GetPlayerAttributeHistory(a.ctx, playerID)
	if err != nil {
		return nil, err
	}
	out := make([]PlayerAttributeSeasonDTO, len(rows))
	for i, r := range rows {
		out[i] = PlayerAttributeSeasonDTO{
			SeasonNum:       r.SeasonNum,
			SeasonID:        r.SeasonID,
			Power:           r.Power,
			Contact:         r.Contact,
			Speed:           r.Speed,
			Fielding:        r.Fielding,
			Arm:             r.Arm,
			Velocity:        r.Velocity,
			Junk:            r.Junk,
			Accuracy:        r.Accuracy,
			PowerPct:        r.PowerPct,
			ContactPct:      r.ContactPct,
			SpeedPct:        r.SpeedPct,
			FieldingPct:     r.FieldingPct,
			ArmPct:          r.ArmPct,
			VelocityPct:     r.VelocityPct,
			JunkPct:         r.JunkPct,
			AccuracyPct:     r.AccuracyPct,
			PowerPctRole:    r.PowerPctRole,
			ContactPctRole:  r.ContactPctRole,
			SpeedPctRole:    r.SpeedPctRole,
			FieldingPctRole: r.FieldingPctRole,
			LgAvgPower:      r.LgAvgPower,
			LgAvgContact:    r.LgAvgContact,
			LgAvgSpeed:      r.LgAvgSpeed,
			LgAvgFielding:   r.LgAvgFielding,
			LgAvgArm:        r.LgAvgArm,
			LgAvgVelocity:   r.LgAvgVelocity,
			LgAvgJunk:       r.LgAvgJunk,
			LgAvgAccuracy:   r.LgAvgAccuracy,
			RoleAvgPower:    r.RoleAvgPower,
			RoleAvgContact:  r.RoleAvgContact,
			RoleAvgSpeed:    r.RoleAvgSpeed,
			RoleAvgFielding: r.RoleAvgFielding,
			RoleAvgArm:      r.RoleAvgArm,
			RoleAvgVelocity: r.RoleAvgVelocity,
			RoleAvgJunk:     r.RoleAvgJunk,
			RoleAvgAccuracy: r.RoleAvgAccuracy,
		}
	}
	return out, nil
}

// GetTeamHistory returns all seasons played by a team with champion flags.
func (a *App) GetTeamHistory(teamID int64) (TeamHistoryDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return TeamHistoryDTO{}, err
	}
	th, err := a.teamQueryStore.GetTeamHistory(a.ctx, teamID)
	if err != nil {
		return TeamHistoryDTO{}, err
	}
	seasons := make([]TeamSeasonSummaryDTO, len(th.Seasons))
	for i, s := range th.Seasons {
		seasons[i] = teamSeasonSummaryToDTO(s)
	}
	return TeamHistoryDTO{
		TeamID:   th.TeamID,
		GameGUID: th.GameGUID,
		Seasons:  seasons,
	}, nil
}

// GetTeamTopPlayers returns the top 25 all-time players for a team, ranked by
// cumulative smbWAR accumulated while with that team.
func (a *App) GetTeamTopPlayers(teamID int64) ([]TeamTopPlayerDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	players, err := a.teamQueryStore.GetTeamTopPlayers(a.ctx, teamID, 25)
	if err != nil {
		return nil, fmt.Errorf("getting top players for team %d: %w", teamID, err)
	}
	out := make([]TeamTopPlayerDTO, len(players))
	for i, p := range players {
		out[i] = TeamTopPlayerDTO{
			PlayerID:       p.PlayerID,
			FirstName:      p.FirstName,
			LastName:       p.LastName,
			IsHallOfFamer:  p.IsHallOfFamer,
			NumSeasons:     p.NumSeasons,
			SeasonNums:     parseSortedSeasonNums(p.SeasonNumsCSV),
			IsPitcher:      p.IsPitcher,
			Position:       p.PrimaryPosition,
			SmbWARWithTeam: p.TotalSmbWAR,
			AvgOpsPlus:     p.AvgOpsPlus,
			AvgEraPlus:     p.AvgEraPlus,
			Awards:         p.Awards,
		}
	}
	return out, nil
}

// GetTeamSeasonDetail returns the roster, schedule, and playoff results for one
// team season. Rate stats are computed on roster players before returning.
// Only teamHistoryID is required — seasonID is derived from the team summary.
func (a *App) GetTeamSeasonDetail(teamHistoryID int64) (TeamSeasonDetailDTO, error) {
	slog.Debug("GetTeamSeasonDetail", "teamHistoryID", teamHistoryID)
	if err := a.requireCompanionDB(); err != nil {
		return TeamSeasonDetailDTO{}, err
	}

	teamSummary, err := a.teamQueryStore.GetTeamSeasonSummaryByHistoryID(a.ctx, teamHistoryID)
	if err != nil {
		slog.Error("GetTeamSeasonDetail: team summary", "teamHistoryID", teamHistoryID, "err", err)
		return TeamSeasonDetailDTO{}, fmt.Errorf("team summary: %w", err)
	}
	seasonID := teamSummary.SeasonID

	roster, err := a.teamQueryStore.GetTeamSeasonRoster(a.ctx, teamHistoryID)
	if err != nil {
		return TeamSeasonDetailDTO{}, fmt.Errorf("roster: %w", err)
	}

	schedule, err := a.teamQueryStore.GetTeamSeasonSchedule(a.ctx, teamHistoryID, seasonID)
	if err != nil {
		return TeamSeasonDetailDTO{}, fmt.Errorf("schedule: %w", err)
	}

	playoffs, err := a.teamQueryStore.GetTeamSeasonPlayoffSchedule(a.ctx, teamHistoryID, seasonID)
	if err != nil {
		return TeamSeasonDetailDTO{}, fmt.Errorf("playoff schedule: %w", err)
	}

	seriesLength, err := a.seasonQueryStore.GetPlayoffSeriesLength(a.ctx, seasonID)
	if err != nil {
		return TeamSeasonDetailDTO{}, fmt.Errorf("playoff series length: %w", err)
	}

	rosterDTOs := make([]RosterPlayerDTO, len(roster))
	for i, r := range roster {
		rosterDTOs[i] = rosterPlayerToDTO(r)
	}
	scheduleDTOs := make([]ScheduleGameDTO, len(schedule))
	for i, g := range schedule {
		scheduleDTOs[i] = scheduleGameToDTO(g)
	}
	playoffDTOs := make([]PlayoffGameDTO, len(playoffs))
	for i, g := range playoffs {
		playoffDTOs[i] = playoffGameToDTO(g)
	}

	return TeamSeasonDetailDTO{
		Team:                teamSeasonSummaryToDTO(teamSummary),
		Roster:              rosterDTOs,
		Schedule:            scheduleDTOs,
		Playoffs:            playoffDTOs,
		PlayoffSeriesLength: seriesLength,
	}, nil
}

// GetTeamSeasonScheduleByHistoryID returns the regular-season game log for a
// single team season. Used by the win-delta chart's division fan-out: each
// division peer is fetched individually rather than in one batch query.
func (a *App) GetTeamSeasonScheduleByHistoryID(teamHistoryID int64) ([]ScheduleGameDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	teamSummary, err := a.teamQueryStore.GetTeamSeasonSummaryByHistoryID(a.ctx, teamHistoryID)
	if err != nil {
		return nil, fmt.Errorf("team summary: %w", err)
	}
	schedule, err := a.teamQueryStore.GetTeamSeasonSchedule(a.ctx, teamHistoryID, teamSummary.SeasonID)
	if err != nil {
		return nil, fmt.Errorf("schedule: %w", err)
	}
	out := make([]ScheduleGameDTO, len(schedule))
	for i, g := range schedule {
		out[i] = scheduleGameToDTO(g)
	}
	return out, nil
}

// GetHistoricalTeams returns one aggregated row per team for the historical
// teams page, covering the inclusive season range [seasonStart, seasonEnd].
// Results are ordered by total wins descending.
func (a *App) GetHistoricalTeams(seasonStart, seasonEnd int) ([]HistoricalTeamDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	rows, err := a.teamQueryStore.GetHistoricalTeams(a.ctx, seasonStart, seasonEnd)
	if err != nil {
		return nil, err
	}
	out := make([]HistoricalTeamDTO, len(rows))
	for i, r := range rows {
		out[i] = HistoricalTeamDTO{
			TeamID:              r.TeamID,
			TeamName:            r.TeamName,
			HistoryID:           r.HistoryID,
			NumSeasons:          r.NumSeasons,
			FirstSeason:         r.FirstSeason,
			LastSeason:          r.LastSeason,
			Wins:                r.Wins,
			Losses:              r.Losses,
			WinPct:              r.WinPct,
			GamesOver500:        r.GamesOver500,
			PlayoffWins:         r.PlayoffWins,
			PlayoffLosses:       r.PlayoffLosses,
			PlayoffAppearances:  r.PlayoffAppearances,
			DivisionTitles:      r.DivisionTitles,
			ConferenceTitles:    r.ConferenceTitles,
			Championships:       r.Championships,
			ChampionshipDrought: r.ChampionshipDrought,
			RunsFor:             r.RunsFor,
			RunsAgainst:         r.RunsAgainst,
			TotalAB:             r.TotalAB,
			TotalHits:           r.TotalHits,
			TotalHR:             r.TotalHR,
			NumPlayers:          r.NumPlayers,
			NumHoF:              r.NumHoF,
			BA:                  r.BA,
			ERA:                 r.ERA,
		}
	}
	return out, nil
}

// ListAllTeamSeasons returns every team-season for the historical teams page.
func (a *App) ListAllTeamSeasons() ([]TeamSeasonListDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	rows, err := a.teamQueryStore.ListAllTeamSeasons(a.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]TeamSeasonListDTO, len(rows))
	for i, r := range rows {
		out[i] = TeamSeasonListDTO{
			SeasonNum:      r.SeasonNum,
			HistoryID:      r.HistoryID,
			TeamID:         r.TeamID,
			TeamName:       r.TeamName,
			ConferenceName: r.ConferenceName,
			DivisionName:   r.DivisionName,
			Wins:           r.Wins,
			Losses:         r.Losses,
			WinPct:         r.WinPct,
			RunsFor:        r.RunsFor,
			RunsAgainst:    r.RunsAgainst,
			PlayoffSeed:    r.PlayoffSeed,
			PlayoffWins:    r.PlayoffWins,
			PlayoffLosses:  r.PlayoffLosses,
			IsChampion:     r.IsChampion,
		}
	}
	return out, nil
}

// parseSortedSeasonNums converts a comma-separated season number string
// (as returned by GROUP_CONCAT) into a sorted integer slice.
func parseSortedSeasonNums(csv string) []int {
	if csv == "" {
		return []int{}
	}
	parts := strings.Split(csv, ",")
	nums := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err == nil {
			nums = append(nums, n)
		}
	}
	sort.Ints(nums)
	return nums
}
