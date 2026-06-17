# Validation Results: The Fix Works

**Confirmed live, on a real SMB4 install: a correctly-encoded `master.sav` registration lets a
cloned league load in-game with no freeze and no errors.** This is the test that turned
`failure-analysis.md`'s leading hypothesis into a confirmed fix.

## What Was Tested

A standalone Python script (stdlib only — `zlib`, `sqlite3`, `uuid`, `shutil` — no dependency on
the archived Smb4LeagueTransferTool repo) performed the following against a real SMB4 install:

1. **Cloned** an existing "Super Mega League" save (`league-99F30082-775B-4547-ADD8-8C7D2C94FCE5.sav`)
   under a brand-new randomly generated GUID, writing the clone to a new file. The source file was
   only ever opened for reading — never modified.
2. **Rewrote every internal reference to the league's GUID** inside the clone so it would be fully
   self-consistent under its new identity, not just renamed at the filename level. A real league
   save's GUID is referenced in exactly 5 places (confirmed via `PRAGMA foreign_key_list` against
   the real schema — out of 91 tables total):
   - `t_leagues.GUID` (the primary key itself)
   - `t_leagues.originalGUID` (self-referential historical link)
   - `t_conferences.leagueGUID`
   - `t_franchise.leagueGUID`
   - `t_seasons.historicalLeagueGUID`
   - `t_league_local_ids.GUID`

   Everything else in the schema hangs off `t_franchise`/local-ID indirection, not directly off the
   league GUID, so these 6 columns (5 tables) were sufficient to keep the clone internally
   consistent.
3. Renamed the clone's display name (`t_leagues.name`) to `"ZZZ TEST CLONE - DO NOT PLAY"` —
   purely cosmetic, so it would be unambiguously identifiable in-game (three other saves already
   share the name "Super Mega League").
4. Cloned `.sav.bak` the same way. Copied `.hash` **unmodified** (its format is still unidentified
   — see below) under the new GUID's filename.
5. **Backed up `master.sav`**, then decompressed it (confirmed zlib, not zip — see
   `failure-analysis.md` Bug #2), inserted a new `t_league_savedatas` row for the new GUID using
   the corrected **16-byte blob** encoding (not the legacy tool's 36-character string — see Bug
   #1), and recompressed/swapped it back in atomically.

## Result

The cloned league appeared in SMB4's in-game league list as "ZZZ TEST CLONE - DO NOT PLAY" and was
navigated into successfully — **no freeze, no errors, no observed problems of any kind.** This is
the exact step that froze the game when the legacy tool's (buggy) registration was used.

The test was then fully reversed: `master.sav` restored from its pre-test backup, clone files
deleted. No original save file was modified at any point in the test, so nothing else needed
restoring.

## What This Confirms

- **Bug #1 (GUID type mismatch) was sufficient on its own to explain the original freeze.** A
  correctly-typed 16-byte blob insert, with everything else about the legacy approach unchanged in
  spirit (hand-editing `t_league_savedatas` directly), works.
- **Bug #2 (zip vs. zlib) is also confirmed fixed** — the test's `master.sav` round-trip used
  zlib inflate/deflate throughout and produced a file the game read back without issue.
- **No sibling-table row in `master.sav` is needed.** Only `t_league_savedatas` was touched; the
  game accepted the new league with nothing else added to any of the other 21 tables.
- **The `.hash` file does not need to be regenerated for this to work** — it was copied verbatim
  from a different GUID's (slightly different content) save and the game didn't care, at least not
  on this load path. Still not reverse-engineered; still worth understanding before broader rollout
  in case some other code path checks it.
- **A full internal-GUID rewrite of a cloned league save is a small, well-scoped operation** — 6
  columns across 5 tables, not a sprawling undertaking. This means a "transfer to a fresh
  identity" implementation (as opposed to re-registering a league under its own original GUID,
  which is what most real cross-machine transfers would actually need) is tractable if it's ever
  needed.

## What This Doesn't Confirm

- This tested registering a **new** GUID that had no prior row in `master.sav` — i.e., the exact
  shape of a real import. It did not test re-registering a league under its *own* original GUID
  after deleting that row (the scenario closer to "this exact league already existed once and is
  being re-added") — though there's no reason to expect that to behave differently.
- Still doesn't address whether `master.sav` editing is strictly necessary versus the game
  discovering `league-*.sav` files by directory scan alone (see `failure-analysis.md`, "Still
  Open"). Not tested because the registered-via-`master.sav` path already works end-to-end.
- Single test, single machine, single game version. Not a substitute for the broader Go
  implementation eventually getting its own test coverage per `internal/CLAUDE.md` standards.

## Reusable Validation Script

The script used for this test (`validate_league_transfer.py`) is not part of the smb-tools
codebase — it lived outside the repo on the testing machine, since it directly mutates a live game
install and isn't something that should ship. The approach (clone → rewrite internal GUIDs →
rename for visibility → register in `master.sav` with a correctly-typed blob → zlib round-trip) is
the template a real Go implementation should follow; see `plan.md` for how that maps onto
smb-tools' architecture.
