# Plan: Issue #171 — Audit qualified player scaling

> Temp planning doc — not part of the permanent docs set. Delete once implemented.

## Context

The career batting/pitching qualification thresholds (used for record-holder and leaderboard
"qualified" filtering) are calibrated off Baseball-Reference's 3000 PA / 1000 IP career minimums,
scaled by season length relative to MLB's 162-game season. That scaling formula is duplicated in
three places and, critically, the same regular-season-scaled threshold is incorrectly applied to
**career playoff** stats too — where playoff PA/IP totals are tiny by comparison, so the scaled
threshold silently returns zero qualifying players. This is flagged by existing TODO comments in
`leaderboard_query.go` and `export_store.go`.

Per Baseball-Reference, career playoff leaders use much lower, **unscaled** minimums: 40 PA or 18
BB+H for batting, 30 IP (90 outs) or 6 decisions for pitching.

Issue #171 also raised scaling thresholds by game length (innings per game), since SMB4 supports
variable-length games. **Initial investigation was wrong**: a follow-up check by decompressing a
real save game snapshot (`internal/db.DecompressAndOpen`-style inspection, via
`PRAGMA table_info`) found that innings-per-game **is** present in the schema —
`t_seasons.innings` (per-season) and `t_franchise_season_creation_params.innings` (franchise
default for new seasons). It's an undocumented column on a table the import pipeline already
queries in several places (`internal/store/sqlite_savegame_reader.go`), not buried in one of the
~20 genuinely unexplored tables. Reading it is therefore in scope for this issue, not deferred.

Decision: career season-length scaling continues to use the franchise's **first season only** (no
averaging/weighting across seasons whose length or innings might change, e.g. across a fork) — this
is documented as a known limitation rather than fixed. Innings-per-game scaling follows the same
first-season-only convention for consistency.

**Legacy migration gap**: the old SmbExplorerCompanion schema has no concept of innings-per-game at
all (`Seasons.NumGamesRegularSeason` exists and is migrated directly, but there is no equivalent
innings column). Since this can't be inferred from legacy data, the legacy migration wizard must
prompt the user for it once per migrated franchise (mirroring the existing per-franchise custom-name
step), defaulting to 9 if the user doesn't know/skips.

## Feature Summary

1. Consolidate the three duplicated "3000 PA / 1000 IP scaled by 162-game season" threshold
   formulas into one shared helper.
2. Fix career playoff qualification to use unscaled MLB postseason minimums (40 PA or 18 BB+H for
   batting; 30 IP/90 outs or 6 decisions for pitching) instead of the misapplied regular-season-
   scaled threshold.
3. Read innings-per-game from the save game (`t_seasons.innings`) during live import, store it on
   the companion `seasons` table, and fold it into the RS-scaled threshold formula alongside games-
   per-season.
4. Prompt the user for innings-per-game once per franchise during legacy migration (no source data
   exists), defaulting to 9.
5. Document the first-season-only scaling limitation (now covering both season length and innings)
   in the franchise-forking docs.

## Affected Areas

- **`internal/db/migrations/companion/`** — new migration adding `innings_per_game INTEGER NOT NULL
  DEFAULT 9` to `seasons`.
- **`internal/store/sqlite_savegame_reader.go`** — read `t_seasons.innings` alongside existing
  season queries (e.g. wherever `GetCurrentSeason`/`GetFranchiseSeasons` already join `t_seasons`);
  add `Innings int` to the relevant `models.SaveGameFranchiseSeason`-style struct.
- **`internal/service/import.go`** — wire the read `Innings` value into the `store.Season{}` upsert
  alongside `NumGames`.
- **`internal/store/season.go`** — add `InningsPerGame int` to `store.Season`; extend `Upsert`'s
  column list/SQL.
- **`internal/store/`** — new shared helper file (e.g. `career_qualification.go`) computing
  RS-scaled batting/pitching thresholds (now factoring in both games-per-season AND innings-per-
  game) plus fixed PO constants. Reuses/extends `GetFranchiseSeasonLength`
  (`internal/store/stat_records.go:344`) with a parallel `GetFranchiseInningsPerGame` (or a single
  query returning both, to avoid two round-trips) — same first-season-only pattern.
