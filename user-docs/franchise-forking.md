---
title: Franchise Forking
---

# Franchise Forking

SMB4 lets you export an existing franchise to a brand new league at any point, which effectively continues your franchise history in a fresh save game. smb-tools calls this a **fork**.

Without forking, smb-tools would treat the new save as an unrelated franchise. Connecting it as a fork source tells smb-tools to number seasons from the new save sequentially after your last synced season, keeping your complete franchise history in one place.

::: info Advanced feature
Forking is labeled **Advanced** in the Setup page because it involves a season offset that you need to get right. Read through this page before using it.
:::

## When to Use Forking

Fork a franchise when you have used SMB4's built-in franchise export to start a new league and you want smb-tools to treat the new save as a continuation of your existing franchise rather than a new one.

Don't use forking just to fix a wrong file selection. Use the **replace** option on the active source instead.

## How It Works

Each save file you connect to a franchise is called a **source**. A franchise can have multiple sources over its lifetime. When smb-tools syncs a season, it reads from the most recently added real source and assigns season numbers based on that source's offset.

When you add a fork source, you set an **offset**: the number of seasons that came before it. Seasons from the fork source are then numbered starting at offset + 1.

## Setting Up a Fork

On the **Setup** page, scroll to **Fork Franchise From New Save Game** and click **Add Fork Source**.

You'll see two fields:

- **Season offset**: pre-filled with your franchise's last synced season number. In almost all cases you should leave this as-is. It tells smb-tools that the fork source picks up directly after your last season.
- **Save file**: the new `.sav` file for your continued franchise.

After selecting the save file, smb-tools confirms the connection. Future syncs will read from this new source, and seasons will be numbered from the offset forward.

## Checking Your Source History

The **Connected Save Files** list on the Setup page shows all sources for the current franchise in order, along with the season ranges each one covers. After adding a fork, the new source will appear at the bottom of the list with no seasons yet (that will populate once you sync).

## Forking After a Legacy Migration

If your franchise was originally imported from SmbExplorerCompanion and you are continuing it in SMB4, the first save file you connect goes through a similar season-mapping step. See [Importing from SmbExplorerCompanion](./legacy-migration) for details on that flow. Once the first real source is connected, you can add fork sources the same way as any other franchise.

## Known Limitations

Career batting/pitching qualification thresholds (the minimums used to decide who counts as a "qualified" career leader or record holder) are calibrated to your franchise's **first season's** games-per-season and innings-per-game. If you fork into a save with a different season length or a different number of innings per game, smb-tools does not re-scale these thresholds — they stay fixed at whatever your first season used. This is a known, accepted limitation rather than a bug.
