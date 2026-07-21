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

// WithTx executes a function within a transaction.
func (t *SQLiteTxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	txCtx := transaction.ContextWithTx(ctx, tx)

	defer func() {
		if p := recover(); p != nil {
			if rbErr := tx.Rollback(); rbErr != nil && t.logger != nil {
				t.logger.Error("transaction rollback failed after panic", "error", rbErr)
			}
			panic(p)
		}
	}()

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil && t.logger != nil {
			t.logger.Error("transaction rollback failed", "error", rbErr, "tx_error", err)
		}
		return err
	}

	return tx.Commit()
}
