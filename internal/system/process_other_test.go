//go:build !windows && !linux

package system_test

import (
	"testing"

	"smb-tools/internal/system"
)

func TestDefaultGameRunningChecker_AlwaysFalseOnOtherPlatforms(t *testing.T) {
	checker := system.DefaultGameRunningChecker{}
	running, err := checker.IsGameRunning()
	if err != nil {
		t.Errorf("IsGameRunning: %v", err)
	}
	if running {
		t.Error("expected running = false on a platform with no confirmed SMB4 support")
	}
}
