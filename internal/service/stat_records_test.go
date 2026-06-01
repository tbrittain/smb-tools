package service

import (
	"slices"
	"testing"

	"smb-tools/internal/store"
)

// ── Batting league leaders ────────────────────────────────────────────────────

func TestComputeBattingLeagueLeaders_SingleLeader(t *testing.T) {
	rows := []store.BattingCountRow{
		{PlayerID: 1, SeasonNum: 1, HomeRuns: 40, Hits: 150},
		{PlayerID: 2, SeasonNum: 1, HomeRuns: 30, Hits: 160},
	}
	leaders := computeBattingLeagueLeaders(rows)
	hrLeaders := leaders[1]["homeRuns"]
	if len(hrLeaders) != 1 || hrLeaders[0] != 1 {
		t.Errorf("homeRuns leader: want [1], got %v", hrLeaders)
	}
	hLeaders := leaders[1]["hits"]
	if len(hLeaders) != 1 || hLeaders[0] != 2 {
		t.Errorf("hits leader: want [2], got %v", hLeaders)
	}
}

func TestComputeBattingLeagueLeaders_Tie(t *testing.T) {
	rows := []store.BattingCountRow{
		{PlayerID: 1, SeasonNum: 1, HomeRuns: 35},
		{PlayerID: 2, SeasonNum: 1, HomeRuns: 35},
		{PlayerID: 3, SeasonNum: 1, HomeRuns: 28},
	}
	leaders := computeBattingLeagueLeaders(rows)
	hrLeaders := leaders[1]["homeRuns"]
	slices.Sort(hrLeaders)
	if len(hrLeaders) != 2 || hrLeaders[0] != 1 || hrLeaders[1] != 2 {
		t.Errorf("tied homeRuns leaders: want [1 2], got %v", hrLeaders)
	}
}

func TestComputeBattingLeagueLeaders_MultiSeason(t *testing.T) {
	rows := []store.BattingCountRow{
		{PlayerID: 1, SeasonNum: 1, HomeRuns: 40},
		{PlayerID: 2, SeasonNum: 1, HomeRuns: 30},
		{PlayerID: 2, SeasonNum: 2, HomeRuns: 45},
		{PlayerID: 1, SeasonNum: 2, HomeRuns: 38},
	}
	leaders := computeBattingLeagueLeaders(rows)
	if got := leaders[1]["homeRuns"]; len(got) != 1 || got[0] != 1 {
		t.Errorf("season 1 HR leader: want [1], got %v", got)
	}
	if got := leaders[2]["homeRuns"]; len(got) != 1 || got[0] != 2 {
		t.Errorf("season 2 HR leader: want [2], got %v", got)
	}
}

func TestComputeBattingLeagueLeaders_ZeroValueExcluded(t *testing.T) {
	// A stat where everyone has 0 should not appear in leaders.
	rows := []store.BattingCountRow{
		{PlayerID: 1, SeasonNum: 1, HomeRuns: 0, Hits: 150},
		{PlayerID: 2, SeasonNum: 1, HomeRuns: 0, Hits: 140},
	}
	leaders := computeBattingLeagueLeaders(rows)
	if _, ok := leaders[1]["homeRuns"]; ok {
		t.Error("homeRuns leader should not exist when all values are 0")
	}
	if len(leaders[1]["hits"]) == 0 {
		t.Error("hits leader should exist when values > 0")
	}
}

// ── Single-season records ─────────────────────────────────────────────────────

func TestComputeBattingSingleSeasonRecords_CorrectSeason(t *testing.T) {
	rows := []store.BattingCountRow{
		{PlayerID: 1, SeasonNum: 1, HomeRuns: 40},
		{PlayerID: 2, SeasonNum: 2, HomeRuns: 50}, // all-time record season
		{PlayerID: 3, SeasonNum: 3, HomeRuns: 45},
	}
	records := computeBattingSingleSeasonRecords(rows)
	hrRecords := records["homeRuns"]
	if len(hrRecords) != 1 {
		t.Fatalf("want 1 record holder, got %d", len(hrRecords))
	}
	if hrRecords[0].PlayerID != 2 || hrRecords[0].SeasonNum != 2 {
		t.Errorf("want {playerID:2, seasonNum:2}, got %+v", hrRecords[0])
	}
}

