---
title: CSV Exports
---

# CSV Exports

::: warning Coming soon
CSV exports are not yet available. This page documents the planned feature.
:::

CSV export is an opt-in capability for power users who want to take their franchise data outside of smb-tools — for analysis in a spreadsheet, sharing with the community, or feeding into external tooling. It is not required for any core app functionality.

## Quick Exports

The fastest path to a CSV is exporting directly from a view you are already looking at.

**Leaderboard exports** — any leaderboard view can be exported as-is, respecting whatever filters and column selections are currently active. What you see is what you get in the file.

**Player career stats** — a player's full career stat line can be exported from their profile page.

## Full Season Export

A full season export produces a structured snapshot of an entire season's data in a format compatible with the original SMB3Explorer CSV output. If you have existing spreadsheets, scripts, or community tools built around the SMB3Explorer export format, this export will slot in without changes.

## Custom Export Builder

For anything beyond the quick exports, the custom export builder lets you construct exactly the dataset you want.

You select which tables and relationships to include — player stats, team records, awards, or combinations across them — then choose the specific columns to export. Once configured, smb-tools runs the query and writes the result to a CSV file.

This is the right tool when you want a non-standard slice of your franchise data: a particular subset of columns, a join across multiple stat types, or a shape that doesn't match any of the built-in views.
