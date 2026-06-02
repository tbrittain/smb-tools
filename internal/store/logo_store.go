package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"smb-tools/internal/models"
)

// TeamLogoWithAssignments bundles a logo with all of its assignment records.
type TeamLogoWithAssignments struct {
	Logo        models.TeamLogo
	Assignments []models.TeamLogoAssignment
}

// LogoStore handles reads and writes for the logos and logo_assignments tables.
type LogoStore struct{}

func NewLogoStore() *LogoStore {
	return &LogoStore{}
}

// InsertLogo inserts a new logo record.
func (s *LogoStore) InsertLogo(ctx context.Context, db DBTX, logo models.TeamLogo) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO logos (id, team_id, file_path, uploaded_at)
		VALUES (?, ?, ?, ?)
	`,
		logo.ID,
		logo.TeamID,
		logo.FilePath,
		logo.UploadedAt.UTC().Format("2006-01-02T15:04:05Z"),
	)
	if err != nil {
		return fmt.Errorf("inserting logo: %w", err)
	}
	return nil
}

// InsertLogoAssignment inserts a new logo assignment record.
func (s *LogoStore) InsertLogoAssignment(ctx context.Context, db DBTX, a models.TeamLogoAssignment) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO logo_assignments (id, logo_id, start_season, end_season, assigned_at)
		VALUES (?, ?, ?, ?, ?)
	`,
		a.ID,
		a.LogoID,
		a.StartSeason,
		a.EndSeason,
		a.AssignedAt.UTC().Format("2006-01-02T15:04:05Z"),
	)
	if err != nil {
		return fmt.Errorf("inserting logo assignment: %w", err)
	}
	return nil
}

// GetLogoForSeason returns the logo whose assignment covers seasonNum for the given team,
// resolved by last-write-wins (most recent assigned_at). Returns nil when no logo applies.
func (s *LogoStore) GetLogoForSeason(ctx context.Context, db DBTX, teamID int, seasonNum int) (*models.TeamLogo, error) {
	var logo models.TeamLogo
	var uploadedAt string
	err := db.QueryRowContext(ctx, `
		SELECT l.id, l.team_id, l.file_path, l.uploaded_at
		FROM logos l
		JOIN logo_assignments la ON la.logo_id = l.id
		WHERE l.team_id = ?
		  AND (la.start_season IS NULL OR la.start_season <= ?)
		  AND (la.end_season IS NULL OR la.end_season >= ?)
		ORDER BY la.assigned_at DESC
		LIMIT 1
	`, teamID, seasonNum, seasonNum).Scan(
		&logo.ID, &logo.TeamID, &logo.FilePath, &uploadedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting logo for team %d season %d: %w", teamID, seasonNum, err)
	}
	logo.UploadedAt, _ = time.Parse("2006-01-02T15:04:05Z", uploadedAt)
	return &logo, nil
}

