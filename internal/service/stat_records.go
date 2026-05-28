package service

import (
	"context"
	"sync"

	"smb-tools/internal/store"
)

// StatRecordsService computes and caches franchise-wide stat highlights:
// per-season league leaders and all-time career/single-season counting stat records.
// The cache lives for the duration of the app session and is invalidated whenever
// a new season is imported.
type StatRecordsService struct {
	store  *store.StatRecordQueryStore
	mu     sync.RWMutex
	cached *statHighlightsCache
}

func NewStatRecordsService(s *store.StatRecordQueryStore) *StatRecordsService {
	return &StatRecordsService{store: s}
}

// PlayerSeasonRef identifies a player-season pair (used for single-season record holders).
type PlayerSeasonRef struct {
	PlayerID  int64
	SeasonNum int
}

// statHighlightsCache is the fully computed highlights state for one franchise.
type statHighlightsCache struct {
	// map[seasonNum][statKey][]playerID — regular season league leaders only.
	// TODO: once qualified-player filtering (minimum PA/IP thresholds) is implemented,
	// revisit this to exclude players below the qualification threshold from bold highlights.
	LeagueLeadersBatting  map[int]map[string][]int64
	LeagueLeadersPitching map[int]map[string][]int64

	// map[statKey][]{PlayerID, SeasonNum} — all-time single regular-season record holders.
	SingleSeasonBatting  map[string][]PlayerSeasonRef
	SingleSeasonPitching map[string][]PlayerSeasonRef

	// map[statKey][]playerID — all-time career record holders, RS and PO tracked separately.
	CareerBattingRS  map[string][]int64
	CareerBattingPO  map[string][]int64
	CareerPitchingRS map[string][]int64
	CareerPitchingPO map[string][]int64
}

