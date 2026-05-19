package store

import (
	"context"
	"fmt"
)

// Player is the base player record identified by its save game GUID.
type Player struct {
	ID           int64
	GameGUID     string
	FirstName    string
	LastName     string
	IsHallOfFamer bool
}

// PlayerSeason is the per-season snapshot of a player.
type PlayerSeason struct {
	ID              int64
	PlayerID        int64
	SeasonID        int
	TeamHistoryID   *int64
	Age             int
	Salary          int
	PrimaryPosition string
	SecondaryPosition string
	PitcherRole     string
	BatHand         string
	ThrowHand       string
	ChemistryType   string
	TraitsJSON      string
	PitchesJSON     string
}

// PlayerSeasonGameStats holds the 1-99 attribute snapshot for a player_season.
type PlayerSeasonGameStats struct {
	PlayerSeasonID int64
	Power          int
	Contact        int
	Speed          int
	Fielding       int
	Arm            int
	Velocity       int
	Junk           int
	Accuracy       int
}

// PlayerSeasonBattingStats holds raw batting counting stats.
type PlayerSeasonBattingStats struct {
	PlayerSeasonID  int64
	IsRegularSeason bool
	GamesPlayed     int
	GamesBatting    int
	AtBats          int
	Runs            int
	Hits            int
	Doubles         int
	Triples         int
	HomeRuns        int
	RBI             int
	StolenBases     int
	CaughtStealing  int
	Walks           int
	Strikeouts      int
	HitByPitch      int
	SacHits         int
	SacFlies        int
	Errors          int
	PassedBalls     int
}

// PlayerSeasonPitchingStats holds raw pitching counting stats.
type PlayerSeasonPitchingStats struct {
	PlayerSeasonID  int64
	IsRegularSeason bool
	Wins            int
	Losses          int
	Games           int
	GamesStarted    int
	CompleteGames   int
	Shutouts        int
	Saves           int
	OutsPitched     int
	HitsAllowed     int
	EarnedRuns      int
	HomeRunsAllowed int
	Walks           int
	Strikeouts      int
	HitBatters      int
	BattersFaced    int
	GamesFinished   int
	RunsAllowed     int
	WildPitches     int
	TotalPitches    int
}

// PlayerSeasonStore manages player and player_season records.
type PlayerSeasonStore struct {
	db DBTX
}

func NewPlayerSeasonStore(db DBTX) *PlayerSeasonStore {
	return &PlayerSeasonStore{db: db}
}

// UpsertPlayer inserts or returns the existing player for a given save game GUID.
// If the player already exists, first/last name are updated to the latest values.
func (s *PlayerSeasonStore) UpsertPlayer(ctx context.Context, p Player) (int64, error) {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO players (game_guid, first_name, last_name)
		VALUES (?, ?, ?)
		ON CONFLICT(game_guid) DO UPDATE SET
			first_name = excluded.first_name,
			last_name  = excluded.last_name
	`, p.GameGUID, p.FirstName, p.LastName)
	if err != nil {
		return 0, fmt.Errorf("upserting player %s: %w", p.GameGUID, err)
	}
	var id int64
	if err := s.db.QueryRowContext(ctx,
		`SELECT id FROM players WHERE game_guid = ?`, p.GameGUID,
	).Scan(&id); err != nil {
		return 0, fmt.Errorf("getting player id for %s: %w", p.GameGUID, err)
	}
	return id, nil
}

// UpsertSeason inserts or replaces a player_season record. Returns the ID.
func (s *PlayerSeasonStore) UpsertSeason(ctx context.Context, ps PlayerSeason) (int64, error) {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO player_seasons (
			player_id, season_id, team_history_id, age, salary,
			primary_position, secondary_position, pitcher_role,
			bat_hand, throw_hand, chemistry_type, traits_json, pitches_json
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(player_id, season_id) DO UPDATE SET
			team_history_id    = excluded.team_history_id,
			age                = excluded.age,
			salary             = excluded.salary,
			primary_position   = excluded.primary_position,
			secondary_position = excluded.secondary_position,
			pitcher_role       = excluded.pitcher_role,
			bat_hand           = excluded.bat_hand,
			throw_hand         = excluded.throw_hand,
			chemistry_type     = excluded.chemistry_type,
			traits_json        = excluded.traits_json,
			pitches_json       = excluded.pitches_json
	`,
		ps.PlayerID, ps.SeasonID, ps.TeamHistoryID, ps.Age, ps.Salary,
		ps.PrimaryPosition, ps.SecondaryPosition, ps.PitcherRole,
		ps.BatHand, ps.ThrowHand, ps.ChemistryType, ps.TraitsJSON, ps.PitchesJSON,
	)
	if err != nil {
		return 0, fmt.Errorf("upserting player season (player=%d season=%d): %w", ps.PlayerID, ps.SeasonID, err)
	}
	var id int64
	if err := s.db.QueryRowContext(ctx,
		`SELECT id FROM player_seasons WHERE player_id = ? AND season_id = ?`,
		ps.PlayerID, ps.SeasonID,
	).Scan(&id); err != nil {
		return 0, fmt.Errorf("getting player season id: %w", err)
	}
	return id, nil
}

