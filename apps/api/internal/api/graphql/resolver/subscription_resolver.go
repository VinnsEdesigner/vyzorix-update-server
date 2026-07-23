// Package resolver provides GraphQL resolver implementations.
package resolver

import (
	"errors"

	"github.com/graphql-go/graphql"
)

// ErrSubscriptionsNotSupported is returned when GraphQL subscriptions are not available.
// Subscriptions in this implementation use WebSocket-based transport instead.
var ErrSubscriptionsNotSupported = errors.New("subscriptions are handled via WebSocket")

// ============================================================.
// Subscription Resolvers.
// ============================================================.

// DeviceUpdated resolves the deviceUpdated subscription field.
// Subscriptions are handled via WebSocket, not GraphQL queries.
func (r *Resolver) DeviceUpdated(p graphql.ResolveParams) (interface{}, error) {
	return nil, ErrSubscriptionsNotSupported
}

// TelemetryReceived resolves the telemetryReceived subscription field.
// Subscriptions are handled via WebSocket, not GraphQL queries.
func (r *Resolver) TelemetryReceived(p graphql.ResolveParams) (interface{}, error) {
	return nil, ErrSubscriptionsNotSupported
}

// CommandStatusChanged resolves the commandStatusChanged subscription field.
// Subscriptions are handled via WebSocket, not GraphQL queries.
func (r *Resolver) CommandStatusChanged(p graphql.ResolveParams) (interface{}, error) {
	return nil, ErrSubscriptionsNotSupported
}

// OrganizationEvent resolves the organizationEvent subscription field.
// Subscriptions are handled via WebSocket, not GraphQL queries.
func (r *Resolver) OrganizationEvent(p graphql.ResolveParams) (interface{}, error) {
	return nil, ErrSubscriptionsNotSupported
}

// MemberEvent resolves the memberEvent subscription field.
// Subscriptions are handled via WebSocket, not GraphQL queries.
func (r *Resolver) MemberEvent(p graphql.ResolveParams) (interface{}, error) {
	return nil, ErrSubscriptionsNotSupported
}