func TestComputeBattingSingleSeasonRecords_TieAcrossSeasons(t *testing.T) {
	rows := []store.BattingCountRow{
		{PlayerID: 1, SeasonNum: 1, HomeRuns: 50},
		{PlayerID: 2, SeasonNum: 3, HomeRuns: 50},
		{PlayerID: 3, SeasonNum: 2, HomeRuns: 45},
	}
	records := computeBattingSingleSeasonRecords(rows)
	hrRecords := records["homeRuns"]
	if len(hrRecords) != 2 {
		t.Fatalf("want 2 tied record holders, got %d: %+v", len(hrRecords), hrRecords)
	}
}

// ── Career records ────────────────────────────────────────────────────────────

func TestComputeBattingCareerRecords_SumsAcrossSeasons(t *testing.T) {
	rows := []store.BattingCountRow{
		{PlayerID: 1, SeasonNum: 1, HomeRuns: 40},
		{PlayerID: 1, SeasonNum: 2, HomeRuns: 38},
		{PlayerID: 2, SeasonNum: 1, HomeRuns: 45},
		{PlayerID: 2, SeasonNum: 2, HomeRuns: 20},
	}
	// Player 1 career: 78 HR; Player 2 career: 65 HR. Threshold 0 = no PA gate.
	records := computeBattingCareerRecords(rows, 0)
	hrRecords := records["homeRuns"]
	if len(hrRecords) != 1 || hrRecords[0] != 1 {
		t.Errorf("career HR record: want [1] (78 HR), got %v", hrRecords)
	}
}

func TestComputeBattingCareerRecords_CareerTie(t *testing.T) {
	rows := []store.BattingCountRow{
		{PlayerID: 1, SeasonNum: 1, HomeRuns: 40},
		{PlayerID: 1, SeasonNum: 2, HomeRuns: 35},
		{PlayerID: 2, SeasonNum: 1, HomeRuns: 50},
		{PlayerID: 2, SeasonNum: 2, HomeRuns: 25},
	}
	// Both players: 75 HR career. Threshold 0 = no PA gate.
	records := computeBattingCareerRecords(rows, 0)
	hrRecords := records["homeRuns"]
	slices.Sort(hrRecords)
	if len(hrRecords) != 2 || hrRecords[0] != 1 || hrRecords[1] != 2 {
		t.Errorf("tied career HR records: want [1 2], got %v", hrRecords)
	}
}

// ── Pitching ─────────────────────────────────────────────────────────────────

func TestComputePitchingLeagueLeaders_Basic(t *testing.T) {
	rows := []store.PitchingCountRow{
		{PlayerID: 10, SeasonNum: 1, Strikeouts: 250, Wins: 20},
		{PlayerID: 11, SeasonNum: 1, Strikeouts: 200, Wins: 18},
	}
	leaders := computePitchingLeagueLeaders(rows)
	if got := leaders[1]["strikeouts"]; len(got) != 1 || got[0] != 10 {
		t.Errorf("K leader: want [10], got %v", got)
	}
	if got := leaders[1]["wins"]; len(got) != 1 || got[0] != 10 {
		t.Errorf("wins leader: want [10], got %v", got)
	}
}

func TestComputePitchingCareerRecords_SumsOutsPitched(t *testing.T) {
	rows := []store.PitchingCountRow{
		{PlayerID: 10, SeasonNum: 1, OutsPitched: 720},
		{PlayerID: 10, SeasonNum: 2, OutsPitched: 690},
		{PlayerID: 11, SeasonNum: 1, OutsPitched: 800},
		{PlayerID: 11, SeasonNum: 2, OutsPitched: 600},
	}
	// P10: 1410; P11: 1400. Threshold 0 = no IP gate.
	records := computePitchingCareerRecords(rows, 0)
	if got := records["outsPitched"]; len(got) != 1 || got[0] != 10 {
		t.Errorf("career IP record: want [10], got %v", got)
	}
}

