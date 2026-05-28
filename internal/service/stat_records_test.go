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
	// Player 1 career: 78 HR; Player 2 career: 65 HR
	records := computeBattingCareerRecords(rows)
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
	// Both players: 75 HR career
	records := computeBattingCareerRecords(rows)
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
	// P10: 1410; P11: 1400
	records := computePitchingCareerRecords(rows)
	if got := records["outsPitched"]; len(got) != 1 || got[0] != 10 {
		t.Errorf("career IP record: want [10], got %v", got)
	}
}
