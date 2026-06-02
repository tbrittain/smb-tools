package models

import "time"

// TeamLogo represents a logo image uploaded for a team.
// FilePath is relative to the franchise directory (e.g. "assets/logos/42/uuid.png").
type TeamLogo struct {
	ID         string
	TeamID     int
	FilePath   string
	UploadedAt time.Time
}

// TeamLogoAssignment associates a logo with a season range for a team.
// Nil StartSeason means "from the very beginning"; nil EndSeason means "no end / ongoing".
type TeamLogoAssignment struct {
	ID          string
	LogoID      string
	StartSeason *int
	EndSeason   *int
	AssignedAt  time.Time
}
