package models

import "github.com/google/uuid"

// LeagueOverview describes a league's structure for discovery/introspection
// in the League Transfer feature — name, conferences, divisions (optional),
// and teams. It deliberately carries no stat history or season data; that's
// franchise-tracker territory (see docs/league-transfer/ux-flow.md).
//
// Unlike SaveGameLeague (which represents GUIDs as hex strings for the
// existing franchise-tracking path), LeagueOverview uses uuid.UUID directly.
// League Transfer reads AND writes GUIDs across save files, and binding a
// uuid.UUID's byte slice is what prevents the exact bug class documented in
// docs/league-transfer/failure-analysis.md (a GUID accidentally stored as
// text instead of a 16-byte blob).
type LeagueOverview struct {
	GUID        uuid.UUID
	Name        string
	Conferences []ConferenceOverview
}

// ConferenceOverview describes one conference and its divisions. Divisions
// are genuinely optional — a conference may have zero divisions, in which
// case Divisions is empty and Teams holds whatever teams could be resolved
// directly (see GetLeagueOverview's doc comment for the current limitation
// here).
type ConferenceOverview struct {
	GUID      uuid.UUID
	Name      string
	Divisions []DivisionOverview
}

// DivisionOverview describes one division and its teams.
type DivisionOverview struct {
	GUID  uuid.UUID
	Name  string
	Teams []TeamOverview
}

// TeamOverview is a minimal team identity for display purposes.
type TeamOverview struct {
	GUID uuid.UUID
	Name string
}
