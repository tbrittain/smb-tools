package store

import "testing"

// TestSaveGamePitcherRole covers the full PitcherRole enum
// and handling for non-pitchers, bare strings, and unseen
// codes.
func TestSaveGamePitcherRole(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "Starter", raw: "1", want: "SP"},
		{name: "StarterReliever", raw: "2", want: "SP/RP"},
		{name: "Reliever", raw: "3", want: "RP"},
		{name: "Closer", raw: "4", want: "CL"},
		{name: "non-pitcher empty", raw: "", want: ""},
		{name: "unknown code passes through", raw: "99", want: "99"},
		{name: "already-translated string passes through", raw: "SP", want: "SP"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := saveGamePitcherRole(tc.raw); got != tc.want {
				t.Errorf("saveGamePitcherRole(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}
