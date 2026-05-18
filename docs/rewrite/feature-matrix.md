# Feature Matrix

Side-by-side mapping of features across SMB3Explorer, SmbExplorerCompanion, and the smb-tools rewrite target.

**Status legend**:
- `✓` — Feature present and complete
- `~` — Partial / planned but incomplete
- `—` — Not present
- `New` — Net new feature not in either original

---

## Save File Handling

| Feature | SMB3Explorer | Companion | smb-tools |
|---------|-------------|-----------|-----------|
| ZLib decompression of .sav file | ✓ | — | ✓ (partial) |
| Auto-detect save file location | ✓ | — | ✓ |
| Manual save file selection | ✓ | — | ✓ |
| Pre-decompressed .sqlite support | ✓ | — | ✓ |
| Read-only SQLite access | ✓ | — | ✓ |
| SMB3 save file support | ✓ | — | ✓ |
| SMB4 save file support | ✓ | — | ✓ |
| Multi-league SMB4 GUID parsing | ✓ | — | ✓ |
| Do not run while game is active (constraint) | ✓ | — | ✓ |

---

## League / Franchise Selection

| Feature | SMB3Explorer | Companion | smb-tools |
|---------|-------------|-----------|-----------|
| List available leagues from save | ✓ | — | ✓ |
| Previously accessed league history | ✓ | — | ✓ |
| Franchise mode support | ✓ | ✓ | ✓ |
| Season mode support | ✓ | — | ✓ |
| Elimination mode support | ✓ | — | ✓ |
| Multi-franchise management | — | ✓ | ✓ |

---

## CSV Export (SMB3Explorer)

In the rewrite, these become in-app views rather than file exports. CSV export may still be offered as an optional download.

| Feature | SMB3Explorer | Companion | smb-tools |
|---------|-------------|-----------|-----------|
| Career batting stats (reg season) | ✓ | — | ✓ |
| Career batting stats (playoffs) | ✓ | — | ✓ |
| Career pitching stats (reg season) | ✓ | — | ✓ |
| Career pitching stats (playoffs) | ✓ | — | ✓ |
| Season batting stats (reg season) | ✓ | — | ✓ |
| Season batting stats (playoffs) | ✓ | — | ✓ |
| Season pitching stats (reg season) | ✓ | — | ✓ |
| Season pitching stats (playoffs) | ✓ | — | ✓ |
| Season standings | ✓ | — | ✓ |
| Playoff standings | ✓ | — | ✓ |
| Most recent season players (attributes + traits) | ✓ | — | ✓ |
| Most recent season teams | ✓ | — | ✓ |
| Most recent season schedule | ✓ | — | ✓ |
| Most recent season playoff schedule | ✓ | — | ✓ |
| Top batting performers | ✓ | — | ✓ |
| Top pitching performers | ✓ | — | ✓ |
| Top rookies (batting + pitching) | ✓ | — | ✓ |

---

## Franchise History Database (Companion)

| Feature | SMB3Explorer | Companion | smb-tools |
|---------|-------------|-----------|-----------|
| Persistent franchise DB | — | ✓ | ✓ |
| Import from CSV files | — | ✓ | — (replaced by direct read) |
| Import directly from save file | — | — | ✓ |
| Multi-season history beyond game's 50-season limit | — | ✓ | ✓ |
| Player career stat accumulation | — | ✓ | ✓ |
| Team name/logo history | — | ✓ | ✓ |
| Intra-season player trade tracking | — | ✓ | ✓ |

---

## Statistics Viewer (Companion)

| Feature | SMB3Explorer | Companion | smb-tools |
|---------|-------------|-----------|-----------|
| Home screen (franchise dashboard) | — | ✓ | ✓ |
| Franchise summary stats | — | ✓ | ✓ |
| Global search (players + teams) | — | ✓ | ✓ |
| Player overview page | — | ✓ | ✓ |
| Player career stats (batting + pitching) | — | ✓ | ✓ |
| Player season-by-season breakdown | — | ✓ | ✓ |
| Player game attribute history | — | ✓ | ✓ |
| Player attribute percentile rankings | — | ✓ | ✓ |
| Player KPI percentile rankings | — | ✓ | ✓ |
| Player visualizations (radar chart, etc.) | — | ✓ | ✓ |
| Similar players recommendations | — | ✓ | ✓ |
| Team overview page | — | ✓ | ✓ |
| Team season detail page | — | ✓ | ✓ |
| Team schedule breakdown | — | ✓ | ✓ |
| Team performance trend visualization | — | ✓ | ✓ |
| Top batting careers leaderboard | — | ✓ | ✓ |
| Top batting seasons leaderboard | — | ✓ | ✓ |
| Top pitching careers leaderboard | — | ✓ | ✓ |
| Top pitching seasons leaderboard | — | ✓ | ✓ |
| Leaderboard filters (position, chemistry, handedness) | — | ✓ | ✓ |
| Leaderboard season range filter | — | ✓ | ✓ |
| HoF-only filter on leaderboards | — | ✓ | ✓ |
| Historical teams list | — | ✓ | ✓ |

---

## Awards & Hall of Fame

| Feature | SMB3Explorer | Companion | smb-tools |
|---------|-------------|-----------|-----------|
| Manual award assignment (MVP, CYA, etc.) | — | ✓ | ✓ |
| Auto title awards (BA, HR, RBI, ERA, W, K) | — | ✓ | ✓ |
| Triple Crown auto-detection | — | ✓ | ✓ |
| Award display on player profiles | — | ✓ | ✓ |
| Hall of Fame management | — | ✓ | ✓ |
| Hall of Fame eligibility evaluation | — | ✓ | ✓ |
| Custom user-created awards | — | ✓ | ✓ |

---

## New Features (smb-tools only)

| Feature | SMB3Explorer | Companion | smb-tools |
|---------|-------------|-----------|-----------|
| Team transfer tool | — | — | New |
| Cross-platform support (macOS, Linux) | — | — | New |
| Unified single-app experience | — | — | New |
| No CSV intermediary required | — | — | New |

---

## Administrative / UX

| Feature | SMB3Explorer | Companion | smb-tools |
|---------|-------------|-----------|-----------|
| Automatic update checking | ✓ | — | TBD |
| Export data purge | ✓ | — | TBD |
| Custom SQL query access | — | ✓ (via external tool) | TBD |
| Dark theme | ~ | ✓ | ✓ |
| Windows installer | ✓ | ✓ | ✓ (NSIS) |
| macOS app bundle | — | — | ✓ |