// Get returns the highlights cache, computing it on the first call.
// Subsequent calls return the cached value without hitting the database.
func (s *StatRecordsService) Get(ctx context.Context) (*statHighlightsCache, error) {
	s.mu.RLock()
	if s.cached != nil {
		defer s.mu.RUnlock()
		return s.cached, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cached != nil {
		return s.cached, nil
	}

	rsBatting, err := s.store.GetBattingCountRows(ctx, true)
	if err != nil {
		return nil, err
	}
	poBatting, err := s.store.GetBattingCountRows(ctx, false)
	if err != nil {
		return nil, err
	}
	rsPitching, err := s.store.GetPitchingCountRows(ctx, true)
	if err != nil {
		return nil, err
	}
	poPitching, err := s.store.GetPitchingCountRows(ctx, false)
	if err != nil {
		return nil, err
	}

	s.cached = &statHighlightsCache{
		LeagueLeadersBatting:  computeBattingLeagueLeaders(rsBatting),
		LeagueLeadersPitching: computePitchingLeagueLeaders(rsPitching),
		SingleSeasonBatting:   computeBattingSingleSeasonRecords(rsBatting),
		SingleSeasonPitching:  computePitchingSingleSeasonRecords(rsPitching),
		CareerBattingRS:       computeBattingCareerRecords(rsBatting),
		CareerBattingPO:       computeBattingCareerRecords(poBatting),
		CareerPitchingRS:      computePitchingCareerRecords(rsPitching),
		CareerPitchingPO:      computePitchingCareerRecords(poPitching),
	}
	return s.cached, nil
}

// Invalidate clears the cached highlights. Call this after any season import so the
// next Get() recomputes from fresh data.
func (s *StatRecordsService) Invalidate() {
	s.mu.Lock()
	s.cached = nil
	s.mu.Unlock()
}

// ── Batting aggregation ───────────────────────────────────────────────────────

// battingStatExtractors maps stat key names to functions that extract a value
// from a BattingCountRow. Keys match the JSON field names in BattingLeaderRowDTO
// and CareerBattingStatsDTO so the frontend can reference them uniformly.
var battingStatExtractors = map[string]func(store.BattingCountRow) int{
	"gamesPlayed": func(r store.BattingCountRow) int { return r.GamesPlayed },
	"atBats":      func(r store.BattingCountRow) int { return r.AtBats },
	"hits":        func(r store.BattingCountRow) int { return r.Hits },
	"doubles":     func(r store.BattingCountRow) int { return r.Doubles },
	"triples":     func(r store.BattingCountRow) int { return r.Triples },
	"homeRuns":    func(r store.BattingCountRow) int { return r.HomeRuns },
	"rbi":         func(r store.BattingCountRow) int { return r.RBI },
	"stolenBases": func(r store.BattingCountRow) int { return r.StolenBases },
	"walks":       func(r store.BattingCountRow) int { return r.Walks },
	"strikeouts":  func(r store.BattingCountRow) int { return r.Strikeouts },
}

// pitchingStatExtractors maps stat key names to functions that extract a value
// from a PitchingCountRow.
var pitchingStatExtractors = map[string]func(store.PitchingCountRow) int{
	"games":        func(r store.PitchingCountRow) int { return r.Games },
	"gamesStarted": func(r store.PitchingCountRow) int { return r.GamesStarted },
	"wins":         func(r store.PitchingCountRow) int { return r.Wins },
	"losses":       func(r store.PitchingCountRow) int { return r.Losses },
	"saves":        func(r store.PitchingCountRow) int { return r.Saves },
	"outsPitched":  func(r store.PitchingCountRow) int { return r.OutsPitched },
	"strikeouts":   func(r store.PitchingCountRow) int { return r.Strikeouts },
	"walks":        func(r store.PitchingCountRow) int { return r.Walks },
	"hitsAllowed":  func(r store.PitchingCountRow) int { return r.HitsAllowed },
	"earnedRuns":   func(r store.PitchingCountRow) int { return r.EarnedRuns },
}

func computeBattingLeagueLeaders(rows []store.BattingCountRow) map[int]map[string][]int64 {
	bySeason := make(map[int][]store.BattingCountRow)
	for _, r := range rows {
		bySeason[r.SeasonNum] = append(bySeason[r.SeasonNum], r)
	}
	out := make(map[int]map[string][]int64, len(bySeason))
	for seasonNum, seasonRows := range bySeason {
		out[seasonNum] = battingSeasonMax(seasonRows)
	}
	return out
}

func battingSeasonMax(rows []store.BattingCountRow) map[string][]int64 {
	leaders := make(map[string][]int64, len(battingStatExtractors))
	for key, fn := range battingStatExtractors {
		maxVal := -1
		for _, r := range rows {
			if v := fn(r); v > maxVal {
				maxVal = v
			}
		}
		if maxVal <= 0 {
			continue
		}
		for _, r := range rows {
			if fn(r) == maxVal {
				leaders[key] = append(leaders[key], r.PlayerID)
			}
		}
	}
	return leaders
}

func computeBattingSingleSeasonRecords(rows []store.BattingCountRow) map[string][]PlayerSeasonRef {
	out := make(map[string][]PlayerSeasonRef, len(battingStatExtractors))
	for key, fn := range battingStatExtractors {
		maxVal := -1
		for _, r := range rows {
			if v := fn(r); v > maxVal {
				maxVal = v
			}
		}
		if maxVal <= 0 {
			continue
		}
		for _, r := range rows {
			if fn(r) == maxVal {
				out[key] = append(out[key], PlayerSeasonRef{PlayerID: r.PlayerID, SeasonNum: r.SeasonNum})
			}
		}
	}
	return out
}

func computeBattingCareerRecords(rows []store.BattingCountRow) map[string][]int64 {
	totals := make(map[int64]map[string]int)
	for _, r := range rows {
		if totals[r.PlayerID] == nil {
			totals[r.PlayerID] = make(map[string]int, len(battingStatExtractors))
		}
		for key, fn := range battingStatExtractors {
			totals[r.PlayerID][key] += fn(r)
		}
	}
	return careerRecordsFromTotals(totals, func(key string) bool { return true })
}

// ── Pitching aggregation ──────────────────────────────────────────────────────

func computePitchingLeagueLeaders(rows []store.PitchingCountRow) map[int]map[string][]int64 {
	bySeason := make(map[int][]store.PitchingCountRow)
	for _, r := range rows {
		bySeason[r.SeasonNum] = append(bySeason[r.SeasonNum], r)
	}
	out := make(map[int]map[string][]int64, len(bySeason))
	for seasonNum, seasonRows := range bySeason {
		out[seasonNum] = pitchingSeasonMax(seasonRows)
	}
	return out
}

func pitchingSeasonMax(rows []store.PitchingCountRow) map[string][]int64 {
	leaders := make(map[string][]int64, len(pitchingStatExtractors))
	for key, fn := range pitchingStatExtractors {
		maxVal := -1
		for _, r := range rows {
			if v := fn(r); v > maxVal {
				maxVal = v
			}
		}
		if maxVal <= 0 {
			continue
		}
		for _, r := range rows {
			if fn(r) == maxVal {
				leaders[key] = append(leaders[key], r.PlayerID)
			}
		}
	}
	return leaders
}

func computePitchingSingleSeasonRecords(rows []store.PitchingCountRow) map[string][]PlayerSeasonRef {
	out := make(map[string][]PlayerSeasonRef, len(pitchingStatExtractors))
	for key, fn := range pitchingStatExtractors {
		maxVal := -1
		for _, r := range rows {
			if v := fn(r); v > maxVal {
				maxVal = v
			}
		}
		if maxVal <= 0 {
			continue
		}
		for _, r := range rows {
			if fn(r) == maxVal {
				out[key] = append(out[key], PlayerSeasonRef{PlayerID: r.PlayerID, SeasonNum: r.SeasonNum})
			}
		}
	}
	return out
}

func computePitchingCareerRecords(rows []store.PitchingCountRow) map[string][]int64 {
	totals := make(map[int64]map[string]int)
	for _, r := range rows {
		if totals[r.PlayerID] == nil {
			totals[r.PlayerID] = make(map[string]int, len(pitchingStatExtractors))
		}
		for key, fn := range pitchingStatExtractors {
			totals[r.PlayerID][key] += fn(r)
		}
	}
	return careerRecordsFromTotals(totals, func(key string) bool { return true })
}

// ── Shared helpers ────────────────────────────────────────────────────────────

// careerRecordsFromTotals finds the all-time max per stat key from pre-summed
// player career totals and returns a map of statKey → []playerID (ties included).
func careerRecordsFromTotals(totals map[int64]map[string]int, include func(string) bool) map[string][]int64 {
	if len(totals) == 0 {
		return nil
	}

	// Collect all stat keys from any player's entry.
	allKeys := make(map[string]struct{})
	for _, t := range totals {
		for k := range t {
			allKeys[k] = struct{}{}
		}
	}

	records := make(map[string][]int64, len(allKeys))
	for key := range allKeys {
		if !include(key) {
			continue
		}
		maxVal := -1
		for _, t := range totals {
			if v := t[key]; v > maxVal {
				maxVal = v
			}
		}
		if maxVal <= 0 {
			continue
		}
		for pid, t := range totals {
			if t[key] == maxVal {
				records[key] = append(records[key], pid)
			}
		}
	}
	return records
}
