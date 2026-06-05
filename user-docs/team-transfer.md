---
title: League Export & Import
---

# League Export & Import

::: warning Coming soon
This feature is not yet available. This page documents the planned design.
:::

The league export and import tool lets you package up an entire SMB4 league — every team and every player — and hand it to someone else so they can load it into their own game. This is the foundation of community league building: one person (or a group) assembles a custom league, exports it, and distributes it so others can play with the same teams.

After import, the recipient's game treats it as a normal franchise save. They can simulate seasons, make trades, run the league however they like — smb-tools is only involved in the transfer itself.

## Exporting a League

Exporting reads from a **snapshot** of one of your franchises. Because the export is based on a snapshot rather than the live save file, you are always working from a stable, known state — the export is reproducible and unaffected by anything that happens in-game after the snapshot was taken.

The export package includes:

- All teams in the league
- Full player rosters (every team)
- Player attributes, salaries, and traits
- Team logos

## Importing a League

Importing takes an export package and writes it into a new SMB4 save game. The result is a playable franchise save that can be opened directly in the game.

smb-tools always writes to a **copy**. Your existing save files are never modified.

## Compatibility

Both export and import require SMB4. The source snapshot and the target save must be from the same game version.