- **`internal/service/stat_records.go`** — replace the duplicated formula (lines ~95-98) with the
  shared helper; extend `computeBattingCareerRecords`/`computePitchingCareerRecords` (and the
  rate-stat equivalents) to take an OR-conditioned qualifier for PO records (PA or BB+H; outs or
  decisions) instead of a single threshold int.
- **`internal/store/leaderboard_query.go`** — `GetBattingCareerLeaders`/`GetPitchingCareerLeaders`
  (inline SQL at lines ~153, ~597): remove the duplicated inline SQL threshold subquery and TODOs;
  branch threshold logic on `f.GameType` (regular_season → scaled by games AND innings, playoffs →
  unscaled OR-condition).
- **`internal/store/export_store.go`** — same fix for `career_batting`/`career_pitching` dataset
  cases (lines ~565, ~569), branching on `CareerStatType`.
- **`internal/service/legacy_migration.go`** — accept an `inningsPerGame int` input (default 9),
  apply it to every migrated season's `store.Season{InningsPerGame: ...}` (there's no per-season
  legacy source, so franchise-level granularity is correct here, consistent with the season-length
  first-season-only convention already used for RS scaling).
- **`app_migration.go`** — `MigrateLegacyFranchise` gains an `inningsPerGame int` parameter (or a
  small options struct if that reads better alongside the existing 4 scalar params).
