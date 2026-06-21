//go:build windows

package system_test

import (
	"testing"

	"smb-tools/internal/system"
)

func TestDefaultGameRunningChecker_DoesNotErrorOnRealMachine(t *testing.T) {
	// We can't assert a specific true/false result — SMB4 may or may not be
	// running on the machine actually executing this test — but the check
	// itself (shelling out to tasklist) must succeed without error.
	checker := system.DefaultGameRunningChecker{}
	if _, err := checker.IsGameRunning(); err != nil {
		t.Errorf("IsGameRunning: %v", err)
	}
}
