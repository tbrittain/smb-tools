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
backend/
  app.go                  # Wails App struct — THIN, delegates only
  db/
    companion.go          # Open companion DB, run migrations
    savegame.go           # Decompress .sav, open read-only connection
    migrations/           # golang-migrate SQL files (*.up.sql / *.down.sql)
  store/                  # Data access — one file per domain, plain structs
  service/                # Business logic that crosses store boundaries
  models/                 # Shared Go structs (no ORM tags)
  config/                 # App data directories, franchise registry
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

**Error handling**: return errors explicitly. Don't panic. Don't swallow errors with `_`. Wrap with context using `fmt.Errorf("doing X: %w", err)`.

**Naming**: follow standard Go conventions. Exported types/functions get doc comments. Unexported helpers do not need comments unless the logic is non-obvious.

## Frontend Coding Standards

**`<script setup>` on every component.** No Options API.

**TypeScript strict mode is non-negotiable.** No `any` types. The Biome rule `noExplicitAny: "error"` enforces this.

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
- **No Options API in Vue** — Composition API with `<script setup>` only
- **No writing to the SMB save game file** (except the team transfer tool, which writes to a copy)
- **No skipping tests** because a feature "seems simple" — edge cases in stat calculations and import logic are where bugs live

## Development Commands

```sh
# Backend
go test ./...          # Run all Go tests
go vet ./...           # Static analysis
golangci-lint run      # Linting

# Frontend  
npm run dev            # Vite dev server (used by Wails in dev mode)
npm run build          # Production build
npm run lint           # Biome check
npm run lint:fix       # Biome auto-fix
npm run storybook      # Launch Storybook
npm run test           # Vitest

# Wails
wails dev              # Full app in dev mode (hot reload)
wails build            # Build for current platform
```

## Key Domain Knowledge

The SMB save game is a **ZLib-compressed SQLite 3 database**. Decompressed, it is a standard SQLite file with ~67 tables and 22 views. The schema is documented in `docs/domain/save-game-schema.md`. The game does not support mods and uses a custom proprietary C++ engine (not Unity/Unreal) — see `docs/game-integration/investigation.md`.

Player attributes are on a **1–99 scale**: Power, Contact, Speed, Fielding, Arm (hitters); Velocity, Junk, Accuracy (pitchers).

**On persisting derived stats**: not all derived stats are equal.

- **Simple rate stats** (BA = H/AB, OBP, SLG, OPS, WHIP, K/9, etc.) are deterministic functions of a single row's own columns. Prefer SQLite [generated columns](https://www.sqlite.org/gencol.html) for these — they are always in sync with their inputs and cannot diverge. Do not store them as independent columns.
- **Context-dependent stats** (wOBA, FIP, ERA+, OPS+, smbWAR) require league-wide context (linear weights, league constants, park factors) that is only available at sync time. These must be computed and persisted during the import pipeline, stored alongside the raw counts that produced them.

The failure mode to avoid (as seen in the original companion app) is storing a derived stat as an independent column that can silently diverge from its inputs due to a bug or re-import. Generated columns for simple rates, careful compute-then-store for complex rates.

Each franchise gets its own **isolated SQLite companion database**. There is no multitenant shared DB. Switching franchises = closing one `*sql.DB` and opening another.
