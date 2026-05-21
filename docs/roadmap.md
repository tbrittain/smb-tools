# Feature Roadmap

High-level phases for the smb-tools rewrite. Each phase should be largely complete before the next begins — later phases depend on the foundations laid earlier. Phases are not sprint boundaries; they are logical groupings of work.

Check off phases as they are completed. Individual phases will be broken into detailed task lists as implementation approaches.

---

## Phase 1 — Foundation
*Re-scaffold the app, establish the full development environment, wire up CI and testing infrastructure. No features — just a solid base to build on.*

- [x] Re-scaffold Wails v2.12.0 with Vue 3 + TypeScript frontend
- [x] Configure full frontend toolchain (Vite 8, PrimeVue, AG Grid, ECharts, Pinia, Biome, Storybook)
- [x] Establish Go backend package structure (store, service, models, config, db)
- [x] Set up SQL-file-based schema migrations with initial registry + companion schemas
- [x] Wire up modernc.org/sqlite for both DB connections
- [x] Set up CI pipeline (go test, go vet, golangci-lint, vitest)
- [x] Establish testutil package (in-memory SQLite, seed helpers)
- [x] App data directory management and franchise registry skeleton

---

## Phase 2 — Save Game Layer
*Everything needed to read an SMB save file. No UI yet — just the Go layer that understands the save game format.*

- [x] ZLib decompression of .sav files
- [x] Read-only SQLite connection via SaveGameReader interface
- [x] Save game schema coverage for the data the original companion app imported: players, teams, schedules, batting stats, pitching stats (scope is the original companion's import surface, not the full save game schema)
- [x] Auto-discovery of save file locations (default paths for SMB3/SMB4, master.sqlite league registry)
- [x] Snapshot persistence: SHA-256 dedup, zstd compression of older snapshots
- [x] Snapshot metadata tracked in companion DB

---

## Phase 3 — Franchise Management
*Create, switch between, and manage franchises. The per-franchise DB architecture in practice.*

- [x] Franchise registry (registry.db): create, list, rename, delete franchises
- [x] Per-franchise companion DB creation with migrations
- [x] Associate a franchise with an SMB save file + league GUID
- [x] Franchise switching (close current DB, open selected)
- [x] Basic franchise management UI (create new, select existing, last-used persistence)
- [x] Save file auto-discovery: scan default paths, probe each .sav for league/franchise metadata (name, player team, season count) so user can identify the right file without a file browser
- [x] Per-franchise save file re-indexing (ProbeFranchiseSaveFile) for live game state in the franchise list
- [x] **Franchise fork / multi-source support**: SMB4 lets users export a franchise snapshot to a new league, which creates a new `leagueGUID` and resets in-game season numbers to 1. The app treats this as a continuation of the same franchise. `franchise_sources` table in `registry.db` (id, franchise_id, save_file_path, league_guid, season_offset, added_at). `season_offset` is added to the save game's season number to produce the display season number, stored at import time. `SyncSeason` reads from the highest-offset source. `seasons` table uses `league_guid + save_game_season_id` as the uniqueness key (preventing PK collision across forks). Player/team GUIDs that change on fork are resolved via `player_alt_guids`/`team_alt_guids` tables with name+handedness+chemistry fuzzy matching as fallback. Dashboard exposes "Replace file" (correction) and "Add fork source" (fork continuation) as two distinct explicit actions.

---

## Phase 4 — Season Sync
*The core pipeline: one action that reads a save game and persists the season to the companion DB. Replaces the entire SMB3Explorer → CSV → Companion import flow.*

- [x] Full player import (attributes, traits, salary, handedness, chemistry, pitch types)
- [x] Full team import (standings, budget, payroll, aggregate attributes)
- [x] Regular season schedule and game results
- [x] Playoff schedule and results
- [x] Career stats for all active players
- [ ] Franchise news events (skill changes, trait changes, trades, retirements — from `t_franchise_news_*` tables; requires new SaveGameReader methods and companion schema columns beyond the original companion app's scope; deferred until companion schema is finalized)
- [ ] Team logo extraction (binary blob storage + rendering; deferred until Phase 5/6 when the UI that displays logos exists)
- [x] Sync UI: trigger button, last-synced indicator, progress feedback
- [x] Season auto-detection: SyncSeason reads the most recent season from the save game; user no longer needs to supply internal season IDs
- [ ] Championship winner detection (post-import query over completed playoff data; deferred to Phase 5 where leaderboard queries will also be written)

---

## Phase 5 — Core Stats Viewer
*The Baseball Reference-style UI. The central reason this app exists.*

- [x] Home/dashboard screen (franchise summary, recent champion, award leaders, standings snapshot)
- [x] Global search (players and teams)
- [x] Player overview page (career stats, season-by-season breakdown, attributes, awards, traits)
- [x] Team overview page (all-time profile, name and logo history)
- [x] Team season detail page (roster with stats, budget/payroll, schedule breakdown, playoff results)
- [x] Historical teams list

---

## Phase 6 — Leaderboards
*Franchise-wide statistical leaderboards — the most-used views for most users.*

- [x] Top batting careers
- [x] Top batting seasons
- [x] Top pitching careers
- [x] Top pitching seasons
- [x] Filters: position, chemistry type, handedness, Hall of Famers only
- [x] Season range selector
- [x] Regular season / playoffs toggle
- [x] Sorting by any stat column
- [x] Pagination

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

---

> **Phases 1–9 constitute MVP.** Phases 10 and beyond are post-MVP enhancements.

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
- [ ] Stat cell highlights in tables (Baseball Reference–style):
  - **Bold** individual stat values when that player led the league in that category for the season
  - **Gold cell background** when a value is a franchise all-time career record
  - Legend displayed near the table explaining both indicators
  - Applies to season stat tables (PlayerStatTable, roster tables); career totals rows are candidates for the gold record highlight

---

## Phase 13 — Game Integration (Optional / Windows-Only)
*Enhancement layer. Does not block any other phase. Core app is fully functional without this.*

- [ ] fsnotify-based auto-sync: detect save file write → trigger sync automatically (eliminating manual sync trigger)
- [ ] Investigate BepInEx feasibility further (blocked on community RE work; see `docs/game-integration/investigation.md`)
