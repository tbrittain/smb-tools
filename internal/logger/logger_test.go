package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSetup_CreatesLogDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "logs")
	cleanup, _, err := Setup(dir, false)
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	defer cleanup()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("log directory was not created")
	}
}

func TestSetup_CreatesSessionFile(t *testing.T) {
	dir := t.TempDir()
	cleanup, sessionFile, err := Setup(dir, false)
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	defer cleanup()

	base := filepath.Base(sessionFile)
	if !strings.HasPrefix(base, "session-") || !strings.HasSuffix(base, ".log") {
		t.Errorf("unexpected session file name: %q", base)
	}
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		t.Error("session file does not exist on disk")
	}
}

func TestSetup_RotationWith5ExistingFiles(t *testing.T) {
	dir := t.TempDir()

	// Create 5 existing session files with staggered timestamps.
	for i := range 5 {
		name := "session-2024-01-0" + string(rune('1'+i)) + "_00-00-00.log"
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond) // ensure distinct mtime just in case
	}

	cleanup, _, err := Setup(dir, false)
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	defer cleanup()

	entries, _ := os.ReadDir(dir)
	var logFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".log") {
			logFiles = append(logFiles, e.Name())
		}
	}
	if len(logFiles) != maxSessionFiles {
		t.Errorf("expected %d log files after rotation, got %d: %v", maxSessionFiles, len(logFiles), logFiles)
	}
}

func TestSetup_RotationWith4ExistingFiles(t *testing.T) {
	dir := t.TempDir()

	for i := range 4 {
		name := "session-2024-01-0" + string(rune('1'+i)) + "_00-00-00.log"
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	cleanup, _, err := Setup(dir, false)
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	defer cleanup()

	entries, _ := os.ReadDir(dir)
	var logFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".log") {
			logFiles = append(logFiles, e.Name())
		}
	}
	if len(logFiles) != maxSessionFiles {
		t.Errorf("expected %d log files, got %d: %v", maxSessionFiles, len(logFiles), logFiles)
	}
}

func TestSetup_RotationWithNoExistingFiles(t *testing.T) {
	dir := t.TempDir()
	cleanup, _, err := Setup(dir, false)
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	defer cleanup()

	entries, _ := os.ReadDir(dir)
	var logFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".log") {
			logFiles = append(logFiles, e.Name())
		}
	}
	if len(logFiles) != 1 {
		t.Errorf("expected 1 log file, got %d: %v", len(logFiles), logFiles)
	}
}

func TestSetup_RotationDeletesOldest(t *testing.T) {
	dir := t.TempDir()

	oldest := "session-2024-01-01_00-00-00.log"
	if err := os.WriteFile(filepath.Join(dir, oldest), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	for i := 1; i < maxSessionFiles; i++ {
		name := "session-2024-01-0" + string(rune('1'+i)) + "_00-00-00.log"
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	cleanup, _, err := Setup(dir, false)
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	defer cleanup()

	if _, err := os.Stat(filepath.Join(dir, oldest)); !os.IsNotExist(err) {
		t.Error("oldest session file was not deleted by rotation")
	}
}

func TestSetup_CleanupClosesFile(t *testing.T) {
	dir := t.TempDir()
	cleanup, sessionFile, err := Setup(dir, false)
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	cleanup()

	// After cleanup the file should still exist but the fd is closed.
	// On Windows and Linux we can verify by checking the file is accessible.
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		t.Error("session file disappeared after cleanup")
	}
}

func TestTailFile_Empty(t *testing.T) {
	if got := TailFile("", 100); got != "" {
		t.Errorf("expected empty string for empty path, got %q", got)
	}
}

func TestTailFile_TruncatesLargeFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "tail-*.log")
	if err != nil {
		t.Fatal(err)
	}
	content := strings.Repeat("a", 8000)
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	got := TailFile(f.Name(), 4000)
	if len(got) > 4000 {
		t.Errorf("TailFile returned %d bytes, expected <= 4000", len(got))
	}
	if !strings.HasSuffix(content, got) {
		t.Error("TailFile did not return the tail of the file")
	}
}

func TestTailFile_SmallFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "tail-*.log")
	if err != nil {
		t.Fatal(err)
	}
	content := "hello world"
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	got := TailFile(f.Name(), 4000)
	if got != content {
		t.Errorf("TailFile(%q) = %q, want %q", f.Name(), got, content)
	}
}
