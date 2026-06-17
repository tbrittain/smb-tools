package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"smb-tools/internal/models"
)

// leagueGUIDReference is one column that holds a league's own GUID,
// somewhere in a league save. Confirmed exhaustive via PRAGMA
// foreign_key_list against a real league save — see
// docs/league-transfer/validation-results.md. RewriteLeagueGUID updates
// exactly these columns/tables, nothing else.
type leagueGUIDReference struct {
	table  string
	column string
}

var leagueGUIDReferences = []leagueGUIDReference{
	{"t_leagues", "GUID"},
	{"t_leagues", "originalGUID"},
	{"t_conferences", "leagueGUID"},
	{"t_franchise", "leagueGUID"},
	{"t_seasons", "historicalLeagueGUID"},
	{"t_league_local_ids", "GUID"},
}

// requiredLeagueSaveTables are the tables a decompressed league .sav must
// contain to be accepted as a valid league save for import. This checks
// table presence only, not row presence — a stock or season/elimination-mode
// league legitimately has zero rows in t_franchise, and must still pass.
var requiredLeagueSaveTables = []string{
	"t_leagues",
	"t_teams",
	"t_franchise",
	"t_conferences",
	"t_divisions",
	"t_division_teams",
	"t_league_local_ids",
}

// LeagueSaveStore handles reads and writes against an already-decompressed,
// already-opened league .sav SQLite connection.
type LeagueSaveStore struct {
	db DBTX
}

func NewLeagueSaveStore(db DBTX) *LeagueSaveStore {
	return &LeagueSaveStore{db: db}
}