// ── Rate stat league leaders ──────────────────────────────────────────────────

func fp(v float64) *float64 { return &v }

func TestComputeBattingRateLeagueLeaders_QualifiedLeaderWins(t *testing.T) {
	// Player 1: .350 BA, qualified (310 PA, 100-game season → threshold 310)
	// Player 2: .420 BA, NOT qualified (50 PA)
	rows := []store.BattingRateRow{
		{PlayerID: 1, SeasonNum: 1, BA: fp(0.350), PlateAppearances: 310, NumGames: 100},
		{PlayerID: 2, SeasonNum: 1, BA: fp(0.420), PlateAppearances: 50, NumGames: 100},
	}
	leaders := computeBattingRateLeagueLeaders(rows)
	got := leaders[1]["ba"]
	if len(got) != 1 || got[0] != 1 {
		t.Errorf("BA leader: want [1] (qualified .350), got %v", got)
	}
}

func TestComputeBattingRateLeagueLeaders_LowerIsBetter_ABPerHR(t *testing.T) {
	// Player 1: 12.5 AB/HR (better — fewer AB per HR), Player 2: 20.0 AB/HR
	rows := []store.BattingRateRow{
		{PlayerID: 1, SeasonNum: 1, ABPerHR: fp(12.5), PlateAppearances: 500, NumGames: 100},
		{PlayerID: 2, SeasonNum: 1, ABPerHR: fp(20.0), PlateAppearances: 500, NumGames: 100},
	}
	leaders := computeBattingRateLeagueLeaders(rows)
	got := leaders[1]["abPerHr"]
	if len(got) != 1 || got[0] != 1 {
		t.Errorf("AB/HR leader: want [1] (lower 12.5), got %v", got)
	}
}

func TestComputeBattingRateLeagueLeaders_NilValueExcluded(t *testing.T) {
	// Player 1: nil BA (no AB), Player 2: .300 qualified
	rows := []store.BattingRateRow{
		{PlayerID: 1, SeasonNum: 1, BA: nil, PlateAppearances: 400, NumGames: 100},
		{PlayerID: 2, SeasonNum: 1, BA: fp(0.300), PlateAppearances: 400, NumGames: 100},
	}
	leaders := computeBattingRateLeagueLeaders(rows)
	got := leaders[1]["ba"]
	if len(got) != 1 || got[0] != 2 {
		t.Errorf("BA leader: want [2] (non-nil .300), got %v", got)
	}
}

func TestComputeBattingRateLeagueLeaders_TiedLeadersBothAppear(t *testing.T) {
	rows := []store.BattingRateRow{
		{PlayerID: 1, SeasonNum: 1, BA: fp(0.350), PlateAppearances: 400, NumGames: 100},
		{PlayerID: 2, SeasonNum: 1, BA: fp(0.350), PlateAppearances: 420, NumGames: 100},
	}
	leaders := computeBattingRateLeagueLeaders(rows)
	got := leaders[1]["ba"]
	slices.Sort(got)
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Errorf("tied BA leaders: want [1 2], got %v", got)
	}
}

func TestComputeBattingRateSingleSeasonRecords_UnqualifiedExcluded(t *testing.T) {
	rows := []store.BattingRateRow{
		{PlayerID: 1, SeasonNum: 1, OPS: fp(1.100), PlateAppearances: 400, NumGames: 100},
		{PlayerID: 2, SeasonNum: 2, OPS: fp(1.250), PlateAppearances: 50, NumGames: 100}, // unqualified
	}
	records := computeBattingRateSingleSeasonRecords(rows)
	got := records["ops"]
	if len(got) != 1 || got[0].PlayerID != 1 {
		t.Errorf("OPS record: want player 1 only, got %v", got)
	}
}

