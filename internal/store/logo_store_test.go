package store_test

import (
	"context"
	"testing"
	"time"

	"smb-tools/internal/models"
	"smb-tools/internal/store"
	"smb-tools/internal/testutil"
)

func newLogoStore() *store.LogoStore {
	return store.NewLogoStore()
}

func intPtr(v int) *int { return &v }

func insertTestLogo(t *testing.T, s *store.LogoStore, db store.DBTX, teamID int, filePath string) models.TeamLogo {
	t.Helper()
	logo := models.TeamLogo{
		ID:         "logo-" + filePath,
		TeamID:     teamID,
		FilePath:   filePath,
		UploadedAt: time.Now().UTC(),
	}
	if err := s.InsertLogo(context.Background(), db, logo); err != nil {
		t.Fatalf("InsertLogo: %v", err)
	}
	return logo
}

func insertTestAssignment(t *testing.T, s *store.LogoStore, db store.DBTX, logoID string, start, end *int, assignedAt time.Time) models.TeamLogoAssignment {
	t.Helper()
	a := models.TeamLogoAssignment{
		ID:          "assign-" + logoID + "-" + assignedAt.Format("15040500"),
		LogoID:      logoID,
		StartSeason: start,
		EndSeason:   end,
		AssignedAt:  assignedAt,
	}
	if err := s.InsertLogoAssignment(context.Background(), db, a); err != nil {
		t.Fatalf("InsertLogoAssignment: %v", err)
	}
	return a
}

func TestLogoStore_NullBothCoversAnySeason(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := newLogoStore()
	ctx := context.Background()

	logo := insertTestLogo(t, s, db, 1, "assets/logos/1/uuid.png")
	insertTestAssignment(t, s, db, logo.ID, nil, nil, time.Now())

	for _, season := range []int{1, 5, 50, 9999} {
		got, err := s.GetLogoForSeason(ctx, db, 1, season)
		if err != nil {
			t.Fatalf("season %d: %v", season, err)
		}
		if got == nil {
			t.Errorf("season %d: expected logo, got nil", season)
		}
	}
}

func TestLogoStore_BoundedRangeCoverage(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := newLogoStore()
	ctx := context.Background()

	logo := insertTestLogo(t, s, db, 1, "assets/logos/1/uuid.png")
	insertTestAssignment(t, s, db, logo.ID, intPtr(3), intPtr(7), time.Now())

	covered := []int{3, 5, 7}
	for _, season := range covered {
		got, err := s.GetLogoForSeason(ctx, db, 1, season)
		if err != nil || got == nil {
			t.Errorf("season %d: expected logo (err=%v, got=%v)", season, err, got)
		}
	}

	uncovered := []int{2, 8, 100}
	for _, season := range uncovered {
		got, err := s.GetLogoForSeason(ctx, db, 1, season)
		if err != nil {
			t.Fatalf("season %d: %v", season, err)
		}
		if got != nil {
			t.Errorf("season %d: expected nil, got logo", season)
		}
	}
}

func TestLogoStore_HalfOpenRange(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := newLogoStore()
	ctx := context.Background()

	logo := insertTestLogo(t, s, db, 1, "assets/logos/1/uuid.png")
	insertTestAssignment(t, s, db, logo.ID, intPtr(5), nil, time.Now())

	for _, season := range []int{5, 6, 100} {
		got, err := s.GetLogoForSeason(ctx, db, 1, season)
		if err != nil || got == nil {
			t.Errorf("season %d: expected logo (err=%v)", season, err)
		}
	}
	got, err := s.GetLogoForSeason(ctx, db, 1, 4)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Error("season 4: expected nil, got logo")
	}
}

func TestLogoStore_SingleSeasonAssignment(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := newLogoStore()
	ctx := context.Background()

	logo := insertTestLogo(t, s, db, 1, "assets/logos/1/uuid.png")
	insertTestAssignment(t, s, db, logo.ID, intPtr(7), intPtr(7), time.Now())

	got, err := s.GetLogoForSeason(ctx, db, 1, 7)
	if err != nil || got == nil {
		t.Fatalf("season 7: expected logo (err=%v)", err)
	}
	for _, season := range []int{6, 8} {
		g, err := s.GetLogoForSeason(ctx, db, 1, season)
		if err != nil {
			t.Fatal(err)
		}
		if g != nil {
			t.Errorf("season %d: expected nil, got logo", season)
		}
	}
}

func TestLogoStore_NoAssignmentReturnsNil(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := newLogoStore()

	got, err := s.GetLogoForSeason(context.Background(), db, 99, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Error("expected nil for team with no logos")
	}
}

func TestLogoStore_LastWriteWins(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := newLogoStore()
	ctx := context.Background()

	logoA := insertTestLogo(t, s, db, 1, "assets/logos/1/a.png")
	logoB := insertTestLogo(t, s, db, 1, "assets/logos/1/b.png")

	earlier := time.Now().Add(-time.Hour)
	later := time.Now()

	insertTestAssignment(t, s, db, logoA.ID, nil, nil, earlier)
	insertTestAssignment(t, s, db, logoB.ID, nil, nil, later)

	got, err := s.GetLogoForSeason(ctx, db, 1, 5)
	if err != nil || got == nil {
		t.Fatalf("expected logo: err=%v", err)
	}
	if got.ID != logoB.ID {
		t.Errorf("last-write-wins: got %q, want %q", got.ID, logoB.ID)
	}
}

func TestLogoStore_AssignmentCountAfterDeleteAndInsert(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := newLogoStore()
	ctx := context.Background()

	logo := insertTestLogo(t, s, db, 1, "assets/logos/1/uuid.png")
	a1 := insertTestAssignment(t, s, db, logo.ID, nil, nil, time.Now())
	insertTestAssignment(t, s, db, logo.ID, intPtr(1), intPtr(3), time.Now().Add(time.Second))

	count, err := s.GetAssignmentCountForLogo(ctx, db, logo.ID)
	if err != nil || count != 2 {
		t.Fatalf("expected 2 assignments, got %d (err=%v)", count, err)
	}

	if err := s.DeleteAssignment(ctx, db, a1.ID); err != nil {
		t.Fatal(err)
	}
	count, err = s.GetAssignmentCountForLogo(ctx, db, logo.ID)
	if err != nil || count != 1 {
		t.Fatalf("after delete: expected 1 assignment, got %d (err=%v)", count, err)
	}
}

func TestLogoStore_GetLogosForTeam(t *testing.T) {
	db := testutil.NewTestDB(t)
	s := newLogoStore()
	ctx := context.Background()

	logoA := insertTestLogo(t, s, db, 1, "assets/logos/1/a.png")
	logoB := insertTestLogo(t, s, db, 1, "assets/logos/1/b.png")
	insertTestAssignment(t, s, db, logoA.ID, nil, nil, time.Now())
	insertTestAssignment(t, s, db, logoB.ID, intPtr(5), intPtr(10), time.Now())
	insertTestAssignment(t, s, db, logoB.ID, intPtr(15), nil, time.Now().Add(time.Second))

	results, err := s.GetLogosForTeam(ctx, db, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 logos, got %d", len(results))
	}

	for _, r := range results {
		switch r.Logo.ID {
		case logoA.ID:
			if len(r.Assignments) != 1 {
				t.Errorf("logoA: expected 1 assignment, got %d", len(r.Assignments))
			}
		case logoB.ID:
			if len(r.Assignments) != 2 {
				t.Errorf("logoB: expected 2 assignments, got %d", len(r.Assignments))
			}
		default:
			t.Errorf("unexpected logo ID %q", r.Logo.ID)
		}
	}
}
