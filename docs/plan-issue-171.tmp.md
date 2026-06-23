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

Separately, issue #171 raised the idea of also scaling thresholds by game length (innings per
game), since SMB4 supports variable-length games. Investigation confirmed innings-per-game is not
documented anywhere in the save game schema and is not read by the import pipeline at all — finding
it requires manually decompressing a real `.sav` file and inspecting ~20 undocumented tables. Per
decision, this is left as a future investigation; no schema or import changes are included here.

Also per decision: career season-length scaling continues to use the franchise's first season only
(no averaging/weighting across seasons whose length might change, e.g. across a fork) — this is
documented as a known limitation rather than fixed.

## Feature Summary

Consolidate the three duplicated "3000 PA / 1000 IP scaled by 162-game season" threshold formulas
into one shared helper, and fix career playoff qualification to use unscaled MLB postseason
minimums (40 PA or 18 BB+H for batting; 30 IP/90 outs or 6 decisions for pitching) instead of the
misapplied regular-season-scaled threshold. No save-game schema or import changes are made —
innings-per-game import is explicitly out of scope and tracked separately. Document the
first-season-only scaling limitation in the franchise-forking docs.

## Affected Areas

- **`internal/store/`** — new shared helper file (e.g. `career_qualification.go`) computing
  RS-scaled batting/pitching thresholds plus fixed PO constants; reuses existing
  `GetFranchiseSeasonLength`.
- **`internal/service/stat_records.go`** — replace the duplicated formula (lines ~95-98) with the
  shared helper; extend `computeBattingCareerRecords`/`computePitchingCareerRecords` (and the
  rate-stat equivalents) to take an OR-conditioned qualifier for PO records (PA or BB+H; outs or
  decisions) instead of a single threshold int.
- **`internal/store/leaderboard_query.go`** — `GetBattingCareerLeaders`/`GetPitchingCareerLeaders`:
  remove the duplicated inline SQL threshold subquery and TODOs; branch threshold logic on
  `f.GameType` (regular_season → scaled, playoffs → unscaled OR-condition).
- **`internal/store/export_store.go`** — same fix for `career_batting`/`career_pitching` dataset
  cases, branching on `CareerStatType`.
- **`user-docs/franchise-forking.md`** — add a Known Limitations note: career qualification
  thresholds (and career stat scaling generally) are calibrated to the franchise's first season
  length and don't adapt if season length changes after a fork; game length (innings) isn't
  tracked yet either.
- **Tests** — `leaderboard_query_test.go`, `export_store_test.go`, `stat_records_test.go`, plus a
  new test file for the shared helper.

## Implementation Tasks

```
1. [Store] Add shared career qualification helper
   What: New function (e.g. CareerQualificationThresholds(ctx) (CareerThresholds, error))
         returning: BattingPAThresholdRS (3000*seasonLength/162), PitchingOutsThresholdRS (same),
         and fixed constants BattingPAThresholdPO=40, BattingBBHThresholdPO=18,
         PitchingOutsThresholdPO=90 (30 IP), PitchingDecisionsThresholdPO=6.
   Why: Single source of truth, replacing 3 duplicated formulas; fixes the playoff bug at the root.
   Depends on: none

2. [Service] Wire stat_records.go to the shared helper
   What: Replace the hardcoded threshold calc; change computeBattingCareerRecords /
         computePitchingCareerRecords (and the *RateRecords variants) to accept an
         OR-conditioned qualifier for PO: batting qualifies if PA>=40 OR (hits+walks)>=18;
         pitching qualifies if outsPitched>=90 OR (wins+losses)>=6.
   Why: Service-layer career record/leader computation respects correct PO minimums.
   Depends on: #1
   Note: verify "hits"/"walks" keys exist in battingStatExtractors and "wins"/"losses" in
         pitchingStatExtractors before wiring — they should, since OBP/decisions already
         require them, but confirm exact key names in code.

3. [Store] Fix GetBattingCareerLeaders / GetPitchingCareerLeaders
   What: Remove the `SELECT ... FROM seasons ORDER BY season_num LIMIT 1` SQL threshold
         subquery and TODO comments. Compute the threshold value in Go via the new helper
         and bind it as a parameter. Branch on f.GameType: "playoffs" → OR-condition unscaled
         minimums; default/regular_season → scaled threshold (existing behavior, just
         de-duplicated). "combined" → use the regular-season scaled threshold (see Open
         Questions).
   Why: Fixes the playoff misapplication bug flagged by the existing TODOs.
   Depends on: #1

4. [Store] Fix export_store.go career_batting/career_pitching extraConds
   What: Same fix, branching on CareerStatType ("regular_season"/"playoffs"/"total_career").
         "total_career" keeps the RS-scaled threshold (see Open Questions).
   Why: Export datasets get the same correction as the leaderboard queries.
   Depends on: #1

5. [Docs] Update user-docs/franchise-forking.md
   What: Add a "Known Limitations" section noting career qualification thresholds (and
         career stat scaling generally) are calibrated to the first season's length and
         don't adapt to season-length changes after a fork; also note game length
         (innings/game) isn't tracked at all yet.
   Why: Sets correct user expectations per decision not to fix multi-season scaling.
   Depends on: none
```

