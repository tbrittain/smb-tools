---
title: Save Game Setup & Season Sync
---

# Save Game Setup & Season Sync

smb-tools reads your Super Mega Baseball 4 save file directly — no CSV exports, no intermediate steps. This page covers how to create a franchise, connect a save file, and keep your stats up to date each season.

## Creating a Franchise

When you open smb-tools for the first time, you'll see the franchise selector. Click **New Franchise** to get started.

You'll be asked for two things:

- **Name** — a label for this franchise in smb-tools. It does not have to match the league name in the game.
- **Save File** — the `.sav` file for your franchise. Only franchise mode saves are shown in the picker. smb-tools looks for saves in the default SMB4 location (`%LOCALAPPDATA%\Metalhead\Super Mega Baseball 4\`), so your file should appear automatically.

After selecting a save file, smb-tools shows the league name and your team name as a confirmation. Click **Create Franchise** to finish.

## Connecting and Replacing a Save File

Your connected save file is shown on the **Setup** page under **Connected Save Files**. Each entry shows the filename and which seasons were synced from it.

If you selected the wrong file when creating the franchise, click the **replace** link next to the active source. This swaps the file in-place — use it only to correct a mistake, not to add a second league.

::: warning
Replacing the active source changes which file future syncs read from. It does not delete any previously synced season data.
:::

## Syncing a Season

Go to **Setup** and click **Sync Season**. smb-tools reads the current state of your save file and imports the season.

**Sync timing matters.** A full season in SMB4 has two phases:

1. **After the regular season ends** — sync to capture regular-season stats.
2. **After the playoffs conclude** — sync again to capture playoff results.

::: danger Sync before advancing to the offseason
Advancing to the offseason in-game triggers data compaction that permanently removes per-game stat detail from the save file. Always sync after the playoffs **before** pressing the button to start the offseason. Missing this window means playoff stats for that season cannot be recovered from the save file.
:::

Each sync result shows how many players, teams, and games were imported. If playoff games appear in the count, the playoff sync was successful.

## Reimporting from a Snapshot

Every time you sync, smb-tools saves a snapshot of the raw save file at that moment. If a sync went wrong — or you want to re-run it after the playoffs when you originally synced mid-season — you can roll back using **Reimport Season from Snapshot**.

On the **Setup** page, select a snapshot from the list and click **Reimport Selected Snapshot**. This replaces the season data for that snapshot's season number. Awards for that season are not affected.

Use this to fix a bad sync without losing any other data.
