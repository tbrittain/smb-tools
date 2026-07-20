package store

import "smb-tools/internal/models"

// resolveLeagueMode is the anti-corruption-layer translation from raw
// t_franchise/t_seasons presence signals to a domain LeagueMode. Mirrors
// SMB3Explorer's LeagueModeExtensions.Parse: a franchise row wins regardless
// of elimination; otherwise any season with elimination=true makes it
// Elimination, any season at all makes it Season, and no franchise + no
// seasons is an empty shell (LeagueModeNone).
//
// Shared by SqliteSaveGameReader.GetLeagues and LeagueSaveStore.resolveLeagueMode,
// which query the same signals via different SQL shapes but must agree on how
// they're interpreted.
func resolveLeagueMode(hasFranchise bool, elimination bool, numSeasons int) models.LeagueMode {
	switch {
	case hasFranchise:
		return models.LeagueModeFranchise
	case elimination:
		return models.LeagueModeElimination
	case numSeasons > 0:
		return models.LeagueModeSeason
	default:
		return models.LeagueModeNone
	}
}
