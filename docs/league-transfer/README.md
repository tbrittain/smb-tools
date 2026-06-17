# League Transfer — Research Index

Exploratory research for [issue #162](https://github.com/tbrittain/smb-tools/issues/162) and the
public-facing design doc at `user-docs/team-transfer.md`. This is **not** an implementation plan
for smb-tools yet — it's the groundwork: what the previous proof-of-concept (POC) did, why it
likely failed, and what we'd need to confirm before attempting this for real.

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
- [`failure-analysis.md`](failure-analysis.md) — ranked hypotheses for what caused the freeze,
  with the evidence for and against each, and what's still unverified.
- [`plan.md`](plan.md) — investigation tasks that should happen before smb-tools attempts an
  implementation, plus a rough shape for how the feature could fit into smb-tools' architecture.

## Key Takeaway Up Front

The POC never wrote anything to the live `.sav` league file's *contents* — it copied that file
byte-for-byte. The risky write was a **single hand-crafted `INSERT`** into a table called
`t_league_savedatas` inside `master.sav`, SMB4's global registry of known leagues.

**Update**: this is no longer just a hypothesis. Using a real `master_sav.sqlite` and a real SMB4
install on this machine, we confirmed two concrete bugs in the legacy code:

1. **The GUID was inserted as a 36-character text string, but the column actually stores a
   16-byte binary blob** (verified by cross-referencing real `t_league_savedatas` rows against
   real `league-{GUID}.sav` filenames — the encoding is just hyphen-stripped hex, no byte
   swapping). This is the leading root-cause candidate for the freeze: native code reading a
   fixed-size 16-byte key and getting a 36-byte value instead is exactly the kind of bug that
   produces a hang rather than a clean error.
2. **`master.sav` is zlib-compressed, not a PKZIP archive** (confirmed via the live file's `78 01`
   header) — the legacy code's `zip` crate handling of it was wrong regardless of bug #1.

The table itself, contrary to initial speculation, only has two columns (`GUID`, `isMissing`) —
there was no missing-column problem. See `failure-analysis.md` for full detail and what's still
unverified (the `.hash` file, and whether `master.sav` needs editing at all).