// GetLogosForTeam returns all logos for a team, each with its assignment slice.
func (s *LogoStore) GetLogosForTeam(ctx context.Context, db DBTX, teamID int) ([]TeamLogoWithAssignments, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, team_id, file_path, uploaded_at
		FROM logos
		WHERE team_id = ?
		ORDER BY uploaded_at ASC
	`, teamID)
	if err != nil {
		return nil, fmt.Errorf("listing logos for team %d: %w", teamID, err)
	}
	defer func() { _ = rows.Close() }()

	var results []TeamLogoWithAssignments
	for rows.Next() {
		var logo models.TeamLogo
		var uploadedAt string
		if err := rows.Scan(&logo.ID, &logo.TeamID, &logo.FilePath, &uploadedAt); err != nil {
			return nil, fmt.Errorf("scanning logo: %w", err)
		}
		logo.UploadedAt, _ = time.Parse("2006-01-02T15:04:05Z", uploadedAt)
		results = append(results, TeamLogoWithAssignments{Logo: logo})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating logos: %w", err)
	}

	for i := range results {
		assignments, err := s.getAssignmentsForLogo(ctx, db, results[i].Logo.ID)
		if err != nil {
			return nil, err
		}
		results[i].Assignments = assignments
	}
	return results, nil
}

func (s *LogoStore) getAssignmentsForLogo(ctx context.Context, db DBTX, logoID string) ([]models.TeamLogoAssignment, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, logo_id, start_season, end_season, assigned_at
		FROM logo_assignments
		WHERE logo_id = ?
		ORDER BY assigned_at ASC
	`, logoID)
	if err != nil {
		return nil, fmt.Errorf("listing assignments for logo %s: %w", logoID, err)
	}
	defer func() { _ = rows.Close() }()

	var assignments []models.TeamLogoAssignment
	for rows.Next() {
		var a models.TeamLogoAssignment
		var assignedAt string
		var startSeason, endSeason sql.NullInt64
		if err := rows.Scan(&a.ID, &a.LogoID, &startSeason, &endSeason, &assignedAt); err != nil {
			return nil, fmt.Errorf("scanning logo assignment: %w", err)
		}
		if startSeason.Valid {
			v := int(startSeason.Int64)
			a.StartSeason = &v
		}
		if endSeason.Valid {
			v := int(endSeason.Int64)
			a.EndSeason = &v
		}
		a.AssignedAt, _ = time.Parse("2006-01-02T15:04:05Z", assignedAt)
		assignments = append(assignments, a)
	}
	return assignments, rows.Err()
}

// GetAssignmentCountForLogo returns how many assignments reference the given logo.
func (s *LogoStore) GetAssignmentCountForLogo(ctx context.Context, db DBTX, logoID string) (int, error) {
	var count int
	err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM logo_assignments WHERE logo_id = ?`, logoID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting assignments for logo %s: %w", logoID, err)
	}
	return count, nil
}

// LogoIDForAssignment returns the logo_id for the given assignment.
func (s *LogoStore) LogoIDForAssignment(ctx context.Context, db DBTX, assignmentID string) (string, error) {
	var logoID string
	err := db.QueryRowContext(ctx,
		`SELECT logo_id FROM logo_assignments WHERE id = ?`, assignmentID,
	).Scan(&logoID)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("assignment %s not found", assignmentID)
	}
	if err != nil {
		return "", fmt.Errorf("looking up logo for assignment %s: %w", assignmentID, err)
	}
	return logoID, nil
}

// GetLogoByID returns the logo with the given id.
func (s *LogoStore) GetLogoByID(ctx context.Context, db DBTX, logoID string) (models.TeamLogo, error) {
	var logo models.TeamLogo
	var uploadedAt string
	err := db.QueryRowContext(ctx,
		`SELECT id, team_id, file_path, uploaded_at FROM logos WHERE id = ?`, logoID,
	).Scan(&logo.ID, &logo.TeamID, &logo.FilePath, &uploadedAt)
	if err == sql.ErrNoRows {
		return models.TeamLogo{}, fmt.Errorf("logo %s not found", logoID)
	}
	if err != nil {
		return models.TeamLogo{}, fmt.Errorf("getting logo %s: %w", logoID, err)
	}
	logo.UploadedAt, _ = time.Parse("2006-01-02T15:04:05Z", uploadedAt)
	return logo, nil
}

// DeleteAssignment removes a single logo assignment record.
func (s *LogoStore) DeleteAssignment(ctx context.Context, db DBTX, assignmentID string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM logo_assignments WHERE id = ?`, assignmentID)
	if err != nil {
		return fmt.Errorf("deleting logo assignment %s: %w", assignmentID, err)
	}
	return nil
}

// DeleteLogoAndAllAssignments removes the logo record and all its assignment records.
func (s *LogoStore) DeleteLogoAndAllAssignments(ctx context.Context, db DBTX, logoID string) error {
	if _, err := db.ExecContext(ctx, `DELETE FROM logo_assignments WHERE logo_id = ?`, logoID); err != nil {
		return fmt.Errorf("deleting assignments for logo %s: %w", logoID, err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM logos WHERE id = ?`, logoID); err != nil {
		return fmt.Errorf("deleting logo %s: %w", logoID, err)
	}
	return nil
}
