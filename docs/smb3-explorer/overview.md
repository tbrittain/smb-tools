# SMB3Explorer: Overview

## What It Is

SMB3Explorer is a **data export tool** for Super Mega Baseball 3 and Super Mega Baseball 4. It reads the game's save file (a ZLib-compressed SQLite database), queries the underlying data, and exports it as CSV files for downstream use.

It is the **first step** in the two-app pipeline: its CSV output is the primary data source for SmbExplorerCompanion. The rewrite (`smb-tools`) aims to eliminate this intermediary step by reading the save file directly.

## Primary Purpose

The game tracks rich statistical data in its save file that it never fully exposes in its own UI. SMB3Explorer surfaces that data:
- Per-season batting and pitching statistics for all players
- Career statistics accumulated across franchise seasons
- Team standings and playoff results
- Full player attribute snapshots (Power, Contact, Speed, etc.) with traits and salary

## Platform & Constraints

- **Windows only** (WPF desktop application)
- **.NET 7 Runtime** required
- **64-bit Windows** required
- Must **not be running while the game is active** (the game locks the save file)
- Operates in **read-only mode** — never modifies the save file

## Supported Games

| Game | Support Level |
|------|--------------|
| Super Mega Baseball 3 | Full support |
| Super Mega Baseball 4 | Full support |

## Role in the Ecosystem

```
SMB save file (.sav)
        ↓
  SMB3Explorer
        ↓
  CSV exports (8 file types)
        ↓
  SmbExplorerCompanion
        ↓
  Franchise history database + stats viewer UI
```

The rewrite (`smb-tools`) collapses this pipeline — it reads the save file directly and presents all functionality in a single application.

## Output

All exports are written to: `%LOCALAPPDATA%\SMB3Explorer\`

Exports can be purged via the app's File menu. The exported CSVs are designed to be imported into SmbExplorerCompanion but can also be opened in Excel, analyzed in Python/R, etc.
