# SmbExplorerCompanion Visualization Analysis

This document inventories every data visualization in the original SmbExplorerCompanion WPF application (ScottPlot.WPF), describes each chart's data model and behavior, and assigns a relative **porting complexity** estimate for a Vue 3 + Apache ECharts implementation in smb-tools.

**Complexity scale:** Low / Medium / High

---

## Summary Table

| # | Chart Name | Type | Location | Complexity |
|---|-----------|------|----------|------------|
| 1 | Player Attribute Radar | Radar / Spider | Player Overview | Medium |
| 2 | Player Attribute Percentile Bars | Horizontal Bar | Player Overview | Low |
| 3 | Player KPI Percentile Bars | Horizontal Bar | Player Overview | Low–Medium |
| 4 | Team Season Schedule | Stepped Line + Error Bars | Team Season Detail | High |

All four charts live in two ViewModels:
- `SmbExplorerCompanion.WPF/ViewModels/PlayerOverviewViewModel.cs`
- `SmbExplorerCompanion.WPF/ViewModels/TeamSeasonDetailViewModel.cs`

---

## Chart 1 — Player Attribute Radar

**Complexity: Medium**

### What it shows
A radar (spider/polygon) chart that overlays two series:
1. The selected player's raw attribute scores for the currently selected season.
2. The league average attribute scores for that same season and role.

This lets the user see at a glance whether a player is above or below average on each dimension.

### Dimensions
| Player type | Axes (in order) |
|-------------|-----------------|
| Batter | Power, Contact, Speed, Fielding, Arm |
| Pitcher | Velocity, Junk, Accuracy, Fielding, Power, Contact, Speed |

All values are on the 1–99 scale. The radar polygon maximum is 99.

### Data sources
| Series | Backend query |
|--------|---------------|
| Player stats | `GetPlayerGameStatOverview` — returns `SeasonStats` for the selected player + season |
| League average | `GetLeagueAverageGameStatsRequest` — returns `LeagueAverageGameStats` aggregated by season, isPitcher, and optional pitcher role |

### Visual configuration
- Size: 300 × 300 px
- Palette: DarkPastel
- Both polygons use hatching (striped fill): upward-diagonal for the player, downward-diagonal for the league average
- Grid lines: dotted
- Legend: visible, labelled "Player Name" and "League Average Pitcher/Batter"
- Zoom: 1.5× applied
- Pan/zoom interaction: **disabled**
- Title: dynamic — e.g. "Pitcher Attributes (Season 1)"

### Porting notes for ECharts
ECharts has a built-in `radar` chart type that maps directly. The dual-series polygon with distinct fill patterns is the only non-trivial part — ECharts supports `areaStyle` with opacity; true hatching requires SVG pattern tricks or can be approximated with semi-transparent fills. The league-average series should use a visually distinct style (dashed border, different color/opacity) rather than a literal hatch.

---

## Chart 2 — Player Attribute Percentile Bars

**Complexity: Low**

### What it shows
A horizontal bar chart showing where the player ranks (0–100 percentile) on each physical attribute relative to all players in the league for the selected season. It answers "is this player's arm in the top 10%?"

### Dimensions
| Player type | Bars |
|-------------|------|
| Batter | Power, Contact, Speed, Fielding, Arm (5 bars) |
| Pitcher | Velocity, Junk, Accuracy, Fielding, Power, Contact, Speed (7 bars) |

### Data source
`GetPlayerGameStatPercentilesRequest` → `PlayerGameStatPercentiles` — returns a pre-computed percentile (0–100) for each attribute for the given player, season, and role.

### Visual configuration
- Size: 300 × 350 px
- Y-axis: fixed 0–100
- **Color gradient per bar**: linear interpolation from Blue (0th percentile) → White (50th) → Red (100th)
  - This is a custom `GetValueColor(value)` function in the ViewModel; each bar is independently colored
- Each bar has a numeric label showing its percentile
- Pan/zoom interaction: **disabled**
- Title: dynamic — e.g. "Pitcher Attribute Percentiles (Season 1)"

