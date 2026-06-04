# SmbExplorerCompanion: CSV Import Flow

## Overview

The import flow is how franchise history gets into the companion app. It is a manual, wizard-driven process that the user runs at the end of every in-game season (before simulating the offseason, which would trigger retirements and clear some data from the save file).

## Prerequisites

- SMB3Explorer must have already been run to export the current season's data
- 8 specific CSV files must be available from that export

## Step-by-Step Flow

### Step 1: Select Season

The user selects which season number they are importing (e.g., "Season 5"). This is the franchise season number as it appears in the game and in the exported CSV files.

### Step 2: Select Files

The user browses to and selects each of the 8 required files:

| CSV File | Source Export from SMB3Explorer |
|----------|--------------------------------|
| Teams | `MostRecentSeasonTeams.csv` |
| Overall Players | `MostRecentSeasonPlayers.csv` |
| Season Batting Stats | `SeasonBattingRegularSeason.csv` |
| Season Pitching Stats | `SeasonPitchingRegularSeason.csv` |
| Season Schedule | `MostRecentSeasonSchedule.csv` |
| Playoff Batting Stats | `SeasonBattingPlayoffs.csv` |
| Playoff Pitching Stats | `SeasonPitchingPlayoffs.csv` |
| Playoff Schedule | `MostRecentSeasonPlayoffSchedule.csv` |

All 8 files must be provided before the import button is enabled — there is no partial import.

### Step 3: Validation

Before executing the import, the app validates:
- All 8 files are selected and readable
- CSV column structure matches expected ClassMap definitions
- Season number does not conflict with an already-imported season (prevents duplicate imports)

### Step 4: Parse

`CsvReaderService` parses each file using CsvHelper with the corresponding `ClassMap`:
- `OverallPlayersCsvMapping` — maps `MostRecentSeasonPlayers.csv` columns to `OverallPlayer` objects
- `TeamCsvMapping` — maps team CSV to `Team` objects
- `SeasonStatBattingCsvMapping` — maps batting stats CSV to `SeasonStatBatting` objects
- `SeasonStatPitchingCsvMapping` — maps pitching stats CSV to `SeasonStatPitching` objects
- `SeasonScheduleCsvMapping` — maps schedule CSV to `SeasonSchedule` objects
- `PlayoffScheduleCsvMapping` — maps playoff schedule CSV to `PlayoffSchedule` objects

(Playoff batting and pitching stats reuse the same ClassMaps as regular season; the `IsRegularSeason = false` flag is set during persistence.)

### Step 5: Persist

`ICsvImportRepository` handles the EF Core persistence:

**Teams**:
- Look up `TeamGameIdHistory` by team GUID from the CSV
- If found: reuse the existing `Team` entity
- If not found: create a new `Team` entity and `TeamGameIdHistory` record
- Create `SeasonTeamHistory` record for this season (wins, losses, budget, payroll, standings, playoff results, etc.)
- Create `TeamNameHistory` record if the team's name has changed

**Players**:
- Look up `PlayerGameIdHistory` by player GUID from the CSV
- If found: reuse the existing `Player` entity
- If not found: create a new `Player` entity and `PlayerGameIdHistory` record
- Create `PlayerSeason` record for this season (age, salary, secondary position, pitch types, traits)
- Create `PlayerSeasonGameStat` (game attributes: Power, Contact, Speed, etc.)
- Create `PlayerSeasonBattingStat` with `IsRegularSeason = true` for regular season stats
- Create `PlayerSeasonBattingStat` with `IsRegularSeason = false` for playoff stats
- Create `PlayerSeasonPitchingStat` with appropriate `IsRegularSeason` flag
- Create `PlayerTeamHistory` records (ordered) to track team changes within the season

**Schedule**:
- Create `TeamSeasonSchedule` records for each regular season game
- Create `TeamPlayoffSchedule` records for each playoff game
- Both reference `SeasonTeamHistory` (home/away) and `PlayerSeason` (starting pitchers)

**Championship**:
- If a playoff champion is identifiable, create a `ChampionshipWinner` record linking the winning `SeasonTeamHistory` to the `Season`

### Step 6: Commit

All changes are committed in a single EF Core transaction. If any step fails, the entire import is rolled back.

## Post-Import

After a successful import:
- The home screen refreshes to reflect the new season
- Awards delegation screen becomes available for the imported season
- Hall of Fame candidates are re-evaluated (any newly retired players become eligible)

## Player Identity: The GUID Matching Problem

Players in the SMB save game are identified by a GUID. The companion app stores these GUIDs in `PlayerGameIdHistory`.

**Normal case**: player GUID matches → existing `Player` entity reused → stats accumulate correctly across seasons.

**Edge case: re-introduced free agents**: the game may reuse player names but assign new GUIDs, or vice versa. When a player's GUID doesn't match any existing record, a new `Player` entity is created — which can result in duplicate player entries for what appears to be the same person. This is a known limitation.

## Timing Recommendation

Import at the **end of each season, before simulating the offseason**. Once the game simulates the offseason:
- Players may retire (and their seasonal stats disappear from the save file)
- The "most recent season" data in the save file will be overwritten with the new season

The companion app cannot recover data that was never imported before these events occur.

---

## Source Files

**SmbExplorerCompanion** (https://github.com/tbrittain/SmbExplorerCompanion):
- `SmbExplorerCompanion.Csv/Services/CsvReaderService.cs` — parses all 8 CSV file types using CsvHelper ClassMaps
- `SmbExplorerCompanion.Database/Services/Imports/CsvImportRepository.cs` — persistence logic: entity matching by GUID, creation of PlayerSeason, SeasonTeamHistory, stats, schedules
- `SmbExplorerCompanion.Csv/Models/OverallPlayer.cs` — player CSV model (maps `MostRecentSeasonPlayers.csv`)
- `SmbExplorerCompanion.Csv/Models/Team.cs` — team CSV model
- `SmbExplorerCompanion.Csv/Models/SeasonStatBatting.cs` — batting stats ClassMap
- `SmbExplorerCompanion.Csv/Models/SeasonStatPitching.cs` — pitching stats ClassMap
- `SmbExplorerCompanion.Csv/Models/SeasonSchedule.cs` — schedule ClassMap
- `SmbExplorerCompanion.Csv/Models/PlayoffSchedule.cs` — playoff schedule ClassMap
- `SmbExplorerCompanion.Database/Entities/PlayerGameIdHistory.cs` — GUID → Player ID mapping
- `SmbExplorerCompanion.Database/Entities/TeamGameIdHistory.cs` — GUID → Team ID mapping
