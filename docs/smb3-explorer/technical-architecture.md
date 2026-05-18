# SMB3Explorer: Technical Architecture

## Technology Stack

| Component | Technology |
|-----------|-----------|
| Language | C# (.NET 7) |
| UI Framework | WPF (Windows Presentation Foundation) |
| UI Pattern | MVVM with partial class data service |
| Database Access | `Microsoft.Data.Sqlite` |
| CSV Output | `CsvHelper` library |
| Logging | Serilog |

## Project Structure

Single C# project (`SMB3Explorer`) with the following key namespaces:

- `Services/DataService/` — database access layer (partial class, split by domain)
- `Models/Exports/` — CSV model classes with CsvHelper attribute mappings
- `ViewModels/` — MVVM view models
- `Enums/` — player traits and position enumerations
- `SqlFiles/` — embedded SQL query files (`.sql` resources)

## Data Access Layer: `DataService`

`DataService` is implemented as a **partial class** split across multiple files, each responsible for a specific domain area:

| File | Responsibility |
|------|---------------|
| `DataServiceInit.cs` | Decompression, SQLite connection setup, schema validation |
| `DataServiceFranchises.cs` | Loading franchise/league lists |
| `DataServiceFranchiseSeasons.cs` | Loading season records |
| `DataServiceFranchiseTeams.cs` | Team standings and playoff data |
| `DataServiceFranchiseCareer.cs` | Career player statistics |
| `DataServiceMostRecentSeason.cs` | Current season player/team/schedule exports |

All files implement the `IDataService` interface, enabling testability via mocking.

## Decompression and Connection Flow

```
User selects .sav file
        ↓
ZLib inflate → raw SQLite bytes
        ↓
Write to %TEMP%/smb3_explorer_yyyyMMddHHmmssfff.sqlite
        ↓
Open Microsoft.Data.Sqlite connection (Mode=ReadOnly)
        ↓
Schema validation: check for t_stats and t_leagues tables
        ↓
Fire ConnectionChanged event → UI reacts
```

If the user provides a pre-decompressed `.sqlite` file, the decompression step is skipped.

## Application State

**`ApplicationContext`** holds runtime state:
- Currently selected league
- Currently loaded franchise seasons
- Currently active game preference (SMB3 or SMB4)

**`ApplicationConfig`** persists user preferences to a JSON file:
- League access history (league ID, name, first accessed, last accessed, access count)
- Last-used game preference

## CSV Export Pipeline

```
DataService query → C# model objects (IEnumerable<T>)
        ↓
CsvWriterWrapper.WriteCsvFile<T>(records, filePath)
        ↓
CsvHelper writes to %LOCALAPPDATA%\SMB3Explorer\{filename}.csv
```

CSV models use `CsvHelper` attributes (`[Index(n)]`, `[Name("...")]`) to define column order and header names. Computed properties (derived stats like BA, OBP, FIP, etc.) are calculated in the model classes.

## SMB3 vs. SMB4 Query Differences

The `DataService` uses separate SQL query files for SMB3 and SMB4 where the schemas differ:
- `MostRecentSeasonPlayersSmb3.sql` — standard player query
- `MostRecentSeasonPlayersSmb4.sql` — adds `chemistryType`, `throwHand`, `batHand`, and pitch repertoire (JSON) columns

The active game mode is tracked in `ApplicationContext` and determines which query variant is used.

## SQL Queries

All SQL is stored as embedded resource `.sql` files (not inline strings). Key queries:
- `CareerStatsBatting.sql` / `CareerStatsPitching.sql`
- `FranchiseSeasonStandings.sql`
- `MostRecentSeasonPlayersSmb3.sql` / `MostRecentSeasonPlayersSmb4.sql`
- `MostRecentSeasonTeams.sql`
- `MostRecentSeasonSchedule.sql` / `MostRecentSeasonPlayoffSchedule.sql`
- `DatabaseTables.sql` (schema validation)

## Navigation

Simple two-screen navigation:
1. **Landing page**: game selection + save file/league selection
2. **Home/export page**: export triggers

`NavigationService` routes between these via events.

## Key Design Decisions

- **Read-only SQLite**: the connection is always opened with `Mode=ReadOnly` to prevent any possibility of corrupting the save file
- **Decompression to temp**: the decompressed file lives in `%TEMP%` and is cleaned up on disconnect; the original `.sav` is never modified
- **Lazy season loading**: franchise seasons are loaded on demand when the user selects a league, not upfront
- **Partial class DataService**: splits a large service class by domain without introducing multiple service dependencies

---

## Source Files

**SMB3Explorer** (`C:\Users\Trey\source\SMB3Explorer`):
- `SMB3Explorer/Services/DataService/DataServiceInit.cs` — decompression, SQLite connection, schema validation
- `SMB3Explorer/Services/DataService/DataService.cs` — main service class and SQL execution helpers
- `SMB3Explorer/Services/DataService/DataServiceFranchises.cs` — franchise/league loading
- `SMB3Explorer/Services/DataService/DataServiceFranchiseSeasons.cs` — season record loading
- `SMB3Explorer/Services/DataService/DataServiceFranchiseTeams.cs` — team standings queries
- `SMB3Explorer/Services/DataService/DataServiceFranchiseCareer.cs` — career player stat queries
- `SMB3Explorer/Services/DataService/DataServiceMostRecentSeason.cs` — current season export queries
- `SMB3Explorer/Services/DataService/IDataService.cs` — data service interface
- `SMB3Explorer/Services/ApplicationContext/ApplicationContext.cs` — runtime state (selected league, seasons, game preference)
- `SMB3Explorer/ApplicationConfig/ApplicationConfig.cs` — JSON-persisted league access history
- `SMB3Explorer/Services/NavigationService/NavigationService.cs` — two-screen navigation
- `SMB3Explorer/App.xaml.cs` — application entry point and DI setup
