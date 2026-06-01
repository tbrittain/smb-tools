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

	// Rate stat equivalents of the above — qualified players only.
	// Stat keys match the JSON field names on the frontend DTOs (e.g. "ba", "era", "k9").
	LeagueLeadersBattingRate  map[int]map[string][]int64
	LeagueLeadersPitchingRate map[int]map[string][]int64
	SingleSeasonBattingRate   map[string][]PlayerSeasonRef
	SingleSeasonPitchingRate  map[string][]PlayerSeasonRef
	CareerBattingRSRate       map[string][]int64
	CareerBattingPORate       map[string][]int64
	CareerPitchingRSRate      map[string][]int64
	CareerPitchingPORate      map[string][]int64
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

	seasonLength, err := s.store.GetFranchiseSeasonLength(ctx)
	if err != nil {
		return nil, err
	}
	// Scale Baseball Reference's 3000 PA / 1000 IP career thresholds by season length.
	// When no seasons exist yet seasonLength is 0; thresholds stay 0 so no one qualifies.
	careerBattingPAThreshold := int(3000 * float64(seasonLength) / 162)
	careerPitchingOutsThreshold := int(3000 * float64(seasonLength) / 162)

	rsBattingRate, err := s.store.GetBattingRateRows(ctx, true)
	if err != nil {
		return nil, err
	}
	rsPitchingRate, err := s.store.GetPitchingRateRows(ctx, true)
	if err != nil {
		return nil, err
	}

	careerBattingRS, err := s.store.GetCareerBattingRateRows(ctx, "regular_season")
	if err != nil {
		return nil, err
	}
	careerBattingPO, err := s.store.GetCareerBattingRateRows(ctx, "playoffs")
	if err != nil {
		return nil, err
	}
	careerPitchingRS, err := s.store.GetCareerPitchingRateRows(ctx, "regular_season")
	if err != nil {
		return nil, err
	}
	careerPitchingPO, err := s.store.GetCareerPitchingRateRows(ctx, "playoffs")
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

		LeagueLeadersBattingRate:  computeBattingRateLeagueLeaders(rsBattingRate),
		LeagueLeadersPitchingRate: computePitchingRateLeagueLeaders(rsPitchingRate),
		SingleSeasonBattingRate:   computeBattingRateSingleSeasonRecords(rsBattingRate),
		SingleSeasonPitchingRate:  computePitchingRateSingleSeasonRecords(rsPitchingRate),
		CareerBattingRSRate:       computeBattingCareerRateRecords(careerBattingRS, careerBattingPAThreshold),
		CareerBattingPORate:       computeBattingCareerRateRecords(careerBattingPO, careerBattingPAThreshold),
		CareerPitchingRSRate:      computePitchingCareerRateRecords(careerPitchingRS, careerPitchingOutsThreshold),
		CareerPitchingPORate:      computePitchingCareerRateRecords(careerPitchingPO, careerPitchingOutsThreshold),
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

