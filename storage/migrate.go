package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Migration represents a single database migration
type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

// RunMigrations executes all pending database migrations
func (s *Store) RunMigrations() error {
	// Get migration directory from environment or default
	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		// Default to ./migrations relative to working directory
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		dir = filepath.Join(wd, "migrations")
	}

	// Load migrations from directory
	migrations, err := LoadMigrations(dir)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	if len(migrations) == 0 {
		return nil // No migrations to run
	}

	// Ensure migrations table exists
	if err := s.ensureMigrationsTable(); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	// Get current version
	currentVersion, err := s.getCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Run pending migrations
	for _, m := range migrations {
		if m.Version <= currentVersion {
			continue // Already applied
		}

		slog.Info("applying migration", "version", m.Version, "name", m.Name)

		tx, err := s.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Execute migration
		if _, err := tx.Exec(m.Up); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to apply migration %d: %w", m.Version, err)
		}

		// Record migration
		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)",
			m.Version, m.Name, sql.NullInt64{Int64: 0, Valid: false},
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration: %w", err)
		}
	}

	return nil
}

// LoadMigrations reads all migration files from a directory
func LoadMigrations(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No migrations directory
		}
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		base := strings.TrimSuffix(entry.Name(), ".up.sql")
		parts := strings.SplitN(base, "_", 2)
		if len(parts) != 2 {
			continue // Invalid filename format
		}

		// Parse version
		var version int
		if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
			continue // Invalid version number
		}

		// Read up migration
		upPath := filepath.Join(dir, entry.Name())
		upContent, err := os.ReadFile(upPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read up migration: %w", err)
		}

		// Read down migration
		downPath := filepath.Join(dir, base+".down.sql")
		downContent, err := os.ReadFile(downPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read down migration: %w", err)
			}
			downContent = []byte{} // Optional
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    parts[1],
			Up:      string(upContent),
			Down:    string(downContent),
		})
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// ensureMigrationsTable creates the migrations tracking table
func (s *Store) ensureMigrationsTable() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at INTEGER
		)
	`)
	return err
}

// getCurrentVersion returns the current schema version
func (s *Store) getCurrentVersion() (int, error) {
	var version int
	err := s.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// RollbackMigration rolls back the most recent migration
func (s *Store) RollbackMigration() error {
	var version int
	var name string
	err := s.db.QueryRow(
		"SELECT version, name FROM schema_migrations ORDER BY version DESC LIMIT 1",
	).Scan(&version, &name)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no migrations to rollback")
		}
		return fmt.Errorf("failed to get latest migration: %w", err)
	}

	// Load the migration
	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		wd, _ := os.Getwd()
		dir = filepath.Join(wd, "migrations")
	}

	migrations, err := LoadMigrations(dir)
	if err != nil {
		return err
	}

	var migration Migration
	for _, m := range migrations {
		if m.Version == version {
			migration = m
			break
		}
	}

	if migration.Down == "" {
		return fmt.Errorf("no down migration for version %d", version)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Execute down migration
	if _, err := tx.Exec(migration.Down); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	// Remove migration record
	if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = ?", version); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	return nil
}