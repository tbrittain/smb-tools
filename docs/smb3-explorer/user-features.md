# SMB3Explorer: User Features

## User Flow

1. **Launch the app** → landing screen prompts game selection
2. **Select game**: SMB3 or SMB4
3. **Select save file**: app auto-detects the default location; manual override available; pre-decompressed `.sqlite` files also accepted
4. **Select league**: list of available leagues/franchises from the save file, with access history (first accessed, last accessed, access count) shown for previously opened saves
5. **Export data**: choose from the export menu; files are written to `%LOCALAPPDATA%\SMB3Explorer\`

For SMB4, the league GUID is parsed from the save filename to distinguish leagues when multiple exist in a single save database.

## Export Categories

### Franchise Career Exports

Stats accumulated across all seasons in the franchise. Only includes **currently active players** (retired player career data is not retrievable from the save file after retirement).

| Export | Description |
|--------|-------------|
| Career Batting Stats (Regular Season) | Career batting statistics for all active players, regular season only |
| Career Batting Stats (Playoffs) | Career playoff batting statistics for all active players |
| Career Pitching Stats (Regular Season) | Career pitching statistics for all active pitchers, regular season only |
| Career Pitching Stats (Playoffs) | Career playoff pitching statistics for all active pitchers |

### Franchise Season Exports

Per-season statistics across all seasons in the franchise.

| Export | Description |
|--------|-------------|
| Season Batting Stats (Regular Season) | Per-season batting stats for each player, regular season |
| Season Batting Stats (Playoffs) | Per-season playoff batting stats |
| Season Pitching Stats (Regular Season) | Per-season pitching stats, regular season |
| Season Pitching Stats (Playoffs) | Per-season playoff pitching stats |
| Season Standings | Team standings for each season (wins, losses, run differential, etc.) |
| Playoff Standings | Team playoff results for each season (ranked by wins; top = champion) |

### Current Season Exports

Data from the most recent season in the franchise. These exports include richer data than the season-aggregate exports (full player attributes, salary, traits).

| Export | Description |
|--------|-------------|
| Most Recent Season Players | All players with full attributes (Power, Contact, Speed, Fielding, Arm, Velocity, Junk, Accuracy), salary, traits, pitch types (SMB4), handedness (SMB4), chemistry (SMB4) |
| Most Recent Season Teams | All teams with aggregate attributes, budget, payroll, standings data |
| Most Recent Season Schedule | Full game schedule with scores and starting pitchers |
| Most Recent Season Playoff Schedule | Playoff bracket games with series numbers, seeds, scores, pitchers |

### Top Performers Exports

Leaderboard-style exports highlighting the best performers.

| Export | Description |
|--------|-------------|
| Top Batting Performers (Regular Season) | Top batting performers from the most recent regular season |
| Top Batting Performers (Playoffs) | Top batting performers from the most recent playoffs |
| Top Pitching Performers (Regular Season) | Top pitching performers from the most recent regular season |
| Top Pitching Performers (Playoffs) | Top pitching performers from the most recent playoffs |
| Top Rookie Batters | Best batting rookies from the most recent season |
| Top Rookie Pitchers | Best pitching rookies from the most recent season |

## Data Ordering

- **Season standings**: sorted by season, then wins, then division, then run differential
- **Playoff standings**: ranked by wins (rank 1 = champion, rank 2 = runner-up, etc.)
- **Player stats**: sortable by games played (batters) or total pitches (pitchers) — this determines "most active" ordering

## Important Limitation: Retired Players

Once a player retires in the game, their **seasonal statistics are no longer accessible** from the save file. Career totals for retired players may still be available, but per-season breakdowns are lost. This is the primary reason SmbExplorerCompanion exists — to capture and persist this data before retirement occurs.

**Workflow implication**: users should export data every season, before simulating the offseason that triggers retirements.

## Other UI Features

- **Automatic update checking** (v1.1.0+)
- **File menu → Purge**: delete all previously exported CSVs from the output directory
- **Previously accessed leagues**: the app remembers leagues and shows access metadata (count, first/last access time) for convenience
