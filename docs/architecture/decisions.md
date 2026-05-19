# Technology Decisions

Committed choices for the smb-tools rewrite. These are not up for re-evaluation during implementation — if a decision needs to be revisited, update this document with the reasoning.

---

## Desktop Framework: Wails v2

**Decision**: Wails v2 (not Tauri, not Electron, not anything else).

**Rationale**:
- An existing, working reference application built with Wails v2 provides a known-good pattern for data-heavy Go + WebView apps
- Go is the preferred backend language (see below); Wails v2 is the most mature Go-native desktop framework
- Electron is explicitly ruled out — too heavy, Chromium-bundled, poor fit for a performance-conscious cross-platform app
- Tauri (Rust) is a strong alternative but carries a learning tax: no existing Rust reference app, longer build times, more boilerplate for SQLite-heavy data work
- Wails v3 exists but is still in RC; v2 is stable and production-proven. Migration to v3 is a future option

**Wails v3 status** (verified May 2026): v3 is at **v3.0.0-alpha.93** (released May 15, 2026) — still explicitly marked alpha with no stable release date. The maintainers' own statement: "When it's ready. And it's nearly ready." (December 2025), with no ETA and no confirmed beta milestone. v3 also carries significant breaking changes from v2 (new application structure, new build system via Taskfile, different API style, different bindings generation). v2.12.0 (released March 26, 2026) is the current stable production release and what this project targets.

**Implication**: The repo will be re-scaffolded from scratch using `wails init` targeting v2.12.0. The previous scaffold is discarded entirely.

---

## Backend Language: Go

**Decision**: Go for all backend logic.

**Rationale**:
- Idiomatic Go is well-suited to this app's core operations: file I/O, SQLite queries, data transformation, concurrent save file processing
- The existing reference application demonstrates Go's viability for data-heavy Wails apps
- Go's standard library, interface model, and testing tooling align well with the testability requirements (see `testing-strategy.md`)
- No Rust, no Node.js in the backend

---

## SQLite Driver: modernc.org/sqlite

**Decision**: `modernc.org/sqlite` as the SQLite driver for all database connections.

**Rationale**:
- Pure Go implementation — no CGO dependency, no C toolchain required on any platform
- Critical for cross-platform builds (Windows, macOS, Linux) without per-platform C compiler setup
- `mattn/go-sqlite3` is more battle-tested and marginally faster but requires CGO, which creates real friction in cross-compilation and CI pipelines
- The performance difference is not meaningful for this application's query patterns

---

## Database Migrations: golang-migrate

**Decision**: `golang-migrate` for companion database schema migrations, driven by plain SQL migration files.

**Rationale**:
- Migrations are plain `.sql` files — readable, diffable, no ORM lock-in
- Runs automatically at app startup against the companion DB
- Pairs naturally with `sqlc` if that tool is adopted (see `open-decisions.md`)
- Gives a clear, auditable history of schema evolution