func TestComputePitchingRateLeagueLeaders_IPThreshold(t *testing.T) {
	// 100-game season → threshold = 300 outs (100 IP)
	// Player 10: 2.80 ERA, 303 outs (qualified)
	// Player 11: 1.50 ERA, 90 outs (not qualified)
	rows := []store.PitchingRateRow{
		{PlayerID: 10, SeasonNum: 1, ERA: fp(2.80), OutsPitched: 303, NumGames: 100},
		{PlayerID: 11, SeasonNum: 1, ERA: fp(1.50), OutsPitched: 90, NumGames: 100},
	}
	leaders := computePitchingRateLeagueLeaders(rows)
	got := leaders[1]["era"]
	if len(got) != 1 || got[0] != 10 {
		t.Errorf("ERA leader: want [10] (qualified), got %v", got)
	}
}

func TestComputePitchingRateSingleSeasonRecords_ZeroERAUnqualifiedExcluded(t *testing.T) {
	rows := []store.PitchingRateRow{
		{PlayerID: 10, SeasonNum: 1, ERA: fp(0.00), OutsPitched: 3, NumGames: 100},   // 1 IP, not qualified
		{PlayerID: 11, SeasonNum: 1, ERA: fp(2.50), OutsPitched: 300, NumGames: 100}, // qualified
	}
	records := computePitchingRateSingleSeasonRecords(rows)
	got := records["era"]
	if len(got) != 1 || got[0].PlayerID != 11 {
		t.Errorf("ERA record: want player 11 (qualified 2.50), got %v", got)
	}
}

func TestComputePitchingRateSingleSeasonRecords_FIPLowerIsBetter(t *testing.T) {
	rows := []store.PitchingRateRow{
		{PlayerID: 10, SeasonNum: 1, FIP: fp(3.50), OutsPitched: 600, NumGames: 100},
		{PlayerID: 11, SeasonNum: 2, FIP: fp(2.80), OutsPitched: 600, NumGames: 100},
	}
	records := computePitchingRateSingleSeasonRecords(rows)
	got := records["fip"]
	if len(got) != 1 || got[0].PlayerID != 11 {
		t.Errorf("FIP record: want player 11 (lower 2.80), got %v", got)
	}
}

func TestComputePitchingRateLeagueLeaders_PPerIPLowerIsBetter(t *testing.T) {
	rows := []store.PitchingRateRow{
		{PlayerID: 10, SeasonNum: 1, PPerIP: fp(14.0), OutsPitched: 600, NumGames: 100},
		{PlayerID: 11, SeasonNum: 1, PPerIP: fp(17.5), OutsPitched: 600, NumGames: 100},
	}
	leaders := computePitchingRateLeagueLeaders(rows)
	got := leaders[1]["pPerIp"]
	if len(got) != 1 || got[0] != 10 {
		t.Errorf("P/IP leader: want [10] (lower 14.0), got %v", got)
	}
}

// ── Rate career records ───────────────────────────────────────────────────────

func TestComputeBattingCareerRateRecords_BelowThresholdExcluded(t *testing.T) {
	// 40-game season → career PA threshold ≈ 740
	seasonLen := 40
	threshold := int(3000 * float64(seasonLen) / 162)
	rows := []store.BattingCareerRateRow{
		{PlayerID: 1, BA: fp(0.310), CareerPA: 800}, // above threshold
		{PlayerID: 2, BA: fp(0.350), CareerPA: 300}, // below threshold
	}
	records := computeBattingCareerRateRecords(rows, threshold)
	got := records["ba"]
	if len(got) != 1 || got[0] != 1 {
		t.Errorf("career BA record: want [1] (above threshold), got %v", got)
	}
}

func TestComputeBattingCareerRateRecords_AboveThresholdBestRateWins(t *testing.T) {
	threshold := 500
	rows := []store.BattingCareerRateRow{
		{PlayerID: 1, OPS: fp(0.900), CareerPA: 600},
		{PlayerID: 2, OPS: fp(1.050), CareerPA: 700},
	}
	records := computeBattingCareerRateRecords(rows, threshold)
	got := records["ops"]
	if len(got) != 1 || got[0] != 2 {
		t.Errorf("career OPS record: want [2] (1.050), got %v", got)
	}
}

