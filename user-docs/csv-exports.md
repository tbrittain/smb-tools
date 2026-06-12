---
title: CSV Exports
---

# CSV Exports

::: warning Coming soon
CSV exports are not yet available. This page documents the planned feature.
:::

CSV export is an opt-in feature for people who want to get their franchise data out of smb-tools, whether that's for analysis in a spreadsheet, sharing with the community, or feeding into other tools. Nothing in the core app depends on it.

## Custom Export Builder

At the heart of the export feature is a flexible query builder. Instead of a fixed set of named exports, you describe exactly the dataset you want:

- Pick the tables and relationships you need: player stats, team records, awards, schedule data, or any combination of those.
- Pick the columns. If all you care about is batting average, home runs, and team name, that's all you'll get.
- Run it, and smb-tools writes the result to a CSV file.

The idea is that you shouldn't have to wonder whether smb-tools can export something. If it's in your franchise data, you can get it out.

## Convenience Exports

For the common cases, you'll be able to export straight from the view you're already looking at:

- From a leaderboard, export it exactly as filtered and configured.
- From a player's profile, export their full career stat line.

Both use the same export engine under the hood; they're just shortcuts to it.

## SMB3Explorer-Compatible Season Export

A full season export will match the original SMB3Explorer CSV schema, so if you've got spreadsheets or community tools built around those exports, they should keep working without changes.
