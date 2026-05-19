# Feature Roadmap

High-level phases for the smb-tools rewrite. Each phase should be largely complete before the next begins — later phases depend on the foundations laid earlier. Phases are not sprint boundaries; they are logical groupings of work.

Check off phases as they are completed. Individual phases will be broken into detailed task lists as implementation approaches.

---

## Phase 1 — Foundation
*Re-scaffold the app, establish the full development environment, wire up CI and testing infrastructure. No features — just a solid base to build on.*

- [ ] Re-scaffold Wails v2.12.0 with Vue 3 + TypeScript frontend
- [ ] Configure full frontend toolchain (Vite 8, PrimeVue, AG Grid, ECharts, Pinia, Biome)
- [ ] Establish Go backend package structure (store, service, models, config, db)
- [ ] Set up golang-migrate with initial empty companion DB schema
- [ ] Wire up modernc.org/sqlite for both DB connections
- [ ] Set up CI pipeline (go test, go vet, golangci-lint, vitest)
- [ ] Establish testutil package (in-memory SQLite, seed helpers)
- [ ] App data directory management and franchise registry skeleton

---

## Phase 2 — Save Game Layer
*Everything needed to read an SMB save file. No UI yet — just the Go layer that understands the save game format.*

- [ ] ZLib decompression of .sav files
- [ ] Read-only SQLite connection via SaveGameReader interface
- [ ] Full save game schema coverage (all tables documented in `docs/domain/save-game-schema.md` plus newly discovered tables from `league-template.sqlite`)
- [ ] Auto-discovery of save file locations (default paths for SMB3/SMB4, master.sqlite league registry)
- [ ] Filesystem watcher (fsnotify) for save file change detection
- [ ] Snapshot persistence: SHA-256 dedup, zstd compression of older snapshots
- [ ] Snapshot metadata tracked in companion DB

---

## Phase 3 — Franchise Management
*Create, switch between, and manage franchises. The per-franchise DB architecture in practice.*

- [ ] Franchise registry (registry.db): create, list, rename, delete franchises
- [ ] Per-franchise companion DB creation with migrations
- [ ] Associate a franchise with an SMB save file + league GUID
- [ ] Franchise switching (close current DB, open selected)
- [ ] Basic franchise management UI (create new, select existing, last-used persistence)

---

## Phase 4 — Season Sync
*The core pipeline: one action that reads a save game and persists the season to the companion DB. Replaces the entire SMB3Explorer → CSV → Companion import flow.*

- [ ] Full player import (attributes, traits, salary, handedness, chemistry, pitch types)
- [ ] Full team import (standings, budget, payroll, aggregate attributes)
- [ ] Regular season schedule and game results
- [ ] Playoff schedule and results
- [ ] Career stats for all active players
- [ ] Franchise news events (skill changes, trait changes, trades, retirements — from newly discovered `t_franchise_news_*` tables)
- [ ] Team logo extraction from save game DB
- [ ] Sync UI: trigger button, last-synced indicator, progress feedback
- [ ] Championship winner detection and recording

---

## Phase 5 — Core Stats Viewer
*The Baseball Reference-style UI. The central reason this app exists.*

- [ ] Home/dashboard screen (franchise summary, recent champion, award leaders, standings snapshot)
- [ ] Global search (players and teams)
- [ ] Player overview page (career stats, season-by-season breakdown, attributes, awards, traits)
- [ ] Team overview page (all-time profile, name and logo history)
- [ ] Team season detail page (roster with stats, budget/payroll, schedule breakdown, playoff results)
- [ ] Historical teams list

---

## Phase 6 — Leaderboards
*Franchise-wide statistical leaderboards — the most-used views for most users.*

- [ ] Top batting careers
- [ ] Top batting seasons
- [ ] Top pitching careers
- [ ] Top pitching seasons
- [ ] Filters: position, chemistry type, handedness, Hall of Famers only
- [ ] Season range selector
- [ ] Regular season / playoffs toggle
- [ ] Sorting by any stat column
- [ ] Pagination

---

## Phase 7 — Awards & Hall of Fame
*Season-level award tracking and Hall of Fame management.*

- [ ] Manual award assignment (MVP, Cy Young, Gold Glove, Silver Slugger, ROY, All-Star, Playoff/Championship MVP)
- [ ] Runner-up award support (MVP-2 through MVP-5, etc.)
- [ ] Auto-calculated title awards (BA, HR, RBI, ERA, W, K leaders)
- [ ] Auto-calculated Triple Crown (batting and pitching)
- [ ] Hall of Fame eligibility evaluation and induction
- [ ] Custom user-defined awards

---

## Phase 8 — Visualizations
*Charts, plots, and percentile rankings throughout the app.*

- [ ] Player attribute radar/spider chart (percentile visualization)
- [ ] Player KPI percentile rankings
- [ ] Team season performance trend chart (margin of victory over season)
- [ ] Similar players recommendations
- [ ] Franchise-level stat trend charts (era averages, season-over-season)

---

## Phase 9 — Legacy Migration
*One-time migration path for existing SmbExplorerCompanion users. Brings historical data forward into the new schema.*

- [ ] Schema analysis and mapping doc (`docs/architecture/schema-analysis.md`)
- [ ] LegacyMigrationService: reads old SmbExplorerCompanion.db, writes to new franchise DB
- [ ] Migration UI: detect existing companion DB, confirm, run, report results
- [ ] Integration tests covering full-franchise, minimal, and edge-case scenarios

---

## Phase 10 — CSV Export
*Opt-in data export for power users, external analysis, and community sharing. Not required for any core app functionality.*

- [ ] Export any leaderboard view to CSV
- [ ] Export player career stats to CSV
- [ ] Full season export (format compatible with original SMB3Explorer output, for existing tooling/workflows)

---

## Phase 11 — Team Transfer Tool
*The most-requested community feature. Never supported by either original app.*

- [ ] Read full team roster from a source save game (players, attributes, salary, traits, logos)
- [ ] Write team into a target save game (as a new or replacement team)
- [ ] UI: select source franchise/team, select target save, confirm transfer
- [ ] Validation: check for compatibility (SMB4 only, same game version)
- [ ] Safety: always write to a copy — never modify the original save file in place

---

## Phase 12 — Polish & Community Features
*Quality-of-life features and items from the original companion app's open issues backlog. Tackle after core is solid.*

- [ ] Stat scaling per 162 games (context for non-162-game seasons)
- [ ] 30/30 and 40/40 season tracking
- [ ] smbWAR custom metric
- [ ] HoF career standards test (Baseball Reference–style)
- [ ] Player nicknames
- [ ] Team colors
- [ ] Player attribute history visualization (season-over-season progression)
- [ ] League-average trend overlays on stat views
- [ ] Bold league leaders in stat tables

---

## Phase 13 — Game Integration (Optional / Windows-Only)
*Enhancement layer. Does not block any other phase. Core app is fully functional without this.*

- [ ] fsnotify-based auto-sync: detect save file write → trigger sync automatically
- [ ] Investigate BepInEx feasibility further (blocked on community RE work; see `docs/game-integration/investigation.md`)
