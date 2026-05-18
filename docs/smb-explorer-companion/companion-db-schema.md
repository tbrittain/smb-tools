# SmbExplorerCompanion: Database Schema

The companion app maintains its own SQLite database at `%LOCALAPPDATA%\SmbExplorerCompanion\SmbExplorerCompanion.db`. This is distinct from the SMB save game file — it is the companion's persistent store for historical franchise data.

Schema is managed via Entity Framework Core code-first migrations. Lookup tables are seeded on first run.

---

## Organization Entities

### `Franchises`

| Property | Type | Notes |
|----------|------|-------|
| Id | int | PK |
| Name | string | User-provided franchise name |
| IsSmb3 | bool | True for SMB3 franchises (partial support) |

Navigation: `Conferences`, `Teams`, `Players`, `Seasons`

### `Seasons`

| Property | Type | Notes |
|----------|------|-------|
| Id | int | PK (DatabaseGenerated: None — set from game data) |
| Number | int | Season number |
| NumGamesRegularSeason | int | Length of the regular season |
| FranchiseId | int | FK to Franchises |
| ChampionshipWinnerId | int? | FK to ChampionshipWinners |

Navigation: `PlayerSeasons`, `SeasonTeamHistory`, `ChampionshipWinner`

### `Conferences`

| Property | Type | Notes |
|----------|------|-------|
| Id | int | PK |
| Name | string | |
| IsDesignatedHitter | bool | Whether DH rule applies |
| FranchiseId | int | FK to Franchises |

Navigation: `Divisions`

### `Divisions`

| Property | Type | Notes |
|----------|------|-------|
| Id | int | PK |
| Name | string | |
| ConferenceId | int | FK to Conferences |

Navigation: `SeasonTeamHistories`

### `Teams`

| Property | Type | Notes |
|----------|------|-------|
| Id | int | PK |
| FranchiseId | int | FK to Franchises |

Navigation: `TeamGameIdHistory`, `SeasonTeamHistory`

---

## Player Entities

### `Players`

| Property | Type | Notes |
|----------|------|-------|
| Id | int | PK |
| FirstName | string | Locked after first import |
| LastName | string | Locked after first import |
| IsHallOfFamer | bool | Set via Hall of Famers screen |
| BatHandednessId | int | FK to BatHandedness |
| ThrowHandednessId | int | FK to ThrowHandedness |
| PrimaryPositionId | int | FK to Positions |
| PitcherRoleId | int? | FK to PitcherRoles; null for non-pitchers |
| ChemistryId | int? | FK to Chemistry; null for SMB3 |
| FranchiseId | int | FK to Franchises |

Navigation: `PlayerSeasons`, `PlayerGameIdHistory`

### `PlayerSeasons`

One record per player per season.

| Property | Type | Notes |
|----------|------|-------|
| Id | int | PK |
| PlayerId | int | FK to Players |
| SeasonId | int | FK to Seasons |
| Age | int | Player's age during this season |
| Salary | int | Display salary (already multiplied) |
| SecondaryPositionId | int? | FK to Positions |
| ChampionshipWinnerId | int? | FK to ChampionshipWinners if on championship team |

Navigation: `PitchTypes` (many-to-many), `PlayerTeamHistory`, `GameStats` (one-to-one), `Traits` (many-to-many), `BattingStats`, `PitchingStats`, `Awards` (many-to-many), schedule collections, `ChampionshipWinner`

### `PlayerSeasonGameStat`

Game attribute snapshot for a player in a season.

| Property | Type | Notes |
|----------|------|-------|
| Id | int | PK |
| PlayerSeasonId | int | FK to PlayerSeasons (one-to-one) |
| Power | int | |
| Contact | int | |
| Speed | int | |
| Fielding | int | |
| Arm | int? | |
| Velocity | int? | Pitchers only |
| Junk | int? | Pitchers only |
| Accuracy | int? | Pitchers only |

### `PlayerSeasonBattingStats`

