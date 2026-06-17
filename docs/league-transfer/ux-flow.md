# UX Flow: League Transfer

This describes the user-facing shape of the league export/import feature for MVP. It assumes the
mechanism validated in `validation-results.md` (correctly-encoded `master.sav` registration) and
builds the UX around it. Implementation-level concerns (Go package layout, migrations, etc.) live
in `plan.md`; this doc is about what the user sees and does.

## Top-Level Split

League Transfer is **not** a feature within a franchise — it's a separate mode from franchise
tracking entirely, surfaced at the same level as franchise selection. On launch, the user picks
one of:

1. **Franchise Tracker** — the existing app, scoped to a franchise's stat history.
2. **League Transfer** — export/import leagues. No franchise context, no companion DB, nothing
   about stat tracking.

This was flagged directly in the issue #162 discussion and is treated as settled, not an open
question: franchise tracking and league transfer are different jobs that happen to read the same
underlying save files, and the UI should reflect that they're different jobs.

## League Transfer Home: Discovery & Introspection

Entering League Transfer shows **every league found on disk** — not just leagues the user has
added to smb-tools as a tracked franchise. This is a deliberate difference from the legacy POC,
which could only export leagues that happened to be in a separate app's config file. smb-tools
already has the building blocks for this: `internal/config.DiscoverSaveFiles()` /
`ScanDirShallow()` find every `league-*.sav` candidate on disk today; this feature reads each one
directly (read-only, same zlib-decompress-then-query pattern already used for franchise sync) to
populate a list, with no dependency on the franchise registry at all.

For each discovered league, show enough to let the user say "this is the one I want":

- **League name** (`t_leagues.name`)
- **Conferences** and, if present, **divisions** within them. Division is genuinely optional —
  a league can have conferences with no divisions underneath, and the UI must not assume a
  conference always has divisions to drill into.
- **Teams** — at minimum the list of team names; team logos could be a nice-to-have if there's
  time, but aren't required for MVP since the goal here is identification, not full
  franchise-style browsing.

This is **introspection of the league as a structure**, not franchise stat history — no
standings, no season records, no player stats. That's franchise tracker territory and out of
scope here. The query surface needed is small: `t_leagues`, `t_conferences`, `t_divisions` (if it
exists — verify against the real schema before assuming the table name), `t_teams`/whatever maps
teams to conferences/divisions.

## Export Flow

1. User selects a league from the discovery list.
2. Single button: **Export League**.
3. App reads the league's `.sav`, `.sav.bak` (if present), and `.hash` (if present) directly from
   the save directory and packages them into a zip — same shape as the legacy POC's export, which
   was already proven safe (read-only, no game-state mutation). A small manifest (league name,
   GUID, export timestamp, smb-tools version) inside the zip is worth adding — it costs nothing
   and gives both the import-time validator and a human opening the zip something to check against
   before relying on filename parsing alone.
4. User picks a save location for the resulting zip; success state shows the path (mirroring what
   the legacy POC's export modal already did reasonably well).

No mutation of any game file happens during export. This stays the lowest-risk part of the
feature, consistent with everything found so far.

## Import Flow

1. User picks a zip file (file picker, not drag-and-drop for MVP — keep it simple).
2. Unzip to a temp location.
3. **Validate shape** before touching anything real:
   - All expected files present and sharing one consistent GUID (same check the legacy POC
     already did correctly).
   - The `.sav` file decompresses via zlib successfully (if it doesn't, this isn't a smb-tools
     export or it's corrupted — reject early with a clear message).
   - The decompressed SQLite file actually contains the tables a league save should have
     (`t_leagues`, `t_teams`, `t_franchise`, `t_conferences` at minimum) — the same spirit as the
     existing "Save Game SQL — Real Schema Required" schema-validation precedent
     (`internal/CLAUDE.md`, and historically SMB3Explorer's own `t_stats`/`t_leagues` check). This
     is a shape/sanity check, not a security boundary — see disclaimers below.
   - The GUID doesn't already exist in this user's `master.sav` (**hard stop with a clear message
     if it does** — "this league is already in your save" — no overwrite support for MVP).
4. **Check whether SMB4 is currently running; refuse to proceed if so**, with a clear message to
   close the game first. `master.sav` is locked by the game while it's open, and writing to it
   concurrently is a real corruption risk, not just an inconvenience — this is worth a real
   process check, not just a warning label.
5. Back up `master.sav` (per the safety requirement in `plan.md`), write the new
   `league-{GUID}.sav`/`.sav.bak`/`.hash` files into the save directory, register the GUID in
   `master.sav` using the validated correct encoding, swap it in atomically.
6. Success state: league name, confirmation it's ready, and a reminder that this only takes
   effect the next time SMB4 is launched.

### Safety Disclaimers (shown to the user, not just documented)

- **smb-tools does not scan imported files for malware.** The shape validation above confirms the
  zip *looks like* a league export — it is not a security boundary. Show a clear, plain-language
  disclaimer before import: only import league files from people you trust, and run your own
  virus scan if you're unsure. This should be visible at the point of import, not buried in docs.
- Recommend (not require) the user keep their own backup of anything important before importing
  unfamiliar content, on top of the automatic `master.sav` backup smb-tools takes itself.

## Resolving the "Writes to a Copy" Promise

`user-docs/team-transfer.md` currently states: *"smb-tools always writes to a copy. Your existing
save files are never modified."* The validated mechanism can't fully honor this as written —
`master.sav` is the game's single shared registry, and there's no way to register a new league
without editing the live one (the open question of whether the game could instead discover saves
by directory scan alone, noted in `plan.md`, remains untested and isn't something to design around
yet).

What's actually true, and what the docs should say instead:

- **No existing league save is ever modified.** Import only ever adds new `league-{GUID}.*` files.
  This part of the original promise holds completely.
- **`master.sav` (the shared league registry, not a "save file" in the franchise sense) is edited
  in place, with a mandatory backup taken immediately before any write.** This should be stated
  plainly rather than implied away — users trust this tool with their save data, and "we back this
  up automatically before touching it" is a stronger, more honest claim than a blanket "never
  modified" that doesn't hold up under what the feature actually has to do.

Recommend updating `user-docs/team-transfer.md`'s "Importing a League" section to reflect this
precisely.

## Explicitly Deferred (Not MVP)

- Overwriting/updating an already-imported league.
- Drag-and-drop import, multi-file batch export.
- Team logos in the league discovery list.
- Anything that assumes `master.sav` editing isn't needed (pending the open question in `plan.md`).
- Cross-platform process detection for "is SMB4 running" — needs its own scoping (Windows is the
  primary target per existing save-path detection; Linux/Proton support already exists for
  save-file discovery, so the process check should at least not silently fail there).
