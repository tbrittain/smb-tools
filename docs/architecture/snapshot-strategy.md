# Save Game Snapshot Strategy

Every time the app reads from an SMB save file and the content differs from the last known snapshot, it persists a permanent copy of the decompressed SQLite database. These snapshots are the canonical source of truth for all game data.

---

## Why Snapshots Are Critical

The SMB game engine compacts franchise data after each season. Once the offseason is simulated:
- Per-season granular data from earlier seasons may be reduced to summary records or lost entirely
- Career stats for retired players may become inaccessible
- Schedule and game result data from prior seasons is typically gone

**The save file as it exists at the end of a season is the only complete record of that season.** Once the user advances past it, that data cannot be recovered from the game — only from a snapshot taken before the advance.

This means:

1. If we later identify data in the game schema that we did not import into our companion DB, we can go back to the snapshot and extract it — provided we captured the snapshot before the season was advanced.
2. Future schema migrations (new companion DB features) can draw from snapshots rather than requiring re-play of the franchise.
3. If our import logic had a bug that corrupted or missed data, the snapshot is the source of recovery.

Snapshots are never deleted lightly. They are the backup that makes every other decision reversible.

---

## Storage Layout

```
{app_data}/
  franchises/
    {franchise_id}/
      companion.db                    # The companion DB for this franchise
      snapshots/
        {season_num}_{sha256_short}.sqlite      # Uncompressed snapshot
        {season_num}_{sha256_short}.sqlite.zst  # Compressed version (if compacted)
```

`{sha256_short}` is the first 12 characters of the SHA-256 hash of the decompressed SQLite content. This makes filenames human-readable (season number visible) while guaranteeing uniqueness.

Example:
```
franchises/
  a3f7c2d1/
    companion.db
    snapshots/
      007_4a8b3c912f1d.sqlite
      008_e2f09a4c7b3e.sqlite
      008_e2f09a4c7b3e.sqlite.zst   # same snapshot, compressed
```

---

## Deduplication by Hash

Before persisting a snapshot, the app computes the SHA-256 hash of the decompressed save game bytes and compares it to the most recent snapshot for that franchise.

- **If hashes match**: the save file has not changed since the last snapshot. No new file is written. The app still reads and processes the data (in case the user wants to re-sync to repair their companion DB), but does not create a duplicate snapshot.
- **If hashes differ**: a new snapshot file is written. This is the normal case after a season has been played.

The hash comparison is against the most recent snapshot only — not all snapshots. Two non-consecutive snapshots with the same hash are an edge case that can coexist without issue.

---

## Compression Strategy

Snapshots accumulate over the life of a franchise. Compression reduces storage cost for older snapshots that are unlikely to be accessed frequently.

**Policy**:
- Snapshots from the **two most recent seasons** are kept uncompressed (fast access for re-sync or debugging)
- Older snapshots are compressed using **zstd** (fast compression, good ratio, supported in pure Go via `github.com/klauspost/compress/zstd`)
- Both the compressed and uncompressed versions are NOT kept simultaneously — compression replaces the uncompressed file in-place
- A compressed snapshot can be decompressed on demand when needed for data recovery

**Trigger**: Compression of eligible snapshots runs as a background task at app startup or after a successful sync. It is never on the critical path of the sync operation itself.

---

## Storage Display

The app surfaces snapshot storage to the user:
- Total size of all snapshots for the active franchise
- Optionally, a breakdown by season (how many snapshots, total size)
- A "Storage" section in franchise settings

This information is **read-only in normal use**. The UI does not offer a "Delete snapshots" button in any primary flow.

---

## Deletion Policy

Snapshots should be very difficult to delete accidentally:

- No bulk delete option in the primary UI
- If deletion is ever supported, it requires: explicit acknowledgment that the data cannot be recovered from the game, confirmation of the specific seasons being deleted, and ideally a second confirmation
- The companion DB is NOT a substitute for the snapshot. The companion DB holds our adapted schema; the snapshot holds the complete raw game data. Deleting a snapshot loses data that may not exist anywhere else.

**Recommendation**: Do not implement snapshot deletion at all in v1. If storage becomes a user concern, address it through better compression, not deletion.

---

## Snapshot Metadata

A lightweight record of each snapshot is maintained in the franchise companion DB (not the snapshot file itself):

```sql
CREATE TABLE save_game_snapshots (
    id          INTEGER PRIMARY KEY,
    season_num  INTEGER NOT NULL,
    captured_at DATETIME NOT NULL,
    file_name   TEXT NOT NULL,      -- relative path within snapshots/
    sha256_hash TEXT NOT NULL,
    file_size   INTEGER NOT NULL,   -- bytes, uncompressed
    compressed  INTEGER NOT NULL DEFAULT 0,  -- boolean
    compressed_size INTEGER         -- bytes, if compressed
);
```

This allows the app to list and reason about snapshots without scanning the filesystem, and provides the hash needed for deduplication without re-reading the snapshot files.

---

## Recovery Use Case (Future)

When a future data migration needs to draw from snapshots:

```go
// Pseudocode for a future recovery operation
snapshots, _ := franchiseStore.ListSnapshots(ctx, franchiseID)
for _, snap := range snapshots {
    db, _ := openSnapshotDB(snap.FilePath)  // decompress if needed, open read-only
    reader := store.NewSqliteSaveGameReader(db)
    // extract whatever new data we need
    // write to companion DB via store layer
    db.Close()
}
```

The `SaveGameReader` interface is the same interface used during normal sync — no special recovery code path needed beyond the snapshot enumeration and opening logic.
