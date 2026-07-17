---
title: Save Game Setup & Season Sync
---

# Save Game Setup & Season Sync

smb-tools reads your Super Mega Baseball 4 save file directly, so there's no CSV exporting and no extra steps in between. This page covers how to create a franchise, connect a save file, and keep your stats up to date each season.

## Creating a Franchise

When you open smb-tools for the first time, you'll see the franchise selector. Click **New Franchise** to get started.

You'll be asked for two things:

- **Name**: a label for this franchise in smb-tools. It doesn't have to match the league name in the game.
- **Save File**: the `.sav` file for your franchise. smb-tools supports both **Franchise Mode** and **Season Mode** saves — see [Season Mode](season-mode.md) if you play Season Mode. It looks specifically for files named `league-<id>.sav` and filters out auxiliary files like `master.sav`, `mugshots-*.sav`, and `season-*.sav`.

On Windows, smb-tools looks in the default SMB4 location (`%LOCALAPPDATA%\Metalhead\Super Mega Baseball 4\`) and your file should appear automatically. On macOS and Linux, whether smb-tools can find it automatically depends on how you're running the game. See [Finding Your Save File on macOS and Linux](#finding-your-save-file-on-macos-and-linux) below.

If your file doesn't appear in the picker, click **Browse for file…** to open a native file picker and select the `.sav` file directly, wherever it lives on disk.

After selecting a save file, smb-tools shows the league name and your team name as confirmation. Click **Create Franchise** to finish.

## Finding Your Save File on macOS and Linux

smb-tools itself is a cross-platform app: it runs natively on Windows, macOS, and Linux. **Super Mega Baseball 4 is not.** It's a Windows/Steam title with no native Mac or Linux build, so on those platforms you're running it through some kind of Windows compatibility layer, and the save file ends up wherever that layer's virtual filesystem puts it.

### Linux: Steam Play / Proton

The most common way to play SMB4 on Linux (including Steam Deck) is through **Steam Play with Proton**, Steam's built-in Windows compatibility layer. Proton runs the game inside a per-game compatibility prefix that mirrors a Windows filesystem layout, and the game writes its saves there exactly as it would on Windows.

In practice, this means your save files typically end up at:

```
~/.local/share/Metalhead/Super Mega Baseball 4/<steam-id>/
```

smb-tools knows about this convention and automatically searches it alongside the standard Windows location, so your save file should appear in the picker without any extra setup, even though SMB4 isn't a native Linux game. If you've moved your Steam library to a non-default location, or you're running SMB4 through a standalone Wine/Proton prefix outside of Steam, auto-discovery may not find it. Use **Browse for file…** to point smb-tools at the prefix's `drive_c` directory directly.

### macOS: compatibility layers (CrossOver, etc.)

There's no confirmed native macOS release of SMB4, and Apple platforms don't support Proton. Players running the game on a Mac typically use a third-party Windows compatibility layer such as **CrossOver** (CodeWeavers' Wine-based tool), or run Windows itself in a virtual machine.

These setups don't follow one consistent save location the way Proton does on Linux, so smb-tools doesn't try to auto-discover save files on macOS, and the picker will come up empty. That's expected, not a bug. Click **Browse for file…** and navigate into your compatibility layer's virtual Windows drive (for example, CrossOver bottles expose a `drive_c` folder) to the same path the game would use on Windows:

```
.../drive_c/users/<you>/AppData/Local/Metalhead/Super Mega Baseball 4/
```

Once you've found the `league-<id>.sav` file for your franchise, select it and smb-tools will read it the same way it does on any platform.

## Connecting and Replacing a Save File

Your connected save file is shown on the **Setup** page under **Connected Save Files**. Each entry shows the filename and which seasons were synced from it.

If you selected the wrong file when creating the franchise, click the **replace** link next to the active source. This swaps the file in place; use it only to correct a mistake, not to add a second league.

::: warning
Replacing the active source changes which file future syncs read from. It does not delete any previously synced season data.
:::

## Syncing a Season

Go to **Setup** and click **Sync Season**. smb-tools reads the current state of your save file and imports the season.

A full season in SMB4 has two phases, and timing matters:

1. Sync after the regular season ends to capture regular-season stats.
2. Sync again after the playoffs conclude to capture playoff results.

::: danger Sync before advancing to the offseason
Advancing to the offseason in-game triggers data compaction that permanently removes per-game stat detail from the save file. Always sync after the playoffs **before** pressing the button to start the offseason. Miss this window and the playoff stats for that season can't be recovered from the save file.
:::

Each sync result shows how many players, teams, and games were imported. If playoff games appear in the count, the playoff sync was successful.

## Reimporting from a Snapshot

Every time you sync, smb-tools saves a snapshot of the raw save file at that moment. If a sync went wrong, or you want to re-run one after the playoffs when you originally synced mid-season, you can roll back using **Reimport Season from Snapshot**.

On the **Setup** page, select a snapshot from the list and click **Reimport Selected Snapshot**. This replaces the season data for that snapshot's season number. Awards for that season aren't affected.

Use this to fix a bad sync without losing any other data.
