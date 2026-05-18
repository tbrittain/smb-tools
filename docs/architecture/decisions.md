# Technology Decisions

Committed choices for the smb-tools rewrite. These are not up for re-evaluation during implementation — if a decision needs to be revisited, update this document with the reasoning.

---

## Desktop Framework: Wails v2

**Decision**: Wails v2 (not Tauri, not Electron, not anything else).

**Rationale**:
- An existing, working reference application built with Wails v2 provides a known-good pattern for data-heavy Go + WebView apps
- Go is the preferred backend language (see below); Wails v2 is the most mature Go-native desktop framework
- Electron is explicitly ruled out — too heavy, Chromium-bundled, poor fit for a performance-conscious cross-platform app
- Tauri (Rust) is a strong alternative but carries a learning tax: no existing Rust reference app, longer build times, more boilerplate for SQLite-heavy data work
- Wails v3 exists but is still in RC; v2 is stable and production-proven. Migration to v3 is a future option

**Implication**: The repo will be re-scaffolded from scratch using `wails init`. The previous scaffold is discarded entirely.

---

## Backend Language: Go

**Decision**: Go for all backend logic.

**Rationale**:
- Idiomatic Go is well-suited to this app's core operations: file I/O, SQLite queries, data transformation, concurrent save file processing
- The existing reference application demonstrates Go's viability for data-heavy Wails apps
- Go's standard library, interface model, and testing tooling align well with the testability requirements (see `testing-strategy.md`)
- No Rust, no Node.js in the backend

---

## SQLite Driver: modernc.org/sqlite

**Decision**: `modernc.org/sqlite` as the SQLite driver for all database connections.

**Rationale**:
- Pure Go implementation — no CGO dependency, no C toolchain required on any platform
- Critical for cross-platform builds (Windows, macOS, Linux) without per-platform C compiler setup
- `mattn/go-sqlite3` is more battle-tested and marginally faster but requires CGO, which creates real friction in cross-compilation and CI pipelines
- The performance difference is not meaningful for this application's query patterns

---

## Database Migrations: golang-migrate

**Decision**: `golang-migrate` for companion database schema migrations, driven by plain SQL migration files.

**Rationale**:
- Migrations are plain `.sql` files — readable, diffable, no ORM lock-in
- Runs automatically at app startup against the companion DB
- Pairs naturally with `sqlc` if that tool is adopted (see `open-decisions.md`)
- Gives a clear, auditable history of schema evolution

**Scope**: golang-migrate manages only the companion database (the app's own schema). The SMB save game database schema is not owned or migrated by this app.

---

## No CQRS, No MediatR-style Patterns

**Decision**: Idiomatic Go. No CQRS, no command/query mediator, no over-abstraction.

**Rationale**:
- The SmbExplorerCompanion used MediatR + CQRS, which was an over-abstraction imported from .NET culture. It added indirection without meaningful benefit for the scale of this application.
- Go's idioms favor explicit, flat, readable code over pattern-driven architecture
- A store layer + service layer is sufficient. Functions call other functions. Dependencies are passed in explicitly.
- If a pattern isn't solving a real problem present in this codebase right now, it doesn't belong here

---

## UI: JavaScript + HTML + CSS (Framework TBD)

**Decision**: The frontend is a web UI rendered via Wails' WebView — JavaScript, HTML, and CSS. No native UI toolkit (no WPF, no Qt, no SwiftUI).

**Rationale**:
- The original WPF applications were a constant source of frustration — limited datagrid capability, no access to JavaScript charting/plotting libraries, Windows-only
- The JavaScript ecosystem has the best tooling for the data-heavy, visualization-rich UI this app requires (AG Grid, Apache ECharts, TanStack Table, etc.)
- Cross-platform consistency: a web UI renders the same on Windows, macOS, and Linux via each platform's WebView

**Framework**: Not yet decided. See `open-decisions.md` for the pending discussion on React vs. Vue vs. alternatives, component library, datagrid library, and charting library.

---

## No Electron

**Decision**: Electron is ruled out.

**Rationale**: Bundles a full Chromium instance (~150MB+), slower startup, higher memory usage, and adds complexity without benefit over Wails which uses the platform's native WebView. This is not negotiable.

---

## Original Schema: Not Preserved

**Decision**: The new companion database schema is designed from scratch. The SmbExplorerCompanion EF Core schema is a reference, not a constraint.

**Rationale**:
- The original schema had shortfalls: computed stats stored redundantly alongside raw counts (two sources of truth), schema choices driven by EF Core conventions rather than SQLite idioms, and areas where denormalization was done for query convenience but created consistency issues
- A schema analysis (comparing the old schema against the actual operations the companion performed) will precede schema design. See `data-layer.md` for principles.

**Non-negotiable**: The new app must provide a migration path from an existing `SmbExplorerCompanion.db` to the new schema. Existing users cannot be left behind. See `data-layer.md`.
