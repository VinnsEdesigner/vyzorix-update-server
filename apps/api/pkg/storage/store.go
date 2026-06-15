// Package storage provides SQLite database operations.
//
// Modular structure:
//   - store.go: Store struct, Open/Close, connection management
//   - migrations.go: Database schema migrations registry
//   - devices.go: Device CRUD operations
//   - telemetry.go: Telemetry storage (UUIDv7)
//   - commands.go: Command dispatch and status
//   - operators.go: Operator CRUD, sessions, verifications
//   - settings.go: System settings, auth sessions, email/password tokens
//   - uuid.go: UUIDv7 generation utilities
//   - crypto.go: Argon2id password hashing utilities
package storage

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// ErrNotFound indicates the requested resource does not exist.
var ErrNotFound = errors.New("not found")

// ErrHijack indicates device_id is already registered to a different firebaseInstallId.
var ErrHijack = errors.New("device_id already registered to a different firebaseInstallId")

// Store represents a SQLite database connection.
type Store struct {
	db   *sql.DB
	path string
	mu   sync.Mutex
}

// Open opens a SQLite database at the given path, creating it if necessary.
// Sets optimal SQLite pragmas for performance and reliability.
func Open(path string) (*Store, error) {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_cache_size=-2000&_busy_timeout=5000&_foreign_keys=1")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, path: path}
	if err := s.setupPragmas(context.Background()); err != nil {
		return nil, err
	}
	if err := RunMigrations(db); err != nil {
		return nil, err
	}
	return s, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// Ping checks the database connection.
func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// setupPragmas configures optimal SQLite settings.
func (s *Store) setupPragmas(ctx context.Context) error {
	pragmas := []string{
		`PRAGMA journal_mode=WAL`,
		`PRAGMA cache_size=-2000`,
		`PRAGMA busy_timeout=5000`,
		`PRAGMA foreign_keys=ON`,
	}
	for _, pragma := range pragmas {
		if _, err := s.db.ExecContext(ctx, pragma); err != nil {
			return err
		}
	}
	return nil
}