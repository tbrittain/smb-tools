# Plan: De-Risking League Transfer

This is an investigation plan, not an implementation plan. Nothing here should be built until the
"Pre-Implementation Investigation" tasks have actually answered the open questions — building the
import side again without doing this is exactly how the POC ended up freezing the game.

## Pre-Implementation Investigation

Two of the four original investigation tasks are now **done**, using a real `master_sav.sqlite`
and a real SMB4 install on this machine (see `failure-analysis.md` for full detail):

1. ~~Inspect the bundled/live `master.sqlite` schema~~ — **done**. `t_league_savedatas` has
   exactly two columns, `GUID BLOB PRIMARY KEY` and `isMissing BOOL`. No other table in
   `master.sav` declares a foreign key against it.
2. ~~Determine `master.sav`'s real on-disk container format~~ — **done**. Confirmed zlib (`78 01`
   header on a live file), same as per-league `.sav` files — not a PKZIP container as the legacy
   code assumed.
3. **Bonus finding, not originally planned**: comparing real `t_league_savedatas` rows against
   real `league-{GUID}.sav` filenames revealed the legacy code's actual bug — it bound the GUID as
   a 36-character uppercase string instead of the 16-byte blob every real row uses. This is now
   the leading root-cause candidate for the freeze, ahead of anything related to schema
   completeness.

Remaining tasks before writing any import code:

1. **Diff `master.sav` before/after the game itself creates a brand-new league.** Lower priority
   than originally scoped (the FK analysis already shows no sibling-table dependency at the SQL
   level), but still worth doing once, mainly to check for non-schema-enforced expectations (e.g.,
   does the game care about insertion order, or about `isMissing` ever being set to anything but
   0 for a freshly added league?).
2. **Reverse-engineer the `.hash` file.** Try standard checksums (CRC32, MD5, SHA-1/256) of the raw
   `.sav` bytes against a real `.hash` file's contents. If a league with no `.hash` file at all
   loads fine in-game, that's evidence the file is optional — worth confirming directly by
   temporarily renaming a `.hash` file aside (after a backup) and checking whether the game still
   loads that league normally.
3. **Confirm whether import even requires a `master.sav` mutation at all**, vs. whether the game
   reconciles `master.sav` against whatever `league-*.sav` files it finds on disk at startup. Test
   by manually copying a known-good `league-*.sav` triplet into the directory *without* touching
   `master.sav` (using a **correctly-formed** triplet this time — same GUID across all three
   files) and see what the game does with it on next launch.
4. **Reproduce with the fix applied.** Given Bug #1 (GUID type) is now understood, the cheapest
   next step toward confidence is simply trying the insert again with the GUID bound as a 16-byte
   blob and seeing whether the league loads without freezing. This is the single most direct test
   available and should happen before any deeper architectural work on the smb-tools side.

If task 3 pans out (no `master.sav` edit needed at all), it would eliminate the riskiest part of
this entire feature — hand-editing a SQLite table the game also reads — and reduce import to
"place validated files in the right location," a much smaller blast radius. But given Bug #1 is a
plausible complete explanation on its own, task 4 (just fix the type and retry) is the faster way
to find out whether further investigation is even necessary.

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

- No code should be written against `master.sav` for smb-tools until the GUID-type fix (Bug #1)
  has actually been retried and confirmed to resolve the freeze, or until the remaining tasks
  above have ruled it out. Guessing at the cause a second time is the exact mistake being
  corrected here — this time there's a concrete, testable fix to verify first.
- This plan does not address the compatibility constraint already documented in
  `user-docs/team-transfer.md` ("source snapshot and target save must be from the same game
  version") beyond noting it — that's a separate concern from the freeze investigation.
