# Testing Strategy

Testability is a first-class design requirement, not something added after the fact. The original two applications had almost no automated testing, which directly caused recurring bugs and untested edge cases that made it into production. This app will not repeat that.

---

## The Core Problem with the Originals

SmbExplorerCompanion had no meaningful test suite. Without tests:
- Import edge cases (re-introduced free agents, partial season data, doubleheader scheduling) caused silent data corruption
- Derived stat calculations (BB/9, walk diff from previous season) had bugs that went undetected for months
- Regressions were caught by users, not by CI

Every structural decision in this app — the thin `app.go`, the store/service split, the `SaveGameReader` interface, the use of in-memory SQLite — is made with testability as a primary motivation.

---

## Test Layers

### Unit Tests

**Target**: Individual functions and methods in isolation.

**What gets unit tested**:
- All derived stat calculations in `service/stats.go` (FIP, wOBA, ERA, Pythagorean win%, OPS+, etc.)
- The legacy migration transformation logic (old model → new model mapping functions)
- Any pure transformation functions in the service layer
- Edge cases: zero at-bats, zero innings pitched, retired players with no career stats, etc.

**How**: Standard `go test`, table-driven tests. No database, no Wails, no file I/O. Pure functions where possible.

```go
// service/stats_test.go
func TestFieldingIndependentPitching(t *testing.T) {
    tests := []struct {
        name     string
        input    models.PitchingStats
        expected float64
    }{
        {"zero innings", models.PitchingStats{OutsPitched: 0}, 0},
        {"typical starter", models.PitchingStats{OutsPitched: 600, HR: 20, BB: 60, K: 180}, 3.47},
        // ...
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := CalculateFIP(tt.input, leagueConstant)
            if math.Abs(got-tt.expected) > 0.01 {
                t.Errorf("got %v, want %v", got, tt.expected)
            }
        })
    }
}
```

### Integration Tests

**Target**: Store methods and service orchestration against a real (in-memory) SQLite database.

**What gets integration tested**:
- All `PlayerStore`, `TeamStore`, `SeasonStore`, etc. methods
- The `ImportService.ImportSeason` flow end-to-end (read from fixture save game data → write to companion DB → verify correct records)
- The `LegacyMigrationService.Migrate` flow (in-memory old-schema DB → in-memory new-schema DB → verify)
- golang-migrate: verify all up/down migrations apply cleanly in sequence
- Query correctness for leaderboard queries (complex joins, ordering, pagination)

**How**: Each integration test opens an in-memory SQLite DB (`"file::memory:?cache=shared"`), runs migrations to bring it to the current schema, seeds fixture data, exercises the code under test, and asserts results. Tests are hermetic — each test gets its own DB.

```go
// store/player_test.go
func TestPlayerStore_GetCareerBattingStats(t *testing.T) {
    db := testutil.NewTestDB(t)         // in-memory SQLite with migrations applied
    store := NewPlayerStore(db)
    testutil.SeedPlayers(t, db, ...)    // seed fixture data

    stats, err := store.GetCareerBattingStats(context.Background(), 1, 1, 10)
    require.NoError(t, err)
    assert.Len(t, stats, 3)
    assert.Equal(t, 42, stats[0].HomeRuns)
}
```

A `testutil` package (internal, not exported) provides helpers: `NewTestDB`, `SeedPlayers`, `SeedSeasons`, fixture data builders, etc.

**SaveGameReader in integration tests**: The `SqliteSaveGameReader` implementation is tested against a committed test fixture `.sqlite` file — a minimal SMB-format SQLite database seeded with known data. This file lives at `testdata/fixtures/savegame_fixture.sqlite` and is checked into the repo. It is NOT a real user save file; it is synthetically constructed to match the SMB schema with controlled test data.

### End-to-End Tests

**Target**: Full flows from Wails binding through service through store to database.

**Reality check**: Full Wails E2E (browser automation against the running app) is expensive to set up and maintain, and the ROI diminishes fast for a desktop app with a native WebView. The strategy here is:

- **Go-layer E2E**: Test the `App` struct methods directly (not via the WebView). Since `app.go` is thin, exercising it in a Go test with a real (in-memory) companion DB and a fixture save game reader effectively tests the full Go stack end-to-end.
- **Frontend E2E**: Use Playwright or Vitest for component-level testing of the most critical UI flows (import wizard, leaderboard filtering, stats display). Not every screen needs E2E — focus on flows that have the most business logic or user-facing complexity.
- **No Wails WebView automation initially**: WebView-based E2E (launching the full packaged app and driving it with a UI automation tool) is aspirational. Don't block progress on it. Add it if/when the integration test coverage proves insufficient.

---

## CI Requirements

Every PR must pass:
- `go test ./...` — all unit and integration tests
- `go vet ./...` — static analysis
- Frontend tests (`vitest run` or equivalent)
- `golangci-lint run` — linting (configured with an agreed-upon ruleset)

The CI pipeline must be able to run on Linux (the lowest-common-denominator CI environment). `modernc.org/sqlite` being pure Go is what makes this possible without installing SQLite system packages.

---

## testutil Package

A shared `backend/testutil` package (internal to the backend module, not exported) provides:

```
testutil/
  db.go           # NewTestDB(t) — in-memory SQLite with migrations applied
  seeds.go        # SeedPlayers, SeedTeams, SeedSeasons, SeedBattingStats, etc.
  fixtures.go     # Fixture structs for common test scenarios
  savegame.go     # NewTestSaveGameReader — in-memory or fixture-based implementation
```

`NewTestDB` registers a `t.Cleanup` to close the DB, so tests never need to manage connection lifecycle manually.

---

## What Is Not Tested (Deliberately)

- The Wails runtime itself — not our code
- WebView rendering — test framework territory, not unit test territory
- The SMB save game schema — we don't own it; we test our reader against a fixture that reflects what we know the schema to be
- golang-migrate internals — we test that our migrations run, not that golang-migrate itself works

---

## Test Coverage Expectations

There is no hard percentage target, because coverage percentages are a poor proxy for test quality. The expectation instead:

- Every stat calculation function has at least one table-driven test covering the happy path and the zero/empty edge case
- Every store method has at least one integration test
- Every service flow that touches both the save game reader and the companion DB has an integration test
- The legacy migration has integration tests covering at least three scenarios: a full franchise (many seasons, all award types), a minimal franchise (one season, no awards), and a franchise with known edge cases from the original app's bug history
