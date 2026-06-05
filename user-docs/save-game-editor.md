---
title: Save Game Editor
---

# Save Game Editor

::: warning Coming soon
The save game editor is planned as a later phase of the team transfer feature, after league export and import is stable. This page documents the intended direction.
:::

Where league export and import is about moving a complete league from one place to another, the save game editor is about making changes to a league. It connects directly to the save game database and writes modifications back — not just packaging existing data, but altering it.

## Planned Capabilities

**Player editing** — update player attributes, traits, and other properties across a roster or an entire league. The kind of large-scale changes that would take hours to make manually through the in-game menus.

**Team splicing** — take a team from one league and insert it into a different save game, replacing an existing team or slotting in alongside them. This builds on the export/import foundation but at the team level, with full control over placement.

**Bulk roster operations** — apply changes across multiple players or teams at once. Useful when building a custom league from scratch and needing to adjust a large number of players to hit a particular balance or theme.

## How It Differs from League Export & Import

League export and import is read-then-write at the file level — it reads a snapshot and produces a new save. The save game editor is a live connection to a save file that supports arbitrary modifications.

This distinction matters for safety. The editor will always work on a copy of the target save, and smb-tools will provide clear confirmation of what will be written before any changes are committed.

## Scope

The save game editor operates on SMB4 saves only. It will not support modifying the internal franchise history database that smb-tools maintains — only the SMB4 save game files themselves.
