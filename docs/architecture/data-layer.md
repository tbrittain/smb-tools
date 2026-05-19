# Data Layer

SQLite strategy, the two-database architecture, schema design principles, migration tooling, and the legacy data migration requirement.

---

## One Database Per Franchise

The original SmbExplorerCompanion used a single SQLite database with a `Franchises` table — essentially a multitenant database where every table had a `franchise_id` foreign key. This was a design flaw:

- Every query required a `WHERE franchise_id = ?` predicate — easy to forget, a potential source of data leakage between franchises
- The schema was more complex than necessary; every entity carried a FK that existed purely for tenant isolation
- Backing up, sharing, or exporting a single franchise required extracting a subset of the database rather than just copying a file

**The new design: one SQLite file per franchise.**

Each franchise gets its own isolated `companion.db`. There is no `franchise_id` column anywhere in the companion schema. A query against a franchise DB implicitly operates on that franchise's data only — there is no other data present.

A lightweight **franchise registry** (a small separate SQLite file or JSON manifest in the app data root) stores the list of franchises and their metadata (name, ID, game version, last synced, DB file path). The registry is the only place franchise identity is tracked across the app.

### App Data Directory Structure

```
{app_data}/
  registry.db                       # Franchise list + metadata only
  franchises/
    {franchise_id}/                 # UUID or stable short ID
      companion.db                  # This franchise's companion DB (full schema)
      snapshots/
        {season}_{hash12}.sqlite    # Decompressed save game snapshots
        {season}_{hash12}.sqlite.zst  # Compressed older snapshots
```

### Implications

- **Migrations run per franchise**: when the app opens a franchise, it runs `golang-migrate` against that franchise's `companion.db`. All franchise DBs are migrated independently.
- **Switching franchises** = closing one `*sql.DB` and opening another. The store layer is re-initialized with the new connection. No query changes needed.
- **The `SaveGameReader` connection** is always separate from the companion DB connection, regardless of which franchise is active.
- **The registry** does not hold any baseball data — only the minimum needed to list and select franchises (name, ID, game version, path to DB, last sync time, snapshot count/size summary).

---

## Two Database Connections, Two Lifetimes

The app works with two SQLite databases simultaneously, and they have fundamentally different characteristics:

| | SMB Save Game DB | Companion DB |
|--|--|--|
| Owned by | Metalhead Software (the game) | This app |
| Access mode | Read-only | Read/write |
| Schema | Fixed, not ours to change | Designed and versioned by us |
| Lifetime | Opened on demand, closed when user unloads | Open for the entire app session |
| Connection | Decompressed `.sav` → temp `.sqlite` file | Persistent file in OS app data directory |
| Migrations | Never | Run automatically at startup |
| Connection var | `saveDB *sql.DB` | `companionDB *sql.DB` |

These are **two separate `*sql.DB` values** threaded explicitly through the code. They are never aliased, never mixed. The store layer makes it structurally impossible to accidentally run a companion write query against the save game DB or vice versa — `SaveGameReader` only has read methods; `PlayerStore` etc. only hold the companion `*sql.DB`.

---

## Save Game Database: Lifecycle

```
User triggers "Sync"
        ↓
db/savegame.go: zlib inflate → decompressed bytes in memory
        ↓
SHA-256 hash of decompressed bytes
        ↓
Compare hash to most recent snapshot for this franchise (from save_game_snapshots table)
        ↓
If hash differs: persist snapshot to franchises/{id}/snapshots/{season}_{hash12}.sqlite
If hash matches: skip snapshot write, continue to read
        ↓
Write decompressed bytes to temp file in OS temp dir
        ↓
sql.Open("sqlite", tempFilePath + "?mode=ro")  ← read-only URI param
        ↓
SqliteSaveGameReader wraps the connection
        ↓
Read season data (players, teams, schedule, stats)
        ↓
Write to franchise companion DB via store layer
        ↓
reader.Close() → temp file deleted
```

The snapshot is persisted before the temp file is opened for reading. If the sync fails partway through, the snapshot still exists and the raw data is preserved. The temp file is always cleaned up on close, whether the sync succeeded or failed.

**The original `.sav` file is never modified. Ever.**

See `snapshot-strategy.md` for the full snapshot lifecycle, compression policy, and storage management.

---

## Companion Database: Lifecycle

```
App startup
        ↓
db/companion.go: resolve path via config/app_directories.go
        ↓
sql.Open("sqlite", companionDBPath)
        ↓
golang-migrate: run any pending up migrations
        ↓
App is ready — companion DB connection lives for the session
        ↓
App shutdown → connection closed
```

