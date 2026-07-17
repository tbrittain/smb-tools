# smb-tools Documentation

Knowledge base for the `smb-tools` rewrite — a consolidated, cross-platform successor to SMB3Explorer and SmbExplorerCompanion. Use this as the authoritative reference during development; there should be no need to dig through the original C# repositories.

## Structure

### `domain/` — Game and data model knowledge

Foundational knowledge about Super Mega Baseball itself and the data structures it uses. Read these first.

| File | Contents |
|------|----------|
| [game-overview.md](domain/game-overview.md) | SMB3/SMB4, game modes (Franchise/Season/Elimination), key differences between versions |
| [save-game-format.md](domain/save-game-format.md) | ZLib+SQLite format, default file locations, decompression approach, constraints |
| [save-game-schema.md](domain/save-game-schema.md) | Every table and view in the SMB save game SQLite database |
| [player-model.md](domain/player-model.md) | Player attributes, positions, handedness, pitcher roles, chemistry |
| [player-stats.md](domain/player-stats.md) | Every batting and pitching stat tracked, including derived/advanced metrics |
| [player-traits.md](domain/player-traits.md) | SMB3 trait list (20), SMB4 trait list (80+), chemistry system, save game encoding |
| [pitch-types.md](domain/pitch-types.md) | The 8 pitch types with abbreviations and context |
| [awards.md](domain/awards.md) | All 30+ awards with importance tiers, categories, and Hall of Fame criteria |

### `smb3-explorer/` — SMB3Explorer app documentation

Documents the first original app: a Windows export tool that reads SMB save files and produces CSV reports.

| File | Contents |
|------|----------|
| [overview.md](smb3-explorer/overview.md) | Purpose, audience, platform constraints, role in the ecosystem |
| [user-features.md](smb3-explorer/user-features.md) | Complete user flow and every export type |
| [technical-architecture.md](smb3-explorer/technical-architecture.md) | C#/.NET 7, WPF MVVM, DataService pattern, SQLite access |
| [csv-export-schemas.md](smb3-explorer/csv-export-schemas.md) | Column-by-column schema for every CSV file the app produces |

### `smb-explorer-companion/` — SmbExplorerCompanion app documentation

Documents the second original app: a Baseball Reference–style franchise history viewer that imports SMB3Explorer CSV exports.

| File | Contents |
|------|----------|
| [overview.md](smb-explorer-companion/overview.md) | Purpose, dependency on SMB3Explorer, limitations, advanced usage |
| [user-features.md](smb-explorer-companion/user-features.md) | Every screen, filter, visualization, and workflow |
| [technical-architecture.md](smb-explorer-companion/technical-architecture.md) | C#/.NET 7, WPF MVVM, CQRS+MediatR, EF Core, ScottPlot |
| [companion-db-schema.md](smb-explorer-companion/companion-db-schema.md) | Full schema of the companion's own SQLite database |
| [import-flow.md](smb-explorer-companion/import-flow.md) | End-to-end CSV import wizard flow |

### `rewrite/` — Rewrite context

Why we're rewriting and what we're building toward.

| File | Contents |
|------|----------|
| [goals.md](rewrite/goals.md) | Motivations, tech stack choices, what's new vs. carried over |

### `game-integration/` — Game internals investigation

| File | Contents |
|------|----------|
| [investigation.md](game-integration/investigation.md) | Engine identification, bundled SQLite schema discovery, auto-sync strategy |

### Root-level

| File | Contents |
|------|----------|
| [companion-issues.md](companion-issues.md) | All open GitHub issues from SmbExplorerCompanion with space for rewrite notes |
| [data-export-roadmap.md](data-export-roadmap.md) | Three-phase plan for the Data Export feature: MVP (done), flexible user filters, full 8-dataset catalog |

## Reading Order

For a new developer on this project:

1. `domain/game-overview.md` — understand the game and its modes
2. `domain/save-game-format.md` + `domain/save-game-schema.md` — understand the raw data source
3. `domain/player-model.md` + `domain/player-stats.md` — understand the player data model
4. `smb3-explorer/user-features.md` — understand what the first app does
5. `smb-explorer-companion/user-features.md` — understand what the second app does
6. `rewrite/goals.md` — understand what we're building