// battingCountHigherIsBetter maps each batting counting stat key to true when a higher
// value is better (HR, H, …) or false when lower is better (strikeouts for a batter).
var battingCountHigherIsBetter = map[string]bool{
	"gamesPlayed": true,
	"atBats":      true,
	"hits":        true,
	"doubles":     true,
	"triples":     true,
	"homeRuns":    true,
	"rbi":         true,
	"stolenBases": true,
	"walks":       true,
	"strikeouts":  false, // fewer batter strikeouts = better
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

// pitchingCountHigherIsBetter maps each pitching counting stat key to true when higher
// is better (W, K, …) or false when lower is better (BB, H, ER allowed).
var pitchingCountHigherIsBetter = map[string]bool{
	"games":        true,
	"gamesStarted": true,
	"wins":         true,
	"losses":       true, // most losses = record holder (a dubious distinction, but consistent)
	"saves":        true,
	"outsPitched":  true,
	"strikeouts":   true,
	"walks":        false, // fewer walks allowed = better
	"hitsAllowed":  false, // fewer hits allowed = better
	"earnedRuns":   false, // fewer earned runs = better
}

// isBetterCount returns true if v beats best in the given direction.
func isBetterCount(v, best int, higherIsBetter bool) bool {
	if higherIsBetter {
		return v > best
	}
	return v < best
}

func computeBattingLeagueLeaders(rows []store.BattingCountRow) map[int]map[string][]int64 {
	bySeason := make(map[int][]store.BattingCountRow)
	for _, r := range rows {
		bySeason[r.SeasonNum] = append(bySeason[r.SeasonNum], r)
	}
	out := make(map[int]map[string][]int64, len(bySeason))
	for seasonNum, seasonRows := range bySeason {
		out[seasonNum] = battingSeasonBest(seasonRows)
	}
	return out
}

func battingSeasonBest(rows []store.BattingCountRow) map[string][]int64 {
	leaders := make(map[string][]int64, len(battingStatExtractors))
	for key, fn := range battingStatExtractors {
		higherIsBetter := battingCountHigherIsBetter[key]
		hasBest := false
		var best int
		for _, r := range rows {
			// Lower-is-better counting stats require the same PA threshold as rate stats
			// to prevent unqualified players (e.g. pitchers with 1 AB) from winning.
			if !higherIsBetter && float64(r.PlateAppearances) < float64(r.NumGames)*3.1 {
				continue
			}
			v := fn(r)
			if !hasBest || isBetterCount(v, best, higherIsBetter) {
				best = v
				hasBest = true
			}
		}
		if !hasBest {
			continue
		}
		if higherIsBetter && best <= 0 {
			continue
		}
		for _, r := range rows {
			if !higherIsBetter && float64(r.PlateAppearances) < float64(r.NumGames)*3.1 {
				continue
			}
			if fn(r) == best {
				leaders[key] = append(leaders[key], r.PlayerID)
			}
		}
	}
	return leaders
}

func computeBattingSingleSeasonRecords(rows []store.BattingCountRow) map[string][]PlayerSeasonRef {
	out := make(map[string][]PlayerSeasonRef, len(battingStatExtractors))
	for key, fn := range battingStatExtractors {
		higherIsBetter := battingCountHigherIsBetter[key]
		hasBest := false
		var best int
		for _, r := range rows {
			if !higherIsBetter && float64(r.PlateAppearances) < float64(r.NumGames)*3.1 {
				continue
			}
			v := fn(r)
			if !hasBest || isBetterCount(v, best, higherIsBetter) {
				best = v
				hasBest = true
			}
		}
		if !hasBest {
			continue
		}
		if higherIsBetter && best <= 0 {
			continue
		}
		for _, r := range rows {
			if !higherIsBetter && float64(r.PlateAppearances) < float64(r.NumGames)*3.1 {
				continue
			}
			if fn(r) == best {
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
	// For lower-is-better stats, exclude players with 0 career AB (never batted).
	qualifies := func(_ int64, t map[string]int) bool { return t["atBats"] > 0 }
	return careerRecordsFromTotals(totals, battingCountHigherIsBetter, qualifies)
}

// ── Pitching aggregation ──────────────────────────────────────────────────────

func computePitchingLeagueLeaders(rows []store.PitchingCountRow) map[int]map[string][]int64 {
	bySeason := make(map[int][]store.PitchingCountRow)
	for _, r := range rows {
		bySeason[r.SeasonNum] = append(bySeason[r.SeasonNum], r)
	}
	out := make(map[int]map[string][]int64, len(bySeason))
	for seasonNum, seasonRows := range bySeason {
		out[seasonNum] = pitchingSeasonBest(seasonRows)
	}
	return out
}

func pitchingSeasonBest(rows []store.PitchingCountRow) map[string][]int64 {
	leaders := make(map[string][]int64, len(pitchingStatExtractors))
	for key, fn := range pitchingStatExtractors {
		higherIsBetter := pitchingCountHigherIsBetter[key]
		hasBest := false
		var best int
		for _, r := range rows {
			// Lower-is-better counting stats (H, ER, BB allowed) require IP qualification
			// to prevent non-starting pitchers from winning on trivially small samples.
			if !higherIsBetter && r.OutsPitched < r.NumGames*3 {
				continue
			}
			v := fn(r)
			if !hasBest || isBetterCount(v, best, higherIsBetter) {
				best = v
				hasBest = true
			}
		}
		if !hasBest {
			continue
		}
		if higherIsBetter && best <= 0 {
			continue
		}
		for _, r := range rows {
			if !higherIsBetter && r.OutsPitched < r.NumGames*3 {
				continue
			}
			if fn(r) == best {
				leaders[key] = append(leaders[key], r.PlayerID)
			}
		}
	}
	return leaders
}

func computePitchingSingleSeasonRecords(rows []store.PitchingCountRow) map[string][]PlayerSeasonRef {
	out := make(map[string][]PlayerSeasonRef, len(pitchingStatExtractors))
	for key, fn := range pitchingStatExtractors {
		higherIsBetter := pitchingCountHigherIsBetter[key]
		hasBest := false
		var best int
		for _, r := range rows {
			if !higherIsBetter && r.OutsPitched < r.NumGames*3 {
				continue
			}
			v := fn(r)
			if !hasBest || isBetterCount(v, best, higherIsBetter) {
				best = v
				hasBest = true
			}
		}
		if !hasBest {
			continue
		}
		if higherIsBetter && best <= 0 {
			continue
		}
		for _, r := range rows {
			if !higherIsBetter && r.OutsPitched < r.NumGames*3 {
				continue
			}
			if fn(r) == best {
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
	// For lower-is-better stats, exclude players with 0 career outs pitched (never pitched).
	qualifies := func(_ int64, t map[string]int) bool { return t["outsPitched"] > 0 }
	return careerRecordsFromTotals(totals, pitchingCountHigherIsBetter, qualifies)
}

// ── Rate stat extractors and direction ────────────────────────────────────────

// battingRateExtractors maps stat key strings (matching frontend DTO JSON keys) to
// functions that extract the corresponding *float64 from a BattingRateRow.
// Includes OPS+ for season-only tracking (not included in career extractors).
var battingRateExtractors = map[string]func(store.BattingRateRow) *float64{
	"ba":      func(r store.BattingRateRow) *float64 { return r.BA },
	"obp":     func(r store.BattingRateRow) *float64 { return r.OBP },
	"slg":     func(r store.BattingRateRow) *float64 { return r.SLG },
	"ops":     func(r store.BattingRateRow) *float64 { return r.OPS },
	"iso":     func(r store.BattingRateRow) *float64 { return r.ISO },
	"babip":   func(r store.BattingRateRow) *float64 { return r.BABIP },
	"kPct":    func(r store.BattingRateRow) *float64 { return r.KPct },
	"bbPct":   func(r store.BattingRateRow) *float64 { return r.BBPct },
	"abPerHr": func(r store.BattingRateRow) *float64 { return r.ABPerHR },
	"opsPlus": func(r store.BattingRateRow) *float64 { return r.OPSPlus },
	"smbWar":  func(r store.BattingRateRow) *float64 { return r.SmbWAR },
}

// battingRateHigherIsBetter maps each batting rate stat key to true when a higher
// value is better (BA, OBP, …) or false when lower is better (kPct, abPerHr).
var battingRateHigherIsBetter = map[string]bool{
	"ba":      true,
	"obp":     true,
	"slg":     true,
	"ops":     true,
	"iso":     true,
	"babip":   true,
	"kPct":    false, // fewer batter strikeouts = better
	"bbPct":   true,
	"abPerHr": false, // fewer AB per HR = more power = better
	"opsPlus": true,  // higher OPS+ = better (100 = league average)
	"smbWar":  true,
}

// pitchingRateExtractors for season-level tracking. Includes ERA+ and FIP- which
// are league-adjusted and tracked at season granularity only (not career).
var pitchingRateExtractors = map[string]func(store.PitchingRateRow) *float64{
	"era":     func(r store.PitchingRateRow) *float64 { return r.ERA },
	"whip":    func(r store.PitchingRateRow) *float64 { return r.WHIP },
	"k9":      func(r store.PitchingRateRow) *float64 { return r.K9 },
	"bb9":     func(r store.PitchingRateRow) *float64 { return r.BB9 },
	"h9":      func(r store.PitchingRateRow) *float64 { return r.H9 },
	"hr9":     func(r store.PitchingRateRow) *float64 { return r.HR9 },
	"kPerBb":  func(r store.PitchingRateRow) *float64 { return r.KPerBB },
	"kPct":    func(r store.PitchingRateRow) *float64 { return r.KPct },
	"winPct":  func(r store.PitchingRateRow) *float64 { return r.WinPct },
	"pPerIp":  func(r store.PitchingRateRow) *float64 { return r.PPerIP },
	"fip":     func(r store.PitchingRateRow) *float64 { return r.FIP },
	"eraPlus": func(r store.PitchingRateRow) *float64 { return r.ERAPlus },
	"fipMinus": func(r store.PitchingRateRow) *float64 { return r.FIPMinus },
	"smbWar":  func(r store.PitchingRateRow) *float64 { return r.SmbWAR },
}

var pitchingRateHigherIsBetter = map[string]bool{
	"era":     false,
	"whip":    false,
	"k9":      true,
	"bb9":     false,
	"h9":      false,
	"hr9":     false,
	"kPerBb":  true,
	"kPct":    true,
	"winPct":  true,
	"pPerIp":  false,
	"fip":     false,
	"eraPlus": true,   // higher ERA+ = better (100 = league average)
	"fipMinus": false, // lower FIP- = better (100 = league average)
	"smbWar":  true,
}

var battingCareerRateExtractors = map[string]func(store.BattingCareerRateRow) *float64{
	"ba":      func(r store.BattingCareerRateRow) *float64 { return r.BA },
	"obp":     func(r store.BattingCareerRateRow) *float64 { return r.OBP },
	"slg":     func(r store.BattingCareerRateRow) *float64 { return r.SLG },
	"ops":     func(r store.BattingCareerRateRow) *float64 { return r.OPS },
	"iso":     func(r store.BattingCareerRateRow) *float64 { return r.ISO },
	"babip":   func(r store.BattingCareerRateRow) *float64 { return r.BABIP },
	"kPct":    func(r store.BattingCareerRateRow) *float64 { return r.KPct },
	"bbPct":   func(r store.BattingCareerRateRow) *float64 { return r.BBPct },
	"abPerHr": func(r store.BattingCareerRateRow) *float64 { return r.ABPerHR },
	"smbWar":  func(r store.BattingCareerRateRow) *float64 { return r.SmbWAR },
}

var pitchingCareerRateExtractors = map[string]func(store.PitchingCareerRateRow) *float64{
	"era":    func(r store.PitchingCareerRateRow) *float64 { return r.ERA },
	"whip":   func(r store.PitchingCareerRateRow) *float64 { return r.WHIP },
	"k9":     func(r store.PitchingCareerRateRow) *float64 { return r.K9 },
	"bb9":    func(r store.PitchingCareerRateRow) *float64 { return r.BB9 },
	"h9":     func(r store.PitchingCareerRateRow) *float64 { return r.H9 },
	"hr9":    func(r store.PitchingCareerRateRow) *float64 { return r.HR9 },
	"kPerBb": func(r store.PitchingCareerRateRow) *float64 { return r.KPerBB },
	"kPct":   func(r store.PitchingCareerRateRow) *float64 { return r.KPct },
	"winPct": func(r store.PitchingCareerRateRow) *float64 { return r.WinPct },
	"pPerIp": func(r store.PitchingCareerRateRow) *float64 { return r.PPerIP },
	"fip":    func(r store.PitchingCareerRateRow) *float64 { return r.FIP },
	"smbWar": func(r store.PitchingCareerRateRow) *float64 { return r.SmbWAR },
}

// isBetterRate returns true if val beats best in the given direction.
func isBetterRate(val, best float64, higherIsBetter bool) bool {
	if higherIsBetter {
		return val > best
	}
	return val < best
}

// ── Rate league leaders ───────────────────────────────────────────────────────

func computeBattingRateLeagueLeaders(rows []store.BattingRateRow) map[int]map[string][]int64 {
	bySeason := make(map[int][]store.BattingRateRow)
	for _, r := range rows {
		if float64(r.PlateAppearances) >= float64(r.NumGames)*3.1 {
			bySeason[r.SeasonNum] = append(bySeason[r.SeasonNum], r)
		}
	}
	out := make(map[int]map[string][]int64, len(bySeason))
	for seasonNum, seasonRows := range bySeason {
		out[seasonNum] = battingRateSeasonBest(seasonRows)
	}
	return out
}

func battingRateSeasonBest(rows []store.BattingRateRow) map[string][]int64 {
	leaders := make(map[string][]int64, len(battingRateExtractors))
	for key, fn := range battingRateExtractors {
		higherIsBetter := battingRateHigherIsBetter[key]
		var best *float64
		for _, r := range rows {
			v := fn(r)
			if v == nil {
				continue
			}
			if best == nil || isBetterRate(*v, *best, higherIsBetter) {
				best = v
			}
		}
		if best == nil {
			continue
		}
		for _, r := range rows {
			if v := fn(r); v != nil && *v == *best {
				leaders[key] = append(leaders[key], r.PlayerID)
			}
		}
	}
	return leaders
}

func computePitchingRateLeagueLeaders(rows []store.PitchingRateRow) map[int]map[string][]int64 {
	bySeason := make(map[int][]store.PitchingRateRow)
	for _, r := range rows {
		if r.OutsPitched >= r.NumGames*3 {
			bySeason[r.SeasonNum] = append(bySeason[r.SeasonNum], r)
		}
	}
	out := make(map[int]map[string][]int64, len(bySeason))
	for seasonNum, seasonRows := range bySeason {
		out[seasonNum] = pitchingRateSeasonBest(seasonRows)
	}
	return out
}

func pitchingRateSeasonBest(rows []store.PitchingRateRow) map[string][]int64 {
	leaders := make(map[string][]int64, len(pitchingRateExtractors))
	for key, fn := range pitchingRateExtractors {
		higherIsBetter := pitchingRateHigherIsBetter[key]
		var best *float64
		for _, r := range rows {
			v := fn(r)
			if v == nil {
				continue
			}
			if best == nil || isBetterRate(*v, *best, higherIsBetter) {
				best = v
			}
		}
		if best == nil {
			continue
		}
		for _, r := range rows {
			if v := fn(r); v != nil && *v == *best {
				leaders[key] = append(leaders[key], r.PlayerID)
			}
		}
	}
	return leaders
}

// ── Rate single-season records ────────────────────────────────────────────────

func computeBattingRateSingleSeasonRecords(rows []store.BattingRateRow) map[string][]PlayerSeasonRef {
	qualified := make([]store.BattingRateRow, 0, len(rows))
	for _, r := range rows {
		if float64(r.PlateAppearances) >= float64(r.NumGames)*3.1 {
			qualified = append(qualified, r)
		}
	}
	out := make(map[string][]PlayerSeasonRef, len(battingRateExtractors))
	for key, fn := range battingRateExtractors {
		higherIsBetter := battingRateHigherIsBetter[key]
		var best *float64
		for _, r := range qualified {
			v := fn(r)
			if v == nil {
				continue
			}
			if best == nil || isBetterRate(*v, *best, higherIsBetter) {
				best = v
			}
		}
		if best == nil {
			continue
		}
		for _, r := range qualified {
			if v := fn(r); v != nil && *v == *best {
				out[key] = append(out[key], PlayerSeasonRef{PlayerID: r.PlayerID, SeasonNum: r.SeasonNum})
			}
		}
	}
	return out
}

func computePitchingRateSingleSeasonRecords(rows []store.PitchingRateRow) map[string][]PlayerSeasonRef {
	qualified := make([]store.PitchingRateRow, 0, len(rows))
	for _, r := range rows {
		if r.OutsPitched >= r.NumGames*3 {
			qualified = append(qualified, r)
		}
	}
	out := make(map[string][]PlayerSeasonRef, len(pitchingRateExtractors))
	for key, fn := range pitchingRateExtractors {
		higherIsBetter := pitchingRateHigherIsBetter[key]
		var best *float64
		for _, r := range qualified {
			v := fn(r)
			if v == nil {
				continue
			}
			if best == nil || isBetterRate(*v, *best, higherIsBetter) {
				best = v
			}
		}
		if best == nil {
			continue
		}
		for _, r := range qualified {
			if v := fn(r); v != nil && *v == *best {
				out[key] = append(out[key], PlayerSeasonRef{PlayerID: r.PlayerID, SeasonNum: r.SeasonNum})
			}
		}
	}
	return out
}

// ── Rate career records ───────────────────────────────────────────────────────

func computeBattingCareerRateRecords(rows []store.BattingCareerRateRow, paThreshold int) map[string][]int64 {
	qualified := make([]store.BattingCareerRateRow, 0, len(rows))
	for _, r := range rows {
		if r.CareerPA >= paThreshold {
			qualified = append(qualified, r)
		}
	}
	if len(qualified) == 0 {
		return nil
	}
	records := make(map[string][]int64, len(battingCareerRateExtractors))
	for key, fn := range battingCareerRateExtractors {
		higherIsBetter := battingRateHigherIsBetter[key]
		var best *float64
		for _, r := range qualified {
			v := fn(r)
			if v == nil {
				continue
			}
			if best == nil || isBetterRate(*v, *best, higherIsBetter) {
				best = v
			}
		}
		if best == nil {
			continue
		}
		for _, r := range qualified {
			if v := fn(r); v != nil && *v == *best {
				records[key] = append(records[key], r.PlayerID)
			}
		}
	}
	return records
}

func computePitchingCareerRateRecords(rows []store.PitchingCareerRateRow, outsThreshold int) map[string][]int64 {
	qualified := make([]store.PitchingCareerRateRow, 0, len(rows))
	for _, r := range rows {
		if r.OutsPitched >= outsThreshold {
			qualified = append(qualified, r)
		}
	}
	if len(qualified) == 0 {
		return nil
	}
	records := make(map[string][]int64, len(pitchingCareerRateExtractors))
	for key, fn := range pitchingCareerRateExtractors {
		higherIsBetter := pitchingRateHigherIsBetter[key]
		var best *float64
		for _, r := range qualified {
			v := fn(r)
			if v == nil {
				continue
			}
			if best == nil || isBetterRate(*v, *best, higherIsBetter) {
				best = v
			}
		}
		if best == nil {
			continue
		}
		for _, r := range qualified {
			if v := fn(r); v != nil && *v == *best {
				records[key] = append(records[key], r.PlayerID)
			}
		}
	}
	return records
}

// ── Shared helpers ────────────────────────────────────────────────────────────

// careerRecordsFromTotals finds the all-time best per stat key from pre-summed
// player career totals and returns a map of statKey → []playerID (ties included).
// higherIsBetter controls the comparison direction per stat key.
// qualifies gates which players are considered for lower-is-better stats
// (e.g. require atBats > 0 for batting, outsPitched > 0 for pitching).
func careerRecordsFromTotals(
	totals map[int64]map[string]int,
	higherIsBetter map[string]bool,
	qualifies func(int64, map[string]int) bool,
) map[string][]int64 {
	if len(totals) == 0 {
		return nil
	}

	allKeys := make(map[string]struct{})
	for _, t := range totals {
		for k := range t {
			allKeys[k] = struct{}{}
		}
	}

	records := make(map[string][]int64, len(allKeys))
	for key := range allKeys {
		isHigher := higherIsBetter[key]
		hasBest := false
		var best int
		for pid, t := range totals {
			if !isHigher && qualifies != nil && !qualifies(pid, t) {
				continue
			}
			v := t[key]
			if !hasBest || isBetterCount(v, best, isHigher) {
				best = v
				hasBest = true
			}
		}
		if !hasBest {
			continue
		}
		if isHigher && best <= 0 {
			continue
		}
		for pid, t := range totals {
			if !isHigher && qualifies != nil && !qualifies(pid, t) {
				continue
			}
			if t[key] == best {
				records[key] = append(records[key], pid)
			}
		}
	}
	return records
}