---

## golang-migrate Setup

Migration files live in `backend/db/migrations/` as plain SQL:

```
001_initial_schema.up.sql
001_initial_schema.down.sql
002_add_player_uniform_numbers.up.sql
002_add_player_uniform_numbers.down.sql
```

Each migration is a pair: `up` applies the change, `down` reverses it. golang-migrate tracks the current version in a `schema_migrations` table it manages itself.

Migrations run automatically at app startup before any store is used. If a migration fails, the app surfaces the error and exits cleanly rather than running against a partially-migrated schema.

---

## Schema Design Principles

The new companion schema is designed from scratch. These principles govern its design:

### 1. Be deliberate about which derived stats to persist

The original companion stored both raw counting stats (AB, H, HR, etc.) *and* pre-computed rate stats as nullable doubles. The problem was not that it stored derived stats — it's that they were stored as independent columns that could silently diverge from the raw counts that produced them.

**Simple rate stats** (BA = H/AB, OBP, SLG, OPS, WHIP, K/9, BB/9, etc.) are deterministic functions of a single row's own columns. Use SQLite [generated columns](https://www.sqlite.org/gencol.html) for these — they are computed by SQLite itself and are always in sync. Never store them as independent writable columns.

**Context-dependent stats** (wOBA, FIP, ERA+, OPS+, smbWAR) require league-wide context — linear weights, league ERA constants, park factors — that is only available at sync time. These must be computed during the import pipeline and persisted alongside the raw counts. Store the league constant or weights used alongside them so the derivation is auditable.

**The rule**: never store a derived stat as an independent writable column where it could diverge from its inputs. Generated columns for simple rates. Computed-then-stored (with audit trail) for complex rates.

### 2. Favor SQLite idioms over ORM convenience

The original schema was shaped by EF Core conventions (auto-increment PKs everywhere, navigation properties, many-to-many junction tables with EF-generated names). Design the new schema for SQLite directly: appropriate use of `INTEGER PRIMARY KEY` (rowid alias), explicit junction table names, and indexes chosen for the actual query patterns this app runs.

### 3. Normalize aggressively, denormalize deliberately

Start fully normalized. Add denormalization only when a specific query has a measured performance problem and the denormalization is explicitly documented with the reason.

### 4. Schema analysis precedes schema design

Before writing a single migration file, conduct a structured analysis:
- Review every repository query in SmbExplorerCompanion (see source files in `docs/smb-explorer-companion/companion-db-schema.md`)
- Identify which stored columns were never queried
- Identify which queries were doing expensive joins that a schema change could simplify
- Identify the `IsRegularSeason` boolean flag pattern — examine whether a separate table per stat type (regular season vs playoffs) is cleaner

This analysis will be documented in `docs/architecture/schema-analysis.md` before implementation begins.

### 5. sqlc is under evaluation

`sqlc` generates type-safe Go from SQL query files. It pairs naturally with golang-migrate (both work with plain `.sql` files) and keeps SQL out of Go strings while producing idiomatic, readable code. It will be evaluated during the initial scaffolding phase. Decision and rationale will be added to `decisions.md` once made.

---

## Legacy Migration: SmbExplorerCompanion.db → New Schema

**This is a first-class feature, not an afterthought.**

Existing users of SmbExplorerCompanion have years of franchise history in their `SmbExplorerCompanion.db` files. The new app must provide a path to bring that data forward.

### Approach

A `service/legacy_migration.go` implements this as a regular Go service:

```go
type LegacyMigrationService struct {
    legacyDB    *sql.DB           // read-only connection to the old SmbExplorerCompanion.db
    players     *store.PlayerStore
    teams       *store.TeamStore
    seasons     *store.SeasonStore
    // ...
}

func (s *LegacyMigrationService) Migrate(ctx context.Context) error {
    // Read from legacyDB in the old schema
    // Transform to new models
    // Write via store layer to new companion DB
}
```

### Requirements

- Opens the old DB read-only — never modifies the original
- Is fully tested: integration tests spin up an in-memory DB seeded with a representative sample of the old schema, run the migration, and assert the new schema contains correct records
- Handles partial/incomplete old data gracefully (some users may have incomplete seasons, missing awards, etc.)
- Is idempotent where possible — safe to run more than once without duplicating data
- Surfaced in the UI as an explicit "Import from previous version" flow, not a silent background operation

### Schema Mapping

The detailed mapping from old EF Core entities to new schema tables will be documented in `docs/architecture/schema-analysis.md` once both schemas are defined.
