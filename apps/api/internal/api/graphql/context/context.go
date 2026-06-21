// Package gqlcontext provides GraphQL context utilities.
package gqlcontext

import (
	"context"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
)

type contextKey string

const (
	operatorKey contextKey = "operator"
	requestIDKey contextKey = "requestID"
)

// WithOperator adds the operator to the context.
func WithOperator(ctx context.Context, op *operator.Operator) context.Context {
	return context.WithValue(ctx, operatorKey, op)
}

// GetOperator retrieves the operator from context.
func GetOperator(ctx context.Context) (*operator.Operator, bool) {
	val := ctx.Value(operatorKey)
	if val == nil {
		return nil, false
	}
	op, ok := val.(*operator.Operator)
	return op, ok
}

// MustGetOperator retrieves the operator or panics.
func MustGetOperator(ctx context.Context) *operator.Operator {
	op, ok := GetOperator(ctx)
	if !ok {
		panic("operator not found in context")
	}
	return op
}

// WithRequestID adds a request ID to the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID retrieves the request ID from context.
func GetRequestID(ctx context.Context) string {
	val := ctx.Value(requestIDKey)
	if val == nil {
		return ""
	}
	return val.(string)
}