func TestComputePitchingCareerRateRecords_OutsThresholdApplied(t *testing.T) {
	seasonLen := 40
	threshold := int(3000 * float64(seasonLen) / 162)
	rows := []store.PitchingCareerRateRow{
		{PlayerID: 10, ERA: fp(3.20), OutsPitched: 900},
		{PlayerID: 11, ERA: fp(2.10), OutsPitched: 100}, // below threshold
	}
	records := computePitchingCareerRateRecords(rows, threshold)
	got := records["era"]
	if len(got) != 1 || got[0] != 10 {
		t.Errorf("career ERA record: want [10] (only qualified), got %v", got)
	}
}

func TestComputePitchingCareerRateRecords_NoQualifiedPlayers(t *testing.T) {
	threshold := 10000
	rows := []store.PitchingCareerRateRow{
		{PlayerID: 10, ERA: fp(2.50), OutsPitched: 300},
	}
	records := computePitchingCareerRateRecords(rows, threshold)
	if records != nil {
		t.Errorf("expect nil when no qualified players, got %v", records)
	}
}

// ── Direction-aware counting stat tests ───────────────────────────────────────

func TestComputeBattingLeagueLeaders_StrikeoutsLowerIsBetter(t *testing.T) {
	// Batter with fewer strikeouts is the "leader" — both must be qualified.
	rows := []store.BattingCountRow{
		{PlayerID: 1, SeasonNum: 1, Strikeouts: 50, PlateAppearances: 520, NumGames: 100},
		{PlayerID: 2, SeasonNum: 1, Strikeouts: 100, PlateAppearances: 540, NumGames: 100},
	}
	leaders := computeBattingLeagueLeaders(rows)
	got := leaders[1]["strikeouts"]
	if len(got) != 1 || got[0] != 1 {
		t.Errorf("K leader (lower): want [1] (50 K), got %v", got)
	}
}

func TestComputeBattingLeagueLeaders_StrikeoutsUnqualifiedExcluded(t *testing.T) {
	// Player 1 has 0 K but only 1 PA — unqualified pitcher pinch-hitting once.
	// Player 2 has 30 K with 400 PA (qualified) and should win.
	rows := []store.BattingCountRow{
		{PlayerID: 1, SeasonNum: 1, Strikeouts: 0, PlateAppearances: 1, NumGames: 100},
		{PlayerID: 2, SeasonNum: 1, Strikeouts: 30, PlateAppearances: 400, NumGames: 100},
	}
	leaders := computeBattingLeagueLeaders(rows)
	got := leaders[1]["strikeouts"]
	if len(got) != 1 || got[0] != 2 {
		t.Errorf("K leader (lower, PA-gated): want [2] (30 K, qualified), got %v", got)
	}
}

func TestComputePitchingLeagueLeaders_HitsAllowedLowerIsBetter(t *testing.T) {
	rows := []store.PitchingCountRow{
		{PlayerID: 10, SeasonNum: 1, HitsAllowed: 160, OutsPitched: 720, NumGames: 100},
		{PlayerID: 11, SeasonNum: 1, HitsAllowed: 210, OutsPitched: 700, NumGames: 100},
	}
	leaders := computePitchingLeagueLeaders(rows)
	got := leaders[1]["hitsAllowed"]
	if len(got) != 1 || got[0] != 10 {
		t.Errorf("H leader (lower): want [10] (160 H), got %v", got)
	}
}

