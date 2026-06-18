//go:build linux

package system

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const procRoot = "/proc"

// gameProcessNeedle is checked against each process's full cmdline, not
// comm — Proton runs the unmodified Windows binary under Wine, and
// /proc/*/comm truncates to 15 characters (cutting "supermegabaseball.exe"
// short), while /proc/*/cmdline is not truncated.
const gameProcessNeedle = "supermegabaseball.exe"

func isGameRunning() (bool, error) {
	return isGameRunningUnderRoot(procRoot)
}

// isGameRunningUnderRoot scans root (normally /proc) for a process whose
// cmdline contains gameProcessNeedle. Split out from isGameRunning so tests
// can point it at a synthetic directory standing in for /proc.
func isGameRunningUnderRoot(root string) (bool, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return false, fmt.Errorf("reading %s: %w", root, err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := strconv.Atoi(e.Name()); err != nil {
			continue // not a PID directory
		}

		data, err := os.ReadFile(filepath.Join(root, e.Name(), "cmdline"))
		if err != nil {
			// Process may have exited between ReadDir and now, or this PID's
			// cmdline may not be readable — neither is fatal to the overall check.
			continue
		}
		if strings.Contains(strings.ToLower(string(data)), gameProcessNeedle) {
			return true, nil
		}
	}
	return false, nil
}