### Porting notes for ECharts
ECharts bar charts support per-bar `itemStyle.color` via a callback, so the blue→white→red gradient logic translates directly. The chart itself is straightforward; complexity is only in replicating the color-mapping function.

---

## Chart 3 — Player KPI Percentile Bars

**Complexity: Low–Medium**

### What it shows
The same percentile bar concept as Chart 2, but applied to performance statistics (counting stats and rate stats) rather than physical attributes. Pitchers have significantly more KPI bars than batters.

### Dimensions
| Player type | KPIs |
|-------------|------|
| Batter | Hits, HR, BA, SB, Strikeouts (batter), OBP, SLG (7 bars) |
| Pitcher | Wins, ERA, WHIP, IP, K/9, K/BB, Hits, HR, BA, SB, K%, OBP, SLG (13 bars) |

Note: pitchers include all batter-side KPIs in addition to pitcher-specific ones. The chart height scales accordingly (~600 px for pitchers vs ~300 px for batters).

### Data source
`GetPlayerKpiPercentilesRequest` → `PlayerKpiPercentiles` — same percentile pattern as Chart 2, but for statistical KPIs.

### Visual configuration
- Size: 300 × 600 px (pitchers) / 300 × ~350 px (batters)
- Same blue→white→red color gradient as Chart 2
- Bar labels showing percentile value per bar
- Pan/zoom interaction: **disabled**
- Title: dynamic — e.g. "Pitcher Stat Percentiles (Season 1)"
- Chart regenerates whenever the season selection or pitcher role filter changes

### Porting notes for ECharts
Essentially the same chart as Chart 2 with more bars and a larger data set. The dynamic height (5 vs 13 bars) needs to be accounted for in the Vue component layout — a fixed container height will cause overcrowding for pitchers. The color gradient logic is shared with Chart 2 and should be extracted into a composable.

The key additional complexity vs Chart 2: the KPI label strings need context-aware display names (e.g. "K/9" for pitchers vs "Strikeouts" for batters), which requires a mapping layer.

---

## Chart 4 — Team Season Schedule (Games Above/Below .500)

**Complexity: High**

### What it shows
A stepped line chart tracking each team's cumulative win-delta (wins minus losses) across every game of the season. A horizontal reference line at Y = 0 represents the .500 threshold. A team trending upward is winning more than they lose; downward is the reverse.

This is the most interactive and configurable chart in the app, and the one with the most direct mapping to GitHub issue #75.

### Axes
| Axis | What it represents |
|------|--------------------|
| X | Game number (1 → ~162) |
| Y | Cumulative wins minus losses (0 = .500 record) |

> **Note:** GitHub issue #75 describes the Y axis as "running win %" rather than raw win delta. The companion app uses win delta (integer), but win % is arguably more readable. This is an open design decision for the smb-tools implementation.

### Data source
`GetTeamScheduleBreakdownRequest` → `TeamScheduleBreakdowns` — returns a list of game-by-game records with:
- `Day` (game number)
- `WinsDelta` (cumulative wins minus losses after this game)
- `TeamScore`
- `OpponentTeamScore`
- `OpponentTeamName`

### Series and toggles
The chart supports two user-controlled toggles that trigger a full redraw:

1. **Division toggle** (`IncludeDivisionTeamsInPlot`): When off, only the current team's line is shown. When on, all division teams are shown simultaneously, each with a distinct color from the Microcharts palette.

2. **Margin of victory toggle** (`IncludeMarginOfVictoryInPlot`): When on, each game node gets a one-sided error bar:
   - **Win**: positive error bar upward showing the run differential
   - **Loss**: negative error bar downward showing the run differential
   This adds a vertical spike to each point indicating how close the game was.

### Hover tooltip
Hovering near a data point triggers a tooltip showing:
```
Game {Day}: {TeamName} {win/lose} against {OpponentTeamName} {TeamScore} - {OpponentTeamScore}
```
The detection uses a proximity threshold (±1 day on X, ±0.5 on Y) with Euclidean-distance tie-breaking for overlapping series (division mode).