func TestComputePitchingLeagueLeaders_EarnedRunsAndWalksLowerIsBetter(t *testing.T) {
	rows := []store.PitchingCountRow{
		{PlayerID: 10, SeasonNum: 1, EarnedRuns: 50, Walks: 40, OutsPitched: 720, NumGames: 100},
		{PlayerID: 11, SeasonNum: 1, EarnedRuns: 80, Walks: 70, OutsPitched: 700, NumGames: 100},
	}
	leaders := computePitchingLeagueLeaders(rows)
	if got := leaders[1]["earnedRuns"]; len(got) != 1 || got[0] != 10 {
		t.Errorf("ER leader (lower): want [10], got %v", got)
	}
	if got := leaders[1]["walks"]; len(got) != 1 || got[0] != 10 {
		t.Errorf("BB leader (lower): want [10], got %v", got)
	}
}

func TestComputePitchingLeagueLeaders_HitsAllowedUnqualifiedExcluded(t *testing.T) {
	// Player 10: 0 H allowed in 2 outs (0.2 IP) — not qualified.
	// Player 11: 180 H in 720 outs (qualified) — should win.
	rows := []store.PitchingCountRow{
		{PlayerID: 10, SeasonNum: 1, HitsAllowed: 0, OutsPitched: 2, NumGames: 100},
		{PlayerID: 11, SeasonNum: 1, HitsAllowed: 180, OutsPitched: 720, NumGames: 100},
	}
	leaders := computePitchingLeagueLeaders(rows)
	got := leaders[1]["hitsAllowed"]
	if len(got) != 1 || got[0] != 11 {
		t.Errorf("H leader (lower, IP-gated): want [11], got %v", got)
	}
}

func TestComputeBattingSingleSeasonRecords_StrikeoutsLowerIsBetter(t *testing.T) {
	rows := []store.BattingCountRow{
		{PlayerID: 1, SeasonNum: 1, Strikeouts: 60, PlateAppearances: 520, NumGames: 100},
		{PlayerID: 2, SeasonNum: 2, Strikeouts: 45, PlateAppearances: 490, NumGames: 100}, // all-time fewest
		{PlayerID: 3, SeasonNum: 3, Strikeouts: 70, PlateAppearances: 510, NumGames: 100},
	}
	records := computeBattingSingleSeasonRecords(rows)
	got := records["strikeouts"]
	if len(got) != 1 || got[0].PlayerID != 2 || got[0].SeasonNum != 2 {
		t.Errorf("K record (fewest): want {2, season 2}, got %v", got)
	}
}

func TestComputeBattingCareerRecords_StrikeoutsLowerIsBetter_PAThreshold(t *testing.T) {
	// 40-game season → career PA threshold ≈ 740.
	seasonLen := 40
	threshold := int(3000 * float64(seasonLen) / 162)
	// Player 1: 80 K, career PA 800 (above threshold) — should win.
	// Player 2: 5 K, career PA 20 (below threshold) — excluded.
	rows := []store.BattingCountRow{
		{PlayerID: 1, SeasonNum: 1, AtBats: 750, Strikeouts: 80, PlateAppearances: 800},
		{PlayerID: 2, SeasonNum: 1, AtBats: 18, Strikeouts: 5, PlateAppearances: 20},
	}
	records := computeBattingCareerRecords(rows, threshold)
	got := records["strikeouts"]
	if len(got) != 1 || got[0] != 1 {
		t.Errorf("career K record (fewest, PA-gated): want [1], got %v", got)
	}
}

func TestComputePitchingCareerRecords_LowerIsBetterStats_OutsThreshold(t *testing.T) {
	// 40-game season → career outs threshold ≈ 740.
	seasonLen := 40
	threshold := int(3000 * float64(seasonLen) / 162)
	// Player 10: qualified (720 + 300 = 1020 outs).
	// Player 11: 0 outs — not qualified.
	rows := []store.PitchingCountRow{
		{PlayerID: 10, SeasonNum: 1, EarnedRuns: 60, Walks: 40, HitsAllowed: 180, OutsPitched: 720},
		{PlayerID: 10, SeasonNum: 2, EarnedRuns: 30, Walks: 20, HitsAllowed: 90, OutsPitched: 300},
		{PlayerID: 11, SeasonNum: 1, EarnedRuns: 0, Walks: 0, HitsAllowed: 0, OutsPitched: 0},
	}
	records := computePitchingCareerRecords(rows, threshold)
	if got := records["earnedRuns"]; len(got) != 1 || got[0] != 10 {
		t.Errorf("career ER record (fewest, IP-gated): want [10], got %v", got)
	}
	if got := records["walks"]; len(got) != 1 || got[0] != 10 {
		t.Errorf("career BB record (fewest, IP-gated): want [10], got %v", got)
	}
	if got := records["hitsAllowed"]; len(got) != 1 || got[0] != 10 {
		t.Errorf("career H record (fewest, IP-gated): want [10], got %v", got)
	}
}

