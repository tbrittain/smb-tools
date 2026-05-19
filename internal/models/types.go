package models

// GameVersion identifies which Super Mega Baseball title a save file or
// franchise belongs to.
type GameVersion string

const (
	GameVersionSMB3 GameVersion = "smb3"
	GameVersionSMB4 GameVersion = "smb4"
)

// Valid reports whether the GameVersion is one of the known values.
func (v GameVersion) Valid() bool {
	return v == GameVersionSMB3 || v == GameVersionSMB4
}

func (v GameVersion) String() string { return string(v) }
