package storage

import (
	"context"
	"database/sql"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
)

// SQLiteTxManager implements transaction.TxManager for SQLite.
type SQLiteTxManager struct {
	db *sql.DB
}

// NewTxManager creates a new SQLite transaction manager.
func NewTxManager(db *sql.DB) transaction.TxManager {
	return &SQLiteTxManager{db: db}
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
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(txCtx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