// ── OPS+/ERA+/FIP- season rate tests ─────────────────────────────────────────

func TestComputeBattingRateLeagueLeaders_OPSPlusHigherIsBetter(t *testing.T) {
	rows := []store.BattingRateRow{
		{PlayerID: 1, SeasonNum: 1, OPSPlus: fp(150), PlateAppearances: 400, NumGames: 100},
		{PlayerID: 2, SeasonNum: 1, OPSPlus: fp(110), PlateAppearances: 400, NumGames: 100},
	}
	leaders := computeBattingRateLeagueLeaders(rows)
	got := leaders[1]["opsPlus"]
	if len(got) != 1 || got[0] != 1 {
		t.Errorf("OPS+ leader: want [1] (150), got %v", got)
	}
}

func TestComputePitchingRateLeagueLeaders_ERAPlus_HigherIsBetter(t *testing.T) {
	rows := []store.PitchingRateRow{
		{PlayerID: 10, SeasonNum: 1, ERAPlus: fp(180), OutsPitched: 600, NumGames: 100},
		{PlayerID: 11, SeasonNum: 1, ERAPlus: fp(130), OutsPitched: 600, NumGames: 100},
	}
	leaders := computePitchingRateLeagueLeaders(rows)
	got := leaders[1]["eraPlus"]
	if len(got) != 1 || got[0] != 10 {
		t.Errorf("ERA+ leader: want [10] (180), got %v", got)
	}
}

func TestComputePitchingRateSingleSeasonRecords_FIPMinus_LowerIsBetter(t *testing.T) {
	rows := []store.PitchingRateRow{
		{PlayerID: 10, SeasonNum: 1, FIPMinus: fp(85), OutsPitched: 600, NumGames: 100},
		{PlayerID: 11, SeasonNum: 2, FIPMinus: fp(70), OutsPitched: 600, NumGames: 100},
	}
	records := computePitchingRateSingleSeasonRecords(rows)
	got := records["fipMinus"]
	if len(got) != 1 || got[0].PlayerID != 11 {
		t.Errorf("FIP- record (lower): want player 11 (70), got %v", got)
	}
}

func TestComputeBattingRateSingleSeasonRecords_OPSPlusIncluded(t *testing.T) {
	rows := []store.BattingRateRow{
		{PlayerID: 1, SeasonNum: 1, OPSPlus: fp(200), PlateAppearances: 400, NumGames: 100},
		{PlayerID: 2, SeasonNum: 2, OPSPlus: fp(160), PlateAppearances: 400, NumGames: 100},
	}
	records := computeBattingRateSingleSeasonRecords(rows)
	got := records["opsPlus"]
	if len(got) != 1 || got[0].PlayerID != 1 {
		t.Errorf("OPS+ record: want player 1 (200), got %v", got)
	}
}

func TestComputeBattingCareerRateRecords_OPSPlusNotTracked(t *testing.T) {
	// OPS+ is excluded from career rate extractors — should not appear in career records.
	threshold := 100
	rows := []store.BattingCareerRateRow{
		{PlayerID: 1, OPS: fp(0.950), CareerPA: 500},
	}
	records := computeBattingCareerRateRecords(rows, threshold)
	if _, ok := records["opsPlus"]; ok {
		t.Error("opsPlus should not appear in career batting rate records")
	}
	if _, ok := records["ops"]; !ok {
		t.Error("ops should appear in career batting rate records")
	}
}
