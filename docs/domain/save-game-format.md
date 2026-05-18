# Save Game Format

## File Format

SMB save game files are **ZLib-compressed SQLite 3 databases**. The decompressed content is a standard SQLite file and can be opened with any SQLite-compatible tool.

Compression: standard ZLib inflate (deflate algorithm, zlib wrapper). No custom encryption or obfuscation.

## Default File Locations

| Game | Path |
|------|------|
| SMB3 | `%LOCALAPPDATA%\Metalhead\Super Mega Baseball 3\{steam_id}\savedata.sav` |
| SMB4 | `%LOCALAPPDATA%\Metalhead\Super Mega Baseball 4\{steam_id}\{filename}.sav` |

The `{steam_id}` directory corresponds to the user's Steam account ID. For SMB4, the filename encodes the league GUID (see below).

## SMB4 Multi-League Save Files

SMB4 stores multiple leagues in a single save file but may also use per-league `.sav` files with the league GUID in the filename (e.g., `league-{guid}.sav`). The GUID is parsed from the filename to identify which league to load when multiple leagues exist in a single database.

## Decompression Approach (from SMB3Explorer)

1. Read the `.sav` file as raw bytes
2. Apply ZLib inflate (standard `zlib` decompress)
3. Write decompressed bytes to a temp file
4. Open the temp file as a read-only SQLite connection

Temp file naming convention used by SMB3Explorer: `smb3_explorer_yyyyMMddHHmmssfff.sqlite` written to `%TEMP%`.

In the Go rewrite (`smb-tools`), the decompressed file is written to the OS-appropriate app data directory via `backend/config/app_directories.go`.

## Critical Constraints

- **Read-only access only.** The tooling must never write to or modify the save file. The SQLite connection should be opened in read-only mode.
- **Do not access while game is running.** The game holds a lock on the save file while active. Attempting to read during gameplay may produce corrupted or incomplete data.
- **SMB3Explorer auto-detects** the save file location but also allows manual file selection and supports pre-extracted (already decompressed) `.sqlite` files.

## Export Output Location (SMB3Explorer)

Exported CSV files are written to: `%LOCALAPPDATA%\SMB3Explorer\`

Users can purge previously exported data via the app's File menu.

## Companion App Database Location

SmbExplorerCompanion maintains its own separate SQLite database at:
`%LOCALAPPDATA%\SmbExplorerCompanion\SmbExplorerCompanion.db`

This database is NOT the game's save file — it is the companion's own persistent store for historical data.
