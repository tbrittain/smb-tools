---
title: Stat Explorer
---

# Stat Explorer

The Stat Explorer is a flexible data export tool that lets you pull any slice of your franchise data out as a CSV file — for spreadsheet analysis, sharing with the community, or feeding into other tools.

Open it from the sidebar. The page is split into a configuration panel on the left and a live preview table on the right.

## Datasets

Pick one of nine datasets to query:

| Dataset | What it contains |
|---|---|
| **Player Season Batting** | Per-player batting stats for a single season, including counting stats, rate stats, OPS+, and smbWAR |
| **Player Season Pitching** | Per-player pitching stats for a single season, including ERA, FIP, FIP-, ERA+, and smbWAR |
| **Team Season Standings** | Win-loss records, run differential, playoff results, budget, and payroll by team and season |
| **Career Batting Stats** | Cumulative and rate batting stats across a player's entire franchise career |
| **Career Pitching Stats** | Cumulative and rate pitching stats across a player's entire franchise career |
| **Player Season Attributes** | Per-player ratings (Power, Contact, Speed, Fielding, Arm, Velocity, Junk, Accuracy) captured at end of season |
| **Season Award Winners** | Award winners and runners-up for every season |
| **Regular Season Schedule** | Game-by-game results for every regular season |
| **Playoff Schedule** | Game-by-game results for every playoff series |

## Configuring Your Export

### Columns

The column selector lists every available field for the active dataset. Toggle individual columns on or off. The CSV export contains exactly the columns you select — nothing else.

### Filters

Add one or more filter rows to narrow the dataset before exporting. Each filter row lets you pick a column, an operator, and a value. Operators are type-aware: numeric columns offer `=`, `≠`, `<`, `≤`, `>`, `≥`; text columns offer `=` and `≠`; enum columns (position, team, chemistry, etc.) show a dropdown of valid values.

Multiple filter rows are combined with AND.

### Stat Type Toggle

For season batting and pitching datasets, toggle between **Regular Season** and **Playoffs**. For career datasets, toggle between **Regular Season**, **Playoffs**, and **Total** (combined).

### Qualified Players Only

Batting and pitching datasets (both season and career) include a **Qualified Players Only** toggle. When on, the results are limited to players who meet the standard qualifying threshold — the same thresholds used in the leaderboards.

### Sort

In the right panel, pick a sort column and toggle ascending or descending. Sorting applies to both the preview and the CSV export.

## Preview

Click **Apply** to run the query and see the first page of results. The preview table is paginated. Player names and team names are clickable links that navigate to the relevant player or team page — these links appear in the preview only and are not included in the CSV.

## Exporting

Click **Export CSV** (in the left panel or the top of the preview area) to download the full result set. The export runs the same query as the preview but returns all rows, not just the current page.

## Presets

Save a named preset to store your current dataset selection, column picks, filters, sort, and stat type toggle. Presets are saved per franchise and persist across sessions. Load a saved preset to restore a configuration without having to rebuild it.
