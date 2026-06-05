---
title: CSV Exports
---

# CSV Exports

::: warning Coming soon
CSV exports are not yet available. This page documents the planned feature.
:::

CSV export is an opt-in capability for power users who want to take their franchise data outside of smb-tools — for analysis in a spreadsheet, sharing with the community, or feeding into external tooling. It is not required for any core app functionality.

## Custom Export Builder

The core of the export feature is a flexible query builder. Rather than offering a fixed set of named exports, smb-tools lets you describe exactly the dataset you want:

- **Choose your tables and relationships** — player stats, team records, awards, schedule data, or combinations across them.
- **Select the columns you want** — include only what is relevant to your analysis. If you want batting average, home runs, and team name but nothing else, you get exactly that.
- **Run and export** — smb-tools executes the query against your franchise data and writes the result to a CSV file.

The goal is that you should never need to ask "can smb-tools export X?" — if the data is in your franchise, you can get it out.

## Convenience Exports

For common cases, a one-click export will be available directly from views you are already looking at:

- **Leaderboards** — export any leaderboard as currently filtered and configured.
- **Player career stats** — export a player's full career stat line from their profile.

These are shortcuts built on the same underlying export engine, not a separate system.

## SMB3Explorer-Compatible Season Export

A full season export will produce output in a format compatible with the original SMB3Explorer CSV schema. If you have existing spreadsheets or community tools built around SMB3Explorer exports, this will slot in without changes.
