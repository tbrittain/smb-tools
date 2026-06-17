package db_test

import (
	"bytes"
	"compress/zlib"
	"os"
	"path/filepath"
	"testing"

	internaldb "smb-tools/internal/db"
)

func writeZlibFile(t *testing.T, path string, content []byte) {
	t.Helper()
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	if _, err := zw.Write(content); err != nil {
		t.Fatalf("compressing test content: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("closing zlib writer: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("writing test zlib file: %v", err)
	}
}

func TestDecompressToTempFile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	savPath := filepath.Join(dir, "test.sav")
	content := []byte("pretend this is a sqlite file's bytes")
	writeZlibFile(t, savPath, content)

	tmpPath, err := internaldb.DecompressToTempFile(savPath)
	if err != nil {
		t.Fatalf("DecompressToTempFile: %v", err)
	}
	defer func() { _ = os.Remove(tmpPath) }()

	got, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("reading decompressed temp file: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("decompressed content = %q, want %q", got, content)
	}

	// Original file must be untouched.
	origBytes, err := os.ReadFile(savPath)
	if err != nil {
		t.Fatalf("reading original file: %v", err)
	}
	var verify bytes.Buffer
	zr, err := zlib.NewReader(bytes.NewReader(origBytes))
	if err != nil {
		t.Fatalf("original file is no longer valid zlib: %v", err)
	}
	if _, err := verify.ReadFrom(zr); err != nil {
		t.Fatalf("decompressing original file post-call: %v", err)
	}
	if !bytes.Equal(verify.Bytes(), content) {
		t.Errorf("original file was modified by DecompressToTempFile")
	}
}

func TestDecompressToTempFile_NotZlib(t *testing.T) {
	dir := t.TempDir()
	savPath := filepath.Join(dir, "not-zlib.sav")
	if err := os.WriteFile(savPath, []byte("not zlib data at all"), 0o600); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	if _, err := internaldb.DecompressToTempFile(savPath); err == nil {
		t.Error("expected an error for non-zlib content, got nil")
	}
}

func TestCompressFileAtomically_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	tmpPath := filepath.Join(dir, "decompressed.sqlite")
	content := []byte("mutated sqlite bytes go here")
	if err := os.WriteFile(tmpPath, content, 0o600); err != nil {
		t.Fatalf("writing test temp file: %v", err)
	}

	destPath := filepath.Join(dir, "master.sav")
	// Pre-populate destPath to confirm it gets fully replaced, not appended to.
	if err := os.WriteFile(destPath, []byte("stale content"), 0o600); err != nil {
		t.Fatalf("writing pre-existing dest file: %v", err)
	}

	if err := internaldb.CompressFileAtomically(tmpPath, destPath); err != nil {
		t.Fatalf("CompressFileAtomically: %v", err)
	}

	// No leftover swap file.
	if _, err := os.Stat(destPath + ".smb-tools-tmp"); !os.IsNotExist(err) {
		t.Errorf("expected swap file to be gone, stat err = %v", err)
	}

	compressed, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("reading dest file: %v", err)
	}
	zr, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("dest file is not valid zlib: %v", err)
	}
	var got bytes.Buffer
	if _, err := got.ReadFrom(zr); err != nil {
		t.Fatalf("decompressing dest file: %v", err)
	}
	if !bytes.Equal(got.Bytes(), content) {
		t.Errorf("round-tripped content = %q, want %q", got.Bytes(), content)
	}
}

func TestCompressThenDecompress_FullRoundTrip(t *testing.T) {
	dir := t.TempDir()
	original := []byte("full round trip content, simulating a mutated master.sav")

	tmpPath := filepath.Join(dir, "mutated.sqlite")
	if err := os.WriteFile(tmpPath, original, 0o600); err != nil {
		t.Fatalf("writing mutated temp file: %v", err)
	}

	savPath := filepath.Join(dir, "master.sav")
	if err := internaldb.CompressFileAtomically(tmpPath, savPath); err != nil {
		t.Fatalf("CompressFileAtomically: %v", err)
	}

	roundTrippedTmp, err := internaldb.DecompressToTempFile(savPath)
	if err != nil {
		t.Fatalf("DecompressToTempFile: %v", err)
	}
	defer func() { _ = os.Remove(roundTrippedTmp) }()

	got, err := os.ReadFile(roundTrippedTmp)
	if err != nil {
		t.Fatalf("reading round-tripped file: %v", err)
	}
	if !bytes.Equal(got, original) {
		t.Errorf("full round trip = %q, want %q", got, original)
	}
}

func TestBackupFileTimestamped(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "master.sav")
	content := []byte("the master save contents")
	if err := os.WriteFile(srcPath, content, 0o600); err != nil {
		t.Fatalf("writing source file: %v", err)
	}

	backupDir := filepath.Join(dir, "backups")

	backup1, err := internaldb.BackupFileTimestamped(srcPath, backupDir, "master", "20260617-150000")
	if err != nil {
		t.Fatalf("first BackupFileTimestamped: %v", err)
	}
	backup2, err := internaldb.BackupFileTimestamped(srcPath, backupDir, "master", "20260617-150405")
	if err != nil {
		t.Fatalf("second BackupFileTimestamped: %v", err)
	}

	if backup1 == backup2 {
		t.Fatalf("expected distinct backup paths for distinct timestamps, got the same path twice: %s", backup1)
	}

	for _, p := range []string{backup1, backup2} {
		got, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("reading backup %s: %v", p, err)
		}
		if !bytes.Equal(got, content) {
			t.Errorf("backup %s content = %q, want %q", p, got, content)
		}
	}
}
