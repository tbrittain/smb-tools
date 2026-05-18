# SmbExplorerCompanion: Open GitHub Issues

Open issues from https://github.com/tbrittain/SmbExplorerCompanion/issues as of 2026-05-18.

Issues are organized by type. Each issue has a **Notes** field for rewrite considerations.

---

## Bugs

### #252 — Error importing playoff data during the "batters" stage
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/252
**Labels**: needs investigation
**Opened**: 2025-10-15

An error occurs when importing playoff data at the batters stage. (Screenshot attached in the original issue — no text description provided by reporter.)

> **Notes**:
> <!-- Your feedback here -->

---

### #207 — Win diff from previous season broken on Teams screen
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/207
**Labels**: bug, confirmed
**Opened**: 2024-08-20

Win differential from previous season isn't calculating correctly. Also needs to only attempt the calculation when a single season range is selected.

> **Notes**:
> <!-- Your feedback here -->

---

### #199 — Explorer fails to import updated season data into existing season
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/199
**Labels**: bug, core functionality, confirmed
**Opened**: 2024-06-19

After creating a season and importing data, attempting to re-import updated data for the same season fails with "An error occurred while saving entity changes." Error occurs on the first player in the database. CSV files attached in the original issue.

> **Notes**:
> <!-- Your feedback here -->

---

### #198 — Explorer parses a team playing multiple games on the same day incorrectly
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/198
**Labels**: bug
**Opened**: 2024-06-19

The schedule chart assumes day number = game number, but the game occasionally schedules a team to play multiple games on the same day (doubleheaders). This causes a rendering error where multiple games appear with the same game number in the schedule visualization.

> **Notes**:
> <!-- Your feedback here -->

---

### #130 — Walks per nine for pitchers may be broken on import
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/130
**Labels**: bug
**Opened**: 2023-10-14

BB/9 may not be calculating or importing correctly. May require a migration of existing imported data to fix historical records.

> **Notes**:
> <!-- Your feedback here -->

---

### #109 — Num division titles doesn't include run differential tiebreaker
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/109
**Labels**: bug
**Opened**: 2023-10-11

Division title counts app-wide do not apply the run differential tiebreaker when determining division winners.

> **Notes**:
> <!-- Your feedback here -->

---

## Enhancements

### #202 — Season mode support
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/202
**Labels**: enhancement
**Opened**: 2024-06-24

The app should theoretically work with Season mode (not just Franchise mode) since the exported data is structurally similar. Needs validation with data exported from SMB3Explorer issue #48.

> **Notes**:
> <!-- Your feedback here -->

---

### #190 — Improved logging to file for debugging and tracing
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/190
**Labels**: (none)
**Opened**: 2024-05-02

Use Serilog (like SMB3Explorer does) and add logging statements throughout the app for debugging.

> **Notes**:
> <!-- Your feedback here -->

---

### #161 — Break "Current Greats" on home into "Recent" and "Sustained" Greats
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/161
**Labels**: enhancement
**Opened**: 2023-12-20

- **Sustained Greats**: current "Current Greats" functionality (career-level greatness)
- **Recent Greats**: top players by smbWAR over the past 3 or 5 seasons

> **Notes**:
> <!-- Your feedback here -->

---

### #154 — Playoff vs regular season career totals on player overview page
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/154
**Labels**: enhancement
**Opened**: 2023-12-14

Career totals currently show a single aggregate. It would be useful to break this down by regular season vs. playoffs on the player overview page.

> **Notes**:
> <!-- Your feedback here -->

---

### #126 — Stat scaling per 162
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/126
**Labels**: (none)
**Opened**: 2023-10-13

For users who don't play 162-game seasons, show stats scaled to a 162-game pace to give context for player performance.

> **Notes**:
> <!-- Your feedback here -->

---

### #125 — Team stadium tracking
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/125
**Labels**: (none)
**Opened**: 2023-10-13

