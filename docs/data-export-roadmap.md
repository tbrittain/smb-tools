# Data Export — Phased Roadmap

Tracks the three-phase delivery plan for the Data Export feature (`/export`). Each phase is independently releasable.

---

## Phase 1 — MVP (Complete, PR #157)

**Shipped:** Dataset picker + column selector + MVP filters + live preview + CSV download + per-franchise presets.

### Datasets (5 of 8)

| Dataset ID | Label | Season filter | Team filter | Career stat type |
|---|---|---|---|---|
| `batting_season` | Player Season Batting | ✓ | ✓ | — |
| `pitching_season` | Player Season Pitching | ✓ | ✓ | — |
| `standings` | Team Season Standings | ✓ | ✓ | — |
| `career_batting` | Career Batting Stats | — | — | ✓ |
| `career_pitching` | Career Pitching Stats | — | — | ✓ |

### Filters (MVP)

- **Season range** — "From Season / To Season" numeric inputs (season datasets only)
- **Team** — single-team dropdown (season datasets only)
- **Career stat type** — Reg Season / Playoffs / Total toggle (career datasets only)

### Architecture choices that carry forward

- **Column allowlist in Go** (`internal/store/export_store.go` `datasetDef.cols`) is the security boundary and query builder source of truth. Adding a dataset means adding a `datasetDef` var there and a matching entry in `frontend/src/lib/exportDatasets.ts`.
- **`FilterRowDTO`** (`column`, `op`, `value`, `value2`) is already wired end-to-end in Go and the frontend composable. The MVP filter panel just generates a fixed set of these rows implicitly; Phase 2 exposes them explicitly.
- **`ExportOptionsDTO.Filters []FilterRowDTO`** accepts any number of rows — no Go changes needed in Phase 2.
- **Preview cap** is 500 rows with a `TotalCount` field so the UI can show "Showing 500 of N."

---

## Phase 2 — Flexible User-Specified Filters (Complete, PR #158)

**Goal:** Let users add arbitrary column comparison rows instead of only the hard-coded season/team/stat-type controls from Phase 1.

### What the user gets

An "Add Filter" button in the `ExportFilterPanel` that appends a new row. Each row has:

```
[Column dropdown]  [Op dropdown]  [Value input]  [× remove]
```

Multiple rows combine with AND logic. The user can layer as many as needed — e.g. "home_runs > 20 AND ops_plus > 130 AND team_name = Bisons".

### Columns available for filtering

Only columns that exist in the selected dataset's `selectedColumnKeys` are available in the dropdown — no pointing at columns the user hasn't pulled in. This keeps the UI coherent and avoids server-side column-lookup errors.

### Ops supported (already validated in Go)

| Op | Label | Applicable to |
|---|---|---|
| `eq` | = | all |
| `neq` | ≠ | all |
| `lt` | < | int, float |
| `lte` | ≤ | int, float |
| `gt` | > | int, float |
| `gte` | ≥ | int, float |
| `contains` | contains | string |

`value2` (reserved in `FilterRowDTO`) is available for a future `between` op without a schema change.

### Implementation scope

**Go:** No changes to `export_store.go` or DTOs — the filter pipeline already handles arbitrary `FilterRowDTO` slices. The `buildExportQuery` function validates each `op` against `validFilterOps` and each `column` against the dataset's allowlist.

