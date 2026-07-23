// Package audit provides audit logging functionality for security events.
package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
)

// SeparateDBConfig holds configuration for separate audit database.
type SeparateDBConfig struct {
	Path string // Path to the separate audit database file
}

// NewSeparateDBAuditRepository creates a new repository that writes to a separate database.

// If the database file doesn't exist, it will be created with the appropriate schema.
func NewSeparateDBAuditRepository(cfg SeparateDBConfig) (*SeparateDBRepository, error) {
	// Ensure directory exists
	dir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Open the separate database
	db, err := sql.Open("sqlite3", cfg.Path)
	if err != nil {
		return nil, err
	}

	// Create schema if needed
	if err := createAuditSchema(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &SeparateDBRepository{db: db}, nil
}

// SeparateDBRepository writes audit logs to a separate SQLite database.

type SeparateDBRepository struct {
	db *sql.DB
}

// createAuditSchema creates the audit log table in the separate database.
func createAuditSchema(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS audit_logs (
			id            TEXT PRIMARY KEY,
			operator_id   TEXT,
			action        TEXT NOT NULL,
			resource_type TEXT,
			resource_id   TEXT,
			ip_address    TEXT,
			user_agent    TEXT,
			result        TEXT NOT NULL,
			metadata      TEXT,
			created_at    INTEGER NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create index for querying by operator
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_audit_operator
		ON audit_logs(operator_id, created_at DESC)
	`)
	if err != nil {
		return err
	}

	// Create index for querying by action type
	_, err = db.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_audit_action
		ON audit_logs(action, created_at DESC)
	`)
	if err != nil {
		return err
	}

	return nil
}

// Log writes an audit entry to the separate database.
func (r *SeparateDBRepository) Log(ctx context.Context, entry *Entry) error {
	metadataJSON := ""

	if entry.Metadata != nil {
		data, err := json.Marshal(entry.Metadata)
		if err != nil {
			return err
		}
		metadataJSON = string(data)
	}

	query := `
		INSERT INTO audit_logs (id, operator_id, action, resource_type, resource_id, ip_address, user_agent, result, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		entry.ID,
		entry.OperatorID,
		string(entry.Action),
		entry.ResourceType,
		entry.ResourceID,
		entry.IPAddress,
		entry.UserAgent,
		string(entry.Result),
		metadataJSON,
		entry.CreatedAt.Unix(),
	)

	return err
}

// Close closes the database connection.
func (r *SeparateDBRepository) Close() error {
	return r.db.Close()
}

// GetDB returns the underlying database connection for testing purposes.
func (r *SeparateDBRepository) GetDB() *sql.DB {
	return r.db
}

// RepositoryForSeparateDB wraps a separate DB repository to implement the Repository interface.
type RepositoryForSeparateDB struct {
	*SeparateDBRepository
}

// Compile-time check that RepositoryForSeparateDB implements Repository interface
var _ interface {
	Log(ctx context.Context, entry *Entry) error
} = (*SeparateDBRepository)(nil)

// DefaultAuditDBPath returns the default path for the audit database.
func DefaultAuditDBPath(dataDir string) string {
	return filepath.Join(dataDir, "audit", "audit.db")
}
