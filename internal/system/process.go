// Package system contains OS-specific checks that don't fit any other
// internal package — currently just detecting whether SMB4 is running, the
// safety gate League Transfer requires before mutating master.sav (see
// docs/league-transfer/ux-flow.md and implementation-plan.md's discussion of
// why a process check, not just a file-lock probe, is the primary check).
package system

// GameRunningChecker reports whether the SMB4 game process is currently
// running. Callers (internal/service/league_transfer.go) depend on this
// interface rather than calling a package-level function directly, so tests
// can inject a fake without needing real OS process APIs.
type GameRunningChecker interface {
	IsGameRunning() (bool, error)
}

// DefaultGameRunningChecker is the platform-appropriate GameRunningChecker.
// Its behavior is implemented per-platform in process_windows.go,
// process_linux.go, and process_other.go.
type DefaultGameRunningChecker struct{}

func (DefaultGameRunningChecker) IsGameRunning() (bool, error) {
	return isGameRunning()
}
