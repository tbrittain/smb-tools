package main

import (
	"strconv"
)

// GetBattingCareerLeaders returns a paginated page of career batting totals for
// players matching the given filters. Rate stats are computed inline by the store.
func (a *App) GetBattingCareerLeaders(filters LeaderboardFiltersDTO) (BattingLeaderPageDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return BattingLeaderPageDTO{}, err
	}
	rows, total, err := a.leaderboardQueryStore.GetBattingCareerLeaders(a.ctx, leaderboardFiltersToDomain(filters))
	if err != nil {
		return BattingLeaderPageDTO{}, err
	}
	out := make([]BattingLeaderRowDTO, len(rows))
	for i := range rows {
		out[i] = battingCareerLeaderToDTO(rows[i])
	}
	return BattingLeaderPageDTO{Rows: out, Total: total}, nil
}

// GetBattingSeasonLeaders returns a paginated page of per-season batting stats
// matching the given filters. Rate stats are read from stored columns.
func (a *App) GetBattingSeasonLeaders(filters LeaderboardFiltersDTO) (BattingLeaderPageDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return BattingLeaderPageDTO{}, err
	}
	rows, total, err := a.leaderboardQueryStore.GetBattingSeasonLeaders(a.ctx, leaderboardFiltersToDomain(filters))
	if err != nil {
		return BattingLeaderPageDTO{}, err
	}
	out := make([]BattingLeaderRowDTO, len(rows))
	for i := range rows {
		out[i] = battingSeasonLeaderToDTO(rows[i])
	}
	return BattingLeaderPageDTO{Rows: out, Total: total}, nil
}

// GetPitchingCareerLeaders returns a paginated page of career pitching totals for
// players matching the given filters. Rate stats are computed inline by the store.
func (a *App) GetPitchingCareerLeaders(filters LeaderboardFiltersDTO) (PitchingLeaderPageDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return PitchingLeaderPageDTO{}, err
	}
	rows, total, err := a.leaderboardQueryStore.GetPitchingCareerLeaders(a.ctx, leaderboardFiltersToDomain(filters))
	if err != nil {
		return PitchingLeaderPageDTO{}, err
	}
	out := make([]PitchingLeaderRowDTO, len(rows))
	for i := range rows {
		out[i] = pitchingCareerLeaderToDTO(rows[i])
	}
	return PitchingLeaderPageDTO{Rows: out, Total: total}, nil
}

// GetPitchingSeasonLeaders returns a paginated page of per-season pitching stats
// matching the given filters. Rate stats are read from stored columns.
func (a *App) GetPitchingSeasonLeaders(filters LeaderboardFiltersDTO) (PitchingLeaderPageDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return PitchingLeaderPageDTO{}, err
	}
	rows, total, err := a.leaderboardQueryStore.GetPitchingSeasonLeaders(a.ctx, leaderboardFiltersToDomain(filters))
	if err != nil {
		return PitchingLeaderPageDTO{}, err
	}
	out := make([]PitchingLeaderRowDTO, len(rows))
	for i := range rows {
		out[i] = pitchingSeasonLeaderToDTO(rows[i])
	}
	return PitchingLeaderPageDTO{Rows: out, Total: total}, nil
}

// GetStatHighlights returns the franchise-wide stat highlight data: per-season
// league leaders and all-time counting-stat records. The result is computed once
// per session and cached until a new season import invalidates it.
func (a *App) GetStatHighlights() (StatHighlightsDTO, error) {
	if err := a.requireCompanionDB(); err != nil {
		return StatHighlightsDTO{}, err
	}
	cache, err := a.statRecordsService.Get(a.ctx)
	if err != nil {
		return StatHighlightsDTO{}, err
	}

	leadersBatting := make(map[string]map[string][]int64, len(cache.LeagueLeadersBatting))
	for seasonNum, stats := range cache.LeagueLeadersBatting {
		leadersBatting[strconv.Itoa(seasonNum)] = stats
	}
	leadersPitching := make(map[string]map[string][]int64, len(cache.LeagueLeadersPitching))
	for seasonNum, stats := range cache.LeagueLeadersPitching {
		leadersPitching[strconv.Itoa(seasonNum)] = stats
	}

	singleBatting := make(map[string][]StatRecordHolderDTO, len(cache.SingleSeasonBatting))
	for key, refs := range cache.SingleSeasonBatting {
		holders := make([]StatRecordHolderDTO, len(refs))
		for i, r := range refs {
			holders[i] = StatRecordHolderDTO{PlayerID: r.PlayerID, SeasonNum: r.SeasonNum}
		}
		singleBatting[key] = holders
	}
	singlePitching := make(map[string][]StatRecordHolderDTO, len(cache.SingleSeasonPitching))
	for key, refs := range cache.SingleSeasonPitching {
		holders := make([]StatRecordHolderDTO, len(refs))
		for i, r := range refs {
			holders[i] = StatRecordHolderDTO{PlayerID: r.PlayerID, SeasonNum: r.SeasonNum}
		}
		singlePitching[key] = holders
	}

	leadersRateBatting := make(map[string]map[string][]int64, len(cache.LeagueLeadersBattingRate))
	for seasonNum, stats := range cache.LeagueLeadersBattingRate {
		leadersRateBatting[strconv.Itoa(seasonNum)] = stats
	}
	leadersRatePitching := make(map[string]map[string][]int64, len(cache.LeagueLeadersPitchingRate))
	for seasonNum, stats := range cache.LeagueLeadersPitchingRate {
		leadersRatePitching[strconv.Itoa(seasonNum)] = stats
	}

	singleRateBatting := make(map[string][]StatRecordHolderDTO, len(cache.SingleSeasonBattingRate))
	for key, refs := range cache.SingleSeasonBattingRate {
		holders := make([]StatRecordHolderDTO, len(refs))
		for i, r := range refs {
			holders[i] = StatRecordHolderDTO{PlayerID: r.PlayerID, SeasonNum: r.SeasonNum}
		}
		singleRateBatting[key] = holders
	}
	singleRatePitching := make(map[string][]StatRecordHolderDTO, len(cache.SingleSeasonPitchingRate))
	for key, refs := range cache.SingleSeasonPitchingRate {
		holders := make([]StatRecordHolderDTO, len(refs))
		for i, r := range refs {
			holders[i] = StatRecordHolderDTO{PlayerID: r.PlayerID, SeasonNum: r.SeasonNum}
		}
		singleRatePitching[key] = holders
	}

	return StatHighlightsDTO{
		LeagueLeadersBatting:  leadersBatting,
		LeagueLeadersPitching: leadersPitching,
		SingleSeasonBatting:   singleBatting,
		SingleSeasonPitching:  singlePitching,
		CareerBattingRS:       cache.CareerBattingRS,
		CareerBattingPO:       cache.CareerBattingPO,
		CareerPitchingRS:      cache.CareerPitchingRS,
		CareerPitchingPO:      cache.CareerPitchingPO,

		LeagueLeadersBattingRate:  leadersRateBatting,
		LeagueLeadersPitchingRate: leadersRatePitching,
		SingleSeasonBattingRate:   singleRateBatting,
		SingleSeasonPitchingRate:  singleRatePitching,
		CareerBattingRSRate:       cache.CareerBattingRSRate,
		CareerBattingPORate:       cache.CareerBattingPORate,
		CareerPitchingRSRate:      cache.CareerPitchingRSRate,
		CareerPitchingPORate:      cache.CareerPitchingPORate,
	}, nil
}

