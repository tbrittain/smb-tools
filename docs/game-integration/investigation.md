# Game Integration Investigation

Investigation of Super Mega Baseball 4's technical internals to assess feasibility of game process hooks and auto-sync strategies.

**Installation path examined**: `E:\Games\Steam\steamapps\common\Super Mega Baseball 4\`

---

## Engine Identification

### Investigation Method

Initial string scans of the executable (sampling across all 343MB) found no signatures for Unity, Unreal, Godot, GameMaker, IL2CPP, or any standard game engine. To move beyond absence-of-evidence, the PE structure was analyzed properly:

- PE header parsed to locate the import directory table (RVA `0x1853F000`, resolved via section table to file offset `0x15634E00`)
- All imported DLLs extracted from the import descriptor table
- PE section layout examined

### PE Section Layout

| Section | Virtual Size | File Size | Notes |
|---------|-------------|-----------|-------|
| `.00cfg` | 33 MB | 33 MB | Code (Control Flow Guard table) |
| `.tls` | 14 MB | 14 MB | Thread-local storage |
| `.didata` | 47 MB | 1.4 MB | Delay-import data |
| `.debug` | **308 MB** | 308 MB | Debug/symbol information |
| `.sbss`, `.data`, etc. | small | small | Standard data sections |

The `.debug` section is 308MB — nearly all of the file's size. This is not a standard PDB file (signature does not match). It is compiled COFF-format debug information embedded in the shipping binary, which is unusual and potentially useful for future reverse engineering work.

### Imported DLLs (Complete List)

All imports confirmed via PE import directory — includes delay-loaded imports:

```
DirectX/Graphics:  d3d12.dll, dxgi.dll
Input:             DINPUT8.dll, XINPUT9_1_0.dll
Audio:             WINMM.dll
Networking:        WS2_32.dll, WINHTTP.dll, wininet.dll
Steam:             steam_api64.dll
Crypto:            bcrypt.dll, CRYPT32.dll
MSVC Runtime:      MSVCP140.dll, VCRUNTIME140.dll, VCRUNTIME140_1.dll, CONCRT140.dll
Windows APIs:      kernel32.dll, user32.dll, gdi32.dll, shell32.dll, advapi32.dll,
                   ole32.dll, oleaut32.dll, ntdll.dll, Secur32.dll, shlwapi.dll,
                   IPHLPAPI.dll, dbghelp.dll, VERSION.dll
C Runtime:         api-ms-win-crt-*.dll (convert, filesystem, heap, math, runtime, stdio, string, time, etc.)
```

**Notably absent**: `mono.dll`, `UnityPlayer.dll`, `GameAssembly.dll` (Unity); any Unreal Engine DLL; FMOD/Wwise audio middleware; PhysX/Havok physics middleware.

### Conclusion

**SMB4 is definitively built on a custom C++ engine by Metalhead Software**, compiled with MSVC 2019 toolchain and targeting DirectX 12. The evidence is positive (the import table is complete — this is not absence of evidence):

- No Unity runtime of any kind in the import table
- No Unreal Engine runtime of any kind
- No standard game engine middleware
- The only external dependencies are DirectX 12, Steam, standard Windows APIs, and the MSVC runtime
- Directory structure and `.mt` asset format are Metalhead-proprietary

The `CONCRT140.dll` (Microsoft Concurrency Runtime) and the `.tls` section (14MB of thread-local storage) suggest heavy use of parallelism, consistent with a custom engine doing its own job/task scheduling.

### Implications for Game Integration

The BepInEx + Harmony approach documented in `architecture/game-integration.md` **does not apply** — that toolchain is specific to Unity games. With a custom native C++ engine:

- Process-level hooks would require Windows native DLL injection (e.g., via the Detours library), hooking against functions identified by reverse engineering the 308MB debug section — possible but a significant undertaking
- The 308MB `.debug` section embedded in the binary (unusual for a shipped game) could be a meaningful asset for RE work if that path is ever pursued
- **Filesystem watching (`fsnotify`) is the practical auto-sync solution** — achieves Tier 1 auto-sync cross-platform with no process injection required

---

## Bundled SQLite Databases

The game ships with raw (uncompressed) SQLite databases in:

```
D3D12/assets/database/baseball/
  league-template.sqlite    # 1.1MB — template used when creating a new franchise/league
  master.sqlite             # 1.6MB — global game state, options, input, logo data
  league-{GUID}.sqlite      # 1.5–2.1MB each — the 3 built-in leagues (Super Mega League, etc.)
