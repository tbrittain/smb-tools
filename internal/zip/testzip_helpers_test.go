package zip_test

import (
	"archive/zip"
	"os"
	"strings"
	"testing"
)

// The helpers in this file build hand-crafted zips for negative test cases
// (missing manifest, missing file) that internalzip.Pack would never
// produce itself — Pack always writes a complete, well-formed package.

func newTestZip(t *testing.T, path string) *zip.Writer {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("creating test zip %s: %v", path, err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return zip.NewWriter(f)
}

func closeTestZip(t *testing.T, zw *zip.Writer) {
	t.Helper()
	if err := zw.Close(); err != nil {
		t.Fatalf("closing test zip: %v", err)
	}
}

func addFileToZip(t *testing.T, zw *zip.Writer, srcPath, nameInZip string) {
	t.Helper()
	data, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("reading %s: %v", srcPath, err)
	}
	addBytesToZip(t, zw, nameInZip, data)
}

func addBytesToZip(t *testing.T, zw *zip.Writer, nameInZip string, data []byte) {
	t.Helper()
	w, err := zw.Create(nameInZip)
	if err != nil {
		t.Fatalf("creating zip entry %s: %v", nameInZip, err)
	}
	if _, err := w.Write(data); err != nil {
		t.Fatalf("writing zip entry %s: %v", nameInZip, err)
	}
}

// rebuildZipWithoutSuffix copies every entry from srcZip to destZip except
// those whose name ends with excludeSuffix.
func rebuildZipWithoutSuffix(t *testing.T, srcZip, destZip, excludeSuffix string) {
	t.Helper()
	r, err := zip.OpenReader(srcZip)
	if err != nil {
		t.Fatalf("opening %s: %v", srcZip, err)
	}
	defer func() { _ = r.Close() }()

	out, err := os.Create(destZip)
	if err != nil {
		t.Fatalf("creating %s: %v", destZip, err)
	}
	defer func() { _ = out.Close() }()

	zw := zip.NewWriter(out)
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, excludeSuffix) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("opening entry %s: %v", f.Name, err)
		}
		w, err := zw.Create(f.Name)
		if err != nil {
			_ = rc.Close()
			t.Fatalf("creating entry %s: %v", f.Name, err)
		}
		buf := make([]byte, f.UncompressedSize64)
		if _, err := rc.Read(buf); err != nil && err.Error() != "EOF" {
			_ = rc.Close()
			t.Fatalf("reading entry %s: %v", f.Name, err)
		}
		if _, err := w.Write(buf); err != nil {
			_ = rc.Close()
			t.Fatalf("writing entry %s: %v", f.Name, err)
		}
		_ = rc.Close()
	}
	closeTestZip(t, zw)
}
