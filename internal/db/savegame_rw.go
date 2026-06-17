package db

import (
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// DecompressToTempFile decompresses a zlib-compressed SMB save file (a
// per-league .sav or master.sav — both use the same format) to a new
// temporary file and returns its path. Unlike DecompressAndOpen, no SQLite
// connection is opened — callers that need read-write access (League
// Transfer) open the returned path themselves. The original file is never
// modified. The caller is responsible for removing the returned temp file.
func DecompressToTempFile(srcPath string) (tmpPath string, err error) {
	f, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("opening save file: %w", err)
	}
	defer func() { _ = f.Close() }()

	zr, err := zlib.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("creating zlib reader: %w", err)
	}
	defer func() { _ = zr.Close() }()

	tmp, err := os.CreateTemp("", "smb-tools-savegame-rw-*.sqlite")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath = tmp.Name()

	if _, err := io.Copy(tmp, zr); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("decompressing save file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("flushing decompressed save file: %w", err)
	}

	return tmpPath, nil
}

// CompressFileAtomically reads the file at tmpPath, zlib-compresses it, and
// writes the result to destPath. The write goes through a sibling temp file
// followed by os.Rename so destPath is never left truncated or partially
// written if the process is interrupted mid-write — required for any code
// that mutates a live, in-use file like master.sav (see
// docs/league-transfer/plan.md's safety requirements).
func CompressFileAtomically(tmpPath, destPath string) error {
	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("reading decompressed file: %w", err)
	}

	swapPath := destPath + ".smb-tools-tmp"
	swapFile, err := os.OpenFile(swapPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("creating swap file: %w", err)
	}

	zw := zlib.NewWriter(swapFile)
	if _, err := zw.Write(data); err != nil {
		_ = zw.Close()
		_ = swapFile.Close()
		_ = os.Remove(swapPath)
		return fmt.Errorf("compressing save file: %w", err)
	}
	if err := zw.Close(); err != nil {
		_ = swapFile.Close()
		_ = os.Remove(swapPath)
		return fmt.Errorf("flushing zlib writer: %w", err)
	}
	if err := swapFile.Close(); err != nil {
		_ = os.Remove(swapPath)
		return fmt.Errorf("closing swap file: %w", err)
	}

	if err := os.Rename(swapPath, destPath); err != nil {
		_ = os.Remove(swapPath)
		return fmt.Errorf("swapping in compressed save file: %w", err)
	}

	return nil
}

// BackupFileTimestamped copies srcPath into destDir under a name that embeds
// the current time, so successive calls never overwrite a previous backup.
// Mirrors the timestamped-history approach already used for franchise
// snapshots (internal/service/snapshot.go), applied here to master.sav.
func BackupFileTimestamped(srcPath, destDir, prefix, timestamp string) (backupPath string, err error) {
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return "", fmt.Errorf("creating backup directory: %w", err)
	}

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return "", fmt.Errorf("reading file to back up: %w", err)
	}

	backupPath = filepath.Join(destDir, fmt.Sprintf("%s_%s%s", prefix, timestamp, filepath.Ext(srcPath)))
	if err := os.WriteFile(backupPath, data, 0o600); err != nil {
		return "", fmt.Errorf("writing backup file: %w", err)
	}

	return backupPath, nil
}