// RewriteLeagueGUID replaces every occurrence of oldGUID across the 6
// confirmed GUID-bearing columns with newGUID, in a single transaction so a
// failure partway through leaves the save file's GUIDs entirely unchanged
// rather than partially rewritten. db must be a *sql.DB (RewriteLeagueGUID
// manages its own transaction); passing a *sql.Tx is not supported.
func (s *LeagueSaveStore) RewriteLeagueGUID(ctx context.Context, db *sql.DB, oldGUID, newGUID uuid.UUID) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning GUID rewrite transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, ref := range leagueGUIDReferences {
		query := fmt.Sprintf("UPDATE [%s] SET [%s] = ? WHERE [%s] = ?", ref.table, ref.column, ref.column)
		if _, err := tx.ExecContext(ctx, query, newGUID[:], oldGUID[:]); err != nil {
			return fmt.Errorf("rewriting %s.%s: %w", ref.table, ref.column, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing GUID rewrite: %w", err)
	}
	return nil
}

// ValidateLeagueSaveShape checks that all tables a league save is expected
// to have are actually present. It is a shape/sanity check, not a security
// boundary — see docs/league-transfer/ux-flow.md's import disclaimers.
func (s *LeagueSaveStore) ValidateLeagueSaveShape(ctx context.Context) error {
	for _, table := range requiredLeagueSaveTables {
		var count int
		err := s.db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?
		`, table).Scan(&count)
		if err != nil {
			return fmt.Errorf("checking for table %s: %w", table, err)
		}
		if count == 0 {
			return fmt.Errorf("missing expected table %s — this does not look like a valid league save", table)
		}
	}
	return nil
}

// GetLeagueOverview returns the structure (name, conferences, divisions,
// teams) of the league identified by guid.
//
// Known limitation: teams are resolved via t_division_teams, the only
// confirmed team-to-division/conference linkage (see
// docs/league-transfer/validation-results.md). A conference with zero
// divisions — a real, valid configuration per
// docs/league-transfer/ux-flow.md — will show no teams under it, because no
// alternative direct conference-to-team linkage has been confirmed against a
// real save. This is flagged as an open question in
// docs/league-transfer/implementation-plan.md, not silently assumed away.
func (s *LeagueSaveStore) GetLeagueOverview(ctx context.Context, guid uuid.UUID) (models.LeagueOverview, error) {
	var overview models.LeagueOverview
	err := s.db.QueryRowContext(ctx, `
		SELECT name FROM t_leagues WHERE GUID = ?
	`, guid[:]).Scan(&overview.Name)
	if err != nil {
		return models.LeagueOverview{}, fmt.Errorf("loading league %s: %w", guid, err)
	}
	overview.GUID = guid

	conferences, err := s.getConferences(ctx, guid)
	if err != nil {
		return models.LeagueOverview{}, err
	}
	overview.Conferences = conferences

	return overview, nil
}

// guidNamePair is scanned from any of the GUID/name queries below. Each
// level's rows are fully drained into a slice of these before any nested
// query is issued for the next level down — issuing a nested query while
// the outer *sql.Rows is still open can hand out a second pooled connection
// to a *different* anonymous ":memory:" database, silently losing the
// schema. Draining first avoids that entirely.
type guidNamePair struct {
	guid []byte
	name string
}

func (s *LeagueSaveStore) getConferences(ctx context.Context, leagueGUID uuid.UUID) ([]models.ConferenceOverview, error) {
	rawConferences, err := s.queryGUIDNamePairs(ctx, `
		SELECT GUID, COALESCE(name, '') FROM t_conferences WHERE leagueGUID = ?
	`, leagueGUID[:])
	if err != nil {
		return nil, fmt.Errorf("querying conferences for league %s: %w", leagueGUID, err)
	}

	var conferences []models.ConferenceOverview
	for _, raw := range rawConferences {
		conferenceGUID, err := uuid.FromBytes(raw.guid)
		if err != nil {
			return nil, fmt.Errorf("parsing conference GUID: %w", err)
		}

		divisions, err := s.getDivisions(ctx, conferenceGUID)
		if err != nil {
			return nil, err
		}

		conferences = append(conferences, models.ConferenceOverview{
			GUID:      conferenceGUID,
			Name:      raw.name,
			Divisions: divisions,
		})
	}
	return conferences, nil
}

func (s *LeagueSaveStore) getDivisions(ctx context.Context, conferenceGUID uuid.UUID) ([]models.DivisionOverview, error) {
	rawDivisions, err := s.queryGUIDNamePairs(ctx, `
		SELECT GUID, COALESCE(name, '') FROM t_divisions WHERE conferenceGUID = ?
	`, conferenceGUID[:])
	if err != nil {
		return nil, fmt.Errorf("querying divisions for conference %s: %w", conferenceGUID, err)
	}

	var divisions []models.DivisionOverview
	for _, raw := range rawDivisions {
		divisionGUID, err := uuid.FromBytes(raw.guid)
		if err != nil {
			return nil, fmt.Errorf("parsing division GUID: %w", err)
		}

		teams, err := s.getTeamsInDivision(ctx, divisionGUID)
		if err != nil {
			return nil, err
		}

		divisions = append(divisions, models.DivisionOverview{
			GUID:  divisionGUID,
			Name:  raw.name,
			Teams: teams,
		})
	}
	return divisions, nil
}

// queryGUIDNamePairs runs a "SELECT GUID, name ... WHERE ? = ?"-shaped query
// and fully drains the result into memory before returning, so callers can
// safely issue further queries against s.db per row without risking a
// second pooled connection seeing an empty database (see guidNamePair).
func (s *LeagueSaveStore) queryGUIDNamePairs(ctx context.Context, query string, arg any) ([]guidNamePair, error) {
	rows, err := s.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var pairs []guidNamePair
	for rows.Next() {
		var p guidNamePair
		if err := rows.Scan(&p.guid, &p.name); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		pairs = append(pairs, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}
	return pairs, nil
}

func (s *LeagueSaveStore) getTeamsInDivision(ctx context.Context, divisionGUID uuid.UUID) ([]models.TeamOverview, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.GUID, t.teamName
		FROM t_division_teams dt
		JOIN t_teams t ON t.GUID = dt.teamGUID
		WHERE dt.divisionGUID = ?
	`, divisionGUID[:])
	if err != nil {
		return nil, fmt.Errorf("querying teams for division %s: %w", divisionGUID, err)
	}
	defer func() { _ = rows.Close() }()

	var teams []models.TeamOverview
	for rows.Next() {
		var (
			guidBytes []byte
			name      string
		)
		if err := rows.Scan(&guidBytes, &name); err != nil {
			return nil, fmt.Errorf("scanning team row: %w", err)
		}
		teamGUID, err := uuid.FromBytes(guidBytes)
		if err != nil {
			return nil, fmt.Errorf("parsing team GUID: %w", err)
		}
		teams = append(teams, models.TeamOverview{GUID: teamGUID, Name: name})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating teams: %w", err)
	}
	return teams, nil
}
