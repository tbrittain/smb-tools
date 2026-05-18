# Game Integration

Investigation and design documents for optional deeper integration with the Super Mega Baseball 4 game.

| File | Contents |
|------|----------|
| [investigation.md](investigation.md) | Findings from examining the local SMB4 installation: engine identification, bundled SQLite databases, full schema discovery, auto-sync strategy |

## Key Findings Summary

- **SMB4 uses a custom Metalhead proprietary engine** — not Unity or Unreal. BepInEx/Harmony do not apply.
- **The game ships raw SQLite files** in `D3D12/assets/database/baseball/`, including a `league-template.sqlite` with the complete 67-table schema.
- **The schema is much richer than SMB3Explorer exposed** — franchise news events (skill changes, trait changes, trades, game results), team logos embedded in the DB, pitch counts, lineups, and more.
- **Filesystem watching (`fsnotify`)** is the practical auto-sync solution — cross-platform, no process injection, achieves Tier 1 auto-sync from `architecture/game-integration.md`.

See [investigation.md](investigation.md) for full details.
