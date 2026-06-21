# Failure Analysis: The Game Freeze

**Status: root cause confirmed and fix validated end-to-end against a live game install.** What
follows below is the investigation trail; see `validation-results.md` for the live test that
proved the fix actually works in-game, not just on paper.

What we knew at the start: import completed without any reported error from the POC tool itself.
The freeze happened in the *game*, specifically when navigating into the newly-registered league
from the SMB4 UI — not at the main menu/league-list level, and not during the import operation
itself.

Using a real `master_sav.sqlite` (decompressed from a live `master.sav`) and a real SMB4 install
on this machine, two concrete bugs in the legacy tool's code were identified directly against
ground truth, and then a live in-game test confirmed that correcting both bugs produces a league
that registers and loads with **no freeze, no errors, no observed problems at all**.

## Confirmed Bug #1: GUID type mismatch in `t_league_savedatas` (confirmed root cause)

The real schema, dumped via `PRAGMA`/`sqlite_master` from a live `master.sav`:

```sql
CREATE TABLE [t_league_savedatas](
  [GUID] BLOB CONSTRAINT [sqlite_autoindex_t_league_savedatas_1] PRIMARY KEY NOT NULL,
  [isMissing] BOOL NOT NULL DEFAULT 0
)
```

Real rows store `GUID` as a **16-byte binary blob**. Cross-referencing against actual
`league-{GUID}.sav` files on disk confirms the encoding exactly: for league
`99F30082-775B-4547-ADD8-8C7D2C94FCE5`, the stored blob is
`99 f3 00 82 77 5b 45 47 ad d8 8c 7d 2c 94 fc e5` — i.e., **strip the hyphens from the standard
uppercase string form and hex-decode in the same left-to-right order**. No byte-swapping, no
.NET/COM mixed-endian GUID layout — it's the plain RFC 4122 byte order, the same order the `uuid`
crate's `Uuid::as_bytes()` (Rust) or a Go `[16]byte` from `uuid.Parse()` would already give you.

The legacy tool's actual insert (`import/database/add_league_reference.rs`):

```rust
let uppercase_id = league_save_file.league_id.to_string().to_uppercase();
statement.bind((1, uppercase_id.as_str()))
```

This binds the GUID as a **36-character TEXT string** (e.g.
`"99F30082-775B-4547-ADD8-8C7D2C94FCE5"`), not a 16-byte blob. Because the column has `BLOB`
affinity (no type coercion in SQLite), the value is stored exactly as bound — a 36-byte UTF-8 TEXT
value sitting in a column where every other row is a clean 16-byte BLOB.

This is a strong, concrete explanation for a freeze rather than a clean error: native C++ code
reading this column almost certainly assumes a fixed-size 16-byte binary key (for hashing,
binary comparison against the GUID parsed from a save filename, or a fixed-size struct copy). A
36-byte value where 16 bytes are expected is exactly the kind of malformed input that produces
undefined behavior — an out-of-bounds read, a comparison that never matches/terminates, or a
struct copy that overruns into adjacent memory — rather than a handled error path, since the game
was never built to expect this column to contain anything but a well-formed 16-byte key.

**Fix, confirmed working**: bind the GUID as its raw 16-byte form, not the stringified-and-
uppercased text (`league_id.as_bytes()` in Rust; `uuid.UUID(s).bytes` in Python; a `[16]byte` from
`uuid.Parse()` in Go). A live test using exactly this encoding — see `validation-results.md` —
registered a cloned league in `master.sav` and loaded it in-game with no freeze and no errors.

## Confirmed Bug #2: `master.sav` is zlib-compressed, not a PKZIP container

The legacy tool's `zip_utils/master_save.rs` opens `master.sav` via `zip::ZipArchive::new()`,
treating it as a one-entry PKZIP archive. Checking the first bytes of a real, live `master.sav` on
this machine:

```
78 01 ec fc 07 58 14 49 d7 36 ...
```

