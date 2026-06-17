# League Transfer — Research Index

Exploratory research for [issue #162](https://github.com/tbrittain/smb-tools/issues/162) and the
public-facing design doc at `user-docs/team-transfer.md`. **The core mechanism is now validated**:
a live in-game test confirmed that the freeze the prior proof-of-concept (POC) hit is fixable, and
the fix works. This is still not a full smb-tools implementation plan — it's the groundwork that
implementation should build on.

## Background

A prior POC, [Smb4LeagueTransferTool](https://github.com/tbrittain/Smb4LeagueTransferTool) (cloned
locally at `../Smb4LeagueTransferTool`), got ~95% of the way to working: export, packaging, and
re-import of a league all completed without error. The failure happened at the very last,
most game-sensitive step — after import "registered" the league with the game, navigating to it
in the SMB4 UI **froze the game**, requiring an ALT+F4 force-close. That is the crux of the
problem this research is trying to de-risk before smb-tools invests in its own implementation.

## Documents

- [`legacy-tool-analysis.md`](legacy-tool-analysis.md) — line-by-line account of what the POC's
  Rust backend actually does for export and import, including the file formats and game paths it
  assumes.
- [`failure-analysis.md`](failure-analysis.md) — the two confirmed bugs that caused the freeze,
  the evidence for each, and what's still unverified (low-priority at this point).
- [`validation-results.md`](validation-results.md) — the live in-game test that proved the fix
  works: a cloned league, correctly registered, loaded with no freeze and no errors.
- [`plan.md`](plan.md) — what's left before smb-tools attempts an implementation, plus a rough
  shape for how the feature could fit into smb-tools' architecture.
- [`ux-flow.md`](ux-flow.md) — the user-facing shape of the feature: top-level mode split, league
  discovery/introspection, export, import, and the safety disclaimers shown at import time.

## Key Takeaway Up Front

The POC never wrote anything to the live `.sav` league file's *contents* — it copied that file
byte-for-byte. The risky write was a **single hand-crafted `INSERT`** into a table called
`t_league_savedatas` inside `master.sav`, SMB4's global registry of known leagues. Two concrete
bugs in that insert were found and confirmed against a real `master.sav` and a real SMB4 install:

1. **The GUID was inserted as a 36-character text string, but the column actually stores a
   16-byte binary blob** (verified by cross-referencing real `t_league_savedatas` rows against
   real `league-{GUID}.sav` filenames — the encoding is just hyphen-stripped hex, no byte
   swapping).
2. **`master.sav` is zlib-compressed, not a PKZIP archive** (confirmed via the live file's `78 01`
   header) — the legacy code's `zip` crate handling of it was wrong regardless of bug #1.

The table itself, contrary to initial speculation, only has two columns (`GUID`, `isMissing`) —
there was no missing-column problem.

**Both fixes were then validated live**: a real league was cloned under a new GUID, registered in
`master.sav` with the corrected encoding, and loaded in-game with **no freeze, no errors, no
observed problems at all.** Full procedure in `validation-results.md`. The one remaining loose
end — the `.hash` sidecar file's exact format — turned out not to matter for this load path either
(the test copied it unmodified and the game didn't care).
