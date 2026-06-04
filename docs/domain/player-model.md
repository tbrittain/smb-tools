# Player Model

## Game Attributes

All attributes are on a **1–99 scale**.

### Hitting Attributes (position players and two-way players)

| Attribute | Description |
|-----------|-------------|
| Power | Home run potential and extra-base hit frequency |
| Contact | Ability to make contact and hit for average |
| Speed | Baserunning and stolen base ability |
| Fielding | Fielding range and defensive skill |
| Arm | Throwing strength (for outfielders and corner infielders) |

### Pitching Attributes

| Attribute | Description |
|-----------|-------------|
| Velocity | Fastball speed; affects power pitches |
| Junk | Breaking ball quality; affects off-speed effectiveness |
| Accuracy | Command and control; affects walk rate |

### Universal Attributes

| Attribute | Description |
|-----------|-------------|
| Fielding | All players have a fielding rating |
| Arm | All players have an arm rating |

Pitchers have all 8 attributes tracked (both hitting and pitching sets), which matters for two-way player traits.

## Positions

### Primary Positions

| Code | Position |
|------|----------|
| P | Pitcher |
| C | Catcher |
| 1B | First Base |
| 2B | Second Base |
| 3B | Third Base |
| SS | Shortstop |
| LF | Left Field |
| CF | Center Field |
| RF | Right Field |

### Secondary Position Groups

These are assigned via `t_baseball_player_options` (optionKey = 55) and represent positional flexibility:

| Code | Meaning |
|------|---------|
| IF | Can play any infield position |
| OF | Can play any outfield position |
| 1B/OF | Can play first base or outfield |
| IF/OF | Can play any infield or outfield |

## Pitcher Roles

| Code | Role |
|------|------|
| SP | Starting Pitcher |
| RP | Relief Pitcher |
| SP/RP | Starter and Reliever (flexible) |
| CL | Closer |

## Handedness

| Type | Values |
|------|--------|
| Bat Hand | R (right), L (left), S (switch) |
| Throw Hand | R (right), L (left) |

## Chemistry Types (SMB4 only)

Each player in SMB4 belongs to one of five chemistry types, which determines which traits they can have:

| Chemistry | Archetype |
|-----------|-----------|
| Competitive | Aggressive, win-at-all-costs mentality |
| Spirited | Energetic, team morale-oriented |
| Disciplined | Patient, process-focused |
| Scholarly | Analytical, situationally aware |
| Crafty | Deceptive, exploits opponent tendencies |

Chemistry types affect which traits are available to a player and factor into team chemistry dynamics.

## Salary

Salary is stored in the save game in **game units**. To convert to the display value (as shown in the game's UI):

```
display_salary = game_units × 200
```

Salary data lives in the `t_salary` table, linked to players and teams.

## Age and Retirement

- Players have an `age` attribute tracked in `t_baseball_players` and `t_stats_players`
- `t_stats_players` includes a `retirementSeason` field indicating when a player retired
- Once retired, seasonal stat data for that player becomes unavailable in the save file — this is why the Companion app must capture data before each player retires

## Two-Way Players

A player can be a two-way player (pitcher who also bats), indicated by a specific trait in SMB4. Two-way players have meaningful ratings for both hitting and pitching attributes.

---

## Source Files

**SMB3Explorer** (https://github.com/tbrittain/SMB3Explorer):
- `SMB3Explorer/Resources/Sql/MostRecentSeasonPlayersSmb3.sql` — all player attribute columns for SMB3
- `SMB3Explorer/Resources/Sql/MostRecentSeasonPlayersSmb4.sql` — adds chemistry, handedness, pitch repertoire for SMB4
- `SMB3Explorer/Models/Exports/SeasonPlayer.cs` — `SeasonPlayer` export model showing all player fields

**SmbExplorerCompanion** (https://github.com/tbrittain/SmbExplorerCompanion):
- `SmbExplorerCompanion.Database/Entities/Player.cs` — core player entity with FK references to lookups
- `SmbExplorerCompanion.Database/Entities/PlayerSeasonGameStat.cs` — per-season attribute snapshot
- `SmbExplorerCompanion.Database/Entities/Lookups/Position.cs` — position definitions with `IsPrimaryPosition` flag
- `SmbExplorerCompanion.Database/Entities/Lookups/Chemistry.cs` — chemistry type definitions
- `SmbExplorerCompanion.Database/Entities/Lookups/PitcherRole.cs` — pitcher role definitions
