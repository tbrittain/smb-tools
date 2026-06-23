# smb-tools

A cross-platform desktop application for **Super Mega Baseball 4** franchise management and statistics.

> **SMB4 only.** Live save game syncing and all domain data (traits, positions, chemistry types) target SMB4 exclusively. SMB3 save files are not supported for import.

## Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| [Go](https://go.dev/dl/) | 1.26+ | Backend language |
| [Node.js](https://nodejs.org/) | 26+ | Frontend toolchain |
| npm | 10+ | Comes with Node 26 |
| [Wails CLI](https://wails.io/docs/gettingstarted/installation) | v2 (not v3) | Desktop app framework |

Install the Wails CLI:
```sh
go install github.com/wailsapp/wails/v2/cmd/wails@v2
```

> **Note:** Do not install `wails@latest` — that may resolve to v3 which has breaking changes. Target the `v2` major version explicitly.

## Development

```sh
# Install frontend dependencies (first time or after package.json changes)
cd frontend && npm install

# Run the full app in dev mode with hot reload
wails dev
```

The Wails dev server embeds both the Go backend and the Vite frontend. Changes to Go files restart the backend; changes to Vue/TypeScript files hot-reload in the WebView.

> **Linux note:** On newer distros (e.g. Ubuntu 24.04+), the system ships `webkit2gtk-4.1` instead of `webkit2gtk-4.0`, which Wails v2 looks for by default. If `wails dev`/`wails build` fails with a `webkit2gtk-4.0` pkg-config error, pass the `webkit2_41` build tag:
> ```sh
> wails dev -tags webkit2_41
> wails build -tags webkit2_41
> ```

You can also run the frontend standalone (without the Go backend) for pure UI work:

```sh
cd frontend && npm run dev
```

> **Note:** Wails bindings will not be available in standalone frontend mode. Mock or stub any backend calls when developing UI in isolation.

## Building

```sh
# Build a native binary for the current platform
wails build
```

Output lands in `build/bin/`.

## Wails TypeScript Bindings

Whenever you add or change exported methods on the `App` struct in Go, regenerate the TypeScript bindings so the frontend stays in sync:

```sh
wails generate module
```

This rewrites `frontend/wailsjs/go/` — commit the result alongside the Go changes.

## Frontend Tooling

```sh
cd frontend

npm run lint          # Biome lint check
npm run lint:fix      # Biome auto-fix
npm run test:run      # Vitest unit tests (one-shot)
npm run test          # Vitest unit tests (watch mode)
npm run storybook     # Launch Storybook component explorer on :6006
npm run build         # TypeScript check + production build
```

### Storybook

Storybook runs independently of the Go backend. Use it to develop, inspect, and visually test UI components in isolation against controlled prop inputs — without needing a live save game or database.

Every non-trivial reusable component in `src/components/` should have a corresponding `.stories.ts` file.

## Go Tooling

```sh
go test ./...          # Run all tests
go vet ./...           # Static analysis
golangci-lint run      # Linting (requires golangci-lint installed)
```

## Further Reading

- [`docs/`](docs/) — architecture decisions, domain knowledge, feature roadmap
- [`CLAUDE.md`](CLAUDE.md) — coding standards and project conventions
