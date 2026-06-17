# Plan: De-Risking League Transfer

**The core hypothesis is now validated, not just theorized.** A live in-game test (full detail in
`validation-results.md`) confirmed that a correctly-encoded `master.sav` registration — 16-byte
blob GUID, zlib round-trip — lets a cloned league load with no freeze and no errors. This plan no
longer needs to convince anyone the approach can work; what's left is scoping the remaining
unknowns and the actual Go implementation.

## Pre-Implementation Investigation

All four originally-planned investigation tasks are now **done**, using a real `master_sav.sqlite`,
a real SMB4 install on this machine, and a live in-game test (see `failure-analysis.md` and
`validation-results.md` for full detail):

1. ~~Inspect the bundled/live `master.sqlite` schema~~ — **done**. `t_league_savedatas` has
   exactly two columns, `GUID BLOB PRIMARY KEY` and `isMissing BOOL`. No other table in
   `master.sav` declares a foreign key against it.
2. ~~Determine `master.sav`'s real on-disk container format~~ — **done**. Confirmed zlib (`78 01`
   header on a live file), same as per-league `.sav` files — not a PKZIP container as the legacy
   code assumed.
3. ~~Reverse-engineer the `.hash` file~~ — **partially done, deprioritized**. The stored 4-byte
   value doesn't match CRC32/Adler32 of the `.sav` bytes (compressed or decompressed). The live
   test copied it unmodified onto a clone with different GUID and content, and the game loaded the
   clone fine regardless — strong evidence this file isn't validated against content on this load
   path. Still not identified, but no longer blocking.
4. ~~Reproduce with the fix applied~~ — **done**. See `validation-results.md`: cloned league
   registered and loaded successfully in-game.

**Bonus finding, not originally planned**: comparing real `t_league_savedatas` rows against real
`league-{GUID}.sav` filenames revealed the legacy code's actual bug — it bound the GUID as a
36-character uppercase string instead of the 16-byte blob every real row uses. This was the
confirmed root cause of the original freeze, now fixed and validated.

## Remaining Open Question (non-blocking)

**Whether `master.sav` editing is strictly necessary**, vs. the game discovering `league-*.sav`
files purely by directory scan. Not tested — the registered-via-`master.sav` path already works
end-to-end, so this is a possible future simplification, not something blocking implementation.

## Shape of the Feature Within smb-tools

Per the discussion on issue #162, this is **not** scoped to an existing franchise — it's a
top-level capability alongside franchise selection, since the source is a *snapshot* (not a live
franchise context) and the destination is a *new save game*, not anything smb-tools is currently
tracking. Rough shape, consistent with `user-docs/team-transfer.md`'s existing design:

- **Export** reads from an existing franchise snapshot (smb-tools already has these) rather than
  a live save — this sidesteps the POC's dependency on a separate app's config file
  (`SMB3Explorer/Config/config.json`) entirely, since smb-tools already has its own
  franchise/league GUID tracked in its registry. Export should remain low-risk: it is read-only
  with respect to the game, matching what the POC already proved works.
- **Import** writes only to a **copy** of a save game, never the live save in place — this is
  already a stated constraint in `user-docs/team-transfer.md` ("smb-tools always writes to a copy")
  and sidesteps a class of risk the POC didn't have, since the POC mutated the live, in-use
  `master.sav` and save directory directly. Exactly how "copy" interacts with the fact that the
  game discovers leagues from a fixed, well-known directory (it's not clear the game can be pointed
  at an arbitrary save directory) is itself one of the open questions — this needs to be resolved
  before the import UX can be finalized, since "writes to a copy" may mean producing a new save
  directory the user then has to manually swap in, rather than something smb-tools can fully
  automate end-to-end.
- Given Go backend conventions (`internal/CLAUDE.md`), any new save-game-mutation logic would live
  under `internal/service/` (orchestration) and `internal/store/` (the actual SQL), translated
  through the same anti-corruption-layer discipline already required for save-game reads — raw
  codes/columns from `master.sav` should not leak into smb-tools' own domain models any more than
  raw save-game codes do today.
- `modernc.org/sqlite` (already in use, no CGO) should be sufficient for read-write access to an
  extracted `master.sav`/`league-*.sav` — no need to reach for a different driver than what's
  already used for read-only access.
- If import ever needs to assign a league a genuinely new identity (as opposed to re-registering
  it under the GUID it already has, which is the more common real-world case), the validated test
  shows that's a small, well-scoped operation: 6 GUID-bearing columns across 5 tables in a league
  save (`t_leagues.GUID`/`originalGUID`, `t_conferences.leagueGUID`, `t_franchise.leagueGUID`,
  `t_seasons.historicalLeagueGUID`, `t_league_local_ids.GUID`), confirmed exhaustive via
  `PRAGMA foreign_key_list` against the real 91-table schema. See `validation-results.md`.

## Safety Requirements For Any Future Implementation

- **`master.sav` must be backed up on its own, immediately before mutation, in addition to any
  broader save-directory backup.** The legacy tool only covered it via a directory-wide zip taken
  earlier in the flow (`backup_save_game.rs`); that's a reasonable general safety net, but it's not
  a substitute for a dedicated, easily-restorable copy of `master.sav` taken right before it's
  edited. `master.sav` is being hand-edited outside of anything the game itself validates on
  write — of all the files this feature touches, it's the one where "restore the last known-good
  copy" needs to be a one-step operation, not "find the right directory-wide zip and re-extract."
- Whatever the eventual mutation path looks like, it should write to a new/temp file and swap it
  into place atomically (rename) rather than truncating and rewriting `master.sav` in place — so a
  failure mid-write can't leave a corrupt file with no good copy immediately at hand.

## Explicit Non-Goals For Now

- This plan does not address the compatibility constraint already documented in
  `user-docs/team-transfer.md` ("source snapshot and target save must be from the same game
  version") beyond noting it — that's a separate concern from the freeze investigation.
- This plan does not cover the actual Wails/Go implementation, UX, or how export reads from a
  franchise snapshot in detail — those are next steps now that the underlying mechanism is
  validated, not something this research phase needed to settle.
