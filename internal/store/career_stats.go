package store

import (
	"context"
	"fmt"

	"smb-tools/internal/models"
)

// CareerStatsStore manages career batting and pitching stat rows in
// player_career_batting_stats and player_career_pitching_stats.
type CareerStatsStore struct {
	db DBTX
}

func NewCareerStatsStore(db DBTX) *CareerStatsStore {
	return &CareerStatsStore{db: db}
}

// CareerBattingAggregate holds career batting counting totals and the
// career-weighted league numerators/denominators needed to compute career OPS+.
// SmbWARSum is the sum of per-season smb_war values.
type CareerBattingAggregate struct {
	// Counting totals
	GamesPlayed    int64
	GamesBatting   int64
	AtBats         int64
	Runs           int64
	Hits           int64
	Doubles        int64
	Triples        int64
	HomeRuns       int64
	RBI            int64
	StolenBases    int64
	CaughtStealing int64
	Walks          int64
	Strikeouts     int64
	HitByPitch     int64
	SacHits        int64
	SacFlies       int64
	Errors         int64
	PassedBalls    int64
	SeasonsPlayed  int64
	SmbWARSum      *float64

	// Career-weighted league batting totals for ComputeLeagueStats.
	LgAtBats   int64
	LgHits     int64
	LgDoubles  int64
	LgTriples  int64
	LgHomeRuns int64
	LgWalks    int64
	LgHBP      int64
	LgSacFlies int64
}

// CareerPitchingAggregate holds career pitching counting totals and the
// career-weighted league numerators/denominators needed to compute career ERA+/FIP.
// SmbWARSum is the sum of per-season smb_war values.
type CareerPitchingAggregate struct {
	// Counting totals
	Wins            int64
	Losses          int64
	Games           int64
	GamesStarted    int64
	CompleteGames   int64
	Shutouts        int64
	Saves           int64
	OutsPitched     int64
	HitsAllowed     int64
	EarnedRuns      int64
	HomeRunsAllowed int64
	Walks           int64
	Strikeouts      int64
	HitBatters      int64
	BattersFaced    int64
	GamesFinished   int64
	RunsAllowed     int64
	WildPitches     int64
	TotalPitches    int64
	SeasonsPlayed   int64
	SmbWARSum       *float64

	// Career-weighted league pitching totals for ComputeLeagueStats.
	LgOutsPitched int64
	LgEarnedRuns  int64
	LgHRAllowed   int64
	LgBBAllowed   int64
	LgHBPAllowed  int64
	LgKPitched    int64
}

// isRegFilter returns the SQL fragment and args for filtering by stat_type.
// For total_career the filter is omitted (all rows included).
func isRegFilter(statType models.CareerStatType) (string, []any) {
	switch statType {
	case models.CareerStatTypeRegularSeason:
		return " AND bs.is_regular_season = 1", nil
	case models.CareerStatTypePlayoffs:
		return " AND bs.is_regular_season = 0", nil
	default: // total_career — no filter
		return "", nil
	}
}

func isRegFilterPitching(statType models.CareerStatType) (string, []any) {
	switch statType {
	case models.CareerStatTypeRegularSeason:
		return " AND pit.is_regular_season = 1", nil
	case models.CareerStatTypePlayoffs:
		return " AND pit.is_regular_season = 0", nil
	default:
		return "", nil
	}
}