| Property | Type | Notes |
|----------|------|-------|
| Id | int | PK |
| PlayerSeasonId | int | FK to PlayerSeasons |
| IsRegularSeason | bool | True = regular season; False = playoffs |
| GamesPlayed | int | |
| GamesBatting | int | |
| AtBats | int | |
| PlateAppearances | int | |
| Runs | int | |
| Hits | int | |
| Singles | int | |
| Doubles | int | |
| Triples | int | |
| HomeRuns | int | |
| RunsBattedIn | int | |
| ExtraBaseHits | int | |
| TotalBases | int | |
| StolenBases | int | |
| CaughtStealing | int | |
| Walks | int | |
| Strikeouts | int | |
| HitByPitch | int | |
| SacrificeHits | int | |
| SacrificeFlies | int | |
| Errors | int | |
| PassedBalls | int | |
| BattingAverage | double? | Computed and stored |
| Obp | double? | |
| Slg | double? | |
| Ops | double? | |
| Woba | double? | |
| Iso | double? | |
| Babip | double? | |
| OpsPlus | double? | |
| PaPerGame | double? | |
| AbPerHomeRun | double? | |
| StrikeoutPercentage | double? | |
| WalkPercentage | double? | |
| ExtraBaseHitPercentage | double? | |

### `PlayerSeasonPitchingStats`

| Property | Type | Notes |
|----------|------|-------|
| Id | int | PK |
| PlayerSeasonId | int | FK to PlayerSeasons |
| IsRegularSeason | bool | True = regular season; False = playoffs |
| Wins | int | |
| Losses | int | |
| CompleteGames | int | |
| Shutouts | int | |
| Saves | int | |
| GamesPlayed | int | |
| GamesStarted | int | |
| GamesFinished | int | |
| Hits | int | |
| EarnedRuns | int | |
| RunsAllowed | int | |
| HomeRuns | int | |
| Walks | int | |
| Strikeouts | int | |
| HitByPitch | int | |
| WildPitches | int | |
| BattersFaced | int | |
| TotalPitches | int | |
| InningsPitched | double? | |
| EarnedRunAverage | double? | |
| BattingAverageAgainst | double? | |
| Fip | double? | |
| Whip | double? | |
| WinPercentage | double? | |
| OpponentObp | double? | |
| StrikeoutsPerWalk | double? | |
| StrikeoutsPerNine | double? | |
| WalksPerNine | double? | |
| HitsPerNine | double? | |
| HomeRunsPerNine | double? | |
| PitchesPerInning | double? | |
| PitchesPerGame | double? | |
| EraMinus | double? | |
| FipMinus | double? | |

### `PlayerTeamHistory`

Tracks intra-season team changes (trades).

| Property | Type | Notes |
|----------|------|-------|
| Id | int | PK |
| PlayerSeasonId | int | FK to PlayerSeasons |
| SeasonTeamHistoryId | int? | FK to SeasonTeamHistory |
| Order | int | Ordering of team changes within the season |

---

## Team History Entities

### `SeasonTeamHistory`

The primary per-team-per-season record.

| Property | Type | Notes |
|----------|------|-------|
| Id | int | PK |
| SeasonId | int | FK to Seasons |
| TeamId | int | FK to Teams |
| DivisionId | int | FK to Divisions |
| TeamNameHistoryId | int | FK to TeamNameHistory |
| Budget | long | |
| Payroll | long | |
| Surplus | long | |
| SurplusPerGame | double | |
| Wins | int | |
| Losses | int | |
| GamesBehind | double | |
| WinPercentage | double | |
| PythagoreanWinPercentage | double | |
| ExpectedWins | int | |
| ExpectedLosses | int | |
| RunsScored | int | |
| RunsAllowed | int | |
| TotalPower | int | Team aggregate |
| TotalContact | int | |
| TotalSpeed | int | |
| TotalFielding | int | |
| TotalArm | int | |
| TotalVelocity | int | |
| TotalJunk | int | |
| TotalAccuracy | int | |
| PlayoffSeed | int? | Null if team did not make playoffs |
| PlayoffWins | int? | |
| PlayoffLosses | int? | |
| PlayoffRunsScored | int? | |
| PlayoffRunsAllowed | int? | |

Navigation: `ChampionshipWinner`, schedule collections, `PlayerTeamHistory`

### `TeamNameHistory`

Tracks team name changes across seasons.

