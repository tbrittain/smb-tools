package main

import (
	"fmt"
	"log/slog"

	"smb-tools/internal/models"
)

// ListAwards returns all award definitions filtered by the playoff flag.
func (a *App) ListAwards(isPlayoff bool) ([]AwardDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	awards, err := a.awardStore.ListAwards(a.ctx, isPlayoff)
	if err != nil {
		return nil, err
	}
	out := make([]AwardDTO, len(awards))
	for i, aw := range awards {
		out[i] = awardToDTO(aw)
	}
	return out, nil
}

// ListAllAwards returns all award definitions regardless of playoff flag.
func (a *App) ListAllAwards() ([]AwardDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	awards, err := a.awardStore.ListAllAwards(a.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]AwardDTO, len(awards))
	for i, aw := range awards {
		out[i] = awardToDTO(aw)
	}
	return out, nil
}

// CreateCustomAward creates a user-defined award and returns it with its new ID.
func (a *App) CreateCustomAward(dto AwardDTO) (AwardDTO, error) {
	slog.Info("CreateCustomAward", "name", dto.Name)
	if err := a.requireCompanionDB(); err != nil {
		return AwardDTO{}, err
	}
	m := models.Award{
		Name: dto.Name, OriginalName: dto.OriginalName,
		Importance: dto.Importance, OmitFromGroupings: dto.OmitFromGroupings,
		IsBattingAward: dto.IsBattingAward, IsPitchingAward: dto.IsPitchingAward,
		IsFieldingAward: dto.IsFieldingAward, IsPlayoffAward: dto.IsPlayoffAward,
		IsUserAssignable: dto.IsUserAssignable,
	}
	id, err := a.awardStore.CreateCustomAward(a.ctx, m)
	if err != nil {
		slog.Error("CreateCustomAward: failed", "name", dto.Name, "err", err)
		return AwardDTO{}, err
	}
	slog.Info("CreateCustomAward: created", "id", id)
	dto.ID = id
	dto.IsBuiltIn = false
	return dto, nil
}

// GetSeasonPlayerAwards returns all player-seasons for a season with their awards.
func (a *App) GetSeasonPlayerAwards(seasonID int64) ([]SeasonPlayerAwardRowDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	rows, err := a.awardStore.GetSeasonPlayerAwards(a.ctx, seasonID)
	if err != nil {
		return nil, err
	}
	out := make([]SeasonPlayerAwardRowDTO, len(rows))
	for i, r := range rows {
		awards := make([]AwardDTO, len(r.Awards))
		for j, aw := range r.Awards {
			awards[j] = awardToDTO(aw)
		}
		out[i] = SeasonPlayerAwardRowDTO{
			PlayerSeasonID: r.PlayerSeasonID,
			PlayerID:       r.PlayerID,
			FirstName:      r.FirstName,
			LastName:       r.LastName,
			TeamName:       r.TeamName,
			PrimaryPos:     r.PrimaryPos,
			PitcherRole:    r.PitcherRole,
			Awards:         awards,
		}
	}
	return out, nil
}

// GetPlayerCareerAwards returns awards grouped by season number for a player.
func (a *App) GetPlayerCareerAwards(playerID int64) (map[string][]AwardDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	bySeasonNum, err := a.awardStore.GetPlayerCareerAwards(a.ctx, playerID)
	if err != nil {
		return nil, err
	}
	out := make(map[string][]AwardDTO, len(bySeasonNum))
	for sn, awards := range bySeasonNum {
		key := fmt.Sprintf("%d", sn)
		dtos := make([]AwardDTO, len(awards))
		for i, aw := range awards {
			dtos[i] = awardToDTO(aw)
		}
		out[key] = dtos
	}
	return out, nil
}

// SetPlayerSeasonAwards replaces the user-assignable awards for one player-season.
func (a *App) SetPlayerSeasonAwards(req SetPlayerAwardsRequestDTO) error {
	if err := a.requireCompanionDB(); err != nil {
		return err
	}
	return a.awardStore.SetPlayerSeasonAwards(a.ctx, req.PlayerSeasonID, req.AwardIDs)
}

// ComputeSeasonStatLeaderAwards computes and stores auto-calculated stat title
// awards (BA, HR, RBI, ERA, W, K, Triple Crown) for the given season.
func (a *App) ComputeSeasonStatLeaderAwards(seasonID int64) error {
	if err := a.requireCompanionDB(); err != nil {
		return err
	}
	return a.awardStore.ComputeAndAssignStatLeaderAwards(a.ctx, seasonID)
}