// UpsertGameStats inserts or replaces player attribute stats for a season.
func (s *PlayerSeasonStore) UpsertGameStats(ctx context.Context, gs PlayerSeasonGameStats) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO player_season_game_stats (
			player_season_id, power, contact, speed, fielding, arm,
			velocity, junk, accuracy
		) VALUES (?,?,?,?,?,?,?,?,?)
		ON CONFLICT(player_season_id) DO UPDATE SET
			power    = excluded.power,
			contact  = excluded.contact,
			speed    = excluded.speed,
			fielding = excluded.fielding,
			arm      = excluded.arm,
			velocity = excluded.velocity,
			junk     = excluded.junk,
			accuracy = excluded.accuracy
	`,
		gs.PlayerSeasonID, gs.Power, gs.Contact, gs.Speed, gs.Fielding, gs.Arm,
		gs.Velocity, gs.Junk, gs.Accuracy,
	)
	if err != nil {
		return fmt.Errorf("upserting game stats for player_season %d: %w", gs.PlayerSeasonID, err)
	}
	return nil
}

// UpsertBattingStats inserts or replaces batting stats for a player_season.
func (s *PlayerSeasonStore) UpsertBattingStats(ctx context.Context, bs PlayerSeasonBattingStats) error {
	isReg := boolToInt(bs.IsRegularSeason)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO player_season_batting_stats (
			player_season_id, is_regular_season,
			games_played, games_batting, at_bats, runs, hits,
			doubles, triples, home_runs, rbi,
			stolen_bases, caught_stealing, walks, strikeouts, hit_by_pitch,
			sac_hits, sac_flies, errors, passed_balls
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(player_season_id, is_regular_season) DO UPDATE SET
			games_played    = excluded.games_played,
			games_batting   = excluded.games_batting,
			at_bats         = excluded.at_bats,
			runs            = excluded.runs,
			hits            = excluded.hits,
			doubles         = excluded.doubles,
			triples         = excluded.triples,
			home_runs       = excluded.home_runs,
			rbi             = excluded.rbi,
			stolen_bases    = excluded.stolen_bases,
			caught_stealing = excluded.caught_stealing,
			walks           = excluded.walks,
			strikeouts      = excluded.strikeouts,
			hit_by_pitch    = excluded.hit_by_pitch,
			sac_hits        = excluded.sac_hits,
			sac_flies       = excluded.sac_flies,
			errors          = excluded.errors,
			passed_balls    = excluded.passed_balls
	`,
		bs.PlayerSeasonID, isReg,
		bs.GamesPlayed, bs.GamesBatting, bs.AtBats, bs.Runs, bs.Hits,
		bs.Doubles, bs.Triples, bs.HomeRuns, bs.RBI,
		bs.StolenBases, bs.CaughtStealing, bs.Walks, bs.Strikeouts, bs.HitByPitch,
		bs.SacHits, bs.SacFlies, bs.Errors, bs.PassedBalls,
	)
	if err != nil {
		return fmt.Errorf("upserting batting stats for player_season %d (reg=%v): %w", bs.PlayerSeasonID, bs.IsRegularSeason, err)
	}
	return nil
}

// UpsertPitchingStats inserts or replaces pitching stats for a player_season.
func (s *PlayerSeasonStore) UpsertPitchingStats(ctx context.Context, ps PlayerSeasonPitchingStats) error {
	isReg := boolToInt(ps.IsRegularSeason)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO player_season_pitching_stats (
			player_season_id, is_regular_season,
			wins, losses, games, games_started, complete_games, shutouts, saves,
			outs_pitched, hits_allowed, earned_runs, home_runs_allowed,
			walks, strikeouts, hit_batters, batters_faced,
			games_finished, runs_allowed, wild_pitches, total_pitches
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(player_season_id, is_regular_season) DO UPDATE SET
			wins               = excluded.wins,
			losses             = excluded.losses,
			games              = excluded.games,
			games_started      = excluded.games_started,
			complete_games     = excluded.complete_games,
			shutouts           = excluded.shutouts,
			saves              = excluded.saves,
			outs_pitched       = excluded.outs_pitched,
			hits_allowed       = excluded.hits_allowed,
			earned_runs        = excluded.earned_runs,
			home_runs_allowed  = excluded.home_runs_allowed,
			walks              = excluded.walks,
			strikeouts         = excluded.strikeouts,
			hit_batters        = excluded.hit_batters,
			batters_faced      = excluded.batters_faced,
			games_finished     = excluded.games_finished,
			runs_allowed       = excluded.runs_allowed,
			wild_pitches       = excluded.wild_pitches,
			total_pitches      = excluded.total_pitches
	`,
		ps.PlayerSeasonID, isReg,
		ps.Wins, ps.Losses, ps.Games, ps.GamesStarted, ps.CompleteGames, ps.Shutouts, ps.Saves,
		ps.OutsPitched, ps.HitsAllowed, ps.EarnedRuns, ps.HomeRunsAllowed,
		ps.Walks, ps.Strikeouts, ps.HitBatters, ps.BattersFaced,
		ps.GamesFinished, ps.RunsAllowed, ps.WildPitches, ps.TotalPitches,
	)
	if err != nil {
		return fmt.Errorf("upserting pitching stats for player_season %d (reg=%v): %w", ps.PlayerSeasonID, ps.IsRegularSeason, err)
	}
	return nil
}