| Property | Type | Notes |
|----------|------|-------|
| Id | int | PK |
| Name | string | Team name for this era |
| TeamLogoHistoryId | int? | FK to TeamLogoHistory |

### `TeamLogoHistory`

Stores team logo images.

| Property | Type | Notes |
|----------|------|-------|
| Id | int | PK |
| LogoFullSize | byte[] | Full-size logo image bytes |
| LogoIconSize | byte[] | Icon-size logo image bytes |
| Order | int | Identity-generated ordering |

### `TeamSeasonSchedules`

Regular season game records.

| Property | Type | Notes |
|----------|------|-------|
| Id | int | PK |
| HomeTeamHistoryId | int | FK to SeasonTeamHistory |
| AwayTeamHistoryId | int | FK to SeasonTeamHistory |
| HomePitcherSeasonId | int? | FK to PlayerSeasons |
| AwayPitcherSeasonId | int? | FK to PlayerSeasons |
| Day | int | Day of the season |
| GlobalGameNumber | int | |
| HomeScore | int? | Null if game not yet played |
| AwayScore | int? | |

### `TeamPlayoffSchedules`

Playoff game records.

| Property | Type | Notes |
|----------|------|-------|
| Id | int | PK |
| HomeTeamHistoryId | int | FK to SeasonTeamHistory |
| AwayTeamHistoryId | int | FK to SeasonTeamHistory |
| HomePitcherSeasonId | int? | FK to PlayerSeasons |
| AwayPitcherSeasonId | int? | FK to PlayerSeasons |
| SeriesNumber | int | Which playoff series |
| GlobalGameNumber | int | |
| HomeScore | int? | |
| AwayScore | int? | |

### `ChampionshipWinners`

| Property | Type | Notes |
|----------|------|-------|
| Id | int | PK |
| SeasonTeamHistoryId | int | FK to SeasonTeamHistory (the winning team) |
| SeasonId | int | FK to Seasons |

Navigation: `PlayerSeasons` (all players on the championship team)

---

## Lookup / Reference Tables

### `BatHandedness`
Values: R, L, S

### `ThrowHandedness`
Values: R, L

### `Positions`
13 total — 9 primary positions (P, C, 1B, 2B, 3B, SS, LF, CF, RF) + 4 secondary groups (IF, OF, 1B/OF, IF/OF).
`IsPrimaryPosition` distinguishes primary from secondary.

### `Chemistry`
5 values: Competitive, Spirited, Disciplined, Scholarly, Crafty

### `Traits`
See [player-traits.md](../domain/player-traits.md) for complete list.
Properties: `Id`, `Name`, `IsSmb3`, `IsPositive`, `ChemistryId` (FK to Chemistry; null for SMB3 traits)

### `PitcherRoles`
Values: SP, RP, SP/RP, CL

### `PitchTypes`
8 values: 4F, 2F, SB, CH, FK, CB, SL, CF

### `PlayerAwards`
See [awards.md](../domain/awards.md) for complete list.
Properties: `Id`, `Name`, `OriginalName`, `IsBuiltIn`, `Importance`, `OmitFromGroupings`, `IsBattingAward`, `IsPitchingAward`, `IsFieldingAward`, `IsPlayoffAward`, `IsUserAssignable`

---

## Identity Tracking Tables

### `PlayerGameIdHistory`

Maps SMB save game player GUIDs to companion `Player` IDs.

| Property | Type | Notes |
|----------|------|-------|
| Id | int | PK |
| PlayerId | int | FK to Players |
| GameId | Guid | The player's GUID from the SMB save file |

### `TeamGameIdHistory`

Maps SMB save game team GUIDs to companion `Team` IDs.

| Property | Type | Notes |
|----------|------|-------|
| Id | int | PK |
| TeamId | int | FK to Teams |
| GameId | Guid | The team's GUID from the SMB save file |

---

## Audit Tables

### `LookupSeeds`

Tracks when lookup tables were seeded to prevent re-seeding on subsequent app launches.

| Property | Type | Notes |
|----------|------|-------|
| Id | Guid | PK (DatabaseGenerated: None) |
| SeededAt | DateTime | Timestamp of seeding |
