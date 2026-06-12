---
title: Getting Started
---

# Getting Started

smb-tools is a franchise history tracker and stat viewer for Super Mega Baseball 4. Think Baseball Reference, but for the franchise you have been playing with but with lots of additional features for maximum immersion: career leaderboards, season-by-season stat lines, Hall of Fame management, the complete picture that you can't get through the in-game screens.

## Where It Comes From

Two apps came before this one.

**[SMB3Explorer](https://github.com/tbrittain/SMB3Explorer)** was a data extraction tool. Super Mega Baseball stores rich per-season statistics in its save file such as batting averages, ERA, player attribute snapshots, schedule results, but does not expose many of these . SMB3Explorer cracked the save file open and exported everything as CSV files so players could actually use it.

**[SmbExplorerCompanion](https://github.com/tbrittain/SmbExplorerCompanion)** was the franchise history viewer built on top of those CSV exports. It maintained its own database of franchise history across seasons, kept stats well beyond the game's built-in 50-season limit, and presented everything through a Baseball Reference-style interface.

Together they worked, but the workflow was cumbersome: run SMB3Explorer to produce CSV exports, import those CSVs into SmbExplorerCompanion, then view the stats. Two separate Windows-only applications, both requiring the .NET 7 runtime, with a manual multi-step handoff every time you wanted to sync a new season.

smb-tools is a re-envisioning of both. It reads your save file directly, meaning no CSV exports, no separate export step, and no intermediary files. Syncing a season is a single button click and is significantly faster in itself. And the smb-tools app itself runs on Windows, macOS, and Linux.

::: tip Playing on macOS or Linux?
smb-tools runs natively on all three platforms, but Super Mega Baseball 4 is a Windows/Steam title with no native Mac or Linux release, so where your save file lives and how you point smb-tools at it depends on how you're running the game. See [Finding Your Save File on macOS and Linux](./save-game-setup#finding-your-save-file-on-macos-and-linux) for specifics.
:::

## What It Does

**Franchise stat tracking** is the core of the app. After a brief first-time setup, you connect smb-tools to your SMB4 franchise save and sync each season as you finish it. The app builds a permanent record of everything the game would eventually discard: career statistics, season-by-season breakdowns, player development over time, award history, Hall of Fame, and franchise leaderboards going back as far as you have been playing.

**The team transfer tool** — coming in a later release — extends the app into territory neither original app ever touched: packaging up entire leagues and moving teams between save games, with a full save game editor for making large-scale changes to rosters and attributes. This is the most-requested community feature for SMB4, and smb-tools is building it as a first-class part of the app.

## A Note on Game Version

smb-tools currently works with **Super Mega Baseball 4** save files. If you tracked your franchise history in SmbExplorerCompanion, you can bring that data forward — see [Importing from SmbExplorerCompanion](./legacy-migration).
