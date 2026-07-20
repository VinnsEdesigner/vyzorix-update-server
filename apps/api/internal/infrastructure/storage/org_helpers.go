package storage

import (
	"database/sql"
	"strings"
	"time"
)

// boolToInt converts a bool to int (0 or 1) for SQLite storage.
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// nullTimePtr returns a sql.NullInt64 from a time.Time pointer (stored as milliseconds).
func nullTimePtr(t *time.Time) sql.NullInt64 {
	if t == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: t.UnixMilli(), Valid: true}
}

// isUniqueConstraintError checks if the error is a unique constraint violation.
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
