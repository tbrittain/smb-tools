---
title: Team Transfer Tool
---

# Team Transfer Tool

::: warning Coming soon
The team transfer tool is not yet available. It is planned as a major feature after the core companion app reaches a stable state. This page documents the intended design.
:::

The team transfer tool is the most-requested community feature for Super Mega Baseball 4 — and the one that neither of the original apps ever supported. It lets you take a team you have built in one franchise and transplant it into a different save game, complete with the full roster, player attributes, salaries, traits, and logos.

The primary use case is community league building: gathering custom teams from different players and assembling them into a shared save game. Historically this has required manual workarounds or external tools with limited reliability. smb-tools is building this as a first-class, integrated feature.

## How It Works

The transfer flow has three steps:

1. **Select the source** — choose a franchise and a specific season snapshot to pull the team from. Because the transfer reads from a snapshot rather than a live save, you are working against a stable, known state.

2. **Select the target** — pick the save game you want to transfer the team into, and whether to add it as a new team or replace an existing one.

3. **Confirm and transfer** — review what will be written and confirm. smb-tools writes the result to a **copy** of the target save. Your original save file is never touched.

## What Gets Transferred

- Full player roster
- Player attributes, salary, and traits
- Team logos

## Compatibility

Transfers are validated before writing. Both the source snapshot and the target save must be SMB4 saves from the same game version. Transferring between versions is not supported.

## Safety Guarantee

The transfer tool is the one part of smb-tools that writes to a save game file. To make this safe, smb-tools always writes to a copy — the original save is never modified in place. The output is a new file you can load separately, verify in-game, and then replace your active save with if everything looks correct.
