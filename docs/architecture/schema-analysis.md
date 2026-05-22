# Legacy Schema Analysis — SmbExplorerCompanion → smb-tools

Documents every mapping decision for `LegacyMigrationService`. The source is
`SmbExplorerCompanion.db` (Entity Framework SQLite, one shared DB for all
franchises). The target is a per-franchise companion DB created by smb-tools.

---

## Database topology difference

| Legacy | smb-tools |
|---|---|
| Single `SmbExplorerCompanion.db` shared by all franchises | One `<franchise-id>.db` per franchise |
| `Franchises` table ties all data together | No franchise_id in companion DB |
| Integer FK lookup tables (positions, hands, chemistry, etc.) | Denormalised TEXT values stored directly |
| Stored derived/rate stats (BA, ERA, OPS, etc.) | Raw counting stats only; views compute rates on read |

---

## Table mapping

| Legacy table | New table | Notes |
|---|---|---|
| `Franchises` | new entry in `registry.db franchises` | One migration = one new franchise |
| `Seasons` | `seasons` | `Id` → `save_game_season_id`; `Number` → `season_num`; `NumGamesRegularSeason` → `num_games` |
| `Teams` + `TeamGameIdHistory` | `teams` + `team_alt_guids` | First `GameId` (lowest `Id`) → `game_guid`; extras → `team_alt_guids` |
| `SeasonTeamHistory` + `TeamNameHistory` + `Divisions` + `Conferences` | `team_season_history` | Team name from `TeamNameHistory.Name`; conference/division resolved via joins |
| `Players` + `PlayerGameIdHistory` | `players` + `player_alt_guids` | First `GameId` → `game_guid`; extras → `player_alt_guids`; `IsHallOfFamer` direct |
| `PlayerSeasons` | `player_seasons` | Lookup IDs resolved to text; `SecondaryPositionId` → `secondary_position` |
| `PlayerTeamHistory` | `player_seasons.team_history_id` | Row where `Order = 1` (current team); NULL if none |
| `PlayerSeasonGameStats` | `player_season_game_stats` | `Arm`/`Velocity`/`Junk`/`Accuracy` NULL → 0 |
| `PlayerSeasonBattingStats` | `player_season_batting_stats` | Counting stats only; see columns below |
| `PlayerSeasonPitchingStats` | `player_season_pitching_stats` | Counting stats only; `InningsPitched` → `outs_pitched` |
| `PlayerSeasonTrait` → `Traits.Name` | `player_seasons.traits_json` | JSON array of trait name strings |
| `PitchTypePlayerSeason` → `PitchTypes.Name` | `player_seasons.pitches_json` | JSON array of pitch type name strings |
| `PlayerAwardPlayerSeason` → `PlayerAwards` | `player_season_awards` | Built-ins matched by `OriginalName`; custom awards inserted |
| `TeamSeasonSchedules` | `team_season_schedules` | IDs remapped via migration ID maps |
| `TeamPlayoffSchedules` | `team_playoff_schedules` | IDs remapped via migration ID maps |

---

## Counting stats to migrate

### Batting (`PlayerSeasonBattingStats` → `player_season_batting_stats`)

| Legacy column | New column |
|---|---|
| `GamesPlayed` | `games_played` |
| `GamesBatting` | `games_batting` |
| `AtBats` | `at_bats` |
| `Runs` | `runs` |
| `Hits` | `hits` |
| `Doubles` | `doubles` |
| `Triples` | `triples` |
| `HomeRuns` | `home_runs` |
| `RunsBattedIn` | `rbi` |
| `StolenBases` | `stolen_bases` |
| `CaughtStealing` | `caught_stealing` |
| `Walks` | `walks` |
| `Strikeouts` | `strikeouts` |
| `HitByPitch` | `hit_by_pitch` |
| `SacrificeHits` | `sac_hits` |
| `SacrificeFlies` | `sac_flies` |
| `Errors` | `errors` |
| `PassedBalls` | `passed_balls` |

**Skipped (computed by view):** `PlateAppearances`, `Singles`, `ExtraBaseHits`, `TotalBases`,
`Obp`, `Slg`, `Ops`, `Woba`, `Iso`, `Babip`, `BattingAverage`, `PaPerGame`,
`AbPerHomeRun`, `StrikeoutPercentage`, `WalkPercentage`, `ExtraBaseHitPercentage`

**Skipped (Phase 8.5):** `OpsPlus`

### Pitching (`PlayerSeasonPitchingStats` → `player_season_pitching_stats`)

| Legacy column | New column | Notes |
|---|---|---|
| `Wins` | `wins` | |
| `Losses` | `losses` | |
| `GamesPlayed` | `games` | Note: column renamed |
| `GamesStarted` | `games_started` | |
| `CompleteGames` | `complete_games` | |
| `Shutouts` | `shutouts` | |
| `Saves` | `saves` | |
| `InningsPitched` | `outs_pitched` | Transform: see below |
| `Hits` | `hits_allowed` | Note: column renamed |
| `EarnedRuns` | `earned_runs` | |
| `HomeRuns` | `home_runs_allowed` | Note: column renamed |
| `Walks` | `walks` | |
| `Strikeouts` | `strikeouts` | |
| `HitByPitch` | `hit_batters` | Note: column renamed |
| `BattersFaced` | `batters_faced` | |
| `GamesFinished` | `games_finished` | |
| `RunsAllowed` | `runs_allowed` | |
| `WildPitches` | `wild_pitches` | |
| `TotalPitches` | `total_pitches` | |

