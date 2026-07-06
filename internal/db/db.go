// Package db is the admin console's read-only data layer over the same
// PostgreSQL the server owns (see docs/database.md). It re-declares its own
// queries against the shared schema — no cross-repo import of the server — and
// exposes typed rows the JSON API serves verbatim. Writes (authoring, fixtures,
// the change-flag reconcile) are a later phase; this layer only reads.
package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned by the single-row lookups (User, Object) when no row
// matches the id, so the API can map it to a 404.
var ErrNotFound = errors.New("not found")

// DB wraps a pgx connection pool. It is safe for concurrent use by the HTTP
// handlers (pgxpool multiplexes connections).
type DB struct {
	pool *pgxpool.Pool
}

// Open dials the DSN and verifies connectivity with a ping. A failure is
// returned (not fatal to the process): main logs it and serves the API as 503s
// so the console still boots and binds — see the "crash on start" note in the
// plan.
func Open(ctx context.Context, dsn string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return &DB{pool: pool}, nil
}

// Close releases the pool.
func (d *DB) Close() {
	if d.pool != nil {
		d.pool.Close()
	}
}
