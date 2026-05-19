package store

import (
	"context"
	"database/sql"
)

// DBTX is the minimal interface satisfied by both *sql.DB and *sql.Tx.
// All store constructors accept DBTX so they can be used inside or outside
// a transaction without any changes to the store code.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
