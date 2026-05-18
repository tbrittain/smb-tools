# Game Integration (Optional / Windows-Only)

An optional enhancement layer that hooks directly into the Super Mega Baseball 4 game process, enabling automatic sync triggers and potentially in-game event capture that never reaches the save file.

**This is strictly optional.** The core smb-tools application is cross-platform (Windows, macOS, Linux) and operates entirely through the save game SQLite database. Nothing in the core app depends on or requires game integration. This is an additive layer for Windows users who want a more seamless experience.

---

## Motivation

The core sync flow requires the user to manually trigger a sync at the right moment — ideally at the end of each season, before simulating the offseason. The game integration layer removes this friction entirely and potentially captures data that the save game never persists.

---

## Capabilities (Tiered by Feasibility)

### Tier 1 — Auto-Sync on Save Event (High confidence, high value)

Hook the method the game calls when it writes to the save database. When detected:
- Automatically trigger a snapshot + sync without any user action
- Surface a notification: "Season 8 synced automatically"
- Eliminates the "remember to sync before advancing the offseason" problem entirely

This is the highest-value, most feasible hook. The save write event is a clear, detectable lifecycle point.

### Tier 2 — In-Game Event Capture During Played Games (Requires investigation)

During games the user actually plays (not simulated), the game engine computes physics and outcomes that are never written to the save file: pitch velocity readings, exit velocity, launch angle, hit trajectory, pitch-by-pitch sequence.

If the game's C# class structure exposes these as hookable events or accessible properties, capturing them would enable:
- Pitch-by-pitch game logs
- Exit velocity and launch angle per batted ball
- Heat maps and spray charts
- A Statcast-equivalent layer on top of franchise data

**Feasibility depends on the game engine architecture.** Requires decompilation and investigation before committing to design. See investigation status below.

### Tier 3 — Simulation Event Capture (Low confidence)

During CPU-vs-CPU simulation (auto-advancing a season), outcomes are almost certainly generated statistically — a dice-roll engine that produces "player X hits a home run" without computing the physics of that home run. There is likely nothing to hook for granular event data during simulation.

This tier is noted for completeness but is not expected to be achievable.

---

## Likely Technical Approach: BepInEx + Harmony

If SMB4 is a Unity title (under investigation — see below), the standard community modding stack applies:

**BepInEx** is the dominant Unity modding framework. It loads a plugin loader alongside the game at startup. It is not memory scanning or process injection in the suspicious sense — it is the mechanism behind mods for Valheim, Risk of Rain 2, Lethal Company, and hundreds of other Unity titles.

**Harmony** is the method-patching library BepInEx uses. It allows you to:
- **Prefix** a method: run code before the original executes
- **Postfix** a method: run code after the original executes
- **Transpile**: modify the IL instructions of the original

The game integration plugin would be a separate C# DLL loaded by BepInEx. It communicates with the smb-tools companion app via a local mechanism (named pipe, local socket, or shared file in the app data directory).

**Important distinction: IL2CPP vs. Mono**

Unity games compile in one of two modes:
- **Mono**: C# IL is preserved in `Assembly-CSharp.dll`. Fully decompilable via dnSpy/ILSpy. Standard BepInEx works directly.
- **IL2CPP**: C# is ahead-of-time compiled to native code. Much harder to mod — requires IL2CPP-specific BepInEx and additional tooling. Decompilation is possible but more involved.

Which mode SMB4 uses is the single most important factor for modding feasibility. Under investigation.

---

## Architecture: Plugin ↔ App Communication

```
SMB4 game process
  └── BepInEx plugin (C# DLL)
        ├── Hooks save write event → notifies smb-tools
        └── Hooks in-game events → streams to smb-tools
              ↓ (named pipe / local socket)
smb-tools companion app (Go)
  └── Listens for plugin events
  └── Triggers sync or records in-game data on receipt
```

The plugin is a thin event emitter. All data processing, storage, and UI happen in smb-tools. The plugin has no schema knowledge and no database access.

---

## Scope Boundaries

| In scope | Out of scope |
|----------|-------------|
| Auto-sync trigger on save event | Any modification of game save data |
| Read-only observation of game state | Any gameplay modification or cheating |
| In-game stat capture (played games) | Online/multiplayer modes |
| Windows only | macOS, Linux (no BepInEx need there anyway — save file polling fills the gap) |

The plugin never writes to the game's save file. It is a passive observer.

---

## Investigation Status

**See `docs/game-integration/investigation.md` for full findings.**

Summary:
- SMB4 is **not a Unity or Unreal title** — it uses a custom Metalhead proprietary engine. BepInEx/Harmony do not apply.
- Process-level hooks would require native Windows DLL injection with no existing community tooling to build on. Not a practical near-term path.
- **Filesystem watching (`fsnotify`) achieves Tier 1 auto-sync** cross-platform with no process injection.
- The bundled SQLite databases reveal a much richer schema than previously documented — franchise news events, team logos, pitch counts, and more. See investigation for the full table list and action items.

---

## Relationship to Core App Timeline

This feature is **not on the critical path**. The core app is designed and built first, entirely independently. The game integration layer is an enhancement added after the core is stable.

When the time comes, the plugin would be distributed separately — an optional download for Windows users who want automatic sync. The core app works fully without it on all platforms.
