package db

import (
	"compress/zlib"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
)

// OpenSnapshot opens an already-decompressed snapshot SQLite file as a
// read-only connection. Unlike DecompressAndOpen, no temp file is created —
// the snapshot is opened in-place. The caller must call db.Close() when done.
func OpenSnapshot(ctx context.Context, snapshotPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", "file:"+snapshotPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("opening snapshot DB: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pinging snapshot DB: %w", err)
	}
	return db, nil
}

// DecompressAndOpen decompresses a zlib-compressed SMB save game file to a
// temporary file and opens it as a read-only SQLite connection.
//
// Returns the opened DB and the path to the temporary file. The caller must:
//  1. Call db.Close() when done reading
//  2. Call os.Remove(tmpPath) to clean up the temp file
//
// The original .sav file is never modified.
func DecompressAndOpen(ctx context.Context, savePath string) (db *sql.DB, tmpPath string, err error) {
	f, err := os.Open(savePath)
	if err != nil {
		return nil, "", fmt.Errorf("opening save file: %w", err)
	}
	defer func() { _ = f.Close() }()

	zr, err := zlib.NewReader(f)
	if err != nil {
		return nil, "", fmt.Errorf("creating zlib reader: %w", err)
	}
	defer func() { _ = zr.Close() }()

	tmp, err := os.CreateTemp("", "smb-tools-savegame-*.sqlite")
	if err != nil {
		return nil, "", fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath = tmp.Name()

	if _, err := io.Copy(tmp, zr); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return nil, "", fmt.Errorf("decompressing save file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return nil, "", fmt.Errorf("flushing decompressed save file: %w", err)
	}

	// Open strictly read-only; the original save is never written to.
	db, err = sql.Open("sqlite", "file:"+tmpPath+"?mode=ro")
	if err != nil {
		_ = os.Remove(tmpPath)
		return nil, "", fmt.Errorf("opening decompressed DB: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		_ = os.Remove(tmpPath)
		return nil, "", fmt.Errorf("pinging save game DB: %w", err)
	}
	return db, tmpPath, nil
}
