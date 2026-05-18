# Pitch Types

SMB features 8 distinct pitch types. Each pitcher can have a repertoire of **up to 5 pitch types**.

## Pitch Type Reference

| Abbreviation | Full Name | Category |
|---|---|---|
| 4F | 4-Seam Fastball | Power |
| 2F | 2-Seam Fastball | Power / Movement |
| SB | Screwball | Off-Speed / Breaking |
| CH | Changeup | Off-Speed |
| FK | Forkball | Off-Speed |
| CB | Curveball | Breaking |
| SL | Slider | Breaking |
| CF | Cutter | Power / Breaking |

## Relevance to Attributes

- **Velocity** primarily governs power pitches (4F, 2F, CF)
- **Junk** primarily governs breaking and off-speed pitches (SB, CH, FK, CB, SL)
- **Accuracy** affects command across all pitch types

## Storage in Save Game

Pitcher repertoires are stored in the `t_baseball_player_options` table and/or `t_baseball_player_traits`. Pitch type data is also available on the player profile in SMB3Explorer's "most recent season players" export and in SmbExplorerCompanion's `PlayerSeason` entity (via a `PitchTypes` collection).

## Traits Associated with Pitch Types

In SMB4, specific traits called "elite pitch" traits correspond to individual pitch types (e.g., "Elite 4-Seamer", "Elite Curveball"). These traits boost the effectiveness of a specific pitch. See [player-traits.md](player-traits.md) for the full trait list.
