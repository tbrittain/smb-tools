package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const maxSessionFiles = 5

// Setup initialises the global slog logger. It creates logDir if absent, rotates
// old session files (keeping the most recent maxSessionFiles-1 so there is room
// for the new one), and opens a timestamped session log file. When devMode is
// true, output is also written to stderr. Returns a cleanup func (closes the log
// file) and the path of the newly created session file.
func Setup(logDir string, devMode bool) (cleanup func(), sessionFile string, err error) {
	if err = os.MkdirAll(logDir, 0o700); err != nil {
		return nil, "", fmt.Errorf("creating log directory: %w", err)
	}
	if err = rotate(logDir); err != nil {
		return nil, "", fmt.Errorf("rotating log files: %w", err)
	}

	name := "session-" + time.Now().Format("2006-01-02_15-04-05") + ".log"
	path := filepath.Join(logDir, name)
	f, err := os.Create(path)
	if err != nil {
		return nil, "", fmt.Errorf("creating session log file: %w", err)
	}

	var w io.Writer = f
	if devMode {
		w = io.MultiWriter(f, os.Stderr)
	}

	h := slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(h))

	return func() { _ = f.Close() }, path, nil
}

// TailFile returns up to maxBytes of the tail of the file at path, reading
// backwards from the end. Returns an empty string if the file cannot be read.
func TailFile(path string, maxBytes int) string {
	if path == "" {
		return ""
	}
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil || info.Size() == 0 {
		return ""
	}

	size := info.Size()
	offset := max(int64(0), size-int64(maxBytes))
	buf := make([]byte, size-offset)
	n, err := f.ReadAt(buf, offset)
	if err != nil && err != io.EOF {
		return ""
	}
	return string(buf[:n])
}

// rotate deletes the oldest .log files in logDir so that at most
// maxSessionFiles-1 remain — leaving room for the new session file.
func rotate(logDir string) error {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".log" {
			files = append(files, filepath.Join(logDir, e.Name()))
		}
	}
	// Lexicographic order equals chronological order because of the timestamp prefix.
	sort.Strings(files)
	for len(files) >= maxSessionFiles {
		if err := os.Remove(files[0]); err != nil {
			return err
		}
		files = files[1:]
	}
	return nil
}
