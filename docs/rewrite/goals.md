# Rewrite Goals

## Why We're Rewriting

### 1. Cross-Platform Compatibility

Both original apps are Windows-only WPF desktop applications. The SMB community spans Windows, macOS, and (increasingly) Linux via Steam. The rewrite targets all three platforms via **Wails** (Go + WebView2/WebKit), which compiles to a native desktop app on each OS.

### 2. Modern Frontend: JavaScript Instead of WPF

WPF is Windows-only and has an increasingly dated developer experience. The rewrite uses **Vue 3 + TypeScript** for the frontend, enabling:
- Richer, more modern UI components (PrimeVue)
- Easier data visualization (compared to ScottPlot/WPF)
- Better developer tooling (Vite, TypeScript, Biome)
- Web-standard CSS and layout primitives

### 3. Consolidated UX: One App Instead of Two

The original two-app pipeline is a significant friction point:
```
User runs SMB3Explorer → exports 8 CSV files → opens SmbExplorerCompanion → imports 8 CSV files
```
This is confusing, error-prone (wrong season, wrong files, wrong order), and requires maintaining two separate applications.

The rewrite collapses this into a single app that reads the SMB save file directly — no CSV intermediary, no two-step import process.

### 4. Direct Save File Reading

The companion app never directly read the SMB save file — it depended on SMB3Explorer's CSV exports. The rewrite implements direct save file reading in Go:
- ZLib decompression (already partially implemented in `backend/smb-connection/`)
- Direct SQLite query against the decompressed database
- No intermediate CSV files

### 5. New Features: Team Transfer Tool

The most-requested community feature that neither original app supported. The team transfer tool would allow users to copy teams (rosters, attributes, salaries) between save files — enabling things like:
- Sharing a custom team with friends
- Importing community-created rosters
- Transferring a franchise team to a new save

This is a write operation against the save file (or rather, creating a modified copy of the save file), which goes beyond what the read-only originals could do.

---

## Tech Stack

| Layer | Technology | Notes |
|-------|-----------|-------|
| Backend language | Go 1.24 | Cross-platform, fast, simple concurrency |
| Desktop framework | Wails v2.10.1 | Bridges Go backend with web frontend |
| Frontend framework | Vue 3 + TypeScript | Component-based, reactive |
| UI components | PrimeVue 4.x | Rich component library with dark theme support |
| State management | Pinia | Vue-idiomatic store |
| Build tool | Vite | Fast HMR dev experience |
| Database (save game) | SQLite via Go driver | Read-only access to decompressed save |
| Linting/formatting | Biome | Replaces ESLint + Prettier |

---

## What Changes

| Area | Original | Rewrite |
|------|----------|---------|
| Platform | Windows only | Windows + macOS + Linux |
| Frontend | WPF (C#) | Vue 3 + TypeScript |
| Backend | C# / .NET 7 | Go 1.24 |
| App count | 2 separate apps | 1 unified app |
| Import workflow | Export CSVs → import CSVs | Direct save file reading |
| New features | None | Team transfer tool |
| Franchise DB | SmbExplorerCompanion.db (EF Core) | TBD (likely SQLite via Go) |

---

## What's Carried Over

All existing functionality from both apps is in scope for the rewrite:

**From SMB3Explorer**:
- ZLib save file decompression
- SQLite read-only access
- League/franchise selection
- All export types (translated to in-app views instead of CSV files)

**From SmbExplorerCompanion**:
- Franchise history database (persistent across seasons)
- Player career and season stats viewer
- Team season detail
- Leaderboards (batting/pitching careers and seasons)
- Hall of Fame management
- Awards tracking
- Percentile rankings

**New in rewrite**:
- Team transfer tool (write operation on save file copy)
- Direct save reading (no CSV intermediary)
- Cross-platform support

---

## Current State of the Rewrite

As of May 2026, `smb-tools` is in early scaffolding:
- Wails project structure established
- Vue 3 + TypeScript + PrimeVue frontend configured
- ZLib decompression partially implemented (`backend/smb-connection/handler.go`)
- App data directory management implemented (`backend/config/app_directories.go`)
- No frontend features implemented yet (still using Wails template)
- No database layer
- No league/player/team data models

The existing backend infrastructure (decompression + file management) provides a foundation, but the majority of the application remains to be built.
