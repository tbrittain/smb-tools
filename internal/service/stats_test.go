package service_test

import (
	"math"
	"testing"

	"smb-tools/internal/models"
	"smb-tools/internal/service"
)

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func ptrVal(p *float64) float64 {
	if p == nil {
		return math.NaN()
	}
	return *p
}

// ── ComputeBattingRates ───────────────────────────────────────────────────────

func TestComputeBattingRates_ZeroAB(t *testing.T) {
	b := &models.CareerBattingStats{} // all zeros
	service.ComputeBattingRates(b)

	if b.BA != nil {
		t.Errorf("BA: want nil, got %v", *b.BA)
	}
	if b.OBP != nil {
		t.Errorf("OBP: want nil, got %v", *b.OBP)
	}
	if b.SLG != nil {
		t.Errorf("SLG: want nil, got %v", *b.SLG)
	}
	if b.OPS != nil {
		t.Errorf("OPS: want nil, got %v", *b.OPS)
	}
	if b.ISO != nil {
		t.Errorf("ISO: want nil, got %v", *b.ISO)
	}
	if b.BABIP != nil {
		t.Errorf("BABIP: want nil, got %v", *b.BABIP)
	}
	if b.KPct != nil {
		t.Errorf("KPct: want nil, got %v", *b.KPct)
	}
	if b.BBPct != nil {
		t.Errorf("BBPct: want nil, got %v", *b.BBPct)
	}
	if b.ABPerHR != nil {
		t.Errorf("ABPerHR: want nil, got %v", *b.ABPerHR)
	}
}

func TestComputeBattingRates_ZeroHR(t *testing.T) {
	b := &models.CareerBattingStats{
		AtBats: 100,
		Hits:   30,
	}
	service.ComputeBattingRates(b)

	if b.ABPerHR != nil {
		t.Errorf("ABPerHR: want nil when HR=0, got %v", *b.ABPerHR)
	}
}

func TestComputeBattingRates_Nominal(t *testing.T) {
	// 500 AB, 150 H (30 2B, 5 3B, 25 HR), 60 BB, 5 HBP, 3 SF, 3 SH, 100 K
	b := &models.CareerBattingStats{
		AtBats:     500,
		Hits:       150,
		Doubles:    30,
		Triples:    5,
		HomeRuns:   25,
		Walks:      60,
		HitByPitch: 5,
		SacFlies:   3,
		SacHits:    3,
		Strikeouts: 100,
	}
	service.ComputeBattingRates(b)

	wantBA := 150.0 / 500.0 // .300
	if b.BA == nil || !approxEqual(*b.BA, wantBA) {
		t.Errorf("BA: want %.4f, got %v", wantBA, ptrVal(b.BA))
	}

	// OBP = (150+60+5) / (500+60+5+3) = 215/568
	wantOBP := 215.0 / 568.0
	if b.OBP == nil || !approxEqual(*b.OBP, wantOBP) {
		t.Errorf("OBP: want %.4f, got %v", wantOBP, ptrVal(b.OBP))
	}

	// TB = (150-30-5-25) + 30*2 + 5*3 + 25*4 = 90+60+15+100 = 265
	// SLG = 265/500 = .530
	wantSLG := 265.0 / 500.0
	if b.SLG == nil || !approxEqual(*b.SLG, wantSLG) {
		t.Errorf("SLG: want %.4f, got %v", wantSLG, ptrVal(b.SLG))
	}

	wantOPS := wantOBP + wantSLG
	if b.OPS == nil || !approxEqual(*b.OPS, wantOPS) {
		t.Errorf("OPS: want %.4f, got %v", wantOPS, ptrVal(b.OPS))
	}

	wantISO := wantSLG - wantBA
	if b.ISO == nil || !approxEqual(*b.ISO, wantISO) {
		t.Errorf("ISO: want %.4f, got %v", wantISO, ptrVal(b.ISO))
	}

	// BABIP = (150-25) / (500-100-25+3) = 125/378
	wantBABIP := 125.0 / 378.0
	if b.BABIP == nil || !approxEqual(*b.BABIP, wantBABIP) {
		t.Errorf("BABIP: want %.4f, got %v", wantBABIP, ptrVal(b.BABIP))
	}

	// PA = 500+60+5+3+3 = 571
	wantKPct := 100.0 / 571.0
	if b.KPct == nil || !approxEqual(*b.KPct, wantKPct) {
		t.Errorf("KPct: want %.4f, got %v", wantKPct, ptrVal(b.KPct))
	}

	wantBBPct := 60.0 / 571.0
	if b.BBPct == nil || !approxEqual(*b.BBPct, wantBBPct) {
		t.Errorf("BBPct: want %.4f, got %v", wantBBPct, ptrVal(b.BBPct))
	}

	wantABPerHR := 500.0 / 25.0 // 20.0
	if b.ABPerHR == nil || !approxEqual(*b.ABPerHR, wantABPerHR) {
		t.Errorf("ABPerHR: want %.1f, got %v", wantABPerHR, ptrVal(b.ABPerHR))
	}
}

