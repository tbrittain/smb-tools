# SmbExplorerCompanion: User Features

## Navigation Overview

The app is organized around a sidebar navigation with the following sections:
- Home (franchise dashboard)
- Players (leaderboards: batting careers, batting seasons, pitching careers, pitching seasons)
- Teams (historical teams list)
- Franchise Management (import CSV, awards delegation, Hall of Famers)

Plus franchise switching from the main menu.

---

## Franchise Selection / Management

- Multiple franchises can coexist in a single installation ("manager mode")
- New franchises created by clicking "New Franchise" and providing a name
- Franchises have an `IsSmb3` flag but SMB3 support is not fully developed
- Switching between franchises updates all views to reflect the selected franchise

---

## Home Screen

The franchise dashboard. Displays on launch after selecting a franchise.

**Franchise Summary** (top panel):
- Total players, total seasons, number of Hall of Famers
- Most recent season's champion team name
- Most recent MVP and Cy Young award winners
- League leaders: home runs, hits, RBIs, wins, saves, strikeouts

**League Summary** (standings panel):
- Current season standings by conference and division
- Wins, losses, win percentage, games behind

**Search**:
- Type a player or team name to search across the entire franchise history
- Results are grouped by type (Players / Teams)
- Click a result to navigate to that player's or team's detail page

**Quick actions**:
- "Go to recent champion team" — navigates directly to the most recent championship team's season detail page
- "Go to random player" — navigates to a randomly selected player profile

---

## Player Overview Screen

A comprehensive per-player profile page. Accessible from search, leaderboards, or team rosters.

**Career Stats tab**:
- Lifetime batting and/or pitching statistics across all seasons
- Separate views for regular season and playoffs
- Awards earned listed chronologically

**Season Breakdown tab**:
- Per-season statistics table with season selector
- Shows team(s) the player was on each season (with intra-season trade tracking)
- Salary per season

**Attributes tab**:
- Game attributes (Power, Contact, Speed, Fielding, Arm, Velocity, Junk, Accuracy) per season
- Percentile rankings relative to all players in the franchise (overall and by position)
- Percentile rankings for statistical KPIs (e.g., BA percentile, ERA percentile)

**Visualizations**:
- Radial percentile chart (spider/radar chart) showing attribute percentiles
- Percentile distribution chart

**Similar players**: Recommended players with comparable attribute profiles (batters compared to batters, pitchers to pitchers)

**Player metadata**: Primary and secondary position, pitcher role, handedness, chemistry type, traits

---

## Team Overview Screen

Top-level team page showing all-time team information.

- Team name history (the team may have been renamed across seasons)
- Logo history with full-size and icon-size logos
- Navigation to individual season pages for any season the team existed

---

## Team Season Detail Screen

In-depth view for a single team in a single season.

**Roster section**:
- Batters: full roster with season batting stats and game attributes
- Pitchers: full roster with season pitching stats and game attributes
- Click any player to navigate to their player overview

**Financials**:
- Season budget, payroll, surplus
- Surplus per game

**Season record**:
- Regular season wins/losses, run differential
- Division and conference placement

**Playoff results** (if applicable):
- Playoff seed, playoff record, playoff runs scored/allowed
- Championship winner indicator

**Schedule breakdown**:
- Game-by-game results for the regular season
- Stats split by division opponent vs. non-division opponent

**Performance trend visualization**:
- Interactive chart of margin of victory across the season
- Toggle to compare against division opponents vs. entire schedule

---

## Top Batting Careers

Franchise-wide career batting leaderboard.

**Filters**:
- Position (filter to specific primary positions)
- Chemistry type (SMB4: Competitive, Spirited, Disciplined, Scholarly, Crafty)
- Bat handedness (R/L/S)
- Hall of Famers only

**Season range**:
- Configurable start season and end season

**Toggle**: Regular season vs. playoffs

**Sortable columns**: all career batting statistics

**Pagination**: 20 results per page

---

## Top Batting Seasons

Franchise-wide single-season batting leaderboard. Same filter and sort options as Top Batting Careers, but shows individual season performances rather than career totals.

---

## Top Pitching Careers

Franchise-wide career pitching leaderboard. Same filter structure as batting (position, chemistry, handedness, HoF only, season range, reg/playoff toggle, pagination).

**Sortable columns**: all career pitching statistics

---

## Top Pitching Seasons

Franchise-wide single-season pitching leaderboard.

---

## Import CSV Screen

Wizard for importing a new season's data from SMB3Explorer CSV exports.

**Step 1**: Select the franchise season number to import into

**Step 2**: Select all 8 required CSV files:

| File | Contents |
|------|----------|
| Teams data | `MostRecentSeasonTeams.csv` from SMB3Explorer |
| Overall player data | `MostRecentSeasonPlayers.csv` from SMB3Explorer |
| Season batting stats | `SeasonBattingRegularSeason.csv` |
| Season pitching stats | `SeasonPitchingRegularSeason.csv` |
| Season schedule | `MostRecentSeasonSchedule.csv` |
| Playoff batting stats | `SeasonBattingPlayoffs.csv` |
| Playoff pitching stats | `SeasonPitchingPlayoffs.csv` |
| Playoff schedule | `MostRecentSeasonPlayoffSchedule.csv` |

**Validation**: All 8 files are required before the import can proceed. The app validates file presence before importing.

**After import**: Player/Team records are created or matched by GUID (tracked in `PlayerGameIdHistory` / `TeamGameIdHistory` tables). Stats are stored with an `IsRegularSeason` flag to distinguish regular season from playoff records.

---

## Hall of Famers Screen

- Lists all players who are Hall of Fame eligible (determined by retirement status and career criteria)
- Allows the user to manually induct eligible players
- `IsHallOfFamer` flag is set on the `Player` entity upon induction
- Hall of Famers appear on the home screen summary and can be filtered for in leaderboards

---

## Awards Delegation Screen

Manual award assignment for each season. The user assigns awards after simulating a season.

**User-assignable awards**:
- MVP (batting or pitching)
- Cy Young
- Silver Slugger
- Rookie of the Year (ROY)
- Gold Glove (fielding)
- Playoff MVP
- Championship MVP
- All-Star
- Secondary runner-up versions: MVP-2 through MVP-5, Cy Young-2 through Cy Young-5, ROY-2 through ROY-5

**Automatically calculated awards** (no user input required):
- **Triple Crown (Batting)**: awarded to the player leading in BA, HR, and RBI
- **Triple Crown (Pitching)**: awarded to the pitcher leading in W, ERA, and K
- **Title awards**: Batting Title (BA leader), Home Run Title (HR leader), RBI Title, ERA Title, Wins Title, Strikeouts Title

Awards are stored in the `Awards` collection on `PlayerSeason` entities. The `Importance` field controls display priority (0 = MVP, highest priority).

---

## Historical Teams Screen

Lists all teams that have existed across all seasons in the franchise, including teams from any era. Each team entry links to team season detail pages for the seasons it was active.
