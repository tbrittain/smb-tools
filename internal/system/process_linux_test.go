//go:build linux

package system

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFakeProc creates a synthetic /proc-shaped directory: one subdirectory
// per pid, each with a cmdline file containing cmdline (NUL-joined args, as
// the real /proc/*/cmdline format does — though a plain string is enough
// for the substring check this code performs).
func writeFakeProc(t *testing.T, entries map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for pid, cmdline := range entries {
		dir := filepath.Join(root, pid)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("creating fake proc dir for pid %s: %v", pid, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "cmdline"), []byte(cmdline), 0o644); err != nil {
			t.Fatalf("writing fake cmdline for pid %s: %v", pid, err)
		}
	}
	return root
}

func TestIsGameRunningUnderRoot_NotRunning(t *testing.T) {
	root := writeFakeProc(t, map[string]string{
		"100": "/usr/bin/some-other-app\x00--flag\x00",
		"200": "steam\x00",
	})

	running, err := isGameRunningUnderRoot(root)
	if err != nil {
		t.Fatalf("isGameRunningUnderRoot: %v", err)
	}
	if running {
		t.Error("expected running = false, got true")
	}
}

func TestIsGameRunningUnderRoot_RunningViaProton(t *testing.T) {
	root := writeFakeProc(t, map[string]string{
		"100": "/usr/bin/some-other-app\x00",
		// Simulates how a Proton-launched process looks: the full Windows
		// exe path and name appear in cmdline, unlike the 15-char-truncated
		// /proc/*/comm.
		"300": "Z:\\home\\user\\.steam\\steamapps\\common\\Super Mega Baseball 4\\supermegabaseball.exe\x00",
	})

	running, err := isGameRunningUnderRoot(root)
	if err != nil {
		t.Fatalf("isGameRunningUnderRoot: %v", err)
	}
	if !running {
		t.Error("expected running = true, got false")
	}
}

func TestIsGameRunningUnderRoot_IgnoresNonPidEntries(t *testing.T) {
	root := t.TempDir()
	// Non-numeric entries (self, net, etc. in a real /proc) must be skipped,
	// not treated as PIDs.
	if err := os.MkdirAll(filepath.Join(root, "self"), 0o755); err != nil {
		t.Fatalf("creating non-pid dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "version"), []byte("fake"), 0o644); err != nil {
		t.Fatalf("creating non-pid file: %v", err)
	}

	running, err := isGameRunningUnderRoot(root)
	if err != nil {
		t.Fatalf("isGameRunningUnderRoot: %v", err)
	}
	if running {
		t.Error("expected running = false for a /proc with no real pid entries")
	}
}

func TestIsGameRunningUnderRoot_TolerantOfVanishedProcess(t *testing.T) {
	// A pid directory with no cmdline file (process exited between ReadDir
	// and the cmdline read) must not be treated as an error.
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "999"), 0o755); err != nil {
		t.Fatalf("creating pid dir: %v", err)
	}

	running, err := isGameRunningUnderRoot(root)
	if err != nil {
		t.Fatalf("isGameRunningUnderRoot should tolerate a missing cmdline file: %v", err)
	}
	if running {
		t.Error("expected running = false")
	}
}
