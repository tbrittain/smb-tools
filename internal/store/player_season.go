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
	ID                int64
	PlayerID          int64
	SeasonID          int64
	Age               int
	Salary            int
	PrimaryPosition   string
	SecondaryPosition string
	PitcherRole       string
	BatHand           string
	ThrowHand         string
	ChemistryType     string
	TraitsJSON        string
	PitchesJSON       string
}

// PlayerSeasonTeam links a player_season to one team in the junction table.
// SortOrder: 0=current/final team, 1=most recently played prior, 2=two teams ago.
type PlayerSeasonTeam struct {
	PlayerSeasonID int64
	TeamHistoryID  int64
	SortOrder      int
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

// PlayerSeasonBattingStats holds batting counting stats and pre-computed rate stats.
// Rate fields are nil when the denominator is zero.
type PlayerSeasonBattingStats struct {
	PlayerSeasonID  int64
	IsRegularSeason bool
	GamesPlayed     int
	GamesBatting    int
	AtBats          int
	PlateAppearances int
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
	// Simple rate stats — computed by the service layer before insert.
	BA      *float64
	OBP     *float64
	SLG     *float64
	OPS     *float64
	ISO     *float64
	BABIP   *float64
	KPct    *float64
	BBPct   *float64
	ABPerHR *float64
}

// PlayerSeasonPitchingStats holds pitching counting stats and pre-computed rate stats.
// Rate fields are nil when the denominator is zero.
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
	// Simple rate stats — computed by the service layer before insert.
	ERA     *float64
	WHIP    *float64
	K9      *float64
	BB9     *float64
	H9      *float64
	HR9     *float64
	KPerBB  *float64
	KPct    *float64
	WinPct  *float64
	PPerIP  *float64
}

// PlayerIdentity holds the GUID for primary lookup plus the semantic fields
// used for fuzzy re-association when a franchise is forked and player GUIDs change.
type PlayerIdentity struct {
	GameGUID      string
	FirstName     string
	LastName      string
	BatHand       string
	ThrowHand     string
	ChemistryType string
}

// PlayerSeasonStore manages player and player_season records.
type PlayerSeasonStore struct {
	db DBTX
}

func NewPlayerSeasonStore(db DBTX) *PlayerSeasonStore {
	return &PlayerSeasonStore{db: db}
}

