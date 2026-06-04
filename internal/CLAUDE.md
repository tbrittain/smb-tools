# Go Backend — Coding Standards

Standards for `internal/`. Read this before working on any Go code.

## Package Responsibilities

- **`app.go`** (root) — wires dependencies and exposes Wails bindings only. Zero business logic. If a method does more than delegate to a store or service, that logic is in the wrong place.
- **`internal/store/`** — data access only. Each store holds a `*sql.DB` and executes queries. No business logic.
- **`internal/service/`** — business logic that crosses store boundaries. Orchestrates reads and writes across multiple stores.
- **`internal/models/`** — shared Go structs. No ORM tags. No SQL.
- **`internal/db/`** — database connection management and migration runner.
- **`internal/config/`** — app data directory resolution and save file paths.
- **`internal/testutil/`** — test helpers. Only imported in `_test.go` files.

## Coding Patterns

**Explicit over implicit.** Dependencies are passed via constructors. No global variables. No `init()` side effects. No package-level singletons.

**Interfaces only where they earn their keep** — primarily where you need to swap implementations for testing (e.g., `SaveGameReader`). Don't create interfaces speculatively.

**No CQRS. No mediator pattern. No MediatR-style dispatch.** These were over-abstractions in the original C# apps. Go has functions; use them.

**SQL belongs in the store layer** — inline in store method bodies or in `.sql` files. Never in service or model files.

**Store methods are dumb.** They execute a query and map to a model struct. No business logic.

**Avoid primitive obsession.** Define named types for values with a specific domain meaning — `GameVersion`, `SHA256Hex`, `SnapshotFileName`, etc. A plain `string` where a constrained type is expected lets the wrong value slip through silently.

**Error handling**: return errors explicitly. Don't panic. Don't swallow with `_`. Wrap with context:

```go
return fmt.Errorf("doing X: %w", err)
```

**Naming**: follow standard Go conventions. Exported types/functions get doc comments. Unexported helpers do not unless the logic is non-obvious.

## Testing

**Testing is not optional.** Test coverage is a first-class requirement.

- **Unit tests** for all pure functions: every stat calculation in `service/stats.go`, all data transformation logic, every business rule. Use table-driven tests.
- **Integration tests** for all store methods: use an in-memory SQLite DB via `testutil.NewTestDB(t)`. Each test gets its own DB with migrations applied.
- **`testutil` package** (`internal/testutil/`) provides: `NewTestDB`, `NewTestRegistryDB`, seed helpers (`SeedPlayers`, `SeedTeams`, etc.), and a test `SaveGameReader` backed by a fixture SQLite file.
- Tests must pass with `go test ./...` on Linux (CI). `modernc.org/sqlite` being pure Go is what makes this possible.
- **No mocking the database.** Use real in-memory SQLite. Mocked DB tests don't catch real query bugs.
- **Non-trivial bug fixes require tests.** If a fix corrects business logic (not a typo or rename), a test that would have caught the bug must accompany it. "Non-trivial" means: nullable data, domain encoding/decoding, multi-step data flow (import pipeline, migration), conditional branching on domain values, or anything where a wrong assumption caused the original bug.

**Designed for testability:**
- Stat calculation functions are pure (input → output, no side effects)
- Store methods receive `*sql.DB` via constructor — swap in `:memory:` for tests
- Service methods receive store structs via constructor — no hidden dependencies
- `app.go` bindings are thin enough that testing the service layer IS effectively testing the binding

## Database Migrations

Migrations live in `internal/db/migrations/` split across two sets:

- `companion/` — schema for the per-franchise companion DB (the bulk of the app's schema)
- `registry/` — schema for the franchise registry DB (franchise list and metadata only)

**Naming convention**: `{version}_{name}.up.sql` where version is a zero-padded 4-digit integer — e.g., `0012_add_salary_column.up.sql`. The migration runner (`internal/db/migrate.go`) parses the integer prefix to determine order.

**Migrations are immutable once deployed.** Applied versions are recorded in a `schema_migrations` table and skipped on subsequent runs. Editing an existing migration file has no effect for users who have already run it — they will never see the change. Always add a new numbered file; never edit an existing one.

**No down migrations.** The runner only processes `.up.sql` files. There is no rollback mechanism. Design migrations to be additive where possible (new columns with defaults, new tables, new indexes). Destructive changes cannot be undone automatically — think carefully before dropping or renaming anything.

**Each migration runs in a single transaction.** If any statement fails, the entire migration rolls back and the app errors at startup. Keep each file focused on one logical change.

**The runner is custom, not golang-migrate.** It lives in `internal/db/migrate.go` and reads from an `embed.FS`. Do not assume golang-migrate behavior — there is no `force`, no `dirty` state, and no CLI tooling.

**Test DB applies all migrations.** `testutil.NewTestDB` runs every migration before each test. Add the migration file before writing tests that depend on the new schema, or the test DB will be missing the columns/tables your test expects.

## What Not to Do

- No CQRS / MediatR-style patterns — not idiomatic Go
- No `mattn/go-sqlite3` — requires CGO, breaks cross-platform builds
- No business logic in `app.go` — it delegates only
- No SQL strings outside the store layer
- No inventing save game column names — verify every name against the SMB3Explorer SQL files (https://github.com/tbrittain/SMB3Explorer, under `SMB3Explorer/Resources/Sql/`) or a real decompressed save. See "Save Game SQL — Real Schema Required" in the root `CLAUDE.md`.
