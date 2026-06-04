# SMB3Explorer: CSV Export Schemas

All exports are CSV files with a header row. Column order matches the `[Index(n)]` attributes on the C# model classes. Derived/computed fields are noted.

---

## Career Batting Statistics

**Files**: `CareerBattingRegularSeason.csv`, `CareerBattingPlayoffs.csv`

| Index | Header | Type | Notes |
|-------|--------|------|-------|
| 0 | First Name | string | |
| 1 | Last Name | string | |
| 2 | Team | string? | Current team |
| 3 | Prev Team | string? | Most recent previous team |
| 4 | 2nd Prev Team | string? | Second most recent team |
| 5 | Retirement Season | int? | Season number when retired; null if active |
| 6 | Age | int | |
| 7 | Position | string | Primary position |
| 8 | Secondary Position | string? | |
| 9 | Pitcher Role | string? | SP/RP/SP-RP/CL; null for non-pitchers |
| 10 | Games Batting | int | Games in which player batted |
| 11 | Games Played | int | Total games appeared |
| 12 | AB | int | At bats |
| 13 | PA | int | Plate appearances (computed) |
| 14 | R | int | Runs |
| 15 | H | int | Hits |
| 16 | BA | double | Batting average (computed) |
| 17 | 1B | int | Singles (computed: H - 2B - 3B - HR) |
| 18 | 2B | int | Doubles |
| 19 | 3B | int | Triples |
| 20 | HR | int | Home runs |
| 21 | RBI | int | Runs batted in |
| 22 | XBH | int | Extra base hits (computed) |
| 23 | TB | int | Total bases (computed) |
| 24 | SB | int | Stolen bases |
| 25 | CS | int | Caught stealing |
| 26 | BB | int | Walks |
| 27 | K | int | Strikeouts |
| 28 | HBP | int | Hit by pitch |
| 29 | OBP | double | On-base percentage (computed) |
| 30 | SLG | double | Slugging percentage (computed) |
| 31 | OPS | double | OBP + SLG (computed) |
| 32 | wOBA | double | Weighted on-base average (computed) |
| 33 | ISO | double | Isolated power: SLG - BA (computed) |
| 34 | BABIP | double | Batting avg on balls in play (computed) |
| 35 | Sac Hits | int | Sacrifice hits (bunts) |
| 36 | Sac Flies | int | Sacrifice flies |
| 37 | Errors | int | Defensive errors |
| 38 | Passed Balls | int | Catcher passed balls |
| 39 | PA/Game | double | Plate appearances per game (computed) |
| 40 | AB/HR | double | At bats per home run (computed) |
| 41 | K% | double | Strikeout rate (computed) |
| 42 | BB% | double | Walk rate (computed) |
| 43 | XBH% | double | Extra base hit rate (computed) |

---

## Season Batting Statistics

**Files**: `SeasonBattingRegularSeason.csv`, `SeasonBattingPlayoffs.csv`

Same columns as Career Batting (indices 0–43) with these **differences**:
- Index 0 is `Season` (int) instead of First Name — season number prepended
- All other indices shift by 1 (Season is prepended at index 0)
- Column order: Season, First Name, Last Name, Team, Prev Team, 2nd Prev Team, Position, Secondary Position, Pitcher Role, Age, Games Batting, Games Played, AB, PA, R, H, BA, 1B, 2B, 3B, HR, RBI, XBH, TB, SB, CS, BB, K, HBP, OBP, SLG, OPS, wOBA, ISO, BABIP, Sac Hits, Sac Flies, Errors, Passed Balls, PA/Game, AB/HR, K%, BB%, XBH%

---

## Most Recent Season Players

**File**: `MostRecentSeasonPlayers.csv`

| Index | Header | Type | Notes |
|-------|--------|------|-------|
| 0 | Season | int | Season number |
| 1 | First Name | string | |
| 2 | Last Name | string | |
| 3 | Position | string | Primary position |
| 4 | Secondary Position | string? | |
| 5 | Pitcher Role | string? | |
| 6 | Team | string? | Current team |
| 7 | Prev Team | string? | Previous team |
| 8 | Power | int | 1–99 |
| 9 | Contact | int | 1–99 |
| 10 | Speed | int | 1–99 |
| 11 | Fielding | int | 1–99 |
| 12 | Arm | int? | 1–99; null if not tracked |
| 13 | Velocity | int? | 1–99; pitchers only |
| 14 | Junk | int? | 1–99; pitchers only |
| 15 | Accuracy | int? | 1–99; pitchers only |
| 16 | Age | int | |
| 17 | Salary | int | In display units (game units × 200) |
| 18 | Trait 1 | string? | Trait name |
| 19 | Trait 2 | string? | Second trait name |
| 20 | Chemistry | string? | SMB4 only; Competitive/Spirited/Disciplined/Scholarly/Crafty |
| 21 | Throw Hand | string | R or L |
| 22 | Bat Hand | string | R, L, or S |
| 23 | Pitch 1 | string? | SMB4 only; pitch type abbreviation |
| 24 | Pitch 2 | string? | |
| 25 | Pitch 3 | string? | |
| 26 | Pitch 4 | string? | |
| 27 | Pitch 5 | string? | |
| 28 | PlayerId | Guid | SMB save game player GUID |
| 29 | SeasonId | int | SMB save game season ID |

