//go:build !windows && !linux

package system

// isGameRunning always reports false on platforms other than Windows and
// Linux. SMB4 has no confirmed macOS support (see
// internal/config/savegame_paths.go), so there's no process to detect there
// — this is an intentional no-op, not an unimplemented stub.
func isGameRunning() (bool, error) {
	return false, nil
}
