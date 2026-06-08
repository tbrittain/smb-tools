---
title: Importing from SmbExplorerCompanion
---

# Importing from SmbExplorerCompanion

If you tracked your franchise history with **SmbExplorerCompanion** (the predecessor to smb-tools), you can bring that data into smb-tools in a few steps. The import reads your existing `SmbExplorerCompanion.db` file and creates one or more franchises in smb-tools — your original data is never modified.

## Starting the Import

From the franchise selector screen, click **Import from SmbExplorerCompanion**. smb-tools will look for your database file in the default location. If it finds one, it will show a banner with the detected path and a **Use this file** button.

If your file is in a different location, click **Browse for SmbExplorerCompanion.db…** and locate it manually.

## Selecting Franchises

After the database is loaded, you'll see a list of every franchise it contains. Each entry is labelled **SMB3** or **SMB4** depending on which game it originated from.

Check the franchises you want to import. You can select as many as you like — each one becomes a separate franchise in smb-tools.

## Renaming Before Import

On the next screen, you can confirm or change the name of each franchise. The fields are pre-filled with the names from your SmbExplorerCompanion database. Edit any that you want to rename.

## Reviewing and Confirming

The confirmation screen lists what will be created. Review the names and game versions, then click **Import** to proceed.

The import runs sequentially. When it finishes, you'll see a summary for each franchise showing how many seasons, teams, players, and awards were migrated.

## After Import: Connecting a Save File

A freshly-imported legacy franchise contains historical data but has no live save file connected to it. To sync future seasons, you need to link a save file.

Open the franchise and go to **Setup**. Because this franchise came from a legacy import, smb-tools needs to know how to number the seasons from the new save game. When you select the save file, it will show you how many seasons it contains and ask you to confirm which franchise season that corresponds to.

For example: if your SmbExplorerCompanion franchise had 10 seasons and your new save game currently contains 3 more, you would confirm that season 3 of the save game corresponds to franchise season 13. From that point forward, synced seasons are numbered sequentially.