> **Note**: In SMB3, Chemistry (20), Throw Hand (21), Bat Hand (22), and Pitch 1–5 (23–27) columns are present but empty for SMB3 saves.

---

## Most Recent Season Teams

**File**: `MostRecentSeasonTeams.csv`

| Index | Header | Type | Notes |
|-------|--------|------|-------|
| 0 | Team | string | |
| 1 | Division | string | |
| 2 | Conference | string | |
| 3 | Season | int | |
| 4 | Budget | int | In display units |
| 5 | Payroll | int | In display units |
| 6 | Surplus | int | Budget - Payroll |
| 7 | Surplus/Game | int | Surplus / games played |
| 8 | W | int | Wins |
| 9 | L | int | Losses |
| 10 | Run Differential | int | Runs For - Runs Against |
| 11 | Runs For | int | |
| 12 | Runs Against | int | |
| 13 | GB | double | Games behind division leader |
| 14 | WPCT | double | Win percentage |
| 15 | Pythagorean WPCT | double | Expected win% based on run diff (computed) |
| 16 | Expected W | int | Wins expected by Pythagorean (computed) |
| 17 | Expected L | int | Losses expected by Pythagorean (computed) |
| 18 | Power | int | Team aggregate power |
| 19 | Contact | int | Team aggregate contact |
| 20 | Speed | int | Team aggregate speed |
| 21 | Fielding | int | Team aggregate fielding |
| 22 | Arm | int | Team aggregate arm |
| 23 | Velocity | int | Team aggregate velocity |
| 24 | Junk | int | Team aggregate junk |
| 25 | Accuracy | int | Team aggregate accuracy |
| 26 | TeamId | Guid | SMB save game team GUID |
| 27 | SeasonId | int | SMB save game season ID |

---

## Most Recent Season Schedule (Regular Season)

**File**: `MostRecentSeasonSchedule.csv`

| Index | Header | Type | Notes |
|-------|--------|------|-------|
| 0 | Season | int | |
| 1 | Game Number | int | Sequential game number |
| 2 | Day | int | Day of the season |
| 3 | Home Team | string | |
| 4 | Away Team | string | |
| 5 | Home Score | int? | Null if game not yet played |
| 6 | Away Score | int? | Null if game not yet played |
| 7 | Home Pitcher | string? | Starting pitcher name |
| 8 | Away Pitcher | string? | Starting pitcher name |
| 9 | HomeTeamId | Guid | |
| 10 | AwayTeamId | Guid | |
| 11 | HomePitcherId | Guid? | |
| 12 | AwayPitcherId | Guid? | |
| 13 | SeasonId | int | |

---

## Most Recent Season Playoff Schedule

**File**: `MostRecentSeasonPlayoffSchedule.csv`

| Index | Header | Type | Notes |
|-------|--------|------|-------|
| 0 | Season | int | |
| 1 | Series | int | Series number within playoffs |
| 2 | Team 1 | string | |
| 3 | Team 1 Seed | int | |
| 4 | Team 2 | string | |
| 5 | Team 2 Seed | int | |
| 6 | Game Number | int | |
| 7 | Home Team | string | |
| 8 | Away Team | string | |
| 9 | Home Score | int? | |
| 10 | Away Score | int? | |
| 11 | Home Pitcher | string? | |
| 12 | Away Pitcher | string? | |
| 13 | Team1Id | Guid | |
| 14 | Team2Id | Guid | |
| 15 | HomeTeamId | Guid | |
| 16 | AwayTeamId | Guid | |
| 17 | HomePitcherId | Guid? | |
| 18 | AwayPitcherId | Guid? | |
| 19 | SeasonId | int | |

---

## Career Pitching Statistics

**Files**: `CareerPitchingRegularSeason.csv`, `CareerPitchingPlayoffs.csv`

