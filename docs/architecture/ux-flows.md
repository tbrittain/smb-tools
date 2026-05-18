# UX Flows

Core user-facing flows for the rewrite. These describe what the user experiences, not implementation detail. Use these as the source of truth when designing the frontend and when the Go service layer needs to know what sequence of operations a given user action triggers.

---

## Guiding Principles

- **One app.** There is no separate export tool. No CSV intermediary. No two-step pipeline. The user opens smb-tools, connects to their save file, and the app does the work.
- **The sync is the import.** Bringing new season data into the app is a single user action — not a multi-file wizard.
- **CSV export is opt-in, not the pipeline.** Exporting CSVs for use in Excel, sharing with others, or external analysis is a secondary feature. It is not required for any core app functionality.

---

## Core Flow: Sync a Season

This replaces the entire SMB3Explorer → 8 CSV exports → SmbExplorerCompanion import wizard pipeline with a single action.

```
User opens app
        ↓
Selects active franchise (or it's pre-selected from last session)
        ↓
App shows "Last synced: Season 7 · 3 days ago" (or "Never synced")
        ↓
User clicks "Sync" (or similar)
        ↓
App locates the .sav file (from remembered path, or prompts once if unknown)
        ↓
Decompresses .sav → raw SQLite bytes
        ↓
Hash the decompressed bytes
        ↓
If hash matches the most recent snapshot for this franchise: show "Already up to date"
If hash differs: persist as a new snapshot (see snapshot-strategy.md)
        ↓
Read from the decompressed save game DB:
  - Current season players (attributes, traits, salary, pitch types)
  - Current season teams (standings, budget, payroll)
  - Regular season schedule + results
  - Playoff schedule + results
  - Career stats for all active players
        ↓
Write to the franchise's companion DB via the store layer
        ↓
App refreshes to show updated data
        ↓
"Last synced: Season 8 · just now"
```

**When to sync**: The user should sync at the end of each in-game season, before simulating the offseason. The app surfaces the last-synced state prominently so the user knows if they are behind. The app does not auto-sync; the trigger is always manual.

**Why manual**: Auto-sync mid-season would capture partial season data. The user controls when data is committed to the franchise DB.

---

## First-Time Setup Flow

The first time a user opens the app, or creates a new franchise:

```
App opens
        ↓
No franchises exist → prompt: "Create your first franchise"
        ↓
User provides:
  - Franchise name
  - Game version (SMB3 / SMB4)
  - Path to their .sav file (or the app auto-detects from known default locations)
        ↓
App creates a new franchise DB (see data-layer.md)
        ↓
App prompts: "Sync now?" → goes to Sync flow above
```

For SMB4, where a single save file can contain multiple leagues, the user selects which league within the save file this franchise corresponds to.

---

## Franchise Switching

```
User selects a different franchise from the sidebar / switcher
        ↓
App closes connection to current franchise DB
        ↓
App opens connection to selected franchise DB
        ↓
All views reload with data from the new franchise DB
```

Each franchise is completely isolated. Switching franchises is instantaneous from the user's perspective — there is no loading/import step.

---

## CSV Export (Secondary / Opt-In)

CSV export is available but is not part of the core flow. It is a "power user" or "sharing" feature.

Where it surfaces:
- A dedicated "Export" section or menu
- Possibly contextual: "Export this leaderboard to CSV", "Export this player's career stats"

What can be exported:
- Any leaderboard view
- A player's full career and season stats
- A team's season history
- Full season batting/pitching stats (replicating the original SMB3Explorer export format for compatibility with external tools)

CSV export never needs to happen for the app to function. A user who only uses the in-app views never needs to touch this feature.

---

## Legacy Migration Flow (Existing Companion Users)

For users migrating from SmbExplorerCompanion:

```
User opens app for the first time
        ↓
App detects existing SmbExplorerCompanion.db at known path
  (or user browses to it manually)
        ↓
App presents: "We found data from a previous version. Import it?"
        ↓
User confirms
        ↓
App reads old SmbExplorerCompanion.db (read-only)
App creates a new franchise DB for each franchise found in the old DB
App runs the legacy migration service: old schema → new schema
        ↓
Success: "X seasons of data imported for Y franchise(s)"
User is placed in the normal app experience with their data intact
```

This flow is separate from the normal sync flow. It is a one-time migration, not a recurring operation. See `data-layer.md` for the technical implementation.

---

## Open Questions (UX)

These need resolution before the frontend is built:

- What is the exact label / verb for the sync action? "Sync", "Import Season", "Update", "Connect"?
- How does the app present the list of franchises (sidebar, dropdown, dedicated screen)?
- What does the app show when no data exists yet for a franchise (empty states)?
- Should the app warn the user if the save file appears to be mid-season (i.e., games are still in progress)?
- For SMB4 multi-league saves: does "Sync" import only the selected franchise's league, or all leagues in the save file simultaneously?
