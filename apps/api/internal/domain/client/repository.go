package client

import (
	"context"
	"time"
)

// Repository defines the interface for client data access.
type Repository interface {
	// FindByID retrieves a client by ID.
	FindByID(ctx context.Context, id string) (*Client, error)
	
	// FindByOperatorID retrieves all clients for an operator.
	FindByOperatorID(ctx context.Context, operatorID string) ([]*Client, error)
	
	// FindAll retrieves all clients (admin use).
	FindAll(ctx context.Context, limit, offset int) ([]*Client, int, error)
	
	// Create creates a new client.
	Create(ctx context.Context, client *Client, secret string) (*Client, string, error)
	
	// Update updates a client.
	Update(ctx context.Context, client *Client) error
	
	// Delete deletes a client.
	Delete(ctx context.Context, id string) error
	
	// RotateSigningKey rotates the signing key with a grace period.
	RotateSigningKey(ctx context.Context, clientID string, gracePeriod time.Duration) (*SigningKey, string, error)
	
	// ValidateSigningKey validates a signing key.
	ValidateSigningKey(ctx context.Context, clientID, signature, payload, timestamp string) error
	
	// GetHmacKey retrieves the HMAC signing key for a client.
	// Returns the key and ok=false if client not found or inactive.
	GetHmacKey(ctx context.Context, clientID string) (string, bool)
}
