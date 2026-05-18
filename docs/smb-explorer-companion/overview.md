# SmbExplorerCompanion: Overview

## What It Is

SmbExplorerCompanion is a **franchise history viewer and statistics tracking application** for Super Mega Baseball 4. Think Baseball Reference, but for your in-game franchise. It stores and visualizes the statistical history that the game itself discards or hides.

## Why It Exists

The game has two critical gaps:
1. **50-season history limit**: franchise history is purged after approximately 50 seasons
2. **No historical leaderboards**: there is no in-game way to see who led the franchise in career home runs, ERA, etc.

SmbExplorerCompanion solves both by maintaining its own persistent SQLite database, populated from CSV exports produced by SMB3Explorer.

## Primary Data Source

The companion app does **not read the SMB save file directly**. It depends entirely on CSV files exported by SMB3Explorer. This means the workflow is:

```
SMB4 save file → SMB3Explorer → CSV exports → SmbExplorerCompanion import wizard → Companion DB → Stats UI
```

Direct save file reading was planned for a future version of the companion but never implemented — it is a core goal of the `smb-tools` rewrite.

## Supported Games

- **SMB4**: primary and only fully supported game
- **SMB3**: not prioritized; some infrastructure exists (`IsSmb3` flag on Franchise) but not developed

## Key Capabilities

- Career and season statistics for all players across the franchise's entire history
- Player attribute tracking (Power, Contact, etc.) per season — watch how players develop
- Hall of Fame management (user-controlled eligibility and induction)
- Award tracking (MVP, Cy Young, Gold Glove, etc.) — manually assigned each season
- Automatic title awards (Batting Title, HR Title, RBI Title, ERA Title, Wins Title, K Title) — computed from stats
- Triple Crown detection (batting and pitching) — auto-calculated
- Team season history with schedule breakdowns and performance trends
- Percentile rankings for player attributes and statistical KPIs within the franchise
- "Similar players" recommendations

## Multi-Franchise Support

The companion supports multiple franchises in a single installation, enabling a "manager mode" where one user tracks multiple independent franchises.

## Advanced Usage

Because the companion's database is a standard SQLite file, power users can run custom SQL queries for analysis beyond the UI. The database is located at:

`%LOCALAPPDATA%\SmbExplorerCompanion\SmbExplorerCompanion.db`

The project maintains a companion scripts repository with example queries (30-30 club identification, league ERA trends over time, position players who pitched, etc.).

## Limitations

- **Player names are locked** after the first import for a player — they cannot be changed within the app
- **Team names can be updated** within the franchise, but division/conference names cannot be renamed
- **Re-introduced free agents** may create duplicate player records if the game reuses player IDs with name mismatches
- **Windows only** (WPF)
- **.NET 7 Runtime** required
