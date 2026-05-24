# CLAUDE.md

Context and coding standards for smb-tools. Read this at the start of every session.

## What This Project Is

smb-tools is a cross-platform desktop application (Wails v2 + Go backend + Vue 3 frontend) that reads Super Mega Baseball 3/4 save game files and provides a Baseball Reference–style franchise history viewer, stat tracker, and team management tool. It is a ground-up rewrite and consolidation of two prior Windows-only C# applications: SMB3Explorer and SmbExplorerCompanion.

**The single most important UX principle**: one button click imports an entire season from the save game. No CSV exports. No multi-step wizards. No two separate apps.

## Read These Docs First

All design decisions, rationale, and domain knowledge live in `docs/`. Before implementing anything, read:

- `docs/architecture/decisions.md` — committed technology choices (don't relitigate these)
- `docs/architecture/backend-structure.md` — Go package layout and patterns
- `docs/architecture/data-layer.md` — two-database architecture, per-franchise DBs, snapshot strategy
- `docs/architecture/testing-strategy.md` — testing requirements and approach
- `docs/architecture/ux-flows.md` — core user flows
- `docs/roadmap.md` — current phase and what's in scope
- `docs/domain/` — save game schema, player model, stats, traits, awards

## Tech Stack

**Backend**
- Go 1.26
- Wails v2.12.0 (NOT v3 — still alpha as of May 2026)
- `modernc.org/sqlite` — pure Go SQLite driver, no CGO
- `golang-migrate` — SQL-file-based schema migrations
- `github.com/fsnotify/fsnotify` — filesystem watching (Phase 13 only)

**Frontend**
- Vue 3.5 with `<script setup>` Composition API throughout
- TypeScript 6, strict mode
- Vite 8
- PrimeVue 4 + `@primeuix/themes` — component library
- AG Grid Community 35 (`ag-grid-vue3`) — data grids
- Apache ECharts 6 + `vue-echarts` 8 — charts and visualizations
- Pinia 3 — global state (franchise context, app state)
- vue-router 5 (hash history — required for Wails)
- Storybook — component development and visual testing
- Biome 2 — linting and formatting

**No Tailwind. Ever.**

## Project Structure

```
app.go                    # Wails App struct — THIN, delegates only
main.go                   # Wails entry point
internal/
  config/                 # App data directories, franchise registry paths
  db/
    companion.go          # Open per-franchise companion DB, run migrations
    registry.go           # Open registry DB, run migrations
    savegame.go           # Decompress .sav, open read-only connection
    migrate.go            # SQL-file migration runner (embed.FS-based)
    migrations/           # SQL migration files ({version}_{name}.up.sql)
      registry/
      companion/
  store/                  # Data access — one file per domain, plain structs
  service/                # Business logic that crosses store boundaries
  models/                 # Shared Go structs (no ORM tags)
  testutil/               # Test helpers: NewTestDB, NewTestRegistryDB, seed helpers
frontend/
  src/
    components/           # Reusable components (each should have a Story)
    pages/                # Route-level page components
    stores/               # Pinia stores
    composables/          # Reusable Composition API logic
    router.ts             # vue-router hash history config
    main.ts               # App bootstrap
  .storybook/             # Storybook configuration
docs/                     # All architecture decisions, domain knowledge, roadmap
```

## Go Coding Standards

**The cardinal rule: `app.go` is thin.** It wires dependencies and exposes Wails bindings. Zero business logic lives there. If an `app.go` method does more than delegate to a store or service, that logic is in the wrong place.

**Explicit over implicit.** Dependencies are passed via constructors. No global variables. No `init()` side effects. No package-level singletons.

**Interfaces only where they earn their keep** — primarily where you need to swap implementations for testing (e.g., `SaveGameReader`). Don't create interfaces speculatively.

**No CQRS. No mediator pattern. No MediatR-style dispatch.** These were over-abstractions in the original C# apps. Go has functions; use them.

**SQL belongs in the store layer** — either inline in store method bodies or in `.sql` files if using sqlc. Never in service or model files.

**Store methods are dumb.** They execute a query and map to a model struct. No business logic in store methods.

**Avoid primitive obsession.** Don't use `string` or `int` for values that have a more specific domain meaning. Define named types for things like `GameVersion`, `SHA256Hex`, `SnapshotFileName`, etc. This produces self-documenting code and prevents accidentally passing the wrong value where a semantically distinct type is expected. The question to ask: "would a plain `string` here let me accidentally pass any string, or does this field have a constrained set of values or a specific format?" If the latter, it warrants a type.

**Error handling**: return errors explicitly. Don't panic. Don't swallow errors with `_`. Wrap with context using `fmt.Errorf("doing X: %w", err)`.

**Naming**: follow standard Go conventions. Exported types/functions get doc comments. Unexported helpers do not need comments unless the logic is non-obvious.

## Frontend Coding Standards

**`<script setup>` on every component.** No Options API.

**TypeScript strict mode is non-negotiable.** No `any` types. The Biome rule `noExplicitAny: "error"` enforces this. `as unknown as T` is equally forbidden — it is an unsafe cast that defeats the type system the same way `any` does. If you find yourself needing it, fix the source type instead (use class constructors, fix prop types, or use proper type guards).

**Formatting is enforced by Biome**: 2-space indent, single quotes, no semicolons, 120-char line width. Run `npm run lint:fix` before committing. Do not manually reformat — let Biome do it.

**Component structure order within `<script setup>`**:
1. Imports
2. Props / emits definitions
3. Injected dependencies (inject, useRoute, useRouter, stores)
4. Reactive state (ref, reactive)
5. Computed properties
6. Watchers
7. Lifecycle hooks
8. Functions / event handlers

**Pinia stores** are for app-wide state (current franchise, app loading state). Component-local state stays in `ref()`/`reactive()` within the component. Don't reach for Pinia for state that doesn't need to be shared.

**Composables** for logic reused across more than one component. Composables live in `src/composables/`.

**No inline styles.** Use scoped CSS in the component's `<style scoped>` block or PrimeVue design tokens.

**PrimeVue components first.** Before hand-rolling any UI primitive — tables, dialogs, dropdowns, paginators, tabs, checkboxes — check whether PrimeVue 4 already provides it. Use `DataTable` + `Column` for any tabular data (never `<table>`/`<thead>`/`<tbody>` by hand), `MultiSelect` for multi-value pickers, `TabView`/`TabPanel` for tabs, etc. Custom HTML primitives are only acceptable when no PrimeVue component fits the use case and the component is too small to warrant a library import.

**Server-side pagination and filtering — always.** smb-tools is a data-heavy application that can accumulate hundreds of seasons and thousands of player-seasons. Pagination and filtering must be implemented in the Go store layer (SQL `LIMIT`/`OFFSET`, `WHERE` clauses, `ORDER BY`), never in the Vue layer by slicing or filtering a full array that was already fetched. Client-side filtering of a server-truncated result set is always wrong — if the backend returns the top 10 rows by OPS and the frontend then filters by team, it will silently miss players ranked 11th+. The only exception is instant UI feedback for a user-typed search that debounces to a real server call; lightweight ephemeral UI state (tab selection, column sort direction on a fully-loaded small dataset) is acceptable. When in doubt, push the predicate to SQL.

**DataTable column widths — always `min-width`, never `width`.** PrimeVue `<Column>` elements must use `style="min-width: Xpx"` so columns can flex and scale. Never use `style="width: Xpx"` — static widths prevent columns from growing and break the layout on wider screens. A `min-width` sets the floor; the column still expands to fill available space.

**Wails bindings** are imported from `../../wailsjs/go/main/App` and called as async functions. Always handle errors explicitly — Wails surfaces Go errors as rejected promises.

## Testing Standards

**Testing is not optional.** The original apps had almost no automated tests and were consistently buggy as a result. Test coverage is a first-class requirement.

### Go (Backend)

- **Unit tests** for all pure functions: every stat calculation in `service/stats.go`, all data transformation logic, every business rule. Use table-driven tests.
- **Integration tests** for all store methods: use an in-memory SQLite DB (`file::memory:?cache=shared`) with migrations applied. Each test gets its own DB via `testutil.NewTestDB(t)`.
- **`testutil` package** (`backend/testutil/`) provides: `NewTestDB`, seed helpers (`SeedPlayers`, `SeedTeams`, etc.), a test `SaveGameReader` implementation backed by a fixture SQLite file.
- The `SaveGameReader` interface exists specifically to enable testing without a real save game file.
- Tests must pass with `go test ./...` on Linux (CI). `modernc.org/sqlite` being pure Go is what makes this possible.
- **No mocking the database.** Use real in-memory SQLite. Mocked DB tests don't catch real query bugs.

### Frontend

- **Vitest** for unit tests on composables, utility functions, and stat calculation logic ported to the frontend.
- **Storybook** for component-level development and visual regression. Every non-trivial component in `src/components/` should have a corresponding Story covering: default state, empty/zero state, loading state, and any notable variants (e.g., a batting stat row with a Hall of Famer flag, a player card with no awards).
- **Storybook is not optional for components.** If you're building a reusable component, build the Story alongside it. This is how component invariants are validated without needing a live database.
- Playwright for E2E flows on the most critical paths (import wizard, franchise creation) — added after core is stable.

### What "designed for testability" means here

- Functions that compute derived stats are pure (input → output, no side effects) so they can be unit tested trivially
- Store methods receive `*sql.DB` via constructor, not from a global — swap in `:memory:` for tests
- Service methods receive store interfaces/structs via constructor — no hidden dependencies
- `app.go` bindings are so thin that testing the service layer IS effectively testing the binding
- Vue components receive data via props, not by directly calling Wails bindings — makes component testing possible without Wails

## What NOT To Do

- **No Tailwind** — ever
- **No CQRS / MediatR-style patterns** — not idiomatic Go
- **No Wails v3** — still alpha, breaking changes, no stable release date
- **No `mattn/go-sqlite3`** — requires CGO, breaks cross-platform builds
- **No business logic in `app.go`** — it delegates only
- **No SQL strings outside the store layer**
- **No `any` types in TypeScript** — Biome will error
- **No `as unknown as T` casts in TypeScript** — this is an unsafe double-cast that defeats the type system as surely as `any`. Fix the source type (use class constructors, adjust prop types, add proper type guards) instead.
- **No hand-editing of autogenerated files** — `wailsjs/go/models.ts`, `wailsjs/go/main/App.js`, and `wailsjs/go/main/App.d.ts` are owned by Wails. After adding or changing any Wails-bound Go method or DTO, run `wails build` to regenerate them. Editing them by hand creates divergence that breaks on the next build.
- **No Options API in Vue** — Composition API with `<script setup>` only
- **No writing to the SMB save game file** (except the team transfer tool, which writes to a copy)
- **No skipping tests** because a feature "seems simple" — edge cases in stat calculations and import logic are where bugs live
- **Non-trivial bug fixes require tests.** If a fix corrects business logic (not a typo or rename), a test that would have caught the bug must accompany it. "Non-trivial" means: any fix involving nullable data, domain encoding/decoding, multi-step data flow (import pipeline, migration), conditional branching on domain values, or anything where a wrong assumption caused the original bug. When in doubt, write the test.
- **No inventing save game column names** — verify every name against the SMB3Explorer SQL files or a real decompressed save before writing it. A fixture built on made-up names produces tests that prove nothing. See "Save Game SQL — Real Schema Required" above.

## Development Commands

```sh
# Backend
go test ./...          # Run all Go tests
go vet ./...           # Static analysis
golangci-lint run      # Linting — run this before every commit (v2, install once: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest)

# Frontend  
npm run dev            # Vite dev server (used by Wails in dev mode)
npm run build          # Production build
npm run lint           # Biome check
npm run lint:fix       # Biome auto-fix
npm run storybook      # Launch Storybook
npm run test           # Vitest

# Wails
wails dev              # Full app in dev mode (hot reload)
wails build            # Build for current platform — also regenerates wailsjs/ bindings
                       # Run this after any Go binding/DTO changes and before opening a PR
```

## Save Game SQL — Real Schema Required

**The only acceptable source of truth for save game table and column names is the real SMB save game schema.** Making tests pass is not the goal — the fixture must mirror the real schema, or passing tests are meaningless.

Before writing any query that touches the save game database, verify every table and column name against one of these two authoritative sources:

1. **SMB3Explorer SQL files** at `C:\Users\Trey\source\SMB3Explorer\SMB3Explorer\Resources\Sql\` — battle-tested queries against the real game. This is the fastest reference; check the relevant `.sql` file before using any column name.

2. **A decompressed real save file** — decompress any `.sav` from `%LOCALAPPDATA%\Metalhead\Super Mega Baseball 4\` using `internal/db.DecompressAndOpen` and run `PRAGMA table_info(<table_name>)` to see the actual columns.

The same requirement applies to the test fixture in `internal/testutil/savegame.go`. The fixture exists to run the import pipeline against a controlled dataset — it must use the real schema's column names, not invented ones. A fixture built with made-up column names produces tests that prove nothing about whether the real game will work.

## Save Game Import — Anti-Corruption Layer

**Raw save game values must never leak into the companion DB or the domain layer.** The SMB save game stores many values as opaque integer codes (position codes, hand codes, chemistry codes, option keys). These are implementation details of Metalhead's engine, not smb-tools domain concepts.

The `SaveGameReader` and the import pipeline in `internal/service/import.go` form the **anti-corruption layer** (ACL): they translate raw codes into meaningful domain values before those values ever touch the companion database or a domain model struct.

**The rule**: every field read from the save game that carries a coded integer value must be translated to a human-readable domain string (or typed constant) in Go — in the reader or import layer — before being stored or returned. SQL `CASE` expressions in the query are not an acceptable substitute; translation belongs in Go code.

Examples of what this looks like in practice:

```go
// Raw integer code from save game → domain string in Go
p.PrimaryPos    = saveGamePosition(rawPrimaryPos)    // "8" → "CF"
p.PitcherRole   = saveGamePitcherRole(rawPitcherRole) // "1" → "SP"
p.ThrowHand     = saveGameHand(throwCode)             // 0   → "L"
p.BatHand       = saveGameHand(batCode)               // 2   → "S"
p.ChemistryType = saveGameChemistry(chemCode)         // 0   → "Competitive"
```

The translation functions (`saveGamePosition`, `saveGamePitcherRole`, `saveGameHand`, `saveGameChemistry`) live in `internal/store/sqlite_savegame_reader.go` and are the canonical mapping between save game codes and domain values.

**The test fixture must also use domain values, not raw codes.** `internal/testutil/savegame.go` seeds the fixture with domain strings (e.g., `"CF"`, `"SP"`, `"L"`, `"Competitive"`) — not the raw integers the save game stores. A fixture seeded with raw codes would only prove that translation was skipped, not that it worked correctly.

## Key Domain Knowledge

The SMB save game is a **ZLib-compressed SQLite 3 database**. Decompressed, it is a standard SQLite file with ~67 tables and 22 views. The schema is documented in `docs/domain/save-game-schema.md`. The game does not support mods and uses a custom proprietary C++ engine (not Unity/Unreal) — see `docs/game-integration/investigation.md`.

Player attributes are on a **1–99 scale**: Power, Contact, Speed, Fielding, Arm (hitters); Velocity, Junk, Accuracy (pitchers).

**On persisting derived stats**: not all derived stats are equal.

- **Simple rate stats** (BA = H/AB, OBP, SLG, OPS, WHIP, K/9, etc.) are deterministic functions of a single row's own columns. Prefer SQLite [generated columns](https://www.sqlite.org/gencol.html) for these — they are always in sync with their inputs and cannot diverge. Do not store them as independent columns.
- **Context-dependent stats** (wOBA, FIP, ERA+, OPS+, smbWAR) require league-wide context (linear weights, league constants, park factors) that is only available at sync time. These must be computed and persisted during the import pipeline, stored alongside the raw counts that produced them.

The failure mode to avoid (as seen in the original companion app) is storing a derived stat as an independent column that can silently diverge from its inputs due to a bug or re-import. Generated columns for simple rates, careful compute-then-store for complex rates.

Each franchise gets its own **isolated SQLite companion database**. There is no multitenant shared DB. Switching franchises = closing one `*sql.DB` and opening another.
