# Game Overview: Super Mega Baseball 3 & 4

## What the Game Is

Super Mega Baseball (SMB) is an arcade-style baseball game developed by Metalhead Software. The franchise is notable for deeply simulated baseball statistics and a robust franchise management layer despite its arcade presentation. The save files it produces are rich, structured databases that far exceed what the game's own UI exposes — which is the entire motivation for the tooling in this repository.

The two relevant versions are:
- **Super Mega Baseball 3** (SMB3) — first supported by the original tooling
- **Super Mega Baseball 4** (SMB4) — the current version; primary focus of the companion app

## Game Modes

### Franchise Mode

The core mode these tools are built around. Key characteristics:

- Multi-season play with full continuity: players age, retire, and accumulate career statistics
- One team is "player-controlled"; the others are CPU-managed
- **Payroll system**: each team has a budget; players have salaries (stored internally in game units, displayed as units × 200)
- **Player lifecycle**: players progress through ages, eventually retiring; the game only surfaces stats for currently active players, making external history tracking essential
- **50-season limit**: the game's own franchise history UI only retains the most recent ~50 seasons of data; stats from earlier seasons are discarded unless externally captured
- Standings, schedules, and playoff brackets are all persisted in the save file

### Season Mode

Fundamentally the same as Franchise Mode — the user plays multiple seasons continuously with full standings, stats, and career accumulation — with one difference: **players and teams never evolve**. No aging, no attribute changes, no retirement, and no free agency; the player pool stays exactly as it started. In the save file, a Season Mode league has no `t_franchise` row (see `docs/domain/save-game-schema.md`), which is also how the app tells the two modes apart. smb-tools supports importing and tracking Season Mode saves; the only feature that doesn't apply is Hall of Fame, since induction is based on career length and retirement, neither of which happens in Season Mode.

### Elimination Mode

Tournament-style competition. Games are played in a bracket/elimination format. Treated as a distinct league type in the save game schema (`t_leagues.teamTypePermissions` differentiates modes).

## SMB3 vs. SMB4: Key Differences for Tooling

| Area | SMB3 | SMB4 |
|------|------|------|
| Save file structure | Single league per save file | Multiple leagues per save file; league selected via GUID |
| League GUID | Not present in filename | Encoded in the `.sav` filename (e.g., `league-{guid}.sav`) |
| Trait system | 20 traits, no chemistry | 80+ traits organized by chemistry type |
| Chemistry types | Not applicable | Competitive, Spirited, Disciplined, Scholarly, Crafty |
| Supported by SMB3Explorer | Yes | Yes |
| Supported by SmbExplorerCompanion | Partial / not prioritized | Primary focus |

## Why External Tooling Exists

1. **The game hides statistics it actually tracks.** The save file contains granular per-season batting, pitching, and team data that the game's UI never fully exposes (e.g., BABIP, FIP, pitch usage, detailed splits).
2. **The 50-season franchise limit.** Long-running franchises lose early history. External persistence captures it before it disappears.
3. **No career statistical aggregation.** The game shows current-season stats; career accumulation requires external computation.
4. **No franchise-wide leaderboards.** There is no in-game equivalent of a "franchise all-time home run leader" or similar.
5. **Awards are not tracked historically.** MVP, Cy Young, etc., are not retained in the save beyond the current season.
