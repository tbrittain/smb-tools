package zip_test

import (
	"compress/zlib"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	internalzip "smb-tools/internal/zip"
)

func writeZlibFile(t *testing.T, path string, content []byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("creating %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()

	zw := zlib.NewWriter(f)
	if _, err := zw.Write(content); err != nil {
		t.Fatalf("compressing %s: %v", path, err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("closing zlib writer for %s: %v", path, err)
	}
}

func samplePackInput(t *testing.T, guid uuid.UUID) (dir string, in internalzip.PackInput) {
	t.Helper()
	dir = t.TempDir()
	upper := guid.String()

	savPath := filepath.Join(dir, "league-"+upper+".sav")
	bakPath := filepath.Join(dir, "league-"+upper+".sav.bak")
	hashPath := filepath.Join(dir, "league-"+upper+".hash")

	writeZlibFile(t, savPath, []byte("pretend sqlite bytes"))
	writeZlibFile(t, bakPath, []byte("pretend sqlite backup bytes"))
	if err := os.WriteFile(hashPath, []byte{0xc0, 0x20, 0xc1, 0x0c}, 0o600); err != nil {
		t.Fatalf("writing hash file: %v", err)
	}

	return dir, internalzip.PackInput{
		GUID:            guid,
		LeagueName:      "Super Mega League",
		SavPath:         savPath,
		BakPath:         bakPath,
		HashPath:        hashPath,
		ExportedAt:      "2026-06-17T12:00:00Z",
		SmbToolsVersion: "test",
	}
}

func TestPackUnpack_RoundTrip(t *testing.T) {
	guid := uuid.New()
	_, in := samplePackInput(t, guid)

	zipPath := filepath.Join(t.TempDir(), "export.zip")
	if err := internalzip.Pack(zipPath, in); err != nil {
		t.Fatalf("Pack: %v", err)
	}

	destDir := t.TempDir()
	result, err := internalzip.Unpack(zipPath, destDir)
	if err != nil {
		t.Fatalf("Unpack: %v", err)
	}

	if result.Manifest.LeagueGUID != guid {
		t.Errorf("manifest GUID = %s, want %s", result.Manifest.LeagueGUID, guid)
	}
	if result.Manifest.LeagueName != "Super Mega League" {
		t.Errorf("manifest LeagueName = %q, want %q", result.Manifest.LeagueName, "Super Mega League")
	}
	if result.Manifest.ExportedAt != "2026-06-17T12:00:00Z" {
		t.Errorf("manifest ExportedAt = %q, want %q", result.Manifest.ExportedAt, "2026-06-17T12:00:00Z")
	}
	if result.Manifest.SmbToolsVersion != "test" {
		t.Errorf("manifest SmbToolsVersion = %q, want %q", result.Manifest.SmbToolsVersion, "test")
	}

	for _, p := range []string{result.SavPath, result.BakPath, result.HashPath} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected extracted file %s to exist: %v", p, err)
		}
	}
}

func TestPackUnpack_NoHashFile(t *testing.T) {
	guid := uuid.New()
	_, in := samplePackInput(t, guid)
	in.HashPath = "" // simulate a source league with no .hash file

	zipPath := filepath.Join(t.TempDir(), "export.zip")
	if err := internalzip.Pack(zipPath, in); err != nil {
		t.Fatalf("Pack: %v", err)
	}

	result, err := internalzip.Unpack(zipPath, t.TempDir())
	if err != nil {
		t.Fatalf("Unpack: %v", err)
	}
	if result.HashPath != "" {
		t.Errorf("expected no HashPath, got %q", result.HashPath)
	}
	if result.SavPath == "" || result.BakPath == "" {
		t.Error("expected .sav and .sav.bak to still be present")
	}
}

func TestUnpack_RejectsMissingManifest(t *testing.T) {
	dir := t.TempDir()
	guid := uuid.New()
	savPath := filepath.Join(dir, "league-"+guid.String()+".sav")
	writeZlibFile(t, savPath, []byte("data"))

	zipPath := filepath.Join(t.TempDir(), "no-manifest.zip")
	zw := newTestZip(t, zipPath)
	addFileToZip(t, zw, savPath, filepath.Base(savPath))
	closeTestZip(t, zw)

	if _, err := internalzip.Unpack(zipPath, t.TempDir()); err == nil {
		t.Error("expected an error for a package with no manifest.json, got nil")
	}
}

