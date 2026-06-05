package service

import (
	"context"
	"fmt"
	"log/slog"

	"smb-tools/internal/models"
	"smb-tools/internal/store"
)

// Scaling constants taken directly from the legacy SmbExplorerCompanion
// WeightedOpsPlusOrEraMinus.cs. The 11500 divisor normalises to a
// franchise-season equivalent; 95 is the replacement-level baseline.
const (
	battingScalingFactor     = 0.9 * 2.75 / 11500.0
	baserunningScalingFactor = 500.0 / 11500.0
	pitchingScalingFactor    = 2.0 * 2.75 / 11500.0
	warBaseline              = 95.0
)

// LeagueStats holds the per-season constants needed to compute OPS+, ERA+,
// FIP, FIP-, and smbWAR for individual players.
type LeagueStats struct {
	LgOBP       float64 // league on-base percentage
	LgSLG       float64 // league slugging percentage
	LgERA       float64 // league ERA (equals lgFIP by construction of FIPConstant)
	FIPConstant float64 // league-adjusted constant so that lgFIP == lgERA
}

// ComputeLeagueStats derives LeagueStats from raw season-level batting and
// pitching totals. Returns zero-value LeagueStats if denominators are zero
// (e.g. a season with no data).
func ComputeLeagueStats(
	lgAB, lgH, lg2B, lg3B, lgHR, lgBB, lgHBP, lgSacF int64,
	lgOutsPitched, lgER, lgHRA, lgBBA, lgHBPA, lgK int64,
) LeagueStats {
	var ls LeagueStats

	obpDenom := lgAB + lgBB + lgHBP + lgSacF
	if obpDenom > 0 {
		ls.LgOBP = float64(lgH+lgBB+lgHBP) / float64(obpDenom)
	}
	if lgAB > 0 {
		singles := lgH - lg2B - lg3B - lgHR
		tb := singles + lg2B*2 + lg3B*3 + lgHR*4
		ls.LgSLG = float64(tb) / float64(lgAB)
	}

	if lgOutsPitched > 0 {
		ls.LgERA = float64(lgER) * 27.0 / float64(lgOutsPitched)
		lgIP := float64(lgOutsPitched) / 3.0
		lgFIPNum := (float64(lgHRA)*13.0 + float64(lgBBA+lgHBPA)*3.0 - float64(lgK)*2.0) / lgIP
		ls.FIPConstant = ls.LgERA - lgFIPNum
	}

	return ls
}

// ComputeOPSPlus returns OPS+ = 100 × (OBP/lgOBP + SLG/lgSLG − 1).
// Returns nil if obp or slg is nil, or league denominators are zero.
func ComputeOPSPlus(lg LeagueStats, obp, slg *float64) *float64 {
	if obp == nil || slg == nil {
		return nil
	}
	if lg.LgOBP == 0 || lg.LgSLG == 0 {
		return nil
	}
	v := 100.0 * (*obp/lg.LgOBP + *slg/lg.LgSLG - 1.0)
	return &v
}

// ComputeERAPlus returns ERA+ = 100 × lgERA / ERA. Higher is better.
// Returns nil if era is nil, zero, or lgERA is zero.
func ComputeERAPlus(lgERA float64, era *float64) *float64 {
	if era == nil || *era == 0 || lgERA == 0 {
		return nil
	}
	v := 100.0 * lgERA / *era
	return &v
}

// ComputeFIP returns FIP = (13×HR + 3×(BB+HBP) − 2×K) / IP + fipConstant.
// Returns nil if outsPitched is zero.
func ComputeFIP(hr, bb, hbp, k, outsPitched int, fipConstant float64) *float64 {
	if outsPitched == 0 {
		return nil
	}
	ip := float64(outsPitched) / 3.0
	v := (float64(hr)*13.0+float64(bb+hbp)*3.0-float64(k)*2.0)/ip + fipConstant
	return &v
}