## Test Coverage Plan

**Test framework**: Go `testing` package, table-driven style, using existing
`internal/testutil` seed helpers (`seedSeason`, `seedPlayer`, `seedPlayerSeason`, `seedBatting`,
etc.) — same patterns already in `leaderboard_query_test.go`.

**Strategy**: Verify the shared helper's scaled-RS math against known season lengths, verify PO
thresholds are fixed regardless of season length, and verify the OR-condition qualifies players
who'd previously been wrongly excluded by the scaled-RS formula.

- [ ] unit: shared helper computes correct RS threshold for 40/81/162-game seasons; returns 0 when
      no seasons exist — File: `internal/store/career_qualification_test.go`
- [ ] unit: shared helper PO constants are season-length-invariant — File:
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
- [ ] regression: existing `TestGetBattingCareerLeaders_QualifiedOnly` /
      `TestGetPitchingCareerLeaders_QualifiedOnly` (regular_season, 40-game season,
      threshold=740) still pass unchanged — File: `leaderboard_query_test.go`
- [ ] integration: `export_store.go` preview with `CareerStatType="playoffs"` applies unscaled
      OR-condition — File: `export_store_test.go`
- [ ] regression: existing `TestPreviewExportData_QualifiedOnly_CareerBatting/Pitching`
      (regular_season, 162-game) unchanged — File: `export_store_test.go`
- [ ] unit: `computeBattingCareerRecords`/`computePitchingCareerRecords` PO qualifier correctly
      OR-gates on totals map — File: `stat_records_test.go`

## Risk & Considerations

- Combining "combined" `GameType` and "total_career" `CareerStatType` qualification basis is an
  assumption (defaulting to the RS-scaled threshold) — flagged as an open question below rather
  than guessed silently into final behavior.
- The OR-condition qualifier change to `computeBattingCareerRecords`/`computePitchingCareerRecords`
  changes their function signature — needs a check of all call sites in `stat_records.go` beyond
  the four documented ones.

## Open Questions / Assumptions

- **`GameType="combined"`** (leaderboard) and **`CareerStatType="total_career"`** (export) are
  assumed to use the RS-scaled threshold, not the unscaled PO minimums — since these blend
  regular-season and playoff totals and playoff PA/IP is comparatively small. Not explicitly
  specified; flag for a quick sanity check after implementation.
- Game length (innings per game) import remains fully out of scope per decision — no schema or
  import pipeline changes. A future investigation (decompress a real `.sav` or the bundled
  `league-template.sqlite`, run `PRAGMA table_info` across the ~20 undocumented tables) is required
  before that work can even be scoped, and should be a separate issue.
- Variable season length across a franchise's history (e.g. across forks) is intentionally left
  unscaled per decision — documented as a limitation, not fixed.