func TestUnpack_RejectsGUIDMismatch(t *testing.T) {
	guid := uuid.New()
	dir, in := samplePackInput(t, guid)

	// Tamper: rename the .sav.bak on disk to embed a different GUID before
	// packing, simulating a corrupted or tampered-with package.
	otherGUID := uuid.New()
	tamperedBak := filepath.Join(dir, "league-"+otherGUID.String()+".sav.bak")
	if err := os.Rename(in.BakPath, tamperedBak); err != nil {
		t.Fatalf("renaming bak file: %v", err)
	}
	in.BakPath = tamperedBak

	zipPath := filepath.Join(t.TempDir(), "mismatch.zip")
	if err := internalzip.Pack(zipPath, in); err != nil {
		t.Fatalf("Pack: %v", err)
	}

	if _, err := internalzip.Unpack(zipPath, t.TempDir()); err == nil {
		t.Error("expected an error for a GUID mismatch between manifest and file name, got nil")
	}
}

func TestUnpack_RejectsMissingSavFile(t *testing.T) {
	guid := uuid.New()
	dir, in := samplePackInput(t, guid)
	in.HashPath = ""

	zipPath := filepath.Join(t.TempDir(), "missing-sav.zip")
	// Pack normally, then manually rebuild a zip with the .sav entry stripped.
	if err := internalzip.Pack(zipPath, in); err != nil {
		t.Fatalf("Pack: %v", err)
	}

	strippedPath := filepath.Join(dir, "stripped.zip")
	rebuildZipWithoutSuffix(t, zipPath, strippedPath, ".sav")

	if _, err := internalzip.Unpack(strippedPath, t.TempDir()); err == nil {
		t.Error("expected an error for a package missing its .sav file, got nil")
	}
}

func TestUnpack_RejectsCorruptZlibStream(t *testing.T) {
	dir := t.TempDir()
	guid := uuid.New()
	upper := guid.String()

	savPath := filepath.Join(dir, "league-"+upper+".sav")
	bakPath := filepath.Join(dir, "league-"+upper+".sav.bak")
	// Not zlib data at all.
	if err := os.WriteFile(savPath, []byte("not zlib data"), 0o600); err != nil {
		t.Fatalf("writing fake sav: %v", err)
	}
	if err := os.WriteFile(bakPath, []byte("not zlib data either"), 0o600); err != nil {
		t.Fatalf("writing fake bak: %v", err)
	}

	in := internalzip.PackInput{
		GUID:            guid,
		LeagueName:      "Corrupt League",
		SavPath:         savPath,
		BakPath:         bakPath,
		ExportedAt:      "2026-06-17T12:00:00Z",
		SmbToolsVersion: "test",
	}
	zipPath := filepath.Join(t.TempDir(), "corrupt.zip")
	if err := internalzip.Pack(zipPath, in); err != nil {
		t.Fatalf("Pack: %v", err)
	}

	if _, err := internalzip.Unpack(zipPath, t.TempDir()); err == nil {
		t.Error("expected an error for a non-zlib .sav file, got nil")
	}
}

func TestUnpack_RejectsMalformedManifest(t *testing.T) {
	dir := t.TempDir()
	guid := uuid.New()
	savPath := filepath.Join(dir, "league-"+guid.String()+".sav")
	bakPath := filepath.Join(dir, "league-"+guid.String()+".sav.bak")
	writeZlibFile(t, savPath, []byte("data"))
	writeZlibFile(t, bakPath, []byte("data"))

	zipPath := filepath.Join(t.TempDir(), "bad-manifest.zip")
	zw := newTestZip(t, zipPath)
	addBytesToZip(t, zw, "manifest.json", []byte("{not valid json"))
	addFileToZip(t, zw, savPath, filepath.Base(savPath))
	addFileToZip(t, zw, bakPath, filepath.Base(bakPath))
	closeTestZip(t, zw)

	if _, err := internalzip.Unpack(zipPath, t.TempDir()); err == nil {
		t.Error("expected an error for a malformed manifest.json, got nil")
	}
}
