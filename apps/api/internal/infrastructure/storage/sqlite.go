package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Config holds SQLite configuration.
type Config struct {
	Path            string
	JournalMode    string // WAL, DELETE, etc.
	CacheSize      int    // KB, negative means MB
	BusyTimeout    int    // milliseconds
	ForeignKeys    bool
}

// DefaultConfig returns the default SQLite configuration.
func DefaultConfig(dbPath string) *Config {
	return &Config{
		Path:         dbPath,
		JournalMode:  "WAL",
		CacheSize:    -2000, // 2GB cache
		BusyTimeout:  5000,  // 5 seconds
		ForeignKeys:  true,
	}
}

// buildDSN builds the SQLite DSN from config.
func (c *Config) buildDSN() string {
	dsn := c.Path
	
	// Ensure directory exists.
	dir := filepath.Dir(c.Path)
	if dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}
	
	// Add query parameters.
	params := "?_journal_mode=" + c.JournalMode
	params += "&_cache_size=" + fmt.Sprintf("%d", c.CacheSize)
	params += "&_busy_timeout=" + fmt.Sprintf("%d", c.BusyTimeout)
	params += "&_foreign_keys="
	if c.ForeignKeys {
		params += "1"
	} else {
		params += "0"
	}
	
	return dsn + params
}

// SQLite provides the base SQLite database connection.
type SQLite struct {
	db *sql.DB
}

// Open opens a SQLite database connection.
func Open(cfg *Config) (*SQLite, error) {
	db, err := sql.Open("sqlite3", cfg.buildDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool.
	db.SetMaxOpenConns(1) // SQLite doesn't handle concurrent writes well
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	// Verify connection.
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &SQLite{db: db}, nil
}

// DB returns the underlying sql.DB.
func (s *SQLite) DB() *sql.DB {
	return s.db
}

// Close closes the database connection.
func (s *SQLite) Close() error {
	return s.db.Close()
}

// Ping checks if the database is reachable.
func (s *SQLite) Ping() error {
	return s.db.Ping()
}

// BeginTx starts a new transaction.
func (s *SQLite) BeginTx() (*sql.Tx, error) {
	return s.db.Begin()
}
