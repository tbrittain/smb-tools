# master.sav Schema

`master.sav` is SMB4's global, per-Steam-profile save file — distinct from the per-league
`league-{GUID}.sav` files. It tracks game-wide state (the league registry, options, achievements,
team logo/customization data) rather than any single league's data. It lives alongside the
per-league `.sav` files at `{LOCALAPPDATA}\Metalhead\Super Mega Baseball 4\{steam_id}\master.sav`
on Windows (and the equivalent Proton path on Linux).

**Format**: zlib-compressed SQLite 3, same as per-league `.sav` files (confirmed via the `78 01`
zlib header on a live file — *not* a PKZIP container, despite an earlier tool incorrectly assuming
otherwise; see `docs/league-transfer/failure-analysis.md` Bug #2 for the full story).

This schema was verified directly against a real, decompressed `master.sav` and a live SMB4
install during the league-transfer research/validation work in `docs/league-transfer/`. It has 22
tables total (excluding `sqlite_sequence`). Only `t_league_savedatas` is written to by the League
Transfer feature; the rest are documented here for completeness and future reference, but are out
of scope for any current smb-tools functionality.

---

## League Registry

### `t_league_savedatas`

**The table League Transfer writes to.** The game's registry of every league it knows about for
this Steam profile — this is what makes a league appear in the in-game league list at all.

| Column | Type | Notes |
|--------|------|-------|
| `GUID` | `BLOB`, primary key, not null | **16-byte binary GUID — not a string.** Encoding confirmed against real rows: take the standard hyphenated uppercase GUID string, strip the hyphens, hex-decode left-to-right. No byte-swapping, no .NET/COM mixed-endian layout. Matches `uuid.UUID.String()`'s byte order directly (Go's `github.com/google/uuid`), and corresponds to the GUID in that league's `league-{GUID}.sav` filename. |
| `isMissing` | `BOOL`, not null, default `0` | Observed always `0` for normal, present leagues in real data. Exact semantics of a `1` value (e.g., a league whose save file the game can no longer find) not confirmed — not exercised by League Transfer. |

**The confirmed root cause of the original league-transfer freeze** (see
`docs/league-transfer/failure-analysis.md`) was a legacy tool binding `GUID` as a 36-character
uppercase string instead of this 16-byte blob. SQLite's `BLOB`-affinity column stores whatever
type is bound with no coercion, so the malformed row was accepted silently by SQLite — the failure
surfaced only in the game's native code when it later read the column expecting a fixed-size
binary key.

No other table in `master.sav` declares a foreign key against `t_league_savedatas.GUID` —
confirmed via `PRAGMA foreign_key_list` against every table in a real `master.sav`. Registering a
league is a single-row insert into this table alone.

---

## Global Options & State

### `t_options`

Generic global game options as key/value pairs.

| Column | Notes |
|--------|-------|
| `ID` | Primary key |
| `name` | `CHAR(20)` |
| `value` | `CHAR(20)` |

### `t_user_preferences`

A large, single-row (`lock` is a `CHECK(lock = 1)` singleton pattern) table of user preferences:
ego/difficulty settings, matchmaking team preferences, new-league defaults
(`customizationNewLeagueNumConferences`/`NumDivisions`/`NumTeams`), and a long list of
`awardLevel*` columns tracking in-game award progression. Not relevant to League Transfer.

### `t_save_data_validity`

Singleton (`lock = 1`) flag table.

| Column | Notes |
|--------|-------|
| `isUserModified` | `BOOL` — the game's own "has this save been tampered with" flag. Not set or read by League Transfer; worth keeping in mind if a future investigation needs to understand how the game detects modified saves. |

### `t_input_mappings`

Controller/keyboard input bindings. Unrelated to league data.

### `t_chat_message_prefs`

Per-context (`contextID`) quick-chat message slot mappings (8 directional slots). Online
multiplayer feature, unrelated to league data.

### `t_achievements`

Key/value achievement progress tracking.

### `t_game_mode_mechanics_seen`, `t_help_stories_seen`, `t_hidden_item_notifications`, `t_logo_editor_notification_seen`, `t_num_promotional_leagues_seen`

Small "has the user seen this tutorial/notification" tracking tables, each effectively a single
integer or boolean primary key with no associated data. Not relevant to League Transfer.

---

## Team & Logo Customization (Global)

These mirror per-league team/logo concepts but at the global/profile level — e.g., custom team
attribute overrides and logos that exist independent of any specific league's teams.

### `t_teams`

Global team registry — includes built-in and user-customized teams independent of any league.

| Column | Notes |
|--------|-------|
| `GUID` | Primary key, `BLOB` |
| `originalGUID` | `BLOB`, nullable — historical link, same pattern as the per-league `t_leagues.originalGUID` |
| `teamName` | `TEXT NOT NULL` |
| `isBuiltIn`, `isGenerated`, `isHistorical` | `BOOL` flags |
| `teamType` | FK to `t_team_types` |
| `templateTeamFamily` | Nullable `INTEGER`; constrained so only `teamType = 2` (template) rows may set it |

### `t_team_types`

Lookup: `teamType` (PK, `INTEGER`) -> `typeName` (`TEXT`).

### `t_team_local_ids`

Maps an autoincrementing `localID` to a team `GUID` (FK to `t_teams`, cascade delete) — the same
local-ID-indirection pattern seen elsewhere in the save format (e.g., `t_league_local_ids` in a
per-league save).

### `t_team_attributes`

Generic key/value attribute overrides per team, keyed by `teamLocalID` (FK to
`t_team_local_ids.localID`). `optionKey`/`colorKey`/`optionType` encode *what* the value means;
`optionValueInt`/`optionValueFloat`/`optionValueText` hold the value itself, only one populated per
row depending on `optionType`.

### `t_team_logos`, `t_team_logo_attributes`, `t_team_logo_types`, `t_team_logo_element_types`

Logo composition data: each logo (`t_team_logos`, keyed by its own `GUID`, FK to a `teamGUID`) is
built from layered elements (`logoType`, `logoElementType`) with position/rotation/scale/color/font
attributes. `t_team_logo_attributes` follows the same generic key/value-override pattern as
`t_team_attributes`. Not relevant to League Transfer — per-league team logos live in the league
save itself, not here.

---

## Misc

### `t_custom_pennant_races`

| Column | Notes |
|--------|-------|
| `raceGUID` | Primary key, `BLOB` |
| `lastSeasonGUID` | `BLOB NOT NULL` — not declared as an FK in the schema, but presumably references a season GUID in some league's save |
| `initialSkillLevel` | `INTEGER` |

### `t_franchise_news_filter`

Singleton (`lock = 1`) table of boolean toggles controlling which franchise news event types are
shown to the user. Unrelated to league registration.

---

## Source

Schema verified by decompressing a live `master.sav` (`internal/db`'s zlib decompression, same as
per-league saves) and querying `sqlite_master`/`PRAGMA foreign_key_list` directly, during the
league-transfer validation work — see `docs/league-transfer/validation-results.md` and
`docs/league-transfer/failure-analysis.md` for the investigation that produced this. Per
`internal/CLAUDE.md`'s "Save Game SQL — Real Schema Required" rule, this document — not guesswork —
is the source of truth for any future code touching `master.sav`.
