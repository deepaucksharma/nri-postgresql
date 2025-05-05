// store/postgres/store.go
package postgres

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/newrelic/nri-postgresql/src/connection"
	"github.com/newrelic/nri-postgresql/src/metrics"
)

// Store is a thin wrapper around *sqlx.DB that provides
// context-aware helpers and sensible time-outs for every call.
type Store struct {
	db      *connection.PGSQLConnection
	timeout time.Duration
}

// New returns a Store with the default (5 s) statement timeout.
func New(db *connection.PGSQLConnection) *Store {
	return &Store{
		db:      db,
		timeout: 5 * time.Second,
	}
}

// Timeout allows callers to override the default statement timeout
// (useful for bulk operations in CLIs / integration tests).
func (s *Store) Timeout(d time.Duration) {
	s.timeout = d
}

// withTimeout guarantees that every query runs with a context that
// is cancelled after s.timeout. If the caller already supplied a
// context, we inherit its deadline and use the shorter of the two.
func (s *Store) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		return context.WithTimeout(context.Background(), s.timeout)
	}

	// Respect the caller's deadline if it's sooner.
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining < s.timeout {
			return context.WithCancel(ctx) // caller's ctx already stricter
		}
	}
	return context.WithTimeout(ctx, s.timeout)
}

// ----------------------------------------------------------------------
// Example queries
// ----------------------------------------------------------------------

// Extension represents a row from pg_extension.
type Extension struct {
	Name    string `db:"extname"`
	Enabled bool   `db:"enabled"`
}

// ListExtensions returns all installed extensions, guarded by a timeout.
func (s *Store) ListExtensions(ctx context.Context) ([]Extension, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	var exts []Extension
	err := s.db.QueryContext(ctx, &exts,
		`SELECT extname, TRUE AS enabled FROM pg_extension`)
	return exts, err
}

// ServerVersion wraps metrics.CollectVersion so callers don't need to
// import that package directly.
func (s *Store) ServerVersion(ctx context.Context) (string, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	ver, err := metrics.CollectVersion(ctx, s.db) // new ctx-aware helper
	if err != nil {
		return "", err
	}
	return ver.String(), nil
}

// ----------------------------------------------------------------------
// Generic helpers
// ----------------------------------------------------------------------

// Get is a generic helper that loads a single row into dest.
// Typical use:
//
//	var user User
//	if err := store.Get(ctx, &user, `SELECT * FROM users WHERE id=$1`, id); err != nil {
//	    ...
//	}
func (s *Store) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	return s.db.QueryContext(ctx, dest, query)
}

// Select is the plural form of Get.
func (s *Store) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	rows, err := s.db.QueryxContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	return sqlx.StructScan(rows, dest)
}
