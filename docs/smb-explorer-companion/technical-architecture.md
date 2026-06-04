# SmbExplorerCompanion: Technical Architecture

## Technology Stack

| Component | Technology |
|-----------|-----------|
| Language | C# (.NET 7) |
| UI Framework | WPF (Windows Presentation Foundation) |
| UI Pattern | MVVM (CommunityToolkit.Mvvm) |
| Application Pattern | CQRS via MediatR |
| Database ORM | Entity Framework Core (SQLite) |
| CSV Parsing | CsvHelper with ClassMap |
| Visualizations | ScottPlot |
| Dependency Injection | Microsoft.Extensions.DependencyInjection |

## Solution Structure (5 Projects)

```
SmbExplorerCompanion.sln
├── SmbExplorerCompanion.Core        # Business logic: MediatR handlers (commands + queries)
├── SmbExplorerCompanion.Database    # EF Core entities, DbContext, migrations
├── SmbExplorerCompanion.Csv         # CSV model classes and ClassMap definitions
├── SmbExplorerCompanion.WPF         # UI: ViewModels, Views, navigation
└── SmbExplorerCompanion.Shared      # Shared utilities and constants
```

## CQRS Pattern (MediatR)

All data operations are expressed as **commands** (write) or **queries** (read) sent via `IMediator`. Each operation has a dedicated handler class in `SmbExplorerCompanion.Core`.

Example query handlers:
- `GetFranchiseSummaryQueryHandler` — home screen summary data
- `GetSearchResultsQueryHandler` — global search
- `GetTopBattingCareersQueryHandler` — leaderboard query with filters
- `GetPlayerGameStatPercentilesRequestHandler` — attribute percentile rankings
- `GetPlayerKpiPercentilesRequestHandler` — statistical KPI rankings
- `GetHallOfFameCandidatesRequestHandler` — HoF eligibility

Example command handlers:
- `ImportCsvCommandHandler` — processes the 8-file CSV import
- `AddHallOfFamersCommandHandler` — marks players as inducted

## Entity Framework Core

- Database: SQLite, stored at `%LOCALAPPDATA%\SmbExplorerCompanion\SmbExplorerCompanion.db`
- Migrations: standard EF Core code-first migrations manage schema evolution
- Seeding: lookup tables (traits, awards, chemistry, positions, etc.) are seeded in `OnModelCreating` via `SeedLookups()`
- `IsRegularSeason` flag on `PlayerSeasonBattingStat` and `PlayerSeasonPitchingStat` distinguishes regular season from playoff rows for the same player season

## CSV Import Pipeline

1. User selects 8 CSV files in the import wizard
2. `CsvReaderService` parses each file using CsvHelper with `ClassMap` configurations
3. Each CSV model class has a corresponding `ClassMap` defining column-name-to-property mappings
4. Parsed data is passed to `ICsvImportRepository` for persistence
5. Player/team identity is tracked via `PlayerGameIdHistory` and `TeamGameIdHistory` (GUID → internal ID mapping)
6. If a player GUID matches an existing record, the existing `Player` entity is used; otherwise a new one is created
7. Intra-season player team changes are tracked in `PlayerTeamHistory` with an `Order` field

## Navigation

`INavigationService` handles screen transitions and passes typed parameters between ViewModels. Each screen receives its parameters via the navigation service rather than a global store.

## Percentile Ranking System

Two query types compute player rankings within the franchise:

- **`GetPlayerGameStatPercentilesRequest`**: ranks game attributes (Power, Contact, etc.) across all players in the franchise (optionally scoped to a season or position)
- **`GetPlayerKpiPercentilesRequest`**: ranks statistical KPIs (BA, ERA, etc.)

Rankings are computed as percentiles — a player at the 95th percentile in home runs has hit more HR than 95% of all players in the franchise's history.

## Award Automation

Three award types are computed automatically without user input:
- **Title awards** (Batting, HR, RBI, ERA, Wins, Strikeouts): stat leaders identified from `PlayerSeasonBattingStat` / `PlayerSeasonPitchingStat`
- **Triple Crown (Batting)**: player leading in all three of BA, HR, RBI
- **Triple Crown (Pitching)**: pitcher leading in all three of W, ERA, K

All other awards (`IsUserAssignable = true`) require manual assignment via the Awards Delegation screen.

## Key Design Decisions

- **CQRS**: separates reads from writes, making it straightforward to add new views or operations without touching existing handlers
- **EF Core migrations**: ensures the companion DB schema can evolve across app versions without data loss
- **ClassMap-based CSV parsing**: decouples CSV column names from C# property names, allowing the CSV format to match the SMB3Explorer output format while the internal model uses idiomatic naming
- **PlayerGameIdHistory / TeamGameIdHistory**: enables re-imports across seasons to match players and teams by their game GUIDs rather than name, which is more reliable than name matching (especially with name changes or re-introduced free agents)
- **Seeded lookup tables**: traits, awards, chemistry, positions are seeded once and referenced by FK, keeping the schema normalized and enabling filtering by these dimensions

---

## Source Files

**SmbExplorerCompanion** (https://github.com/tbrittain/SmbExplorerCompanion):

*Database / EF Core:*
- `SmbExplorerCompanion.Database/SmbExplorerCompanionDbContext.cs` — DbContext, relationships, seeding
- `SmbExplorerCompanion.Database/DependencyInjection.cs` — DI registration for database layer

*CSV parsing:*
- `SmbExplorerCompanion.Csv/Services/CsvReaderService.cs` — CsvHelper-based parsing orchestrator
- `SmbExplorerCompanion.Csv/Models/OverallPlayer.cs` — player CSV model with ClassMap
- `SmbExplorerCompanion.Csv/Models/SeasonStatBatting.cs` — batting stats CSV model
- `SmbExplorerCompanion.Csv/Models/SeasonStatPitching.cs` — pitching stats CSV model
- `SmbExplorerCompanion.Csv/Models/SeasonSchedule.cs` — schedule CSV model
- `SmbExplorerCompanion.Csv/Models/PlayoffSchedule.cs` — playoff schedule CSV model
- `SmbExplorerCompanion.Csv/Models/Team.cs` — team CSV model

*Repositories:*
- `SmbExplorerCompanion.Database/Services/Imports/CsvImportRepository.cs` — CSV import persistence logic
- `SmbExplorerCompanion.Database/Services/Players/GeneralPlayerRepository.cs`
- `SmbExplorerCompanion.Database/Services/Players/PositionPlayerCareerRepository.cs`
- `SmbExplorerCompanion.Database/Services/Players/PitcherCareerRepository.cs`
- `SmbExplorerCompanion.Database/Services/SearchRepository.cs`
- `SmbExplorerCompanion.Database/Services/SummaryRepository.cs`
- `SmbExplorerCompanion.Database/Services/TeamRepository.cs`

*WPF services:*
- `SmbExplorerCompanion.WPF/Services/NavigationService.cs` — typed screen-to-screen navigation
- `SmbExplorerCompanion.WPF/Services/ApplicationContext.cs` — global WPF app state
- `SmbExplorerCompanion.WPF/Services/MappingService.cs` — DTO → ViewModel mapping
- `SmbExplorerCompanion.WPF/App.xaml.cs` — application entry point and DI setup
