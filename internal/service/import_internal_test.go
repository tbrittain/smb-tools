package service

import (
	"testing"

	"smb-tools/internal/models"
)

// ── scheduledGamesPerTeam ──────────────────────────────────────────────────────

func TestScheduledGamesPerTeam_TwoTeams(t *testing.T) {
	// With exactly 2 teams, total matchups happens to equal per-team games —
	// this is the case the shared test fixture exercises, and why a bug here
	// went undetected by the higher-level import tests.
	games := []models.SaveGameGame{
		{HomeTeamGUID: "A", AwayTeamGUID: "B"},
		{HomeTeamGUID: "B", AwayTeamGUID: "A"},
	}
	if got := scheduledGamesPerTeam(games); got != 2 {
		t.Errorf("scheduledGamesPerTeam: got %d, want 2", got)
	}
}

func TestScheduledGamesPerTeam_MultiTeamLeague(t *testing.T) {
	// 4 teams (A, B, C, D), each playing 3 games — 6 total matchups (round robin).
	// A naive len(games) count would report 6, wildly overstating the per-team
	// season length and breaking qualified-player PA/IP thresholds downstream.
	games := []models.SaveGameGame{
		{HomeTeamGUID: "A", AwayTeamGUID: "B"},
		{HomeTeamGUID: "A", AwayTeamGUID: "C"},
		{HomeTeamGUID: "A", AwayTeamGUID: "D"},
		{HomeTeamGUID: "B", AwayTeamGUID: "C"},
		{HomeTeamGUID: "B", AwayTeamGUID: "D"},
		{HomeTeamGUID: "C", AwayTeamGUID: "D"},
	}
	if got := scheduledGamesPerTeam(games); got != 3 {
		t.Errorf("scheduledGamesPerTeam: got %d, want 3 (per-team games), not %d (total matchups)", got, len(games))
	}
}

func TestScheduledGamesPerTeam_ByeWeekDoesNotSkewMode(t *testing.T) {
	// 5 teams normally play 4 games each in a full round robin; team E sits out
	// one round (a bye), playing only 3. The mode across teams should still
	// reflect the standard 4-game schedule, not the bye-affected outlier.
	games := []models.SaveGameGame{
		{HomeTeamGUID: "A", AwayTeamGUID: "B"},
		{HomeTeamGUID: "A", AwayTeamGUID: "C"},
		{HomeTeamGUID: "A", AwayTeamGUID: "D"},
		{HomeTeamGUID: "A", AwayTeamGUID: "E"},
		{HomeTeamGUID: "B", AwayTeamGUID: "C"},
		{HomeTeamGUID: "B", AwayTeamGUID: "D"},
		{HomeTeamGUID: "B", AwayTeamGUID: "E"},
		{HomeTeamGUID: "C", AwayTeamGUID: "D"},
		{HomeTeamGUID: "D", AwayTeamGUID: "E"},
		// C and E only get 3 games each; A, B, D get 4.
	}
	if got := scheduledGamesPerTeam(games); got != 4 {
		t.Errorf("scheduledGamesPerTeam: got %d, want 4 (mode)", got)
	}
}

func TestScheduledGamesPerTeam_Empty(t *testing.T) {
	if got := scheduledGamesPerTeam(nil); got != 0 {
		t.Errorf("scheduledGamesPerTeam(nil): got %d, want 0", got)
	}
}
