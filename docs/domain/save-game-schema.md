# Save Game Schema

The SMB save game is a SQLite 3 database. These are all known tables and views, derived from the queries used by SMB3Explorer. Schema validation checks for the presence of `t_stats` and `t_leagues` as required tables.

---

## League & Franchise Tables

### `t_leagues`

Stores all leagues in the save file.

| Column | Notes |
|--------|-------|
| leagueId | Primary key |
| leagueName | Display name of the league |
| leagueTeamTypeId | References team type; used to distinguish franchise/season/elimination mode |

### `t_team_types`

Lookup table for league team type definitions (used to identify Franchise vs. Season vs. Elimination mode).

| Column | Notes |
|--------|-------|
| leagueTypeName | Name of the mode type |

### `t_franchise`

Franchise records linked to leagues.

| Column | Notes |
|--------|-------|
| franchiseId | Primary key |
| leagueId (FK) | References `t_leagues` |
| playerTeamId | The team ID that the user controls |
| playerTeamName | Display name of the user's team |

### `t_franchise_seasons`

Individual season records within a franchise.

| Column | Notes |
|--------|-------|
| seasonID | Primary key |
| leagueGUID | Historical league identifier (used for multi-league SMB4 saves) |
| seasonNum | Computed via RANK() over franchise seasons |

### `t_franchise_season_creation_params`

Franchise configuration at season creation.

| Column | Notes |
|--------|-------|
| franchiseId (FK) | |
| numGamesInSeason | Season length |
| incomePerTick | Payroll budget tick rate |

---

## Team Tables

### `t_teams`

Team definitions.

| Column | Notes |
|--------|-------|
| teamGUID | GUID identifier |
| teamName | Display name |

### `t_team_local_ids`

Maps local numeric IDs to team GUIDs.

| Column | Notes |
|--------|-------|
| teamLocalId | Numeric local ID |
| teamGUID (FK) | References `t_teams` |

### `t_conferences`

Conference definitions.

| Column | Notes |
|--------|-------|
| conferenceName | Display name |

### `t_divisions`

Division definitions scoped to conferences.

| Column | Notes |
|--------|-------|
| divisionName | Display name |
| conferenceId (FK) | |

### `t_division_teams`

Maps teams to divisions.

---

## Player Tables

### `t_baseball_players`

Core player attribute data.

| Column | Notes |
|--------|-------|
| baseballPlayerGUID | GUID identifier |
| power | 1–99 |
| contact | 1–99 |
| speed | 1–99 |
| fielding | 1–99 |
| arm | 1–99 |
| velocity | 1–99 |
| junk | 1–99 |
| accuracy | 1–99 |
| age | Player age |

### `t_baseball_player_local_ids`

Maps local numeric IDs to player GUIDs.

| Column | Notes |
|--------|-------|
| baseballPlayerLocalId | Numeric local ID |
| baseballPlayerGUID (FK) | |

### `t_baseball_player_traits`

Player traits stored as JSON.

| Column | Notes |
|--------|-------|
| baseballPlayerGUID (FK) | |
| traits | JSON array of `{traitId, subtypeId}` objects; up to 2 traits per player |

### `t_baseball_player_options`

Player options, used for secondary position and pitch repertoire.

| Column | Notes |
|--------|-------|
| baseballPlayerGUID (FK) | |
| optionKey | 55 = secondary position; other keys used for pitch types (SMB4) |
| optionValue | The option value (position code, pitch type code, etc.) |

### `t_salary`

Player salary data.

| Column | Notes |
|--------|-------|
| baseballPlayerGUID (FK) | |
| salary | In game units; multiply by 200 for display value |
| teamId (FK) | The team currently paying this salary |

---

## Statistics Tables

### `t_stats`

Base stats record. All stats records have a corresponding row here.

| Column | Notes |
|--------|-------|
| aggregatorID | Primary key; referenced by all stats tables |
| currentTeamName | Name of current team |
| mostRecentTeamName | Previous team name |
| secondMostRecentTeamName | Second previous team name |

### `t_stats_players`

Player metadata for stats (joined to `t_stats`).

| Column | Notes |
|--------|-------|
| aggregatorID (FK) | |
| statsPlayerID | Secondary identifier |
| baseballPlayerGUIDIfKnown | Player GUID (may be null for historical/unknown players) |
| firstName | |
| lastName | |
| primaryPosition | Position code |
| secondaryPosition | Secondary position code |
| pitcherRole | SP/RP/SP-RP/CL |
| age | At time of stat record |
| retirementSeason | Season number when retired; null if active |

### `t_stats_batting`

Batting statistics (one row per player per stats aggregator).

| Column | Notes |
|--------|-------|
| aggregatorID (FK) | |
| gamesPlayed | |
| gamesBatting | |
| atBats | |
| runs | |
| hits | |
| doubles | |
| triples | |
| homeruns | |
| rbi | |
| stolenBases | |
| caughtStealing | |
| baseOnBalls | Walks |
| strikeOuts | |
| hitByPitch | |
| sacrificeHits | |
| sacrificeFlies | |
| errors | |
| passedBalls | |

### `t_stats_pitching`

Pitching statistics.

| Column | Notes |
|--------|-------|
| aggregatorID (FK) | |
| wins | |
| losses | |
| games | Total appearances |
| gamesStarted | |
| completeGames | |
| totalPitches | |
| shutouts | |
| saves | |
| outsPitched | Multiply by (1/3) for innings pitched display |
| hits | Hits allowed |
| earnedRuns | |
| homeRuns | Home runs allowed |
| baseOnBalls | Walks issued |
| strikeOuts | |
| battersHitByPitch | |
| battersFaced | |
| gamesFinished | |
| runsAllowed | Total runs (earned + unearned) |
| wildPitches | |

### `t_season_stats`

Links stats records to specific franchise seasons.

| Column | Notes |
|--------|-------|
| aggregatorID (FK) | |
| seasonID (FK) | References `t_franchise_seasons` |

### `t_career_season_stats`

Career stat aggregations across seasons within a franchise.

---

## Schedule & Game Tables

### `t_season_schedule`

Game schedule entries.

| Column | Notes |
|--------|-------|
| gameNumber | Global sequential game number |
| day | Day within the season |
| homeTeamID (FK) | Local team ID |
| awayTeamID (FK) | Local team ID |

### `t_game_results`

Results for completed games.

| Column | Notes |
|--------|-------|
| gameNumber (FK) | |
| homeRunsScored | |
| awayRunsScored | |
| homePitcherLocalID | Starting pitcher local ID |
| awayPitcherLocalID | Starting pitcher local ID |

### `t_season_games`

Links games to specific seasons.

| Column | Notes |
|--------|-------|
| gameNumber (FK) | |
| seasonID (FK) | |

### `t_playoffs`

Playoff bracket data.

| Column | Notes |
|--------|-------|
| seasonID (FK) | |
| seriesNumber | Which round/series |
| team1GUID | |
| team2GUID | |
| team1Seed | Playoff seeding |
| team2Seed | |

---

## Views

### `v_season_standings`

Computed standings including wins, losses, runs for/against, games back. Calculated from `t_game_results` joined to team and schedule data.

### `v_baseball_player_info`

Denormalized player information joining `t_baseball_players`, `t_baseball_player_local_ids`, `t_stats_players`, and salary data. Used to simplify player queries.

### `v_playoff_games_won_lost`

Playoff game win/loss calculations, used to compute playoff standings (champion = most wins, runner-up = second most, etc.).
