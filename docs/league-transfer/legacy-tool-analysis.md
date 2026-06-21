# Legacy Tool Analysis: Smb4LeagueTransferTool

Source: `../Smb4LeagueTransferTool` (Tauri v1 + Rust backend + SolidJS frontend). This was a
standalone desktop app, separate from SMB3Explorer/SmbExplorerCompanion and from smb-tools.
Version 1.0.0, never released publicly per the issue thread ("had a pretty good POC... 80% of the
way there").

## Stack

- Tauri 1.5.2 (Rust backend, webview frontend)
- `sqlite` crate 0.33.0 — a thin libsqlite3 binding (not `rusqlite`, not `modernc.org/sqlite`)
- `zip` crate 0.6.6 (PKZIP/DEFLATE, via the `zip-rs` crate) for packaging
- `walkdir`, `uuid`, `chrono`, `serde`/`serde_json`, `thiserror`

The frontend (SolidJS) only ever wired up the **export** flow. The "Import League" tab in
`src/App.tsx` is a literal placeholder div (`<div>Hey this is the import page</div>`) — there is
no import UI. The Rust `import_league` Tauri command exists and is fully implemented and
registered in `main.rs`'s `invoke_handler!`, but nothing in the shipped frontend calls it. This
means the freeze the user observed was very likely triggered by invoking the command directly
(e.g., via devtools `invoke("import_league", ...)`) rather than through a finished UI — consistent
with this being an in-progress POC, not a finished feature.

## File & Path Model

| File | Location | Purpose |
|---|---|---|
| `master.sav` | `{LOCALAPPDATA}/Metalhead/Super Mega Baseball 4/{steam_id}/` | Game's global registry of known leagues (`t_league_savedatas` table, among ~22 others) |
| `league-{GUID}.sav` | same directory | One league's actual save data (the file smb-tools already reads) |
| `league-{GUID}.sav.bak` | same directory | Game-maintained backup of the above |
| `league-{GUID}.hash` | same directory | Unknown-format sidecar file, present for at least some leagues — never reverse-engineered by the POC |

`get_steam_directory_path()` (`windows_os/paths.rs`) locates the per-steam-ID subdirectory by
scanning for the one child directory under `Metalhead/Super Mega Baseball 4` whose name parses as
a `u64`. It explicitly bails out if it finds more than one (multi-account machines unsupported).
This matches smb-tools' own `internal/config/savegame_paths.go`, which independently arrived at
the same `{steam_id}` layout and explicitly excludes `master.sav` from its `league-*.sav` file
scan — i.e., smb-tools already treats `master.sav` as out of scope today.

**Confirmed bug**: SMB4's per-league save files (`league-{GUID}.sav`) are zlib-compressed SQLite
databases (per `docs/domain/save-game-format.md`). The POC's `master.sav` handling
(`zip_utils/master_save.rs`) instead treats `master.sav` as a **PKZIP archive containing exactly
one entry** — it calls `zip::ZipArchive::new` directly on the file. Checking a live `master.sav`
on disk confirms it is in fact zlib-compressed (`78 01` header), the same as every other `.sav`
file, not a PKZIP container. There's no indication this was an intentional, considered choice —
nothing in the repo or its single-line README documents a reason `master.sav` would be packaged
differently from every other save file, and the simplest explanation is that it's simply wrong.
See `failure-analysis.md` for the full confirmation and how it cross-checks against real bytes.

## Export Flow

Entry point: `commands/export_league.rs` → `export::export_league()`.

1. **Discover exportable leagues** — `get_league_export_options()` does **not** read SMB4's own
   `master.sav` registry. It reads `{LOCALAPPDATA}/SMB3Explorer/Config/config.json`, parses a
   `smb4Leagues` array (`Smb4League { name, id, num_times_accessed, num_seasons,
   first_accessed, last_accessed }`), and returns that list. This is **SMB3Explorer's own
   config file** (the legacy C# app), not anything written by the game. The POC is entirely
   dependent on SMB3Explorer having been run first to populate this file — there is no
   independent mechanism for discovering which leagues exist. This dependency does not carry
   over to smb-tools, which has no equivalent config file and discovers leagues by scanning for
   `league-*.sav` directly.
2. **Validate the three on-disk files exist** for the selected league ID — `{league-name}.sav`,
   `.sav.bak`, and (optionally) `.hash` — naming them `league-{UPPERCASE-GUID}.sav` etc. in the
   Steam save directory. The `.hash` file is treated as optional (`has_hash_file` flag) — some
   leagues apparently don't have one.
3. **Zip them up** (`zip_utils/league.rs::zip_league`) into a single archive alongside a
   bundled `README.txt`, written to
   `{LOCALAPPDATA}/Smb4LeagueTransferTool/Exports/{sanitized-name}_{timestamp}.zip`.

Nothing about export touches `master.sav` or writes to the game's data at all — export is fully
read-only with respect to the game. This means **export is safe to reattempt/reimplement with low
risk**; the risk is entirely concentrated in import.

## Import Flow

Entry point: `commands/import_league.rs::import_league()`. Sequence:

1. Unzip the user-provided archive to a temp directory (`zip_utils/league.rs::unzip_league`).
2. `validate_league()` — parses the three extracted file names, extracts the GUID from each
   filename, and requires **all files present to share the exact same GUID** (`HashSet` of parsed
   GUIDs must have length 1). This is the only validation performed — no validation of the `.sav`
   file's internal contents, no checksum/hash verification against the `.hash` file's actual
   contents (the file is just copied, never read).
3. **Back up the user's entire SMB4 save directory** (`backup_save_game.rs`) — zips the whole
   `Metalhead/Super Mega Baseball 4` tree to
   `{LOCALAPPDATA}/Smb4LeagueTransferTool/Backups/save_file_backup_{timestamp}.zip`. This runs
   *before* any mutation — a reasonable safety net, and one smb-tools should keep regardless of
   implementation approach.
4. Locate `master.sav` in the Steam directory; error if missing.
5. **Unzip `master.sav`** to a temp `master.sav.db` file (treating it as a one-entry PKZIP
   archive — see the wrinkle noted above).
6. **`add_league_reference()`** (`import/database/add_league_reference.rs`) — opens that
   extracted SQLite file directly (read-write, no backup of the extracted DB itself) and runs:

   ```sql
   INSERT INTO t_league_savedatas (GUID, isMissing) VALUES (?, 0);
   ```

   binding the new league's GUID (uppercased string form). **This is the entire "registration"
   step** — two columns, no schema introspection, no check of what other columns exist or what
   values the game itself would populate for those columns on a normal "new league" flow.
7. Re-zip the mutated `master.sav.db` back into a new `master.sav` (`zip_master_save`),
   **overwriting the original** at its original path. No second backup is taken of the pre-edit
   `master.sav` specifically — only the directory-wide backup from step 3 covers it, and that
   backup is a separate zip archive, not a same-directory file the user could quickly swap back in
   by hand. A future smb-tools implementation should not repeat this: `master.sav` is being
   hand-edited outside of anything the game itself validates on write, so it should get its own
   explicit, easily-restorable backup immediately before mutation, not just rely on being covered
   by a broader directory snapshot taken earlier in the flow. See `plan.md`'s safety requirements.
8. Copy the three league files (`.sav`, `.sav.bak`, `.hash` if present) into the Steam directory
   under the canonical `{UPPERCASE-GUID}.*` naming the game expects.

After step 8, the import command returns success. The freeze happened when the user subsequently
launched the game and tried to navigate into the newly-registered league from its UI.

## What the POC Got Right

- Read-only export with no game-state mutation.
- Backs up the whole save directory before any write during import.
- GUID-consistency validation across the three files before touching anything.
- Correct understanding of the `{steam_id}` directory layout and `league-{GUID}.*` naming —
  consistent with what smb-tools independently verified for its own save-discovery code.

## What's Missing or Unverified

- No schema introspection of `t_league_savedatas` — column list, types, NOT NULL constraints,
  foreign keys, and defaults were never determined.
- No comparison against what the game itself writes to that table when a league is created
  through normal play (e.g., via a save-file diff: snapshot `master.sav` before/after creating a
  brand-new league in-game).
- The `.hash` file's format, algorithm, and what it's computed over were never investigated —
  it's only ever copied as opaque bytes, never read or regenerated.
- No verification that `master.sav` is genuinely a one-entry PKZIP container rather than a raw
  zlib stream like the per-league `.sav` files.
- No verification of whether other master-level tables (e.g., anything that maps Steam profile ↔
  league, save slot ordering/UI position, achievements, custom pennant races) need a
  corresponding row for a league to be navigable without the game choking.
