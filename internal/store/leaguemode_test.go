package store

import (
	"testing"

	"smb-tools/internal/models"
)

func TestResolveLeagueMode(t *testing.T) {
	tests := []struct {
		name         string
		hasFranchise bool
		elimination  bool
		numSeasons   int
		want         models.LeagueMode
	}{
		{"franchise wins over elimination", true, true, 5, models.LeagueModeFranchise},
		{"franchise wins over seasons", true, false, 10, models.LeagueModeFranchise},
		{"franchise with zero seasons", true, false, 0, models.LeagueModeFranchise},
		{"elimination without franchise", false, true, 3, models.LeagueModeElimination},
		{"season mode: no franchise, seasons played", false, false, 4, models.LeagueModeSeason},
		{"season mode: single season", false, false, 1, models.LeagueModeSeason},
		{"empty shell: no franchise, no seasons", false, false, 0, models.LeagueModeNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveLeagueMode(tt.hasFranchise, tt.elimination, tt.numSeasons)
			if got != tt.want {
				t.Errorf("resolveLeagueMode(%v, %v, %d) = %q, want %q",
					tt.hasFranchise, tt.elimination, tt.numSeasons, got, tt.want)
			}
		})
	}
}