// GetSeasonAwardSummary returns personal-performance awards delegated for the
// given season, grouped by award type, for display in read-only view mode.
// Championship and team awards are excluded.
func (a *App) GetSeasonAwardSummary(seasonID int64) (SeasonAwardSummaryDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return SeasonAwardSummaryDTO{}, err
	}
	summary, err := a.awardStore.GetSeasonAwardSummary(a.ctx, seasonID)
	if err != nil {
		return SeasonAwardSummaryDTO{}, err
	}
	return seasonAwardSummaryToDTO(summary), nil
}

// GetSeasonChampionTeamHistoryID returns the team_season_history_id of the
// playoff champion for the season, or nil if not yet determinable.
func (a *App) GetSeasonChampionTeamHistoryID(seasonID int64) (*int64, error) {
	if err := a.requireCompanionDB(); err != nil {
		return nil, err
	}
	return a.awardStore.GetSeasonChampionTeam(a.ctx, seasonID)
}

// GetHoFCandidates returns a paginated list of retired players eligible for Hall
// of Fame induction. page is 1-based; pageSize controls rows per page;
// lastSeasons limits results to players whose last season falls within the past
// lastSeasons seasons.
func (a *App) GetHoFCandidates(page, pageSize, lastSeasons int) (HoFPageDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return HoFPageDTO{}, err
	}
	result, err := a.awardStore.GetHoFCandidates(a.ctx, page, pageSize, lastSeasons)
	if err != nil {
		return HoFPageDTO{}, err
	}
	items := make([]HoFCandidateDTO, len(result.Items))
	for i, c := range result.Items {
		items[i] = hofCandidateToDTO(c)
	}
	return HoFPageDTO{Items: items, Total: result.Total}, nil
}

// GetHoFInducted returns a paginated list of current Hall of Fame members,
// filtered and paginated with the same semantics as GetHoFCandidates.
func (a *App) GetHoFInducted(page, pageSize, lastSeasons int) (HoFPageDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return HoFPageDTO{}, err
	}
	result, err := a.awardStore.GetHoFInducted(a.ctx, page, pageSize, lastSeasons)
	if err != nil {
		return HoFPageDTO{}, err
	}
	items := make([]HoFCandidateDTO, len(result.Items))
	for i, c := range result.Items {
		items[i] = hofCandidateToDTO(c)
	}
	return HoFPageDTO{Items: items, Total: result.Total}, nil
}

// GetSeasonAwardCandidates returns all award delegation candidate groups for a season:
// top batters/pitchers overall, rookies, by team, and by position. Award IDs are
// pre-populated from existing assignments, or auto-suggested if none exist yet.
func (a *App) GetSeasonAwardCandidates(seasonID int64) (SeasonAwardCandidatesDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return SeasonAwardCandidatesDTO{}, err
	}
	m, err := a.awardStore.GetSeasonAwardCandidates(a.ctx, seasonID)
	if err != nil {
		return SeasonAwardCandidatesDTO{}, err
	}
	return seasonAwardCandidatesToDTO(m), nil
}

// SubmitSeasonAwards replaces user-assignable awards for all specified player-seasons
// in a single transaction. Players omitted from the request are not modified.
func (a *App) SubmitSeasonAwards(req SubmitSeasonAwardsDTO) error {
	slog.Info("SubmitSeasonAwards", "players", len(req.PlayerAwards))
	if err := a.requireCompanionDB(); err != nil {
		return err
	}
	entries := make([]models.PlayerAwardEntry, len(req.PlayerAwards))
	for i, e := range req.PlayerAwards {
		entries[i] = models.PlayerAwardEntry{
			PlayerSeasonID: e.PlayerSeasonID,
			AwardIDs:       e.AwardIDs,
		}
	}
	if err := a.awardStore.SubmitMultiplePlayerAwards(a.ctx, entries); err != nil {
		slog.Error("SubmitSeasonAwards: failed", "err", err)
		return err
	}
	slog.Info("SubmitSeasonAwards: complete")
	return nil
}

// SetHallOfFamer updates the Hall of Fame status for a player.
func (a *App) SetHallOfFamer(playerID int64, isHoF bool) error {
	if err := a.requireCompanionDB(); err != nil {
		return err
	}
	return a.playerQueryStore.SetHallOfFamer(a.ctx, playerID, isHoF)
}