// UpsertPlayer resolves or creates a player record using a three-tier lookup:
//  1. Exact match on players.game_guid.
//  2. Exact match on player_alt_guids.game_guid (GUIDs from prior forks).
//  3. Fuzzy match on first_name + last_name + bat_hand + throw_hand + chemistry_type
//     (handles the case where a league fork assigned new GUIDs to existing players).
//     On a fuzzy match, the new GUID is added to player_alt_guids so subsequent
//     imports skip the fuzzy scan.
//
// If none of the three tiers match, a new player row is inserted.
// Always updates first_name/last_name to the latest values from the save game.
func (s *PlayerSeasonStore) UpsertPlayer(ctx context.Context, identity PlayerIdentity) (int64, error) {
	// Tier 1: primary GUID
	var id int64
	err := s.db.QueryRowContext(ctx,
		`SELECT id FROM players WHERE game_guid = ?`, identity.GameGUID,
	).Scan(&id)
	if err == nil {
		_, _ = s.db.ExecContext(ctx,
			`UPDATE players SET first_name = ?, last_name = ? WHERE id = ?`,
			identity.FirstName, identity.LastName, id)
		return id, nil
	}

	// Tier 2: alt GUID table
	err = s.db.QueryRowContext(ctx,
		`SELECT player_id FROM player_alt_guids WHERE game_guid = ?`, identity.GameGUID,
	).Scan(&id)
	if err == nil {
		_, _ = s.db.ExecContext(ctx,
			`UPDATE players SET first_name = ?, last_name = ? WHERE id = ?`,
			identity.FirstName, identity.LastName, id)
		return id, nil
	}

	// Tier 3: fuzzy match — only when chemistry and handedness are present
	if identity.BatHand != "" && identity.ThrowHand != "" && identity.ChemistryType != "" {
		err = s.db.QueryRowContext(ctx, `
			SELECT p.id FROM players p
			JOIN player_seasons ps ON ps.player_id = p.id
			WHERE p.first_name     = ?
			  AND p.last_name      = ?
			  AND ps.bat_hand      = ?
			  AND ps.throw_hand    = ?
			  AND ps.chemistry_type = ?
			LIMIT 1
		`, identity.FirstName, identity.LastName,
			identity.BatHand, identity.ThrowHand, identity.ChemistryType,
		).Scan(&id)
		if err == nil {
			// Register the new GUID so future lookups hit tier 1/2
			_, _ = s.db.ExecContext(ctx,
				`INSERT OR IGNORE INTO player_alt_guids (player_id, game_guid) VALUES (?, ?)`,
				id, identity.GameGUID)
			_, _ = s.db.ExecContext(ctx,
				`UPDATE players SET first_name = ?, last_name = ? WHERE id = ?`,
				identity.FirstName, identity.LastName, id)
			return id, nil
		}
	}

	// No match — new player
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO players (game_guid, first_name, last_name) VALUES (?, ?, ?)`,
		identity.GameGUID, identity.FirstName, identity.LastName)
	if err != nil {
		return 0, fmt.Errorf("inserting player %s: %w", identity.GameGUID, err)
	}
	newID, _ := res.LastInsertId()
	return newID, nil
}

// UpsertSeason inserts or replaces a player_season record. Returns the ID.
func (s *PlayerSeasonStore) UpsertSeason(ctx context.Context, ps PlayerSeason) (int64, error) {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO player_seasons (
			player_id, season_id, age, salary,
			primary_position, secondary_position, pitcher_role,
			bat_hand, throw_hand, chemistry_type, traits_json, pitches_json
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(player_id, season_id) DO UPDATE SET
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
		ps.PlayerID, ps.SeasonID, ps.Age, ps.Salary,
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

// ReplaceSeasonTeams atomically replaces all team associations for a player_season.
// Passing an empty slice leaves the player as a free agent for that season.
func (s *PlayerSeasonStore) ReplaceSeasonTeams(ctx context.Context, playerSeasonID int64, teams []PlayerSeasonTeam) error {
	if _, err := s.db.ExecContext(ctx,
		`DELETE FROM player_season_teams WHERE player_season_id = ?`, playerSeasonID,
	); err != nil {
		return fmt.Errorf("clearing season teams for player_season %d: %w", playerSeasonID, err)
	}
	for _, t := range teams {
		if _, err := s.db.ExecContext(ctx,
			`INSERT OR IGNORE INTO player_season_teams (player_season_id, team_history_id, sort_order) VALUES (?,?,?)`,
			playerSeasonID, t.TeamHistoryID, t.SortOrder,
		); err != nil {
			return fmt.Errorf("inserting season team (player_season=%d team_history=%d): %w",
				playerSeasonID, t.TeamHistoryID, err)
		}
	}
	return nil
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

// UpsertBattingStats inserts or replaces batting stats for a player_season,
// including pre-computed simple rate stats.
func (s *PlayerSeasonStore) UpsertBattingStats(ctx context.Context, bs PlayerSeasonBattingStats) error {
	isReg := boolToInt(bs.IsRegularSeason)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO player_season_batting_stats (
			player_season_id, is_regular_season,
			games_played, games_batting, at_bats, plate_appearances, runs, hits,
			doubles, triples, home_runs, rbi,
			stolen_bases, caught_stealing, walks, strikeouts, hit_by_pitch,
			sac_hits, sac_flies, errors, passed_balls,
			ba, obp, slg, ops, iso, babip, k_pct, bb_pct, ab_per_hr
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(player_season_id, is_regular_season) DO UPDATE SET
			games_played      = excluded.games_played,
			games_batting     = excluded.games_batting,
			at_bats           = excluded.at_bats,
			plate_appearances = excluded.plate_appearances,
			runs              = excluded.runs,
			hits              = excluded.hits,
			doubles           = excluded.doubles,
			triples           = excluded.triples,
			home_runs         = excluded.home_runs,
			rbi               = excluded.rbi,
			stolen_bases      = excluded.stolen_bases,
			caught_stealing   = excluded.caught_stealing,
			walks             = excluded.walks,
			strikeouts        = excluded.strikeouts,
			hit_by_pitch      = excluded.hit_by_pitch,
			sac_hits          = excluded.sac_hits,
			sac_flies         = excluded.sac_flies,
			errors            = excluded.errors,
			passed_balls      = excluded.passed_balls,
			ba                = excluded.ba,
			obp               = excluded.obp,
			slg               = excluded.slg,
			ops               = excluded.ops,
			iso               = excluded.iso,
			babip             = excluded.babip,
			k_pct             = excluded.k_pct,
			bb_pct            = excluded.bb_pct,
			ab_per_hr         = excluded.ab_per_hr
	`,
		bs.PlayerSeasonID, isReg,
		bs.GamesPlayed, bs.GamesBatting, bs.AtBats, bs.PlateAppearances, bs.Runs, bs.Hits,
		bs.Doubles, bs.Triples, bs.HomeRuns, bs.RBI,
		bs.StolenBases, bs.CaughtStealing, bs.Walks, bs.Strikeouts, bs.HitByPitch,
		bs.SacHits, bs.SacFlies, bs.Errors, bs.PassedBalls,
		bs.BA, bs.OBP, bs.SLG, bs.OPS, bs.ISO, bs.BABIP, bs.KPct, bs.BBPct, bs.ABPerHR,
	)
	if err != nil {
		return fmt.Errorf("upserting batting stats for player_season %d (reg=%v): %w", bs.PlayerSeasonID, bs.IsRegularSeason, err)
	}
	return nil
}