// ComputeFIPMinus returns FIP− = 100 × FIP / lgERA. Lower is better.
// Returns nil if fip is nil or lgERA is zero.
func ComputeFIPMinus(fip *float64, lgERA float64) *float64 {
	if fip == nil || lgERA == 0 {
		return nil
	}
	v := 100.0 * *fip / lgERA
	return &v
}

// ComputeBattingWAR computes smbWAR for a batter using the legacy formula:
//
//	(OPS+ − 95) × PA × BattingScalingFactor + (SB − CS) × BaserunningScalingFactor
//
// Returns nil if opsPlus is nil.
func ComputeBattingWAR(opsPlus *float64, pa, sb, cs int) *float64 {
	if opsPlus == nil {
		return nil
	}
	v := (*opsPlus-warBaseline)*float64(pa)*battingScalingFactor +
		float64(sb-cs)*baserunningScalingFactor
	return &v
}

// ComputePitchingWAR computes smbWAR for a pitcher using the legacy formula:
//
//	((ERA+ + FIPplus) / 2 − 95) × IP × PitchingScalingFactor
//
// where FIPplus = 10000 / FIP− (exact conversion from lower=better to
// higher=better). Returns nil if eraPlus or fipMinus is nil or fipMinus is zero.
func ComputePitchingWAR(eraPlus, fipMinus *float64, outsPitched int) *float64 {
	if eraPlus == nil || fipMinus == nil || *fipMinus == 0 {
		return nil
	}
	ip := float64(outsPitched) / 3.0
	fipPlus := 10000.0 / *fipMinus
	v := ((*eraPlus+fipPlus)/2.0-warBaseline) * ip * pitchingScalingFactor
	return &v
}

// ApplyContextStats computes and persists league-level context stats for one
// season × is_regular_season combination. It aggregates totals from the store,
// computes LeagueStats, then updates every player batting/pitching row with
// OPS+, ERA+, FIP, FIP−, and smbWAR. All DB calls use db (typically a *sql.Tx).
func ApplyContextStats(ctx context.Context, db store.DBTX, seasonID int64, isRegularSeason bool) error {
	slog.Debug("ApplyContextStats: starting", "seasonID", seasonID, "regularSeason", isRegularSeason)
	cs := store.NewContextStatsStore(db)

	bt, err := cs.GetLeagueBattingTotals(ctx, seasonID, isRegularSeason)
	if err != nil {
		return fmt.Errorf("ApplyContextStats: %w", err)
	}
	pt, err := cs.GetLeaguePitchingTotals(ctx, seasonID, isRegularSeason)
	if err != nil {
		return fmt.Errorf("ApplyContextStats: %w", err)
	}

	lg := ComputeLeagueStats(
		bt.AB, bt.Hits, bt.Doubles, bt.Triples, bt.HomeRuns, bt.Walks, bt.HBP, bt.SacFlies,
		pt.OutsPitched, pt.EarnedRuns, pt.HRAllowed, pt.BBAllowed, pt.HBPAllowed, pt.KPitched,
	)

	rec := store.LeagueSeasonStatsRecord{
		SeasonID:        seasonID,
		IsRegularSeason: isRegularSeason,
		BattingTotals:   bt,
		PitchingTotals:  pt,
	}
	if lg.LgOBP != 0 {
		v := lg.LgOBP
		rec.LgOBP = &v
	}
	if lg.LgSLG != 0 {
		v := lg.LgSLG
		rec.LgSLG = &v
	}
	if lg.LgERA != 0 {
		v := lg.LgERA
		rec.LgERA = &v
		c := lg.FIPConstant
		rec.FIPConstant = &c
	}
	if err := cs.UpsertLeagueSeasonStats(ctx, rec); err != nil {
		return fmt.Errorf("ApplyContextStats: %w", err)
	}

	if err := applyBattingContextStats(ctx, cs, seasonID, isRegularSeason, lg); err != nil {
		return fmt.Errorf("ApplyContextStats batting: %w", err)
	}
	if err := applyPitchingContextStats(ctx, cs, seasonID, isRegularSeason, lg); err != nil {
		return fmt.Errorf("ApplyContextStats pitching: %w", err)
	}
	slog.Debug("ApplyContextStats: complete", "seasonID", seasonID, "regularSeason", isRegularSeason)
	return nil
}