### Visual configuration
- Size: 300 × 1000 px (very tall to show the full season arc)
- Line style: stepped (ScatterStep — each game is a discrete step, not a smooth curve)
- Markers: filled circles, size 2, one per game
- Reference line: black dashed horizontal at Y = 0
- X-axis labels: rotated 45°
- Axis labels: "Day" (X), "Games >.500" (Y)
- Legend: visible, lower-left, one entry per team
- Palette: Microcharts (distinct colors for up to ~4–5 division teams)
- Pan/zoom interaction: **disabled**

### Porting notes for ECharts
This is the most complex chart to port because of:

1. **Stepped line rendering** — ECharts supports `step: 'start'` on line series, which approximates ScottPlot's ScatterStep behavior.

2. **Error bars (margin of victory)** — ECharts does not have a native one-sided error bar primitive. The standard approach is to use a `custom` renderer or overlay a separate `bar` series with negative/positive offset bars styled to look like error bars. Alternatively, the Candlestick series can simulate this with some effort.

3. **Multi-series division mode** — Straightforward in ECharts; each team becomes a `line` series. The toggle simply adds/removes series from the chart config reactively.

4. **Hover tooltips** — ECharts tooltip `formatter` can produce the required label. The proximity-based detection in ScottPlot is replaced by ECharts' native tooltip trigger, which already handles overlapping series via `axisPointer`.

5. **Scale** — Up to ~162 data points × 5 teams = ~810 points total. ECharts handles this comfortably.

---

## Additional Items from Issue #47 (Not Yet in Companion App)

Issue #47 lists five visualization areas. Four of them map to the charts above. Two are **not implemented** in the original companion app:

| Item from issue #47 | Status in companion app |
|---------------------|------------------------|
| Player attribute radar chart | ✅ Implemented (Chart 1) |
| Player KPI percentile rankings | ✅ Implemented (Charts 2 & 3) |
| Team season performance trend chart | ✅ Implemented (Chart 4) |
| Similar players | ❌ Not in companion app — net-new feature |
| Franchise-level stat trend charts (season-over-season) | ❌ Not in companion app — net-new feature |

---

## Porting Complexity Summary

| Chart | Complexity | Key challenges |
|-------|------------|----------------|
| Player Attribute Radar | **Medium** | Dual-series radar with distinct fill styles; ECharts radar type maps well but hatch patterns require workarounds |
| Player Attribute Percentile Bars | **Low** | Per-bar color gradient is the only non-trivial piece; standard ECharts bar chart otherwise |
| Player KPI Percentile Bars | **Low–Medium** | Same as above; added complexity is dynamic bar count (5 vs 13), dynamic height, and KPI label mappings |
| Team Season Schedule | **High** | Stepped line, one-sided error bars (no native ECharts primitive), multi-series division toggle, proximity hover tooltips |

### Recommended porting order
1. **Chart 2** (Attribute Percentile Bars) — lowest complexity, validates the color-gradient composable
2. **Chart 3** (KPI Percentile Bars) — reuses Chart 2's component and composable
3. **Chart 1** (Radar) — standalone chart type, medium complexity, high visual impact
4. **Chart 4** (Team Schedule) — highest complexity, most user-facing value; tackle last when the other patterns are established

---

## Data Availability Check

Before implementing any chart, confirm the following backend endpoints/queries exist or need to be added:

| Data needed | Likely location in smb-tools |
|-------------|------------------------------|
| Player attribute stats per season | `app_season.go` / `internal/store/` |
| League average attributes per season + role | May need new aggregation query |
| Player attribute percentiles (0–100) | Likely new query — percentile calculation in SQL or Go |
| Player KPI percentiles (0–100) | Likely new query |
| Game-by-game team schedule breakdown | May exist in season store; verify it includes score data |
| Division teams for a given season | Likely exists; confirm it returns per-game schedule for each team |
