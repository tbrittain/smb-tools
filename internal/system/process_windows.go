//go:build windows

package system

import (
	"fmt"
	"os/exec"
	"strings"
)

// gameProcessName is SMB4's executable name, confirmed live via
// Get-Process during the league-transfer research work.
const gameProcessName = "supermegabaseball.exe"

func isGameRunning() (bool, error) {
	out, err := exec.Command("tasklist", "/FI", "IMAGENAME eq "+gameProcessName, "/NH").Output()
	if err != nil {
		return false, fmt.Errorf("checking running processes: %w", err)
	}
	return strings.Contains(strings.ToLower(string(out)), gameProcessName), nil
}
