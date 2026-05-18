# Open Decisions

Decisions that are explicitly pending. Nothing in this file should be assumed to have a chosen answer during implementation. Update this file (and `decisions.md`) when a decision is made.

---

## Frontend Framework

**Status**: Under discussion. No decision made.

**The question**: React, Vue 3, Svelte, or something else?

**Context**:
- The previous smb-tools scaffold used Vue 3 + TypeScript. That scaffold is being discarded, so there is no sunk cost.
- The previous companion app was WPF. The primary motivation for the rewrite is access to the JavaScript ecosystem — specifically, rich datagrid and charting libraries that were unavailable in WPF.
- The app is **very table and visualization heavy**: leaderboards, per-player stat grids, season breakdowns, trend charts, radar/spider charts for player attributes, schedule breakdowns.
- Wails generates TypeScript bindings for Go methods regardless of which JS framework is used.

**Key considerations**:
- React has a wider ecosystem and more examples for data-heavy dashboard applications
- Vue 3 is what was previously started with; the developer has some prior familiarity
- Svelte compiles to vanilla JS, smaller runtime, good DX — but smaller ecosystem for the specialized libraries this app needs
- Wails v2 has official templates for Vue, React, Svelte, and others

**What the decision unlocks**: component library choice, datagrid library choice, charting library choice, routing approach.

---

## Datagrid Library

**Status**: Pending frontend framework decision.

**The question**: AG Grid Community, TanStack Table, or something else?

**Context**:
- This app's primary UI surface is data grids: career leaderboards (sortable, filterable, paginated), per-player season-by-season breakdowns, team roster views, schedule grids
- The original WPF DataTable was a constant source of frustration — limited capability
- **AG Grid Community** (free tier): the gold standard for web data grids. Handles very large datasets, built-in sorting/filtering/pagination, virtual scrolling, good accessibility. More opinionated about styling.
- **TanStack Table**: headless — provides logic, not UI. Pairs with your own rendering. Maximum flexibility, more implementation work.

**Leaning**: AG Grid Community, unless the frontend framework decision makes TanStack a more natural fit (e.g., if going with a fully headless component philosophy).

---

## Charting / Visualization Library

**Status**: Pending frontend framework decision.

**The question**: Apache ECharts, Chart.js, Recharts (React), or something else?

**Context**:
- The companion app used ScottPlot (C# library) and was limited to basic charts
- This app needs: line/bar charts (team performance trends over a season), radar/spider charts (player attribute percentile visualization), possibly scatter plots (player comparisons)
- **Apache ECharts**: very capable, good Vue and React integrations, handles complex chart types including radar, good performance on large datasets
- **Chart.js**: simpler, fine for basic charts, not as capable for complex types
- **Recharts**: React-specific, good for dashboards, built on D3
- **Highcharts**: powerful but commercial license for commercial use

**Leaning**: Apache ECharts regardless of framework choice — it has strong bindings for both Vue (`vue-echarts`) and React (`echarts-for-react`).

---

## UI Component Library

**Status**: Pending frontend framework decision.

**The question**: What provides the base UI components (buttons, inputs, modals, tabs, etc.)?

**Context**:
- The previous scaffold used PrimeVue. The user is not tied to it.
- The app has a dark theme requirement and is desktop-context (not mobile-responsive)
- Some strong options depending on framework:
  - **shadcn/ui** (React) + Tailwind: copy-paste component model, highly customizable, very active ecosystem, dark mode excellent
  - **shadcn-vue** (Vue port): same philosophy, slightly less mature than the React original
  - **PrimeVue** (Vue): comprehensive, has its own DataTable (less powerful than AG Grid)
  - **Radix UI / Ark UI**: headless primitives, framework-specific
  - **Mantine** (React): batteries-included, good dark mode

**The datagrid and charting libraries are more important than the component library** for this app. The component library handles chrome (nav, buttons, forms, modals); the datagrid and charting libraries handle the core UI surface. Don't let component library preference override datagrid/charting capability.

---

## sqlc

**Status**: Under evaluation during initial scaffolding.

**The question**: Use `sqlc` to generate type-safe Go from SQL query files, or write store methods by hand using `database/sql`?

**Context**:
- `sqlc` reads `.sql` query files and generates Go structs + functions. Works naturally alongside golang-migrate.
- Keeps SQL in `.sql` files (diffable, reviewable) rather than embedded in Go strings
- Generated code uses `database/sql` directly — no ORM magic, transparent behavior
- Adds a code generation step to the build workflow (not automatic — must run `sqlc generate` when queries change)
- The two-database situation (save game DB + companion DB) needs to be modeled correctly in sqlc config — two separate `sqlc.yaml` configs or a single config with multiple schemas

**Decision trigger**: Evaluate during the initial backend scaffolding. If sqlc setup for the two-DB scenario is clean, adopt it. If it adds friction, write queries by hand.

---

## New Companion Database Schema

**Status**: Deferred pending schema analysis.

**The question**: What is the new companion database schema?

**Process**:
1. Conduct schema analysis: compare old SmbExplorerCompanion EF Core schema against the actual queries and operations the app performed. Identify shortfalls, redundancies, and improvement opportunities. Document in `docs/architecture/schema-analysis.md`.
2. Design new schema against the principles in `data-layer.md`.
3. Write initial golang-migrate SQL files.
4. Document the old → new column mapping for the legacy migration service.

**This must happen before any store or service code is written.**

---

## Frontend Testing Tooling

**Status**: Pending frontend framework decision.

**The question**: Vitest + Testing Library? Playwright? Both?

**Baseline expectation** (from `testing-strategy.md`): component-level tests for the most critical UI flows (import wizard, leaderboard filtering, stats display). Playwright for any flows that require multi-step interaction. Exact tooling depends on the framework decision.
