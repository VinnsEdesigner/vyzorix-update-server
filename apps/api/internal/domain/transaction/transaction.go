package transaction

import (
	"context"
	"database/sql"
)

// ctxKey is a custom type for context keys to avoid collisions.
type ctxKey string

const (
	// TxKey is the context key for storing the transaction.
	TxKey ctxKey = "tx"
)

// ContextWithTx returns a new context with the transaction attached.
func ContextWithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, TxKey, tx)
}

// TxFromContext retrieves the transaction from context, if any.
func TxFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(TxKey).(*sql.Tx)
	return tx, ok
}

// TxManager defines the interface for transaction management.
type TxManager interface {
	// BeginTx starts a new transaction.
	BeginTx() (*sql.Tx, error)
	// WithTx executes a function within a transaction.
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
