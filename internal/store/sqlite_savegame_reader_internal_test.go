package store

import (
	"reflect"
	"testing"
)

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

// TestParsePitchesJSON covers populated arsenals,
// empty/null inputs, and malformed JSON.
func TestParsePitchesJSON(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{name: "empty array", raw: "[]", want: nil},
		{name: "blank", raw: "", want: nil},
		{name: "null", raw: "null", want: nil},
		{name: "single pitch", raw: `[{"optionKey":58}]`, want: []string{"4F"}},
		{name: "arsenal in optionKey order", raw: `[{"optionKey":58},{"optionKey":63},{"optionKey":64}]`, want: []string{"4F", "CB", "SL"}},
		{name: "full arsenal", raw: `[{"optionKey":58},{"optionKey":59},{"optionKey":60},{"optionKey":61},{"optionKey":62}]`, want: []string{"4F", "2F", "SB", "CH", "FK"}},
		{name: "unknown optionKey skipped", raw: `[{"optionKey":58},{"optionKey":99},{"optionKey":64}]`, want: []string{"4F", "SL"}},
		{name: "malformed JSON returns nil", raw: `[not json`, want: nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parsePitchesJSON(tc.raw)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parsePitchesJSON(%q) = %v, want %v", tc.raw, got, tc.want)
			}
		})
	}
}