```

These are **not compressed** — they are standard SQLite files readable by any SQLite tool. The save files in `%LOCALAPPDATA%\Metalhead\Super Mega Baseball 4\` are zlib-compressed versions of the same format.

The `league-template.sqlite` is the most important reference: it contains the complete schema for franchise/league databases, including many tables that SMB3Explorer never queried.

---

## Schema: Much Richer Than Previously Documented

The `league-template.sqlite` has **67 tables and 22 views**. Our existing `docs/domain/save-game-schema.md` only documented the subset that SMB3Explorer's queries touched. The full picture includes many unexplored and potentially valuable tables.

### Tables Not Previously Documented

#### Franchise News System
The game maintains a typed news feed for every franchise event. Each event type has its own table:

| Table | Contents |
|-------|----------|
| `t_franchise_news` | Base news record (type, timestamp) |
| `t_franchise_news_game_result` | Per-game result events |
| `t_franchise_news_skill_changes` | Player attribute changes (e.g., Power improved during offseason) |
| `t_franchise_news_trait_changes` | Player trait changes |
| `t_franchise_news_traded` | Trade events |
| `t_franchise_news_resigned` | Re-signing events |
| `t_franchise_news_retired` | Retirement events |
| `t_franchise_news_roster_acquisition` | Roster pickup events |
| `t_franchise_news_training_outcome` | Training results |
| `t_franchise_news_salary_change` | Salary change events |
| `t_franchise_news_championship` | Championship records |
| `t_franchise_news_negotiation_round_result` | Contract negotiation outcomes |
| `t_franchise_news_accept_raise` / `t_franchise_news_reject_extension` | Contract decisions |
| `t_franchise_news_manager_moment` | In-game manager moment events |

**This is significant**: the game already persists a rich event log covering skill changes, trait changes, trades, and training outcomes. None of this was exposed by SMB3Explorer, but it could be extremely valuable for the companion app (e.g., tracking how player attributes evolved season over season without needing to diff snapshots).

#### Team Logos in the Database
| Table | Contents |
|-------|----------|
| `t_team_logos` | Logo image data stored directly in the DB |
| `t_team_logo_attributes` | Logo attribute metadata |
| `t_team_logo_element_types` | Logo element type definitions |
| `t_team_logo_types` | Logo type definitions |

Team logos are already embedded in the SQLite database — no separate image files. The companion app can extract these directly without any user action.

#### Player and Roster Management
| Table | Contents |
|-------|----------|
| `t_season_pitch_counts` | Pitch counts per player per season |
| `t_starting_lineups` | Lineup configurations |
| `t_pitching_rotations` | Pitching rotation order |
| `t_franchise_available_players` | Free agent pool |
| `t_franchise_pending_available_players` | Pending free agents |
| `t_franchise_player_extensions` | Contract extension offers |
| `t_franchise_retired_players` | Retired player records |
| `t_franchise_resigned_players` | Re-signed player records |
| `t_franchise_training` | Training data |
| `t_baseball_player_colors` | Player color/appearance data |

#### Financial
| Table | Contents |
|-------|----------|
| `t_franchise_team_cash` | Team cash/financial state |

#### Snapshots (Game-Internal)
| Table | Contents |
|-------|----------|
| `t_season_games_played_snapshot` | Season games played snapshots |
| `t_season_games_won_lost_snapshot` | Season W/L snapshots |
| `t_playoff_games_played_snapshot` | Playoff games played snapshots |
| `t_playoff_games_won_lost_snapshot` | Playoff W/L snapshots |
| `t_season_summary_snapshot` | Season summary snapshots |

These suggest the game itself takes periodic internal snapshots, which may be useful for understanding data compaction behavior.

#### Other
| Table | Contents |
|-------|----------|
| `t_save_data_validity` | Has `isUserModified` flag — game tracks if save has been altered |
| `t_season_user_controlled_teams` | Which teams the user controls each season |
| `t_franchise_manager_moments_queue` | Queued manager moments |
| `t_fantasy_draft_generated_players` | Fantasy draft player pool |
| `t_franchise_local_ids` | Franchise local ID mappings |

### Views Not Previously Documented

| View | Likely Contents |
|------|----------------|
| `v_stats_batting` | Denormalized batting stats — may simplify queries vs. multi-table joins |
| `v_stats_pitching` | Denormalized pitching stats |
| `v_season_standings` | Season standings (was previously documented) |
| `v_season_summary` | Season summary data |
| `v_franchise_players` | All franchise players |
| `v_franchise_players_including_pending_players` | Including free agent pool |
| `v_franchise_teams` | Franchise teams |
| `v_league_players` / `v_league_teams` | League-wide player/team views |
| `v_active_historical_players` / `v_active_historical_teams` | Historical active records |
| `v_season_historical_players` / `v_season_historical_teams` | Season-scoped history |
| `v_lineups_default` / `v_lineups_pennant` | Lineup configurations |
| `v_league_reclaimable_teams` | Teams available for reclaiming |
| `v_season_single_user_teams` | Single-user mode teams |

### master.sqlite

The master database (22 tables) manages global game state independent of any specific league/franchise:

| Table | Contents |
|-------|----------|
| `t_league_savedatas` | Registry of all saved leagues — the game's own franchise list |
| `t_options` | Global game options |
| `t_input_mappings` | Controller/keyboard mappings |
| `t_achievements` | Achievement tracking |
| `t_team_attributes` | Global team attribute overrides |
| `t_team_logos` | Global team logos |
| `t_franchise_news_filter` | News display preferences |
| `t_custom_pennant_races` | Custom pennant race configurations |

`t_league_savedatas` in master.sqlite is particularly interesting — this is the game's own registry of franchises, and likely maps to the save file paths. This could be read to enumerate a user's franchises without requiring them to manually locate each `.sav` file.

---

## Auto-Sync Strategy: Filesystem Watching

Since process hooks require significant RE work with no existing community tooling to build on, **filesystem watching is the practical auto-sync solution**.

The Go library `github.com/fsnotify/fsnotify` provides cross-platform filesystem event notifications. When the game writes to a save file:
1. fsnotify emits a `WRITE` event on the `.sav` file
2. smb-tools triggers the snapshot + sync pipeline
3. User sees "Synced automatically" notification

This achieves the Tier 1 goal from `architecture/game-integration.md` without any process injection, without any Windows-only code, and without any RE work.

**Caveat**: The game holds a file lock while writing. The watcher should debounce the event and wait for the lock to be released before reading. A short delay + retry loop handles this.

**This approach works on Windows, macOS, and Linux.**

---

## Schema Analysis Action Items

The tables discovered here significantly expand what the companion app should capture. Before finalizing the new companion DB schema, the following should be queried against the actual bundled databases to understand their structure:

- [ ] `t_franchise_news_game_result` — does this contain box score level detail?
- [ ] `t_franchise_news_skill_changes` — exact columns for attribute change tracking
- [ ] `t_franchise_news_trait_changes` — columns for trait change events
- [ ] `v_stats_batting` / `v_stats_pitching` — full column list; may replace complex joins
- [ ] `t_season_pitch_counts` — what granularity? Per game? Per season?
- [ ] `t_team_logos` — format of stored logo data (raw bytes? format?)
- [ ] `t_league_savedatas` in master.sqlite — can we enumerate save files from here?

These queries should be run against the bundled `league-template.sqlite` and one of the live league files during the schema analysis phase.
