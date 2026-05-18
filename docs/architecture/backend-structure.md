# Backend Structure

How the Go backend is organized and the patterns that govern it.

## Guiding Principles

**`app.go` is thin.** The Wails `App` struct is a façade. It wires up the store and service layers, exposes a small set of methods as Wails bindings, and does nothing else. All real logic lives below it. This is the single most important structural rule — it is what makes the business logic testable without Wails in the loop.

**Explicit over implicit.** Dependencies are passed in via constructors or function parameters. No global state, no package-level singletons, no `init()` side effects.

**Interfaces only where they earn their keep.** An interface is warranted when you need to swap implementations — primarily for testing (e.g., the save game reader, where you want a real SQLite connection in production and a test fixture in tests). Don't create interfaces "just in case."

**Flat packages, clear names.** Go favors fewer, broader packages over deep hierarchies. Name packages for what they contain, not for abstract layer names.

---

## Proposed Package Layout

```
backend/
  app.go                    # Wails App struct — thin façade, wires everything together
  app_test.go               # Smoke tests for Wails bindings

  db/
    companion.go            # Open companion DB, run migrations on startup
    savegame.go             # Open/close SMB save game DB (decompress → temp → read-only)
    migrations/             # golang-migrate SQL files for the companion schema
      001_initial.up.sql
      001_initial.down.sql
      ...

  store/                    # Data access layer — one file per domain area
    player.go               # PlayerStore: read/write player records in companion DB
    team.go                 # TeamStore
    franchise.go            # FranchiseStore
    season.go               # SeasonStore
    savegame_reader.go      # SaveGameReader: read-only queries against SMB save DB
    savegame_reader_test.go

  service/                  # Business logic that crosses store boundaries
    import.go               # Orchestrates reading from save game → writing to companion DB
    stats.go                # Derived stat calculations (FIP, wOBA, Pythagorean W%, etc.)
    legacy_migration.go     # SmbExplorerCompanion.db → new schema migration

  models/                   # Shared data types
    player.go
    team.go
    stats.go
    savegame.go             # Types representing the SMB save game domain

  config/
    app_directories.go      # Cross-platform app data directory resolution (already exists)
```

---

## Store Layer

The store layer owns all SQL. Each store is a struct holding a `*sql.DB`, constructed with a constructor function.

```go
// store/player.go
type PlayerStore struct {
    db *sql.DB
}

func NewPlayerStore(db *sql.DB) *PlayerStore {
    return &PlayerStore{db: db}
}

func (s *PlayerStore) GetByID(ctx context.Context, id int64) (models.Player, error) { ... }
func (s *PlayerStore) ListBySeason(ctx context.Context, seasonID int64) ([]models.Player, error) { ... }
func (s *PlayerStore) Upsert(ctx context.Context, p models.Player) error { ... }
```

Store methods return plain Go structs from `models/`. No ORM types leak out of the store layer.

Stores are **not** responsible for business logic. A store method executes a query and maps the result to a model. That's it.

---

## Service Layer

Services orchestrate across multiple stores or implement logic that doesn't belong in a single store. They also receive dependencies via constructor injection.

```go
// service/import.go
type ImportService struct {
    saveReader  SaveGameReader   // interface — can be swapped for tests
    players     *store.PlayerStore
    teams       *store.TeamStore
    seasons     *store.SeasonStore
}

func NewImportService(
    reader SaveGameReader,
    players *store.PlayerStore,
    teams *store.TeamStore,
    seasons *store.SeasonStore,
) *ImportService { ... }

func (s *ImportService) ImportSeason(ctx context.Context, seasonID int) error { ... }
```

---

## The SaveGameReader Interface

The save game database is the one place where an interface is clearly warranted. The real implementation queries a decompressed SQLite file. Test implementations can use either an in-memory SQLite DB seeded with fixture data, or a committed test fixture `.sqlite` file.

```go
// store/savegame_reader.go
type SaveGameReader interface {
    GetLeagues(ctx context.Context) ([]models.League, error)
    GetFranchises(ctx context.Context, leagueID int) ([]models.Franchise, error)
    GetSeasonPlayers(ctx context.Context, seasonID int) ([]models.Player, error)
    GetSeasonTeams(ctx context.Context, seasonID int) ([]models.Team, error)
    GetSchedule(ctx context.Context, seasonID int) ([]models.Game, error)
    GetBattingStats(ctx context.Context, seasonID int) ([]models.BattingStat, error)
    GetPitchingStats(ctx context.Context, seasonID int) ([]models.PitchingStat, error)
    // ... etc
    Close() error
}

// SqliteSaveGameReader implements SaveGameReader against a real SQLite file
type SqliteSaveGameReader struct {
    db *sql.DB
}
```

---

## Wails Bindings (app.go)

The `App` struct exposes methods that Wails generates TypeScript bindings for. These methods should be nearly trivial — they call into a service or store and return the result (or an error).

```go
// app.go
type App struct {
    ctx      context.Context
    players  *store.PlayerStore
    teams    *store.TeamStore
    importer *service.ImportService
    // ...
}

func (a *App) GetCareerBattingStats(franchiseID, seasonStart, seasonEnd int) ([]models.CareerBattingStat, error) {
    return a.players.GetCareerBattingStats(a.ctx, franchiseID, seasonStart, seasonEnd)
}
```

If a binding method is doing anything more than delegating to a store or service, that logic belongs lower in the stack.

---

## What Does NOT Go in This Structure

- No CQRS commands/queries/handlers
- No mediator or event bus
- No global `db` variable
- No `init()` functions with side effects
- No business logic in `app.go`
- No SQL strings in service or model files — SQL belongs in the store layer (or in `.sql` files if using sqlc)