func applyBattingContextStats(
	ctx context.Context, cs *store.ContextStatsStore,
	seasonID int64, isRegularSeason bool, lg LeagueStats,
) error {
	rows, err := cs.GetBattingRowsForContext(ctx, seasonID, isRegularSeason)
	if err != nil {
		return err
	}
	for _, r := range rows {
		stats := models.CareerBattingStats{
			AtBats:     r.AtBats,
			Hits:       r.Hits,
			Doubles:    r.Doubles,
			Triples:    r.Triples,
			HomeRuns:   r.HomeRuns,
			Walks:      r.Walks,
			HitByPitch: r.HitByPitch,
			SacHits:    r.SacHits,
			SacFlies:   r.SacFlies,
		}
		ComputeBattingRates(&stats)
		opsPlus := ComputeOPSPlus(lg, stats.OBP, stats.SLG)
		pa := r.AtBats + r.Walks + r.HitByPitch + r.SacHits + r.SacFlies
		smbWAR := ComputeBattingWAR(opsPlus, pa, r.StolenBases, r.CaughtStealing)
		if err := cs.UpdateBattingContextStats(ctx, r.ID, opsPlus, smbWAR); err != nil {
			return err
		}
	}
	return nil
}

func applyPitchingContextStats(
	ctx context.Context, cs *store.ContextStatsStore,
	seasonID int64, isRegularSeason bool, lg LeagueStats,
) error {
	rows, err := cs.GetPitchingRowsForContext(ctx, seasonID, isRegularSeason)
	if err != nil {
		return err
	}
	for _, r := range rows {
		stats := models.CareerPitchingStats{
			OutsPitched: r.OutsPitched,
			EarnedRuns:  r.EarnedRuns,
		}
		ComputePitchingRates(&stats)
		eraPlus := ComputeERAPlus(lg.LgERA, stats.ERA)
		fip := ComputeFIP(r.HomeRunsAllowed, r.Walks, r.HitBatters, r.Strikeouts, r.OutsPitched, lg.FIPConstant)
		fipMinus := ComputeFIPMinus(fip, lg.LgERA)
		smbWAR := ComputePitchingWAR(eraPlus, fipMinus, r.OutsPitched)
		if err := cs.UpdatePitchingContextStats(ctx, r.ID, eraPlus, fip, fipMinus, smbWAR); err != nil {
			return err
		}
	}
	return nil
}

// ApplyCareerStats computes and persists all three stat_type career rows
// (regular_season, playoffs, total_career) for each player in playerIDs.
// Must be called after ApplyContextStats so per-season smb_war values are set.
func ApplyCareerStats(ctx context.Context, db store.DBTX, playerIDs []int64) error {
	slog.Debug("ApplyCareerStats: starting", "players", len(playerIDs))
	cs := store.NewCareerStatsStore(db)
	statTypes := []models.CareerStatType{
		models.CareerStatTypeRegularSeason,
		models.CareerStatTypePlayoffs,
		models.CareerStatTypeTotalCareer,
	}
	for _, playerID := range playerIDs {
		for _, st := range statTypes {
			if err := applyCareerBatting(ctx, cs, playerID, st); err != nil {
				return fmt.Errorf("ApplyCareerStats batting (player=%d type=%s): %w", playerID, st, err)
			}
			if err := applyCareerPitching(ctx, cs, playerID, st); err != nil {
				return fmt.Errorf("ApplyCareerStats pitching (player=%d type=%s): %w", playerID, st, err)
			}
		}
	}
	slog.Debug("ApplyCareerStats: complete", "players", len(playerIDs))
	return nil
}

