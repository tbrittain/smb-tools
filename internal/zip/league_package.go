// Package zip provides the export/import container format for League
// Transfer: a zip file holding a league's .sav/.sav.bak/.hash plus a JSON
// manifest, per docs/league-transfer/ux-flow.md.
package zip

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const manifestFileName = "manifest.json"

// Manifest is the JSON file packed alongside the save files. It exists so
// import-time validation and a human opening the zip both have something to
// check against beyond filename parsing alone.
type Manifest struct {
	LeagueGUID      uuid.UUID `json:"leagueGuid"`
	LeagueName      string    `json:"leagueName"`
	ExportedAt      string    `json:"exportedAt"` // RFC 3339
	SmbToolsVersion string    `json:"smbToolsVersion"`
}

// PackInput names the on-disk files to package. HashPath may be empty —
// not every league save has a .hash file (see
// docs/league-transfer/legacy-tool-analysis.md).
type PackInput struct {
	GUID            uuid.UUID
	LeagueName      string
	SavPath         string
	BakPath         string
	HashPath        string // optional
	ExportedAt      string // RFC 3339
	SmbToolsVersion string
}

// Pack writes a league export zip to outPath containing the manifest and
// the league's save files, named the same way they appear in a real Steam
// save directory so Unpack's output can be copied there directly.
func Pack(outPath string, in PackInput) error {
	manifest := Manifest{
		LeagueGUID:      in.GUID,
		LeagueName:      in.LeagueName,
		ExportedAt:      in.ExportedAt,
		SmbToolsVersion: in.SmbToolsVersion,
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling manifest: %w", err)
	}

	out, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("creating output zip: %w", err)
	}
	defer func() { _ = out.Close() }()

	zw := zip.NewWriter(out)
	defer func() { _ = zw.Close() }()

	if err := writeZipEntry(zw, manifestFileName, manifestBytes); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}

	files := []struct {
		path string
		name string
	}{
		{in.SavPath, filepath.Base(in.SavPath)},
		{in.BakPath, filepath.Base(in.BakPath)},
	}
	if in.HashPath != "" {
		files = append(files, struct {
			path string
			name string
		}{in.HashPath, filepath.Base(in.HashPath)})
	}

	for _, f := range files {
		data, err := os.ReadFile(f.path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", f.path, err)
		}
		if err := writeZipEntry(zw, f.name, data); err != nil {
			return fmt.Errorf("writing %s: %w", f.name, err)
		}
	}

	if err := zw.Close(); err != nil {
		return fmt.Errorf("finalizing zip: %w", err)
	}
	return nil
}

func writeZipEntry(zw *zip.Writer, name string, data []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// UnpackResult holds the manifest and the paths of the extracted save
// files, ready to be validated further (zlib/shape) or copied into a real
// Steam save directory.
type UnpackResult struct {
	Manifest Manifest
	SavPath  string
	BakPath  string
	HashPath string // empty if the package had no .hash file
}

// Unpack extracts a league export zip to destDir and validates its shape:
// a well-formed manifest, a .sav and .sav.bak file whose names embed the
// same GUID as the manifest, and (if present) a .hash file with that GUID
// too. It does not decompress or otherwise inspect the .sav file's
// contents — that's internal/store.LeagueSaveStore.ValidateLeagueSaveShape,
// run separately once the file is decompressed.
func Unpack(zipPath, destDir string) (UnpackResult, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return UnpackResult{}, fmt.Errorf("opening zip: %w", err)
	}
	defer func() { _ = r.Close() }()

	manifest, err := readManifest(r.File)
	if err != nil {
		return UnpackResult{}, err
	}

	expectedGUID := strings.ToUpper(manifest.LeagueGUID.String())
	var result UnpackResult
	result.Manifest = manifest

	for _, f := range r.File {
		if f.Name == manifestFileName {
			continue
		}
		if !strings.Contains(strings.ToUpper(f.Name), expectedGUID) {
			return UnpackResult{}, fmt.Errorf(
				"file %q does not match the manifest's league GUID (%s) — this package may be corrupt or tampered with",
				f.Name, expectedGUID,
			)
		}

		extractedPath := filepath.Join(destDir, filepath.Base(f.Name))
		if err := extractZipFile(f, extractedPath); err != nil {
			return UnpackResult{}, fmt.Errorf("extracting %s: %w", f.Name, err)
		}

		switch {
		case strings.HasSuffix(f.Name, ".sav.bak"):
			result.BakPath = extractedPath
		case strings.HasSuffix(f.Name, ".hash"):
			result.HashPath = extractedPath
		case strings.HasSuffix(f.Name, ".sav"):
			result.SavPath = extractedPath
		}
	}

	if result.SavPath == "" {
		return UnpackResult{}, fmt.Errorf("package is missing the league .sav file")
	}
	if result.BakPath == "" {
		return UnpackResult{}, fmt.Errorf("package is missing the league .sav.bak file")
	}

	for _, path := range []string{result.SavPath, result.BakPath} {
		if err := checkZlibHeader(path); err != nil {
			return UnpackResult{}, fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
	}

	return result, nil
}

// checkZlibHeader is a cheap sanity check that path starts with a zlib
// stream header (confirmed CMF byte 0x78 on every real .sav file examined —
// see docs/league-transfer/failure-analysis.md Bug #2). It is not a
// substitute for actually decompressing and validating the save's table
// shape (internal/store.LeagueSaveStore.ValidateLeagueSaveShape) — it only
// catches obviously wrong input (a renamed unrelated file, a truncated
// download) before any further processing is attempted.
func checkZlibHeader(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening for header check: %w", err)
	}
	defer func() { _ = f.Close() }()

	header := make([]byte, 1)
	if _, err := io.ReadFull(f, header); err != nil {
		return fmt.Errorf("reading header: %w", err)
	}
	if header[0] != 0x78 {
		return fmt.Errorf("does not look like a zlib-compressed save file (unexpected header byte 0x%02x)", header[0])
	}
	return nil
}

func readManifest(files []*zip.File) (Manifest, error) {
	for _, f := range files {
		if f.Name != manifestFileName {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return Manifest{}, fmt.Errorf("opening manifest: %w", err)
		}
		defer func() { _ = rc.Close() }()

		var m Manifest
		if err := json.NewDecoder(rc).Decode(&m); err != nil {
			return Manifest{}, fmt.Errorf("manifest is malformed: %w", err)
		}
		if m.LeagueName == "" {
			return Manifest{}, fmt.Errorf("manifest is missing leagueName")
		}
		return m, nil
	}
	return Manifest{}, fmt.Errorf("package is missing %s", manifestFileName)
}

func extractZipFile(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()

	out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	_, err = io.Copy(out, rc)
	return err
}