**Frontend:**
- `ExportFilterPanel.vue` — replace (or augment) the hard-coded season/team/stat-type controls with a dynamic filter list. The MVP controls can remain as convenience shortcuts; the generic rows are additive.
- `useExportConfig.ts` — replace `seasonMin`/`seasonMax`/`selectedTeamName` refs with a `filterRows: Ref<FilterRowDTO[]>` array. The `buildOptions()` function assembles the `FilterRowDTO[]` directly from this array instead of constructing them from individual refs.
- `ExportPresetConfig` interface — replace `seasonMin`, `seasonMax`, `selectedTeamName` fields with a `filters: FilterRowDTO[]` field. Existing presets saved in Phase 1 will not deserialize filter rows (they'll just default to empty), which is fine.

### What doesn't change

- Go store layer — zero changes
- `ExportPreviewTable`, `ExportDatasetPicker`, `ExportColumnSelector`, `ExportPresetManager` — untouched
- The `careerStatType` control stays as a dedicated toggle (it maps to a column filter internally but has fixed semantics that deserve their own UI)

---

## Phase 3 — Additional Datasets (9 total, Complete)

**Shipped:** 4 new datasets added (`player_season_attributes`, `award_winners`, `regular_season_schedule`, `playoff_schedule`). Also fixed a bug in `buildExportQuery` where filter conditions on datasets ending with a LEFT JOIN were incorrectly appended to the ON clause instead of as a WHERE clause.

### Datasets added

| Dataset ID | Label | Primary source table(s) | Filter support |
|---|---|---|---|
| `player_season_attributes` | Player Season Attributes | `player_season_game_stats`, `player_seasons`, `players`, `seasons` | All columns including enum filters |
| `award_winners` | Season Award Winners | `player_season_awards`, `awards`, `player_seasons`, `players`, `seasons` | `award_type` enum (Winner/Runner-Up) |
| `regular_season_schedule` | Regular Season Schedule | `team_season_schedules`, `team_season_history`, `seasons` | All columns |
| `playoff_schedule` | Playoff Schedule | `team_playoff_schedules`, `team_season_history`, `seasons` | All columns |

#### `player_season_attributes`

The 8 game attributes on a 1–99 scale, plus computed percentile rankings from `player_season_attribute_percentiles`. Representative columns:

```
player_name, season_num, team_name, primary_position, pitcher_role,
power, contact, speed, fielding, arm, velocity, junk, accuracy,
power_pct, contact_pct, speed_pct, fielding_pct, arm_pct, velocity_pct, junk_pct, accuracy_pct,
power_pct_role, ...
```

#### `award_winners`

One row per player-season-award combination. Representative columns:

```
player_name, season_num, team_name, award_name, award_original_name
```

The `award_original_name` is the canonical DB value; `award_name` is the human-readable label. This dataset has no natural sort column — default sort is `season_num DESC, award_name ASC`.

#### `season_schedule`

One row per game. Representative columns:

```
season_num, game_num, home_team, away_team, home_score, away_score, game_type (regular/playoff)
```

`game_type` is an enum column; the existing `careerStatType`-style toggle pattern could work here, or it can be a generic filter row (Phase 2 enables this for free).

### Implementation scope

**Go — `internal/store/export_store.go`:** Add one `datasetDef` package-level var per new dataset, following the exact same pattern as the 5 existing defs. No changes to `buildExportQuery`, `PreviewExportData`, or `ExportToCSV`.

**Frontend — `frontend/src/lib/exportDatasets.ts`:** Add one `ExportDatasetDef` entry per new dataset to `EXPORT_DATASETS`. Column keys must match the Go `datasetDef.cols` keys exactly.

**No other files change.** The dataset picker, column selector, filter panel, preview table, and composable all work generically — they'll pick up new datasets automatically.

### Go tests to add

Following the pattern in `internal/store/export_store_test.go`:
- `TestPreviewExportData_PlayerSeasonAttributes` — returns rows with attribute columns
- `TestPreviewExportData_AwardWinners` — returns rows; verify award_name present
- `TestPreviewExportData_SeasonSchedule` — returns rows; verify game_num present

---


## Phase 4 — shorthand legacy SMB3Explorer export support

Support exporting data in the exact format that CSV exports from SMB3Explorer exported (either by exposing default seeded preset things or equivalent)

## Summary

| Phase | Scope | Go changes | Frontend changes | Blocking dependencies |
|---|---|---|---|---|
| 1 — MVP | 5 datasets, fixed filters, preview, CSV, presets | New store + migration + bindings | New page + 5 components + composable | — |
| 2 — Flexible filters | Arbitrary filter rows per user | None | `ExportFilterPanel` + `useExportConfig` refactor | Phase 1 |
| 3 — Full dataset catalog | 3 more datasets (8 total) | 3 new `datasetDef` vars | 3 entries in `exportDatasets.ts` | Phase 1 (Phase 2 optional but nice) |

Phases 2 and 3 are independent — either can ship first. Phase 3 is lower effort; Phase 2 has more UX work but no backend changes.
