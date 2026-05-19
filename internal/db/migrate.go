package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

// runMigrations applies any pending SQL migration files from the given embedded
// filesystem. Files must follow the naming convention {version}_{name}.up.sql
// where version is a zero-padded integer (e.g. 0001_initial.up.sql).
//
// Each migration runs inside a transaction. The applied version is recorded in
// a schema_migrations table that is created automatically.
func runMigrations(ctx context.Context, db *sql.DB, migrations embed.FS, dir string) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INTEGER PRIMARY KEY NOT NULL,
			applied_at DATETIME NOT NULL DEFAULT (datetime('now'))
		)
	`); err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	applied := map[int]bool{}
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("querying applied migrations: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return fmt.Errorf("scanning migration version: %w", err)
		}
		applied[v] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating applied migrations: %w", err)
	}

	entries, err := fs.ReadDir(migrations, dir)
	if err != nil {
		return fmt.Errorf("reading migrations directory %q: %w", dir, err)
	}

	type pending struct {
		version  int
		filename string
	}
	var queue []pending
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			continue
		}
		v, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		if !applied[v] {
			queue = append(queue, pending{version: v, filename: name})
		}
	}
	sort.Slice(queue, func(i, j int) bool { return queue[i].version < queue[j].version })

	for _, m := range queue {
		sqlBytes, err := migrations.ReadFile(dir + "/" + m.filename)
		if err != nil {
			return fmt.Errorf("reading migration file %q: %w", m.filename, err)
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("beginning transaction for migration %d: %w", m.version, err)
		}
		if _, err := tx.ExecContext(ctx, string(sqlBytes)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("applying migration %d (%s): %w", m.version, m.filename, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES (?)`, m.version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("recording migration %d: %w", m.version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %d: %w", m.version, err)
		}
	}
	return nil
}
