package store

import (
	"context"

	"smb-tools/internal/models"
)

// SaveGameReader is a read-only view over a decompressed SMB save game database.
// The interface exists to allow the real SQLite implementation to be swapped
// for a test implementation backed by synthetic in-memory data.
//
// All methods are read-only. The underlying database must never be written to.
type SaveGameReader interface {
	// GetLeagues returns all leagues present in the save file.
	GetLeagues(ctx context.Context) ([]models.SaveGameLeague, error)

	// GetFranchiseSeasons returns the season records for a given league,
	// ordered by season number ascending. leagueGUID is the hex-encoded GUID
	// blob from t_leagues (same value stored on Franchise.LeagueGUID).
	GetFranchiseSeasons(ctx context.Context, leagueGUID string) ([]models.SaveGameFranchiseSeason, error)

	// GetCurrentSeasonPlayers returns the full player roster for a given season
	// with all attributes, traits, salary, and (for SMB4) handedness, chemistry,
	// and pitch repertoire populated.
	GetCurrentSeasonPlayers(ctx context.Context, seasonID int) ([]models.SaveGamePlayer, error)

	// GetCurrentSeasonTeams returns all teams for a given season with standings,
	// budget/payroll, and aggregate attribute data.
	GetCurrentSeasonTeams(ctx context.Context, seasonID int) ([]models.SaveGameTeam, error)

	// GetSeasonSchedule returns the full regular season schedule and results
	// for a given season, including starting pitcher assignments.
	GetSeasonSchedule(ctx context.Context, seasonID int) ([]models.SaveGameGame, error)

	// GetPlayoffSchedule returns the playoff schedule and results for a given season.
	GetPlayoffSchedule(ctx context.Context, seasonID int) ([]models.SaveGamePlayoffGame, error)

	// GetSeasonBattingStats returns regular season batting stats for the given
	// season (rows linked via t_season_stats).
	GetSeasonBattingStats(ctx context.Context, seasonID int) ([]models.SaveGameBattingStat, error)

	// GetPlayoffBattingStats returns playoff batting stats for the given season
	// (rows linked via t_playoff_stats).
	GetPlayoffBattingStats(ctx context.Context, seasonID int) ([]models.SaveGameBattingStat, error)

	// GetSeasonPitchingStats returns regular season pitching stats for the given
	// season (rows linked via t_season_stats).
	GetSeasonPitchingStats(ctx context.Context, seasonID int) ([]models.SaveGamePitchingStat, error)

	// GetPlayoffPitchingStats returns playoff pitching stats for the given season
	// (rows linked via t_playoff_stats).
	GetPlayoffPitchingStats(ctx context.Context, seasonID int) ([]models.SaveGamePitchingStat, error)

	// GetCareerBattingStats returns career-aggregated batting stats for all
	// currently active players in the franchise.
	GetCareerBattingStats(ctx context.Context) ([]models.SaveGameBattingStat, error)

	// GetCareerPitchingStats returns career-aggregated pitching stats for all
	// currently active pitchers.
	GetCareerPitchingStats(ctx context.Context) ([]models.SaveGamePitchingStat, error)

	// GetCurrentSeason returns the most recent season for the franchise
	// identified by leagueGUID. If leagueGUID is empty (SMB3 single-league
	// saves), returns the latest season across all franchise seasons.
	GetCurrentSeason(ctx context.Context, leagueGUID string) (models.SaveGameSeasonInfo, error)

	// Close releases the underlying database connection.
	Close() error
}
