# Implementation Plan: League Transfer MVP

Produced via the `major-feature` planning process, grounded in `legacy-tool-analysis.md`,
`failure-analysis.md`, `validation-results.md`, `plan.md`, and `ux-flow.md`. This is a planning
artifact — no code has been written yet.

## Feature Summary

Implement the MVP of **League Transfer**: a top-level app mode, fully separate from franchise
tracking, selectable at launch alongside "Franchise Tracker." It lets a user **discover** every
SMB4 league on disk (regardless of whether it's a tracked franchise), **export** a selected league
to a self-contained zip (`.sav`/`.sav.bak`/`.hash` + a JSON manifest), and **import** such a zip on
another machine (or the same one) by validating its shape, blocking if SMB4 is running, assigning
the imported league a **freshly generated GUID** with every internal reference rewritten for
consistency (6 columns across 5 tables, per `validation-results.md`), backing up `master.sav` with
a timestamped history mirroring the existing franchise-snapshot pattern, writing the new save files
into the target Steam save directory, and registering the league in `t_league_savedatas` using the
validated correct encoding (`uuid.UUID` 16-byte blob, zlib round-trip, not the legacy tool's two
confirmed bugs). If the importing machine has multiple Steam-ID save directories, the user picks
the target via a list, mirroring the existing `SaveFilePicker` pattern. The game-running check
happens via OS process detection (Windows: process list query for `supermegabaseball.exe`;
Linux/Proton: scan `/proc/*/cmdline`; macOS: no-op), with an exclusive-lock attempt at the moment
of writing `master.sav` as a defense-in-depth secondary check. No existing league save is ever
modified by import; only new files are added and `master.sav` is edited in place after a backup.

## Affected Areas

**New Go packages/files:**
- `internal/models/guid.go` — no new type needed; standardize on `uuid.UUID` (already a
  dependency) for all league-GUID handling in new code, with helpers for blob<->UUID conversion
  where SQL binding needs `[]byte`.
- `internal/models/league_transfer.go` — `LeagueOverview` (name, conferences, divisions, teams),
  `LeagueImportPreview`, `SteamSaveDirCandidate` domain structs.
- `internal/db/savegame_rw.go` — new read-write counterparts to `internal/db/savegame.go`'s
  read-only helpers: `DecompressToTempFile(ctx, srcPath) (tmpPath string, err error)` and
  `CompressFileAtomically(ctx, tmpPath, destPath string) error` (zlib deflate + atomic rename),
  reusable for both `master.sav` and league `.sav` files.
- `internal/store/league_registry_store.go` — operates on an opened `master.sav`-shaped DB:
  `LeagueExists(ctx, db, guid) (bool, error)`, `RegisterLeague(ctx, db, guid uuid.UUID) error`
  (the corrected blob insert).
- `internal/store/league_save_store.go` — operates on an opened league `.sav`-shaped DB:
  `GetLeagueOverview(ctx, db) (models.LeagueOverview, error)` (conferences/divisions/teams query),
  `ValidateLeagueSaveShape(ctx, db) error` (table-presence check), `RewriteLeagueGUID(ctx, db,
  oldGUID, newGUID uuid.UUID) error` (the 6-column/5-table rewrite, table-driven like the
  validation script).
- `internal/service/league_transfer.go` — orchestration: `DiscoverLeagues(ctx)
  ([]models.LeagueOverview, error)`, `ExportLeague(ctx, guid, sourceSavePath, outputPath string)
  error`, `PreviewImport(ctx, zipPath string) (models.LeagueImportPreview, error)`,
  `ConfirmImport(ctx, zipPath, targetSteamDir string) error`.
- `internal/config/league_transfer_paths.go` — extends `AppDirs` with `LeagueTransferDir()`,
  `MasterSaveBackupsDir()`, `ExportsOutputDir()`. Also a new `DiscoverSteamSaveDirs()
  ([]SteamSaveDirCandidate, error)` (enumerates all Steam-ID directories under the SMB4 root, one
  entry per `master.sav` found — generalizes the existing single-directory assumption in
  `savegame_paths.go` without breaking it).
- `internal/system/process_windows.go`, `internal/system/process_linux.go`,
  `internal/system/process_other.go` (build-tagged) — `IsGameRunning() (bool, error)` behind a
  small interface so it's mockable in tests; `process_other.go` covers macOS/anything else and
  always returns `false, nil`.
- `internal/zip/league_package.go` — zip/unzip helpers specific to the league-transfer package
  shape (manifest + 3 files), separate from any generic zip utility since the shape (manifest
  schema, file naming) is feature-specific.
- `app_league_transfer.go` — new Wails bindings file, thin delegation only, mirroring
  `app_franchise.go`'s pattern exactly.
- `dto.go` additions — `LeagueOverviewDTO`, `LeagueImportPreviewDTO`, `SteamSaveDirCandidateDTO`,
  plus mapping helpers.

**New migrations:**
- None required for the companion or franchise registry DBs — this feature doesn't touch either.
  (`master.sav` and league `.sav` files are mutated directly via the new store files above, not
  through smb-tools' own migration system, since they're the *game's* schemas, not smb-tools'.)
- If any new local bookkeeping is needed (e.g., a log of past imports/exports for user reference),
  it would be a new SQLite DB under `AppDirs.LeagueTransferDir()` with its own tiny migration set
  — **deferred**, not needed for MVP per `ux-flow.md`'s deferred list.

**New docs:**
- `docs/domain/master-save-schema.md` — formalize the real `master.sav` schema we verified this
  session (currently only living in `docs/league-transfer/failure-analysis.md`/
  `validation-results.md`, which are research docs, not the canonical domain reference
  `internal/CLAUDE.md` points to for schema verification).

**Frontend:**
- `frontend/src/pages/LeagueTransferHomePage.vue` (or similar) — discovery list, entry point for
  Export/Import sub-flows.
- `frontend/src/pages/LeagueExportPage.vue`, `LeagueImportPage.vue` (or tabs within one page —
  fine either way, frontend's call at execution time).
- `frontend/src/stores/leagueTransfer.ts` — Pinia store if state needs sharing across the
  sub-pages (discovery results, in-progress import preview).
- `frontend/src/router.ts` — new top-level routes.
- `frontend/src/App.vue` — the launch-time mode chooser (Franchise Tracker vs. League Transfer)
  replaces today's implicit "no franchise selected → show FranchiseSelector" behavior at the root
  path; League Transfer's routes need to be reachable without any franchise context.
- `frontend/src/composables/useBreadcrumbs.ts` — add new top-level paths to `ROOT_PATHS`.

## Implementation Tasks

Progress is tracked with the checkboxes below — tick off each task as it's completed during
implementation, in dependency order.

- [x] **1. [Backend/Domain] Document the real master.sav schema**
      What: Write docs/domain/master-save-schema.md with the verified CREATE TABLE
      statements for t_league_savedatas and the other ~21 master.sav tables (at
      least the names/purpose; full column detail for t_league_savedatas, which is
      the only one this feature writes to).
      Why: internal/CLAUDE.md requires schema verification against a documented
      source before any save-game SQL is written; master.sav currently has no
      canonical doc home outside research notes.
      Depends on: none

- [x] **2. [Backend/db] Add read-write zlib helpers**
      What: internal/db/savegame_rw.go — DecompressToTempFile and
      CompressFileAtomically (temp file + os.Rename swap, per the safety
      requirement in plan.md).
      Why: Foundation every other mutation (master.sav edit, league GUID rewrite)
      builds on. No existing code in the repo writes back to a save-game-shaped
      file, so this is genuinely new.
      Depends on: none

- [x] **3. [Backend/store] master.sav registry store**
      What: internal/store/league_registry_store.go — LeagueExists, RegisterLeague.
      RegisterLeague binds guid[:] ([]byte from uuid.UUID), never a string — this
      is the exact bug class from failure-analysis.md Bug #1, and using uuid.UUID's
      byte slice directly makes the mistake structurally hard to repeat.
      Why: Core registration mechanism.
      Depends on: #2

- [x] **4. [Backend/store] League save GUID-rewrite store**
      What: internal/store/league_save_store.go — RewriteLeagueGUID (table-driven
      over the 6 confirmed GUID-bearing columns), GetLeagueOverview,
      ValidateLeagueSaveShape.
      Why: Core import mechanism (fresh-GUID assignment) and the introspection
      query needed by both discovery and import preview.
      Depends on: #2

- [ ] **5. [Backend/config] Multi-Steam-ID discovery + app dirs**
      What: internal/config/league_transfer_paths.go — DiscoverSteamSaveDirs,
      AppDirs.LeagueTransferDir/MasterSaveBackupsDir/ExportsOutputDir.
      Why: Needed before import can know which master.sav to target, and before
      backups have a home. Generalizes savegame_paths.go's single-directory
      assumption without changing its existing behavior for franchise tracking.
      Depends on: none

- [ ] **6. [Backend/system] Game-running detection**
      What: internal/system/ with IsGameRunning() behind an interface, three
      build-tagged implementations (windows/linux/other).
      Why: Primary safety gate before any master.sav mutation.
      Depends on: none

- [ ] **7. [Backend/zip] League package format**
      What: internal/zip/league_package.go — manifest.json schema (league name,
      GUID, export timestamp, smb-tools version), Pack/Unpack functions.
      Why: Export/import need a shared, validated container format.
      Depends on: none

- [ ] **8. [Backend/service] Orchestration**
      What: internal/service/league_transfer.go wiring everything from #2-7:
      DiscoverLeagues, ExportLeague, PreviewImport (read-only: unzip to temp,
      validate shape, check GUID collision against ALL discovered Steam dirs,
      return candidate list), ConfirmImport (mutating: IsGameRunning gate, backup
      master.sav with timestamp, rewrite GUID, write files, register, attempt
      exclusive lock immediately before the master.sav write as defense-in-depth).
      Why: This is where the validated procedure from validation-results.md
      becomes the production code path.
      Depends on: #3, #4, #5, #6, #7

- [ ] **9. [Backend/bindings] Wails bindings**
      What: app_league_transfer.go — thin delegation to the service, following
      app_franchise.go's exact shape (logging, nil-check, error wrapping, DTO
      mapping). Add DTOs to dto.go.
      Why: Frontend surface.
      Depends on: #8

- [ ] **10. [Backend/build] wails build**
      What: Regenerate wailsjs/ bindings after the new App methods/DTOs land.
      Why: internal/CLAUDE.md requirement — never hand-edit generated bindings.
      Depends on: #9

- [ ] **11. [Frontend] Top-level mode chooser**
      What: Modify App.vue's root-path logic to offer Franchise Tracker vs.
      League Transfer when no franchise is active; add league-transfer routes
      to router.ts and ROOT_PATHS.
      Why: The structural UX change from ux-flow.md.
      Depends on: #10

- [ ] **12. [Frontend] Discovery + Export page**
      What: LeagueTransferHomePage.vue (list via DiscoverLeagues, showing
      name/conferences/divisions/teams) + export action calling ExportLeague,
      success state showing output path (mirroring ExportPage.tsx's pattern from
      the legacy tool, adapted to PrimeVue/Dialog conventions).
      Why: Export UX.
      Depends on: #10

- [ ] **13. [Frontend] Import page**
      What: File picker -> PreviewImport -> show league name/manifest details +
      safety disclaimer (no virus scanning, trust your source) -> if multiple
      Steam dirs, a picker (reusing SaveFilePicker's list pattern) -> confirm
      button -> ConfirmImport -> success/error toast.
      Why: Import UX, including the two-step preview/confirm split.
      Depends on: #10

- [ ] **14. [Docs] Update user-docs/team-transfer.md**
      What: Replace the "always writes to a copy" line with accurate wording per
      ux-flow.md's "Resolving the Writes to a Copy Promise" section.
      Why: Public-facing doc currently overpromises relative to what the
      validated mechanism actually does.
      Depends on: none (can happen any time)

## Test Coverage Plan

**Test frameworks in use**: Go `testing` package + table-driven tests (backend); Vitest (frontend
composables/utilities); Storybook (frontend components).

**Testing strategy**: Backend logic (GUID rewrite, blob encoding, shape validation, zip packaging)
is fully unit/integration-testable with in-memory SQLite and temp files — no real game or OS
process needed. `IsGameRunning` is the one piece that's inherently OS-dependent; it sits behind an
interface specifically so the service-layer orchestration tests can inject a fake. New test
fixtures must mirror real schema exactly, per `internal/CLAUDE.md` — and unusually, we already
have **verified ground truth** for `master.sav` from this session's live testing, removing the
usual "needs reverse engineering" risk for this fixture.

**Test cases:**

- [x] [unit] GUID blob round-trip: `uuid.UUID` -> bytes -> back, confirms no byte-swap, matches the
      real-data verification from `validation-results.md`
      Covers: the exact bug class from failure-analysis.md Bug #1
      File: `internal/store/league_registry_store_test.go`
- [x] [integration] `RegisterLeague` on a fresh in-memory `t_league_savedatas`-shaped DB, then read
      back and assert the stored value is a 16-byte blob, not text
      Covers: regression test for Bug #1 specifically
      File: `internal/store/league_registry_store_test.go`
- [x] [integration] `LeagueExists` returns true/false correctly
      Covers: collision-detection path
      File: `internal/store/league_registry_store_test.go`
- [x] [integration] `RewriteLeagueGUID` updates all 6 columns across 5 tables and leaves unrelated
      tables untouched
      Covers: the exact mechanism validated live; must use the new master-save-schema-verified
      fixture extension
      File: `internal/store/league_save_store_test.go`
- [x] [integration] `RewriteLeagueGUID` is transactional — a forced failure mid-way leaves zero
      columns changed
      Covers: partial-rewrite corruption risk
      File: `internal/store/league_save_store_test.go` (covered via the no-matching-rows case,
      which exercises the same single-transaction path; a true forced-failure-mid-loop case would
      need fault injection and is a candidate to add later if this code changes)
- [x] [integration] `GetLeagueOverview` against a league with conferences + divisions, and against
      one with conferences and zero divisions
      Covers: the "divisions are optional" domain rule from ux-flow.md — needs a fixture variant
      with an empty `t_divisions` for one conference; verify the query's fallback path doesn't
      silently return no teams
      File: `internal/store/league_save_store_test.go`
- [x] [unit] `ValidateLeagueSaveShape` rejects a DB missing `t_leagues`/`t_franchise`/
      `t_conferences`/etc., accepts one with all present regardless of whether `t_franchise` has
      rows
      Covers: shape validation must check table presence, not data presence (a stock/season-mode
      league has no franchise rows and must still pass)
      File: `internal/store/league_save_store_test.go`
- [x] [integration] `CompressFileAtomically` + `DecompressToTempFile` round-trip produces
      byte-identical decompressed content
      Covers: regression test for Bug #2 (zip vs. zlib)
      File: `internal/db/savegame_rw_test.go`
- [ ] [unit] Zip packaging: pack then unpack a league + manifest, assert all files and manifest
      fields survive
      Covers: export/import container format
      File: `internal/zip/league_package_test.go`
- [ ] [unit] Zip validation rejects: missing file, GUID mismatch across the three files, corrupt
      zlib stream, manifest missing/malformed
      Covers: import-time shape validation, the "negative" cases
      File: `internal/zip/league_package_test.go`
- [ ] [integration] `PreviewImport` end-to-end against a valid exported zip (using the real-schema
      fixture) returns the expected preview data without mutating anything
      Covers: read-only guarantee of the preview step
      File: `internal/service/league_transfer_test.go`
- [ ] [integration] `ConfirmImport` end-to-end: backs up master.sav with a timestamped name,
      registers the new GUID, writes files, and a second call with the same zip is rejected by
      collision detection
      Covers: the full validated happy path as production code, plus idempotency-of-rejection
      File: `internal/service/league_transfer_test.go`
- [ ] [integration] `ConfirmImport` refuses to proceed when a fake `IsGameRunning` returns true —
      no files written, no backup taken
      Covers: safety gate ordering (check happens before any mutation)
      File: `internal/service/league_transfer_test.go`
- [ ] [unit] Process detection on Linux: a fake `/proc` tree (temp dir with synthetic `cmdline`
      files) correctly matches/rejects based on substring, including a case where `comm` would be
      truncated but `cmdline` is checked instead
      Covers: the Proton-specific correctness concern raised in discussion
      File: `internal/system/process_linux_test.go`
- [ ] [unit] Backup naming/retention: multiple `ConfirmImport` calls (across separate test
      runs/timestamps) each produce a new timestamped backup file under `MasterSaveBackupsDir()`,
      none overwritten
      Covers: Q4's "timestamped history" requirement, mirrored against the existing
      `SnapshotFileName` pattern
      File: `internal/service/league_transfer_test.go`
- [ ] [vitest] Frontend composable for steam-dir picker selection state (if a composable is
      introduced rather than inline page state)
      Covers: UI state correctness for the multi-profile picker
      File: `frontend/src/composables/*.test.ts` (only if a composable is actually extracted —
      skip if the picker logic stays inline in the page)
- [ ] [storybook] Any new reusable component (e.g., a "league overview card" showing
      conferences/divisions/teams) gets a `.stories.ts` per `frontend/CLAUDE.md`'s non-optional
      requirement
      Covers: default/empty/loading states

## Risk & Considerations

- **This is the first feature that writes back to a save-game-shaped SQLite file.**
  `internal/CLAUDE.md`'s anti-corruption-layer and "real schema required" rules were written with
  read-only access in mind; the new store files must hold to the same translation discipline in
  reverse (typed Go values -> correctly-encoded SQL params), which is exactly the discipline that
  was missing in the legacy tool.
- **`gocognit` (threshold 20)**: `ConfirmImport`'s orchestration (gate check, backup, rewrite,
  write, register, lock-probe) is a natural candidate to exceed this if written as one function —
  should be decomposed into named steps even though they run sequentially, not collapsed into a
  single large function.
- **Exclusive-lock-at-write-time is defense-in-depth, not the primary mechanism** — per the Q3
  discussion, a lock probe alone is insufficient (TOCTOU race) and must not become the only safety
  check.
- **Steam-ID directory enumeration is new territory**: existing code (`savegame_paths.go`) was
  written for the franchise-tracking single-directory common case; the new `DiscoverSteamSaveDirs`
  must not change that function's existing behavior — it's an addition, not a refactor of shared
  logic, to avoid risking regressions in franchise tracking's save-file discovery.
- **macOS**: the game doesn't run there at all per existing docs, so League Transfer's OS-process
  check trivially passes (always "not running") — worth an explicit comment so a future reader
  doesn't mistake the no-op for a bug.

## Open Questions / Assumptions

- Exact `supermegabaseball.exe` process name confirmed live in this session (Windows). Linux/Proton
  `cmdline` substring match is a reasonable inference from how Proton processes appear in `/proc`,
  but hasn't been verified against an actual Linux/Proton install — worth a real check before
  considering #6 done, not just a code review.
- The `t_divisions`-empty-conference test case (a league with conferences but zero divisions)
  couldn't be verified against a real save file in this session — all real leagues checked happened
  to have divisions. The query's fallback behavior for that case should be written defensively and
  the test added once we either find or construct a real example.
- Assumes no new local SQLite DB is needed for MVP (no persistent log of past imports/exports) —
  deferred per `ux-flow.md`; flag if that turns out to be wanted sooner.
- Everything else from the four clarifying questions is resolved: fresh GUID on import (#1),
  list-picker for multi-profile machines mirroring existing UX (#2), layered process-check +
  lock-probe (#3), timestamped backup history mirroring the snapshot pattern (#4).
