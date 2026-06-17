package service_test

import (
	"math"
	"testing"

	"smb-tools/internal/service"
)

func ptr(v float64) *float64 { return &v }

// ── ComputeLeagueStats ────────────────────────────────────────────────────────

func TestComputeLeagueStats_Normal(t *testing.T) {
	// Minimal one-batter, one-pitcher season.
	// Batter: AB=180, H=54, 2B=10, 3B=2, HR=12, BB=20, HBP=0, SacF=0
	// Pitcher: outs=540 (180 IP), ER=55, HR_a=15, BB_a=40, HBP_a=0, K=180
	ls := service.ComputeLeagueStats(
		180, 54, 10, 2, 12, 20, 0, 0,
		540, 55, 15, 40, 0, 180,
	)

	// lgOBP = (54+20) / (180+20) = 74/200 = 0.370
	wantOBP := 74.0 / 200.0
	if math.Abs(ls.LgOBP-wantOBP) > 1e-9 {
		t.Errorf("LgOBP: got %.6f, want %.6f", ls.LgOBP, wantOBP)
	}

	// lgSLG = TB/AB; TB = (54-10-2-12) + 2*10 + 3*2 + 4*12 = 30+20+6+48=104
	wantSLG := 104.0 / 180.0
	if math.Abs(ls.LgSLG-wantSLG) > 1e-9 {
		t.Errorf("LgSLG: got %.6f, want %.6f", ls.LgSLG, wantSLG)
	}

	// lgERA = 55*27/540 = 2.75
	wantERA := 55.0 * 27.0 / 540.0
	if math.Abs(ls.LgERA-wantERA) > 1e-9 {
		t.Errorf("LgERA: got %.6f, want %.6f", ls.LgERA, wantERA)
	}

	// lgFIPNum = (13*15 + 3*40 - 2*180) / 180 = (195+120-360)/180 = -45/180 = -0.25
	// FIPConstant = lgERA - lgFIPNum = 2.75 - (-0.25) = 3.00
	wantFIPConst := wantERA - (-45.0/180.0)
	if math.Abs(ls.FIPConstant-wantFIPConst) > 1e-9 {
		t.Errorf("FIPConstant: got %.6f, want %.6f", ls.FIPConstant, wantFIPConst)
	}
}