// UpsertPitchingStats inserts or replaces pitching stats for a player_season,
// including pre-computed simple rate stats.
func (s *PlayerSeasonStore) UpsertPitchingStats(ctx context.Context, ps PlayerSeasonPitchingStats) error {
	isReg := boolToInt(ps.IsRegularSeason)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO player_season_pitching_stats (
			player_season_id, is_regular_season,
			wins, losses, games, games_started, complete_games, shutouts, saves,
			outs_pitched, hits_allowed, earned_runs, home_runs_allowed,
			walks, strikeouts, hit_batters, batters_faced,
			games_finished, runs_allowed, wild_pitches, total_pitches,
			era, whip, k_per_9, bb_per_9, h_per_9, hr_per_9, k_per_bb, k_pct, win_pct, p_per_ip
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
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
			total_pitches      = excluded.total_pitches,
			era                = excluded.era,
			whip               = excluded.whip,
			k_per_9            = excluded.k_per_9,
			bb_per_9           = excluded.bb_per_9,
			h_per_9            = excluded.h_per_9,
			hr_per_9           = excluded.hr_per_9,
			k_per_bb           = excluded.k_per_bb,
			k_pct              = excluded.k_pct,
			win_pct            = excluded.win_pct,
			p_per_ip           = excluded.p_per_ip
	`,
		ps.PlayerSeasonID, isReg,
		ps.Wins, ps.Losses, ps.Games, ps.GamesStarted, ps.CompleteGames, ps.Shutouts, ps.Saves,
		ps.OutsPitched, ps.HitsAllowed, ps.EarnedRuns, ps.HomeRunsAllowed,
		ps.Walks, ps.Strikeouts, ps.HitBatters, ps.BattersFaced,
		ps.GamesFinished, ps.RunsAllowed, ps.WildPitches, ps.TotalPitches,
		ps.ERA, ps.WHIP, ps.K9, ps.BB9, ps.H9, ps.HR9, ps.KPerBB, ps.KPct, ps.WinPct, ps.PPerIP,
	)
	if err != nil {
		return fmt.Errorf("upserting pitching stats for player_season %d (reg=%v): %w", ps.PlayerSeasonID, ps.IsRegularSeason, err)
	}
	return nil
}