**Scope**: golang-migrate manages only the companion database (the app's own schema). The SMB save game database schema is not owned or migrated by this app.

---

## No CQRS, No MediatR-style Patterns

**Decision**: Idiomatic Go. No CQRS, no command/query mediator, no over-abstraction.

**Rationale**:
- The SmbExplorerCompanion used MediatR + CQRS, which was an over-abstraction imported from .NET culture. It added indirection without meaningful benefit for the scale of this application.
- Go's idioms favor explicit, flat, readable code over pattern-driven architecture
- A store layer + service layer is sufficient. Functions call other functions. Dependencies are passed in explicitly.
- If a pattern isn't solving a real problem present in this codebase right now, it doesn't belong here

---

## Frontend Stack

**Decision**: Vue 3 + TypeScript + Vite, with PrimeVue as the component library, AG Grid Community for data grids, and Apache ECharts + vue-echarts for charting. Pinia for global state. Biome for linting and formatting.

**Rationale**:
- **Vue 3**: developer has an existing Wails + Vue reference app (git-analytics) with proven patterns. Vue's Composition API reactivity model is preferred over React's useState/useEffect model. Svelte was considered but its ecosystem for data grids and charting is notably weaker, and SvelteKit's primary features (SSR, file-based routing) provide no benefit in a Wails context.
- **PrimeVue**: preferred over Naive UI and Element Plus based on aesthetics and existing familiarity from the previous smb-tools scaffold. Among the highest-download Vue component libraries. Tailwind-based alternatives (shadcn-vue) were explicitly ruled out.
- **AG Grid Community**: the gold standard for web data grids. Handles the franchise-wide leaderboards, sortable/filterable stat tables, and paginated season breakdowns that plain HTML tables cannot. The reference app (git-analytics) used plain tables; smb-tools data volume and interactivity requirements exceed what that approach supports.
- **Apache ECharts + vue-echarts**: already used in the reference app. Covers every chart type needed — bar (season trends), radar/spider (player attribute percentiles), scatter (player comparisons). Official Vue integration via vue-echarts.
- **Pinia**: not used in git-analytics (simpler app, no shared global state). smb-tools has franchise-level global state (which franchise DB is open, current franchise metadata) that needs to be accessible across many components without deep prop-drilling. Pinia is the idiomatic Vue 3 answer.
- **Biome**: same config as reference app (2-space indent, single quotes, no semicolons, 120-char line width).

**Pinned versions** (current stable at time of scaffolding):

| Package | Version |
|---------|---------|
| vue | 3.5.34 |
| vue-router | 5.0.7 |
| vite | 8.0.13 |
| @vitejs/plugin-vue | 6.0.7 |
| typescript | 6.0.3 |
| vue-tsc | 3.3.0 |
| primevue | 4.5.5 |
| @primeuix/themes | 2.0.3 |
| echarts | 6.0.0 |
| vue-echarts | 8.0.1 |
| ag-grid-community | 35.3.0 |
| ag-grid-vue3 | 35.3.0 |
| pinia | 3.0.4 |
| @biomejs/biome | 2.4.15 |

**Go / Wails versions** (match at scaffolding time):

| | Version |
|--|---------|
| Go | 1.26.3 |
| Wails | v2.12.0 |

---

## No Electron

**Decision**: Electron is ruled out.

**Rationale**: Bundles a full Chromium instance (~150MB+), slower startup, higher memory usage, and adds complexity without benefit over Wails which uses the platform's native WebView. This is not negotiable.

---

## Direct Save Game Import: No CSV Intermediary

**Decision**: The core import flow reads directly from the SMB save game database. CSV files are not required for any core app functionality.

**Rationale**:
- The original two-app pipeline (SMB3Explorer exports 8 CSV files → user imports 8 CSV files into Companion) was the most significant UX failure of the original apps: confusing, error-prone, and entirely unnecessary given direct SQLite access
- Issue #36 of SmbExplorerCompanion ("Import current season data directly from game db") was always the right direction; this app implements it as the default
- CSV export remains available as a secondary, opt-in feature for power users, external analysis, or community sharing — it is not the pipeline

**Implication**: The `service/import.go` reads from `SaveGameReader` (the SMB save game DB) and writes directly to the franchise companion DB. There is no CSV parsing step in the normal sync flow.

---

## Save Game Snapshots: Every Sync Persisted Permanently

**Decision**: Every time the app syncs from an SMB save file and the content differs from the last snapshot (determined by SHA-256 hash), the full decompressed SQLite database is persisted as a permanent snapshot.

**Rationale**:
- The SMB game engine compacts franchise data after each season. Data that exists in the save file at the end of a season may be partially or fully gone after the offseason is simulated. Once lost from the save file, it cannot be recovered from the game.
- If our companion schema misses data we later want, or if an import bug corrupts data, the snapshot is the only recovery path.
- Snapshots are the canonical source of truth; the companion DB is a derived view over them.
- Deduplication by hash prevents storing redundant copies when the save file hasn't changed.
- Older snapshots are compressed (zstd) to reduce storage; recent snapshots are kept uncompressed for fast access.

**See**: `snapshot-strategy.md` for the full design.

---

## One SQLite Database Per Franchise

**Decision**: Each franchise has its own isolated SQLite database file. There is no shared multitenant companion database.

**Rationale**:
- The original SmbExplorerCompanion used a single database with a `Franchises` table and `franchise_id` FKs on every entity — a classic multitenant antipattern that added complexity to every query and every schema table
- Per-franchise DBs mean: no `WHERE franchise_id = ?` anywhere, no risk of cross-franchise data leakage, simpler schema, trivial backup/export of a single franchise (copy one file), and complete isolation between franchises
- Migrations run independently per franchise DB, which is actually simpler than coordinating a shared schema

**Structure**:
```
{app_data}/
  registry.db               # Franchise list + metadata only (no baseball data)
  franchises/
    {franchise_id}/
      companion.db          # This franchise's full schema
      snapshots/            # Save game snapshots for this franchise
```

**Implication**: Switching franchises in the UI = closing one `*sql.DB` and opening another. The store layer is re-initialized with the new connection. This is a trivial operation.

---

## Original Schema: Not Preserved

**Decision**: The new companion database schema is designed from scratch. The SmbExplorerCompanion EF Core schema is a reference, not a constraint.

**Rationale**:
- The original schema had shortfalls: computed stats stored redundantly alongside raw counts (two sources of truth), schema choices driven by EF Core conventions rather than SQLite idioms, and areas where denormalization was done for query convenience but created consistency issues
- A schema analysis (comparing the old schema against the actual operations the companion performed) will precede schema design. See `data-layer.md` for principles.

**Non-negotiable**: The new app must provide a migration path from an existing `SmbExplorerCompanion.db` to the new schema. Existing users cannot be left behind. See `data-layer.md`.
