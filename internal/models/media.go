package models

import "time"

// MediaType distinguishes image files from video files.
type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
)

// Media is a single uploaded file (image or video) within a franchise.
// FilePath is relative to the franchise directory (e.g. "assets/media/uuid.mp4").
type Media struct {
	ID          string
	FilePath    string
	MediaType   MediaType
	Name        string
	Description string
	UploadedAt  time.Time
}

// MediaTeamSeason links a media item to a specific team-season record.
type MediaTeamSeason struct {
	ID            string
	MediaID       string
	TeamHistoryID int64
}

// MediaPlayer links a media item to a player.
type MediaPlayer struct {
	ID       string
	MediaID  string
	PlayerID int64
}
