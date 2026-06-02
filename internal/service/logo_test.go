package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"smb-tools/internal/config"
	"smb-tools/internal/service"
	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

func intPtr(v int) *int { return &v }

// writeTempImage writes a file with the given magic bytes to a temp file
// and returns its path.
func writeTempImage(t *testing.T, magic []byte, name string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, magic, 0o600); err != nil {
		t.Fatalf("writeTempImage: %v", err)
	}
	return path
}

// pngMagic is a minimal valid PNG header (8-byte signature + padding).
var pngMagic = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}

// jpegMagic is a minimal valid JPEG header.
var jpegMagic = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}

// invalidMagic is not a valid image.
var invalidMagic = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}

func newTestLogoService(t *testing.T) (*service.LogoService, *config.AppDirs, string) {
	t.Helper()
	franchisesDir := t.TempDir()
	dirs := &config.AppDirs{
		DataDir:       t.TempDir(),
		FranchisesDir: franchisesDir,
	}
	dirs.RegistryPath = filepath.Join(dirs.DataDir, "registry.db")
	svc := service.NewLogoService(store.NewLogoStore(), dirs)
	return svc, dirs, "test-franchise-id"
}

func TestLogoService_UploadAndAssign_FileCopied(t *testing.T) {
	svc, dirs, franchiseID := newTestLogoService(t)
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	src := writeTempImage(t, pngMagic, "logo.png")

	logo, _, err := svc.UploadAndAssign(ctx, db, franchiseID, 42, src, nil, nil)
	if err != nil {
		t.Fatalf("UploadAndAssign: %v", err)
	}

	fullPath := filepath.Join(dirs.FranchiseDir(franchiseID), filepath.FromSlash(logo.FilePath))
	if _, err := os.Stat(fullPath); err != nil {
		t.Fatalf("logo file not found at %s: %v", fullPath, err)
	}
}

func TestLogoService_UploadAndAssign_RelativeFilePath(t *testing.T) {
	svc, dirs, franchiseID := newTestLogoService(t)
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	src := writeTempImage(t, pngMagic, "logo.png")

	logo, _, err := svc.UploadAndAssign(ctx, db, franchiseID, 42, src, nil, nil)
	if err != nil {
		t.Fatalf("UploadAndAssign: %v", err)
	}

	if filepath.IsAbs(logo.FilePath) {
		t.Errorf("FilePath should be relative, got absolute: %s", logo.FilePath)
	}
	if !filepath.IsAbs(filepath.Join(dirs.FranchiseDir(franchiseID), filepath.FromSlash(logo.FilePath))) {
		t.Error("joined path should be absolute")
	}
}

func TestLogoService_UploadAndAssign_RejectsInvalidMagicBytes(t *testing.T) {
	svc, _, franchiseID := newTestLogoService(t)
	db := testutil.NewTestDB(t)

	src := writeTempImage(t, invalidMagic, "not_an_image.png")

	_, _, err := svc.UploadAndAssign(context.Background(), db, franchiseID, 42, src, nil, nil)
	if err == nil {
		t.Error("expected error for non-PNG/JPEG file, got nil")
	}
}

func TestLogoService_UploadAndAssign_AcceptsJPEG(t *testing.T) {
	svc, _, franchiseID := newTestLogoService(t)
	db := testutil.NewTestDB(t)

	src := writeTempImage(t, jpegMagic, "logo.jpg")

	_, _, err := svc.UploadAndAssign(context.Background(), db, franchiseID, 42, src, nil, nil)
	if err != nil {
		t.Fatalf("expected JPEG to be accepted: %v", err)
	}
}

func TestLogoService_GetLogoURLForSeason_ReturnsURL(t *testing.T) {
	svc, _, franchiseID := newTestLogoService(t)
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	src := writeTempImage(t, pngMagic, "logo.png")
	_, _, err := svc.UploadAndAssign(ctx, db, franchiseID, 42, src, nil, nil)
	if err != nil {
		t.Fatalf("UploadAndAssign: %v", err)
	}

	url, err := svc.GetLogoURLForSeason(ctx, db, 42, 5)
	if err != nil {
		t.Fatalf("GetLogoURLForSeason: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty URL for season with logo")
	}
	if len(url) < 2 || url[:8] != "/_logos/" {
		t.Errorf("URL should start with /_logos/, got %q", url)
	}
}

func TestLogoService_GetLogoURLForSeason_ReturnsEmptyWhenNone(t *testing.T) {
	svc, _, _ := newTestLogoService(t)
	db := testutil.NewTestDB(t)

	url, err := svc.GetLogoURLForSeason(context.Background(), db, 99, 1)
	if err != nil {
		t.Fatal(err)
	}
	if url != "" {
		t.Errorf("expected empty URL, got %q", url)
	}
}

func TestLogoService_DeleteAssignment_HardDeletesFileWhenLastRef(t *testing.T) {
	svc, dirs, franchiseID := newTestLogoService(t)
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	src := writeTempImage(t, pngMagic, "logo.png")
	logo, assignment, err := svc.UploadAndAssign(ctx, db, franchiseID, 42, src, nil, nil)
	if err != nil {
		t.Fatalf("UploadAndAssign: %v", err)
	}

	fullPath := filepath.Join(dirs.FranchiseDir(franchiseID), filepath.FromSlash(logo.FilePath))
	if _, err := os.Stat(fullPath); err != nil {
		t.Fatalf("file should exist before delete: %v", err)
	}

	if err := svc.DeleteAssignment(ctx, db, franchiseID, assignment.ID); err != nil {
		t.Fatalf("DeleteAssignment: %v", err)
	}

	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		t.Error("file should be deleted when last assignment is removed")
	}
}

func TestLogoService_DeleteAssignment_LeavesFileWhenOtherAssignmentsRemain(t *testing.T) {
	svc, dirs, franchiseID := newTestLogoService(t)
	db := testutil.NewTestDB(t)
	ctx := context.Background()

	src := writeTempImage(t, pngMagic, "logo.png")
	logo, firstAssignment, err := svc.UploadAndAssign(ctx, db, franchiseID, 42, src, intPtr(1), intPtr(5))
	if err != nil {
		t.Fatalf("UploadAndAssign: %v", err)
	}
	// Add a second assignment for the same logo.
	if _, err := svc.AssignExisting(ctx, db, logo.ID, intPtr(10), nil); err != nil {
		t.Fatalf("AssignExisting: %v", err)
	}

	fullPath := filepath.Join(dirs.FranchiseDir(franchiseID), filepath.FromSlash(logo.FilePath))

	// Deleting the first assignment should NOT delete the file.
	if err := svc.DeleteAssignment(ctx, db, franchiseID, firstAssignment.ID); err != nil {
		t.Fatalf("DeleteAssignment: %v", err)
	}

	if _, err := os.Stat(fullPath); err != nil {
		t.Errorf("file should still exist when another assignment references it: %v", err)
	}
}