- **`frontend/src/pages/LegacyMigrationPage.vue`** — new wizard step (mirroring the existing
  `'names'` step's per-franchise form pattern) collecting innings-per-game per selected franchise,
  inserted between `'names'` and `'confirm'`; default the input to 9.
- **`user-docs/franchise-forking.md`** — add a Known Limitations note: career qualification
  thresholds (and career stat scaling generally) are calibrated to the franchise's first season's
  games-per-season AND innings-per-game, and don't adapt if either changes after a fork.
- **Tests** — `leaderboard_query_test.go`, `export_store_test.go`, `stat_records_test.go`,
  `legacy_migration_test.go`, `season_test.go`, plus a new test file for the shared helper.

## Implementation Tasks

```
1. [DB] Add innings_per_game column to companion seasons table
   What: New migration file adding `innings_per_game INTEGER NOT NULL DEFAULT 9` to `seasons`.
   Why: Storage for both live-imported and legacy-migrated innings-per-game values.
   Depends on: none

2. [Store] Read t_seasons.innings in the save game reader
   What: Extend the existing season queries in sqlite_savegame_reader.go (wherever t_seasons is
         already joined) to also SELECT `innings`; add the field to the populated struct.
   Why: Source of truth for live-imported franchises; column already exists on a table the
        reader already touches, so this is additive, not a new query shape.
   Depends on: none

3. [Service] Wire innings into the season import write path
   What: In import.go, pass the read Innings value into store.Season{} alongside NumGames when
         calling seasonStore.Upsert. Add InningsPerGame field to store.Season and extend the
         Upsert SQL/column list (season.go).
   Why: Persists per-season innings for live imports.
   Depends on: #1, #2

4. [Store] Add shared career qualification helper
   What: New function (e.g. CareerQualificationThresholds(ctx) (CareerThresholds, error))
         returning: BattingPAThresholdRS (3000 * numGames/162 * innings/9),
         PitchingOutsThresholdRS (same), and fixed constants BattingPAThresholdPO=40,
         BattingBBHThresholdPO=18, PitchingOutsThresholdPO=90 (30 IP),
         PitchingDecisionsThresholdPO=6. Add GetFranchiseInningsPerGame (or extend
         GetFranchiseSeasonLength to return both values in one query) using the same
         first-season-only pattern.
   Why: Single source of truth, replacing 3 duplicated formulas; fixes the playoff bug at the
        root; folds in the innings factor everywhere at once instead of 3x.
   Depends on: #1

5. [Service] Wire stat_records.go to the shared helper
   What: Replace the hardcoded threshold calc; change computeBattingCareerRecords /
         computePitchingCareerRecords (and the *RateRecords variants) to accept an
         OR-conditioned qualifier for PO: batting qualifies if PA>=40 OR (hits+walks)>=18;
         pitching qualifies if outsPitched>=90 OR (wins+losses)>=6.
   Why: Service-layer career record/leader computation respects correct PO minimums and the
        innings-scaled RS threshold.
   Depends on: #4
   Note: verify "hits"/"walks" keys exist in battingStatExtractors and "wins"/"losses" in
         pitchingStatExtractors before wiring — they should, since OBP/decisions already
         require them, but confirm exact key names in code.

6. [Store] Fix GetBattingCareerLeaders / GetPitchingCareerLeaders
   What: Remove the inline SQL threshold subquery and TODO comments. Compute the threshold
         value in Go via the new helper (games AND innings factored in) and bind it as a
         parameter. Branch on f.GameType: "playoffs" → OR-condition unscaled minimums;
         default/regular_season → scaled threshold (existing behavior, de-duplicated, now also
         innings-scaled). "combined" → use the regular-season scaled threshold (see Open
         Questions).
   Why: Fixes the playoff misapplication bug flagged by the existing TODOs; applies innings
        scaling that wasn't previously possible.
   Depends on: #4

7. [Store] Fix export_store.go career_batting/career_pitching extraConds
   What: Same fix, branching on CareerStatType ("regular_season"/"playoffs"/"total_career").
         "total_career" keeps the RS-scaled threshold (see Open Questions).
   Why: Export datasets get the same correction as the leaderboard queries.
   Depends on: #4

8. [Service+Binding] Prompt for innings-per-game during legacy migration
   What: Add `inningsPerGame int` param to MigrateLegacyFranchise (app_migration.go) and to
         LegacyMigrationService.Migrate. Apply it uniformly to every migrated season's
         store.Season{InningsPerGame: inningsPerGame}, since the legacy schema has no per-season
         (or any) source for this value. Default to 9 if caller passes 0/unset.
   Why: Legacy-migrated franchises need a value for the new column; it can't be derived, so the
        user must supply it once per franchise.
   Depends on: #1, #3 (for the store.Season shape)

9. [Frontend] Add innings-per-game step to legacy migration wizard
   What: New wizard step in LegacyMigrationPage.vue between 'names' and 'confirm', mirroring the
         existing per-franchise custom-name form (v-for over selected franchises, one numeric
         input each, default 9). Pass values through to MigrateLegacyFranchise.
   Why: Surfaces the prompt from #8 to the user.
   Depends on: #8

10. [Docs] Update user-docs/franchise-forking.md
    What: Add a "Known Limitations" section noting career qualification thresholds (and career
          stat scaling generally) are calibrated to the first season's games-per-season AND
          innings-per-game, and don't adapt to changes after a fork.
    Why: Sets correct user expectations per decision not to fix multi-season/multi-length
         scaling.
    Depends on: none
```

## Test Coverage Plan

**Test framework**: Go `testing` package, table-driven style, using existing
`internal/testutil` seed helpers (`seedSeason`, `seedPlayer`, `seedPlayerSeason`, `seedBatting`,
etc.) — same patterns already in `leaderboard_query_test.go`.

**Strategy**: Verify the shared helper's scaled-RS math against known season lengths AND innings
combinations, verify PO thresholds are fixed regardless of either, verify the OR-condition
qualifies players who'd previously been wrongly excluded by the scaled-RS formula, and verify the
legacy migration prompt path persists the supplied value.

- [ ] unit: reader populates `Innings` from `t_seasons.innings` for a fixture season — File:
      `sqlite_savegame_reader_test.go`
- [ ] integration: import.go writes `innings_per_game` to the companion `seasons` row matching the
      save game fixture — File: `import_test.go`
- [ ] unit: shared helper computes correct RS threshold for combinations of 40/81/162-game seasons
      × 6/9/12-inning games; returns 0 when no seasons exist — File:
      `internal/store/career_qualification_test.go`
- [ ] unit: shared helper PO constants are season-length/innings-invariant — File:
      `internal/store/career_qualification_test.go`
- [ ] integration: `GetBattingCareerLeaders` with `GameType="playoffs"`, player with 45 career
      playoff PA (below old 740-threshold bug, above new 40-PA threshold) is included — File:
      `leaderboard_query_test.go`
- [ ] integration: same query, player with PA<40 but hits+walks>=18 qualifies via OR — File:
      `leaderboard_query_test.go`
- [ ] integration: same query, player below both PA and BB+H minimums is excluded — File:
      `leaderboard_query_test.go`
- [ ] integration: `GetPitchingCareerLeaders` playoffs — outs>=90 OR decisions>=6 qualifies; below
      both excludes — File: `leaderboard_query_test.go`
- [ ] integration: regular-season threshold with a non-9-inning first season scales correctly
      (e.g. 6-inning games halves the RS threshold relative to 9-inning, holding games constant) —
      File: `leaderboard_query_test.go`
- [ ] regression: existing `TestGetBattingCareerLeaders_QualifiedOnly` /
      `TestGetPitchingCareerLeaders_QualifiedOnly` (regular_season, 40-game/9-inning season,
      threshold=740) still pass unchanged — File: `leaderboard_query_test.go`
- [ ] integration: `export_store.go` preview with `CareerStatType="playoffs"` applies unscaled
      OR-condition — File: `export_store_test.go`
- [ ] regression: existing `TestPreviewExportData_QualifiedOnly_CareerBatting/Pitching`
      (regular_season, 162-game/9-inning) unchanged — File: `export_store_test.go`
- [ ] unit: `computeBattingCareerRecords`/`computePitchingCareerRecords` PO qualifier correctly
      OR-gates on totals map — File: `stat_records_test.go`
- [ ] integration: legacy migration with a supplied `inningsPerGame` value writes it to every
      migrated season's companion row — File: `legacy_migration_test.go`
- [ ] integration: legacy migration with `inningsPerGame=0`/unset defaults to 9 — File:
      `legacy_migration_test.go`

## Risk & Considerations

- Combining "combined" `GameType` and "total_career" `CareerStatType` qualification basis is an
  assumption (defaulting to the RS-scaled threshold) — flagged as an open question below rather
  than guessed silently into final behavior.
- The OR-condition qualifier change to `computeBattingCareerRecords`/`computePitchingCareerRecords`
  changes their function signature — needs a check of all call sites in `stat_records.go` beyond
  the four documented ones.
- Legacy migration's franchise-level (not per-season) innings prompt is a simplification: if a
  legacy franchise's innings-per-game actually changed across seasons in the original game, that
  nuance is lost. This mirrors the existing first-season-only limitation for season length and is
  considered acceptable for the same reason — documented, not solved.
- Adding a parameter to `MigrateLegacyFranchise` changes its Wails-bound signature — requires
  `wails build` to regenerate `wailsjs/` bindings before the frontend step can compile.

## Progress Tracking

- [x] 1. [DB] Add innings_per_game column to companion seasons table
- [x] 2. [Store] Read t_seasons.innings in the save game reader
- [x] 3. [Service] Wire innings into the season import write path
- [x] 4. [Store] Add shared career qualification helper
- [x] 5. [Service] Wire stat_records.go to the shared helper
- [x] 6. [Store] Fix GetBattingCareerLeaders / GetPitchingCareerLeaders
- [x] 7. [Store] Fix export_store.go career_batting/career_pitching extraConds
- [x] 8. [Service+Binding] Prompt for innings-per-game during legacy migration
- [ ] 9. [Frontend] Add innings-per-game step to legacy migration wizard
- [ ] 10. [Docs] Update user-docs/franchise-forking.md
- [ ] 11. Test coverage (see Test Coverage Plan checklist above) implemented and passing
- [ ] 12. `wails build` run to regenerate bindings after Wails signature change
- [ ] 13. `golangci-lint run` and `go test ./...` clean before commit

## Open Questions / Assumptions

- **`GameType="combined"`** (leaderboard) and **`CareerStatType="total_career"`** (export) are
  assumed to use the RS-scaled threshold, not the unscaled PO minimums — since these blend
  regular-season and playoff totals and playoff PA/IP is comparatively small. Not explicitly
  specified; flag for a quick sanity check after implementation.
- Variable season length or innings-per-game across a franchise's history (e.g. across forks, or
  within legacy-migrated data) is intentionally left unscaled/franchise-level per decision —
  documented as a limitation, not fixed.
- Whether `t_franchise_season_creation_params.innings` should ever be used as a fallback when
  `t_seasons.innings` is somehow null for a given season is unconfirmed — current snapshots show
  both always populated and in agreement, so this is treated as a non-issue unless real-world data
  proves otherwise.
