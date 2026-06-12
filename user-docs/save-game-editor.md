---
title: Save Game Editor
---

# Save Game Editor

::: warning Coming soon
The save game editor is planned as a later phase of the team transfer feature, after league export and import is stable. This page documents the intended direction.
:::

League export and import moves a complete league from one place to another. The save game editor goes further: it connects directly to the save game database and writes changes back, altering the data rather than just repackaging it.

## Planned Capabilities

**Player editing.** Update attributes, traits, and other properties across a roster or an entire league, the kind of large-scale change that would otherwise take hours through the in-game menus.

**Team splicing.** Take a team from one league and drop it into a different save game, either replacing an existing team or adding it alongside the others. It builds on the export/import groundwork, but at the team level, with full control over placement.

**Bulk roster operations.** Apply changes across multiple players or teams at once, which will be handy when you're building a custom league from scratch and need to nudge a lot of players toward a particular balance or theme.

## How It Differs from League Export & Import

League export and import is read-then-write at the file level: it reads a snapshot and produces a new save. The save game editor is a live connection to a save file that supports arbitrary modifications.

This distinction matters for safety. The editor will always work on a copy of the target save, and smb-tools will give you a clear summary of what's about to be written before any changes are committed.

## Scope

The save game editor operates on SMB4 saves only. It won't touch the internal franchise history database that smb-tools maintains, just the SMB4 save game files themselves.