| Index | Header | Type | Notes |
|-------|--------|------|-------|
| 0 | First Name | string | |
| 1 | Last Name | string | |
| 2 | Team | string? | |
| 3 | Prev Team | string? | |
| 4 | 2nd Prev Team | string? | |
| 5 | Retirement Season | int? | |
| 6 | Age | int | |
| 7 | Pitcher Role | string? | |
| 8 | W | int | Wins |
| 9 | L | int | Losses |
| 10 | CG | int | Complete games |
| 11 | CGSO | int | Complete game shutouts |
| 12 | H | int | Hits allowed |
| 13 | ER | int | Earned runs |
| 14 | HR | int | Home runs allowed |
| 15 | BB | int | Walks |
| 16 | K | int | Strikeouts |
| 17 | IP | double | Innings pitched (computed from outs recorded) |
| 18 | ERA | double | Earned run average (computed) |
| 19 | TP | int | Total pitches thrown |
| 20 | SV | int | Saves |
| 21 | HBP | int | Hit batters |
| 22 | Batters Faced | int | |
| 23 | Games Played | int | |
| 24 | Games Started | int | |
| 25 | Games Finished | int | |
| 26 | Runs Allowed | int | Total runs (earned + unearned) |
| 27 | WP | int | Wild pitches |
| 28 | BAA | double | Batting average against (computed) |
| 29 | FIP | double | Fielding independent pitching (computed) |
| 30 | WHIP | double | Walks + hits per inning (computed) |
| 31 | WPCT | double | Win percentage (computed) |
| 32 | Opp OBP | double | Opponent on-base percentage (computed) |
| 33 | K/BB | double | Strikeout to walk ratio (computed) |
| 34 | K/9 | double | Strikeouts per 9 innings (computed) |
| 35 | BB/9 | double | Walks per 9 innings (computed) |
| 36 | H/9 | double | Hits per 9 innings (computed) |
| 37 | HR/9 | double | Home runs per 9 innings (computed) |
| 38 | Pitches Per Inning | double | (computed) |
| 39 | Pitches Per Game | double | (computed) |

---

## Season Pitching Statistics

**Files**: `SeasonPitchingRegularSeason.csv`, `SeasonPitchingPlayoffs.csv`

Same columns as Career Pitching but with `Season` prepended at index 0, and two additional trailing columns (most recent season only):

| 40 | ERA- | double | Park/league adjusted ERA (computed) |
| 41 | FIP- | double | Park/league adjusted FIP (computed) |
| 42 | PlayerId | Guid | |
| 43 | SeasonId | int | |
| 44 | TeamId | Guid? | |
| 45 | MostRecentTeamId | Guid? | |
| 46 | PreviousTeamId | Guid? | |

---

## Franchise Season Standings

**File**: `FranchiseSeasonStandings.csv`

| Index | Header | Type | Notes |
|-------|--------|------|-------|
| 0 | Index | int | Row number (ranking position within season) |
| 1 | Season | int | |
| 2 | Team | string | |
| 3 | Division | string | |
| 4 | Conference | string | |
| 5 | W | int | |
| 6 | L | int | |
| 7 | Runs For | int | |
| 8 | Runs Against | int | |
| 9 | Run Differential | int | |
| 10 | WPCT | double | |
| 11 | GB | double | Games behind division leader |

---

## Franchise Playoff Standings

**File**: `FranchisePlayoffStandings.csv`

| Index | Header | Type | Notes |
|-------|--------|------|-------|
| 0 | Index | int | Rank (1 = champion, 2 = runner-up, etc.) |
| 1 | Season | int | |
| 2 | Team | string | |
| 3 | Division | string | |
| 4 | Conference | string | |
| 5 | W | int | Playoff wins |
| 6 | L | int | Playoff losses |
| 7 | Runs For | int | Playoff runs scored |
| 8 | Runs Against | int | Playoff runs allowed |
| 9 | Run Differential | int | Playoff run differential |

---

## Source Files

**SMB3Explorer** (https://github.com/tbrittain/SMB3Explorer):
- `SMB3Explorer/Models/Exports/BattingStatistic.cs` — abstract base with all batting stat properties and `[Index]`/`[Name]` attributes
- `SMB3Explorer/Models/Exports/BattingMostRecentSeasonStatistic.cs` — adds OPS+, PlayerId, SeasonId, TeamId columns
- `SMB3Explorer/Models/Exports/CareerStatistic.cs` — career batting base class
- `SMB3Explorer/Models/Exports/CareerBattingStatistic.cs` — career batting full model
- `SMB3Explorer/Models/Exports/PitchingStatistic.cs` — abstract base with all pitching stat properties
- `SMB3Explorer/Models/Exports/PitchingMostRecentSeasonStatistic.cs` — adds ERA-, FIP-, PlayerId, TeamId columns
- `SMB3Explorer/Models/Exports/CareerPitchingStatistic.cs` — career pitching full model
- `SMB3Explorer/Models/Exports/SeasonPlayer.cs` — most recent season player model (attributes, traits, pitches)
- `SMB3Explorer/Models/Exports/SeasonTeam.cs` — most recent season team model
- `SMB3Explorer/Models/Exports/SeasonSchedule.cs` — regular season schedule model
- `SMB3Explorer/Models/Exports/SeasonPlayoffSchedule.cs` — playoff schedule model
- `SMB3Explorer/Models/Exports/FranchiseSeasonStanding.cs` — franchise season standings model
- `SMB3Explorer/Models/Exports/FranchisePlayoffStanding.cs` — franchise playoff standings model
- `SMB3Explorer/Constants/FileExports.cs` — output filename constants