func applyCareerBatting(ctx context.Context, cs *store.CareerStatsStore, playerID int64, statType models.CareerStatType) error {
	agg, err := cs.GetCareerBattingTotalsWithLeague(ctx, playerID, statType)
	if err != nil {
		return err
	}
	if agg == nil {
		return nil
	}

	lg := ComputeLeagueStats(
		agg.LgAtBats, agg.LgHits, agg.LgDoubles, agg.LgTriples, agg.LgHomeRuns,
		agg.LgWalks, agg.LgHBP, agg.LgSacFlies,
		0, 0, 0, 0, 0, 0, // no pitching league stats needed for batting
	)

	b := &models.CareerBattingStats{
		GamesPlayed: int(agg.GamesPlayed), GamesBatting: int(agg.GamesBatting),
		AtBats: int(agg.AtBats), Runs: int(agg.Runs), Hits: int(agg.Hits),
		Doubles: int(agg.Doubles), Triples: int(agg.Triples), HomeRuns: int(agg.HomeRuns),
		RBI: int(agg.RBI), StolenBases: int(agg.StolenBases), CaughtStealing: int(agg.CaughtStealing),
		Walks: int(agg.Walks), Strikeouts: int(agg.Strikeouts), HitByPitch: int(agg.HitByPitch),
		SacHits: int(agg.SacHits), SacFlies: int(agg.SacFlies),
		Errors: int(agg.Errors), PassedBalls: int(agg.PassedBalls),
		SmbWAR: agg.SmbWARSum,
	}
	ComputeBattingRates(b)
	b.OPSPlus = ComputeOPSPlus(lg, b.OBP, b.SLG)
	// Career smbWAR is the sum of per-season values, not recomputed from career OPS+.

	return cs.UpsertCareerBattingStats(ctx, playerID, statType, int(agg.SeasonsPlayed), b)
}

func applyCareerPitching(ctx context.Context, cs *store.CareerStatsStore, playerID int64, statType models.CareerStatType) error {
	agg, err := cs.GetCareerPitchingTotalsWithLeague(ctx, playerID, statType)
	if err != nil {
		return err
	}
	if agg == nil {
		return nil
	}

	lg := ComputeLeagueStats(
		0, 0, 0, 0, 0, 0, 0, 0, // no batting league stats needed for pitching
		agg.LgOutsPitched, agg.LgEarnedRuns, agg.LgHRAllowed,
		agg.LgBBAllowed, agg.LgHBPAllowed, agg.LgKPitched,
	)

	p := &models.CareerPitchingStats{
		Wins: int(agg.Wins), Losses: int(agg.Losses), Games: int(agg.Games),
		GamesStarted: int(agg.GamesStarted), CompleteGames: int(agg.CompleteGames),
		Shutouts: int(agg.Shutouts), Saves: int(agg.Saves),
		OutsPitched: int(agg.OutsPitched), HitsAllowed: int(agg.HitsAllowed),
		EarnedRuns: int(agg.EarnedRuns), HomeRunsAllowed: int(agg.HomeRunsAllowed),
		Walks: int(agg.Walks), Strikeouts: int(agg.Strikeouts), HitBatters: int(agg.HitBatters),
		BattersFaced: int(agg.BattersFaced), GamesFinished: int(agg.GamesFinished),
		RunsAllowed: int(agg.RunsAllowed), WildPitches: int(agg.WildPitches),
		TotalPitches: int(agg.TotalPitches),
		SmbWAR:       agg.SmbWARSum,
	}
	ComputePitchingRates(p)

	p.ERAPlus = ComputeERAPlus(lg.LgERA, p.ERA)
	p.FIP = ComputeFIP(int(agg.HomeRunsAllowed), int(agg.Walks), int(agg.HitBatters), int(agg.Strikeouts), int(agg.OutsPitched), lg.FIPConstant)
	p.FIPMinus = ComputeFIPMinus(p.FIP, lg.LgERA)
	// Career smbWAR is the sum of per-season values, not recomputed from career ERA+/FIP−.

	return cs.UpsertCareerPitchingStats(ctx, playerID, statType, int(agg.SeasonsPlayed), p)
}