// GetCareerBattingTotalsWithLeague aggregates career batting counting stats and
// the corresponding league batting totals for the given player and stat type.
// Returns nil when the player has no batting rows for this stat type.
func (s *CareerStatsStore) GetCareerBattingTotalsWithLeague(
	ctx context.Context,
	playerID int64,
	statType models.CareerStatType,
) (*CareerBattingAggregate, error) {
	regFilter, _ := isRegFilter(statType)

	//nolint:gosec // statType values are controlled constants, not user input
	row := s.db.QueryRowContext(ctx, `
SELECT
    COUNT(DISTINCT ps.season_id),
    COALESCE(SUM(bs.games_played),    0),
    COALESCE(SUM(bs.games_batting),   0),
    COALESCE(SUM(bs.at_bats),         0),
    COALESCE(SUM(bs.runs),            0),
    COALESCE(SUM(bs.hits),            0),
    COALESCE(SUM(bs.doubles),         0),
    COALESCE(SUM(bs.triples),         0),
    COALESCE(SUM(bs.home_runs),       0),
    COALESCE(SUM(bs.rbi),             0),
    COALESCE(SUM(bs.stolen_bases),    0),
    COALESCE(SUM(bs.caught_stealing), 0),
    COALESCE(SUM(bs.walks),           0),
    COALESCE(SUM(bs.strikeouts),      0),
    COALESCE(SUM(bs.hit_by_pitch),    0),
    COALESCE(SUM(bs.sac_hits),        0),
    COALESCE(SUM(bs.sac_flies),       0),
    COALESCE(SUM(bs.errors),          0),
    COALESCE(SUM(bs.passed_balls),    0),
    SUM(bs.smb_war),
    COALESCE(SUM(lss.total_at_bats),  0),
    COALESCE(SUM(lss.total_hits),     0),
    COALESCE(SUM(lss.total_doubles),  0),
    COALESCE(SUM(lss.total_triples),  0),
    COALESCE(SUM(lss.total_home_runs),0),
    COALESCE(SUM(lss.total_walks),    0),
    COALESCE(SUM(lss.total_hbp),      0),
    COALESCE(SUM(lss.total_sac_flies),0)
FROM player_season_batting_stats bs
JOIN player_seasons ps ON ps.id = bs.player_season_id
LEFT JOIN league_season_stats lss
    ON lss.season_id = ps.season_id
   AND lss.is_regular_season = bs.is_regular_season
WHERE ps.player_id = ?`+regFilter,
		playerID,
	)

	var a CareerBattingAggregate
	if err := row.Scan(
		&a.SeasonsPlayed,
		&a.GamesPlayed, &a.GamesBatting, &a.AtBats, &a.Runs, &a.Hits,
		&a.Doubles, &a.Triples, &a.HomeRuns, &a.RBI,
		&a.StolenBases, &a.CaughtStealing, &a.Walks, &a.Strikeouts, &a.HitByPitch,
		&a.SacHits, &a.SacFlies, &a.Errors, &a.PassedBalls,
		&a.SmbWARSum,
		&a.LgAtBats, &a.LgHits, &a.LgDoubles, &a.LgTriples, &a.LgHomeRuns,
		&a.LgWalks, &a.LgHBP, &a.LgSacFlies,
	); err != nil {
		return nil, fmt.Errorf("career batting aggregate (player=%d type=%s): %w", playerID, statType, err)
	}
	if a.AtBats == 0 && a.GamesPlayed == 0 {
		return nil, nil
	}
	return &a, nil
}

// GetCareerPitchingTotalsWithLeague aggregates career pitching counting stats and
// the corresponding league pitching totals for the given player and stat type.
// Returns nil when the player has no pitching rows for this stat type.
func (s *CareerStatsStore) GetCareerPitchingTotalsWithLeague(
	ctx context.Context,
	playerID int64,
	statType models.CareerStatType,
) (*CareerPitchingAggregate, error) {
	regFilter, _ := isRegFilterPitching(statType)

	//nolint:gosec // statType values are controlled constants, not user input
	row := s.db.QueryRowContext(ctx, `
SELECT
    COUNT(DISTINCT ps.season_id),
    COALESCE(SUM(pit.wins),              0),
    COALESCE(SUM(pit.losses),            0),
    COALESCE(SUM(pit.games),             0),
    COALESCE(SUM(pit.games_started),     0),
    COALESCE(SUM(pit.complete_games),    0),
    COALESCE(SUM(pit.shutouts),          0),
    COALESCE(SUM(pit.saves),             0),
    COALESCE(SUM(pit.outs_pitched),      0),
    COALESCE(SUM(pit.hits_allowed),      0),
    COALESCE(SUM(pit.earned_runs),       0),
    COALESCE(SUM(pit.home_runs_allowed), 0),
    COALESCE(SUM(pit.walks),             0),
    COALESCE(SUM(pit.strikeouts),        0),
    COALESCE(SUM(pit.hit_batters),       0),
    COALESCE(SUM(pit.batters_faced),     0),
    COALESCE(SUM(pit.games_finished),    0),
    COALESCE(SUM(pit.runs_allowed),      0),
    COALESCE(SUM(pit.wild_pitches),      0),
    COALESCE(SUM(pit.total_pitches),     0),
    SUM(pit.smb_war),
    COALESCE(SUM(lss.total_outs_pitched),0),
    COALESCE(SUM(lss.total_earned_runs), 0),
    COALESCE(SUM(lss.total_hr_allowed),  0),
    COALESCE(SUM(lss.total_bb_allowed),  0),
    COALESCE(SUM(lss.total_hbp_allowed), 0),
    COALESCE(SUM(lss.total_k_pitched),   0)
FROM player_season_pitching_stats pit
JOIN player_seasons ps ON ps.id = pit.player_season_id
LEFT JOIN league_season_stats lss
    ON lss.season_id = ps.season_id
   AND lss.is_regular_season = pit.is_regular_season
WHERE ps.player_id = ?`+regFilter,
		playerID,
	)

	var a CareerPitchingAggregate
	if err := row.Scan(
		&a.SeasonsPlayed,
		&a.Wins, &a.Losses, &a.Games, &a.GamesStarted, &a.CompleteGames,
		&a.Shutouts, &a.Saves, &a.OutsPitched, &a.HitsAllowed, &a.EarnedRuns,
		&a.HomeRunsAllowed, &a.Walks, &a.Strikeouts, &a.HitBatters, &a.BattersFaced,
		&a.GamesFinished, &a.RunsAllowed, &a.WildPitches, &a.TotalPitches,
		&a.SmbWARSum,
		&a.LgOutsPitched, &a.LgEarnedRuns, &a.LgHRAllowed,
		&a.LgBBAllowed, &a.LgHBPAllowed, &a.LgKPitched,
	); err != nil {
		return nil, fmt.Errorf("career pitching aggregate (player=%d type=%s): %w", playerID, statType, err)
	}
	if a.OutsPitched == 0 && a.Games == 0 {
		return nil, nil
	}
	return &a, nil
}