`78 01` is a standard zlib header (CMF/FLG for default compression). This is **the same format
already documented** for per-league `.sav` files in `docs/domain/save-game-format.md` — `master.sav`
is not special-cased; it's zlib-compressed SQLite like everything else. It is not a PKZIP file, and
`zip::ZipArchive::new()` should fail outright on it (PKZIP parsing looks for an end-of-central-directory
signature it won't find in a zlib stream).

**This creates an inconsistency worth resolving before trusting the rest of this analysis**: if
the code in this repo snapshot genuinely cannot open `master.sav` at all, the import flow should
have failed loudly at that step, before ever reaching the `t_league_savedatas` insert or copying
files into the save directory — contradicting the report that registration "worked" up until the
in-game freeze. Possible explanations, none confirmed:

- The user reached the freeze via a different mechanism than running this exact `import_league`
  Tauri command end-to-end (recall from `legacy-tool-analysis.md` that the import UI was never
  wired up in the shipped frontend — it's plausible the registration was done by hand, e.g.
  directly editing a decompressed `master.sav` with a SQLite tool, which would reproduce Bug #1
  exactly without ever hitting this zip-vs-zlib code path).
- An earlier/different version of the tool (not what's in this checkout) handled `master.sav`
  correctly.

Regardless of how it was reached, **this is still a real bug** that must be fixed before this
code can be reused as-is — it's just not certain it's the bug that caused the specific freeze the
user originally observed with the legacy tool. Bug #1 remains the stronger candidate for that
specific incident.

**Fix, confirmed working**: decompress/recompress `master.sav` the same way per-league `.sav`
files are handled (zlib inflate/deflate), not via the `zip` crate. The validated approach in
`validation-results.md` does exactly this and round-trips `master.sav` without issue.

## Hypotheses Now Closed

- ~~Incomplete row in `t_league_savedatas` (missing columns)~~ — **closed**. The real table has
  exactly two columns, `GUID` and `isMissing`. The legacy insert populated both. There was no
  missing-column problem; the problem was the *type* of the GUID value (Bug #1).
- ~~A sibling table in `master.sav` also needs a row~~ — **closed, with caveat**. None of the other
  21 tables in `master.sav` (`t_achievements`, `t_custom_pennant_races`, `t_team_attributes`,
  `t_team_logos`, `t_user_preferences`, etc.) declare a foreign key against
  `t_league_savedatas.GUID`. There's no SQL-level evidence the game requires a corresponding row
  anywhere else in `master.sav` for a league to register. (This doesn't rule out logic enforced in
  game code rather than the schema, but it removes the most direct way that would show up.)

## Confirmed by Live Test

A live in-game test (full procedure and script in `validation-results.md`) cloned a real league
under a brand-new GUID, rewrote every internal GUID reference inside the clone, registered it in
`master.sav` with the corrected 16-byte blob encoding, and recompressed `master.sav` as zlib (not
zip). **Result: the cloned league appeared in the in-game league list and loaded with no freeze,
no errors, no observed problems at all.** Both bugs above are now confirmed fixes, not just
plausible theories.

## Still Open

### The `.hash` sidecar file

Not reverse-engineered, but now has a concrete negative result: the stored 4-byte hash for the
source league doesn't match CRC32 or Adler32 of either the compressed or decompressed `.sav`
bytes (checked directly). The validation test copied this file **unmodified** onto a clone whose
GUID and content both differ from the original, and the clone still loaded with no problem. That's
solid evidence the `.hash` file is not validated against content on this load path — at least not
in a way that blocks loading. Still unidentified: what it actually is (a version counter? a stat
unrelated to content?) and whether some other path the test didn't exercise (e.g., online
multiplayer/leaderboards) cares about it. Low priority given the test result, but worth keeping in
mind.

### Whether `master.sav` needs editing at all

Not tested directly — the validation test did register via `master.sav`, so it confirms that path
works, but doesn't rule out a simpler alternative (the game discovering `league-*.sav` files purely
by scanning the directory, with no registry edit needed). Since the tested path already works
end-to-end, this is now a nice-to-have simplification to explore later, not a blocker.
