// Package gqlcontext provides GraphQL context utilities.
package gqlcontext

import (
	"context"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
)

type contextKey string

const (
	operatorKey       contextKey = "operator"
	requestIDKey      contextKey = "requestID"
	metadataKey       contextKey = "requestMetadata"
	organizationIDKey contextKey = "organizationID"
)

// RequestMetadata holds request metadata.
type RequestMetadata struct {
	ClientIP  string
	UserAgent string
}

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

	str, ok := val.(string)
	if !ok {
		return ""
	}

	return str
}

// WithRequestMetadata adds request metadata to the context.
func WithRequestMetadata(ctx context.Context, clientIP, userAgent string) context.Context {
	return context.WithValue(ctx, metadataKey, RequestMetadata{
		ClientIP:  clientIP,
		UserAgent: userAgent,
	})
}

// GetRequestMetadata retrieves the request metadata from context.
func GetRequestMetadata(ctx context.Context) (RequestMetadata, bool) {
	val := ctx.Value(metadataKey)
	if val == nil {
		return RequestMetadata{}, false
	}

	meta, ok := val.(RequestMetadata)

	return meta, ok
}

// GetClientIP retrieves the client IP from context.
func GetClientIP(ctx context.Context) string {
	if meta, ok := GetRequestMetadata(ctx); ok {
		return meta.ClientIP
	}

	return ""
}

// GetUserAgent retrieves the user agent from context.
func GetUserAgent(ctx context.Context) string {
	if meta, ok := GetRequestMetadata(ctx); ok {
		return meta.UserAgent
	}

	return ""
}

// WithOrganizationID adds the organization ID to the context.
func WithOrganizationID(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, organizationIDKey, orgID)
}

// GetOrganizationID retrieves the organization ID from context.
func GetOrganizationID(ctx context.Context) string {
	val := ctx.Value(organizationIDKey)
	if val == nil {
		return ""
	}

	orgID, ok := val.(string)
	if !ok {
		return ""
	}

	return orgID
}