// UpsertCareerBattingStats inserts or replaces a career batting stat row.
func (s *CareerStatsStore) UpsertCareerBattingStats(
	ctx context.Context,
	playerID int64,
	statType models.CareerStatType,
	seasonsPlayed int,
	b *models.CareerBattingStats,
) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO player_career_batting_stats (
    player_id, stat_type, seasons_played,
    games_played, games_batting, at_bats, runs, hits,
    doubles, triples, home_runs, rbi,
    stolen_bases, caught_stealing, walks, strikeouts, hit_by_pitch,
    sac_hits, sac_flies, errors, passed_balls,
    ba, obp, slg, ops, iso, babip, k_pct, bb_pct, ab_per_hr,
    ops_plus, smb_war
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
ON CONFLICT(player_id, stat_type) DO UPDATE SET
    seasons_played  = excluded.seasons_played,
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
    passed_balls    = excluded.passed_balls,
    ba              = excluded.ba,
    obp             = excluded.obp,
    slg             = excluded.slg,
    ops             = excluded.ops,
    iso             = excluded.iso,
    babip           = excluded.babip,
    k_pct           = excluded.k_pct,
    bb_pct          = excluded.bb_pct,
    ab_per_hr       = excluded.ab_per_hr,
    ops_plus        = excluded.ops_plus,
    smb_war         = excluded.smb_war
`,
		playerID, string(statType), seasonsPlayed,
		b.GamesPlayed, b.GamesBatting, b.AtBats, b.Runs, b.Hits,
		b.Doubles, b.Triples, b.HomeRuns, b.RBI,
		b.StolenBases, b.CaughtStealing, b.Walks, b.Strikeouts, b.HitByPitch,
		b.SacHits, b.SacFlies, b.Errors, b.PassedBalls,
		b.BA, b.OBP, b.SLG, b.OPS, b.ISO, b.BABIP, b.KPct, b.BBPct, b.ABPerHR,
		b.OPSPlus, b.SmbWAR,
	)
	if err != nil {
		return fmt.Errorf("upserting career batting stats (player=%d type=%s): %w", playerID, statType, err)
	}
	return nil
}

// UpsertCareerPitchingStats inserts or replaces a career pitching stat row.
func (s *CareerStatsStore) UpsertCareerPitchingStats(
	ctx context.Context,
	playerID int64,
	statType models.CareerStatType,
	seasonsPlayed int,
	p *models.CareerPitchingStats,
) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO player_career_pitching_stats (
    player_id, stat_type, seasons_played,
    wins, losses, games, games_started, complete_games, shutouts, saves,
    outs_pitched, hits_allowed, earned_runs, home_runs_allowed,
    walks, strikeouts, hit_batters, batters_faced,
    games_finished, runs_allowed, wild_pitches, total_pitches,
    era, whip, k_per_9, bb_per_9, h_per_9, hr_per_9, k_per_bb, k_pct, win_pct, p_per_ip,
    era_plus, fip, fip_minus, smb_war
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
ON CONFLICT(player_id, stat_type) DO UPDATE SET
    seasons_played     = excluded.seasons_played,
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
    p_per_ip           = excluded.p_per_ip,
    era_plus           = excluded.era_plus,
    fip                = excluded.fip,
    fip_minus          = excluded.fip_minus,
    smb_war            = excluded.smb_war
`,
		playerID, string(statType), seasonsPlayed,
		p.Wins, p.Losses, p.Games, p.GamesStarted, p.CompleteGames, p.Shutouts, p.Saves,
		p.OutsPitched, p.HitsAllowed, p.EarnedRuns, p.HomeRunsAllowed,
		p.Walks, p.Strikeouts, p.HitBatters, p.BattersFaced,
		p.GamesFinished, p.RunsAllowed, p.WildPitches, p.TotalPitches,
		p.ERA, p.WHIP, p.K9, p.BB9, p.H9, p.HR9, p.KPerBB, p.KPct, p.WinPct, p.PPerIP,
		p.ERAPlus, p.FIP, p.FIPMinus, p.SmbWAR,
	)
	if err != nil {
		return fmt.Errorf("upserting career pitching stats (player=%d type=%s): %w", playerID, statType, err)
	}
	return nil
}