// ── ComputePitchingRates ──────────────────────────────────────────────────────

func TestComputePitchingRates_ZeroIP(t *testing.T) {
	p := &models.CareerPitchingStats{} // all zeros
	service.ComputePitchingRates(p)

	if p.ERA != nil {
		t.Errorf("ERA: want nil, got %v", *p.ERA)
	}
	if p.WHIP != nil {
		t.Errorf("WHIP: want nil, got %v", *p.WHIP)
	}
	if p.K9 != nil {
		t.Errorf("K9: want nil, got %v", *p.K9)
	}
	if p.KPerBB != nil {
		t.Errorf("KPerBB: want nil (zero walks), got %v", *p.KPerBB)
	}
	if p.WinPct != nil {
		t.Errorf("WinPct: want nil (0 decisions), got %v", *p.WinPct)
	}
}

func TestComputePitchingRates_ZeroWalks(t *testing.T) {
	p := &models.CareerPitchingStats{
		OutsPitched: 27,
		Strikeouts:  9,
		Walks:       0,
	}
	service.ComputePitchingRates(p)

	if p.KPerBB != nil {
		t.Errorf("KPerBB: want nil when walks=0, got %v", *p.KPerBB)
	}
}

func TestComputePitchingRates_Nominal(t *testing.T) {
	// 9 complete innings = 27 outs, 3 ER, 7 H, 2 BB, 10 K, 1 HR, 100 pitches, 8 BF
	p := &models.CareerPitchingStats{
		Wins:            1,
		Losses:          0,
		OutsPitched:     27,
		EarnedRuns:      3,
		HitsAllowed:     7,
		Walks:           2,
		Strikeouts:      10,
		HomeRunsAllowed: 1,
		TotalPitches:    100,
		BattersFaced:    33,
	}
	service.ComputePitchingRates(p)

	// ERA = 3*27/27 = 3.000
	if p.ERA == nil || !approxEqual(*p.ERA, 3.0) {
		t.Errorf("ERA: want 3.000, got %v", ptrVal(p.ERA))
	}

	// WHIP = (2+7)*3/27 = 27/27 = 1.000
	if p.WHIP == nil || !approxEqual(*p.WHIP, 1.0) {
		t.Errorf("WHIP: want 1.000, got %v", ptrVal(p.WHIP))
	}

	// K/9 = 10*27/27 = 10.0
	if p.K9 == nil || !approxEqual(*p.K9, 10.0) {
		t.Errorf("K/9: want 10.0, got %v", ptrVal(p.K9))
	}

	// BB/9 = 2*27/27 = 2.0
	if p.BB9 == nil || !approxEqual(*p.BB9, 2.0) {
		t.Errorf("BB/9: want 2.0, got %v", ptrVal(p.BB9))
	}

	// H/9 = 7*27/27 = 7.0
	if p.H9 == nil || !approxEqual(*p.H9, 7.0) {
		t.Errorf("H/9: want 7.0, got %v", ptrVal(p.H9))
	}

	// HR/9 = 1*27/27 = 1.0
	if p.HR9 == nil || !approxEqual(*p.HR9, 1.0) {
		t.Errorf("HR/9: want 1.0, got %v", ptrVal(p.HR9))
	}

	// K/BB = 10/2 = 5.0
	if p.KPerBB == nil || !approxEqual(*p.KPerBB, 5.0) {
		t.Errorf("KPerBB: want 5.0, got %v", ptrVal(p.KPerBB))
	}

	// K% = 10/33
	wantKPct := 10.0 / 33.0
	if p.KPct == nil || !approxEqual(*p.KPct, wantKPct) {
		t.Errorf("KPct: want %.4f, got %v", wantKPct, ptrVal(p.KPct))
	}

	// Win% = 1/1 = 1.0
	if p.WinPct == nil || !approxEqual(*p.WinPct, 1.0) {
		t.Errorf("WinPct: want 1.0, got %v", ptrVal(p.WinPct))
	}

	// P/IP = 100*3/27
	wantPPerIP := 100.0 * 3.0 / 27.0
	if p.PPerIP == nil || !approxEqual(*p.PPerIP, wantPPerIP) {
		t.Errorf("PPerIP: want %.4f, got %v", wantPPerIP, ptrVal(p.PPerIP))
	}
}

func TestComputePitchingRates_FractionalIP(t *testing.T) {
	// 97 outs = 32.1 IP; ERA should use outs directly, not truncated IP
	p := &models.CareerPitchingStats{
		OutsPitched: 97,
		EarnedRuns:  10,
	}
	service.ComputePitchingRates(p)

	// ERA = 10*27/97 ≈ 2.7835
	wantERA := 10.0 * 27.0 / 97.0
	if p.ERA == nil || !approxEqual(*p.ERA, wantERA) {
		t.Errorf("ERA with fractional IP: want %.6f, got %v", wantERA, ptrVal(p.ERA))
	}
}
