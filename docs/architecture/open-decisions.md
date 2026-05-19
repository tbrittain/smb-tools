# Open Decisions

Decisions that are explicitly pending. Nothing in this file should be assumed to have a chosen answer during implementation. Update this file (and `decisions.md`) when a decision is made.

---

## Frontend Testing Tooling

**Status**: Pending — follow-up after initial scaffolding.

**The question**: Vitest + Testing Library? Playwright? Both?

**Baseline expectation** (from `testing-strategy.md`): component-level tests for the most critical UI flows (import wizard, leaderboard filtering, stats display). Playwright for any flows that require multi-step interaction.

**Likely answer**: Vitest for unit/component tests (native Vite integration, fast), Playwright for E2E flows. Decision to be confirmed during scaffolding.

---

## sqlc

**Status**: Under evaluation during initial scaffolding.

**The question**: Use `sqlc` to generate type-safe Go from SQL query files, or write store methods by hand using `database/sql`?

**Context**:
- `sqlc` reads `.sql` query files and generates Go structs + functions. Works naturally alongside golang-migrate.
- Keeps SQL in `.sql` files (diffable, reviewable) rather than embedded in Go strings
- Generated code uses `database/sql` directly — no ORM magic, transparent behavior
- Adds a code generation step to the build workflow (not automatic — must run `sqlc generate` when queries change)
- The two-database situation (save game DB + companion DB) needs to be modeled correctly in sqlc config

**Decision trigger**: Evaluate during the initial backend scaffolding. If sqlc setup for the two-DB scenario is clean, adopt it. If it adds friction, write queries by hand.

---

## New Companion Database Schema

**Status**: Deferred — requires schema analysis before any store code is written.

**Process**:
1. Conduct schema analysis: compare old SmbExplorerCompanion EF Core schema against the actual queries performed. Identify shortfalls, redundancies, improvement opportunities. Also incorporate newly discovered tables from the bundled `league-template.sqlite` (franchise news events, team logos, pitch counts, etc.). Document in `docs/architecture/schema-analysis.md`.
2. Design new schema against the principles in `data-layer.md`.
3. Write initial golang-migrate SQL files.
4. Document the old → new column mapping for the legacy migration service.

---

## Future: Intra-Season Event Tracking

**Status**: Not planned for MVP. Noted for future discussion.

**The idea**: Currently the app captures a *snapshot* at one point in time per season — whatever the save game contains when the user runs a sync. Player attributes (Power, Contact, Velocity, etc.) are stored as the value at sync time; there is no history of how those values changed *during* the season.

In franchise mode, players' attributes can change mid-season through:
- Training outcomes
- Manager moments (small attribute adjustments)
- Trait changes
- Injuries affecting performance

The `t_franchise_news_*` tables in the save game (documented in `docs/game-integration/investigation.md`) already persist these as typed events. A future enhancement could:
1. Import these events into the companion DB (new tables: `player_attribute_events`, `player_trait_events`, etc.)
2. Track the full trajectory of a player's attributes across a season, not just the end-of-season snapshot
3. Enable visualizations like "Power over time" charts on player profile pages

**Interaction with fsnotify (Phase 13)**: Once the app can auto-sync on every save file write, the granularity of captured data increases from "per season" to "per game" or "per training session." This would be the natural enabler for intra-season tracking — the data is already in the save game; it just requires more frequent syncing and new companion schema tables to store the event log.

**Why not now**: Requires additional `SaveGameReader` methods for the news tables, new companion DB schema tables, and UI to display the richer data. Scoped to a post-MVP phase after the core Baseball Reference–style views are working.