func TestComputeLeagueStats_ZeroAB(t *testing.T) {
	ls := service.ComputeLeagueStats(0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	if ls.LgOBP != 0 || ls.LgSLG != 0 || ls.LgERA != 0 || ls.FIPConstant != 0 {
		t.Errorf("expected zero-value LeagueStats for zero inputs, got %+v", ls)
	}
}

// ── ComputeOPSPlus ────────────────────────────────────────────────────────────

func TestComputeOPSPlus_LeagueAverage(t *testing.T) {
	ls := service.LeagueStats{LgOBP: 0.32, LgSLG: 0.40}
	got := service.ComputeOPSPlus(ls, ptr(0.32), ptr(0.40))
	if got == nil {
		t.Fatal("expected non-nil OPS+")
		return
	}
	// (0.32/0.32 + 0.40/0.40 - 1) * 100 = 100
	if math.Abs(*got-100.0) > 1e-9 {
		t.Errorf("league-average OPS+: got %.4f, want 100", *got)
	}
}

func TestComputeOPSPlus_AboveAverage(t *testing.T) {
	ls := service.LeagueStats{LgOBP: 0.32, LgSLG: 0.40}
	got := service.ComputeOPSPlus(ls, ptr(0.38), ptr(0.55))
	if got == nil {
		t.Fatal("expected non-nil OPS+")
		return
	}
	want := 100.0 * (0.38/0.32 + 0.55/0.40 - 1.0)
	if math.Abs(*got-want) > 1e-9 {
		t.Errorf("OPS+: got %.4f, want %.4f", *got, want)
	}
}

func TestComputeOPSPlus_NilInputs(t *testing.T) {
	ls := service.LeagueStats{LgOBP: 0.32, LgSLG: 0.40}
	if got := service.ComputeOPSPlus(ls, nil, ptr(0.40)); got != nil {
		t.Error("expected nil OPS+ when OBP is nil")
	}
	if got := service.ComputeOPSPlus(ls, ptr(0.32), nil); got != nil {
		t.Error("expected nil OPS+ when SLG is nil")
	}
}

func TestComputeOPSPlus_ZeroLeague(t *testing.T) {
	ls := service.LeagueStats{LgOBP: 0, LgSLG: 0}
	if got := service.ComputeOPSPlus(ls, ptr(0.32), ptr(0.40)); got != nil {
		t.Error("expected nil OPS+ when league denominators are zero")
	}
}

// ── ComputeERAPlus ────────────────────────────────────────────────────────────

func TestComputeERAPlus_LeagueAverage(t *testing.T) {
	got := service.ComputeERAPlus(4.00, ptr(4.00))
	if got == nil {
		t.Fatal("expected non-nil ERA+")
		return
	}
	if math.Abs(*got-100.0) > 1e-9 {
		t.Errorf("league-average ERA+: got %.4f, want 100", *got)
	}
}

func TestComputeERAPlus_BetterThanAverage(t *testing.T) {
	got := service.ComputeERAPlus(4.00, ptr(2.00))
	if got == nil {
		t.Fatal("expected non-nil ERA+")
		return
	}
	// 100 * 4.00/2.00 = 200
	if math.Abs(*got-200.0) > 1e-9 {
		t.Errorf("ERA+: got %.4f, want 200", *got)
	}
}

func TestComputeERAPlus_NilAndZero(t *testing.T) {
	if got := service.ComputeERAPlus(4.00, nil); got != nil {
		t.Error("expected nil ERA+ for nil era")
	}
	if got := service.ComputeERAPlus(4.00, ptr(0.0)); got != nil {
		t.Error("expected nil ERA+ for era=0")
	}
	if got := service.ComputeERAPlus(0.0, ptr(3.00)); got != nil {
		t.Error("expected nil ERA+ for lgERA=0")
	}
}

// ── ComputeFIP ────────────────────────────────────────────────────────────────

func TestComputeFIP_Basic(t *testing.T) {
	// FIP = (13*15 + 3*40 - 2*180) / 180 + 3.00 = -0.25 + 3.00 = 2.75
	got := service.ComputeFIP(15, 40, 0, 180, 540, 3.00)
	if got == nil {
		t.Fatal("expected non-nil FIP")
		return
	}
	if math.Abs(*got-2.75) > 1e-9 {
		t.Errorf("FIP: got %.4f, want 2.75", *got)
	}
}

func TestComputeFIP_ZeroIP(t *testing.T) {
	if got := service.ComputeFIP(0, 0, 0, 0, 0, 3.10); got != nil {
		t.Error("expected nil FIP for zero outs_pitched")
	}
}

// ── ComputeFIPMinus ───────────────────────────────────────────────────────────

func TestComputeFIPMinus_LeagueAverage(t *testing.T) {
	got := service.ComputeFIPMinus(ptr(2.75), 2.75)
	if got == nil {
		t.Fatal("expected non-nil FIP-")
		return
	}
	if math.Abs(*got-100.0) > 1e-9 {
		t.Errorf("FIP- league average: got %.4f, want 100", *got)
	}
}

func TestComputeFIPMinus_NilAndZero(t *testing.T) {
	if got := service.ComputeFIPMinus(nil, 2.75); got != nil {
		t.Error("expected nil FIP- for nil fip")
	}
	if got := service.ComputeFIPMinus(ptr(2.75), 0); got != nil {
		t.Error("expected nil FIP- for lgERA=0")
	}
}

// ── ComputeBattingWAR ─────────────────────────────────────────────────────────

func TestComputeBattingWAR_LeagueAverage(t *testing.T) {
	// OPS+=100, 200 PA, 0 net SB → small positive (baseline 95 not 100)
	got := service.ComputeBattingWAR(ptr(100.0), 200, 0, 0)
	if got == nil {
		t.Fatal("expected non-nil batting smbWAR")
		return
	}
	want := (100.0-95.0)*200.0*(0.9*2.75/11500.0) + 0
	if math.Abs(*got-want) > 1e-9 {
		t.Errorf("batting smbWAR: got %.6f, want %.6f", *got, want)
	}
}

func TestComputeBattingWAR_Nil(t *testing.T) {
	if got := service.ComputeBattingWAR(nil, 200, 0, 0); got != nil {
		t.Error("expected nil batting smbWAR when OPS+ is nil")
	}
}

func TestComputeBattingWAR_BaserunningComponent(t *testing.T) {
	// 30 SB, 5 CS at exact league average OPS+ (100)
	got := service.ComputeBattingWAR(ptr(100.0), 600, 30, 5)
	if got == nil {
		t.Fatal("expected non-nil batting smbWAR")
		return
	}
	battingPart := (100.0-95.0)*600.0*(0.9*2.75/11500.0)
	runningPart := float64(30-5) * (500.0 / 11500.0)
	want := battingPart + runningPart
	if math.Abs(*got-want) > 1e-9 {
		t.Errorf("smbWAR with baserunning: got %.6f, want %.6f", *got, want)
	}
}

// ── ComputePitchingWAR ────────────────────────────────────────────────────────

func TestComputePitchingWAR_LeagueAverage(t *testing.T) {
	// ERA+=100, FIP-=100, 540 outs (180 IP)
	got := service.ComputePitchingWAR(ptr(100.0), ptr(100.0), 540)
	if got == nil {
		t.Fatal("expected non-nil pitching smbWAR")
		return
	}
	// FIPplus = 10000/100 = 100
	// ((100+100)/2 - 95) * 180 * scale = 5 * 180 * scale
	scale := 2.0 * 2.75 / 11500.0
	want := 5.0 * 180.0 * scale
	if math.Abs(*got-want) > 1e-9 {
		t.Errorf("pitching smbWAR: got %.6f, want %.6f", *got, want)
	}
}

func TestComputePitchingWAR_NilInputs(t *testing.T) {
	if got := service.ComputePitchingWAR(nil, ptr(100.0), 540); got != nil {
		t.Error("expected nil when eraPlus is nil")
	}
	if got := service.ComputePitchingWAR(ptr(100.0), nil, 540); got != nil {
		t.Error("expected nil when fipMinus is nil")
	}
	if got := service.ComputePitchingWAR(ptr(100.0), ptr(0.0), 540); got != nil {
		t.Error("expected nil when fipMinus is zero")
	}
}