Track stadium changes in a similar manner to team name changes. Referenced by the team management issue (#33).

> **Notes**:
> <!-- Your feedback here -->

---

### #117 — Bold players stats for league leaders
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/117
**Labels**: enhancement
**Opened**: 2023-10-11

Highlight/bold statistical leaders in stat tables, similar to how Baseball Reference indicates league leaders.

> **Notes**:
> <!-- Your feedback here -->

---

### #103 — More advanced team season screen
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/103
**Labels**: enhancement
**Opened**: 2023-10-08

Additions to the team season detail screen:
- Team aggregate stats (total power, contact, payroll, expected W/L — already stored)
- Most improved players from previous season (by individual game stats like power, contact, traits)
- Stadium tracking (#125)

> **Notes**:
> <!-- Your feedback here -->

---

### #95 — HoF career standards test
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/95
**Labels**: (none)
**Opened**: 2023-09-10

Implement a Hall of Fame career standards test similar to Baseball Reference's HoF standards page (https://www.baseball-reference.com/about/leader_glossary.shtml#hof_standard).

> **Notes**:
> <!-- Your feedback here -->

---

### #92 — Integration of player game stats in the Season grid view
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/92
**Labels**: enhancement
**Opened**: 2023-09-05

Show player game attributes (Power, Contact, etc.) in the season grid view. Also proposes a composite stat that combines smbWAR with game stat ratings — a "playing up to / exceeding potential" metric.

> **Notes**:
> <!-- Your feedback here -->

---

### #86 — 30/30 and 40/40 seasons
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/86
**Labels**: (none)
**Opened**: 2023-09-01

Track and highlight notable power-speed seasons (30 HR + 30 SB, 40 HR + 40 SB). Power-speed players should receive recognition for these achievements.

> **Notes**:
> <!-- Your feedback here -->

---

### #80 — Summary rows on player overview grids
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/80
**Labels**: enhancement
**Opened**: 2023-08-28

Add career summary rows at the bottom of player overview stat grids (below all season/playoff rows, not sorted with them).

> **Notes**:
> <!-- Your feedback here -->

---

### #60 — Awards delegation - Gold Gloves and playoff awards
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/60
**Labels**: enhancement
**Opened**: 2023-08-17

Add Gold Glove and playoff award delegation to the awards management flow.

> **Notes**:
> <!-- Your feedback here -->

---

### #52 — Player nicknames
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/52
**Labels**: (none)
**Opened**: 2023-08-09

Allow users to assign nicknames to players for fun/flavor.

> **Notes**:
> <!-- Your feedback here -->

---

### #40 — Improved TeamNames handling and SeasonTeam navigation from Player Overview
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/40
**Labels**: (none)
**Opened**: 2023-08-03

Track team names better on entities (as a list of team histories). From player season rows:
1. If played for one team, navigate directly on click
2. If played for multiple teams, right-click to select which team's season page to navigate to

> **Notes**:
> <!-- Your feedback here -->

---

### #36 — Import current season data directly from game db rather than going through SMB Explorer
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/36
**Labels**: (none)
**Opened**: 2023-08-02

Read directly from the SMB save game file instead of requiring SMB3Explorer CSV exports. Noted as potentially faster, better UX, but complex to implement with two DbContexts. The smb-tools rewrite addresses this as a core architectural goal.

> **Notes**:
> <!-- Your feedback here -->

---

### #35 — Awards management screen
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/35
**Labels**: enhancement, core functionality
**Opened**: 2023-08-01

- Edit the name and priority of existing awards
- Add custom awards

> **Notes**:
> <!-- Your feedback here -->

---

### #33 — Team management
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/33
**Labels**: enhancement
**Opened**: 2023-08-01

Umbrella issue for team management features, referencing:
- #6 (team logo upload)
- #30 (team colors)
- #125 (stadium tracking)

> **Notes**:
> <!-- Your feedback here -->

---

### #32 — Player uniform numbers
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/32
**Labels**: (none)
**Opened**: 2023-08-01

Track player uniform numbers, likely as a property of `PlayerSeason` or `PlayerTeamHistory`.

> **Notes**:
> <!-- Your feedback here -->

---

### #31 — Team opponent season series breakdown chart
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/31
**Labels**: enhancement
**Opened**: 2023-08-01

A record-against-opponents breakdown (similar to Baseball Reference's schedule/results table). May get complex with cross-conference opponents as a single column.

> **Notes**:
> <!-- Your feedback here -->

---

### #30 — Add team colors to logo table
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/30
**Labels**: enhancement
**Opened**: 2023-08-01

Allow users to define team colors (manual entry, color picker, or image-based color extraction from a screenshot). Use these colors to theme team names and table elements throughout the app.

> **Notes**:
> <!-- Your feedback here -->

---

### #15 — Conference management
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/15
**Labels**: enhancement
**Opened**: 2023-07-24

No way currently to set whether a conference uses the DH rule or not from the SMB3Explorer CSV exports. Needs a screen to manually toggle this within the app.

> **Notes**:
> <!-- Your feedback here -->

---

### #6 — Ability to upload team logo image for use app-wide
**URL**: https://github.com/tbrittain/SmbExplorerCompanion/issues/6
**Labels**: enhancement
**Opened**: 2023-07-16

Allow users to upload team logo images:
- Full-size and icon-size variants
- Not creating logos from the logos table; user-supplied images only
- No player headshots (too intensive)
- Infrastructure already exists in `TeamLogoHistory` entity

> **Notes**:
> <!-- Your feedback here -->
