# Backend (Go)

Go backend for smb-tools. The backend exposes data to the Vue frontend via [Wails v2](https://wails.io/) bindings and reads SMB4 save game files directly.

## Setup

Requires Go 1.26

## Commands

```sh
go test ./...          # Run all tests
go vet ./...           # Static analysis
golangci-lint run      # Linting — run before every commit
```

Install `golangci-lint` once (v2 module path):
```sh
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

## Database layout

Each franchise has its own isolated SQLite companion database stored in the app data directory. There is no shared multitenant DB. Switching franchises closes one `*sql.DB` and opens another.

The SMB4 save game is a ZLib-compressed SQLite 3 file. The backend decompresses it in-memory and opens a read-only connection — the save game is never written to. Schema reference: [`docs/domain/save-game-schema.md`](../docs/domain/save-game-schema.md).

## Migrations

SQL migrations live in `internal/db/migrations/` as `{version}_{name}.up.sql` files and are embedded via `embed.FS`. The custom runner in `internal/db/migrate.go` applies them in order on DB open. There are no down migrations.

## Wails bindings

After adding or changing any exported method on the `App` struct, regenerate the TypeScript bindings so the frontend stays in sync:

```sh
wails build
```

This rewrites `frontend/wailsjs/go/`. Commit the result alongside the Go changes. Never hand-edit those generated files.
