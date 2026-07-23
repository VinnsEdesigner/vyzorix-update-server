package storage

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
)

// SQLiteTxManager implements transaction.TxManager for SQLite.
type SQLiteTxManager struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewTxManager creates a new SQLite transaction manager.
func NewTxManager(db *sql.DB) transaction.TxManager {
	return &SQLiteTxManager{db: db}
}

// NewTxManagerWithLogger creates a new SQLite transaction manager with a logger.
func NewTxManagerWithLogger(db *sql.DB, logger *slog.Logger) transaction.TxManager {
	return &SQLiteTxManager{db: db, logger: logger}
}

// BeginTx starts a new transaction.
func (t *SQLiteTxManager) BeginTx() (*sql.Tx, error) {
	return t.db.Begin()
}

// SafeRollback handles transaction rollback safely, logging any errors.
// This should be used in defer statements to ensure rollback errors are not silently discarded.
func SafeRollback(tx *sql.Tx, logger *slog.Logger) {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone && logger != nil {
		logger.Error("transaction rollback failed", "error", err)
	}
}

// SafeRollbackNoLog handles transaction rollback safely without logging.
// Use this when you don't have access to a logger but still want to check for errors.
func SafeRollbackNoLog(tx *sql.Tx) {
	_ = tx.Rollback() // We intentionally ignore the error here as we can't do much about it
	// Note: sql.ErrTxDone is returned if the transaction was already committed or rolled back
	// which is not an error condition in most cases
}

// WithTx executes a function within a transaction.
func (t *SQLiteTxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	txCtx := transaction.ContextWithTx(ctx, tx)

	defer func() {
		if p := recover(); p != nil {
			SafeRollback(tx, t.logger)
			panic(p)
		}
	}()

	if err := fn(txCtx); err != nil {
		SafeRollback(tx, t.logger)
		return err
	}

	return tx.Commit()
}