**Skipped (computed by view):** `EarnedRunAverage`, `BattingAverageAgainst`, `Whip`,
`WinPercentage`, `OpponentObp`, `StrikeoutsPerWalk`, `StrikeoutsPerNine`,
`WalksPerNine`, `HitsPerNine`, `HomeRunsPerNine`, `PitchesPerInning`, `PitchesPerGame`

**Skipped (Phase 8.5):** `Fip`, `EraMinus`, `FipMinus`

### Team history (`SeasonTeamHistory` → `team_season_history`)

| Legacy column | New column | Notes |
|---|---|---|
| `Wins` | `wins` | |
| `Losses` | `losses` | |
| `GamesBehind` | `games_back` | |
| `RunsScored` | `runs_for` | Note: renamed |
| `RunsAllowed` | `runs_against` | |
| `Budget` | `budget` | |
| `Payroll` | `payroll` | |
| `TotalPower` | `total_power` | |
| `TotalContact` | `total_contact` | |
| `TotalSpeed` | `total_speed` | |
| `TotalFielding` | `total_fielding` | |
| `TotalArm` | `total_arm` | |
| `TotalVelocity` | `total_velocity` | |
| `TotalJunk` | `total_junk` | |
| `TotalAccuracy` | `total_accuracy` | |
| `PlayoffSeed` | `playoff_seed` | |
| `PlayoffWins` | `playoff_wins` | |
| `PlayoffLosses` | `playoff_losses` | |
| `PlayoffRunsScored` | `playoff_runs_for` | Note: renamed |
| `PlayoffRunsAllowed` | `playoff_runs_against` | |

**Skipped:** `Surplus`, `SurplusPerGame`, `WinPercentage`, `PythagoreanWinPercentage`,
`ExpectedWins`, `ExpectedLosses`

---

## Key transforms

### InningsPitched → outs_pitched

`InningsPitched` is a nullable `REAL`. The decimal digit represents additional
outs (0, 1, or 2), not fractions. `6.2` = 6 complete innings + 2 outs = 20 outs.

```go
func ipToOuts(ip float64) int64 {
    whole := math.Floor(ip)
    frac  := ip - whole
    return int64(whole)*3 + int64(math.Round(frac*10))
}
// NULL → 0 outs
```

### Lookup IDs → text

Legacy uses integer FK lookup tables. The reader resolves these to text strings
at read time by pre-loading all lookup tables into maps.

| Column | Table | Example |
|---|---|---|
| `BatHandednessId` | `BatHandedness.Name` | 1 → "R" |
| `ThrowHandednessId` | `ThrowHandedness.Name` | 2 → "L" |
| `PrimaryPositionId` | `Positions.Name` | 8 → "CF" |
| `SecondaryPositionId` | `Positions.Name` | nullable → "" |
| `PitcherRoleId` | `PitcherRoles.Name` | nullable → "" |
| `ChemistryId` | `Chemistry.Name` | nullable → "" |

### Traits / pitch types → JSON arrays

`PlayerSeasonTrait` join → `Traits.Name` slice → `["Clutch","Tough Out"]`
`PitchTypePlayerSeason` join → `PitchTypes.Name` slice → `["4F","SL"]`

Empty slices serialize as `"[]"`.

### league_guid for migrated seasons

No real league GUID exists in the legacy DB. A new UUID is generated at migration
start and registered as a `franchise_sources` entry with `save_file_path = "(legacy migration)"`,
`season_offset = 0`. All migrated seasons use this synthetic league_guid.

### Award matching

Built-in awards are matched by `PlayerAwards.OriginalName` → `awards.original_name`
(identical across both apps, verified from seed data).

Custom awards (`IsBuiltIn = 0`) that don't match any existing `original_name` are
inserted into `awards` with `is_built_in = 0` using the legacy `Name` for both
`name` and `original_name`.

### Player/team identity for migration

`PlayerGameIdHistory` / `TeamGameIdHistory` track GUIDs across franchise fork
events. For migration into a fresh companion DB:
- First entry (lowest `Id`) → `players.game_guid` / `teams.game_guid`
- Additional entries → `player_alt_guids` / `team_alt_guids`

---

## Data not migrated

| Data | Reason |
|---|---|
| `TeamLogoHistory.LogoFullSize` / `LogoIconSize` | Logo storage not yet in companion schema |
| All derived/rate batting stats | Computed by `v_batting_stats` view on read |
| All derived/rate pitching stats | Computed by `v_pitching_stats` view on read |
| `OpsPlus`, `EraMinus`, `FipMinus`, `Fip` | Phase 8.5 |
| `SeasonTeamHistory.WinPercentage`, `PythagoreanWinPercentage`, etc. | Not in new schema |
| `SeasonTeamHistory.Surplus`, `SurplusPerGame` | Not in new schema |
| `SeasonTeamHistory.GamesBehind` | Not in new schema — `games_back` for standings |
| `ChampionshipWinners` table | Championship implicit via playoff_wins in team_season_history |
| `LookupSeeds` | Internal migration sentinel |
| `PlayerSeasons.ChampionshipWinnerId` | No equivalent; championship tracked by playoff results |
